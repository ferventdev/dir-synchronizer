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
	// todo
	return ctx.Err()
}
