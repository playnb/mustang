package fsm

//NewStateMachine 创建状态机
func NewStateMachine(transitions ...Transition) *StateMachine {
	sm := &StateMachine{
		transitions: transitions,
	}
	return sm
}

//StateMachine 状态机
type StateMachine struct {
	transitions []Transition
}

//Trigger 出发状态机事件
func (m *StateMachine) Trigger(event string, target ITarget) Error {
	trans := m.findTransMatching(target.GetCurrentState(), event)
	if trans == nil {
		return NewEventError(event, target.GetCurrentState())
	}

	if trans.From.GetName() != trans.To.GetName() {
		trans.From.OnLeave(target)
		trans.To.OnEnter(target)
		target.SetCurrentState(trans.To.GetName())
	}
	trans.To.Action(trans.Action, target)

	return nil
}

func (m *StateMachine) findTransMatching(fromState string, event string) *Transition {
	for _, v := range m.transitions {
		if v.From.GetName() == fromState && v.Event == event {
			return &v
		}
	}
	return nil
}
