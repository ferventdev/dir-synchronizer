package iout

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// also implicitly tests successful cases for EnsureDirExists() and ReplaceFile()
func TestCopyFile(t *testing.T) {
	requires := require.New(t)

	// 1. arrange
	wd, err := os.Getwd()
	requires.NoError(err)

	srcDir := filepath.Join(wd, "testdata/src")
	copyDir, err := os.MkdirTemp(filepath.Join(wd, "testdata"), "copy")
	requires.NoError(err)
	defer os.RemoveAll(copyDir)

	fileName := "some_file.txt"
	srcAbsPath := filepath.Join(srcDir, fileName)
	srcFileInfo, err := os.Stat(srcAbsPath)
	requires.NoError(err)

	// 2. act
	destAbsPath := filepath.Join(copyDir, fileName)
	err = CopyFile(context.Background(), srcAbsPath, destAbsPath, srcFileInfo.ModTime())

	// 3. assert that the original and the copied files have same names (in their dirs), size and modTime
	requires.NoError(err)
	copiedFileInfo, err := os.Stat(destAbsPath)
	requires.NoError(err)
	requires.Equal(fileName, copiedFileInfo.Name())
	requires.False(copiedFileInfo.IsDir())
	requires.Equal(srcFileInfo.IsDir(), copiedFileInfo.IsDir()) // both are not dirs
	requires.Equal(srcFileInfo.Size(), copiedFileInfo.Size())
	requires.Equal(srcFileInfo.ModTime(), copiedFileInfo.ModTime())
}

func TestEnsureDirExistsCannotMakeDir(t *testing.T) {
	requires := require.New(t)
	wd, err := os.Getwd()
	requires.NoError(err)

	err = EnsureDirExists(context.Background(), filepath.Join(wd, "testdata/src/some_file.txt")) // file is not a directory!

	requires.Error(err)
	requires.ErrorContains(err, "cannot make dir")
	requires.True(IsErrNotDir(err))
}

func TestRemoveIgnoresNonEmptyDir(t *testing.T) {
	requires := require.New(t)
	wd, err := os.Getwd()
	requires.NoError(err)
	nonEmptyDirPath := filepath.Join(wd, "testdata/src")

	err = Remove(nonEmptyDirPath) // dir is not empty, so won't be removed (no error)

	requires.NoError(err)

	// check that dir still exists (was not removed)
	dirInfo, err := os.Stat(nonEmptyDirPath)
	requires.NoError(err)
	requires.True(dirInfo.IsDir())
}

func BenchmarkCopyFile(b *testing.B) {
	b.StopTimer()

	// 1. arrange
	requires := require.New(b)
	wd, err := os.Getwd()
	requires.NoError(err)
	srcDir := filepath.Join(wd, "testdata/src")
	copyDir, err := os.MkdirTemp(filepath.Join(wd, "testdata"), "copy")
	requires.NoError(err)
	defer os.RemoveAll(copyDir)

	fileName := "some_file.txt"
	srcAbsPath := filepath.Join(srcDir, fileName)
	destAbsPath := filepath.Join(copyDir, fileName)
	srcFileInfo, err := os.Stat(srcAbsPath)
	requires.NoError(err)
	ctx := context.Background()

	// 2. bench
	b.ReportAllocs()
	b.ResetTimer()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		_ = CopyFile(ctx, srcAbsPath, destAbsPath, srcFileInfo.ModTime())

		// remove the copied file so that we can copy it again on the next iteration
		b.StopTimer()
		requires.NoError(Remove(destAbsPath))
		b.StartTimer()
	}
}
