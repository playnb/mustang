package counter

import (
	"github.com/playnb/mustang/util"
	"github.com/go-redis/redis"
)

var (
	KeyCounter   = "C:"
	RedisCounter *redis.Client
)

func counterKey(kind string, key string) string {
	return KeyCounter + kind + ":I:" + key
}
func counterRecKey(kind string, key string) string {
	return KeyCounter + kind + ":R:" + key
}
func counterSortKey(kind string, key string) string {
	return KeyCounter + kind + ":S:" + key
}

//清空计数器
func CountNormalClear(kind string, key string) {
	RedisCounter.Del(counterKey(kind, key))
	RedisCounter.Del(counterRecKey(kind, key))
}

//访问资源按天计数
func CountNormalByDay(kind string, key string, visitor string) bool {
	key = counterKey(kind, key)
	n, _ := RedisCounter.SAdd(key, visitor).Result()
	if n == 1 {
		RedisCounter.ExpireAt(key, util.GetTomorrowZeroTime())
	}
	return n == 1
}
func CountNormalByDayDel(kind string, key string, visitor string) bool {
	key = counterKey(kind, key)
	n, _ := RedisCounter.SRem(key, visitor).Result()
	return n == 1
}
func CountNormalByDayHistory(kind string, key string, visitor string, visitorData string, maxLimit int64) bool {
	newOne := CountNormalByDay(kind, key, visitor)

	key = counterRecKey(kind, key)
	l := RedisCounter.LPush(key, visitorData).Val()
	if l >= maxLimit*2 && maxLimit > 0 {
		RedisCounter.LTrim(key, 0, maxLimit-1)
	}
	RedisCounter.ExpireAt(key, util.GetTomorrowZeroTime())
	return newOne
}

//访问资源永久计数
func CountNormalForever(kind string, key string, visitor string) bool {
	key = counterKey(kind, key)
	n, _ := RedisCounter.SAdd(key, visitor).Result()
	return n == 1
}
func CountNormalForeverDel(kind string, key string, visitor string) bool {
	key = counterKey(kind, key)
	n, _ := RedisCounter.SRem(key, visitor).Result()
	return n == 1
}
func CountNormalForeverHistory(kind string, key string, visitor string, visitorData string, maxLimit int64) bool {
	newOne := CountNormalForever(kind, key, visitor)

	key = counterRecKey(kind, key)
	l := RedisCounter.LPush(key, visitorData).Val()
	if l >= maxLimit*2 && maxLimit > 0 {
		RedisCounter.LTrim(key, 0, maxLimit-1)
	}
	return newOne
}

//获取计数
func CountNormalCounter(kind string, key string) int64 {
	key = counterKey(kind, key)
	return RedisCounter.SCard(key).Val()
}

//获取计数集合
func CountNormalAllVisitors(kind string, key string) []string {
	key = counterKey(kind, key)
	return RedisCounter.SMembers(key).Val()
}

//获取计数序列
func CountNormalRecord(kind string, key string, index int64, num int64) []string {
	if num == 0 {
		num = 1
	}
	key = counterRecKey(kind, key)
	begin := index
	end := index + num - 1
	if end < 0 {
		end = -1
	}
	return RedisCounter.LRange(key, begin, end).Val()
}

//判断visitor是否再Counter中
func CountNormalExist(kind string, key string, visitor string) bool {
	key = counterKey(kind, key)
	return RedisCounter.SIsMember(key, visitor).Val()
}

//----------------------

//按照天计数
func CountSortByDay(kind string, key string, visitor string, score float64) bool {
	key = counterSortKey(kind, key)
	n := RedisCounter.ZAdd(key, redis.Z{Score: score, Member: visitor}).Val()
	if n == 1 {
		RedisCounter.ExpireAt(key, util.GetTomorrowZeroTime())
	}
	return n == 1
}

//永久计数
func CountSortForever(kind string, key string, visitor string, score float64) bool {
	key = counterSortKey(kind, key)
	return RedisCounter.ZAdd(key, redis.Z{Score: score, Member: visitor}).Val() == 1
}

//是否在计数器中
func CountSortExist(kind string, key string, visitor string) bool {
	key = counterSortKey(kind, key)
	_, err := RedisCounter.ZScore(key, visitor).Result()
	return err == nil
}

//删除计数器的一个元素
func CountSortDelete(kind string, key string, visitor string) bool {
	key = counterSortKey(kind, key)
	return RedisCounter.ZRem(key, visitor).Val() == 1
}

//获取所有访问者(时间复杂度O(n*log(m)))
func CountSortAllVisitors(kind string, key string) []string {
	key = counterSortKey(kind, key)
	return RedisCounter.ZRange(key, 0, -1).Val()
}

//获取计数器数量
func CountSortCount(kind string, key string) int64 {
	key = counterSortKey(kind, key)
	return RedisCounter.ZCard(key).Val()
}

type IdScore struct {
	ID    string
	Score float64
}

//获取计数器中的序列
func CountSortRange(kind string, key string, index int, num int) []*IdScore {
	var ok bool
	key = counterSortKey(kind, key)
	end := int64(index + num - 1)
	if num == -1 {
		end = -1
	}
	val := RedisCounter.ZRevRangeWithScores(key, int64(index), end).Val()
	if len(val) > 0 {
		ret := make([]*IdScore, 0, len(val))
		for _, v := range val {
			item := &IdScore{}
			item.ID, ok = v.Member.(string)
			if !ok {
				continue
			}
			item.Score = v.Score
			ret = append(ret, item)
		}
		return ret
	}
	return nil
}

func CountSortCounter(kind string, key string) int64 {
	key = counterSortKey(kind, key)
	return RedisCounter.ZCard(key).Val()
}

func CountSortClear(kind string, key string) {
	RedisCounter.Del(counterSortKey(kind, key))
}
