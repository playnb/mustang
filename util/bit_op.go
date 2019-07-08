package util

func SetBit(x uint64, y uint64) uint64 {
	return (x | (uint64(1) << y))
}

func ClearBit(x uint64, y uint64) uint64 {
	return (x & (^(uint64(1) << y)))
}

func IsBitSet(x uint64, y uint64) bool {
	return (((x) >> (y)) & uint64(1)) == uint64(1)
}

func UnpackBitSet(x uint64) []uint64 {
	var ret []uint64
	i := uint64(0)
	for x > 0 {
		if x&1 == 1 {
			ret = append(ret, i)
		}
		i++
		x = x >> 1
	}
	return ret
}
