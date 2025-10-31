package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type InputSource interface {
	Load() (string, error)
}

type FileInput struct {
	Config *Config
	Path   string
}

func (f *FileInput) Load() (string, error) {
	if f.Path == "-" {
		data, err := io.ReadAll(os.Stdin)
		return string(data), err
	}

	fullPath := f.Path
	if !filepath.IsAbs(f.Path) {
		fullPath = filepath.Join(f.Config.SandboxRoot, f.Path)
	}

	data, err := os.ReadFile(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to read %s: %w", f.Path, err)
	}

	return string(data), nil
}

type ExprInput struct {
	Expr      string
	WithStdin bool
}

func (e *ExprInput) Load() (string, error) {
	expr := e.Expr

	if e.WithStdin {
		stdinData, err := io.ReadAll(os.Stdin)
		if err != nil {
			return "", fmt.Errorf("error reading stdin: %w", err)
		}
		expr = string(stdinData) + "\n" + expr
	}

	return expr, nil
}
