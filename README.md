# Viro

Viro is a homoiconic programming language interpreter implemented in Go, featuring a type-based dispatch system and an interactive REPL.

## Overview

Viro's key features:
- **Left-to-right evaluation** (no operator precedence)
- **Local-by-default scoping** (safe, predictable variable scoping)
- **Bash-style refinements** (`--flag`, `--option value`)
- **Paren type for immediate evaluation** distinct from deferred blocks

## Feature 002: Deferred Language Capabilities

**Status**: In Progress

This release adds advanced capabilities deferred from the initial implementation:
- **High-precision decimals** (IEEE 754 decimal128) for financial and scientific calculations
- **Sandboxed ports** for file and network I/O (HTTP, TCP) with TLS controls
- **Objects and paths** for structured data organization and nested access
- **Parse dialect** for declarative pattern matching and data transformation
- **Observability** including tracing, debugging, and reflection capabilities

See [specs/002-implement-deferred-features/](specs/002-implement-deferred-features/) for detailed specifications.

## Quick Start

See [specs/001-implement-the-core/quickstart.md](specs/001-implement-the-core/quickstart.md) for detailed build, run, and test instructions.

### Build

```bash
go build -o viro ./cmd/viro
```

### Run REPL

```bash
./viro
```

### Example Session

```viro
>> 42
42
>> "hello"
"hello"
>> x: 10
10
>> x
10
>> 3 + 4 * 2
14
>> 3 + (4 * 2)
11
>> [1 2 3]
[1 2 3]
>> exit
Goodbye!
```

### Run Tests

```bash
go test ./...
```

## Project Structure

- `internal/value/` - Value types (integer, string, word, block, function)
- `internal/eval/` - Core evaluator with type-based dispatch
- `internal/stack/` - Unified stack for data and frames
- `internal/frame/` - Frame and context system for scoping
- `internal/native/` - Native function implementations
- `internal/error/` - Structured error handling
- `internal/parse/` - Parser with left-to-right evaluation
- `cmd/viro/` - CLI entry point and REPL
- `test/contract/` - Contract tests for native functions
- `test/integration/` - End-to-end interpreter tests

## Features

### Phase 1: Core Interpreter ✅ COMPLETED

**User Story 1: Basic Expression Evaluation** ✅
- ✅ Literal evaluation (integers, strings, true/false/none)
- ✅ Variable binding and retrieval (set-word, word)
- ✅ Arithmetic operations (+, -, *, /) with overflow detection
- ✅ Left-to-right evaluation (no operator precedence)
- ✅ Block and paren evaluation
- ✅ Interactive REPL with readline support
- ✅ Parser with left-to-right evaluation
- ✅ 45 contract tests passing

**Remaining for Phase 1:**
- User Story 2: Control flow (when, if, loop, while)
- User Story 3: Series operations (first, last, append, insert, length?)
- User Story 4: User-defined functions with refinements
- User Story 5: I/O operations (print, input, read, write)
- User Story 6: Error handling (try/catch, throw)

### Future Phases (Planned)

- Parse dialect
- Object system
- Module system
- File I/O
- Network operations
- Decimal/floating-point arithmetic

## Design Principles

1. **TDD (Non-Negotiable)** - Tests written before implementation
2. **Incremental Layering** - Architecture built layer by layer
3. **Type Dispatch Fidelity** - Type-based evaluation semantics
4. **Stack and Frame Safety** - Index-based access prevents pointer invalidation
5. **Structured Errors** - Category-based errors with diagnostic context
6. **Observable Behavior** - REPL provides clear feedback
7. **YAGNI** - Minimal feature set, no premature optimization

## Technology Stack

- **Language**: Go 1.21+
- **REPL Library**: github.com/chzyer/readline
- **Testing**: Go standard library (table-driven tests)
- **Platform**: macOS primary, Linux/Windows compatible

## Documentation

- [Specification](specs/001-implement-the-core/spec.md)
- [Implementation Plan](specs/001-implement-the-core/plan.md)
- [Data Model](specs/001-implement-the-core/data-model.md)
- [Quickstart Guide](specs/001-implement-the-core/quickstart.md)
- [Native Function Contracts](specs/001-implement-the-core/contracts/)

## License

See LICENSE file for details.

## Contributing

This project follows strict TDD methodology. All contributions must include tests written before implementation.

See [specs/001-implement-the-core/tasks.md](specs/001-implement-the-core/tasks.md) for the complete task breakdown.
