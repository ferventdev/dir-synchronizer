package dirsyncer

import (
	"context"
	"dsync/internal/log"
	"dsync/internal/model"
	"dsync/internal/settings"
)

type task struct {
	path      string          // it's a key in DirEntriesMap
	entryInfo model.EntryInfo // it's a value in DirEntriesMap
}

//taskScheduler service is responsible for scheduling sync operations that should be done in order to eliminate
//the difference between the source and copy directories.
type taskScheduler struct {
	log        log.Logger
	settings   settings.Settings
	entriesMap *model.DirEntriesMap
	queue      chan<- task
}

func newTaskScheduler(logger log.Logger, stg settings.Settings, eMap *model.DirEntriesMap, tasks chan<- task) *taskScheduler {
	return &taskScheduler{log: logger, settings: stg, entriesMap: eMap, queue: tasks}
}

func (s *taskScheduler) scheduleOnce(ctx context.Context) error {
	var tasksToEnqueue []task
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
					tasksToEnqueue = append(tasksToEnqueue, task{path: key, entryInfo: entry})
				}
				return ctx.Err()
			}

			// if sync not required while it's already in progress we should cancel it
			if op != nil && op.Status == model.OperationInProgress {
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
		op := model.NewOperation(t.entryInfo.ChooseOperationKind())
		t.entryInfo.OperationPtr = op
		select {
		case <-childCtx.Done():
			return childCtx.Err()
		case s.queue <- t: // enqueue new task with a scheduled operation inside to the queue of tasks
			s.entriesMap.UpdateValueByKey(t.path, func(entry *model.EntryInfo) { entry.SetOperation(op) })
		}
	}

	return nil
}
