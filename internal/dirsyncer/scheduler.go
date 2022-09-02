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
					tasksToEnqueue = append(tasksToEnqueue, Task{Path: key, EntryInfo: entry})
				}
				return ctx.Err()
			}

			// if sync not required while it's already in progress we should cancel it
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
		opKind := t.EntryInfo.ResolveOperationKind()
		if opKind == model.OpKindNone {
			s.log.Error("sync operation kind cannot be properly resolved", log.Any("task", t))
			continue
		}
		op := model.NewOperation(opKind)
		t.EntryInfo.OperationPtr = op
		select {
		case <-childCtx.Done():
			if errors.Is(childCtx.Err(), context.Canceled) { // may happen only at the shutdown
				close(s.queue)
			}
			return childCtx.Err()
		case s.queue <- t: // enqueue new task with a scheduled operation inside to the queue of tasks
			s.entriesMap.UpdateValueByKey(t.Path, func(entry *model.EntryInfo) { entry.SetOperation(op) })
			s.log.Debug("new sync task enqueued by scheduler", log.Any("task", t))
		}
	}

	return nil
}
