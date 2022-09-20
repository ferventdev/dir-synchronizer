package ut

import (
	"errors"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateUint64IDGenerator(t *testing.T) {
	const (
		numOfGenerators = 5
		numOfCalls      = 100
	)
	generators := make([]func() uint64, 0, numOfGenerators)
	for i := 0; i < numOfGenerators; i++ {
		generators = append(generators, CreateUint64IDGenerator())
	}

	var gensWg sync.WaitGroup
	for _, gen := range generators {
		gensWg.Add(1)

		go func(getNext func() uint64) {
			defer gensWg.Done()

			var wg sync.WaitGroup
			for i := 0; i < numOfCalls; i++ {
				wg.Add(1)
				go func() {
					_ = getNext() // we're incrementing same counter in parallel in different goroutines
					wg.Done()
				}()
			}
			wg.Wait()
		}(gen)
	}
	gensWg.Wait()

	asserts := assert.New(t)
	for _, gen := range generators {
		getNext := gen
		for i := 1; i <= 3; i++ {
			asserts.Equal(uint64(numOfCalls+i), getNext())
		}
	}
}

func TestIsSameError(t *testing.T) {
	type args struct {
		err error
		tgt error
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{name: "both nil", args: args{err: nil, tgt: nil}, want: true},
		{name: "err nil", args: args{err: nil, tgt: errors.New("some")}, want: false},
		{name: "tgt nil", args: args{err: errors.New("some"), tgt: nil}, want: false},
		{name: "differ", args: args{err: errors.New("one"), tgt: errors.New("two")}, want: false},
		{name: "same", args: args{err: errors.New("one"), tgt: errors.New("one")}, want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, IsSameError(tt.args.err, tt.args.tgt))
		})
	}
}
