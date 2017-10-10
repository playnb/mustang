/*
给log做一层封装
以后用什么log库再说
*/

package log

/*
import (
	"flag"
	"fmt"

	"github.com/golang/glog"
)

var DevLog = false
var Release = false
var panic_dir string
var log_dir string

func init() {
	flag.StringVar(&panic_dir, "panic_dir", `C:\code\server\panic`, "日志位置")

	flag.Set("alsologtostderr", "true")
	flag.Set("log_dir", `C:\code\server\log`)
	flag.Set("panic_dir", `C:\code\server\panic`)
	flag.Set("v", "3")
}

func Init() {
	fmt.Println("Log Init")
}

func Dev(format string, a ...interface{}) {
	if DevLog {
		glog.Infof(format, a...)
	}
}

func Debug(format string, a ...interface{}) {
	glog.Infof(format, a...)
}

func Info(format string, a ...interface{}) {
	glog.Infof(format, a...)
}

func Trace(format string, a ...interface{}) {
	glog.Infof(format, a...)
}

func Warning(format string, a ...interface{}) {
	glog.Warningf(format, a...)
}

func Error(format string, a ...interface{}) {
	glog.Errorf(format, a...)
}

func Fatal(format string, a ...interface{}) {
	glog.Fatalf(format, a...)
}

func Flush() {
	//gLogger.Close()
	glog.Flush()
}
*/
