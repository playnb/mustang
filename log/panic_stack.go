package log

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"runtime/debug"
	"strings"
	"time"
)

var _panic_dir string

func InitPanic() {
	_panic_dir = "./panic"
	flag.VisitAll(func(fg *flag.Flag) {
		if fg.Name == "log_dir" {
			os.MkdirAll(fg.Value.String(), os.ModePerm)
		}
		if fg.Name == "panic_dir" {
			os.MkdirAll(fg.Value.String(), os.ModePerm)
			_panic_dir = fg.Value.String()
		}
	})

	reDirectErr()
}

func RemoveSlice(s []interface{}, index int) []interface{} {
	if index >= len(s) {
		return s
	} else if index == 0 {
		return s[1:]
	} else if index+1 == len(s) {
		return s[:index]
	} else {
		return append(s[:index], s[index+1:]...)
	}
}

func PrintArray(arr ...interface{}) string {
	str := ""
	for i, v := range arr {
		str = fmt.Sprintf("%s, (%d,%v)", str, i, v)
	}
	return str
}

var panic_time int64
var panic_count uint32
var app_name string

func MakePanic() {
	c := make(chan int, 10)
	close(c)
	c <- 1
}

func PrintPanicStack() {
	if err := recover(); err != nil {
		PrintPanicError(err)
	}
}

func PrintPanicError(err interface{}) {
	now := time.Now()
	if now.Unix() != panic_time {
		panic_time = now.Unix()
		panic_count = 0
	}
	panic_count++

	_, app_name := path.Split(strings.Replace(os.Args[0], "\\", "/", -1))
	app_name = strings.Split(app_name, ".")[0]

	fileName := fmt.Sprintf("panic_%s_%s_%d.log", app_name, now.Format("2006-01-02-15_04_05"), panic_count)
	fmt.Println(fileName)

	fmt.Println(_panic_dir + "/" + fileName)
	fmt.Println(err)
	fmt.Println(string(debug.Stack()))

	file, ferr := os.OpenFile(_panic_dir+"/"+fileName, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0666)
	defer file.Close()

	if ferr != nil {
		fmt.Println(ferr)
	}
	if file != nil {
		io.WriteString(file, fmt.Sprintln(err))
		io.WriteString(file, "\n==============\n")
		io.WriteString(file, string(debug.Stack()))
	} else {
		fmt.Println("./panic/" + fileName + "===========打开文件失败")
	}

	Trace("宕机了!: " + _panic_dir + "/" + fileName)
	Trace("%s", err)
	Trace(string(debug.Stack()))
	Flush()
}

func substr(s string, pos, length int) string {
	runes := []rune(s)
	l := pos + length
	if l > len(runes) {
		l = len(runes)
	}
	return string(runes[pos:l])
}
func GetCurrentDirectory() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		//log.Error(err.Error())
	}
	return strings.Replace(dir, "\\", "/", -1)
}
func GetParentDirectory(dirctory string) string {
	return substr(dirctory, 0, strings.LastIndex(dirctory, "/"))
}
