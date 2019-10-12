package scheduler

import (
	"net/http"
)

// 调度器接口类型
type Scheduler interface {
	// 初始化调度器
	// 参数requestArgs代表请求相关参数
	// 参数dataArgs代表数据相关参数
	// 参数moduleArgs代表组件相关参数
	Init(requestArgs RequestArgs, dataArgs DataArgs, moduleArgs ModuleArgs) (err error)
	// 启动调度器并执行爬取流程
	// 参数firstHTTPReq 代表首次请求，调度器会以此为起点开始执行爬去流程
	Start(firstHTTPReq *http.Request) (err error)
	// 停止调度器的运行
	// 所有处理模块执行的流程都会被终止
	Stop()
	// 获取调度器状态
	Status() Status
	// 获取错误通道
	// 调度器以及各个处理模块运行过程中出现的所有错误都会被发送到该通道
	// 若结果值为nil, 则说明错误通道不可用或调度器已停止
	ErrorChan() <-chan error
	Idle() bool
	Summary() SchedSummary
}

// 调度器实现类型
type myScheduler struct {

}