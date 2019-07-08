package util

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"cell/common/mustang/log"
	"strconv"
	"strings"
)

type YamlConf struct {
	data     map[interface{}]interface{}
	baseConf *YamlConf
	baseName []string
}

func (ym *YamlConf) GetConfig(ns ...string) *YamlConf {
	mp := ym.Get(ns...)
	if mp == nil {
		return nil
	}
	if mi, ok := mp.(map[interface{}]interface{}); ok {
		ny := &YamlConf{}
		ny.data = mi
		ny.baseConf = ym
		ny.baseName = ns
		return ny
	}
	return nil
}

func (ym *YamlConf) LoadBytes(data []byte) {
	ym.data = make(map[interface{}]interface{})
	err := yaml.Unmarshal(data, ym.data)
	if err != nil {
		log.Error(err.Error())
		return
	}
	log.Debug("YamlConf:Load:%v", ym.data)
}

func (ym *YamlConf) LoadFile(fileName string) {
	yamlFile, err := ioutil.ReadFile(fileName)
	if err != nil {
		log.Error(err.Error())
		return
	}
	ym.LoadBytes(yamlFile)
}

func (ym *YamlConf) GetString(ns ...string) string {
	mp := ym.Get(ns...)
	if mp == nil {
		return ""
	}
	if str, ok := mp.(string); ok {
		return str
	} else if num,ok:=mp.(int);ok {
		return strconv.FormatInt(int64(num),10)
	} else if b,ok:=mp.(bool);ok {
		return strconv.FormatBool(b)
	}
	return ""
}

func (ym *YamlConf) GetUint64(ns ...string) uint64 {
	mp := ym.Get(ns...)
	if mp == nil {
		return 0
	}
	if str, ok := mp.(string); ok {
		num, err := strconv.ParseUint(str, 10, 64)
		if err != nil {
			return 0
		}
		return uint64(num)
	} else if num, ok := mp.(int); ok {
		return uint64(num)
	}
	return 0
}

func (ym *YamlConf) GetInt(ns ...string) int {
	mp := ym.Get(ns...)
	if mp == nil {
		return 0
	}
	if str, ok := mp.(string); ok {
		num, err := strconv.ParseInt(str, 10, 64)
		if err != nil {
			return 0
		}
		return int(num)
	} else if num, ok := mp.(int); ok {
		return int(num)
	}
	return 0
}

func (ym *YamlConf) GetBool(ns ...string) bool {
	str := ym.GetString(ns...)
	return strings.ToUpper(str) == "TRUE"
}

func (ym *YamlConf) Get(ns ...string) interface{} {
	data := interface{}(nil)
	mp := ym.data
	for _, name := range ns {
		data = nil
		if mp == nil {
			return data
		}
		if m1, ok := mp[name]; ok {
			data = m1
			m2, ok := m1.(map[interface{}]interface{})
			if ok {
				mp = m2
			}
		}
	}
	return data
}
