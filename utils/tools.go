package utils

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"runtime/debug"
	"strings"
	"time"
)

func init() {
	os.MkdirAll("log", os.ModePerm)
	os.MkdirAll("panic", os.ModePerm)
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

func PrintPanicStack() {
	if err := recover(); err != nil {
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

		fmt.Println(err)
		fmt.Println(string(debug.Stack()))

		file, _ := os.OpenFile("./panic/"+fileName, os.O_CREATE, 0666)
		if file != nil {
			file.WriteString(fmt.Sprintln(err))
			file.WriteString("\n==============\n")
			file.WriteString(string(debug.Stack()))
			file.Close()
		}
	}
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
