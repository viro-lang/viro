package main

import (
	"fmt"
	"os"

	"github.com/marcin-radoszewski/viro/internal/core"
	"github.com/marcin-radoszewski/viro/internal/eval"
	"github.com/marcin-radoszewski/viro/internal/frame"
	"github.com/marcin-radoszewski/viro/internal/parse"
	"github.com/marcin-radoszewski/viro/internal/value"
)

type ExecutionContext struct {
	Config      *Config
	Input       InputSource
	Args        []string
	PrintResult bool
	ParseOnly   bool
}

func runExecution(cfg *Config, mode Mode) int {
	var ctx *ExecutionContext

	switch mode {
	case ModeCheck:
		ctx = &ExecutionContext{
			Config:      cfg,
			Input:       &FileInput{Config: cfg, Path: cfg.ScriptFile},
			Args:        nil,
			PrintResult: false,
			ParseOnly:   true,
		}
	case ModeEval:
		ctx = &ExecutionContext{
			Config:      cfg,
			Input:       &ExprInput{Expr: cfg.EvalExpr, WithStdin: cfg.ReadStdin},
			Args:        []string{},
			PrintResult: !cfg.NoPrint,
			ParseOnly:   false,
		}
	case ModeScript:
		ctx = &ExecutionContext{
			Config:      cfg,
			Input:       &FileInput{Config: cfg, Path: cfg.ScriptFile},
			Args:        cfg.Args,
			PrintResult: false,
			ParseOnly:   false,
		}
	}

	_, exitCode := executeViroCode(ctx)
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
