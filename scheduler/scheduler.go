package scheduler

import (
	"context"
	"net/http"
	"sync"

	"web_crawler/module"
	"web_crawler/helper/log"
	"web_crawler/toolkit/buffer"
	"web_crawler/toolkit/cmap"
)
// 日志记录器
var logger = log.DLogger()

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
	// 爬取的最大深度。首次深度为0
	maxDepth uint32
	// 代表可以接收的URL的主域名的字典 (cmap:并发安全字典)
	acceptedDomainMap cmap.ConcurrentMap
	registrar module.Registrar
	reqBufferPool buffer.Pool
	respBufferPool buffer.Pool
	itemBufferPool buffer.Pool
	errorBufferPool buffer.Pool
	// 已处理的url
	urlMap cmap.ConcurrentMap
	// 上下文，用于感知调度器的停止
	ctx context.Context
	// 取消函数，用于停止调度器
	cancelFunc context.CancelFunc
	status Status
	// 专用于状态的读写锁
	statusLock sync.RWMutex
	summary SchedSummary
}

// 初始化调度器
// 参数requestArgs代表请求相关参数
// 参数dataArgs代表数据相关参数
// 参数moduleArgs代表组件相关参数
func (scheduler *myScheduler) Init(
	requestArgs RequestArgs,
	dataArgs DataArgs,
	moduleArgs ModuleArgs) (err error) {
	// 检查状态
	logger.Info("Check status for intialization...")
	var oldStatus Status
	oldStatus, err =scheduler.checkAndSetStatus(SCHED_STATUS_INITIALIZED)
	defer func() {
		scheduler.statusLock.Lock()
		if err != nil {
			scheduler.status = oldStatus
		} else {
			scheduler.status = SCHED_STATUS_INITIALIZED
		}
		scheduler.statusLock.Unlock()
	}()
	// 检查参数
	logger.Info("check request arguments...")
	if err = requestArgs.Check(); err != nil {
		return err
	}
	logger.Info("Check data arguments...")
	if err = dataArgs.Check(); err != nil {
		return err
	}
	logger.Info("Data arguments are vaild.")
	logger.Info("Check modules arguments...")
	if err = moduleArgs.Check(); err != nil {
		return err
	}
	logger.Info("Modules arguments are vaild.")
	// 初始化内部字段
	logger.Info("Initialize scheduler's fields...")
	if scheduler.registrar == nil {
		scheduler.registrar = module.NewRegistrar()
	} else {
		scheduler.registrar.Clear()
	}
	scheduler.maxDepth = requestArgs.MaxDepth
	logger.Infof("-- Max depth: %d", scheduler.maxDepth)

	scheduler.acceptedDomainMap, _ = cmap.NewConcurrentMap(1, nil)
	for _, domain := range requestArgs.AcceptedDomains {
		scheduler.acceptedDomainMap.Put(domain, struct {}{})
	}
	logger.Infof("-- Accepted primary domains: %v", requestArgs.AcceptedDomains)

	scheduler.urlMap, _ = cmap.NewConcurrentMap(16, nil)
	logger.Infof("-- URL map: length: %d, concurrency: %d",
		scheduler.urlMap.Len(), scheduler.urlMap.Concurrency())

	scheduler.initBufferPool(dataArgs)
	scheduler.resetContext()

}

// 启动调度器并执行爬取流程
// 参数firstHTTPReq 代表首次请求，调度器会以此为起点开始执行爬去流程
func (scheduler *myScheduler) Start(firstHTTPReq *http.Request) (err error) {

}
// 停止调度器的运行
// 所有处理模块执行的流程都会被终止
func (scheduler *myScheduler) Stop() {

}

// 获取调度器状态
func (scheduler *myScheduler) Status() Status {

}

// 获取错误通道
// 调度器以及各个处理模块运行过程中出现的所有错误都会被发送到该通道
// 若结果值为nil, 则说明错误通道不可用或调度器已停止
func (scheduler *myScheduler) ErrorChan() <-chan error {

}

func (scheduler *myScheduler) Idle() bool {

}

func (scheduler *myScheduler) Summary() SchedSummary {

}

// 检查状态，并在条件满足是设置状态
func (scheduler *myScheduler) checkAndSetStatus(
	wantedStatus Status) (oldStatus Status, err error) {
		scheduler.statusLock.Lock()
		defer scheduler.statusLock.Unlock()
		oldStatus = scheduler.status
		err = checkStatus(oldStatus, wantedStatus, nil)
		if err == nil {
			scheduler.status = wantedStatus
		}
		return
}

// 按照给定参数初始化缓冲池
// 如果某个缓冲池可用且未关闭， 就先关闭该缓冲池
func (scheduler *myScheduler) initBufferPool(dataArgs DataArgs) {
	// 初始化请求缓冲池
	if scheduler.reqBufferPool != nil && !scheduler.reqBufferPool.Closed() {
		scheduler.reqBufferPool.Closed()
	}
	scheduler.reqBufferPool, _ = buffer.NewPool(dataArgs.ReqBufferCap, dataArgs.ReqMaxBufferNumber)
	logger.Infof("-- Request buffer pool: bufferCap: %d, maxBufferNumber: %d",
		dataArgs.ReqBufferCap, dataArgs.ReqMaxBufferNumber)
	// 初始化响应缓冲池。
	if scheduler.respBufferPool != nil && !scheduler.respBufferPool.Closed() {
		scheduler.respBufferPool.Close()
	}
	scheduler.respBufferPool, _ = buffer.NewPool(
		dataArgs.RespBufferCap, dataArgs.RespMaxBufferNumber)
	logger.Infof("-- Response buffer pool: bufferCap: %d, maxBufferNumber: %d",
		scheduler.respBufferPool.BufferCap(), scheduler.respBufferPool.MaxBufferNumber())
	// 初始化条目缓冲池。
	if scheduler.itemBufferPool != nil && !scheduler.itemBufferPool.Closed() {
		scheduler.itemBufferPool.Close()
	}
	scheduler.itemBufferPool, _ = buffer.NewPool(
		dataArgs.ItemBufferCap, dataArgs.ItemMaxBufferNumber)
	logger.Infof("-- Item buffer pool: bufferCap: %d, maxBufferNumber: %d",
		scheduler.itemBufferPool.BufferCap(), scheduler.itemBufferPool.MaxBufferNumber())
	// 初始化错误缓冲池。
	if scheduler.errorBufferPool != nil && !scheduler.errorBufferPool.Closed() {
		scheduler.errorBufferPool.Close()
	}
	scheduler.errorBufferPool, _ = buffer.NewPool(
		dataArgs.ErrorBufferCap, dataArgs.ErrorMaxBufferNumber)
	logger.Infof("-- Error buffer pool: bufferCap: %d, maxBufferNumber: %d",
		scheduler.errorBufferPool.BufferCap(), scheduler.errorBufferPool.MaxBufferNumber())

}

// 重置调度器上下文
func (scheduler *myScheduler) resetContext() {
	scheduler.ctx, scheduler.cancelFunc = context.WithCancel(context.Background())
}

