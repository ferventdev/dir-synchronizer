package ut

import "sync/atomic"

func CreateUint64IDGenerator() func() uint64 {
	counter := new(uint64)
	return func() uint64 {
		return atomic.AddUint64(counter, 1)
	}
}
