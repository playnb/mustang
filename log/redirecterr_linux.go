// +build linux

package log

import (
	"fmt"
	"os"
	"syscall"
)

func reDirectErr() {
	logFile, err := os.OpenFile(fmt.Sprintf("%s/cell-%d.err", _panic_dir, os.Getpid()), os.O_WRONLY|os.O_CREATE|os.O_SYNC, 0755)
	if err == nil {
		os.Stderr = logFile
		//syscall.Dup2(int(logFile.Fd()), 1)
		syscall.Dup2(int(logFile.Fd()), int(os.Stderr.Fd()))
	} else {
		panic("打开错误重定向文件失败")
	}
}
