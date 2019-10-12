package module

// 组件ID模板
var midTemplate = "%s%d|%s"

// 组件ID
type MID string

// 计算组件评分函数
type CalculateScore func(counts Counts) uint64

// SplitMID 用于分解组件ID。
// 第一个结果值表示分解是否成功。
// 若分解成功，则第二个结果值长度为3，
// 并依次包含组件类型字母、序列号和组件网络地址（如果有的话）。
func SplitMID(mid MID) ([]string, error) {

}