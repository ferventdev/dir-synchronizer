package main

import (
	"dsync/internal/log"
	"dsync/internal/settings"
	"fmt"
	"os"
	"time"
)

func main() {
	fmt.Println("start")
	stg, err := settings.New(os.Args[1:])
	if err != nil {
		exit(err, 2)
	}
	if err := stg.Validate(); err != nil {
		exit(err, 1)
	}

	if err := run(stg); err != nil {
		exit(err, 1)
	}
}

func exit(err error, returnCode int) {
	fmt.Println(err)
	os.Exit(returnCode)
}

func run(stg *settings.Settings) error {
	fmt.Printf("settings: %+v\n", stg)
	logger, err := log.New(stg.LogLevel, stg.LogToStd)
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
