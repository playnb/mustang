package fsm

import (
	"strings"
	"github.com/playnb/mustang/log"
)

type InitState struct {
	ActionState
}

func (is *InitState) GetName() string {
	return InitStateName
}
func (is *InitState) OnEnter(target ITarget) {
}
func (is *InitState) OnLeave(target ITarget) {
}
func (is *InitState) Init() {
}

//NewStateMachine 创建状态机
func NewStateMachine(transitions ...Transition) *StateMachine {
	allStates := make(map[IState]bool)
	for i := 0; i < len(transitions); i++ {
		transitions[i].Action = strings.ToLower(transitions[i].Action)
		transitions[i].Event = strings.ToLower(transitions[i].Event)
		allStates[transitions[i].From] = true
		allStates[transitions[i].To] = true
	}
	for s, _ := range allStates {
		s.Init()
	}
	for s, _ := range allStates {
		transitions = append(transitions, Transition{
			From:   s,
			To:     s,
			Event:  strings.ToLower(TimerEvent),
			Action: strings.ToLower(TimerAction),
		})
	}
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
	event = strings.ToLower(event)
	//log.Debug("%s Trigger %s", target.GetCurrentState(), event)

	trans := m.findTransMatching(target.GetCurrentState(), event)
	if trans == nil {
		return NewEventError(event, target.GetCurrentState())
	}
	//log.Debug("StateMachine: Trigger %s", trans)

	if trans.From.GetName() != trans.To.GetName() {
		log.Debug("OnLeave ===> " + trans.From.GetName())
		trans.From.OnLeave(target)

		log.Debug("OnEnter ===> " + trans.To.GetName())
		trans.To.OnEnter(target)
		target.SetCurrentState(trans.To.GetName())
	}

	//log.Debug("Action ===> " + trans.To.GetName())
	return trans.To.Action(trans.Action, target)
}

func (m *StateMachine) Timer(target ITarget) Error {
	return m.Trigger(TimerEvent, target)
}

func (m *StateMachine) findTransMatching(fromState string, event string) *Transition {
	event = strings.ToLower(event)

	for _, v := range m.transitions {
		if v.From.GetName() == fromState {
			if v.Event == event {
				return &v
			}
		}
	}
	return nil
}
