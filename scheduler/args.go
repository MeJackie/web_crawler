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

func (args *DataArgs) Check() error  {
	if args.ReqBufferCap == 0 {
		return genError("zero request buffer capacity")
	}
	if args.ReqMaxBufferNumber == 0 {
		return genError("zero max request buffer number")
	}
	if args.RespBufferCap == 0 {
		return genError("zero response buffer capacity")
	}
	if args.RespMaxBufferNumber == 0 {
		return genError("zero max response buffer number")
	}
	if args.ItemBufferCap == 0 {
		return genError("zero item buffer capacity")
	}
	if args.ItemMaxBufferNumber == 0 {
		return genError("zero max item buffer number")
	}
	if args.ErrorBufferCap == 0 {
		return genError("zero error buffer capacity")
	}
	if args.ErrorMaxBufferNumber == 0 {
		return genError("zero max error buffer number")
	}
	return nil
}

// 提供组件实例列表
type ModuleArgs struct {
	Downloaders []module.Downloader
	Analyzers []module.Analyzer
	Pipelines []module.Pipeline
}

// ModuleArgsSummary 代表组件相关的参数容器的摘要类型。
type ModuleArgsSummary struct {
	DownloaderListSize int `json:"downloader_list_size"`
	AnalyzerListSize   int `json:"analyzer_list_size"`
	PipelineListSize   int `json:"pipeline_list_size"`
}

func (args *ModuleArgs) Check() error  {
	if len(args.Downloaders) == 0 {
		return genError("empty downloader list")
	}
	if len(args.Analyzers) == 0 {
		return genError("empty analyzer list")
	}
	if len(args.Pipelines) == 0 {
		return genError("empty pipeline list")
	}
	return nil
}

func (args *ModuleArgs) Summary() ModuleArgsSummary {
	return ModuleArgsSummary{
		DownloaderListSize: len(args.Downloaders),
		AnalyzerListSize:   len(args.Analyzers),
		PipelineListSize:   len(args.Pipelines),
	}
}