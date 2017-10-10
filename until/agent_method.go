package until

import (
	"github.com/playnb/mustang/log"
	"fmt"
	"reflect"
	"sync"
)

type agentMethod struct {
	method    reflect.Method
	ArgType   reflect.Type
	ReplyType reflect.Type
}

//MethodAgent 函数代理
type MethodAgent struct {
	methods map[string]*agentMethod
	orgType reflect.Type
	sync.RWMutex
}

//InitMethodAgent 初始化函数代理
func (ma *MethodAgent) InitMethodAgent(ins interface{}) {
	ma.methods = make(map[string]*agentMethod)
	ma.orgType = reflect.TypeOf(ins)
}

//RegisterFuncName 注册函数
func (ma *MethodAgent) RegisterFuncName(name string) {
	//fmt.Println(ma.orgType.NumMethod())
	ma.Lock()
	defer ma.Unlock()
	for m := 0; m < ma.orgType.NumMethod(); m++ {
		method := ma.orgType.Method(m)
		if name == method.Name {
			if method.Type.NumIn() == 2 && method.Type.NumOut() == 1 {
				//fmt.Println(method)
				ma.methods[method.Name] = &agentMethod{
					method:    method,
					ArgType:   method.Type.In(1),
					ReplyType: method.Type.Out(0),
				}
				return
			} else {
				errString := "MethodAgent注册的函数参数不对(必须有且仅有一个输入和输出参数): " + ma.orgType.String() + "." + name
				fmt.Println(errString)
				panic(errString)
			}
		}
	}
	errString := "MethodAgent注册" + ma.orgType.String() + "找不函数(只有可以被导出的函数才可以注册):" + name
	fmt.Println(errString)
	panic(errString)
}

//RegisterAll 注册所有函数
func (ma *MethodAgent) RegisterAll() {
	ma.Lock()
	defer ma.Unlock()
	for m := 0; m < ma.orgType.NumMethod(); m++ {
		method := ma.orgType.Method(m)

		if method.Type.NumIn() == 2 && method.Type.NumOut() == 1 {
			//fmt.Printf("注册函数: Name:%s ===>%v\n", method.Name, method)
			ma.methods[method.Name] = &agentMethod{
				method:    method,
				ArgType:   method.Type.In(1),
				ReplyType: method.Type.Out(0),
			}
		}

	}
}

//Call 调用函数
func (ma *MethodAgent) Call(agent interface{}, name string, in interface{}) interface{} {
	ma.RLock()
	defer ma.RUnlock()

	if agent != nil && reflect.TypeOf(agent) != ma.orgType {
		return nil
	}

	if m, ok := ma.methods[name]; ok {
		out := m.method.Func.Call([]reflect.Value{
			reflect.ValueOf(agent),
			reflect.ValueOf(in),
		})
		if len(out) == 1 {
			return out[0].Interface()
		}
		return nil

	}
	log.Error("MethodAgent.Call 未知函数: " + ma.orgType.String() + "." + name)
	return nil
}
