# Agent Guidelines for Viro

## Build & Test Commands
- Generate grammar: `make grammar` or `pigeon -o internal/parse/peg/parser.go grammar/viro.peg`
- Build: `make build` (includes grammar generation) or `go build -o viro ./cmd/viro`
- Test all: `go test ./...` or `make test`
- Test summary: `make test-summary` (shows total, passed, and failed test counts)
- Test package: `go test -v ./test/contract/...` or `go test -v ./internal/native/...`
- Single test: `go test -v ./test/contract -run TestNativeAdd`
- Test with JSON output: `go test -json ./... | jq` (structured output for better analysis)
- Coverage: `go test -coverprofile=coverage.out ./...`

## Code Style
- **NO COMMENTS** in code; documentation belongs in package docs only
- Use constructor functions: `value.IntVal()`, `value.StrVal()`, never direct struct creation
- Index-based refs: Frame.Parent is `int` index, NOT pointer (prevents stack expansion bugs)
- Table-driven tests: Always use `tests := []struct{name, args, want, wantErr}`
- Errors: Use `verror.NewScriptError()`, `verror.NewMathError()` with category/ID/args
- Imports: Group stdlib → external → internal, alphabetically within groups
- Naming: Use REBOL-style native names (`first`, `length?`, `type-of`) not Go-style

## Workflow
- **TDD mandatory**: Write tests FIRST in `test/contract/`, then implement in `internal/native/`
- Consult specs: `specs/*/contracts/*.md` before implementation
- Every code change MUST have test coverage
- Automated tests ONLY; avoid running `./viro` manually
- No real network calls in tests; use 127.0.0.1 mocked servers only

## Planning
- Store plans in `plans/` directory with sequential numbering: `001_description.md`, `002_description.md`, etc.
- Check existing plans to determine the next sequence number

## Debugging

### Quick Start
Use enhanced trace system for LLM-friendly debugging:
```viro
trace --on --verbose --include-args --step-level 1
; Your code here
trace --off
```

### Complete Instructions
**IMPORTANT**: For detailed debugging workflow, trace output format, parsing examples, and troubleshooting, see:
- **[LLM Debugging Instructions](/.github/instructions/debugging-with-trace.instruction.md)** - Complete guide for LLM agents
- [Debugging Guide](/docs/debugging-guide.md) - User-facing documentation
- [Debugging Examples](/docs/debugging-examples.md) - Practical scenarios

### Common Patterns
- **Infinite recursion**: Use `--max-depth 10` and check depth progression
- **Variable tracking**: Use `--verbose` to see frame state changes
- **Function arguments**: Use `--include-args` to verify parameter values
- **Performance**: Use `--step-level 0` and analyze duration field
- **Parse JSON output**: Write Python/JS scripts to analyze trace events
