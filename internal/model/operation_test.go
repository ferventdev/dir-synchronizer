package model

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOperation_IsNotNilAndOver(t *testing.T) {
	tests := []struct {
		name      string
		operation *Operation
		want      bool
	}{
		{name: "nil", operation: nil, want: false},
		{name: "new", operation: NewOperation(OpKindCopyFile), want: false},
		{name: "scheduled", operation: &Operation{Status: OpStatusScheduled}, want: false},
		{name: "in_progress", operation: &Operation{Status: OpStatusInProgress}, want: false},
		{name: "canceled", operation: &Operation{Status: OpStatusCanceled}, want: true},
		{name: "completed", operation: &Operation{Status: OpStatusCompleted}, want: true},
		{name: "failed", operation: &Operation{Status: OpStatusFailed}, want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, tt.operation.IsNotNilAndOver())
		})
	}
}
