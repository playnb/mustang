package redislock

import (
	"cell/common/mustang/util"
	"errors"
	"github.com/go-redis/redis"
	"strconv"
	"time"
)

const (
	//默认锁失效时间
	DefaultExpiry = 8 * time.Second
	DefaultDelay  = 50 * time.Millisecond

	luaCode = `
if redis.call("get",KEYS[1]) == ARGV[1]
then
    return redis.call("del",KEYS[1])
else
    return 0
end
`
)

type Mutex struct {
	Name   string        //加锁资源名
	Expiry time.Duration //失效时间
	Delay  time.Duration
	Tries  int
	value  string
	node   *redis.Client
}

func NewMutex(name string, node *redis.Client) *Mutex {
	m := &Mutex{
		Name:  name,
		node:  node,
		value: strconv.Itoa(int(util.NextUniqueID())),
	}
	return m
}

func (m *Mutex) Lock() error {
	expiry := m.Expiry
	if expiry == 0 {
		expiry = DefaultExpiry
	}
	delay := m.Delay
	if delay == 0 {
		delay = DefaultDelay
	}
	tries := m.Tries

	for m.Tries == 0 || tries > 0 {
		tries--
		if m.node.SetNX(m.Name, m.value, expiry).Val() {
			return nil
		}
		time.Sleep(delay)
	}
	return errors.New("max tries")
}

func (m *Mutex) UnLock() {
	m.node.Eval(luaCode, []string{m.Name}, m.value)
}
