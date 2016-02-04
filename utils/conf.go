package utils

import (
	"encoding/json"
	"github.com/playnb/mustang/log"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

func LoadJsonFile(fileName string, conf interface{}) bool {
	//confName := ""
	if !strings.HasSuffix(fileName, ".json") {
		fileName = fileName + ".json"
		//confName = fileName
	} else {
		//confName = fileName[:len(fileName)-len(".json")]
	}
	r, err := os.Open(fileName)
	if err != nil {
		log.Error(err.Error())
		return false
	}
	buf := make([]byte, 1024*1024)
	n, err := r.Read(buf)
	if err != nil {
		log.Error(err.Error())
		return false
	}
	str := string(buf[:n])
	//str = strings.Replace(str, "//", "######", -1)
	//str = strings.Replace(str, "/", "\\/", -1)
	//str = strings.Replace(str, "######", "//", -1)
	err = json.Unmarshal([]byte(str), conf)
	//decoder := json.NewDecoder(r)
	//err = decoder.Decode(conf)
	if err != nil {
		log.Error(err.Error())
		return false
	}
	return true
}

func LoadJsonURL(url string, conf interface{}) bool {

	resp, err := http.Get(url)
	if err != nil {
		panic(err.Error())
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err.Error())
	}

	//str = strings.Replace(str, "//", "######", -1)
	//str = strings.Replace(str, "/", "\\/", -1)
	//str = strings.Replace(str, "######", "//", -1)
	err = json.Unmarshal(body, conf)
	//decoder := json.NewDecoder(r)
	//err = decoder.Decode(conf)
	if err != nil {
		log.Error(err.Error())
		return false
	}
	return true
}
