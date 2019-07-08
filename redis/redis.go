package redis

import (
	"fmt"
	"github.com/go-redis/redis"
	"sync"
	"time"
)

var _MaxPoolSize = 10
var redisPool chan *redis.Client
var redisAddress = "192.168.175.16:6179"
var redisPassword = ""
var redisMutex sync.Mutex
var redisClient *redis.Client

const NilReplyError = redis.Nil

func IsNilData(err error) bool {
	return err == NilReplyError
}

var redisClientMap map[uint64]*redis.Client = make(map[uint64]*redis.Client)
var redisMapMutex sync.Mutex

func GetPoolClient() *redis.Client {
	return redisClient
}

/*
func GetDBClient(db uint64) *redis.Client {
	redisMapMutex.Lock()
	defer redisMapMutex.Unlock()
	if client, ok := redisClientMap[db]; ok {
		return client
	}
	client := redis.NewClient(&redis.Options{
		Addr:     redisAddress,
		Password: "",      // no password set
		DB:       int(db), // use default DB
		PoolSize: 100,
	})
	redisClientMap[db] = client
	return client
}
*/
type getputRedisStruct struct {
	Count    int
	FuncName string
}

func (r *getputRedisStruct) String() string {
	return fmt.Sprintf("%d %v", r.Count, r.FuncName)
}

//var getputRedis map[string]*getputRedisStruct
//var getputRedisMutex sync.Mutex

//初始化Redis连接池
func InitRedisPool(address string, password string) {
	redisAddress = address
	redisPassword = password
	makeRedisConn()

	//	getputRedisMutex.Lock()
	//	getputRedis = make(map[string]*getputRedisStruct)
	//	getputRedisMutex.Unlock()
}

//放回Redis连接
func PutRedis(conn *redis.Client) {
	redisMutex.Lock()
	defer redisMutex.Unlock()

	if redisPool == nil {
		//redisPool = make(chan *redis.Client, _MaxPoolSize)
		panic("Redis do not inited!")
	}
	if len(redisPool) >= _MaxPoolSize {
		conn.Close()
		return
	}

	//	getputRedisMutex.Lock()
	//	pName := fmt.Sprintf("%p", conn)
	//	if _, ok := getputRedis[pName]; ok == true {
	//		getputRedis[pName].Count--
	//		getputRedis[pName].FuncName = ""
	//	}
	//	getputRedisMutex.Unlock()

	redisPool <- conn
}

//获取Redis连接
func GetRedis() *redis.Client {
	//	getputRedisMutex.Lock()
	//	if len(redisPool) == 0 {
	//		fmt.Printf("=======>GetRedis %d %v\n", len(redisPool), getputRedis)
	//	}
	//	getputRedisMutex.Unlock()

	client := <-redisPool

	//	getputRedisMutex.Lock()
	//	pName := fmt.Sprintf("%p", client)
	//	if _, ok := getputRedis[pName]; ok == false {
	//		getputRedis[pName] = &getputRedisStruct{}
	//	}
	//	getputRedis[pName].Count++
	//	getputRedis[pName].FuncName = util.GetFuncName(1)
	//	getputRedisMutex.Unlock()

	return client
}

func makeOneRedisConn() {
	client := redis.NewClient(&redis.Options{
		Addr:     redisAddress,
		Password: redisPassword, // no password set
		DB:       0,             // use default DB
		PoolSize: 100,
	})

	if client == nil {
		panic("Redis no alive")
	}
	//redisPool <- client

	if redisClient == nil {
		redisClient = client
		makeOneRedisConn()
	} else {
		PutRedis(client)
	}
}

//创建Redis连接
func makeRedisConn() {
	if redisPool == nil {
		redisPool = make(chan *redis.Client, _MaxPoolSize)
	}
	if len(redisPool) == 0 {
		//go func() {
		for i := 0; i < _MaxPoolSize-1; i++ {
			makeOneRedisConn()
		}
		//}()

		go func() {
			for {
				time.Sleep(time.Millisecond * 10)

				if len(redisPool) == 0 {
					//fmt.Println("------------>补充RedisClient")
					for i := 0; i < _MaxPoolSize/2; i++ {
						makeOneRedisConn()
					}
				} else {
					time.Sleep(time.Second)
				}
			}
		}()
	}
}

/*
//使用原有的缓冲区
var _MaxPoolSize = 100
var client *redis.Client
var redis_address = "192.168.175.16:6179"

func InitRedisPool(address string) {
	redis_address = address
	client = redis.NewClient(&redis.Options{
		Addr:     redis_address,
		Password: "", // no password set
		DB:       0,  // use default DB
		PoolSize: _MaxPoolSize,
	})
}

func PutRedis(conn *redis.Client) {
}

func GetRedis() *redis.Client {
	return client
}
*/

/*
角色 U:USER_ID:P
*/
