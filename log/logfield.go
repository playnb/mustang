package log

import (
	"fmt"
	"github.com/cihub/seelog"
	"github.com/json-iterator/go"
)

type Fields map[string]interface{}

var userFieldLog *TFieldLog
var fieldLog *TFieldLog

func NewFieldLog(log seelog.LoggerInterface) *TFieldLog {
	l := &TFieldLog{}
	l.log = log
	return l
}

type TFieldLog struct {
	log    seelog.LoggerInterface
	fields Fields
}

func (fl *TFieldLog) WithField(fields Fields) *TFieldLog {
	if fl == nil {
		return NewFieldLog(nil)
	}
	clone := NewFieldLog(fl.log)
	clone.fields = fields
	return clone
}

func (fl *TFieldLog) Log(message string) {
	if fl.log != nil {
		if fl.fields != nil {
			str, _ := jsoniter.MarshalToString(fl.fields)
			fl.log.Tracef("%s|%s", str, message)
		} else {
			fl.log.Tracef("%s|%s", "{}", message)
		}
	} else {
		if fl.fields != nil {
			str, _ := jsoniter.MarshalToString(fl.fields)
			fmt.Printf("%s|%s\n", str, message)
		} else {
			fmt.Printf("%s|%s\n", "{}", message)
		}
	}
}

func UserLog() *TFieldLog {
	return userFieldLog
}
func FieldLog() *TFieldLog {
	return fieldLog
}
