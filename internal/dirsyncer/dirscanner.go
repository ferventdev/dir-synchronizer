package dirsyncer

import (
	"context"
	"dsync/internal/log"
	"dsync/internal/model"
	"dsync/internal/settings"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
	"time"
)

type dirScanner struct {
	log        log.Logger
	settings   settings.Settings
	entriesMap *model.DirEntriesMap
}

func newDirScanner(logger log.Logger, stg settings.Settings, eMap *model.DirEntriesMap) *dirScanner {
	return &dirScanner{log: logger, settings: stg, entriesMap: eMap}
}

func (d *dirScanner) scanOnce(ctx context.Context) error {
	d.log.Debug("scanOnce started")
	start := time.Now()

	// here we recursively walk through the source dir file tree and save these files' info into the map
	if err := d.walk(d.settings.SrcDir, (*model.EntryInfo).SetSrcPathInfo); err != nil {
		return fmt.Errorf("cannot walk through the source dir file tree: %v", err)
	}
	// here we recursively walk through the copy dir file tree and save these files' info into the map
	if err := d.walk(d.settings.CopyDir, (*model.EntryInfo).SetCopyPathInfo); err != nil {
		return fmt.Errorf("cannot walk through the copy dir file tree: %v", err)
	}

	// todo
	d.log.Debug("scanOnce finished", log.Duration("tookTime", time.Since(start)))
	return nil
}

func (d *dirScanner) walk(root string, pathInfoSetter func(*model.EntryInfo, model.PathInfo)) error {
	return filepath.WalkDir(root, func(fullPath string, de fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("cannot visit the entry %q: %v", fullPath, err)
		}
		path, err := filepath.Rel(root, fullPath)
		if err != nil {
			return fmt.Errorf("cannot get a relative path: %v", err)
		}
		if path == "." || (!d.settings.IncludeHidden && strings.HasPrefix(de.Name(), ".")) {
			return nil
		}
		info, err := de.Info()
		if err != nil && !errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("cannot fetch entry's %q info: %v", fullPath, err)
		}

		pi := model.PathInfo{
			Exists:   err == nil, // means that err != fs.ErrNotExist
			FullPath: fullPath,
			IsDir:    de.IsDir(),
			Size:     info.Size(),
			ModTime:  info.ModTime(),
		}
		d.entriesMap.UpdateValueByKey(path, func(entry *model.EntryInfo) { pathInfoSetter(entry, pi) })

		d.log.Debug("entry", log.String("path", path), log.Any("info", pi))
		return nil
	})
}
