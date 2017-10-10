package redis

import (
	"github.com/go-redis/redis"
)

var MAX_POOL_SIZE = 100
var redisPool chan *redis.Client
var redis_address = "192.168.175.16:6179"

//初始化Redis连接池
func InitRedisPool(address string) {
	redis_address = address
	makeRedisConn()
}

//放回Redis连接
func PutRedis(conn *redis.Client) {
	if redisPool == nil {
		redisPool = make(chan *redis.Client, MAX_POOL_SIZE)
	}
	if len(redisPool) >= MAX_POOL_SIZE {
		conn.Close()
		return
	}
	redisPool <- conn
}

//获取Redis连接
func GetRedis() *redis.Client {
	//	makeRedisConn()
	return <-redisPool
}

//创建Redis连接
func makeRedisConn() {
	if redisPool == nil {
		redisPool = make(chan *redis.Client, MAX_POOL_SIZE)
	}
	if len(redisPool) == 0 {
		go func() {
			for i := 0; i < MAX_POOL_SIZE/2; i++ {
				client := redis.NewClient(&redis.Options{
					Addr:     redis_address,
					Password: "", // no password set
					DB:       0,  // use default DB
					PoolSize: 100,
				})

				if client == nil {
					panic("Redis no alive")
				}
				PutRedis(client)
			}
		}()
	}
}

/*
//使用原有的缓冲区
var MAX_POOL_SIZE = 100
var client *redis.Client
var redis_address = "192.168.175.16:6179"

func InitRedisPool(address string) {
	redis_address = address
	client = redis.NewClient(&redis.Options{
		Addr:     redis_address,
		Password: "", // no password set
		DB:       0,  // use default DB
		PoolSize: MAX_POOL_SIZE,
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
