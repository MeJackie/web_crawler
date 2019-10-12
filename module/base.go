package module

import "net/http"

// 组件类型
type Type string

// 当前组件类型
const (
	//下载器
	TYPE_DOWNLOADER Type = "downloader"
	//分析器
	TYPE_ANALYZER Type = "analyzer"
	//条目处理管道
	TYPE_PIPELINE Type = "pipeline"
)

// 组件内计数类型
type Counts struct {
	// 组件被调用的计数
	CalledCount uint64
	// 组件接收调用计数
	AcceptedCount uint64
	// 组件成功调用计数
	CompletedCount uint64
	// 组件正在处理的调用的数量
	HandlingNumber uint64
}

// 组件摘要结构类型
type SummaryStruct struct {
	ID MID `json:"id"`
	Called uint64 `json:"called"`
	Accepted uint64 `json:"accepted"`
	Completed uint64 `json:"completed"`
	Handling uint64 `json:"handling"`
	Extra interface{} `json:"extra,omitempty"`
}

// Module代表组件的基础接口类型
// 该接口的实现类型必须是并发安全的
type Module interface {
	// 组件ID
	ID() MID
	// 当前组件的网络地址
	Addr() string
	// 获取当前组件的评分
	Score() uint64
	// 设置当前组件的评分
	SetScore(score uint64)
	// 获取评分计算器
	ScoreCalculator() CalculateScore
	// 获取当前组件被调用的计数
	CalledCount() uint64
	// 获取当前组件接收调用计数
	// 组件一般会由于超负荷或参数有误而拒绝调用
	AcceptedCount() uint64
	// 组件成功调用计数
	CompletedCount() uint64
	// 组件正在处理的调用的数量
	HandlingNumber() uint64
	// 获取所有计数
	Counts() Counts
	// 获取组件摘要
	Summary() SummaryStruct
}

// 下载器接口类型
// 该接口的实现类型必须是并发安全的
type Downloader interface {
	Module
	// 根据请求获取内容并返回响应
	Download(req *Request) (*Response, error)
}

// 分析器接口类型
// 该接口的实现类型必须是并发安全的
type Analyzer interface {
	Module
	// 返回当前分析器使用的响应解析函数的列表
	ResqParsers() []ParseResponse
	// 根据规则分析响应并返回请求和条目，
	// 响应需要分别经过若干响应解析函数的处理，然后合并结果
	Analyze(resq *Response) ([]Data, []error)
}

// 用于解析http响应的函数类型
type ParseResponse func(httpResp *http.Response, httpDepth uint32)([]Data, []error)

// 条目处理管道的接口类型
// 该接口的实现类型必须是并发安全的
type Pipeline interface {
	Module
	// 返回当前条目处理管道使用的条目处理函数的列表
	ItemProcessors() []ProcessItem
	// 向条目管道发送条目
	// 条目需要依次经过弱冠条目处理函数的处理
	Send(item Item) []error
	// 方法返回布尔值，该值表示当前条目管道是否是快速失败的
	// 这里的快速失败是指：只要在处理某个条目时在某一个步骤上出错，
	// 那么条目处理管道就会忽略掉后续的所有处理步骤并报告错误
	FailFast() bool
	// 设置是否快速失败
	SetFailFast(failFast bool)
}

// 用于处理条目的函数类型
type ProcessItem func(item Item) (result Item, err error)