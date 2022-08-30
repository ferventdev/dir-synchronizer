package dirsyncer

import (
	"context"
	"dsync/internal/log"
	"dsync/internal/model"
	"dsync/internal/settings"
	"dsync/pkg/helpers/run"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
)

//dirScanner service is responsible for scanning source and copy directories for files (recursively) and
//saving their entry's info into the DirEntriesMap.
type dirScanner struct {
	log        log.Logger
	settings   settings.Settings
	entriesMap *model.DirEntriesMap
}

func newDirScanner(logger log.Logger, stg settings.Settings, eMap *model.DirEntriesMap) *dirScanner {
	return &dirScanner{log: logger, settings: stg, entriesMap: eMap}
}

func (d *dirScanner) scanOnce(parentCtx context.Context) error {
	//d.log.Debug("scanOnce started")
	//start := time.Now()
	d.entriesMap.PrepareForScan()

	ctx, cancel := context.WithCancel(parentCtx)
	defer cancel()

	errCh1 := run.AsyncWithError(func() error {
		// here we recursively walk through the source dir file tree and save these files' info into the map
		if err := d.walk(ctx, d.settings.SrcDir, (*model.EntryInfo).SetSrcPathInfo); err != nil {
			return fmt.Errorf("cannot walk through the source dir file tree: %w", err)
		}
		return nil
	})
	errCh2 := run.AsyncWithError(func() error {
		// here we recursively walk through the copy dir file tree and save these files' info into the map
		if err := d.walk(ctx, d.settings.CopyDir, (*model.EntryInfo).SetCopyPathInfo); err != nil {
			return fmt.Errorf("cannot walk through the copy dir file tree: %w", err)
		}
		return nil
	})

	for i := 0; i < 2; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case err1 := <-errCh1:
			if err1 != nil {
				return err1
			}
			errCh1 = nil // this excludes errCh1 from select on the next iteration
		case err2 := <-errCh2:
			if err2 != nil {
				return err2
			}
			errCh2 = nil // this excludes errCh2 from select on the next iteration
		}
	}

	d.entriesMap.RemoveObsolete()
	//d.log.Debug("scanOnce finished", log.Duration("tookTime", time.Since(start)))
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

		// in case of decision to sync empty dirs as well this if-clause below should be removed;
		// yet so far, I decided not to sync empty dirs, i.e. only all files (recursively) are synchronized
		if de.IsDir() {
			return nil
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

		d.log.Debug("entry scanned",
			log.String("path", path),
			log.Int64("size", pi.Size),
			log.Time("modTime", pi.ModTime),
		)
		return nil
	})
}
