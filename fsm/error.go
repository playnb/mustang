package fsm

import "fmt"

//Error 状态机错误
type Error interface {
	error
	BadThing() string
	CurrentState() string
}

func NewEventError(badThing string, currentState string) Error {
	return &eventError{badThing: badThing, currentState: currentState}
}
func NewActionError(badThing string, currentState string) Error {
	return &actionError{badThing: badThing, currentState: currentState}
}

type eventError struct {
	badThing     string
	currentState string
}

func (e eventError) Error() string {
	return fmt.Sprintf("no transition (event[%s] state[%s])", e.badThing, e.currentState)
}
func (e eventError) BadThing() string {
	return e.badThing
}
func (e eventError) CurrentState() string {
	return e.currentState
}

type actionError struct {
	badThing     string
	currentState string
}

func (e actionError) Error() string {
	return fmt.Sprintf("no action (action[%s] state[%s])", e.badThing, e.currentState)
}
func (e actionError) BadThing() string {
	return e.badThing
}
func (e actionError) CurrentState() string {
	return e.currentState
}
