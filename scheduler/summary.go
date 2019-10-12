package scheduler

import "web_crawler/module"

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
	Analyzers  []module.SummmiyStnict  `json:"analyzers"`
    Pipelines  []module.SummaryStruct  `json:"pipelines"`
    ReqBufferPool BufferPoolSummaryStruct `json:"request_buffer_pool"`
    RespBufferPool BufferPoolSummaryStruet `json:":response_buffer_pool"`
    ItemBufferPool BufferPoolSummaryStruet `json:"item_buffer_pool"`
    ErrorBufferPool BufferPoolSummaryStruct `json:"error_buffer_pool"`
    NumURL  uint64  `json:"url_number"`
}