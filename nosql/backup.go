package nosql

import (
	"fmt"
	"github.com/robfig/cron"
	"os"
	"time"
	"io"
)

func CopyFile(src, des string) (w int64, err error) {
	srcFile, err := os.Open(src)
	if err != nil {
		fmt.Println(err)
	}
	defer srcFile.Close()

	desFile, err := os.Create(des)
	if err != nil {
		fmt.Println(err)
	}
	defer desFile.Close()

	return io.Copy(desFile, srcFile)
}

func BackupRedis(path string, files []string) {
	err := os.MkdirAll(path+"/bak", os.ModePerm)
	if err != nil {
		fmt.Println(err)
	}

	c := cron.New()
	spec := "0 0 * * *"
	c.AddFunc(spec, func() {
		dirName := fmt.Sprintf("%s", time.Now().Format("2006-01-02-15_04_05"))
		err := os.MkdirAll(path+"/bak/"+dirName, os.ModePerm)
		if err == nil {
			for _,file := range files{
				CopyFile(path+"/"+file, path+"/bak/"+dirName+"/"+file)
			}
		}
	})
	c.Start()
}
