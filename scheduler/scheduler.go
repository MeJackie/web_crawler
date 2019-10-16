package scheduler

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"errors"

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
	Stop() (err error)
	// 获取调度器状态
	Status() Status
	// 获取错误通道
	// 调度器以及各个处理模块运行过程中出现的所有错误都会被发送到该通道
	// 若结果值为nil, 则说明错误通道不可用或调度器已停止
	ErrorChan() <-chan error
	//用于判断所有处理模块是否都处于空闲状态。
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
	scheduler.summary =
		newSchedSummary(requestArgs,dataArgs,moduleArgs,scheduler)
	// 注册组件
	logger.Info("Register modules...")
	if err := scheduler.registerModules(moduleArgs); err != nil {
		return err
	}
	logger.Info("Scheduler has been initialized.")
	return nil
}

// 启动调度器并执行爬取流程
// 参数firstHTTPReq 代表首次请求，调度器会以此为起点开始执行爬去流程
func (scheduler *myScheduler) Start(firstHTTPReq *http.Request) (err error) {
	defer func() {
		if p := recover(); p != nil {
			errMsg := fmt.Sprintf("Fatal scheduler error: %s", p)
			logger.Fatal(errMsg)
			err = genError(errMsg)
		}
	}()
	logger.Info("Start scheduler...")
	// 检查状态。
	logger.Info("Check status for start...")
	var oldStatus Status
	oldStatus, err =
		scheduler.checkAndSetStatus(SCHED_STATUS_STARTING)
	defer func() {
		scheduler.statusLock.Lock()
		if err != nil {
			scheduler.status = oldStatus
		} else {
			scheduler.status = SCHED_STATUS_STARTED
		}
		scheduler.statusLock.Unlock()
	}()
	if err != nil {
		return
	}
	// 检查参数。
	logger.Info("Check first HTTP request...")
	if firstHTTPReq == nil {
		err = genParameterError("nil first HTTP request")
		return
	}
	logger.Info("The first HTTP request is valid.")
	// 获得首次请求的主域名，并将其添加到可接受的主域名的字典。
	logger.Info("Get the primary domain...")
	logger.Infof("-- Host: %s", firstHTTPReq.Host)
	var primaryDomain string
	primaryDomain, err = getPrimaryDomain(firstHTTPReq.Host)
	if err != nil {
		return
	}
	logger.Infof("-- Primary domain: %s", primaryDomain)
	scheduler.acceptedDomainMap.Put(primaryDomain, struct{}{})
	// 开始调度数据和组件。
	if err = scheduler.checkBufferPoolForStart(); err != nil {
		return
	}
	scheduler.download()
	scheduler.analyze()
	scheduler.pick()
	logger.Info("Scheduler has been started.")
	// 放入第一个请求。
	firstReq := module.NewRequest(firstHTTPReq, 0)
	scheduler.sendReq(firstReq)
	return nil
}
// 停止调度器的运行
// 所有处理模块执行的流程都会被终止
func (scheduler *myScheduler) Stop() (err error) {
	logger.Info("Stop scheduler...")
	// 检查状态。
	logger.Info("Check status for stop...")
	var oldStatus Status
	oldStatus, err =
		scheduler.checkAndSetStatus(SCHED_STATUS_STOPPING)
	defer func() {
		scheduler.statusLock.Lock()
		if err != nil {
			scheduler.status = oldStatus
		} else {
			scheduler.status = SCHED_STATUS_STOPPED
		}
		scheduler.statusLock.Unlock()
	}()
	if err != nil {
		return
	}
	scheduler.cancelFunc()
	scheduler.reqBufferPool.Close()
	scheduler.respBufferPool.Close()
	scheduler.itemBufferPool.Close()
	scheduler.errorBufferPool.Close()
	logger.Info("Scheduler has been stopped.")
	return nil
}

// 获取调度器状态
func (scheduler *myScheduler) Status() Status {
	var status Status
	scheduler.statusLock.RLock()
	status = scheduler.status
	scheduler.statusLock.RUnlock()
	return status
}

// 获取错误通道
// 调度器以及各个处理模块运行过程中出现的所有错误都会被发送到该通道
// 若结果值为nil, 则说明错误通道不可用或调度器已停止
func (scheduler *myScheduler) ErrorChan() <-chan error {
	errBuffer := scheduler.errorBufferPool
	errCh := make(chan error, errBuffer.BufferCap())
	go func(errBuffer buffer.Pool, errCh chan error) {
		for {
			if scheduler.canceled() {
				close(errCh)
				break
			}
			datum, err := errBuffer.Get()
			if err != nil {
				logger.Warnln("The error buffer pool was closed. Break error reception.")
				close(errCh)
				break
			}
			err, ok := datum.(error)
			if !ok {
				errMsg := fmt.Sprintf("incorrect error type: %T", datum)
				sendError(errors.New(errMsg), "", scheduler.errorBufferPool)
				continue
			}
			if scheduler.canceled() {
				close(errCh)
				break
			}
			errCh <- err
		}
	}(errBuffer, errCh)
	return errCh
}

func (scheduler *myScheduler) Idle() bool {
	moduleMap := scheduler.registrar.GetAll()
	for _, module := range moduleMap {
		if module.HandlingNumber() > 0 {
			return false
		}
	}
	if scheduler.reqBufferPool.Total() > 0 ||
		scheduler.respBufferPool.Total() > 0 ||
		scheduler.itemBufferPool.Total() > 0 {
		return false
	}
	return true
}

func (scheduler *myScheduler) Summary() SchedSummary {
	return scheduler.summary
}

// download 会从请求缓冲池取出请求并下载，
// 然后把得到的响应放入响应缓冲池。
func (scheduler *myScheduler) download() {
	go func() {
		for {
			if scheduler.canceled() {
				break
			}
			datum, err := scheduler.reqBufferPool.Get()
			if err != nil {
				logger.Warnln("The request buffer pool was closed. Break request reception.")
				break
			}
			req, ok := datum.(*module.Request)
			if !ok {
				errMsg := fmt.Sprintf("incorrect request type: %T", datum)
				sendError(errors.New(errMsg), "", scheduler.errorBufferPool)
			}
			scheduler.downloadOne(req)
		}
	}()
}

// downloadOne 会根据给定的请求执行下载并把响应放入响应缓冲池。
func (scheduler *myScheduler) downloadOne(req *module.Request) {
	if req == nil {
		return
	}
	if scheduler.canceled() {
		return
	}
	m, err := scheduler.registrar.Get(module.TYPE_DOWNLOADER)
	if err != nil || m == nil {
		errMsg := fmt.Sprintf("couldn't get a downloader: %s", err)
		sendError(errors.New(errMsg), "", scheduler.errorBufferPool)
		scheduler.sendReq(req)
		return
	}
	downloader, ok := m.(module.Downloader)
	if !ok {
		errMsg := fmt.Sprintf("incorrect downloader type: %T (MID: %s)",
			m, m.ID())
		sendError(errors.New(errMsg), m.ID(), scheduler.errorBufferPool)
		scheduler.sendReq(req)
		return
	}
	resp, err := downloader.Download(req)
	if resp != nil {
		sendResp(resp, scheduler.respBufferPool)
	}
	if err != nil {
		sendError(err, m.ID(), scheduler.errorBufferPool)
	}
}

// analyze 会从响应缓冲池取出响应并解析，
// 然后把得到的条目或请求放入相应的缓冲池。
func (scheduler *myScheduler) analyze() {
	go func() {
		for {
			if scheduler.canceled() {
				break
			}
			datum, err := scheduler.respBufferPool.Get()
			if err != nil {
				logger.Warnln("The response buffer pool was closed. Break response reception.")
				break
			}
			resp, ok := datum.(*module.Response)
			if !ok {
				errMsg := fmt.Sprintf("incorrect response type: %T", datum)
				sendError(errors.New(errMsg), "", scheduler.errorBufferPool)
			}
			scheduler.analyzeOne(resp)
		}
	}()
}

// analyzeOne 会根据给定的响应执行解析并把结果放入相应的缓冲池。
func (scheduler *myScheduler) analyzeOne(resp *module.Response) {
	if resp == nil {
		return
	}
	if scheduler.canceled() {
		return
	}
	m, err := scheduler.registrar.Get(module.TYPE_ANALYZER)
	if err != nil || m == nil {
		errMsg := fmt.Sprintf("couldn't get an analyzer: %s", err)
		sendError(errors.New(errMsg), "", scheduler.errorBufferPool)
		sendResp(resp, scheduler.respBufferPool)
		return
	}
	analyzer, ok := m.(module.Analyzer)
	if !ok {
		errMsg := fmt.Sprintf("incorrect analyzer type: %T (MID: %s)",
			m, m.ID())
		sendError(errors.New(errMsg), m.ID(), scheduler.errorBufferPool)
		sendResp(resp, scheduler.respBufferPool)
		return
	}
	dataList, errs := analyzer.Analyze(resp)
	if dataList != nil {
		for _, data := range dataList {
			if data == nil {
				continue
			}
			switch d := data.(type) {
			case *module.Request:
				scheduler.sendReq(d)
			case module.Item:
				sendItem(d, scheduler.itemBufferPool)
			default:
				errMsg := fmt.Sprintf("Unsupported data type %T! (data: %#v)", d, d)
				sendError(errors.New(errMsg), m.ID(), scheduler.errorBufferPool)
			}
		}
	}
	if errs != nil {
		for _, err := range errs {
			sendError(err, m.ID(), scheduler.errorBufferPool)
		}
	}
}

// pick 会从条目缓冲池取出条目并处理。
func (scheduler *myScheduler) pick() {
	go func() {
		for {
			if scheduler.canceled() {
				break
			}
			datum, err := scheduler.itemBufferPool.Get()
			if err != nil {
				logger.Warnln("The item buffer pool was closed. Break item reception.")
				break
			}
			item, ok := datum.(module.Item)
			if !ok {
				errMsg := fmt.Sprintf("incorrect item type: %T", datum)
				sendError(errors.New(errMsg), "", scheduler.errorBufferPool)
			}
			scheduler.pickOne(item)
		}
	}()
}

// pickOne 会处理给定的条目。
func (scheduler *myScheduler) pickOne(item module.Item) {
	if scheduler.canceled() {
		return
	}
	m, err := scheduler.registrar.Get(module.TYPE_PIPELINE)
	if err != nil || m == nil {
		errMsg := fmt.Sprintf("couldn't get a pipeline: %s", err)
		sendError(errors.New(errMsg), "", scheduler.errorBufferPool)
		sendItem(item, scheduler.itemBufferPool)
		return
	}
	pipeline, ok := m.(module.Pipeline)
	if !ok {
		errMsg := fmt.Sprintf("incorrect pipeline type: %T (MID: %s)",
			m, m.ID())
		sendError(errors.New(errMsg), m.ID(), scheduler.errorBufferPool)
		sendItem(item, scheduler.itemBufferPool)
		return
	}
	errs := pipeline.Send(item)
	if errs != nil {
		for _, err := range errs {
			sendError(err, m.ID(), scheduler.errorBufferPool)
		}
	}
}

// sendReq 会向请求缓冲池发送请求。
// 不符合要求的请求会被过滤掉。
func (scheduler *myScheduler) sendReq(req *module.Request) bool {
	if req == nil {
		return false
	}
	if scheduler.canceled() {
		return false
	}
	httpReq := req.HTTPReq()
	if httpReq == nil {
		logger.Warnln("Ignore the request! Its HTTP request is invalid!")
		return false
	}
	reqURL := httpReq.URL
	if reqURL == nil {
		logger.Warnln("Ignore the request! Its URL is invalid!")
		return false
	}
	scheme := strings.ToLower(reqURL.Scheme)
	if scheme != "http" && scheme != "https" {
		logger.Warnf("Ignore the request! Its URL scheme is %q, but should be %q or %q. (URL: %s)\n",
			scheme, "http", "https", reqURL)
		return false
	}
	if v := scheduler.urlMap.Get(reqURL.String()); v != nil {
		logger.Warnf("Ignore the request! Its URL is repeated. (URL: %s)\n", reqURL)
		return false
	}
	pd, _ := getPrimaryDomain(httpReq.Host)
	if scheduler.acceptedDomainMap.Get(pd) == nil {
		if pd == "bing.net" {
			panic(httpReq.URL)
		}
		logger.Warnf("Ignore the request! Its host %q is not in accepted primary domain map. (URL: %s)\n",
			httpReq.Host, reqURL)
		return false
	}
	if req.Depth() > scheduler.maxDepth {
		logger.Warnf("Ignore the request! Its depth %d is greater than %d. (URL: %s)\n",
			req.Depth(), scheduler.maxDepth, reqURL)
		return false
	}
	go func(req *module.Request) {
		if err := scheduler.reqBufferPool.Put(req); err != nil {
			logger.Warnln("The request buffer pool was closed. Ignore request sending.")
		}
	}(req)
	scheduler.urlMap.Put(reqURL.String(), struct{}{})
	return true
}

// sendResp 会向响应缓冲池发送响应。
func sendResp(resp *module.Response, respBufferPool buffer.Pool) bool {
	if resp == nil || respBufferPool == nil || respBufferPool.Closed() {
		return false
	}
	go func(resp *module.Response) {
		if err := respBufferPool.Put(resp); err != nil {
			logger.Warnln("The response buffer pool was closed. Ignore response sending.")
		}
	}(resp)
	return true
}

// sendItem 会向条目缓冲池发送条目。
func sendItem(item module.Item, itemBufferPool buffer.Pool) bool {
	if item == nil || itemBufferPool == nil || itemBufferPool.Closed() {
		return false
	}
	go func(item module.Item) {
		if err := itemBufferPool.Put(item); err != nil {
			logger.Warnln("The item buffer pool was closed. Ignore item sending.")
		}
	}(item)
	return true
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

// 注册给定的组件
func (scheduler *myScheduler) registerModules(moduleArgs ModuleArgs) error {
	// 注册所有下载器
	for _, d := range moduleArgs.Downloaders {
		if d == nil {
			continue
		}
		ok, err := scheduler.registrar.Register(d)
		if err != nil {
			return genErrorByError(err)
		}
		if !ok {
			return genError(fmt.Sprintf("Could't register analyer instance with MID: %q !", d.ID()))
		}
	}
	logger.Infof("All analyzers have been registered. (number: %d)",
		len(moduleArgs.Downloaders))
	// 注册分析器
	for _, a := range moduleArgs.Analyzers {
		if a == nil {
			continue
		}
		ok, err := scheduler.registrar.Register(a)
		if err != nil {
			return genErrorByError(err)
		}
		if !ok {
			errMsg := fmt.Sprintf("Couldn't register analyzer instance with MID %q!", a.ID())
			return genError(errMsg)
		}
	}
	logger.Infof("All analyzers have been registered. (number: %d)",
		len(moduleArgs.Analyzers))
	// 注册管道
	for _, p := range moduleArgs.Pipelines {
		if p == nil {
			continue
		}
		ok, err := scheduler.registrar.Register(p)
		if err != nil {
			return genErrorByError(err)
		}
		if !ok {
			errMsg := fmt.Sprintf("Couldn't register pipeline instance with MID %q!", p.ID())
			return genError(errMsg)
		}
	}
	logger.Infof("All pipelines have been registered. (number: %d)",
		len(moduleArgs.Pipelines))
	return nil
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

// checkBufferPoolForStart 会检查缓冲池是否已为调度器的启动准备就绪。
// 如果某个缓冲池不可用，就直接返回错误值报告此情况。
// 如果某个缓冲池已关闭，就按照原先的参数重新初始化它。
func (scheduler *myScheduler) checkBufferPoolForStart() error {
	// 检查请求缓冲池。
	if scheduler.reqBufferPool == nil {
		return genError("nil request buffer pool")
	}
	if scheduler.reqBufferPool != nil && scheduler.reqBufferPool.Closed() {
		scheduler.reqBufferPool, _ = buffer.NewPool(
			scheduler.reqBufferPool.BufferCap(), scheduler.reqBufferPool.MaxBufferNumber())
	}
	// 检查响应缓冲池。
	if scheduler.respBufferPool == nil {
		return genError("nil response buffer pool")
	}
	if scheduler.respBufferPool != nil && scheduler.respBufferPool.Closed() {
		scheduler.respBufferPool, _ = buffer.NewPool(
			scheduler.respBufferPool.BufferCap(), scheduler.respBufferPool.MaxBufferNumber())
	}
	// 检查条目缓冲池。
	if scheduler.itemBufferPool == nil {
		return genError("nil item buffer pool")
	}
	if scheduler.itemBufferPool != nil && scheduler.itemBufferPool.Closed() {
		scheduler.itemBufferPool, _ = buffer.NewPool(
			scheduler.itemBufferPool.BufferCap(), scheduler.itemBufferPool.MaxBufferNumber())
	}
	// 检查错误缓冲池。
	if scheduler.errorBufferPool == nil {
		return genError("nil error buffer pool")
	}
	if scheduler.errorBufferPool != nil && scheduler.errorBufferPool.Closed() {
		scheduler.errorBufferPool, _ = buffer.NewPool(
			scheduler.errorBufferPool.BufferCap(), scheduler.errorBufferPool.MaxBufferNumber())
	}
	return nil
}

// 重置调度器上下文
func (scheduler *myScheduler) resetContext() {
	scheduler.ctx, scheduler.cancelFunc = context.WithCancel(context.Background())
}

// canceled 用于判断调度器的上下文是否已被取消。
func (sched *myScheduler) canceled() bool {
	select {
	case <-sched.ctx.Done():
		return true
	default:
		return false
	}
}



