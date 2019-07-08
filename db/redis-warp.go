package db

import (
	"github.com/go-redis/redis"
	"time"
)

//=================================
type RedisBase struct {
	rdb *redis.Client
}

func (r *RedisBase) Init(c *redis.Client) {
	r.rdb = c
}

func (r *RedisBase) Ping() *redis.StatusCmd {
	return r.rdb.Ping()
}
func (r *RedisBase) String() string {
	return r.rdb.String()
}

//=================================
type RedisRead struct {
	*RedisBase
}

func (r *RedisRead) HMGet(key string, fields ...string) *redis.SliceCmd {
	return r.rdb.HMGet(key, fields...)
}

func (r *RedisRead) SMembers(key string) *redis.StringSliceCmd {
	return r.rdb.SMembers(key)
}

func (r *RedisRead) Get(key string) *redis.StringCmd {
	return r.rdb.Get(key)
}

//=================================
type RedisMQ struct {
	*RedisBase
}

func (r *RedisMQ) Publish(channel string, message interface{}) *redis.IntCmd {
	return r.rdb.Publish(channel, message)
}

func (r *RedisMQ) Subscribe(channels ...string) *redis.PubSub {
	return r.rdb.Subscribe(channels...)
}

func (r *RedisMQ) RPush(key string, values ...interface{}) *redis.IntCmd {
	return r.rdb.RPush(key, values...)
}

func (r *RedisMQ) BLPop(timeout time.Duration, keys ...string) *redis.StringSliceCmd {
	return r.rdb.BLPop(timeout, keys...)
}
