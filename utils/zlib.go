package utils

import (
	"bytes"
	"compress/zlib"
	"io"
)

var dict = []byte("{Name wx http HTTP jpg www Token token jpg OpenID}")

//进行zlib压缩
func DoZlibCompress(src []byte) []byte {
	var in bytes.Buffer
	w, _ := zlib.NewWriterLevelDict(&in, zlib.BestCompression, dict)
	w.Write(src)
	w.Close()
	//	fmt.Printf("%d -> %d\n", len(src), len(in.Bytes()))
	return in.Bytes()
}

//进行zlib解压缩
func DoZlibUnCompress(compressSrc []byte) ([]byte, error) {
	b := bytes.NewReader(compressSrc)
	var out bytes.Buffer
	r, err := zlib.NewReaderDict(b, dict)
	if err != nil {
		return nil, err
	}
	io.Copy(&out, r)
	return out.Bytes(), nil
}
