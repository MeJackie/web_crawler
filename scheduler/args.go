package scheduler

import (
	"web_crawler/module"
)

type Args interface {
	// 检查参数有效性
	Check() error
}

type RequestArgs struct {
	// 可接收的 URL 的主域名列表
	AcceptedDomains []string `json:"accepted_primary_domains"`
	MaxDepth uint32 `json:"max_depth"`
}

func (args *RequestArgs) Check() error  {
	if args.AcceptedDomains == nil {
		return genError("nil accepted primary domain list")
	}
}


type DataArgs struct {
	//请求缓冲器的容量
	ReqBufferCap uint32 `json:":req_buffer_cap"`
    //请求缓冲器的最大数量
	ReqMaxBufferNumber uint32 `json:"req_max_buffer_number"`
    //响应缓冲器的容量
	RespBufferCap uint32 `json:":resp_buffer_cap"`
    //响应缓冲器的最大数量
	RespMaxBufferNumber uint32 `json:"resp_max_buffer_number"`
    //条目缓冲器的容量
	ItemBufferCap uint32 `json:"item_buffer_cap"`
    //条目缓冲器的最大数量
	ItemMaxBufferNumber uint32 `json:"item_max_buffer_number"`
    //错误缓冲器的容量
	ErrorBufferCap uint32 `json:"error_buffer_cap"`
    //错误缓冲器的最大数量
	ErrorMaxBufferNumber uint32 `json:"error_max_buffer_number"`
}

// 提供组件实例列表
type ModuleArgs struct {
	Downloaders []module.Downloader
	Analyzer []module.Analyzer
	Pipeline []module.Pipeline
}