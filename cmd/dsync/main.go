package main

import (
	"dsync/internal/log"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Settings struct {
	srcDir   string
	copyDir  string
	logLevel log.Level
	logToStd bool
	once     bool
}

func main() {
	fmt.Println("start")
	settings, err := parseCommandArgs(os.Args[1:])
	if err != nil {
		fmt.Println(err)
		os.Exit(2)
	}
	if err := validateSettings(settings); err != nil {
		exit(err)
	}
	if err := run(settings); err != nil {
		exit(err)
	}
}

func exit(err error) {
	fmt.Println(err)
	os.Exit(1)
}

func run(stg *Settings) error {
	fmt.Println("settings:", stg)
	logger, err := log.New(stg.logLevel, stg.logToStd)
	if err != nil {
		return fmt.Errorf("can't initialize the logger: %v", err)
	}
	defer logger.Sync()
	logger.Info("logger initialized")

	// todo
	time.Sleep(1 * time.Second)

	logger.Info("finish")
	return nil
}

func parseCommandArgs(args []string) (*Settings, error) {
	stg := new(Settings)

	flagSet := flag.NewFlagSet("Directories Synchronizer CLI", flag.ExitOnError)
	flagSet.BoolVar(&stg.logToStd, "log2std", false,
		"if true, then logs are written to the console, otherwise - to the text file")
	flagSet.BoolVar(&stg.once, "once", false,
		"if true, then directories are synchronized only once (i.e. the program has finite execution), "+
			"otherwise - the process is started and lasts indefinitely (until interruption)")
	var level string
	flagSet.StringVar(&level, "loglvl", log.InfoLevel,
		fmt.Sprintf("level of logging, permitted values are: %v, %v, %v, %v",
			log.DebugLevel, log.InfoLevel, log.WarnLevel, log.ErrorLevel),
	)

	flagSet.Parse(args)

	if flagSet.NArg() < 2 {
		return nil, errors.New("at least two arguments (for the directories for synchronization) must present")
	}

	var err error = nil
	if stg.srcDir, err = filepath.Abs(flagSet.Arg(0)); err != nil {
		return nil, fmt.Errorf("path %q cannot be converted to absolute: %v", flagSet.Arg(0), err)
	}
	if stg.copyDir, err = filepath.Abs(flagSet.Arg(1)); err != nil {
		return nil, fmt.Errorf("path %q cannot be converted to absolute: %v", flagSet.Arg(1), err)
	}
	if stg.srcDir == stg.copyDir {
		return nil, errors.New("the directories for synchronization cannot be the same")
	}
	if !log.Level(level).IsValid() {
		return nil, fmt.Errorf("logging level %q does not exist", level)
	}
	stg.logLevel = log.Level(strings.ToLower(level))

	return stg, nil
}

func validateSettings(stg *Settings) error {
	if err := validateDirectoryPath(stg.srcDir); err != nil {
		return fmt.Errorf("the first (source) directory is invalid: %v", err)
	}
	if err := validateDirectoryPath(stg.copyDir); err != nil {
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
