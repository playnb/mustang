package redislock

import (
	"github.com/go-redis/redis"
	"sync"
	"testing"
)

const LockName = "I_am_Lock"

var wg = &sync.WaitGroup{}
var f = func(r *redis.Client, M int) {
	defer wg.Done()
	m := NewMutex(LockName, r)
	for i := 0; i < M; i++ {
		m.Lock()
		n, _ := r.Get("DATA").Int()
		n++
		r.Set("DATA", n, -1)
		m.UnLock()
	}
}

func TestMutex(t *testing.T) {
	r := redis.NewClient(&redis.Options{
		Addr:     "app.playnb.net:8001",
		Password: "Ztgame+123", // no password set
		DB:       1,            // use default DB
		PoolSize: 100,
	})
	if r.Ping().Err() != nil {
		t.Fail()
	}
	r.Del("DATA")

	N := 100
	M := 10
	for i := 0; i < N; i++ {
		wg.Add(1)
		go f(r, M)
	}
	wg.Wait()
	n, _ := r.Get("DATA").Int()
	t.Log(n)
	if n != N*M {
		t.Fail()
	}
}
