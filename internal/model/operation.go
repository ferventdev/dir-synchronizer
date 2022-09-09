package model

import (
	"context"
	"dsync/pkg/helpers/ut"
	"time"
)

//OperationStatus is a status of a sync operation.
type OperationStatus string

const (
	OpStatusScheduled  OperationStatus = "scheduled"
	OpStatusInProgress OperationStatus = "in_progress"
	OpStatusCanceled   OperationStatus = "canceled"
	OpStatusCompleted  OperationStatus = "completed"
)

type OperationKind string

const (
	OpKindNone               OperationKind = "none"
	OpKindCopyFile           OperationKind = "copy_file"
	OpKindRemoveFile         OperationKind = "remove_file"
	OpKindRemoveDir          OperationKind = "remove_dir"
	OpKindReplaceFile        OperationKind = "replace_file"
	OpKindReplaceDirWithFile OperationKind = "replace_dir_with_file"
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
