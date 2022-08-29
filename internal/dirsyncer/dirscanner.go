package dirsyncer

import (
	"context"
	"dsync/internal/log"
	"dsync/internal/model"
	"dsync/internal/settings"
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

	d.entriesMap.PrepareForScan()

	// here we recursively walk through the source dir file tree and save these files' info into the map
	if err := d.walk(ctx, d.settings.SrcDir, (*model.EntryInfo).SetSrcPathInfo); err != nil {
		return fmt.Errorf("cannot walk through the source dir file tree: %w", err)
	}
	// here we recursively walk through the copy dir file tree and save these files' info into the map
	if err := d.walk(ctx, d.settings.CopyDir, (*model.EntryInfo).SetCopyPathInfo); err != nil {
		return fmt.Errorf("cannot walk through the copy dir file tree: %w", err)
	}

	d.entriesMap.RemoveObsolete()

	d.log.Debug("scanOnce finished", log.Duration("tookTime", time.Since(start)))
	return ctx.Err()
}

func (d *dirScanner) walk(ctx context.Context, root string, pathInfoSetter func(*model.EntryInfo, model.PathInfo)) error {
	return filepath.WalkDir(root, func(fullPath string, de fs.DirEntry, err error) error {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return ctxErr
		}
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
		if err != nil {
			return fmt.Errorf("cannot fetch entry's %q info: %v", fullPath, err)
		}

		pi := model.PathInfo{
			Exists:   true,
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
