package run

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	panicPrefix = "panic: "
	someErrText = "some error"
)

var errInner = errors.New(someErrText)

func TestWithError(t *testing.T) {
	type args struct{ fn func() error }
	tests := []struct {
		name string
		args args
		want error
	}{
		{
			name: "no err",
			args: args{func() error { return nil }},
			want: nil,
		},
		{
			name: "usual err",
			args: args{func() error { return errInner }},
			want: errInner,
		},
		{
			name: "panic with text",
			args: args{func() error { panic(someErrText) }},
			want: fmt.Errorf(panicPrefix+"%v", someErrText),
		},
		{
			name: "panic with err",
			args: args{func() error { panic(errInner) }},
			want: errInner,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			requires := require.New(t)

			var err error
			requires.NotPanics(func() {
				err = WithError(tt.args.fn)
			})

			if tt.want == nil {
				requires.NoError(err)
			} else {
				requires.Error(err)
				if strings.HasPrefix(err.Error(), panicPrefix) {
					requires.EqualError(err, tt.want.Error())
				} else {
					requires.ErrorIs(err, tt.want)
				}
			}
		})
	}
}

func TestAsyncWithError(t *testing.T) {
	type args struct{ fn func() error }
	tests := []struct {
		name string
		args args
		want error
	}{
		{
			name: "no err",
			args: args{func() error { return nil }},
			want: nil,
		},
		{
			name: "usual err",
			args: args{func() error { return errInner }},
			want: errInner,
		},
		{
			name: "panic with text",
			args: args{func() error { panic(someErrText) }},
			want: fmt.Errorf(panicPrefix+"%v", someErrText),
		},
		{
			name: "panic with err",
			args: args{func() error { panic(errInner) }},
			want: errInner,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			requires := require.New(t)

			var err error
			requires.NotPanics(func() {
				err = <-AsyncWithError(tt.args.fn)
			})

			if tt.want == nil {
				requires.NoError(err)
			} else {
				requires.Error(err)
				if strings.HasPrefix(err.Error(), panicPrefix) {
					requires.EqualError(err, tt.want.Error())
				} else {
					requires.ErrorIs(err, tt.want)
				}
			}
		})
	}
}
