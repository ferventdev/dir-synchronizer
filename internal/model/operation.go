package model

import (
	"context"
	"dsync/pkg/helpers/ut"
	"time"
)

//OperationStatus is a status of a sync operation.
type OperationStatus string

const (
	OpStatusScheduled  = "scheduled"
	OpStatusInProgress = "in_progress"
	OpStatusCanceled   = "canceled"
	OpStatusCompleted  = "completed"
)

type OperationKind string

const (
	OpKindNone    = "none"
	OpKindCopy    = "copy"
	OpKindRemove  = "remove"
	OpKindReplace = "replace"
)

var generateOperationID = ut.CreateUint64IDGenerator()

// Operation - synchronization operation between the dir entry in the source directory and same entry in the copy directory.
type Operation struct {
	ID          uint64             `json:"id"`
	Status      OperationStatus    `json:"status"`
	Kind        OperationKind      `json:"kind"`
	CancelFn    context.CancelFunc `json:"-"`
	ScheduledAt time.Time          `json:"scheduledAt"`
	StartedAt   *time.Time         `json:"startedAt,omitempty"`
	CanceledAt  *time.Time         `json:"canceledAt,omitempty"`
	CompletedAt *time.Time         `json:"completedAt,omitempty"`
}

func NewOperation(kind OperationKind) *Operation {
	return &Operation{ID: generateOperationID(), Status: OpStatusScheduled, Kind: kind, ScheduledAt: time.Now()}
}

func (op *Operation) IsNotNilAndOver() bool {
	return op != nil && (op.Status == OpStatusCanceled || op.Status == OpStatusCompleted)
}
