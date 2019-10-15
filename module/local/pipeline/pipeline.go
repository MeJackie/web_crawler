package pipeline

import (
	"fmt"
	"web_crawler/module"
	"web_crawler/module/stub"
)

// 条目处理管道实现类型
type myPipeline struct {
	// 组件基础实例
	stub.ModuleInternal
	// 代表条目处理器列表
	itemProcessors []module.ProcessItem
	// 代表处理是否需要快速失败
	failFast bool
}

func New(
	mid module.MID,
	itemProcessors []module.ProcessItem,
	scoreCalculator module.CalculateScore) (module.Pipeline, error) {
		moduleBase, err := stub.NewModuleInternal(mid, scoreCalculator)
		if err != nil {
			return nil, err
		}

		if itemProcessors == nil {
			return nil, genParametarError("nil item processor list")
		}

		if len(itemProcessors) == 0 {
			return nil, genParametarError("empty item processor list")
		}

		var innerProcessors []module.ProcessItem
		for i, pipeline := range itemProcessors {
			if pipeline == nil {
				return nil, genParametarError(fmt.Sprintf("nil item processor[%d]", i))
			}
			innerProcessors = append(innerProcessors, pipeline)
		}

		return &myPipeline{
			ModuleInternal: moduleBase,
			itemProcessors: innerProcessors,
		}, nil
}

// 返回当前条目处理管道使用的条目处理函数的列表
func (pipeline *myPipeline) ItemProcessors() []module.ProcessItem {
	processors := make([]module.ProcessItem, len(pipeline.itemProcessors))
	copy(processors, pipeline.itemProcessors)
	return processors
}
// 向条目管道发送条目
// 条目需要依次经过若干条目处理函数的处理
func (pipeline *myPipeline) Send(item module.Item) []error {
	pipeline.ModuleInternal.IncrHandlingCount()
	defer pipeline.ModuleInternal.DecrHandlingCount()
	pipeline.ModuleInternal.IncrCalledCount()
	var errs []error
	if item == nil {
		errs = append(errs, genParametarError("nil item"))
		return errs
	}

	pipeline.IncrAcceptedCount()
	var currentItem = item
	for _, processor := range pipeline.itemProcessors {
		processedItem, err := processor(currentItem)
		if err != nil {
			errs = append(errs, err)
			if pipeline.FailFast() {
				break
			}
		}
		if processedItem != nil {
			currentItem = processedItem
		}
	}
	if len(errs) == 0 {
		pipeline.ModuleInternal.IncrCompletedCount()
	}
	return nil
}
// 方法返回布尔值，该值表示当前条目管道是否是快速失败的
// 这里的快速失败是指：只要在处理某个条目时在某一个步骤上出错，
// 那么条目处理管道就会忽略掉后续的所有处理步骤并报告错误
func (pipeline *myPipeline) FailFast() bool {
	return pipeline.failFast
}
// 设置是否快速失败
func (pipeline *myPipeline) SetFailFast(failFast bool) {
	pipeline.failFast = failFast
}