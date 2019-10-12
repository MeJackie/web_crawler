package module

// 合法组件类型-字母映射表
var legalTypeLetterMap = map[Type]string{
	TYPE_DOWNLOADER: "D",
	TYPE_ANALYZER: "A",
	TYPE_PIPELINE: "P",
}

// legalLetterTypeMap 代表合法的字母-组件类型的映射。
var legalLetterTypeMap = map[string]Type{
	"D": TYPE_DOWNLOADER,
	"A": TYPE_ANALYZER,
	"P": TYPE_PIPELINE,
}