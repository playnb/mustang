package log

import (
	"fmt"
	log "github.com/cihub/seelog"
)

func init() {
	//gLogger, _ = New("debug", "")
	logger, err := log.LoggerFromConfigAsFile("seelog.xml")
	if err != nil {
		fmt.Println("err parsing config log file", err)
		return
	}
	log.ReplaceLogger(logger)
}

var Release = false

func Dev(format string, a ...interface{}) {
	if Release {
		return
	}
	//gLogger.Debug(format, a...)
	log.Infof(format, a...)
}

func Debug(format string, a ...interface{}) {
	//gLogger.Debug(format, a...)
	log.Debugf(format, a...)
}

func Trace(format string, a ...interface{}) {
	//gLogger.Trace(format, a...)
	log.Tracef(format, a...)
}

func Error(format string, a ...interface{}) {
	//gLogger.Error(format, a...)
	log.Errorf(format, a...)
}

func Fatal(format string, a ...interface{}) {
	//gLogger.Fatal(format, a...)
	log.Criticalf(format, a...)
}

func Close() {
	//gLogger.Close()
	log.Flush()
}
