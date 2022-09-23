package main

import (
	"context"
	"dsync/internal/dirsyncer"
	"dsync/internal/log"
	"dsync/internal/settings"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	stg, err := settings.New(os.Args[1:], flag.ExitOnError)
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

	// this print proves normal (externally initialized) and graceful shutdown
	fmt.Printf("Directories Synchronizer process (PID = %d) has been stopped\n", pid)
}

func exit(err error, code int) {
	fmt.Println(err)
	os.Exit(code)
}

//run returns only most critical errors that make further work impossible,
//otherwise returns nil (including the case when an external OS signal, e.g. SIGTERM, was received).
func run(stg *settings.Settings, pid int) error {
	logger, err := log.New(stg.LogLevel, stg.LogToStd)
	if err != nil {
		return fmt.Errorf("cannot initialize the logger: %v", err)
	}
	defer logger.Sync()

	logger.Debug("logger initialized")
	logger.Debug("cli args successfully parsed", log.Any("settings", *stg))

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)
	defer stop()

	if stg.PrintPID {
		// this doesn't go to the log intentionally, only to the console, may be useful in case of background running
		fmt.Println("Directories Synchronizer process started, its PID:", pid)
	}

	return dirsyncer.New(logger, *stg).Start(ctx, stop)
}
