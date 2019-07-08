package time_guard

import (
	"github.com/playnb/mustang/util"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

type Guard struct {
	UserData   interface{}
	createTime int64
	uniqueID   uint64
}

type _GuardMgr struct {
	guards     sync.Map
	sequenceID uint64
	Threshold  int64
}

func (mgr *_GuardMgr) Add(g *Guard) {
	g.createTime = util.AppTimestamp()
	g.uniqueID = atomic.AddUint64(&mgr.sequenceID, 1)
	mgr.guards.Store(g.uniqueID, g)
}

func (mgr *_GuardMgr) Remove(g *Guard) {
	mgr.guards.Delete(g.uniqueID)
}

func (mgr *_GuardMgr) Check() {
	now := util.AppTimestamp()
	mgr.guards.Range(func(key, value interface{}) bool {
		g := value.(*Guard)
		if now-g.createTime > mgr.Threshold {
			fmt.Println(g.UserData)
		}
		return true
	})
}

var _gm *_GuardMgr

func init() {
	_gm = &_GuardMgr{}
	_gm.Threshold = int64(60 * 1000)	//60S
	go func() {
		t := time.NewTicker(time.Minute)

		for {
			select {
			case <-t.C:
				_gm.Check()
			}
		}
	}()
}

func NewGuard(userData interface{}) *Guard {
	g := &Guard{}
	g.UserData = userData
	_gm.Add(g)
	return g
}

func RemoveGuard(g *Guard) {
	_gm.Remove(g)
}
