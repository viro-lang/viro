# Plan 023: Complete Execution Logic Deduplication

## Context
Plan 022 successfully eliminated the duplicate `RuntimeContext` type definition between `cmd/viro/run.go` and `internal/api/api.go`. However, approximately 150-200 lines of execution logic remain duplicated between these files. This plan completes the deduplication effort to establish `internal/api` as the single source of truth for ALL execution logic.

## Issues Identified

### Priority 1: Function Duplication (100% Duplicates)
The following functions are identically duplicated:
- `executeViroCodeWithContext()` - run.go:173-207 vs api.go:212-247
- `setupEvaluatorWithContext()` - run.go:210-229 vs api.go:249-269
- `handleErrorWithContext()` - run.go:232-242 vs api.go:288-298
- `printErrorToWriter()` - run.go:244-252 vs api.go:300-308
- `categoryToExitCode()` - exit.go:34-45 vs api.go:310-321
- `ExprInputWithContext` type - run.go:153-171 vs api.go:57-75

### Priority 2: InputSource Type Consolidation
Multiple implementations exist:
- `cmd/viro/run.go`: `ExprInputWithContext` (with `Stdin io.Reader`)
- `cmd/viro/input.go`: `ExprInput` (hardcoded `os.Stdin`) and `FileInput` (hardcoded `os.Stdin`)
- `internal/api/api.go`: Better versions with `Stdin io.Reader` for testability

### Priority 3: Exit Code Consolidation
Exit codes defined in two places:
- `cmd/viro/exit.go` (lines 10-18)
- `internal/api/api.go` (lines 323-331)

### Priority 4: Dead Code Elimination
Files containing potentially obsolete code:
- `cmd/viro/execution.go` - Old `runExecution()` and `executeViroCode()` superseded by `*WithContext` functions
- `cmd/viro/input.go` - Superseded by `api.FileInput` and `api.ExprInputWithContext`
- `cmd/viro/evaluator.go` - Old `setupEvaluator()` superseded by `setupEvaluatorWithContext()`
- `cmd/viro/exit.go` - Functions duplicated in `internal/api`

### Priority 5: Simplify cmd/viro/run.go
After deduplication, `cmd/viro/run.go` should become a thin wrapper.

## Target Architecture
```
internal/api/         [Single source of truth for ALL execution logic]
       ↑
       |
cmd/viro/            [Thin CLI wrapper - arg parsing, calling api functions]
  ├── main.go        [Entry point, signal handling]
  ├── run.go         [Mode routing, calls api functions]
  ├── help.go        [CLI help text]
  └── version.go     [Version info]
```

## Implementation Steps

### Step 1: Remove Dead Files
Delete obsolete files that are no longer used:
- [x] `cmd/viro/execution.go` - Superseded by `api.RunExecutionWithContext()`
- [x] `cmd/viro/input.go` - Superseded by `api.FileInput` and `api.ExprInputWithContext`
- [x] `cmd/viro/evaluator.go` - Superseded by `api.setupEvaluatorWithContext()`
- [x] `cmd/viro/exit.go` - Superseded by `api` exit codes and functions

### Step 2: Update cmd/viro/main.go to Use API Exit Codes
- [x] Replace `ExitInterrupt` with `api.ExitInterrupt`

### Step 3: Refactor cmd/viro/run.go
- [x] Remove all duplicated function implementations
- [x] Remove `ExprInputWithContext` type definition
- [x] Import and use only `api` functions
- [x] Use `api.NewFileInput()` for file inputs
- [x] Use `api.ExprInputWithContext` for expression inputs
- [x] Call `api.RunExecutionWithContext()` instead of local implementation

### Step 4: Export Additional API Functions
Ensure all needed functions are exported from `internal/api/api.go`:
- [x] `RunExecutionWithContext()` - Already exists as `runExecutionWithContext()`
- [x] Exit code constants - Already exported
- [x] Input types - Already exported

### Step 5: Verify Tests Pass
- [x] Run `make test` to ensure no regressions
- [x] Test all CLI modes (REPL, script, eval, check)
- [x] Verify exit codes are correct

## Expected Outcome
- Eliminate ~150-200 lines of duplicate code
- Achieve clear separation: CLI (cmd/viro) vs core logic (internal/api)
- Establish single source of truth for execution logic
- Maintain all existing functionality and test coverage
- `cmd/viro/` reduced from 10 files to 4-5 files

## Testing Strategy
1. Automated tests: `make test`
2. CLI mode testing:
   - REPL: `./viro`
   - Script: `./viro examples/01_basics.viro`
   - Eval: `./viro -c "3 + 4"`
   - Check: `./viro --check examples/01_basics.viro`
3. Exit code verification: `echo $?` after each command
4. Integration tests in `test/integration/`

## Success Criteria
- [x] All duplicate functions removed from `cmd/viro/`
- [x] All duplicate types removed from `cmd/viro/`
- [x] All tests pass
- [x] All CLI modes work correctly
- [x] Exit codes preserved
- [x] Code size reduced by ~150-200 lines

## Status: COMPLETED

All steps completed successfully. The architectural goal has been achieved with `internal/api` as the single source of truth and `cmd/viro` as a thin CLI wrapper.

## Results

### Files Deleted (5 total, 373 lines)
- `cmd/viro/execution.go` - 150 lines (old execution logic)
- `cmd/viro/input.go` - 57 lines (duplicate InputSource types)
- `cmd/viro/evaluator.go` - 27 lines (duplicate evaluator setup)
- `cmd/viro/exit.go` - 55 lines (duplicate exit codes and functions)
- `cmd/viro/exit_test.go` - 84 lines (tests moved to integration)

### Files Modified
- `cmd/viro/run.go` - Reduced from 253 lines to 87 lines (-166 lines)
  - Removed all duplicate function implementations
  - Removed `ExprInputWithContext` type definition
  - Now calls `api.RunExecutionWithContext()` for script/eval/check modes
  - Uses `api` exit codes throughout
  
- `cmd/viro/main.go` - Changed signal handler to use `api.ExitInterrupt`

- `internal/api/api.go` - Exported 2 functions (+4 lines)
  - `RunExecutionWithContext()` - was `runExecutionWithContext()`
  - `HandleErrorWithContext()` - was `handleErrorWithContext()`

### Total Impact
- **Lines removed: 549**
- **Lines added: 11**
- **Net reduction: 538 lines**
- **Files reduced: From 10 to 5 in cmd/viro/**

### Architecture Achievement
```
internal/api/api.go (331 lines)
  ├── RuntimeContext type
  ├── InputSource types (FileInput, ExprInputWithContext)
  ├── Exit code constants
  ├── RunExecutionWithContext() - orchestrates execution
  ├── executeViroCodeWithContext() - core execution logic
  ├── setupEvaluatorWithContext() - evaluator setup
  ├── HandleErrorWithContext() - error handling
  ├── printErrorToWriter() - error formatting
  └── categoryToExitCode() - error category mapping

cmd/viro/ (244 lines across 5 files)
  ├── main.go (32 lines) - Entry point, signal handling
  ├── run.go (87 lines) - Config loading, mode routing, REPL setup
  ├── help.go (95 lines) - CLI help text
  ├── help_test.go (11 lines) - Basic tests
  └── version.go (19 lines) - Version info
```

### Test Results
- All 19 test packages pass
- Integration tests verify exit codes (syntax=2, error=1, success=0)
- CLI modes tested and working:
  - REPL: `./viro`
  - Script: `./viro script.viro`
  - Eval: `./viro -c "expr"`
  - Check: `./viro --check script.viro`
  - Version: `./viro --version`
  - Help: `./viro --help`

### Key Benefits
1. **Single Source of Truth**: All execution logic now in `internal/api`
2. **Clean Separation**: CLI layer is now just argument parsing and mode routing
3. **Maintainability**: No more duplicate code to keep in sync
4. **Testability**: API functions can be tested independently
5. **Reusability**: API can be used by other Go programs, not just CLI
