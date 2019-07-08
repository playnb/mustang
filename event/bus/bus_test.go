package bus

import "testing"

var n = 1

func add(x int) {
	n = n + x
}

func mul(x int) {
	n = n * x
}

func TestDefaultDispatcher(t *testing.T) {
	d := New()
	d.Subscribe("calc", add)
	d.SubscribeAsync("calc_a", add, false)
	if d.HasCallback("calc") == false {
		t.Fail()
	}
	if d.HasCallback("calc_a") == false {
		t.Fail()
	}
	if d.HasCallback("xxx") == true {
		t.Fail()
	}

	if d.Unsubscribe("xxx", add) == nil {
		t.Fail()
	}
	if d.Unsubscribe("calc_a", mul) == nil {
		t.Fail()
	}
	if d.Unsubscribe("calc", mul) == nil {
		t.Fail()
	}
	if d.Unsubscribe("calc", add) != nil {
		t.Fail()
	}
	if d.Unsubscribe("calc_a", add) != nil {
		t.Fail()
	}
	if d.Unsubscribe("calc_a", add) == nil {
		t.Fail()
	}
	if d.Unsubscribe("calc", add) == nil {
		t.Fail()
	}
}

func TestDispatcher_Publish1(t *testing.T) {
	n = 1
	d := New()
	d.Subscribe("calc", add)
	d.Subscribe("calc", add)
	d.Publish("calc", 10)
	if n != 21 {
		t.Fail()
	}
}

func TestDispatcher_Publish2(t *testing.T) {
	n = 1
	d := New()
	d.Subscribe("calc", add)
	d.Subscribe("calc", mul)
	d.Publish("calc", 10)
	if n != 110 {
		t.Fail()
	}
}

func TestDispatcher_Publish3(t *testing.T) {
	n = 0
	d := New()
	d.SubscribeAsync("calc", add, true) //10
	d.SubscribeAsync("calc", add, true) //20
	d.SubscribeAsync("calc", mul, true) //200
	d.SubscribeAsync("calc", add, true) //210

	d.Publish("calc", 10)

	d.WaitAsync()
	if n != 210 {
		t.Fail()
	}
}


func _Publish4(t *testing.T) {
	n = 0
	d := New()
	d.SubscribeAsync("calc", add, false) //10
	d.SubscribeAsync("calc", add, false) //20
	d.SubscribeAsync("calc", mul, false) //200
	d.SubscribeAsync("calc", add, false) //210
	d.SubscribeAsync("calc", mul, false) //200
	d.SubscribeAsync("calc", add, false) //210
	d.SubscribeAsync("calc", add, false) //210
	d.SubscribeAsync("calc", add, false) //210

	d.Publish("calc", 10)

	d.WaitAsync()
	t.Logf("TestDispatcher_Publish4 %d", n)
}

func TestDispatcher_Publish4(t *testing.T) {
	_Publish4(t)
	_Publish4(t)
	_Publish4(t)
	_Publish4(t)
	_Publish4(t)
}


func TestDispatcher_Publish5(t *testing.T) {
	n = 0
	d := New()
	d.Subscribe("calc", add)
	d.Unsubscribe("calc", mul)
	d.Publish("calc", 10)
	if n != 10 {
		t.Fail()
	}
}

func TestDispatcher_Publish6(t *testing.T) {
	n = 0
	d := New()
	d.Subscribe("calc", add)
	d.Unsubscribe("calc", add)
	d.Publish("calc", 10)
	if n != 0 {
		t.Fail()
	}
}