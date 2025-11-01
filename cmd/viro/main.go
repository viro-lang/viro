package main

import (
	"os"
	"os/signal"
	"syscall"
)

func main() {
	setupSignalHandler()

	ctx := &RuntimeContext{
		Args:   os.Args[1:],
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}

	exitCode := Run(ctx)
	os.Exit(exitCode)
}

func setupSignalHandler() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		os.Exit(ExitInterrupt)
	}()
}
