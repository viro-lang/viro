package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/marcin-radoszewski/viro/internal/repl"
)

func main() {
	setupSignalHandler()

	cfg, err := loadConfiguration()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Configuration error: %v\n", err)
		os.Exit(ExitUsage)
	}

	exitCode := executeMode(cfg)
	os.Exit(exitCode)
}

func loadConfiguration() (*Config, error) {
	cfg := NewConfig()
	if err := cfg.LoadFromEnv(); err != nil {
		return nil, err
	}
	if err := cfg.LoadFromFlags(); err != nil {
		return nil, err
	}
	if err := cfg.ApplyDefaults(); err != nil {
		return nil, err
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

func executeMode(cfg *Config) int {
	mode, err := cfg.DetectMode()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return ExitUsage
	}

	switch mode {
	case ModeREPL:
		return runREPL(cfg)
	case ModeScript, ModeEval, ModeCheck:
		return runExecution(cfg, mode)
	case ModeVersion:
		printVersion()
		return ExitSuccess
	case ModeHelp:
		printHelp()
		return ExitSuccess
	default:
		fmt.Fprintf(os.Stderr, "Unknown mode: %v\n", mode)
		return ExitUsage
	}
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

	r, err := repl.NewREPL(cfg.Args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing REPL: %v\n", err)
		return ExitError
	}

	if err := r.Run(); err != nil {
		return handleError(err)
	}

	return ExitSuccess
}
