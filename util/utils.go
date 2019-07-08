package util

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"reflect"
	"runtime"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func IsLinux() bool {
	return runtime.GOOS == "linux"
}

func IsWindows() bool {
	return runtime.GOOS == "windows"
}

func SelectByTenThousand(per int) bool {
	return (RandInt(9999) + 1) <= per
}

func SelectByPercent(per int) bool {
	return (RandInt(99) + 1) <= per
}

func GetFuncName(i int) string {
	pc, _, _, _ := runtime.Caller(i - 1)
	funcInfo := runtime.FuncForPC(pc)
	return funcInfo.Name()
}

const indentStr = "    "

//把一个类型的描述作为字符串输出
func DumpTypeToString(t reflect.Type) string {
	return dumpTypeToString(t, 0, make(map[reflect.Type]bool))
}
func dumpTypeToString(t reflect.Type, indentation int, typeList map[reflect.Type]bool) string {
	retStr := ""
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	indent := ""
	for i := 0; i < indentation; i++ {
		indent += indentStr
	}

	if typeList[t] == true {
		retStr += fmt.Sprintf("{\n")
		retStr += fmt.Sprintf("%s\"__type__\":\"%s\"", indent+indentStr, t)
		retStr += fmt.Sprintf("\n%s}", indent)
		return retStr
	}

	typeList[t] = true
	if t.Kind() == reflect.Struct {
		indentIn := indent + indentStr

		retStr += fmt.Sprintf("{\n")
		retStr += fmt.Sprintf("%s\"__type__\":\"%s\"", indentIn, t)
		for i := 0; i < t.NumField(); i++ {
			j := t.Field(i).Tag.Get(`json`)
			if len(j) > 0 && j != "-" {
				if t.Field(i).Type.Kind() == reflect.Ptr {
					retStr += fmt.Sprintf(",\n%s\"%s\":%s", indentIn, j, dumpTypeToString(t.Field(i).Type.Elem(), indentation+1, typeList))
				} else if t.Field(i).Type.Kind() == reflect.Slice {
					retStr += fmt.Sprintf(",\n%s\"%s[]\":%s", indentIn, j, dumpTypeToString(t.Field(i).Type.Elem(), indentation+1, typeList))
				} else if t.Field(i).Type.Kind() == reflect.Struct {
					retStr += fmt.Sprintf(",\n%s\"%s\":%s", indentIn, j, dumpTypeToString(t.Field(i).Type, indentation+1, typeList))
				} else {
					retStr += fmt.Sprintf(",\n%s\"%s\":\"%s\"", indentIn, j, t.Field(i).Type)
				}
			}
		}
		retStr += fmt.Sprintf("\n%s}", indent)
	} else if t.Kind() == reflect.Slice {
		retStr += fmt.Sprintf("%s[%s]", indent, dumpTypeToString(t.Elem(), indentation+1, typeList))
	} else {
		retStr += fmt.Sprintf("\"%s\"", t)
	}
	return retStr
}

func DumpTypeToJson(t reflect.Type) string {
	desc := dumpTypeToJson(t, make(map[reflect.Type]bool))
	str, _ := json.Marshal(desc)
	return string(str)
}

func dumpTypeToJson(t reflect.Type, typeList map[reflect.Type]bool) interface{} {
	description := make(map[string]interface{})
	retStr := ""
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if typeList[t] == true {
		retStr += fmt.Sprintf("{\n")
		retStr += fmt.Sprintf("%s\"__type__\":\"%s\"", indentStr, t)
		retStr += fmt.Sprintf("\n}", )
		return retStr
	}
	typeList[t] = true

	if t.Kind() == reflect.Struct {
		retStr += fmt.Sprintf("%s{\n", t)
		for i := 0; i < t.NumField(); i++ {
			j := t.Field(i).Tag.Get(`json`)
			if len(j) > 0 && j != "-" {
				if t.Field(i).Type.Kind() == reflect.Ptr {
					description[j] = dumpTypeToJson(t.Field(i).Type.Elem(), typeList)
				} else if t.Field(i).Type.Kind() == reflect.Slice {
					description[j+"[]"] = dumpTypeToJson(t.Field(i).Type.Elem(), typeList)
				} else if t.Field(i).Type.Kind() == reflect.Struct {
					description[j] = dumpTypeToJson(t.Field(i).Type, typeList)
				} else {
					description[j] = t.Field(i).Type.String()
				}
			}
		}
		return description
	} else {
		return t.String()
	}
}
