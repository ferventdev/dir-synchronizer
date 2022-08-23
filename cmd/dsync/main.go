package main

import (
	"context"
	"dsync/internal/log"
	"dsync/internal/settings"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	stg, err := settings.New(os.Args[1:])
	if err != nil {
		exit(err, 2)
	}
	if err := stg.Validate(); err != nil {
		exit(err, 1)
	}

	pid := os.Getpid()
	if err := run(stg, pid); err != nil {
		exit(err, 1)
	}

	time.Sleep(100 * time.Millisecond)
	fmt.Printf("Directories Synchronizer process (PID = %d) has been stopped\n", pid)
}

func exit(err error, returnCode int) {
	fmt.Println(err)
	os.Exit(returnCode)
}

func run(stg *settings.Settings, pid int) error {
	fmt.Printf("settings: %+v\n", stg) // todo: comment this print later
	logger, err := log.New(stg.LogLevel, stg.LogToStd)
	if err != nil {
		return fmt.Errorf("can't initialize the logger: %v", err)
	}
	defer logger.Sync()
	logger.Info("logger initialized")

	if stg.PrintPID {
		// this doesn't go to the log intentionally, only to the console, may be useful in case of background running
		fmt.Println("Directories Synchronizer process started, its PID:", pid)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer stop()

	select {
	case <-ctx.Done():
		stop() // stop receiving signal notifications as soon as possible
		return nil
	case <-time.After(5 * time.Second):
		fmt.Println("Timeout expired")
	}

	// todo
	return nil
}
