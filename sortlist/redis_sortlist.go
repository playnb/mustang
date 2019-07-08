package sortlist

import (
	"github.com/playnb/mustang/redis"
	"github.com/playnb/mustang/util"
	R "github.com/go-redis/redis"
	"reflect"
	"strconv"
)

func NewRedisSortList(key string) SortList {
	sl := &RedisSortList{}
	sl.users = make(map[uint64]*RedisSortElem)
	sl.key = "SORT:" + key
	return sl
}

type RedisSortElem struct {
	uniqueID uint64
	score    uint64
}

func (elem *RedisSortElem) GetUniqueID() uint64 { return elem.uniqueID }
func (elem *RedisSortElem) GetScore() uint64    { return elem.score }

type RedisSortList struct {
	users map[uint64]*RedisSortElem
	key   string
}

func (sl *RedisSortList) Save() {}
func (sl *RedisSortList) Load() {
	ret := redis.GetPoolClient().ZRevRangeWithScores(sl.key, 0, -1).Val()
	for _, s := range ret {
		uniqueID := uint64(0)
		uniqueID, err := strconv.ParseUint(reflect.ValueOf(s.Member).String(), 10, 64)
		if err == nil {
			elem := &RedisSortElem{}
			elem.uniqueID = uniqueID
			elem.score = uint64(s.Score)
			sl.users[elem.GetUniqueID()] = elem
		}
	}
}

/*
func (sl *RedisSortList) AddUser(user Sortable) {
	elem, _ := sl.users[user.GetUniqueID()]
	if elem == nil {
		elem = &RedisSortElem{}
		sl.users[user.GetUniqueID()] = elem
	}
	elem.Sortable = user
	redis.GetPoolClient().ZAdd(sl.key, R.Z{Score: float64(elem.GetScore()), Member: elem.GetUniqueID()})
}
*/

func (sl *RedisSortList) GetUser(uniqueID uint64) Sortable {
	if elem, ok := sl.users[uniqueID]; ok {
		return elem
	}
	return nil
}

func (sl *RedisSortList) RemoveUser(uniqueID uint64) {
	redis.GetPoolClient().ZRem(sl.key, uniqueID)
	delete(sl.users, uniqueID)
}

func (sl *RedisSortList) UpdateUser(uniqueID uint64, score uint64) {
	t := util.TimeStrToTimestamp("2030-12-12 12:00:00")
	t = t - uint64(util.NowTimestamp())

	elem, _ := sl.users[uniqueID]
	if elem == nil {
		elem = &RedisSortElem{}
		elem.uniqueID = uniqueID
		sl.users[elem.GetUniqueID()] = elem
	}
	elem.score = (score << 32) + t
	redis.GetPoolClient().ZAdd(sl.key, R.Z{Score: float64(elem.GetScore()), Member: elem.GetUniqueID()})
}

func (sl *RedisSortList) GetRank(uniqueID uint64) uint64 {
	if elem, ok := sl.users[uniqueID]; ok {
		ret, err := redis.GetPoolClient().ZRevRank(sl.key, strconv.FormatUint(elem.GetUniqueID(), 10)).Result()
		if err != nil {
			return 0
		}
		return uint64(ret + 1)
	}
	return 0
}

func (sl *RedisSortList) GetRankUser(rank uint64) Sortable {
	if rank == 0 {
		rank = 1
	}
	ret, err := redis.GetPoolClient().ZRevRange(sl.key, int64(rank)-1, int64(rank)-1).Result()
	if err != nil && len(ret) != 1 {
		return nil
	}
	uniqueID, err := strconv.ParseUint(ret[0], 10, 64)
	if err != nil {
		return nil
	}
	return sl.GetUser(uniqueID)
}

func (sl *RedisSortList) GetRankUsers(fromRnk int64, toRank int64) []Sortable {
	if fromRnk == 0 {
		fromRnk = 1
		toRank++
	}

	result := make([]Sortable, 0)
	ret, err := redis.GetPoolClient().ZRevRange(sl.key, int64(fromRnk)-1, int64(toRank)-1).Result()
	if err != nil {
		return nil
	}
	for _, s := range ret {
		uniqueID, err := strconv.ParseUint(s, 10, 64)
		if err == nil {
			elem := sl.GetUser(uniqueID)
			if elem != nil {
				result = append(result, elem)
			}
		}
	}
	return result
}

func (sl *RedisSortList) GetUserCount() uint64 {
	return uint64(redis.GetPoolClient().ZCard(sl.key).Val())
}
