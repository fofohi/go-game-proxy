package util

import (
	"github.com/panjf2000/ants/v2"
	"sync"
)

var (
	smallBufferSize  = 4 * 1024  // 4KB small buffer
	mediumBufferSize = 10 * 1024  // 8KB medium buffer

	SPool = sync.Pool{
		New: func() interface{} {
			return make([]byte, smallBufferSize)
		},
	}
	MPool = sync.Pool{
		New: func() interface{} {
			return make([]byte, mediumBufferSize)
		},
	}
	threadPool,_ = ants.NewPool(100)
)


func GetSmallPool() []byte  {
	b := SPool.Get().([]byte)
	defer SPool.Put(b)
	return b
}

func GetThreadPool()  *ants.Pool{
	return threadPool
}

func GetMiddlePool() []byte  {
	b := MPool.Get().([]byte)
	return b
}
