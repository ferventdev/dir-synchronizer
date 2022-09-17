package iout

import (
	"context"
	"dsync/pkg/helpers/ut"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"time"
)

var (
	// this is for cross-platformity (afraid to use syscall.ENOTEMPTY, because it seems to be unix-only)
	errDirNotEmpty = errors.New("directory not empty")
	// because stdlib's fsys.errNotDir is not exported
	errNotDir = errors.New("not a directory")
)

func IsErrNotDir(err error) bool {
	return ut.IsSameError(err, errNotDir)
}

//readerWithContext allows to perform a cancellable read operation.
type readerWithContext struct {
	ctx context.Context
	r   io.Reader
}

func newReaderWithContext(ctx context.Context, r io.Reader) io.Reader {
	return &readerWithContext{ctx: ctx, r: r}
}

func (r *readerWithContext) Read(p []byte) (int, error) {
	select {
	case <-r.ctx.Done():
		return 0, r.ctx.Err()
	default:
		return r.r.Read(p)
	}
}

//Remove removes a file or an empty directory. It silently ignores non-empty directory.
func Remove(path string) error {
	if err := os.Remove(path); err != nil {
		var pErr *fs.PathError
		if errors.As(err, &pErr) && ut.IsSameError(pErr.Err, errDirNotEmpty) {
			return nil
		}
		return fmt.Errorf("cannot remove entry: %w", err)
	}
	return nil
}

//CopyFile copies the entry at the source path (must be a regular file) to the specified destination.
//It sets for the copied file the same modTime as the source file modTime.
func CopyFile(ctx context.Context, srcPath, dstPath string, srcModTime time.Time) error {
	if err := EnsureDirExists(ctx, filepath.Dir(dstPath)); err != nil {
		return err
	}
	return ReplaceFile(ctx, srcPath, dstPath, srcModTime)
}

func EnsureDirExists(ctx context.Context, dirPath string) error {
	if err := os.MkdirAll(dirPath, os.ModePerm); err != nil {
		if !IsErrNotDir(err) {
			return fmt.Errorf("cannot make dir: %w", err)
		}
		// 'not a directory' error may happen due to a rare collision - when there's a file at dirPath,
		// and it has not yet been removed (i.e. 'remove_file' task was scheduled and another worker may even
		// start its execution, but has not yet completed), so we just retry mkdir in a moment
		const retryTime = 20 * time.Millisecond
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(retryTime):
		}
		if err := os.MkdirAll(dirPath, os.ModePerm); err != nil {
			return fmt.Errorf("cannot make dir: %w", err)
		}
	}
	return nil
}

//ReplaceFile truncates the file at dstPath (or creates it, if absent) and writes the source file's content into it.
//It sets for the "replaced" file the same modTime as the source file modTime.
func ReplaceFile(ctx context.Context, srcPath string, dstPath string, srcModTime time.Time) error {
	if err := copyFileContents(ctx, srcPath, dstPath); err != nil {
		return fmt.Errorf("cannot copy file contents: %w", err)
	}
	if err := os.Chtimes(dstPath, time.Now(), srcModTime); err != nil {
		return fmt.Errorf("cannot set file modification time: %w", err)
	}
	return nil
}

//ReplaceDirWithFile removes an empty directory dstPath and, if succeeded, copies the source file to its place.
//Method does nothing in case of non-empty dstPath.
func ReplaceDirWithFile(ctx context.Context, srcPath string, dstPath string, srcModTime time.Time) error {
	if err := os.Remove(dstPath); err != nil {
		var pErr *fs.PathError
		if errors.As(err, &pErr) && ut.IsSameError(pErr.Err, errDirNotEmpty) {
			return nil
		}
		return fmt.Errorf("cannot remove dir: %w", err)
	}
	return ReplaceFile(ctx, srcPath, dstPath, srcModTime)
}

func copyFileContents(ctx context.Context, src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("cannot open file: %w", err)
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("cannot create file: %w", err)
	}
	defer out.Close()

	if _, err = io.Copy(out, newReaderWithContext(ctx, in)); err != nil {
		return fmt.Errorf("cannot read/write file content: %w", err)
	}
	return out.Sync()
}
