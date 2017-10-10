package fsm

//IState 状态接口
type IState interface {
	GetName() string                            //获取状态名
	OnEnter(target ITarget)                     //进入状态
	OnLeave(target ITarget)                     //结束状态
	Action(action string, target ITarget) Error //动作
}

//Transition 状态跳转关系
type Transition struct {
	Event  string
	From   IState
	To     IState
	Action string
}

//ITarget 状态机目标接口
type ITarget interface {
	GetCurrentState() string
	SetCurrentState(string)
}

type actionFunc func(target ITarget)

//ActionState 状态
type ActionState struct {
	AllFunc map[string]actionFunc
}

//RegAction 注册Action
func (state *ActionState) RegAction(action string, callBack actionFunc) {
	if state.AllFunc == nil {
		state.AllFunc = make(map[string]actionFunc)
	}
	state.AllFunc[action] = callBack
}

//Action 调用Action
func (state *ActionState) Action(action string, target ITarget) Error {
	if state.AllFunc == nil {
		return NewActionError(action, target.GetCurrentState())
	}
	if f, ok := state.AllFunc[action]; ok == true {
		f(target)
		return nil
	}
	return NewActionError(action, target.GetCurrentState())
}
