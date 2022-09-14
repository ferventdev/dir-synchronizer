package dirsyncer

import (
	"context"
	"dsync/internal/log"
	"dsync/internal/model"
	"dsync/internal/settings"
	"dsync/pkg/helpers/iout"
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
						now := time.Now()
						task.EntryInfo.OperationPtr.FailedAt = &now
						task.EntryInfo.OperationPtr.Status = model.OpStatusFailed
						e.log.Error("failed to execute operation", log.Cause(err), log.Any("task", task))
					}
					e.entriesMap.SetValueByKey(task.Path, &(task.EntryInfo))
				}
			}
		}()
	}
}

//Stop awaits this executor's workers to finish their processing. But it doesn't wait forever - there is a timeout.
func (e *taskExecutor) Stop() {
	done := make(chan struct{})
	e.log.Debug("taskExecutor awaiting its workers to finish processing")
	go func() {
		defer close(done)
		e.wg.Wait()
	}()
	const executorStopTimeout = 5 * time.Second
	select {
	case <-done:
		e.log.Debug("taskExecutor normally stopped")
	case <-time.After(executorStopTimeout):
		e.log.Error("taskExecutor abnormally stopped on timeout (awaiting for some its worker(s) failed)")
	}
}

func (e *taskExecutor) process(ctx context.Context, task Task) error {
	entry := &(task.EntryInfo)
	op := entry.OperationPtr
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
		e.log.Debug("operation taken into processing", task.log()...)
	}

	// as long as some time passed since the task was created, we need to recheck the entry info before proceeding
	wasUpdated, err := e.actualizeEntryPathsInfo(task.Path, entry)
	if err != nil {
		return fmt.Errorf("cannot actualize entry info: %v", err)
	}

	now := time.Now()
	if wasUpdated {
		// as long as entry paths info has changed, the operation may become not actual anymore,
		// and in such case we may need to cancel or redefine it
		if entry.IsSyncRequired() {
			opKind := entry.ResolveOperationKind()
			if opKind == model.OpKindNone {
				op.CanceledAt, op.Status = &now, model.OpStatusCanceled
				e.log.Debug("entry actualized, sync not required now, operation will be canceled", task.log()...)
			} else {
				if op.Kind != opKind {
					op.Kind = opKind
					e.log.Debug("entry actualized, operation kind changed", task.log()...)
				}
			}
		} else { // sync is not required anymore, so we have to cancel the operation
			op.CanceledAt, op.Status = &now, model.OpStatusCanceled
			e.log.Debug("entry actualized, sync already achieved, operation will be canceled", task.log()...)
		}
	} else {
		// as long as entry paths info hasn't changed, we don't need to cancel or redefining the operation
		//e.log.Debug("entry info has not got any changes since task creation", log.Any("task", task))
	}

	if op.Status != model.OpStatusCanceled {
		if ctx.Err() != nil {
			op.CanceledAt, op.Status = &now, model.OpStatusCanceled
		} else {
			op.StartedAt, op.Status = &now, model.OpStatusInProgress
		}
	}
	// we need to update the entry info (with operation inside) in the main common data structure
	e.entriesMap.SetValueByKey(task.Path, entry)
	if op.Status == model.OpStatusCanceled {
		return nil // because no processing actually required
	}

	//e.log.Debug("operation execution will start now", log.Any("task", task))
	if err := e.executeOperation(ctx, entry); err != nil {
		return err
	}
	now = time.Now()
	op.CompletedAt, op.Status = &now, model.OpStatusCompleted
	e.log.Info("operation successfully executed", task.log()...)
	return nil
}

func (e *taskExecutor) actualizeEntryPathsInfo(path string, entry *model.EntryInfo) (bool, error) {
	updated := false
	srcPath, copyPath := filepath.Join(e.settings.SrcDir, path), filepath.Join(e.settings.CopyDir, path)

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
	} else if !(srcInfo.IsDir() || srcInfo.Mode().IsRegular()) { // is non-regular entry (i.e. symlink, device, etc.)
		if entry.SrcPathInfo.Exists {
			entry.SrcPathInfo = model.PathInfo{}
			updated = true
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
	} else if !(copyInfo.IsDir() || copyInfo.Mode().IsRegular()) { // is non-regular entry (i.e. symlink, device, etc.)
		if entry.CopyPathInfo.Exists {
			entry.CopyPathInfo = model.PathInfo{}
			updated = true
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

	return updated, nil
}

func (e *taskExecutor) executeOperation(ctx context.Context, entry *model.EntryInfo) error {
	opKind := entry.OperationPtr.Kind
	switch opKind {
	case model.OpKindCopyFile:
		return iout.CopyFile(ctx, entry.SrcPathInfo.FullPath, entry.CopyPathInfo.FullPath)
	case model.OpKindRemoveFile, model.OpKindRemoveDir:
		return iout.Remove(entry.CopyPathInfo.FullPath)
	case model.OpKindReplaceFile:
		// todo
	case model.OpKindReplaceDirWithFile:
		// todo
	default: // should never happen
		panic("invalid operation kind: " + opKind)
	}
	return nil
}
