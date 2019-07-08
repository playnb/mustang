package util_db

import (
	time_guard "cell/common/mustang/time-guard"
	"gopkg.in/mgo.v2"
)

type Collection struct {
	*mgo.Collection
	session *mgo.Session
	guard   *time_guard.Guard
}

func (c *Collection) CloseSession() {
	//c.session.Close()
	time_guard.RemoveGuard(c.guard)
}

func NewCollection(session *mgo.Session, mc *mgo.Collection, desc string) *Collection {
	c := &Collection{
		Collection: mc,
		session:    session,
	}
	c.guard = time_guard.NewGuard(desc)
	return c
}