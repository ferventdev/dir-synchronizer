package dirsyncer

import (
	"context"
	logmock "dsync/generated/mocks"
	"dsync/internal/log"
	"dsync/internal/model"
	"dsync/internal/settings"
	"dsync/pkg/helpers/iout"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestDirSyncerByRunningOnce(t *testing.T) {
	requires := require.New(t)
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	// 1. arrange
	_ = os.Chdir("testdata")
	wd, _ := os.Getwd()
	srcDir := filepath.Join(wd, "src")
	copyDir, err := os.MkdirTemp(wd, "copy")
	requires.NoError(err)
	defer os.RemoveAll(copyDir)

	prepareCopyDir(requires, copyDir, srcDir)

	loggerMock := getMockLogger(mockCtrl, gomock.Any())
	stg := settings.Settings{
		SrcDir:           srcDir,
		CopyDir:          copyDir,
		ScanPeriod:       2 * time.Second,
		IncludeHidden:    false,
		IncludeEmptyDirs: true,
		LogLevel:         log.DebugLevel,
		LogToStd:         true,
		Once:             true,
		WorkersCount:     1,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 2. act
	err = New(loggerMock, stg).Start(ctx, cancel)

	// 3. assert
	requires.NoError(err)

	dirEntriesMap := model.NewDirEntriesMap()
	requires.NoError(newDirScanner(loggerMock, stg, dirEntriesMap).scanOnce(context.Background()))

	err = dirEntriesMap.ForEach(func(key string, eMap map[string]model.EntryInfo) error {
		entry := eMap[key]
		requires.False(entry.IsSyncRequired())
		return nil
	})
	requires.NoError(err)
}

func TestDirSyncerByRunningRegularly(t *testing.T) {
	requires := require.New(t)
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	// 1. arrange
	_ = os.Chdir("testdata")
	wd, _ := os.Getwd()
	srcDir := filepath.Join(wd, "src")
	copyDir, err := os.MkdirTemp(wd, "copy")
	requires.NoError(err)
	defer os.RemoveAll(copyDir)

	prepareCopyDir(requires, copyDir, srcDir)

	loggerMock := getMockLogger(mockCtrl, gomock.Any())
	scanPeriod := time.Second
	stg := settings.Settings{
		SrcDir:           srcDir,
		CopyDir:          copyDir,
		ScanPeriod:       scanPeriod,
		IncludeHidden:    true,
		IncludeEmptyDirs: true,
		LogLevel:         log.DebugLevel,
		LogToStd:         true,
		Once:             false,
		WorkersCount:     4,
	}

	timeout := 2*scanPeriod + 200*time.Millisecond
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// 2. act
	err = New(loggerMock, stg).Start(ctx, cancel)

	// 3. assert
	requires.NoError(err)

	dirEntriesMap := model.NewDirEntriesMap()
	requires.NoError(newDirScanner(loggerMock, stg, dirEntriesMap).scanOnce(context.Background()))

	err = dirEntriesMap.ForEach(func(key string, eMap map[string]model.EntryInfo) error {
		entry := eMap[key]
		requires.False(entry.IsSyncRequired())
		return nil
	})
	requires.NoError(err)
}

func prepareCopyDir(req *require.Assertions, copyDir string, srcDir string) {
	createDir(req, copyDir, "subdir1/subdir2")
	createDir(req, copyDir, "subdir1/wrong_dir")    // this dir has to be removed
	createDir(req, copyDir, "subdir1/missing_file") // this dir has to be replaced with the same named file
	copyFileIntoDir(req, srcDir, "old_file.txt", copyDir)
	// this file has to be replaced due to the different mod time
	copyFileIntoDirWithChangedModTime(req, srcDir, "subdir1/old_file.txt", copyDir)
	// the wrong file has to be removed, because such name is absent in the src dir
	copyFile(req, copyDir, "subdir1/old_file.txt", filepath.Join(copyDir, "subdir1/wrong_file.txt"), false)
	copyFileIntoDir(req, srcDir, "subdir1/subdir2/old_file.txt", copyDir)
}

func getMockLogger(mockCtrl *gomock.Controller, any gomock.Matcher) *logmock.MockLogger {
	loggerMock := logmock.NewMockLogger(mockCtrl)
	loggerMock.EXPECT().Debug(any).AnyTimes()
	loggerMock.EXPECT().Debug(any, any).AnyTimes()
	loggerMock.EXPECT().Debug(any, any, any).AnyTimes()
	loggerMock.EXPECT().Debug(any, any, any, any).AnyTimes()
	loggerMock.EXPECT().Info(any, any, any).AnyTimes()
	loggerMock.EXPECT().Info(any, any, any, any).AnyTimes()
	loggerMock.EXPECT().Error(any, any, any).AnyTimes()
	return loggerMock
}

func createDir(req *require.Assertions, baseDir string, relDirPath string) {
	req.NoError(iout.EnsureDirExists(context.Background(), filepath.Join(baseDir, relDirPath)))
}

func copyFileIntoDir(req *require.Assertions, srcBaseDir string, fileRelPath string, copyBaseDir string) {
	copyFile(req, srcBaseDir, fileRelPath, filepath.Join(copyBaseDir, fileRelPath), true)
}

func copyFileIntoDirWithChangedModTime(
	req *require.Assertions, srcBaseDir string, fileRelPath string, copyBaseDir string,
) {
	copyFile(req, srcBaseDir, fileRelPath, filepath.Join(copyBaseDir, fileRelPath), false)
}

func copyFile(req *require.Assertions, srcBaseDir string, srcFileRelPath string, copyAbsPath string, sameModTime bool) {
	srcAbsPath := filepath.Join(srcBaseDir, srcFileRelPath)
	fileStat, err := os.Stat(srcAbsPath)
	req.NoError(err)
	modTime := fileStat.ModTime()
	if !sameModTime {
		modTime = modTime.Add(24 * time.Hour)
	}
	req.NoError(iout.CopyFile(context.Background(), srcAbsPath, copyAbsPath, modTime))
}
