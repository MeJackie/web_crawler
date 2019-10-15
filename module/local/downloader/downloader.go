package downloader

import (
	"net/http"
	"web_crawler/module"
	"web_crawler/module/stub"
)

// 日志记录器
//var logger = log.DLogger()

// 下载器的实现类型
type myDownloader struct {
	// 组件基础接口实现  (匿名段)
	stub.ModuleInternal
	// 下载用的HTTP客户端
	httpClient http.Client
}

// 创建下载实例
// 注意，这里还隐藏着一个 Go语言的命名惯例。由于下载器的实现代码独占一个代码包，所以可以让这个函数的名称足够简单，只有一个单词 New。这不同于前面提到的函数 NewPool 和 NewMultipleReader，这两个函数所创建的实例的含义无法由其所在代码包的名称 buffer 和 reader 表达。
func New (
	mid module.MID,
	client *http.Client,
	scoreCalculator module.CalculateScore) (module.Downloader, error) {
		moduleBase, err := stub.NewModuleInternal(mid, scoreCalculator)
		if err != nil {
			return nil, err
		}

		if client == nil {
			return nil, genParametarError("nil http client")
		}

		return &myDownloader{
			ModuleInternal: moduleBase,
			httpClient:     *client,
		}, nil
}

func (downloader *myDownloader) Download(req *module.Request) (*module.Response, error) {
	downloader.ModuleInternal.IncrHandlingCount()
	defer downloader.ModuleInternal.DecrHandlingCount()
	downloader.ModuleInternal.IncrCalledCount()
	if req == nil {
		return nil, genParametarError("nil request")
	}

	httpReq := req.HTTPReq()
	if httpReq == nil {
		return nil, genParametarError("nil http request")
	}
	downloader.ModuleInternal.IncrAcceptedCount()
	//logger.Infof("Do the request (URL: %s, Depth: %d)... \n", httpReq.URL, req.Depth())
	httpResp, err := downloader.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	downloader.ModuleInternal.IncrCompletedCount()
	return module.NewRespone(httpResp, req.Depth()), nil
}