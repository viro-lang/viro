package main

import (
	"fmt"
	"os"

	"github.com/marcin-radoszewski/viro/internal/parse"
)

func runCheck(cfg *Config) int {
	scriptPath := cfg.ScriptFile
	content, err := loadScriptFile(cfg, scriptPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading script: %v\n", err)
		return ExitError
	}

	values, err := parse.Parse(content)
	if err != nil {
		printParseError(err)
		return ExitSyntax
	}

	if cfg.Verbose {
		fmt.Printf("✓ Syntax valid: %s\n", scriptPath)
		fmt.Printf("  Parsed %d expressions\n", len(values))
	}

	return ExitSuccess
}
