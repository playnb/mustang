package util

import (
	"fmt"
	"testing"
)

type _mockData struct {
	id   int
	name string
}

func buildSlice() []*_mockData {
	var v []*_mockData
	for i := 0; i < 10; i++ {
		v = append(v, &_mockData{
			id:   i,
			name: fmt.Sprintf("Name_%d", i),
		})
	}
	return v
}
func TestDelSilce(t *testing.T) {
	{
		s := buildSlice()

		DelSlice(&s, func(d interface{}) bool {
			return d.(*_mockData).id == 3
		})

		for _, v := range s {
			if v.id == 3 {
				t.Fail()
			}
		}
		if len(s) != 9 {
			t.Fail()
		}
	}

	{
		s := buildSlice()
		s[0].id = 3
		DelSlice(&s, func(d interface{}) bool {
			return d.(*_mockData).id == 3
		})

		for _, v := range s {
			if v.id == 3 {
				t.Fail()
			}
		}
		if len(s) != 8 {
			t.Fail()
		}
	}

	{
		s := buildSlice()
		s[0].id = 3
		s[9].id = 3
		DelSlice(&s, func(d interface{}) bool {
			return d.(*_mockData).id == 3
		})

		for _, v := range s {
			if v.id == 3 {
				t.Fail()
			}
		}
		if len(s) != 7 {
			t.Fail()
		}
	}

	{
		s := buildSlice()
		s[0].id = 3
		s[6].id = 3
		s[7].id = 3
		s[8].id = 3
		s[9].id = 3
		DelSlice(&s, func(d interface{}) bool {
			return d.(*_mockData).id == 3
		})

		for _, v := range s {
			if v.id == 3 {
				t.Fail()
			}
		}
		if len(s) != 4 {
			t.Fail()
		}
	}

	{
		s := buildSlice()
		for _, v := range s {
			v.id = 3
		}
		DelSlice(&s, func(d interface{}) bool {
			return d.(*_mockData).id == 3
		})

		for _, v := range s {
			if v.id == 3 {
				t.Fail()
			}
		}
		if len(s) != 0 {
			t.Fail()
		}
	}
}

func TestSubSlice(t *testing.T) {
	s1 := []int{0, 1, 2, 3, 4, 5}

	//t.Log(s1[1:6])
	t.Log(SubSlice(s1, 0, 2).([]int))
	t.Log(SubSlice(s1, 3, 5).([]int))
	t.Log(SubSlice(s1, 3, 6).([]int))
	t.Log(SubSlice(s1, 3, 9).([]int))
	t.Log(SubSlice(s1, 9, 6).([]int))
	t.Log(SubSlice(s1, 3, 1).([]int))
	t.Log(SubSlice(s1, 6, 6).([]int))
}
