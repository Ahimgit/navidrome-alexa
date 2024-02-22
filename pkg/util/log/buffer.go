package log

import (
	"sync"
)

const (
	initialSize = 0x400
	maxSize     = 0x4000
)

var pool = sync.Pool{
	New: func() any {
		buffer := make(PooledBuffer, 0, initialSize)
		return &buffer
	},
}

type PooledBuffer []byte

func GetBuffer() *PooledBuffer {
	return pool.Get().(*PooledBuffer)
}

func (wb *PooledBuffer) Release() {
	if cap(*wb) < maxSize {
		*wb = (*wb)[0:0] // reset to 0
		pool.Put(wb)
	}
}

func (wb *PooledBuffer) AddChr(char byte) *PooledBuffer {
	*wb = append(*wb, char)
	return wb
}

func (wb *PooledBuffer) AddStr(str string) *PooledBuffer {
	*wb = append(*wb, str...)
	return wb
}

func (wb *PooledBuffer) AddBytes(data []byte) *PooledBuffer {
	*wb = append(*wb, data...)
	return wb
}
