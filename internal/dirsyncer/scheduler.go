package dirsyncer

import (
	"context"
	"dsync/internal/log"
	"dsync/internal/model"
	"dsync/internal/settings"
	"errors"
)

//Task is a sync task. taskScheduler puts it into its queue.
type Task struct {
	Path      string          `json:"path"`  // it's a key in DirEntriesMap
	EntryInfo model.EntryInfo `json:"entry"` // it's a value in DirEntriesMap
	ready     chan struct{}   // is task ready to be processed by a worker
}

func NewTask(path string, ei model.EntryInfo) Task {
	return Task{Path: path, EntryInfo: ei, ready: make(chan struct{})}
}

//setReady tells a worker (that will process this task) that this task is ready for processing.
func (t *Task) setReady() {
	close(t.ready)
}

func (t *Task) log() []log.Field {
	opPtr := t.EntryInfo.OperationPtr
	var opField log.Field
	if opPtr == nil {
		opField = log.NilField("operation")
	} else {
		opField = log.Any("operation", *opPtr)
	}
	return []log.Field{log.String("path", t.Path), opField}
}

//taskScheduler service is responsible for scheduling sync operations that should be done in order to eliminate
//the difference between the source and copy directories.
type taskScheduler struct {
	log        log.Logger
	settings   settings.Settings
	entriesMap *model.DirEntriesMap
	queue      chan<- Task // only taskScheduler can write to this channel
}

func newTaskScheduler(logger log.Logger, stg settings.Settings, eMap *model.DirEntriesMap, tasks chan<- Task) *taskScheduler {
	return &taskScheduler{log: logger, settings: stg, entriesMap: eMap, queue: tasks}
}

func (s *taskScheduler) scheduleOnce(ctx context.Context) error {
	var tasksToEnqueue []Task
	if err := s.entriesMap.ForEach(
		func(key string, eMap map[string]model.EntryInfo) error {
			entry := eMap[key] // entry may have zero value
			op := entry.OperationPtr

			// if the operation took place earlier, and it's over now, we should clear it
			if op.IsNotNilAndOver() {
				entry.OperationPtr = nil
				eMap[key] = entry
				return ctx.Err()
			}

			if entry.IsSyncRequired() {
				// here we create new sync task
				if op == nil {
					tasksToEnqueue = append(tasksToEnqueue, NewTask(key, entry))
				}
				return ctx.Err()
			}

			// if sync not required, but the sync operation is already in progress - we should cancel it
			if op != nil && op.Status == model.OpStatusInProgress {
				if op.CancelFn != nil {
					op.CancelFn()
				} else {
					s.log.Error("operation's cancel function is not set", log.Uint64("opID", op.ID))
				}
			}

			return ctx.Err()
		},
	); err != nil {
		return err
	}

	// we don't want to be blocked forever if s.queue is full
	timeout := s.settings.ScanPeriod
	if s.settings.Once {
		timeout = s.settings.ScanPeriod * 1000 // if this func is run only once, is makes sense to wait much longer
	}
	childCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	for _, t := range tasksToEnqueue {
		t := t
		opKind := t.EntryInfo.ResolveOperationKind()
		if opKind == model.OpKindCopyDir {
			// do not copy empty dir, and non-empty dir will be copied automatically on the file copying
			continue
		}
		if opKind == model.OpKindNone {
			s.log.Warn("sync operation kind cannot be properly resolved", t.log()...)
			continue
		}
		op := model.NewOperation(opKind)
		t.EntryInfo.OperationPtr = op
		select {
		case <-childCtx.Done():
			// if timeout is exceeded we don't consider that as an error,
			// because we'll be back to this method on the next sync cycle
			if !s.settings.Once && errors.Is(childCtx.Err(), context.DeadlineExceeded) {
				return nil
			}
			return childCtx.Err()
		case s.queue <- t: // enqueue new task with a scheduled operation inside to the queue of tasks
			s.entriesMap.UpdateValueByKey(t.Path, func(entry *model.EntryInfo) { entry.SetOperation(op) })
			s.log.Debug("new task enqueued by scheduler", t.log()...)
			t.setReady() // tell the worker that task is ready for processing
		}
	}

	return nil
}
