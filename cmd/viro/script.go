package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/marcin-radoszewski/viro/internal/parse"
)

func runScript(cfg *Config) int {
	content, err := loadScriptFile(cfg, cfg.ScriptFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading script: %v\n", err)
		return ExitError
	}

	return executeScript(cfg, content)
}

func loadScriptFile(cfg *Config, path string) (string, error) {
	if path == "-" {
		return readStdin()
	}

	fullPath := resolveScriptPath(cfg.SandboxRoot, path)

	data, err := os.ReadFile(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to read %s: %w", path, err)
	}

	return string(data), nil
}

func resolveScriptPath(sandboxRoot, path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(sandboxRoot, path)
}

func readStdin() (string, error) {
	data, err := io.ReadAll(os.Stdin)
	return string(data), err
}

func executeScript(cfg *Config, content string) int {
	values, err := parse.Parse(content)
	if err != nil {
		printParseError(err)
		return ExitSyntax
	}

	evaluator := setupEvaluator(cfg)

	_, err = evaluator.DoBlock(values)
	if err != nil {
		printRuntimeError(err)
		return handleError(err)
	}

	return ExitSuccess
}
