package settings

import (
	"dsync/internal/log"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const (
	minScanPeriod   = time.Second
	maxScanPeriod   = 10 * time.Second
	minWorkersCount = 1
	maxWorkersCount = 1000
)

type Settings struct {
	SrcDir           string
	CopyDir          string
	ScanPeriod       time.Duration
	IncludeHidden    bool
	IncludeEmptyDirs bool
	LogLevel         log.Level
	LogToStd         bool
	Once             bool
	PrintPID         bool
	WorkersCount     int
}

func New(commandArgs []string, handling flag.ErrorHandling) (*Settings, error) {
	stg := new(Settings)
	flagSet := flag.NewFlagSet("Directories Synchronizer CLI", handling)

	flagSet.BoolVar(&stg.IncludeHidden, "hidden", false,
		"if true, then hidden files (that start with dot) are included in synchronization")
	flagSet.BoolVar(&stg.IncludeEmptyDirs, "copydirs", false,
		"if true, then empty directories are included in synchronization as well (non-empty are synced anyway)")
	flagSet.BoolVar(&stg.LogToStd, "log2std", false,
		"if true, then logs are written to the console, otherwise - to the text file")
	flagSet.BoolVar(&stg.Once, "once", false,
		"if true, then directories are synchronized only once (i.e. the program has finite execution), "+
			"otherwise - the process is started and lasts indefinitely (until interruption)")
	flagSet.BoolVar(&stg.PrintPID, "pid", false,
		"if true, then the PID is printed at the startup (may be useful in case of background running)")
	var level string
	flagSet.StringVar(&level, "loglvl", log.InfoLevel,
		fmt.Sprintf("level of logging, permitted values are: %v, %v, %v, %v",
			log.DebugLevel, log.InfoLevel, log.WarnLevel, log.ErrorLevel),
	)
	flagSet.DurationVar(&stg.ScanPeriod, "scanperiod", time.Second,
		fmt.Sprintf("period of directories scanning, must be a value between %v and %v", minScanPeriod, maxScanPeriod))
	flagSet.IntVar(&stg.WorkersCount, "workers", runtime.NumCPU(),
		fmt.Sprintf("the number of workers that will be started to execute all sync operations, "+
			"must be a value between %d and %d", minWorkersCount, maxWorkersCount))

	flagSet.Parse(commandArgs)

	if flagSet.NArg() < 2 {
		return nil, errors.New("at least two arguments (for the directories for synchronization) must present")
	}

	var err error = nil
	if stg.SrcDir, err = filepath.Abs(flagSet.Arg(0)); err != nil {
		return nil, fmt.Errorf("path %q cannot be converted to absolute: %v", flagSet.Arg(0), err)
	}
	if stg.CopyDir, err = filepath.Abs(flagSet.Arg(1)); err != nil {
		return nil, fmt.Errorf("path %q cannot be converted to absolute: %v", flagSet.Arg(1), err)
	}
	if stg.SrcDir == stg.CopyDir {
		return nil, errors.New("the directories for synchronization cannot be the same")
	}
	if !log.Level(level).IsValid() {
		return nil, fmt.Errorf("logging level %q does not exist", level)
	}
	stg.LogLevel = log.Level(strings.ToLower(level))

	return stg, nil
}

func (stg *Settings) Validate() error {
	if err := validateDirectoryPath(stg.SrcDir); err != nil {
		return fmt.Errorf("the first (source) directory is invalid: %v", err)
	}
	if err := validateDirectoryPath(stg.CopyDir); err != nil {
		return fmt.Errorf("the second (copy) directory is invalid: %v", err)
	}
	if stg.ScanPeriod < minScanPeriod || stg.ScanPeriod > maxScanPeriod {
		return fmt.Errorf("period of directories scanning must be a value between %v and %v, while it is %v",
			minScanPeriod, maxScanPeriod, stg.ScanPeriod)
	}
	if stg.WorkersCount < minWorkersCount || stg.WorkersCount > maxWorkersCount {
		return fmt.Errorf("number of workers must be a value between %d and %d, while it is %d",
			minWorkersCount, maxWorkersCount, stg.WorkersCount)
	}
	return nil
}

func validateDirectoryPath(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("path %q is not a directory path", path)
	}
	return nil
}
