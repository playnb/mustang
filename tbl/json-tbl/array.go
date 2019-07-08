package json_tbl

import (
	"strconv"
	"strings"
)

type ArrayUint64 struct {
	Data []uint64
}

func (j *ArrayUint64) UnmarshalJSON(data []byte) error {
	str := string(data)
	str = strings.Trim(str, `"`)
	ss := strings.Split(str, ",")
	for _, s := range ss {
		n, _ := strconv.ParseUint(s, 10, 64)
		j.Data = append(j.Data, n)
	}
	return nil
}
