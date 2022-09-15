package iout

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"strings"
)

// this is for cross-platformity (afraid to use syscall.ENOTEMPTY, because it seems to be unix-only)
var errDirNotEmpty = errors.New("directory not empty")

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
		if errors.As(err, &pErr) && pErr.Err != nil && strings.Contains(pErr.Err.Error(), errDirNotEmpty.Error()) {
			return nil
		}
		return fmt.Errorf("cannot remove entry: %w", err)
	}
	return nil
}

//CopyFile copies the entry at the source path (must be a regular file) to the specified destination.
func CopyFile(ctx context.Context, srcPath, dstPath string) error {
	if err := copyFileContents(ctx, srcPath, dstPath); err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}
	//todo
	return nil
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
