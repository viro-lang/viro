package eval

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// SandboxRoot holds the configured sandbox root directory for file operations.
// Set via --sandbox-root CLI flag, defaults to current working directory per FR-006.
var SandboxRoot string

// ResolveSandboxPath resolves a user-provided path relative to the sandbox root.
// Enforces sandbox restrictions per research.md security considerations:
// - Cleans path to normalize separators and remove ".." sequences
// - Joins with sandbox root to create absolute path
// - Evaluates symlinks to detect escape attempts
// - Verifies final path has sandbox root prefix
//
// Returns absolute path within sandbox, or error if path escapes sandbox.
func ResolveSandboxPath(userPath string) (string, error) {
	if SandboxRoot == "" {
		return "", fmt.Errorf("sandbox root not configured")
	}

	// Clean the user path to normalize it
	cleaned := filepath.Clean(userPath)

	// If path is absolute, verify it's within sandbox
	var candidate string
	if filepath.IsAbs(cleaned) {
		candidate = cleaned
	} else {
		// Relative path: join with sandbox root
		candidate = filepath.Join(SandboxRoot, cleaned)
	}

	// Evaluate symlinks to detect escape attempts
	resolved, err := filepath.EvalSymlinks(candidate)
	if err != nil {
		// Path doesn't exist yet - check parent directory chain
		// This allows creating new files within sandbox
		dir := filepath.Dir(candidate)
		resolvedDir, err := filepath.EvalSymlinks(dir)
		if err != nil {
			// Parent doesn't exist either - verify candidate path is within sandbox
			if !strings.HasPrefix(candidate, SandboxRoot) {
				return "", fmt.Errorf("path escapes sandbox: %s", userPath)
			}
			return candidate, nil
		}
		// Reconstruct with resolved parent + filename
		resolved = filepath.Join(resolvedDir, filepath.Base(candidate))
	}

	// Verify resolved path is within sandbox root
	if !strings.HasPrefix(resolved, SandboxRoot) {
		return "", fmt.Errorf("path escapes sandbox: %s resolves to %s", userPath, resolved)
	}

	return resolved, nil
}

// InitSandbox initializes the sandbox root directory.
// Called during REPL initialization with the value from --sandbox-root flag.
func InitSandbox(root string) error {
	if root == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}
		root = cwd
	}

	// Resolve to absolute path
	abs, err := filepath.Abs(root)
	if err != nil {
		return fmt.Errorf("failed to resolve sandbox root: %w", err)
	}

	// Verify directory exists
	info, err := os.Stat(abs)
	if err != nil {
		return fmt.Errorf("sandbox root does not exist: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("sandbox root is not a directory: %s", abs)
	}

	SandboxRoot = abs
	return nil
}
