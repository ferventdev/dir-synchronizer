package ut

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strings"
)

// this is for cross-platformity (afraid to use syscall.ENOTEMPTY, because it seems to be unix-only)
var errDirNotEmpty = errors.New("directory not empty")

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
