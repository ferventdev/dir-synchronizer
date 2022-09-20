package ut

import (
	"strings"
	"sync/atomic"
)

//CreateUint64IDGenerator returns an ID generator, which is a function that is safe to be called concurrently.
func CreateUint64IDGenerator() func() uint64 {
	counter := new(uint64)
	return func() uint64 {
		return atomic.AddUint64(counter, 1)
	}
}

//IsSameError should be used only when it's impossible to compare errors by errors.Is()
//because original target is unexported or platform dependent.
func IsSameError(err, tgt error) bool {
	return (err == nil && tgt == nil) || (err != nil && tgt != nil && strings.Contains(err.Error(), tgt.Error()))
}
