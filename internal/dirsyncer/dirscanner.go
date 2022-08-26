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
)

type dirScanner struct {
	log        log.Logger
	settings   settings.Settings
	entriesMap *dirEntriesMap
}

func newDirScanner(logger log.Logger, stg settings.Settings, eMap *dirEntriesMap) *dirScanner {
	return &dirScanner{log: logger, settings: stg, entriesMap: eMap}
}

func (d *dirScanner) scanOnce(ctx context.Context) error {
	d.log.Debug("scanOnce started")
	start := time.Now()

	// here we recursively walk through the source dir file tree and save these files' info into the map
	if err := d.walk(d.settings.SrcDir, (*EntryInfo).setSrcPathInfo); err != nil {
		return err
	}
	// here we recursively walk through the copy dir file tree and save these files' info into the map
	if err := d.walk(d.settings.CopyDir, (*EntryInfo).setCopyPathInfo); err != nil {
		return err
	}

	// todo
	d.log.Debug("scanOnce finished", log.Duration("tookTime", time.Since(start)))
	return nil
}

func (d *dirScanner) walk(root string, pathInfoSetter func(*EntryInfo, PathInfo)) error {
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

		pi := PathInfo{
			Exists:  err == nil, // means that err != fs.ErrNotExist
			IsDir:   de.IsDir(),
			Size:    info.Size(),
			ModTime: info.ModTime(),
		}
		d.entriesMap.updateValueByKey(path, func(entry *EntryInfo) { pathInfoSetter(entry, pi) })

		d.log.Debug("entry", log.String("path", path), log.Any("info", pi))
		return nil
	})
}
