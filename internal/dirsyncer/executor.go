package dirsyncer

import (
	"context"
	"dsync/internal/log"
	"dsync/internal/model"
	"dsync/internal/settings"
	"dsync/pkg/helpers/run"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
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
	entry := &(task.EntryInfo)
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
		e.log.Debug("operation has been taken into processing", log.Uint64("opID", entry.OperationPtr.ID))
	}

	wasUpdated, err := e.actualizeEntryInfo(task.Path, entry)
	if err != nil {
		return fmt.Errorf("cannot actualize entry info: %v", err)
	}
	if wasUpdated {
		e.log.Debug("entry info was actualized", log.Any("task", task))
	}

	//todo
	return nil
}

func (e *taskExecutor) actualizeEntryInfo(path string, entry *model.EntryInfo) (bool, error) {
	updated := false
	srcPath := filepath.Join(e.settings.SrcDir, path)
	copyPath := filepath.Join(e.settings.CopyDir, path)

	// 1. actualize the source file info
	srcInfo, err := os.Stat(srcPath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) { // the source file at the path does NOT exist right now
			if entry.SrcPathInfo.Exists {
				entry.SrcPathInfo = model.PathInfo{}
				updated = true
			}
		} else {
			return false, err
		}
	} else { // the source file at the path exists right now
		if entry.SrcPathInfo.Exists && entry.SrcPathInfo.FullPath != srcPath { // normally this should never happen
			e.log.Warn("entry.SrcPathInfo has inactual FullPath",
				log.String("fullPath", entry.SrcPathInfo.FullPath), log.String("srcPath", srcPath))
			entry.SrcPathInfo.FullPath = srcPath
			updated = true
		}
		if !entry.SrcPathInfo.Exists {
			entry.SrcPathInfo.Exists = true
			entry.SrcPathInfo.FullPath = srcPath
			updated = true
		}
		if entry.SrcPathInfo.IsDir != srcInfo.IsDir() {
			entry.SrcPathInfo.IsDir = srcInfo.IsDir()
			updated = true
		}
		if entry.SrcPathInfo.Size != srcInfo.Size() {
			entry.SrcPathInfo.Size = srcInfo.Size()
			updated = true
		}
		if entry.SrcPathInfo.ModTime != srcInfo.ModTime() {
			entry.SrcPathInfo.ModTime = srcInfo.ModTime()
			updated = true
		}
	}

	// 2. actualize the copy file info
	copyInfo, err := os.Stat(copyPath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) { // the copy file at the path does NOT exist right now
			if entry.CopyPathInfo.Exists {
				entry.CopyPathInfo = model.PathInfo{}
				updated = true
			}
		} else {
			return false, err
		}
	} else { // the copy file at the path exists right now
		if entry.CopyPathInfo.Exists && entry.CopyPathInfo.FullPath != copyPath { // normally this should never happen
			e.log.Warn("entry.CopyPathInfo has inactual FullPath",
				log.String("fullPath", entry.CopyPathInfo.FullPath), log.String("copyPath", copyPath))
			entry.CopyPathInfo.FullPath = copyPath
			updated = true
		}
		if !entry.CopyPathInfo.Exists {
			entry.CopyPathInfo.Exists = true
			entry.CopyPathInfo.FullPath = copyPath
			updated = true
		}
		if entry.CopyPathInfo.IsDir != copyInfo.IsDir() {
			entry.CopyPathInfo.IsDir = copyInfo.IsDir()
			updated = true
		}
		if entry.CopyPathInfo.Size != copyInfo.Size() {
			entry.CopyPathInfo.Size = copyInfo.Size()
			updated = true
		}
		if entry.CopyPathInfo.ModTime != copyInfo.ModTime() {
			entry.CopyPathInfo.ModTime = copyInfo.ModTime()
			updated = true
		}
	}

	// 3. actualize the sync operation of the task
	// todo

	if updated {
		// we need to update the entry info in the main common data structure
		e.entriesMap.SetValueByKey(path, entry)
	}
	return updated, nil
}
