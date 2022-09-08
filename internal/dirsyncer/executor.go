package dirsyncer

import (
	"context"
	"dsync/internal/log"
	"dsync/internal/model"
	"dsync/internal/settings"
	"dsync/pkg/helpers/run"
	"errors"
	"sync"
	"time"
)

var errTaskCannotGetReady = errors.New("task can't get ready for processing, so it is discarded")

//taskExecutor service is responsible for executing sync operations in order to eliminate
//the difference between the source and copy directories.
type taskExecutor struct {
	log        log.Logger
	settings   settings.Settings
	entriesMap *model.DirEntriesMap
	queue      <-chan Task
	wg         sync.WaitGroup
}

func newTaskExecutor(logger log.Logger, stg settings.Settings, eMap *model.DirEntriesMap, tasks <-chan Task) *taskExecutor {
	return &taskExecutor{log: logger, settings: stg, entriesMap: eMap, queue: tasks}
}

//Start starts this executor's workers in different goroutines.
//All workers start processing the tasks that were enqueued by the scheduler.
func (e *taskExecutor) Start(ctx context.Context) {
	for i := 0; i < e.settings.WorkersCount; i++ {
		e.wg.Add(1)
		go func() {
			defer e.wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case task, ok := <-e.queue:
					if !ok {
						return
					}
					if err := run.WithError(func() error { return e.process(ctx, task) }); err != nil {
						e.log.Error("failed to execute the task", log.Cause(err), log.Any("task", task))
					}
				}
			}
		}()
	}
}

//Stop awaits this executor's workers to finish their processing. But it doesn't wait forever - there is a timeout.
func (e *taskExecutor) Stop() {
	done := make(chan struct{})
	e.log.Debug("taskExecutor is awaiting its workers to finish processing")
	go func() {
		defer close(done)
		e.wg.Wait()
	}()
	const executorStopTimeout = 5 * time.Second
	select {
	case <-done:
		e.log.Debug("taskExecutor has been normally stopped")
	case <-time.After(executorStopTimeout):
		e.log.Error("taskExecutor has been abnormally stopped on timeout (awaiting for some its worker(s) failed)")
	}
}
func (e *taskExecutor) process(ctx context.Context, task Task) error {
	select {
	case <-ctx.Done():
		return nil
	case <-time.After(e.settings.ScanPeriod): // just in case, normally this shouldn't happen
		// but if this happens, this case will prevent a deadlock or a worker goroutine leak;
		// also, in such case, this task will become inactual, so we should discard the scheduled operation -
		// this will allow the scheduler to enqueue a new task (if there will be a need)
		e.entriesMap.UpdateValueByKey(task.Path, func(entry *model.EntryInfo) { entry.SetOperation(nil) })
		return errTaskCannotGetReady
	case <-task.ready: // usually this will be true instantly or as soon as possible
		e.log.Debug("operation has been taken into processing", log.Uint64("opID", task.EntryInfo.OperationPtr.ID))
	}

	//todo
	return nil
}
