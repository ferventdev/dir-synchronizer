package dirsyncer

import (
	"context"
	"dsync/internal/log"
	"time"
)

type dirScanner struct {
	log log.Logger
}

func newDirScanner(logger log.Logger) *dirScanner {
	return &dirScanner{log: logger}
}

func (d *dirScanner) scanOnce(ctx context.Context) {
	d.log.Debug("scanOnce started")
	start := time.Now()
	time.Sleep(100 * time.Millisecond) // todo
	d.log.Debug("scanOnce finished, it took " + time.Since(start).String())
}
