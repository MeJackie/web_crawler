package analyzer

import "web_crawler/errors"

// 生成爬虫参数错误值
func genParametarError(errMsg string) error {
	return errors.NewCrawlerErrorBy(errors.ERROR_TYPE_ANALYER,
		errors.NewIllegalParameterError(errMsg))
}