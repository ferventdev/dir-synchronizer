package dirsyncer

import (
	"context"
	"dsync/internal/log"
	"dsync/internal/settings"
	"fmt"
	"time"
)

type DirSyncer struct {
	log      log.Logger
	settings settings.Settings
}

func New(logger log.Logger, stg settings.Settings) *DirSyncer {
	return &DirSyncer{log: logger, settings: stg}
}

//Start returns only most critical errors that make further work impossible, otherwise returns nil.
func (d *DirSyncer) Start(ctx context.Context, stop context.CancelFunc) (err error) {
	defer func() {
		if p := recover(); p != nil {
			stop()
			if perr, ok := p.(error); ok {
				err = perr
			} else {
				err = fmt.Errorf("panic! %v", p)
			}
		}
	}()

	eMap := newDirEntriesMap()
	dirScanner := newDirScanner(d.log, d.settings, eMap)

	if d.settings.Once {
		dirScanner.scanOnce(ctx)
		return nil
	}

	ticker := time.NewTicker(d.settings.ScanPeriod)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			stop() // stop receiving signal notifications as soon as possible
			return nil
		case <-ticker.C:
			dirScanner.scanOnce(ctx)
		}
	}
}
