# Plan 017: cmd/viro Code Deduplication and Complexity Reduction

## Overview
Analysis of `cmd/viro/` reveals several opportunities for code unification, deduplication, and complexity reduction. This plan proposes refactorings to improve maintainability and reduce redundancy.

## Current State Analysis

### File Organization (14 files)
- **argparse.go** (47 lines): Command-line argument splitting logic
- **config.go** (140 lines): Configuration struct and loading (env + flags)
- **config_test.go** (408 lines): Configuration tests
- **evaluator.go** (27 lines): Evaluator setup and native registration
- **execution.go** (117 lines): Execution context and code execution
- **exit.go** (64 lines): Exit codes and error handling
- **exit_test.go** (85 lines): Error handling tests
- **help.go** (85 lines): Help text
- **help_test.go** (12 lines): Help smoke tests
- **input.go** (56 lines): Input sources (FileInput, ExprInput)
- **main.go** (89 lines): Main entry point and orchestration
- **mode.go** (86 lines): Mode detection logic
- **mode_test.go** (99 lines): Mode detection tests
- **version.go** (17 lines): Version information

**Total: ~1,300 lines**

## Identified Issues

### 1. **Duplicated Error Printing Logic**

**Location:** `exit.go:47-63`

Three nearly identical functions for error printing:
```go
func printError(err error, prefix string)       // Generic
func printParseError(err error)                 // Calls printError("Parse")
func printRuntimeError(err error)               // Calls printError("Runtime")
```

**Problem:** The two wrapper functions add minimal value (just prefix string).

**Proposed Fix:** 
- Remove `printParseError` and `printRuntimeError`
- Call `printError` directly with appropriate prefix at call sites
- Reduces 3 functions to 1

**Impact:** -12 lines, improved clarity

---

### 2. **Duplicated Flag Processing Pattern**

**Location:** `config.go:96-126`

Repetitive conditional assignment pattern:
```go
if *sandboxRoot != "" {
    c.SandboxRoot = *sandboxRoot
}
c.AllowInsecureTLS = c.AllowInsecureTLS || *allowInsecureTLS
c.Quiet = *quiet
// ... repeated 10+ times
```

**Problem:** Mix of different assignment patterns (conditional, boolean OR, direct).

**Proposed Fix:**
Create helper method:
```go
func (c *Config) applyFlags(fs *flagSet) {
    c.applyStringFlag(&c.SandboxRoot, fs.sandboxRoot)
    c.applyBoolFlag(&c.AllowInsecureTLS, fs.allowInsecureTLS, true) // OR mode
    c.applyBoolFlag(&c.Quiet, fs.quiet, false)                      // direct
    // ...
}
```

**Impact:** -20 lines, more consistent pattern

---

### 3. **Duplicated Mode Detection and Validation**

**Location:** `mode.go:35-85` and `config.go:136-139`

Mode detection logic includes validation (e.g., "check requires script"), but `Config.Validate()` just calls `detectMode()`.

**Problem:**
- `Config.Validate()` is a redundant wrapper
- Mode detection and validation are mixed
- Validation happens twice in main flow: `cfg.Validate()` then `executeMode(cfg)` calls `detectMode()` again

**Proposed Fix:**
```go
// In config.go
func (c *Config) Validate() error {
    // Inline simple validations without mode detection
    if c.CheckOnly && c.ScriptFile == "" {
        return fmt.Errorf("--check flag requires a script file")
    }
    if c.ReadStdin && c.EvalExpr == "" {
        return fmt.Errorf("--stdin flag requires -c flag")
    }
    if c.NoPrint && c.EvalExpr == "" {
        return fmt.Errorf("--no-print flag requires -c flag")
    }
    return nil
}

// In mode.go - detectMode() only detects, doesn't validate
```

**Impact:** Clearer separation of concerns, removes redundant detectMode() call

---

### 4. **ExecutionMode vs Mode Duplication**

**Location:** `execution.go:14-20` and `mode.go:5-14`

Two separate enum types for similar concepts:
```go
// execution.go
type ExecutionMode int
const (
    ExecuteModeCheck ExecutionMode = iota
    ExecuteModeEval
    ExecuteModeScript
)

// mode.go
type Mode int
const (
    ModeREPL Mode = iota
    ModeScript
    ModeEval
    ModeCheck
    ModeVersion
    ModeHelp
)
```

**Problem:** 
- `ExecutionMode` is subset of `Mode`
- Conversion happens in `main.go:46-52` via map
- Extra abstraction layer with no clear benefit

**Proposed Fix:**
- Remove `ExecutionMode` enum
- Change `runExecution(cfg *Config, mode ExecutionMode)` to `runExecution(cfg *Config, mode Mode)`
- Build `ExecutionContext` based on `Mode` directly

**Impact:** -20 lines, simpler type system

---

### 5. **Input Source Abstraction Underused**

**Location:** `input.go:10-56`

`InputSource` interface with `FileInput` and `ExprInput` implementations:
```go
type InputSource interface {
    Load() (string, error)
}
```

**Problem:**
- Only used in `execution.go`, not exposed elsewhere
- Interface defined but never used polymorphically (types used directly in construction)
- Small abstraction over simple operations

**Evaluation:** 
- **Keep** - abstraction is reasonable and may grow
- **Improvement:** Could add `StdinInput` type to replace `FileInput{Path: "-"}` special case

**Impact:** No change recommended, but document the abstraction's purpose

---

### 6. **Configuration Loading Chain**

**Location:** `main.go:25-36`

```go
func loadConfiguration() (*Config, error) {
    cfg := NewConfig()
    if err := cfg.LoadFromEnv(); err != nil {
        return nil, err
    }
    if err := cfg.LoadFromFlags(); err != nil {
        return nil, err
    }
    if err := cfg.Validate(); err != nil {
        return nil, err
    }
    return cfg, nil
}
```

**Problem:**
- Sequential error checking is verbose
- `Validate()` is redundant (see issue #3)

**Proposed Fix:**
```go
func loadConfiguration() (*Config, error) {
    cfg := NewConfig()
    
    loaders := []func() error{
        cfg.LoadFromEnv,
        cfg.LoadFromFlags,
        cfg.Validate,
    }
    
    for _, load := range loaders {
        if err := load(); err != nil {
            return nil, err
        }
    }
    
    return cfg, nil
}
```

**Evaluation:** Current form is actually clearer. **No change recommended.**

---

### 7. **Handler Map Pattern in executeMode**

**Location:** `main.go:46-53`

```go
handlers := map[Mode]func(*Config) int{
    ModeREPL:    runREPL,
    ModeScript:  func(cfg *Config) int { return runExecution(cfg, ExecuteModeScript) },
    ModeEval:    func(cfg *Config) int { return runExecution(cfg, ExecuteModeEval) },
    ModeCheck:   func(cfg *Config) int { return runExecution(cfg, ExecuteModeCheck) },
    ModeVersion: func(cfg *Config) int { printVersion(); return ExitSuccess },
    ModeHelp:    func(cfg *Config) int { printHelp(); return ExitSuccess },
}
```

**Problem:**
- Unnecessary indirection for simple cases (Version, Help)
- Anonymous lambdas for execution modes could be clearer

**Proposed Fix:** 
After removing `ExecutionMode` (issue #4):
```go
switch mode {
case ModeREPL:
    return runREPL(cfg)
case ModeScript, ModeEval, ModeCheck:
    return runExecution(cfg, mode)
case ModeVersion:
    printVersion()
    return ExitSuccess
case ModeHelp:
    printHelp()
    return ExitSuccess
default:
    fmt.Fprintf(os.Stderr, "Unknown mode: %v\n", mode)
    return ExitUsage
}
```

**Impact:** +4 lines but clearer control flow, eliminates map allocation

---

### 8. **Sandbox Root Default Assignment**

**Location:** `config.go:120-126`

```go
if c.SandboxRoot == "" {
    cwd, err := os.Getwd()
    if err != nil {
        return fmt.Errorf("error getting current directory: %w", err)
    }
    c.SandboxRoot = cwd
}
```

**Problem:** Happens in `LoadFromFlags()` which violates single responsibility (loading vs. defaulting).

**Proposed Fix:**
Move to separate method:
```go
func (c *Config) ApplyDefaults() error {
    if c.SandboxRoot == "" {
        cwd, err := os.Getwd()
        if err != nil {
            return fmt.Errorf("error getting current directory: %w", err)
        }
        c.SandboxRoot = cwd
    }
    return nil
}
```

Call from `loadConfiguration()` chain.

**Impact:** Better separation of concerns

---

### 9. **Argument Parsing Flags Hardcoded**

**Location:** `argparse.go:25-36`

```go
if arg == "-c" {
    if i+1 < len(args) {
        i++
    }
    continue
}

if arg == "--sandbox-root" || arg == "--history-file" || arg == "--prompt" {
    if i+1 < len(args) {
        i++
    }
    continue
}
```

**Problem:**
- Flag names duplicated between `argparse.go` and `config.go` flag definitions
- Adding new flag with value requires updating both files
- Brittle maintenance

**Proposed Fix:**
```go
var flagsWithValues = map[string]bool{
    "-c":             true,
    "--sandbox-root": true,
    "--history-file": true,
    "--prompt":       true,
}

func splitCommandLineArgs(args []string) *ParsedArgs {
    // ...
    if flagsWithValues[arg] {
        if i+1 < len(args) {
            i++
        }
        continue
    }
    // ...
}
```

**Impact:** Single source of truth for flags

---

### 10. **Test Organization**

**Current:**
- `config_test.go`: 408 lines
- Mix of unit tests, integration tests, table-driven tests

**Observation:** Tests are well-structured but could benefit from:
- Subtests for environment variable cleanup (already doing this)
- Helper functions for common test setup (flag reset)

**Recommendation:** Add test helpers:
```go
func setupTestArgs(t *testing.T, args []string) {
    t.Helper()
    flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
    os.Args = args
}
```

**Impact:** -50 lines in tests, improved readability

---

## Summary of Proposed Changes

### High Priority (Immediate Impact)
1. **Remove `printParseError`/`printRuntimeError` wrappers** (-12 lines)
2. **Eliminate `ExecutionMode` enum** (-20 lines, cleaner types)
3. **Fix mode detection/validation duplication** (clearer flow)
4. **Simplify `executeMode` with switch** (+4 lines but clearer)

### Medium Priority (Maintainability)
5. **Extract sandbox root defaulting** (better SRP)
6. **Centralize flags-with-values in argparse** (DRY)
7. **Add test helper functions** (-50 test lines)

### Low Priority (Nice to Have)
8. **Flag assignment helper methods** (-20 lines)

## Estimated Impact
- **Lines removed:** ~100 lines (7% reduction)
- **Complexity reduction:** Fewer enums, clearer control flow
- **Maintainability:** Centralized flag definitions, clearer separation of concerns

## Implementation Order
1. Remove error wrapper functions (exit.go)
2. Eliminate ExecutionMode (execution.go, main.go)
3. Simplify executeMode to switch statement (main.go)
4. Fix mode detection/validation split (mode.go, config.go)
5. Extract sandbox root defaulting (config.go)
6. Centralize flag definitions (argparse.go)
7. Add test helpers (config_test.go)
8. Optional: Flag assignment helpers (config.go)

## Files Affected
- **Modified:** main.go, config.go, execution.go, exit.go, mode.go, argparse.go
- **Tests:** config_test.go, exit_test.go
- **Unchanged:** evaluator.go, input.go, help.go, version.go

## Validation
- All existing tests must pass
- No behavioral changes
- Test coverage maintained at current levels
