# Plan 020: Refactor Integration Tests to Use Testable API

## Overview
Refactor integration tests from using `exec.Command()` with the compiled binary to calling an exported API function directly. This fixes GitHub CI failures and follows Go best practices for CLI testing.

## Problem Statement
- Integration tests in `test/integration/` require the compiled `viro` binary at `../../viro`
- In GitHub CI, tests run before the build step, causing tests to skip
- Binary-based testing is slower and less isolated than direct API calls
- User wants "some API covering the app usage" instead of relying on the binary

## Goals
1. Create an exported `Run()` function in `cmd/viro` package
2. Refactor integration tests to call `Run()` directly instead of `exec.Command()`
3. Ensure tests work in CI without requiring the binary to be built first
4. Maintain backward compatibility with existing behavior
5. Improve test speed and isolation

## Non-Goals
- Changing the CLI behavior or user-facing functionality
- Modifying the REPL implementation
- Refactoring contract/unit tests (only integration tests)
- Removing binary-based tests entirely (keep as optional smoke tests)

## Current State Analysis

### Strengths
- Code already uses `flag.NewFlagSet()` (ready for custom args)
- Logic well-separated from `main()` into functions
- Exit codes properly defined as constants
- `splitCommandLineArgs()` already accepts custom args slice
- Config structure cleanly separated

### Challenges
1. **I/O Dependencies**: Many functions use `os.Stdout`, `os.Stderr`, `fmt.Fprintf(os.Stderr, ...)`
2. **Signal Handler**: `setupSignalHandler()` uses global channels - needs to be optional for tests
3. **Flag Parsing**: `LoadFromFlags()` reads `os.Args[1:]` - needs to accept custom args
4. **Input Sources**: `FileInput` and `ExprInput` structs may need stdin injection
5. **Print Statements**: Direct printing scattered throughout execution flow

## Implementation Plan

### Phase 1: Create Core API Function

**File: `cmd/viro/run.go` (NEW)**

```go
package main

import "io"

// Run executes the viro interpreter with the given runtime context.
// It returns the exit code that would normally be passed to os.Exit().
//
// Parameters:
//   - ctx: RuntimeContext containing args and I/O streams
//
// Returns:
//   - Exit code (0 for success, non-zero for errors)
//
// Example:
//   ctx := &RuntimeContext{
//       Args:   []string{"-c", "3 + 4"},
//       Stdin:  nil,
//       Stdout: &stdout,
//       Stderr: &stderr,
//   }
//   exitCode := Run(ctx)
func Run(ctx *RuntimeContext) int {
	return runWithContext(ctx)
}

// RuntimeContext holds I/O dependencies for execution
type RuntimeContext struct {
	Args   []string
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
}

func runWithContext(ctx *RuntimeContext) int {
	cfg, err := loadConfigurationWithContext(ctx)
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "Configuration error: %v\n", err)
		return ExitUsage
	}
	
	return executeModeWithContext(cfg, ctx)
}
```

### Phase 2: Refactor Configuration Loading

**File: `cmd/viro/config.go` (MODIFY)**

Add method to accept custom args:

```go
// LoadFromFlagsWithArgs parses command-line arguments from the provided slice
func (c *Config) LoadFromFlagsWithArgs(args []string) error {
	fs := flag.NewFlagSet("viro", flag.ContinueOnError)
	
	// ... existing flag definitions (lines 60-78) ...
	
	// Use provided args instead of os.Args[1:]
	parsed := splitCommandLineArgs(args)
	
	// ... rest of existing logic (lines 83-131) ...
}

// LoadFromFlags uses os.Args for backward compatibility
func (c *Config) LoadFromFlags() error {
	return c.LoadFromFlagsWithArgs(os.Args[1:])
}
```

### Phase 3: Thread I/O Context Through Execution

**File: `cmd/viro/execution.go` (MODIFY)**

Update functions to accept and use RuntimeContext:

```go
func runExecution(cfg *Config, mode Mode) int {
	return runExecutionWithContext(cfg, mode, &RuntimeContext{
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	})
}

func runExecutionWithContext(cfg *Config, mode Mode, ctx *RuntimeContext) int {
	// ... existing trace initialization ...
	
	execCtx := &ExecutionContext{
		Config:      cfg,
		// ... existing fields ...
		RuntimeCtx:  ctx,  // NEW
	}
	
	// ... rest of logic using ctx.Stderr instead of os.Stderr ...
}

func executeViroCode(ctx *ExecutionContext) (core.Value, int) {
	content, err := ctx.Input.Load()
	if err != nil {
		// OLD: fmt.Fprintf(os.Stderr, "Error loading input: %v\n", err)
		// NEW:
		fmt.Fprintf(ctx.RuntimeCtx.Stderr, "Error loading input: %v\n", err)
		return nil, ExitError
	}
	
	// ... rest of logic using ctx.RuntimeCtx for I/O ...
}
```

**Add to ExecutionContext struct:**
```go
type ExecutionContext struct {
	Config      *Config
	Input       InputSource
	Args        []string
	PrintResult bool
	ParseOnly   bool
	Profiler    *profile.Profiler
	RuntimeCtx  *RuntimeContext  // NEW
}
```

### Phase 4: Update Main Entry Point

**File: `cmd/viro/main.go` (MODIFY)**

```go
func main() {
	setupSignalHandler()
	ctx := &RuntimeContext{
		Args:   os.Args[1:],
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}
	exitCode := Run(ctx)
	os.Exit(exitCode)
}
```

Keep existing `loadConfiguration()` and `executeMode()` as wrappers for backward compatibility.

### Phase 5: Refactor Integration Tests

**Strategy**: Convert tests one file at a time, starting with simplest.

**Order of Conversion**:
1. `eval_test.go` - Simple eval mode tests
2. `eval_stdin_test.go` - Tests with stdin
3. `script_exec_test.go` - Script execution tests
4. `check_test.go` - Syntax check tests
5. Continue with remaining files

**Example Refactored Test** (`test/integration/eval_test.go`):

```go
package integration

import (
	"bytes"
	"strings"
	"testing"
	
	"github.com/marcin-radoszewski/viro/cmd/viro"
)

func TestEvalModeAPI(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		want     string
		wantExit int
	}{
		{
			name:     "simple math",
			args:     []string{"-c", "3 + 4"},
			want:     "7",
			wantExit: 0,
		},
		{
			name:     "multiplication",
			args:     []string{"-c", "10 * 5"},
			want:     "50",
			wantExit: 0,
		},
		// ... rest of test cases ...
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			ctx := &viro.RuntimeContext{
				Args:   tt.args,
				Stdin:  nil,
				Stdout: &stdout,
				Stderr: &stderr,
			}
			
			exitCode := viro.Run(ctx)
			
			if exitCode != tt.wantExit {
				t.Errorf("exit code = %d, want %d\nStderr: %s", 
					exitCode, tt.wantExit, stderr.String())
			}
			
			output := stdout.String()
			if !strings.Contains(output, tt.want) {
				t.Errorf("output = %q, want to contain %q", output, tt.want)
			}
		})
	}
}
```

**Example with Stdin** (`test/integration/eval_stdin_test.go`):

```go
func TestEvalModeWithStdinAPI(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		stdin    string
		want     string
		wantExit int
	}{
		{
			name:     "read from stdin",
			args:     []string{"-c", "first data", "--stdin"},
			stdin:    "[1 2 3]",
			want:     "1",
			wantExit: 0,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			stdin := strings.NewReader(tt.stdin)
			ctx := &viro.RuntimeContext{
				Args:   tt.args,
				Stdin:  stdin,
				Stdout: &stdout,
				Stderr: &stderr,
			}
			
			exitCode := viro.Run(ctx)
			
			if exitCode != tt.wantExit {
				t.Errorf("exit code = %d, want %d", exitCode, tt.wantExit)
			}
			
			if !strings.Contains(stdout.String(), tt.want) {
				t.Errorf("output = %q, want %q", stdout.String(), tt.want)
			}
		})
	}
}
```

### Phase 6: Handle Special Cases

#### 6.1 Signal Handler
Make signal handler optional in tests:

```go
func setupSignalHandlerIfNeeded(enable bool) {
	if !enable {
		return
	}
	// ... existing signal handler code ...
}
```

#### 6.2 REPL Tests
For REPL tests, may need special handling since REPL is interactive:
- Keep some REPL tests binary-based (mark with build tags)
- Or create mock REPL input/output streams

#### 6.3 File-based Tests
Script execution tests that create temp files should work as-is:
```go
func TestScriptExecutionAPI(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "test_*.viro")
	// ... write script content ...
	
	var stdout, stderr bytes.Buffer
	ctx := &viro.RuntimeContext{
		Args:   []string{tmpfile.Name()},
		Stdin:  nil,
		Stdout: &stdout,
		Stderr: &stderr,
	}
	exitCode := viro.Run(ctx)
	// ... assertions ...
}
```

### Phase 7: Optional Binary-Based Smoke Tests

Keep minimal binary-based tests for end-to-end validation:

**File: `test/integration/binary_smoke_test.go` (NEW)**

```go
// +build integration

package integration

import (
	"os"
	"os/exec"
	"testing"
)

func TestBinarySmoke(t *testing.T) {
	viroPath := "../../viro"
	if _, err := os.Stat(viroPath); os.IsNotExist(err) {
		t.Skip("viro binary not found - this is optional smoke test")
	}
	
	// Just one or two critical end-to-end tests
	cmd := exec.Command(viroPath, "-c", "3 + 4")
	output, err := cmd.CombinedOutput()
	
	if err != nil {
		t.Fatalf("binary execution failed: %v", err)
	}
	
	if !strings.Contains(string(output), "7") {
		t.Errorf("unexpected output: %s", output)
	}
}
```

Run with: `go test -tags=integration ./test/integration`

## Checklist

### Phase 1: Core API
- [ ] Create `cmd/viro/run.go` with `Run()` function
- [ ] Create `RuntimeContext` struct
- [ ] Implement `runWithContext()` function
- [ ] Add unit tests for `Run()` function

### Phase 2: Configuration
- [ ] Add `LoadFromFlagsWithArgs()` method to Config
- [ ] Update existing `LoadFromFlags()` to call new method
- [ ] Test config loading with custom args
- [ ] Ensure backward compatibility

### Phase 3: I/O Context
- [ ] Add `RuntimeContext` field to `ExecutionContext`
- [ ] Create `runExecutionWithContext()` function
- [ ] Update `executeViroCode()` to use context I/O
- [ ] Update `printError()` to accept stderr writer
- [ ] Thread context through all print statements

### Phase 4: Main Entry Point
- [ ] Update `main()` to call `Run()`
- [ ] Update `loadConfiguration()` to use context
- [ ] Update `executeMode()` to use context
- [ ] Test binary still works correctly

### Phase 5: Integration Tests
- [ ] Refactor `eval_test.go`
- [ ] Refactor `eval_stdin_test.go`
- [ ] Refactor `script_exec_test.go`
- [ ] Refactor `check_test.go`
- [ ] Refactor remaining integration test files
- [ ] Remove binary path checks from tests
- [ ] Verify all tests pass

### Phase 6: Special Cases
- [ ] Make signal handler optional
- [ ] Handle REPL tests appropriately
- [ ] Verify file-based tests work
- [ ] Test error cases and exit codes

### Phase 7: CI and Documentation
- [ ] Verify tests pass in CI without build step
- [ ] Update `.github/workflows/ci.yml` if needed
- [ ] Add documentation to AGENTS.md about new testing approach
- [ ] Create optional smoke tests with build tag

### Phase 8: Cleanup and Validation
- [ ] Run full test suite: `go test ./...`
- [ ] Run tests without binary: `rm viro && go test ./test/integration`
- [ ] Test binary execution: `make build && ./viro -c "3 + 4"`
- [ ] Check test coverage: `go test -coverprofile=coverage.out ./...`
- [ ] Verify CI passes

## Testing Strategy

### Unit Tests
- Test `Run()` function with various argument combinations
- Test config loading with custom args
- Test I/O redirection works correctly

### Integration Tests
- All existing integration tests converted to API-based
- Tests should pass without binary present
- Tests should be faster than binary-based tests

### Smoke Tests (Optional)
- Minimal binary-based tests with build tag
- Only run when binary is present
- Validate end-to-end binary functionality

## Rollback Plan

If issues arise:
1. Keep old binary-based tests in separate files with different names
2. Can revert to old approach by uncommenting old tests
3. No changes to actual CLI behavior, so users unaffected

## Success Criteria

- [ ] All integration tests pass without requiring binary
- [ ] Tests work in GitHub CI without building binary first
- [ ] Tests run faster than before (no process spawning)
- [ ] CLI binary behavior unchanged
- [ ] Test coverage maintained or improved
- [ ] Code follows existing style (no comments in code)

## Estimated Effort

- Phase 1-2: 1-2 hours (API function and config)
- Phase 3-4: 2-3 hours (I/O threading and main refactor)
- Phase 5: 3-4 hours (test conversion)
- Phase 6-8: 1-2 hours (special cases, CI, validation)

**Total: 7-11 hours**

## Benefits After Completion

1. **CI Reliability**: Tests run without binary, no build ordering issues
2. **Test Speed**: 2-5x faster tests (no process spawning)
3. **Test Isolation**: Better control over stdin/stdout/stderr
4. **Debugging**: Easier to debug tests (direct function calls)
5. **Maintainability**: Standard Go CLI testing pattern
6. **Flexibility**: Easy to test edge cases with mocked I/O

## References

- Similar pattern: https://github.com/kubernetes/kubectl/blob/master/pkg/cmd/cmd.go
- Go testing best practices: https://golang.org/doc/code.html#Testing
- Flag package documentation: https://pkg.go.dev/flag
