package main

import (
	"fmt"
	"io"

	"github.com/marcin-radoszewski/viro/internal/api"
	"github.com/marcin-radoszewski/viro/internal/config"
	"github.com/marcin-radoszewski/viro/internal/eval"
	"github.com/marcin-radoszewski/viro/internal/native"
	"github.com/marcin-radoszewski/viro/internal/parse"
	"github.com/marcin-radoszewski/viro/internal/profile"
	"github.com/marcin-radoszewski/viro/internal/repl"
	"github.com/marcin-radoszewski/viro/internal/trace"
	"github.com/marcin-radoszewski/viro/internal/verror"
)

func Run(ctx *api.RuntimeContext) int {
	cfg, err := loadConfigurationWithContext(ctx)
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "Configuration error: %v\n", err)
		return ExitUsage
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
		return ExitUsage
	}

	switch mode {
	case config.ModeREPL:
		return runREPLWithContext(cfg, ctx)
	case config.ModeScript, config.ModeEval, config.ModeCheck:
		return runExecutionWithContext(cfg, mode, ctx)
	case config.ModeVersion:
		fmt.Fprintf(ctx.Stdout, "%s\n", getVersionString())
		return ExitSuccess
	case config.ModeHelp:
		fmt.Fprintf(ctx.Stdout, "%s", getHelpText())
		return ExitSuccess
	default:
		fmt.Fprintf(ctx.Stderr, "Unknown mode: %v\n", mode)
		return ExitUsage
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
		return ExitError
	}

	if err := r.Run(); err != nil {
		return handleErrorWithContext(err)
	}

	return ExitSuccess
}

func runExecutionWithContext(cfg *config.Config, mode config.Mode, ctx *api.RuntimeContext) int {
	var err error
	if cfg.Profile {
		err = trace.InitTraceSilent()
	} else {
		err = trace.InitTrace("", defaultTraceMaxSizeMB)
	}

	if err != nil {
		fmt.Fprintf(ctx.Stderr, "Error initializing trace: %v\n", err)
		return ExitInternal
	}

	var profiler *profile.Profiler
	if cfg.Profile && trace.GlobalTraceSession != nil {
		profiler = profile.NewProfiler()
		profile.EnableProfilingWithTrace(trace.GlobalTraceSession, profiler)
	}

	var input InputSource

	switch mode {
	case config.ModeCheck:
		input = &FileInput{Config: cfg, Path: cfg.ScriptFile}
	case config.ModeEval:
		input = &ExprInputWithContext{Expr: cfg.EvalExpr, WithStdin: cfg.ReadStdin, Stdin: ctx.Stdin}
	case config.ModeScript:
		input = &FileInput{Config: cfg, Path: cfg.ScriptFile}
	}

	var args []string
	if mode == config.ModeScript {
		args = cfg.Args
	} else {
		args = []string{}
	}

	printResult := (mode == config.ModeEval && !cfg.NoPrint)
	parseOnly := (mode == config.ModeCheck)

	exitCode := executeViroCodeWithContext(cfg, input, args, printResult, parseOnly, profiler, ctx)

	if profiler != nil {
		profiler.Disable()
		if !cfg.Quiet {
			report := profiler.GetReport()
			report.FormatText(ctx.Stderr)
		}
	}

	if trace.GlobalTraceSession != nil {
		trace.GlobalTraceSession.Close()
	}

	return exitCode
}

type ExprInputWithContext struct {
	Expr      string
	WithStdin bool
	Stdin     io.Reader
}

func (e *ExprInputWithContext) Load() (string, error) {
	expr := e.Expr

	if e.WithStdin {
		stdinData, err := io.ReadAll(e.Stdin)
		if err != nil {
			return "", fmt.Errorf("error reading stdin: %w", err)
		}
		expr = string(stdinData) + "\n" + expr
	}

	return expr, nil
}

func executeViroCodeWithContext(cfg *config.Config, input InputSource, args []string, printResult bool, parseOnly bool, profiler *profile.Profiler, ctx *api.RuntimeContext) int {
	content, err := input.Load()
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "Error loading input: %v\n", err)
		return ExitError
	}

	values, err := parse.Parse(content)
	if err != nil {
		printErrorToWriter(err, "Parse", ctx.Stderr)
		return ExitSyntax
	}

	if parseOnly {
		if cfg.Verbose {
			fmt.Fprintf(ctx.Stdout, "✓ Syntax valid\n")
			fmt.Fprintf(ctx.Stdout, "  Parsed %d expressions\n", len(values))
		}
		return ExitSuccess
	}

	evaluator := setupEvaluatorWithContext(cfg, ctx)
	initializeSystemObjectInEvaluator(evaluator, args)

	result, err := evaluator.DoBlock(values)
	if err != nil {
		printErrorToWriter(err, "Runtime", ctx.Stderr)
		return handleErrorWithContext(err)
	}

	if printResult && !cfg.Quiet {
		fmt.Fprintln(ctx.Stdout, result.Form())
	}

	return ExitSuccess
}

func setupEvaluatorWithContext(cfg *config.Config, ctx *api.RuntimeContext) *eval.Evaluator {
	evaluator := eval.NewEvaluator()

	if cfg.Quiet {
		evaluator.SetOutputWriter(io.Discard)
	} else {
		evaluator.SetOutputWriter(ctx.Stdout)
	}
	evaluator.SetErrorWriter(ctx.Stderr)
	evaluator.SetInputReader(ctx.Stdin)

	rootFrame := evaluator.GetFrameByIndex(0)
	native.RegisterMathNatives(rootFrame)
	native.RegisterSeriesNatives(rootFrame)
	native.RegisterDataNatives(rootFrame)
	native.RegisterIONatives(rootFrame, evaluator)
	native.RegisterControlNatives(rootFrame)
	native.RegisterHelpNatives(rootFrame)

	return evaluator
}

func handleErrorWithContext(err error) int {
	if err == nil {
		return ExitSuccess
	}

	if vErr, ok := err.(*verror.Error); ok {
		return categoryToExitCode(vErr.Category)
	}

	return ExitError
}

func printErrorToWriter(err error, prefix string, w io.Writer) {
	if vErr, ok := err.(*verror.Error); ok {
		fmt.Fprintf(w, "%v", vErr)
	} else if prefix != "" {
		fmt.Fprintf(w, "%s error: %v\n", prefix, err)
	} else {
		fmt.Fprintf(w, "Error: %v\n", err)
	}
}
