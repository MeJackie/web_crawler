package stub

import "web_crawler/module"

// 组件内部基础接口类型
type ModuleInternal interface {
	module.Module
	IncrCalledCount()
	IncrAcceptedCount()
	IncrCompletedCount()
	IncrHandlingCount()
	DecrHandlingCount()
	// 清空统计
	Clear()
}

