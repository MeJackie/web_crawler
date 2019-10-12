package buffer

import (
	"fmt"
	"sync"
	"sync/atomic"
	"web_crawler/errors"
)

// FIFO
type Buffer interface {
	Cap() uint32
	// 获取缓冲器中数据的数量
	Len() uint32
	// 向缓存器中写入数据 ！本方法是非堵塞的
	Put(datum interface{}) (bool, error)
	// 从缓冲器中读取数据 ！本方法是非堵塞的
	Get()(datum interface{}, err error)
	Close() bool
	// 用于判断缓存器是否已关闭
	Closed() bool
}

// 缓冲器接口的实现类型
type myBuffer struct {
	ch chan interface{}
	// 0-未关闭 1-已关闭
	closed uint32
	// 为了消除因关闭缓冲器而产生的竞态条件的读写锁
	closingLock sync.RWMutex
}

func NewBuffer(size uint32) (Buffer, error)  {
	if size == 0 {
		errMsg := fmt.Sprintf("illegal size for buffer : %d", size)
		return nil, errors.NewIllegalParameterError(errMsg)
	}
	
	return &myBuffer{
		ch:          make(chan interface{}, size),
	}, nil
}

func (buf *myBuffer) Cap() uint32 {
	return uint32(cap(buf.ch))
}

func (buf *myBuffer) Len() uint32 {
	return uint32(len(buf.ch))
}

func (buf *myBuffer) Put(datum interface{}) (ok bool, err error) {
	buf.closingLock.RLock()
	defer buf.closingLock.RUnlock()
	if buf.Closed() {
		return false, ErrClosedBuffer
	}
	select {
	case buf.ch <- datum:
		ok = true
	default:
		ok = false
	}

	return
}


func (buf *myBuffer) Get() (datum interface{}, err error) {
	select {
	case datum, ok := <- buf.ch:
		if !ok {
			return nil, ErrClosedBuffer
		}
		return datum, nil
	default:
		return nil, nil
	}
}

func (buf *myBuffer) Close() bool {
	if atomic.CompareAndSwapUint32(&buf.closed, 0, 1) {
		buf.closingLock.Lock()
		close(buf.ch)
		buf.closingLock.Unlock()
		return true
	}
	return false
}

func (buf *myBuffer) Closed() bool {
	if atomic.LoadUint32(&buf.closed) == 0 {
		return false
	}
	return true
}