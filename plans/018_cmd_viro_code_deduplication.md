# Plan 018: cmd/viro Code Deduplication and Complexity Reduction

## Overview
This plan analyzes similar code patterns in `cmd/viro/` and proposes unification strategies to reduce duplication and complexity.

## Analysis Summary

### Files Analyzed
- `main.go` (92 lines) - Main entry point
- `config.go` (172 lines) - Configuration management
- `argparse.go` (47 lines) - Argument parsing
- `mode.go` (67 lines) - Mode detection
- `execution.go` (109 lines) - Execution orchestration
- `evaluator.go` (27 lines) - Evaluator setup
- `input.go` (56 lines) - Input sources
- `exit.go` (56 lines) - Error handling & exit codes
- `help.go` (85 lines) - Help text
- `version.go` (17 lines) - Version info

**Total LOC**: ~728 lines across 10 files

### Code Duplication Findings

## 1. Mode Detection Duplication ⭐ HIGH VALUE

**Issue**: Mode validation logic duplicated between `config.go` and `mode.go`

**Location**:
- `config.go:139-159` - Validate() counts modes
- `mode.go:35-66` - detectMode() counts modes

**Code Comparison**:
```go
// config.go:139-159
func (c *Config) Validate() error {
    modeCount := 0
    if c.ShowVersion { modeCount++ }
    if c.ShowHelp { modeCount++ }
    if c.EvalExpr != "" { modeCount++ }
    if c.CheckOnly { modeCount++ }
    if !c.CheckOnly && c.ScriptFile != "" { modeCount++ }
    
    if modeCount > 1 {
        return fmt.Errorf("multiple modes specified...")
    }
    // ... more validation
}

// mode.go:35-66
func detectMode(cfg *Config) (Mode, error) {
    modeCount := 0
    var detectedMode Mode
    
    modes := []struct{
        condition bool
        mode      Mode
    }{
        {cfg.ShowVersion, ModeVersion},
        {cfg.ShowHelp, ModeHelp},
        {cfg.EvalExpr != "", ModeEval},
        {cfg.CheckOnly, ModeCheck},
        {!cfg.CheckOnly && cfg.ScriptFile != "", ModeScript},
    }
    
    for _, m := range modes {
        if m.condition {
            modeCount++
            detectedMode = m.mode
        }
    }
    
    if modeCount > 1 {
        return ModeREPL, fmt.Errorf("multiple modes specified...")
    }
    // ...
}
```

**Duplication**: 
- Same error message in both places
- Same counting logic
- Same conditions checked

**Proposal**: Unify into single method
```go
// mode.go
func (c *Config) DetectMode() (Mode, error) {
    modes := []struct{
        condition bool
        mode      Mode
    }{
        {c.ShowVersion, ModeVersion},
        {c.ShowHelp, ModeHelp},
        {c.EvalExpr != "", ModeEval},
        {c.CheckOnly, ModeCheck},
        {!c.CheckOnly && c.ScriptFile != "", ModeScript},
    }
    
    var detectedMode Mode
    modeCount := 0
    
    for _, m := range modes {
        if m.condition {
            modeCount++
            detectedMode = m.mode
        }
    }
    
    if modeCount > 1 {
        return ModeREPL, fmt.Errorf("multiple modes specified; use only one of: --version, --help, -c, or script file")
    }
    
    if modeCount == 0 {
        return ModeREPL, nil
    }
    
    return detectedMode, nil
}

// config.go - Remove mode counting from Validate()
func (c *Config) Validate() error {
    // Only keep mode-specific validations
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

// main.go - Use unified method
func executeMode(cfg *Config) int {
    mode, err := cfg.DetectMode()  // Instead of detectMode(cfg)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        return ExitUsage
    }
    // ...
}
```

**Benefits**:
- ✅ Eliminates 20+ lines of duplicate logic
- ✅ Single source of truth for mode detection
- ✅ Easier to add new modes
- ✅ Clearer separation: Config.Validate() validates constraints, Config.DetectMode() detects mode

**Recommendation**: ⭐ **IMPLEMENT** - High value, low risk, cleaner architecture


## 2. Error Printing Duplication ⭐ MEDIUM VALUE

**Issue**: Similar error printing patterns scattered across files

**Locations**:
- `main.go:17` - `fmt.Fprintf(os.Stderr, "Configuration error: %v\n", err)`
- `main.go:45` - `fmt.Fprintf(os.Stderr, "Error: %v\n", err)`
- `main.go:61` - `fmt.Fprintf(os.Stderr, "Unknown mode: %v\n", mode)`
- `main.go:82` - `fmt.Fprintf(os.Stderr, "Error initializing REPL: %v\n", err)`
- `execution.go:59` - `fmt.Fprintf(os.Stderr, "Error loading input: %v\n", err)`
- `exit.go:26` - `fmt.Fprintf(os.Stderr, "%v", vErr)`
- `exit.go:30` - `fmt.Fprintf(os.Stderr, "Error: %v\n", err)`
- `exit.go:48-54` - printError() function

**Current State**: 
- `printError()` exists in exit.go but not consistently used
- Inconsistent formatting (sometimes "Error: %v", sometimes "%v", sometimes with context)

**Proposal**: Standardize error reporting
```go
// exit.go - Enhance existing functions
func printError(err error, context string) {
    if vErr, ok := err.(*verror.Error); ok {
        fmt.Fprintf(os.Stderr, "%v", vErr)
        return
    }
    
    if context != "" {
        fmt.Fprintf(os.Stderr, "%s: %v\n", context, err)
    } else {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
    }
}

// Usage examples:
// main.go
printError(err, "Configuration error")
printError(err, "Error initializing REPL")

// execution.go
printError(err, "Error loading input")
```

**Benefits**:
- ✅ Consistent error formatting across codebase
- ✅ Reduces ~8 duplicate fmt.Fprintf calls
- ✅ Easier to change error format globally

**Recommendation**: ⭐ **IMPLEMENT** - Good value, improves consistency


## 3. Config Field Organization 🟡 LOW VALUE

**Issue**: Config struct has 18 fields with unclear grouping

**Current State** (config.go:9-29):
```go
type Config struct {
    SandboxRoot      string  // Global
    AllowInsecureTLS bool    // Global
    Quiet            bool    // Global
    Verbose          bool    // Global
    
    ShowVersion bool        // Mode selection
    ShowHelp    bool        // Mode selection
    EvalExpr    string      // Mode: Eval
    CheckOnly   bool        // Mode: Check
    ScriptFile  string      // Mode: Script
    Args        []string    // Script/REPL args
    
    NoHistory   bool        // REPL only
    HistoryFile string      // REPL only
    Prompt      string      // REPL only
    NoWelcome   bool        // REPL only
    
    NoPrint   bool          // Eval only
    ReadStdin bool          // Eval only
}
```

**Proposal**: Group with embedded structs
```go
type Config struct {
    // Global settings
    SandboxRoot      string
    AllowInsecureTLS bool
    Quiet            bool
    Verbose          bool
    
    // Mode selection
    Mode ModeConfig
    
    // Script/Eval arguments
    Args []string
    
    // REPL settings
    REPL REPLConfig
}

type ModeConfig struct {
    ShowVersion bool
    ShowHelp    bool
    EvalExpr    string
    CheckOnly   bool
    ScriptFile  string
    NoPrint     bool
    ReadStdin   bool
}

type REPLConfig struct {
    NoHistory   bool
    HistoryFile string
    Prompt      string
    NoWelcome   bool
}
```

**Impact Analysis**:
- ⚠️ Requires updating all field references: `cfg.EvalExpr` → `cfg.Mode.EvalExpr`
- ⚠️ Affects ~50+ references across codebase
- ⚠️ All tests need updates (config_test.go has ~30 test cases)
- ✅ Better conceptual organization
- ✅ Clearer which fields affect which modes

**Recommendation**: ❌ **SKIP** - Low value, high churn, breaks tests unnecessarily


## 4. Environment Variable Loading Pattern 🟡 LOW VALUE

**Issue**: Repetitive if-statement pattern for env vars

**Current State** (config.go:39-52):
```go
func (c *Config) LoadFromEnv() error {
    if root := os.Getenv("VIRO_SANDBOX_ROOT"); root != "" {
        c.SandboxRoot = root
    }
    
    if tls := os.Getenv("VIRO_ALLOW_INSECURE_TLS"); tls == "1" || tls == "true" {
        c.AllowInsecureTLS = true
    }
    
    if history := os.Getenv("VIRO_HISTORY_FILE"); history != "" {
        c.HistoryFile = history
    }
    
    return nil
}
```

**Proposal**: Table-driven loader
```go
func (c *Config) LoadFromEnv() error {
    envMappings := []struct{
        envKey   string
        setter   func(string)
    }{
        {"VIRO_SANDBOX_ROOT", func(v string) { c.SandboxRoot = v }},
        {"VIRO_HISTORY_FILE", func(v string) { c.HistoryFile = v }},
        {"VIRO_ALLOW_INSECURE_TLS", func(v string) { 
            c.AllowInsecureTLS = v == "1" || v == "true" 
        }},
    }
    
    for _, m := range envMappings {
        if val := os.Getenv(m.envKey); val != "" {
            m.setter(val)
        }
    }
    
    return nil
}
```

**Impact Analysis**:
- ✅ 15 lines → 18 lines (3 line increase!)
- ✅ More complex with closures
- ⚠️ Harder to read for simple case
- ⚠️ No actual duplication (only 3 env vars)

**Recommendation**: ❌ **SKIP** - Overengineering, current code is clearer


## 5. ExecutionContext Construction 🟡 LOW VALUE

**Issue**: Repetitive switch cases for context creation

**Current State** (execution.go:22-50):
```go
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
```

**Proposal**: Factory function per mode
```go
func newCheckContext(cfg *Config) *ExecutionContext {
    return &ExecutionContext{
        Config:      cfg,
        Input:       &FileInput{Config: cfg, Path: cfg.ScriptFile},
        Args:        nil,
        PrintResult: false,
        ParseOnly:   true,
    }
}

func newEvalContext(cfg *Config) *ExecutionContext {
    return &ExecutionContext{
        Config:      cfg,
        Input:       &ExprInput{Expr: cfg.EvalExpr, WithStdin: cfg.ReadStdin},
        Args:        []string{},
        PrintResult: !cfg.NoPrint,
        ParseOnly:   false,
    }
}

func newScriptContext(cfg *Config) *ExecutionContext {
    return &ExecutionContext{
        Config:      cfg,
        Input:       &FileInput{Config: cfg, Path: cfg.ScriptFile},
        Args:        cfg.Args,
        PrintResult: false,
        ParseOnly:   false,
    }
}

func runExecution(cfg *Config, mode Mode) int {
    var ctx *ExecutionContext
    
    switch mode {
    case ModeCheck:
        ctx = newCheckContext(cfg)
    case ModeEval:
        ctx = newEvalContext(cfg)
    case ModeScript:
        ctx = newScriptContext(cfg)
    }
    
    _, exitCode := executeViroCode(ctx)
    return exitCode
}
```

**Impact Analysis**:
- ⚠️ 28 lines → 48 lines (71% increase)
- ⚠️ Adds indirection for no clear benefit
- ⚠️ Each mode context is only used once
- ⚠️ Current switch is already clear

**Recommendation**: ❌ **SKIP** - Increases LOC without improving clarity


## 6. Test Helper Duplication ⭐ LOW-MEDIUM VALUE

**Issue**: setupTestArgs() only in config_test.go but could be reused

**Current State**:
- `config_test.go:9-13` - setupTestArgs() helper
- Used in 8+ test cases in config_test.go
- Could be useful for other test files

**Proposal**: Extract to shared test utilities
```go
// testing_helpers.go (new file in cmd/viro/)
// +build test

package main

import (
    "flag"
    "os"
    "testing"
)

func setupTestArgs(t *testing.T, args []string) {
    t.Helper()
    flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
    os.Args = args
}

func setupTestEnv(t *testing.T, envVars map[string]string) func() {
    t.Helper()
    
    // Save old values
    oldVals := make(map[string]string)
    for k := range envVars {
        oldVals[k] = os.Getenv(k)
    }
    
    // Set new values
    for k, v := range envVars {
        os.Setenv(k, v)
    }
    
    // Return cleanup function
    return func() {
        for k, v := range oldVals {
            os.Setenv(k, v)
        }
    }
}
```

**Benefits**:
- ✅ Reusable across test files
- ✅ setupTestEnv eliminates manual defer cleanup in tests
- ✅ Could reduce ~20 lines in config_test.go (lines 26-33 pattern repeated 7 times)

**Recommendation**: 🟡 **OPTIONAL** - Nice to have, minor improvement


## 7. Flag Definition Duplication 🟡 LOW VALUE

**Issue**: flagsWithValues map hardcodes flag names that are also in LoadFromFlags()

**Locations**:
- `argparse.go:5-10` - flagsWithValues map
- `config.go:55-74` - flag.String() definitions

**Duplication**: Flag names like "-c", "--sandbox-root", "--history-file" appear in both

**Proposal**: Generate flagsWithValues from flag definitions
```go
// config.go
type flagDef struct {
    name     string
    hasValue bool
    // ... other flag metadata
}

var flags = []flagDef{
    {name: "-c", hasValue: true},
    {name: "--sandbox-root", hasValue: true},
    {name: "--history-file", hasValue: true},
    {name: "--prompt", hasValue: true},
    {name: "--version", hasValue: false},
    // ...
}

func getFlagsWithValues() map[string]bool {
    m := make(map[string]bool)
    for _, f := range flags {
        if f.hasValue {
            m[f.name] = true
        }
    }
    return m
}
```

**Impact Analysis**:
- ⚠️ Only 4 flags with values currently
- ⚠️ flagsWithValues is for arg parsing, not flag definition
- ⚠️ Would tightly couple argparse and config packages
- ⚠️ Adds complexity to simple lookup table

**Recommendation**: ❌ **SKIP** - Not worth the coupling for 4 items


## Summary of Recommendations

### ⭐ HIGH PRIORITY - Implement
1. **Unify mode detection** (Proposal #1)
   - Impact: -20 LOC, better architecture
   - Effort: 2-3 hours
   - Risk: Low (good test coverage)

2. **Standardize error printing** (Proposal #2)
   - Impact: -8 duplications, consistency
   - Effort: 1 hour
   - Risk: Very low

### 🟡 OPTIONAL - Consider if time permits
3. **Test helper utilities** (Proposal #6)
   - Impact: -20 LOC in tests, better reusability
   - Effort: 1 hour
   - Risk: Very low

### ❌ SKIP - Not worth implementing
4. Config field organization (Proposal #3) - High churn, low value
5. Env var table-driven loader (Proposal #4) - Overengineering
6. ExecutionContext factories (Proposal #5) - Increases complexity
7. Flag definition deduplication (Proposal #7) - Premature abstraction

## Implementation Priority

**Phase 1: High Value Changes**
1. Unify mode detection logic (#1)
2. Standardize error printing (#2)

**Phase 2: Optional Improvements** (if time permits)
3. Extract test helpers (#6)

**Total Estimated Effort**: 3-4 hours for Phase 1, +1 hour for Phase 2

## Metrics

### Before
- Total LOC: ~728
- Duplicated logic: ~30 lines
- Error print sites: 8
- Test helper copies: 2

### After (Phase 1)
- Total LOC: ~700 (-28 lines, -3.8%)
- Duplicated logic: ~2 lines (-93%)
- Error print sites: 1 (centralized)
- Improved maintainability score

### After (Phase 1 + 2)
- Total LOC: ~680 (-48 lines, -6.6%)
- Test LOC: -20 lines
- Reusable test utilities: 2 helpers

## Risk Assessment

**Phase 1 Risks**: ⚠️ LOW
- Good test coverage exists (config_test.go, mode_test.go, exit_test.go)
- Changes are refactorings, not behavioral changes
- All existing tests should pass without modification

**Phase 2 Risks**: ⚠️ VERY LOW
- Test-only changes
- No impact on production code

## Conclusion

**Implement Phase 1** - Clear value proposition with low risk. The codebase is already well-organized; the main issues are the duplicated mode detection logic and inconsistent error printing. Fixing these two issues will improve maintainability without introducing unnecessary abstraction.

**Skip most other proposals** - They either increase complexity without clear benefit, or have high churn for low value.

The `cmd/viro` package is generally well-structured. The biggest wins come from eliminating the mode detection duplication and standardizing error handling patterns.
