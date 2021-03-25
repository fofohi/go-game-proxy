package util

import (
	"sync"
)

var (
	smallBufferSize  = 5 * 1024  // 4KB small buffer
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
)


func GetSmallPool() []byte  {
	b := SPool.Get().([]byte)
	defer SPool.Put(b)
	return b
}

func GetMiddlePool() []byte  {
	b := MPool.Get().([]byte)
	return b
}
