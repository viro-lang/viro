package main

import (
	"fmt"
	"os"

	"github.com/marcin-radoszewski/viro/internal/config"
	"github.com/marcin-radoszewski/viro/internal/core"
	"github.com/marcin-radoszewski/viro/internal/eval"
	"github.com/marcin-radoszewski/viro/internal/frame"
	"github.com/marcin-radoszewski/viro/internal/parse"
	"github.com/marcin-radoszewski/viro/internal/profile"
	"github.com/marcin-radoszewski/viro/internal/trace"
	"github.com/marcin-radoszewski/viro/internal/value"
)

const (
	defaultTraceMaxSizeMB = 50
)

type ExecutionContext struct {
	Config      *config.Config
	Input       InputSource
	Args        []string
	PrintResult bool
	ParseOnly   bool
	Profiler    *profile.Profiler
}

func runExecution(cfg *config.Config, mode config.Mode) int {
	var err error
	if cfg.Profile {
		err = trace.InitTraceSilent()
	} else {
		err = trace.InitTrace("", defaultTraceMaxSizeMB)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing trace: %v\n", err)
		return ExitInternal
	}

	var profiler *profile.Profiler
	if cfg.Profile && trace.GlobalTraceSession != nil {
		profiler = profile.NewProfiler()
		profile.EnableProfilingWithTrace(trace.GlobalTraceSession, profiler)
	}

	var ctx *ExecutionContext

	switch mode {
	case config.ModeCheck:
		ctx = &ExecutionContext{
			Config:      cfg,
			Input:       &FileInput{Config: cfg, Path: cfg.ScriptFile},
			Args:        nil,
			PrintResult: false,
			ParseOnly:   true,
			Profiler:    profiler,
		}
	case config.ModeEval:
		ctx = &ExecutionContext{
			Config:      cfg,
			Input:       &ExprInput{Expr: cfg.EvalExpr, WithStdin: cfg.ReadStdin},
			Args:        []string{},
			PrintResult: !cfg.NoPrint,
			ParseOnly:   false,
			Profiler:    profiler,
		}
	case config.ModeScript:
		ctx = &ExecutionContext{
			Config:      cfg,
			Input:       &FileInput{Config: cfg, Path: cfg.ScriptFile},
			Args:        cfg.Args,
			PrintResult: false,
			ParseOnly:   false,
			Profiler:    profiler,
		}
	}

	_, exitCode := executeViroCode(ctx)

	if ctx.Profiler != nil {
		ctx.Profiler.Disable()
		if !cfg.Quiet {
			report := ctx.Profiler.GetReport()
			report.FormatText(os.Stderr)
		}
	}

	if trace.GlobalTraceSession != nil {
		trace.GlobalTraceSession.Close()
	}

	return exitCode
}

func executeViroCode(ctx *ExecutionContext) (core.Value, int) {
	content, err := ctx.Input.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading input: %v\n", err)
		return nil, ExitError
	}

	values, err := parse.Parse(content)
	if err != nil {
		printError(err, "Parse")
		return nil, ExitSyntax
	}

	if ctx.ParseOnly {
		if ctx.Config.Verbose {
			fmt.Printf("✓ Syntax valid\n")
			fmt.Printf("  Parsed %d expressions\n", len(values))
		}
		return nil, ExitSuccess
	}

	evaluator := setupEvaluator(ctx.Config)
	initializeSystemObjectInEvaluator(evaluator, ctx.Args)

	result, err := evaluator.DoBlock(values)
	if err != nil {
		printError(err, "Runtime")
		return nil, handleError(err)
	}

	if ctx.PrintResult && !ctx.Config.Quiet {
		fmt.Println(result.Form())
	}

	return result, ExitSuccess
}

func initializeSystemObjectInEvaluator(evaluator *eval.Evaluator, args []string) {
	viroArgs := make([]core.Value, len(args))
	for i, arg := range args {
		viroArgs[i] = value.NewStringValue(arg)
	}

	argsBlock := value.NewBlockValue(viroArgs)

	ownedFrame := frame.NewFrame(frame.FrameObject, -1)
	ownedFrame.Bind("args", argsBlock)

	systemObj := value.NewObject(ownedFrame)

	rootFrame := evaluator.GetFrameByIndex(0)
	rootFrame.Bind("system", systemObj)
}
