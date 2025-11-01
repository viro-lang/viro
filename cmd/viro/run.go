package main

import (
	"fmt"

	"github.com/marcin-radoszewski/viro/internal/api"
	"github.com/marcin-radoszewski/viro/internal/config"
	"github.com/marcin-radoszewski/viro/internal/repl"
)

func Run(ctx *api.RuntimeContext) int {
	cfg, err := loadConfigurationWithContext(ctx)
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "Configuration error: %v\n", err)
		return api.ExitUsage
	}

	return executeModeWithContext(cfg, ctx)
}

func loadConfigurationWithContext(ctx *api.RuntimeContext) (*config.Config, error) {
	cfg := config.NewConfig()
	if err := cfg.LoadFromEnv(); err != nil {
		return nil, err
	}
	if err := cfg.LoadFromFlagsWithArgs(ctx.Args); err != nil {
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

func executeModeWithContext(cfg *config.Config, ctx *api.RuntimeContext) int {
	mode, err := cfg.DetectMode()
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "Error: %v\n", err)
		return api.ExitUsage
	}

	switch mode {
	case config.ModeREPL:
		return runREPLWithContext(cfg, ctx)
	case config.ModeScript, config.ModeEval, config.ModeCheck:
		return api.RunExecutionWithContext(cfg, mode, ctx)
	case config.ModeVersion:
		fmt.Fprintf(ctx.Stdout, "%s\n", getVersionString())
		return api.ExitSuccess
	case config.ModeHelp:
		fmt.Fprintf(ctx.Stdout, "%s", getHelpText())
		return api.ExitSuccess
	default:
		fmt.Fprintf(ctx.Stderr, "Unknown mode: %v\n", mode)
		return api.ExitUsage
	}
}

func runREPLWithContext(cfg *config.Config, ctx *api.RuntimeContext) int {
	if cfg.AllowInsecureTLS {
		fmt.Fprintf(ctx.Stderr, "WARNING: TLS certificate verification disabled globally. Use with caution.\n")
	}

	opts := &repl.Options{
		Prompt:      cfg.Prompt,
		NoWelcome:   cfg.NoWelcome,
		NoHistory:   cfg.NoHistory,
		HistoryFile: cfg.HistoryFile,
		TraceOn:     cfg.TraceOn,
		Args:        cfg.Args,
	}

	r, err := repl.NewREPLWithOptions(opts)
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "Error initializing REPL: %v\n", err)
		return api.ExitError
	}

	if err := r.Run(); err != nil {
		return api.HandleErrorWithContext(err)
	}

	return api.ExitSuccess
}
