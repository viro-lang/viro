# Plan 022: RuntimeContext Deduplication

## Problem

The `RuntimeContext` type is duplicated between `cmd/viro/run.go` and `internal/api/api.go`. This violates DRY principles and creates a maintenance burden where changes must be mirrored in both locations.

### Current State

**cmd/viro/run.go (lines 17-22):**
```go
type RuntimeContext struct {
	Args   []string
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
}
```

**internal/api/api.go (lines 25-30):**
```go
type RuntimeContext struct {
	Args   []string
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
}
```

### Why Duplication Exists

- Originally, `cmd/viro` had its own implementation
- `internal/api` package was created to allow integration tests to invoke Viro programmatically
- `RuntimeContext` was duplicated to avoid circular dependencies
- Both definitions are identical (4 fields)

### Current Usage

**cmd/viro/run.go:**
- Defines `RuntimeContext` struct
- `Run(ctx *RuntimeContext) int` function uses it
- Various helper functions accept `*RuntimeContext` parameter

**internal/api/api.go:**
- Defines identical `RuntimeContext` struct
- `Run(ctx *RuntimeContext, cfg *Config) int` function uses it
- Various helper functions accept `*RuntimeContext` parameter

**Test files:**
- Integration tests import `internal/api` and use `api.RuntimeContext`
- Tests cannot import `cmd/viro` (it's a main package)

## Solution

Keep `RuntimeContext` in `internal/api` package (single definition) and have `cmd/viro` import it.

### Rationale

1. `internal/api` is designed to be importable by both `cmd` and tests
2. Following the pattern established by `Config` deduplication (Plan 021)
3. `RuntimeContext` is a pure data structure with no behavior
4. `cmd/viro` can import `internal/api` without creating circular dependencies

### Architecture

```
internal/api/
  api.go         - Contains RuntimeContext definition (keep)

cmd/viro/
  run.go         - Import api.RuntimeContext (change)
  main.go        - Use api.RuntimeContext (change if needed)

test/integration/
  *_test.go      - Already use api.RuntimeContext (no change)
```

## Implementation Steps

### Phase 1: Update cmd/viro

1. Update `cmd/viro/run.go`:
   - Add import: `"github.com/marcin-radoszewski/viro/internal/api"`
   - Remove local `RuntimeContext` struct definition (lines 17-22)
   - Change all `*RuntimeContext` references to `*api.RuntimeContext`
   - Verify all function signatures using RuntimeContext

2. Update `cmd/viro/main.go` if needed:
   - Check if `RuntimeContext` is used
   - Update to `api.RuntimeContext` if necessary

### Phase 2: Verify

1. Build test: `go build ./cmd/viro`
2. Run full test suite: `go test ./...`
3. Manual CLI testing:
   - `./viro --version`
   - `./viro -c "3 + 4"`
   - `./viro examples/01_basics.viro`

## Benefits

1. **Single source of truth** - RuntimeContext defined once in `internal/api`
2. **Eliminates duplication** - No need to maintain two identical structs
3. **Consistent API** - Tests and cmd use same type definition
4. **Easier maintenance** - Changes made in one location
5. **Follows established pattern** - Mirrors Config deduplication approach (Plan 021)

## Risks & Mitigations

**Risk: Import cycle**
- Mitigation: `cmd/viro` → `internal/api` is valid (cmd can import internal packages)
- Mitigation: `internal/api` does not import `cmd/viro` (cannot import main packages)

**Risk: Breaking existing behavior**
- Mitigation: `RuntimeContext` is identical in both locations
- Mitigation: Only changing import location, not structure or behavior
- Mitigation: Comprehensive test suite verifies compatibility

## Alternative Considered

**Move RuntimeContext to separate internal/context package:**
- Rejected: Over-engineering for a simple 4-field struct
- Rejected: Creates unnecessary package proliferation
- Rejected: `internal/api` is already the public API for programmatic access

**Keep duplication, add comments:**
- Rejected: Doesn't solve maintenance burden
- Rejected: Still violates DRY principles

## Success Criteria

1. Zero duplication - RuntimeContext defined once in `internal/api`
2. All tests pass - no behavioral regressions
3. CLI works correctly - all modes function as before
4. Code is cleaner - eliminated duplicate struct definition

## Implementation Status

### ✅ COMPLETED - All Phases Finished

**Phase 1: Update cmd/viro** - ✅ Complete
- Updated `cmd/viro/run.go`:
  - Added import: `"github.com/marcin-radoszewski/viro/internal/api"`
  - Removed local `RuntimeContext` struct definition (lines 17-22)
  - Changed all function signatures from `*RuntimeContext` to `*api.RuntimeContext`:
    - `Run(ctx *api.RuntimeContext)`
    - `loadConfigurationWithContext(ctx *api.RuntimeContext)`
    - `executeModeWithContext(cfg *config.Config, ctx *api.RuntimeContext)`
    - `runREPLWithContext(cfg *config.Config, ctx *api.RuntimeContext)`
    - `runExecutionWithContext(cfg *config.Config, mode config.Mode, ctx *api.RuntimeContext)`
    - `executeViroCodeWithContext(..., ctx *api.RuntimeContext)`
    - `setupEvaluatorWithContext(cfg *config.Config, ctx *api.RuntimeContext)`
- Updated `cmd/viro/main.go`:
  - Added import: `"github.com/marcin-radoszewski/viro/internal/api"`
  - Changed `&RuntimeContext{...}` to `&api.RuntimeContext{...}`

**Phase 2: Verification** - ✅ Complete
- ✅ Build successful: `go build ./cmd/viro`
- ✅ All tests pass: `go test ./...`
- ✅ CLI manual testing:
  - `./viro --version` → "Viro 0.1.0"
  - `./viro -c "3 + 4"` → "7"
  - `./viro examples/01_basics.viro` → script execution works

### Results

**Success Criteria Met:**
1. ✅ Zero duplication - RuntimeContext in `internal/api` only
2. ✅ All tests pass - no regressions
3. ✅ CLI works correctly - all modes functional
4. ✅ Code is cleaner - eliminated duplicate struct definition

**Final State:**
- Single source of truth: `RuntimeContext` defined in `internal/api/api.go` (lines 25-30)
- `cmd/viro` imports and uses `api.RuntimeContext`
- Tests continue to use `api.RuntimeContext` (no changes needed)
- Clean architecture: follows pattern from Config deduplication (Plan 021)

**Completed:** 2025-11-01
