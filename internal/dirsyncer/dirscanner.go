package dirsyncer

import (
	"context"
	"dsync/internal/log"
	"dsync/internal/settings"
	"errors"
	"io/fs"
	"path/filepath"
	"strings"
	"time"

	"go.uber.org/zap"
)

type dirScanner struct {
	log      log.Logger
	settings settings.Settings
}

func newDirScanner(logger log.Logger, stg settings.Settings) *dirScanner {
	return &dirScanner{log: logger, settings: stg}
}

func (d *dirScanner) scanOnce(ctx context.Context) error {
	d.log.Debug("scanOnce started")
	start := time.Now()

	const entriesMapMinCapacity = 10
	entriesMap := make(map[string]EntryInfo, entriesMapMinCapacity)

	if err := d.walk(d.settings.SrcDir, entriesMap); err != nil {
		return err
	}

	// todo
	d.log.Debug("scanOnce finished, it took " + time.Since(start).String())
	return nil
}

func (d *dirScanner) walk(root string, entriesMap map[string]EntryInfo) error {
	return filepath.WalkDir(root, func(fullPath string, de fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		path, err := filepath.Rel(root, fullPath)
		if err != nil {
			return err
		}
		if path == "." || (!d.settings.IncludeHidden && strings.HasPrefix(de.Name(), ".")) {
			return nil
		}
		info, err := de.Info()
		if err != nil && !errors.Is(err, fs.ErrNotExist) {
			return nil
		}

		entry := entriesMap[path]
		pi := PathInfo{
			Exists:  err == nil, // means that err != fs.ErrNotExist
			IsDir:   de.IsDir(),
			Size:    info.Size(),
			ModTime: info.ModTime(),
		}
		entry.SrcPathInfo = pi
		entriesMap[path] = entry

		d.log.Debug("entry", zap.String("path", path), zap.Any("info", pi))
		return nil
	})
}
