package dirsyncer

import (
	"context"
	"dsync/internal/log"
	"dsync/internal/model"
	"dsync/internal/settings"
)

type task struct {
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
				if op == nil {
					//todo: create new task (with required sync operation) and enqueue it
				}
			} else {
				if op != nil && op.Status == model.OperationInProgress {
					//todo: try to cancel the task in progress, because sync is not required anymore
				}
			}

			return ctx.Err()
		},
	); err != nil {
		return err
	}

	return ctx.Err()
}
