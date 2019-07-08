package util
/*
import (
	"common/protocol/msg"
	"strconv"
	"strings"
)

func NewGmCommandProp(cmd *msg.MSG_GMCommand, toLower bool) *GmCommandProp {
	gm := &GmCommandProp{}
	gm.Init(cmd, toLower)
	return gm
}

type GmCommandProp struct {
	prop    map[string]string
	command string
}

func (gm *GmCommandProp) GetCommand() string {
	return gm.command
}

func (gm *GmCommandProp) Init(cmd *msg.MSG_GMCommand, toLower bool) {
	gm.prop = make(map[string]string)
	for _, v := range cmd.GetParams() {
		if toLower {
			gm.prop[strings.ToLower(v.GetKey())] = strings.ToLower(v.GetParam())
		}else{
			gm.prop[v.GetKey()] = v.GetParam()
		}
	}

	gm.command = cmd.GetCommand()
}

func (gm *GmCommandProp) GetUint64(key string) uint64 {
	if v, ok := gm.prop[strings.ToLower(key)]; ok == true {
		n, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			return 0
		}
		return n
	}
	return 0
}

func (gm *GmCommandProp) GetString(key string) string {
	if v, ok := gm.prop[strings.ToLower(key)]; ok == true {
		return v
	}
	return ""
}

func (gm *GmCommandProp) Foreach(foo func(k, v string)) {
	for k, v := range gm.prop {
		foo(k, v)
	}
}
*/