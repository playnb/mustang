package until

import (
	"math/rand"
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
