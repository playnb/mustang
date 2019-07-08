package util

import (
	"os"
	"testing"
)

type TestMethodAgentInt interface {
}

type TestMethodAgentType2 struct {
	name    string
	testing *testing.T
	n       int
}

func (t *TestMethodAgentType2) Foo1(n int) int {
	t.testing.Logf("Foo1: %s(%d)\n", t.name, n)
	return n + t.n
}

type TestMethodAgentType struct {
	name    string
	testing *testing.T
	n       int
}

func (t *TestMethodAgentType) Foo1(n int) int {
	t.testing.Logf("Foo1: %s(%d)\n", t.name, n)
	return n + t.n
}

func (t *TestMethodAgentType) Foo2() {
	t.testing.Logf("Foo2:" + t.name)
}

func (t *TestMethodAgentType) Foo3(n int) {
	t.testing.Logf("Foo3: %s(%d)\n", t.name, n)
	return
}

func TestMethodAgent(t *testing.T) {
	m := &MethodAgent{}
	m.InitMethodAgent(&TestMethodAgentType{})
	m.RegisterAll()

	var ins TestMethodAgentInt
	ins = &TestMethodAgentType{name: "CALL_TYPE", testing: t, n: 1}
	for i := 0; i < 100; i++ {
		t.Run("有函数", func(t *testing.T) {
			n := m.Call(ins, "Foo1", 99)
			if n == nil {
				t.Fatal("函数调用结果为nil")
			} else if n.(int) != 99+1 {
				t.Fatal("函数调用结果不正确")
			}
		})

		t.Run("没函数", func(t *testing.T) {
			os.Stdout = nil
			n := m.Call(ins, "FooM", 99)
			if n != nil {
				t.Fail()
				return
			}
		})
	}
}
