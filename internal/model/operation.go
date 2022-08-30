package model

import "time"

//OperationStatus is a status of a sync operation.
type OperationStatus string

const (
	OperationScheduled  = "scheduled"
	OperationInProgress = "in_progress"
	OperationCanceled   = "canceled"
	OperationCompleted  = "completed"
)

// Operation - synchronization operation between the dir entry in the source directory and same entry in the copy directory.
type Operation struct {
	Status      OperationStatus
	ScheduledAt time.Time
	StartedAt   time.Time
	CanceledAt  time.Time
	CompletedAt time.Time
}

func (op *Operation) IsNotNilAndOver() bool {
	return op != nil && (op.Status == OperationCanceled || op.Status == OperationCompleted)
}
