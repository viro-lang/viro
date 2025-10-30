package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	setupSignalHandler()

	cfg := NewConfig()
	if err := cfg.LoadFromEnv(); err != nil {
		fmt.Fprintf(os.Stderr, "Configuration error: %v\n", err)
		os.Exit(ExitUsage)
	}

	if err := cfg.LoadFromFlags(); err != nil {
		fmt.Fprintf(os.Stderr, "Configuration error: %v\n", err)
		os.Exit(ExitUsage)
	}

	if err := cfg.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(ExitUsage)
	}

	mode, err := detectMode(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(ExitUsage)
	}

	var exitCode int
	switch mode {
	case ModeREPL:
		exitCode = runREPL(cfg)
	case ModeScript:
		exitCode = runScript(cfg)
	case ModeEval:
		exitCode = runEval(cfg)
	case ModeCheck:
		exitCode = runCheck(cfg)
	case ModeVersion:
		printVersion()
		exitCode = ExitSuccess
	case ModeHelp:
		printHelp()
		exitCode = ExitSuccess
	default:
		fmt.Fprintf(os.Stderr, "Unknown mode: %v\n", mode)
		exitCode = ExitUsage
	}

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

func runREPL(cfg *Config) int {
	if cfg.AllowInsecureTLS {
		fmt.Fprintf(os.Stderr, "WARNING: TLS certificate verification disabled globally. Use with caution.\n")
	}

	repl, err := NewREPL()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing REPL: %v\n", err)
		return ExitError
	}

	if err := repl.Run(); err != nil {
		return handleError(err)
	}

	return ExitSuccess
}
