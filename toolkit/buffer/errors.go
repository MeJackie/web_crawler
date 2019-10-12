package buffer

import "errors"

// 表示缓存器已关闭错误
var ErrClosedBuffer = errors.New("closed buffer")

// 表示缓存器已关闭错误
var ErrClosedBufferPool = errors.New("closed buffer poll")