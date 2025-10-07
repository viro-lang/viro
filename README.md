# Viro - REBOL-Inspired Interpreter

Viro is a REBOL R3-inspired interpreter implemented in Go, featuring a type-based dispatch system and an interactive REPL.

## Overview

Viro implements a subset of REBOL's evaluation semantics with modern improvements:
- **Traditional operator precedence** (not left-to-right)
- **Local-by-default scoping** (safer than REBOL's global-by-default)
- **Bash-style refinements** (`--flag`, `--option value`)
- **Paren type for immediate evaluation** distinct from deferred blocks

## Quick Start

See [specs/001-implement-the-core/quickstart.md](specs/001-implement-the-core/quickstart.md) for detailed build, run, and test instructions.

### Build

```bash
go build -o viro cmd/viro/main.go
```

### Run REPL

```bash
./viro
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
- `internal/parse/` - Parser with operator precedence
- `cmd/viro/` - CLI entry point and REPL
- `test/contract/` - Contract tests for native functions
- `test/integration/` - End-to-end interpreter tests

## Features

### Phase 1: Core Interpreter ✅ (In Progress)

- Basic expression evaluation (literals, arithmetic, words)
- Control flow (when, if, loop, while)
- Series operations (first, last, append, insert, length?)
- User-defined functions with refinements
- Structured error handling with context
- Interactive REPL with command history

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
3. **Type Dispatch Fidelity** - Type-based evaluation per REBOL semantics
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
