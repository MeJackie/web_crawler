package reader

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
)

type MultipleReader interface {
	// Reader用于获取一个可关闭读取器的实例。
	// 后者会持有该多重读取器中的数据
	Reader() io.ReadCloser
}

// 多重读取器的实现类型
type myMultipleReader struct {
	data []byte
}

func NewMultipleReader(reader io.Reader) (MultipleReader,error) {
	var data []byte
	var err error
	if reader != nil {
		data, err = ioutil.ReadAll(reader)
		if err != nil {
			return nil, fmt.Errorf("multipie reader: could't create a new one: %s", err)
		} else {
			data = []byte{}
		}

		return &myMultipleReader{data:data},nil
	}
}

func (mr *myMultipleReader) Reader() io.ReadCloser {
	return ioutil.NopCloser(bytes.NewReader(mr.data))
}