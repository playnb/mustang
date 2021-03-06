package fsm

import "fmt"

import "strings"

const TimerEvent = "__TimerEvent__"     //定时器事件
const TimeOutEvent = "__TimeOutEvent__" //超时事件

const DefaultAction = "__DefaultAction__"
const TimerAction = "__TimerAction__"
const TimeOutAction = "__TimeOutAction__" //超时事件

const InitStateName = "__InitState__"

//IState 状态接口
type IState interface {
	GetName() string                            //获取状态名
	OnEnter(target ITarget)                     //进入状态
	OnLeave(target ITarget)                     //结束状态
	Action(action string, target ITarget) Error //动作
	Init()
}

//Transition 状态跳转关系
type Transition struct {
	Event  string
	From   IState
	To     IState
	Action string
}

func (trans *Transition) String() string {
	return fmt.Sprintf("[Event:%v From:%s To:%s Action:%s]", trans.Event, trans.From.GetName(), trans.To.GetName(), trans.Action)
}

//ITarget 状态机目标接口
type ITarget interface {
	GetCurrentState() string
	SetCurrentState(string)
	GetChangeStateTime() uint64
	SetChangeStateTime(uint64)
	GetFSM() *StateMachine
}

type actionFunc func(target ITarget)

//ActionState 状态
type ActionState struct {
	AllFunc map[string]actionFunc
}

//RegAction 注册Action
func (state *ActionState) RegAction(action string, callBack actionFunc) {
	action = strings.ToLower(action)

	if state.AllFunc == nil {
		state.AllFunc = make(map[string]actionFunc)
	}
	state.AllFunc[action] = callBack
}

//Action 调用Action
func (state *ActionState) Action(action string, target ITarget) Error {
	action = strings.ToLower(action)

	if state.AllFunc == nil {
		return NewActionError(action, target.GetCurrentState())
	}
	if f, ok := state.AllFunc[action]; ok == true {
		f(target)
		//log.Debug("========> f:%v", f)
		return nil
	}
	return NewActionError(action, target.GetCurrentState())
}
