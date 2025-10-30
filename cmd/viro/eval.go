package main

import (
	"fmt"
	"io"
	"os"

	"github.com/marcin-radoszewski/viro/internal/core"
	"github.com/marcin-radoszewski/viro/internal/parse"
)

func runEval(cfg *Config) int {
	expr := cfg.EvalExpr

	if cfg.ReadStdin {
		stdinData, err := io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading stdin: %v\n", err)
			return ExitError
		}
		expr = string(stdinData) + "\n" + expr
	}

	values, err := parse.Parse(expr)
	if err != nil {
		printParseError(err)
		return ExitSyntax
	}

	evaluator := setupEvaluator(cfg)

	result, err := evaluator.DoBlock(values)
	if err != nil {
		printRuntimeError(err)
		return handleError(err)
	}

	if !cfg.NoPrint && !cfg.Quiet {
		fmt.Println(formatResult(result, cfg))
	}

	return ExitSuccess
}

func formatResult(val core.Value, cfg *Config) string {
	return val.Form()
}
