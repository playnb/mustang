package until

type ITimeValue interface {
	Refresh(cur uint64)
}

type TimeValueMgr struct {
	values []ITimeValue
}

func (mgr *TimeValueMgr) Add(tv ITimeValue) {
	if mgr.values == nil {
		mgr.values = make([]ITimeValue, 0, 5)
	}
	mgr.values = append(mgr.values, tv)
}
func (mgr *TimeValueMgr) Refresh(cur uint64) {
	if mgr.values != nil {
		for _, v := range mgr.values {
			v.Refresh(cur)
		}
	}
}

//==============================================================================

type TimeBoolean struct {
	timeOut   uint64
	Value     bool
	OnTimeOut func()
}

func (tb *TimeBoolean) Set(value bool, timeOut uint64) {
	tb.Value = value
	tb.timeOut = timeOut
}

func (tb *TimeBoolean) Refresh(cur uint64) {
	if tb.timeOut != 0 && tb.timeOut < cur {
		tb.Value = false
		tb.timeOut = 0
		if tb.OnTimeOut != nil {
			tb.OnTimeOut()
		}
	}
}
