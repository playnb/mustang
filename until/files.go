package until

//TODO  名字拼写错误until==>util

import (
	"bufio"
	"github.com/playnb/mustang/log"
	//	"github.com/playnb/protocol/msg"
	"io"
	//	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func substr(s string, pos, length int) string {
	runes := []rune(s)
	l := pos + length
	if l > len(runes) {
		l = len(runes)
	}
	return string(runes[pos:l])
}

func GetParentDirectory(dirctory string) string {
	index := strings.LastIndex(dirctory, "/")
	if index != -1 {
		return substr(dirctory, 0, index)
	}
	return dirctory
}

func GetCurrentDirectory() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Error(err.Error())
	}
	return strings.Replace(dir, "\\", "/", -1)
}

func GetFilelist(path string, f func(string)) {
	filepath.Walk(path, func(fileName string, finfo os.FileInfo, err error) error {
		if finfo == nil {
			return err
		}
		if finfo.IsDir() {
			return nil
		}
		f(fileName)
		return nil
	})
}

func ReadFileLine(fileName string, handler func(string)) error {
	f, err := os.Open(fileName)
	if err != nil {
		return err
	}
	buf := bufio.NewReader(f)
	for {
		line, err := buf.ReadString('\n')
		line = strings.TrimSpace(line)
		handler(line)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
	}
}
