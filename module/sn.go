package module

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