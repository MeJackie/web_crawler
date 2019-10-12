package stub

import (
	"fmt"
	"web_crawler/errors"
	"web_crawler/module"
)

// 组件内部基础接口的实现类型
type myModule struct {
	//组件ID
	mid module.MID
	//组件的网络地址
	addr string
	//组件评分
	score uint64
	//评分计算器
	scoreCalculator module.CalculateScore
	//调用计数
	calledCount uint64
	//接受计数
	acceptedCount uint64
	//成功完成计数
	completedCount uint64
	//实时处理数
	handlingNumber uint64
}

func NewModuleInternal(
	mid module.MID,
	scoreCalculator module.CalculateScore) (ModuleInternal, error)  {
		parts, err := module.SplitMID(mid)
		if err != nil {
			return nil, errors.NewIllegalParameterError(
				fmt.Sprintf("illegel ID %s: %s", mid, err))
		}

		return &myModule{
			mid:             mid,
			addr:            parts[2],
			scoreCalculator: scoreCalculator,
		}, nil
}

// 组件ID
func (m *myModule) ID() MID {

}
// 当前组件的网络地址
func (m *myModule) Addr() string {

}
// 获取当前组件的评分
func (m *myModule) Score() uint64 {

}
// 设置当前组件的评分
func (m *myModule) SetScore(score uint64) {

}
// 获取评分计算器
func (m *myModule) ScoreCalculator() CalculateScore {

}
// 获取当前组件被调用的计数
func (m *myModule) CalledCount() uint64 {

}
// 获取当前组件接收调用计数
// 组件一般会由于超负荷或参数有误而拒绝调用
func (m *myModule) AcceptedCount() uint64 {

}
// 组件成功调用计数
func (m *myModule) CompletedCount() uint64 {

}
// 组件正在处理的调用的数量
func (m *myModule) HandlingNumber() uint64 {

}
// 获取所有计数
func (m *myModule) Counts() Counts {

}
// 获取组件摘要
func (m *myModule) Summary() SummaryStruct {

}

func (m *myModule) IncrCalledCount() {

}
func (m *myModule) IncrAcceptedCount() {

}
func (m *myModule) IncrCompletedCount() {

}
func (m *myModule) IncrHandlingCount() {

}
func (m *myModule) DecrHandlingCount() {

}
// 清空统计
func (m *myModule) Clear() {

}