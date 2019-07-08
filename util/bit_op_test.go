package util

import "testing"

func IsSliceEqual(a []uint64, b []uint64) bool {
	for _, v := range a {
		found := false
		for _, v2 := range b {
			if v == v2 {
				found = true
				break
			}
		}
		if found == false {
			return false
		}
	}
	for _, v := range b {
		found := false
		for _, v2 := range a {
			if v == v2 {
				found = true
				break
			}
		}
		if found == false {
			return false
		}
	}
	return true
}

func TestUnpackBitSet(t *testing.T) {
	x := uint64(0)
	r := []uint64{3, 4, 6, 0, 17, 6, 63}

	for _, v := range r {
		x = SetBit(x, v)
	}

	x = ClearBit(x, 0)
	x = ClearBit(x, 1)
	x = ClearBit(x, 6)

	for _, v := range r {
		x = SetBit(x, v)
	}
	x = ClearBit(x, 17+63)


	o := UnpackBitSet(x)
	t.Log(r)
	t.Log(o)
	if IsSliceEqual(r, o) == false {
		t.Fail()
	}
}
