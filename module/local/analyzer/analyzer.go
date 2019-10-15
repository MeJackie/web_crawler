package analyzer

import (
	"fmt"
	"web_crawler/module"
	"web_crawler/module/stub"
	"web_crawler/toolkit/reader"
)

// logger 代表日志记录器。
//var logger = log.DLogger()

type myAnalyzer struct {
	// 组件基础实例
	stub.ModuleInternal
	// 响应解析器列表
	respParse []module.ParseResponse
}

func New(
	mid module.MID,
	respParsers []module.ParseResponse,
	scoreCalculator module.CalculateScore) (module.Analyzer, error) {
		moduleBase, err := stub.NewModuleInternal(mid, scoreCalculator)
		if err != nil {
			return nil, err
		}

		if respParsers == nil {
			return nil, genParametarError("nil response parsers")
		}

		if len(respParsers) == 0 {
			return nil, genParametarError("nil response parser list")
		}

		var innerParsers []module.ParseResponse
		for i, parser := range respParsers {
			if parser == nil {
				return nil, genParametarError(fmt.Sprintf("nil response parser[%d]", i))
			}
			innerParsers = append(innerParsers, parser)
		}

		return &myAnalyzer{
			ModuleInternal: moduleBase,
			respParse:      innerParsers,
		}, nil
}

func (analyzer *myAnalyzer) ResqParsers() []module.ParseResponse{

}

func (analyzer *myAnalyzer) Analyze(
	resq *module.Response) (dataList []module.Data, errorList []error) {
		analyzer.ModuleInternal.IncrHandlingCount()
		defer analyzer.ModuleInternal.DecrHandlingCount()
		analyzer.ModuleInternal.IncrCalledCount()
		if resq == nil {
			errorList = append(errorList,
				genParametarError("nil response"))
			return
		}

		httpResp := resq.HTTPResp()
		if httpResp == nil {
			errorList = append(errorList,
				genParametarError("nil http response"))
			return
		}

		httpReq := httpResp.Request
		if httpReq == nil {
			errorList = append(errorList,
				genParametarError("nil http request"))
			return
		}

		reqUrl := httpReq.URL
		if reqUrl == nil {
			errorList = append(errorList,
				genParametarError("nil http reqeust url"))
			return
		}

		analyzer.ModuleInternal.IncrAcceptedCount()
		respDepth := resq.Depth()
		//logger.Infof("Parse the response (URL: %s, depth: %d)... \n",
		//	reqURL, respDepth)

		// 解析http响应
		if httpResp.Body != nil {
			defer httpResp.Body.Close()
		}

		multipleReader, err := reader.NewMultipleReader(httpResp.Body)
		if err != nil {
			errorList = append(errorList, err)
			return
		}

		dataList = []module.Data{}
		for _, respParser := range analyzer.respParse {
			httpResp.Body = multipleReader.Reader()
			pDataList, pErrorList := respParser(httpResp, respDepth)
			if pDataList != nil {
				for _, pData := range pDataList {
					if pData == nil {
						continue
					}
					dataList = appendDataList(pDataList,pData,respDepth)
				}
			}
			if pErrorList != nil {
				for _, pError := range pErrorList {
					if pError == nil {
						continue
					}
					errorList = append(errorList, pError)
				}
			}

			if len(errorList) == 0 {
				analyzer.ModuleInternal.IncrCompletedCount()
			}
			return dataList, errorList
		}
}

// 添加请求值或条目值到列表
func appendDataList(dataList []module.Data, data module.Data, respDepth uint32) []module.Data {
	if data == nil {
		return dataList
	}
	req, ok := data.(*module.Request)
	if !ok {
		return append(dataList, data)
	}

	newDepth := respDepth + 1
	if req.Depth() != newDepth {
		req = module.NewRequest(req.HTTPReq(), newDepth)
	}
	return append(dataList, req)
}
