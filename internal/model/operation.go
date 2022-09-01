package model

import (
	"context"
	"dsync/pkg/helpers/ut"
	"time"
)

//OperationStatus is a status of a sync operation.
type OperationStatus string

const (
	OperationScheduled  = "scheduled"
	OperationInProgress = "in_progress"
	OperationCanceled   = "canceled"
	OperationCompleted  = "completed"
)

type OperationKind string

var generateOperationID = ut.CreateUint64IDGenerator()

// Operation - synchronization operation between the dir entry in the source directory and same entry in the copy directory.
type Operation struct {
	ID          uint64
	Status      OperationStatus
	Kind        OperationKind
	CancelFn    context.CancelFunc
	ScheduledAt time.Time
	StartedAt   time.Time
	CanceledAt  time.Time
	CompletedAt time.Time
}

func NewOperation(kind OperationKind) *Operation {
	return &Operation{ID: generateOperationID(), Status: OperationScheduled, Kind: kind, ScheduledAt: time.Now()}
}

func (op *Operation) IsNotNilAndOver() bool {
	return op != nil && (op.Status == OperationCanceled || op.Status == OperationCompleted)
}
