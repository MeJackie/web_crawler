package scheduler

import "web_crawler/errors"

// 生成爬虫错误值
func genError(errMsg string) error {
	return errors.NewCrawlerError(errors.ERROR_TYPE_SCHEDULER,
		errMsg)
}