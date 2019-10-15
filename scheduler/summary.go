package scheduler

import (
	"encoding/json"
	"web_crawler/module"
)

// 调度器摘要的接口类型
type SchedSummary interface {
	// 获取摘要信息的结构化形式
	Struct() SummaryStruct
	// 获取摘要信息的字符串形式
	String() string
}

// 表示调度器摘要的结构
type SummaryStruct struct {
	RequestArgs  RequestArgs `json:"dequest_args"`
    DataArgs  DataArgs  `json:"data_args"`
    ModuleArgs  ModuleArgsSummary  `json:"module_args"`
    Status  string  `json:"status"`
    Downloaders  []module.SummaryStruct  `json:"downloaders"`
	Analyzers  []module.SummaryStruct  `json:"analyzers"`
    Pipelines  []module.SummaryStruct  `json:"pipelines"`
    ReqBufferPool BufferPoolSummaryStruct `json:"request_buffer_pool"`
    RespBufferPool BufferPoolSummaryStruct `json:":response_buffer_pool"`
    ItemBufferPool BufferPoolSummaryStruct `json:"item_buffer_pool"`
    ErrorBufferPool BufferPoolSummaryStruct `json:"error_buffer_pool"`
    NumURL  uint64  `json:"url_number"`
}

// BufferPoolSummaryStruct 代表缓冲池的摘要类型。
type BufferPoolSummaryStruct struct {
	BufferCap       uint32 `json:"buffer_cap"`
	MaxBufferNumber uint32 `json:"max_buffer_number"`
	BufferNumber    uint32 `json:"buffer_number"`
	Total           uint64 `json:"total"`
}

// mySchedSummary 代表调度器摘要的实现类型。
type mySchedSummary struct {
	// requestArgs 代表请求相关的参数。
	requestArgs RequestArgs
	// dataArgs 代表数据相关参数的容器实例。
	dataArgs DataArgs
	// moduleArgs 代表组件相关参数的容器实例。
	moduleArgs ModuleArgs
	// maxDepth 代表爬取的最大深度。
	maxDepth uint32
	// sched 代表调度器实例。
	sched *myScheduler
}

func NewSchedSummary(
	requestArgs RequestArgs,
	dataArgs DataArgs,
	moduleArgs ModuleArgs,
	sched *myScheduler) SchedSummary {
	if sched == nil {
		return nil
	}
	return &mySchedSummary{
		requestArgs: requestArgs,
		dataArgs:    dataArgs,
		moduleArgs:  moduleArgs,
		sched:       sched,
	}
}


// 获取摘要信息的结构化形式
func (ss *mySchedSummary) Struct() SummaryStruct {
	registrar := ss.sched.registrar
	return SummaryStruct{
		RequestArgs:     ss.requestArgs,
		DataArgs:        ss.dataArgs,
		ModuleArgs:      ss.moduleArgs.Summary(),
		Status:          GetStatusDescription(ss.sched.status),
		Downloaders:     get,
		Analyzers:       nil,
		Pipelines:       nil,
		ReqBufferPool:   BufferPoolSummaryStruct{},
		RespBufferPool:  BufferPoolSummaryStruct{},
		ItemBufferPool:  BufferPoolSummaryStruct{},
		ErrorBufferPool: BufferPoolSummaryStruct{},
		NumURL:          0,
	}
}
// 获取摘要信息的字符串形式
func (ss *mySchedSummary) String() string {
	b, err := json.MarshalIndent(ss.Struct(), "", "    ")
	if err != nil {
		logger.Errorf("An error", err)
	}
}