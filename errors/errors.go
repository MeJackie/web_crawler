package errors

import (
	"bytes"
	"fmt"
	"strings"
)

// 代表错误类型
type ErrorType string

// 错误类型常量
const (
	// 下载器错误
	ERROR_TYPE_DOWNLOADER ErrorType = "downloader error"
	// 分析器错误
	ERROR_TYPE_ANALYER ErrorType = "analyzer error"
	// 条目处理管道错误
	ERROR_TYPE_PIPELINE ErrorType = "pipeline error"
	// 调度器错误
	ERROR_TYPE_SCHEDULER ErrorType = "scheduler error"
)

// 爬虫错误接口类型
type CrawlerError interface {
	// 错误类型
	Type() ErrorType
	// 错误提示信息
	Error() string
}

// 爬虫错误的实现类型
type myCrawlerError struct {
	// 错误类型
	errType ErrorType
	// 错误提示信息
	errMsg string
	// 完整错误提示信息
	fullErrMsg string
}

// 创建爬虫错误
func NewCrawlerError(errType ErrorType, errMsg string) CrawlerError {
	return &myCrawlerError{
		errType:    errType,
		errMsg:     strings.TrimSpace(errMsg),
	}
}

func (ce *myCrawlerError) Type() ErrorType {
	return ce.errType
}

func (ce *myCrawlerError) Error() string {
	if ce.fullErrMsg == "" {
		ce.genFullErrMsg()
	}
	return ce.fullErrMsg
}

// 生成错误提示信息，并给相应字段赋值
func (ce *myCrawlerError) genFullErrMsg() {
	var buffer bytes.Buffer
	buffer.WriteString("crawler error : ")
	if ce.errType != "" {
		buffer.WriteString(string(ce.errType))
		buffer.WriteString(" : ")
	}
	buffer.WriteString(ce.errMsg)
	ce.fullErrMsg = fmt.Sprintf("%s", buffer.String())
	return
}