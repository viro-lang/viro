# Agent Guidelines for Viro

## Build & Test Commands
- Build: `go build -o viro ./cmd/viro` or `make build`
- Test all: `go test ./...` or `make test`
- Test package: `go test -v ./test/contract/...` or `go test -v ./internal/native/...`
- Single test: `go test -v ./test/contract -run TestNativeAdd`
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
