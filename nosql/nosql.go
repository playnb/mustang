package nosql

import (
	"github.com/playnb/mustang/log"
	"gopkg.in/redis.v3"
)

var Redis *redis.Client

func InitRedis(addr string, password string, db int64) {
	Redis = redis.NewClient(&redis.Options{
		Addr:     addr, // "localhost:6379",
		Password: password,
		DB:       db,
	})
	//Redis.Auth(password)
	str, err := Redis.Ping().Result()
	if err != nil {
		log.Error(err.Error())
	} else {
		log.Trace("REDIS: %s", str)
	}
}
