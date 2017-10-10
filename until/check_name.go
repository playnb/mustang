package until

import (
	"bufio"
	"github.com/playnb/mustang/log"
	"io"
	"os"
	"strings"

	"github.com/huichen/sego"
	"github.com/smtc/godat"
)

type CheckBadWord struct {
	segmenter sego.Segmenter
	goDic     godat.GoDat
}

//Init 初始化
func (cn *CheckBadWord) Init(dictionary string, badword string) bool {
	if !cn.initSegment(dictionary) {
		return false
	}
	if !cn.initFilter(badword) {
		return false
	}
	return true
}

//Check 屏蔽字处理(返回true包含屏蔽字，false是正常文字)
func (cn *CheckBadWord) Check(checkStr string) bool {
	text := []byte(checkStr)
	segments := cn.segmenter.Segment(text)

	str := sego.SegmentsToString(segments, true)

	nameFrag := strings.Split(str, " ")
	for _, n := range nameFrag {
		name := strings.Split(n, "/")
		if len(name) == 0 {
			continue
		}

		if cn.goDic.Match(name[0], 0) == true {
			//log.Debug("%s 中包含屏蔽字 %s", check_name, name[0])
			return true
		} else {
			//log.Debug("%s 不是屏蔽字", name[0])
		}
	}

	return false
}

func (cn *CheckBadWord) initSegment(dictionary string) bool {
	return cn.segmenter.LoadDictionary(dictionary)
}

// 生成屏蔽字字典
func (cn *CheckBadWord) initFilter(badword string) bool {
	cn.goDic = godat.GoDat{}

	f, err := os.Open(badword)
	defer f.Close()

	if err != nil {
		log.Error("open BadWord file failed!\n")
		return false
	} else {
		rd := bufio.NewReader(f)

		for line, err := rd.ReadString(byte('\n')); err == nil || err != io.EOF; line, err = rd.ReadString(byte('\n')) {
			str_reduce_n := strings.Replace(line, "\n", "", -1)
			str_reduce_r := strings.Replace(str_reduce_n, "\r", "", -1)
			cn.goDic.Add(str_reduce_r)
		}

		cn.goDic.Initialize(true)
		cn.goDic.BuildWithoutConflict()

		//log.Debug("GD列表 %s", self.GD.GetPats())
	}

	return true
}
