package worker

import (
	"fmt"
	"testing"
	"time"
)

func Test_DefaultPool(t *testing.T) {
	DefaultPool().Start(100, func() {
		if err := recover(); err != nil {
			fmt.Println(err)
		}
	})

	toDie := func() {
		fmt.Println("我是找死的")
		i := &struct {
			Data uint32
		}{}
		i = nil
		i.Data = 100
	}
	DefaultPool().Do(0, false, toDie)

	DefaultPool().Dump()

	DefaultPool().Do(0, false, toDie)
	DefaultPool().Do(0, false, toDie)

	for _, v := range DefaultPool().workers {
		if v.working == false {
			t.Fail()
		}
	}

	fmt.Println("少少等待1s")
	time.Sleep(time.Second)
}
