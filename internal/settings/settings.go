package settings

import (
	"dsync/internal/log"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Settings struct {
	SrcDir   string
	CopyDir  string
	LogLevel log.Level
	LogToStd bool
	Once     bool
}

func New(commandArgs []string) (*Settings, error) {
	stg := new(Settings)
	flagSet := flag.NewFlagSet("Directories Synchronizer CLI", flag.ExitOnError)

	flagSet.BoolVar(&stg.LogToStd, "log2std", false,
		"if true, then logs are written to the console, otherwise - to the text file")
	flagSet.BoolVar(&stg.Once, "once", false,
		"if true, then directories are synchronized only once (i.e. the program has finite execution), "+
			"otherwise - the process is started and lasts indefinitely (until interruption)")
	var level string
	flagSet.StringVar(&level, "loglvl", log.InfoLevel,
		fmt.Sprintf("level of logging, permitted values are: %v, %v, %v, %v",
			log.DebugLevel, log.InfoLevel, log.WarnLevel, log.ErrorLevel),
	)

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
