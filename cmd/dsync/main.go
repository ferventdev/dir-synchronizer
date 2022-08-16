package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"time"
)

type Settings struct {
	srcDir   string
	copyDir  string
	logToStd bool
	once     bool
}

func main() {
	fmt.Println("start")
	settings, err := parseCommandArgs()
	if err != nil {
		fmt.Println(err)
		os.Exit(2)
	}
	fmt.Println("settings:", settings)

	time.Sleep(1 * time.Second)
	fmt.Println("finish")
}

func parseCommandArgs() (*Settings, error) {
	stg := new(Settings)

	flagSet := flag.NewFlagSet("Directories Synchronizer CLI", flag.ExitOnError)
	flagSet.BoolVar(&stg.logToStd, "log2std", false,
		"if true, then logs are written to the console, otherwise - to the text file")
	flagSet.BoolVar(&stg.once, "once", false,
		"if true, then directories are synchronized only once (i.e. the program has finite execution), "+
			"otherwise - the process is started and lasts indefinitely (until interruption)")
	flagSet.Parse(os.Args[1:])

	if flagSet.NArg() < 2 {
		return nil, errors.New("at least two arguments (for the directories for synchronization) must present")
	}
	stg.srcDir, stg.copyDir = flagSet.Arg(0), flagSet.Arg(1)
	if stg.srcDir == stg.copyDir {
		return nil, errors.New("the directories for synchronization cannot be the same")
	}
	return stg, nil
}
