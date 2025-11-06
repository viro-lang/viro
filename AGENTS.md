# Agent Guidelines for Viro

## Build & Test Commands

- Build: `make build` or `go build -o viro ./cmd/viro`
- Test all: `go test ./...` or `make test`
- Test summary: `make test-summary` (shows total, passed, and failed test counts)
- Test package: `go test -v ./test/contract/...` or `go test -v ./internal/native/...`
- Single test: `go test -v ./test/contract -run TestNativeAdd`
- Test with JSON output: `go test -json ./... | jq` (structured output for better analysis)
- Coverage: `go test -coverprofile=coverage.out ./...`

## CLI Usage

The `viro` binary supports multiple execution modes:

### Modes

- **REPL** (default): `./viro` - Interactive Read-Eval-Print Loop
- **REPL with args**: `./viro -- arg1 arg2` - REPL with `system.args` populated
- **Script execution**: `./viro script.viro arg1 arg2` - Execute file with arguments
- **Expression eval**: `./viro -c "3 + 4"` - Evaluate and print result
- **Syntax check**: `./viro --check script.viro` - Parse only, no execution
- **Version**: `./viro --version` - Show version information
- **Help**: `./viro --help` - Show usage information

### Global Options

- `--sandbox-root PATH` - Sandbox root for file operations (default: current directory)
- `--allow-insecure-tls` - Disable TLS certificate verification (security risk!)
- `--quiet` - Suppress non-error output
- `--verbose` - Enable verbose output

### Eval Mode Options (with `-c`)

- `--stdin` - Read additional input from stdin: `echo "[1 2 3]" | viro -c "first" --stdin`
- `--no-print` - Don't print evaluation result: `viro -c "pow 2 10" --no-print`

### REPL Mode Options

- `--no-history` - Disable command history
- `--history-file PATH` - Custom history file location
- `--prompt STRING` - Custom REPL prompt
- `--no-welcome` - Skip welcome message

### Environment Variables

- `VIRO_SANDBOX_ROOT` - Default sandbox root directory
- `VIRO_ALLOW_INSECURE_TLS` - Allow insecure TLS (set to "1" or "true")
- `VIRO_HISTORY_FILE` - REPL history file location

### Exit Codes

- `0` - Success
- `1` - General error (script/math error)
- `2` - Syntax error (parse failure)
- `3` - Access error (permission denied, sandbox violation)
- `64` - Usage error (invalid CLI arguments)
- `70` - Internal error (interpreter crash)
- `130` - Interrupted (Ctrl+C)

## Code Style

- **NO COMMENTS** in code; documentation belongs in package docs only
- Use constructor functions: `value.IntVal()`, `value.StrVal()`, never direct struct creation
- Index-based refs: Frame.Parent is `int` index, NOT pointer (prevents stack expansion bugs)
- Table-driven tests: Always use `tests := []struct{name, args, want, wantErr}`
- Errors: Use `verror.NewScriptError()`, `verror.NewMathError()` with category/ID/args
- Imports: Group stdlib → external → internal, alphabetically within groups
- Naming: Use Viro-style native names (`first`, `length?`, `type-of` with ?, ! suffixes) not Go-style

## Workflow

- **ALWAYS use viro-coder agent**: When editing ANY Viro interpreter code, you MUST use the viro-coder agent via the Task tool. This agent has specialized expertise in the Viro codebase architecture and prevents common mistakes.
- **MANDATORY code review process**: After the viro-coder agent finishes updates, you MUST use the viro-reviewer agent to review the code. The main agent may then decide whether to apply the suggested changes. If changes are applied, ALWAYS use the viro-coder agent again (never edit directly).
- **TDD mandatory**: Write tests FIRST in `test/contract/`, then implement in `internal/native/`
- Consult specs: `specs/*/contracts/*.md` before implementation
- Every code change MUST have test coverage
- **Prefer automated tests** over manual execution
- For manual testing, use CLI modes:
  - Quick syntax check: `./viro --check script.viro`
  - One-off eval: `./viro -c "expression"`
  - Script with args: `./viro script.viro arg1 arg2`
  - Pipeline testing: `echo "data" | ./viro -c "expression" --stdin`
- No real network calls in tests; use 127.0.0.1 mocked servers only

## Planning

- Store plans in `plans/` directory with sequential numbering: `001_description.md`, `002_description.md`, etc.
- Check existing plans to determine the next sequence number

## Debugging

### Quick Start

Use enhanced trace system for LLM-friendly debugging:

```bash
# In script file
./viro -c 'trace --on --verbose --include-args --step-level 1'
./viro script.viro

# Or inline
./viro -c 'do [trace --on --verbose --include-args --step-level 1  your-code  trace --off]'
```

### CLI-Based Debugging Workflow

1. **Syntax check first**: `./viro --check script.viro` - Validates parse without execution
2. **Expression testing**: `./viro -c "expression"` - Quick REPL-less evaluation
3. **Script with trace**: Add trace commands to script, run with `./viro script.viro`
4. **Verbose mode**: `./viro --verbose script.viro` - CLI-level diagnostic output
5. **Exit code testing**: Check `$?` after execution to verify error handling

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
- **Pipeline debugging**: Use `-c` to wrap code: `./viro -c "do [trace --on --verbose  your-code  trace --off]"`

## Documentation

### Quick Reference for Viro Development

For rapid understanding of Viro's architecture and core concepts:

- **[Viro Architecture Knowledge Base](/docs/viro_architecture_knowledge.md)** - Comprehensive guide covering:
  - Language design principles and core concepts
  - Interpreter architecture and key components
  - Value system and type-based dispatch
  - Native functions implementation patterns
  - Error handling and extension guidelines
  - Performance characteristics and testing strategies

- **[Viro Core Knowledge RAG](/docs/viro_core_knowledge_rag.md)** - Searchable knowledge base with:
  - FAQ-style explanations of common concepts
  - Code examples and usage patterns
  - Integration with other tools and systems
  - Troubleshooting guides and best practices

**When to use these documents**:
- Understanding the codebase before making changes
- Implementing new native functions or language features
- Debugging complex issues requiring architectural knowledge
- Learning Viro's unique design patterns and conventions