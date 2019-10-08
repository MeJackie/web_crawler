package module

// 组件类型
type Type string

// 当前组件类型
const (
	//下载器
	TYPE_DOWNLOADER Type = "downloader"
	//分析器
	TYPE_ANALYZER Type = "analyzer"
	//条目处理管道
	TYPE_PIPELINE Type = "pipeline"
)

// 合法组件类型-字母映射表
var legalTypeLetterMap = map[Type]string{
	TYPE_DOWNLOADER: "D",
	TYPE_ANALYZER: "A",
	TYPE_PIPELINE: "P",
}

// Module代表组件的基础接口类型
// 该接口的实现类型必须是并发安全的
type Module interface {
	// 组件ID
	ID() MID
	// 当前组件的网络地址
	Addr() string
	// 获取当前组件的评分
	Score() uint64
	// 设置当前组件的评分
	SetScore(score uint64)
	// 获取评分计算器
	ScoreCalculator() CalculateScore
	// 获取当前组件被调用的计数
	CalledCount() uint64
	// 获取当前组件接收调用计数
	// 组件一般会由于超负荷或参数有误而拒绝调用
	AcceptedCount() uint64
	// 组件成功调用计数
	CompletedCount() uint64
	// 组件正在处理的调用的数量
	HandlingNumber() uint64
	// 获取所有计数
	Counts() Counts
	// 获取组件摘要
	Summary() SummaryStruet
}

// 序列号生成器接口类型
type SNGenertor interface {
	// 获取预设的最小序列号
	Start() uint64
	// 获取预设的最大序列号
	Max() uint64
	// 获取当前序列号，并准备下一个序列号
	Get() uint64
	// 获取下一个序列号
	Next() uint64
	// 循环计数
	CycleCount() uint64
}