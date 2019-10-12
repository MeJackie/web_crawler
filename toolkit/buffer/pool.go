package buffer

import (
	"fmt"
	"sync"
	"sync/atomic"
	"web_crawler/errors"
)

type Pool interface {
	BufferCap() uint32
	MaxBufferNumber() uint32
	// 实际持有缓存器数量
	BufferNumber() uint32
	// 获取数据总数
	Total() uint64
	// !方法是阻塞的
	Put(datum interface{}) error
	// !方法是阻塞的
	Get() (datum interface{}, err error)
	Close() bool
	// 判断缓冲池是否已关闭
	Closed() bool
}

// 数据缓存池接口实现类型
type myPool struct {
	// 缓冲器统一容量
	bufferCap uint32
	maxBufferNumber uint32
	bufferNmber uint32
	// 池子中数据总数
	total uint64
	bufCh chan Buffer
	// 0-未关闭  1-已关闭
	closed uint32
	// 保护内部共享资源的读写锁
	rwLock sync.RWMutex
}

func NewPool(
	bufferCap uint32,
	maxBufferNumber uint32) (Pool, error) {
	if bufferCap == 0 {
		errMsg := fmt.Sprintf("illegal buffer cap for buffer pool: %d", bufferCap)
		return nil, errors.NewIllegalParameterError(errMsg)
	}

	if maxBufferNumber == 0 {
		errMsg := fmt.Sprintf("illegal max buffer number for buffer pool: %d", maxBufferNumber)
		return nil, errors.NewIllegalParameterError(errMsg)
	}

	bufCh := make(chan Buffer, maxBufferNumber)
	buf, _ := NewBuffer(bufferCap)
	bufCh <- buf
	return &myPool{
		bufferCap:       bufferCap,
		maxBufferNumber: maxBufferNumber,
		bufferNmber:     1,
		total:           0,
		bufCh:           bufCh,
		closed:          0,
	}, nil
}

func (pool *myPool)  BufferCap() uint32 {
	return pool.bufferCap
}

func (pool *myPool)  MaxBufferNumber() uint32 {
	return pool.maxBufferNumber
}

func (pool *myPool)  BufferNumber() uint32 {
	return atomic.LoadUint32(&pool.bufferNmber)
}

func (pool *myPool)  Total() uint64 {
	return atomic.LoadUint64(&pool.total)
}

func (pool *myPool)  Put(datum interface{}) (err error) {
	if pool.Closed() {
		return ErrClosedBufferPool
	}

	// 当前已满尝试次数
	var count uint32
	// 允许已满尝试次数
	var maxCount = pool.BufferNumber() * 5
	for buf := range pool.bufCh{
		ok, err := pool.putDdata(buf, datum, &count, maxCount)
		if ok || err != nil {
			break
		}
	}

	return
}

func (pool *myPool) putDdata(buf Buffer, datum interface{}, count *uint32, maxCount uint32) (ok bool, err error){
	if pool.Closed() {
		return false, ErrClosedBufferPool
	}

	// 归还缓冲器
	defer func() {
		pool.rwLock.RLock()
		if pool.Closed() {    // 池子已关闭，不归还缓冲器
			// ^uint32(0) = -1
			atomic.AddUint32(&pool.bufferNmber, ^uint32(0))
			err = ErrClosedBufferPool
		} else {
			pool.bufCh <- buf
		}
		pool.rwLock.RUnlock()
	}()

	// 向缓冲器写入数据
	ok, err = buf.Put(datum)
	if ok {
		atomic.AddUint64(&pool.total, 1)
	}

	if err != nil {
		return
	}

	// 若缓冲器已满而未放入数据，就递增计数
	(*count)++

	// 如果尝试向缓冲器放入数据的失败次数达到阈值，
	// 并且池中缓冲器的数量未达到最大值，
	// 那么就尝试创建一个新的缓冲器，先放入数据再把它放入池
	if *count >= maxCount && pool.BufferNumber() < pool.MaxBufferNumber() {
		// 第一：防止向已关闭的缓冲池追加缓冲器。
		// 第二：防止缓冲器的数量超过最大值。在确保这两种情况不会发生后，把一个已放入那个数据的缓冲器追加到缓冲池中。
		pool.rwLock.Lock()
		if pool.bufferNmber < pool.MaxBufferNumber() {
			if pool.Closed() {
				pool.rwLock.Unlock()
				return
			}
			newBuf, err := NewBuffer(pool.bufferCap)
			if err != nil {
				return
			}
			_, err = newBuf.Put(datum)
			if err != nil {
				return
			}
			pool.bufCh <- newBuf
			atomic.AddUint32(&pool.bufferNmber, 1)
			atomic.AddUint64(&pool.total, 1)
			ok = true
		}
		pool.rwLock.Unlock()
		*count = 0
	}
}

func (pool *myPool)  Get() (datum interface{}, err error) {
	if pool.Closed() {
		return nil, ErrClosedBufferPool
	}
	// 当前已空尝试次数
	var count uint32
	// 准许已空尝试次数
	maxCount := pool.BufferNumber() * 10
	for buf := range pool.bufCh {
		datum, err := pool.getData(buf, &count, maxCount)
		if datum != nil || err != nil {
			break
		}
	}
	return
}

func (pool *myPool) getData(buf Buffer, count *uint32, maxCount uint32) (datum interface{}, err error)  {
	if pool.Closed() {
		return nil, ErrClosedBufferPool
	}

	// 归还缓冲器
	defer func() {
		// 如果尝试从缓冲器获取数据的失败次数达到阈值，
		// 同时当前缓冲器已空且池中缓冲器的数量大于1,
		// 那么就直接关掉当前缓冲器，并不归还给池
		if *count >= maxCount && pool.BufferNumber() > 1 {
			buf.Close()
			atomic.AddUint32(&pool.bufferNmber, ^uint32(0))
			*count = 0
			return
		}

		pool.rwLock.RLock()
		if pool.Closed() {
			atomic.AddUint32(&pool.bufferNmber, ^uint32(0))
			err = ErrClosedBufferPool
		} else {
			pool.bufCh <- buf
		}
		pool.rwLock.RUnlock()
	}()

	datum, err = buf.Get()
	if datum != nil {
		atomic.AddUint64(&pool.total, ^uint64(0))
		return
	}
	if err != nil {
		return
	}

	// 若因缓冲器已空未取出数据，就递增计数
	(*count)++
	return
}

func (pool *myPool)  Close() bool {
	if !atomic.CompareAndSwapUint32(&pool.closed, 0, 1) {
		return false
	}
	pool.rwLock.Lock()
	defer pool.rwLock.Unlock()
	close(pool.bufCh)
	for buf := range pool.bufCh {
		buf.Closed()
	}
	return true
}

func (pool *myPool)  Closed() bool {
	if atomic.LoadUint32(&pool.closed) == 0 {
		return false
	}

	return true
}

