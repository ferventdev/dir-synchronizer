package dirsyncer

import (
	"context"
	"dsync/internal/log"
	"dsync/internal/model"
	"dsync/internal/settings"
	"errors"
	"fmt"
	"time"
)

//DirSyncer is the main service of the app, that is responsible for the synchronization between the source and copy dirs.
type DirSyncer struct {
	log      log.Logger
	settings settings.Settings
}

func New(logger log.Logger, stg settings.Settings) *DirSyncer {
	return &DirSyncer{log: logger, settings: stg}
}

//Start returns only most critical errors (unless App runs with the -once flag) that make further work impossible,
//otherwise returns nil. However, if any inner error repeatedly happens during the execution, then such last error is
//returned after 3 consecutive occasions.
func (d *DirSyncer) Start(ctx context.Context, stop context.CancelFunc) (err error) {
	defer func() {
		if p := recover(); p != nil {
			stop()
			if perr, ok := p.(error); ok {
				err = perr
			} else {
				err = fmt.Errorf("panic: %v", p)
			}
		}
	}()

	eMap := model.NewDirEntriesMap()
	dirScanner := newDirScanner(d.log, d.settings, eMap)

	if d.settings.Once {
		err := scanAndScheduleTasks(ctx, dirScanner)
		if errors.Is(err, context.Canceled) {
			return nil
		}
		return err // no need to count inner errors in case of only one execution cycle (-once flag)
	}

	const maxConsecutiveErrors = 3
	errCount := 0
	ticker := time.NewTicker(d.settings.ScanPeriod)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			stop() // stop receiving signal notifications as soon as possible
			return nil
		case <-ticker.C:
			err := scanAndScheduleTasks(ctx, dirScanner)
			if errors.Is(err, context.Canceled) {
				return nil
			}
			if errCount++; errCount >= maxConsecutiveErrors {
				return err
			}
		}
	}
}

func scanAndScheduleTasks(ctx context.Context, dirScanner *dirScanner) error {
	if err := dirScanner.scanOnce(ctx); err != nil {
		return err
	}
	//todo: schedule tasks
	return nil
}
