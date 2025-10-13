# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Viro is a REBOL-inspired language interpreter implemented in Go. **It is NOT a REBOL interpreter** - it draws inspiration from REBOL's design but implements its own semantics with modern improvements. Key differences and features:
- Type-based dispatch system with left-to-right evaluation (no operator precedence)
- Local-by-default scoping (safer than REBOL's global-by-default)
- Bash-style refinements (`--flag`, `--option value`)
- Distinction between blocks `[...]` (deferred) and parens `(...)` (immediate evaluation)

## Essential Commands

### Building and Running
```bash
# Build the interpreter
go build -o viro ./cmd/viro
make build  # Alternative using Makefile

# Run REPL
./viro

# Run all tests
go test ./...
make test

# Run specific test package
go test -v ./test/contract/...
go test -v ./internal/native/...

# Run single test
go test -v ./test/contract -run TestNativeAdd

# Run with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Development Workflow
- **TDD is non-negotiable**: Always write tests BEFORE implementation
- **Contract tests first**: Define behavior in `specs/*/contracts/`, then implement in `test/contract/`
- **Automated tests preferred**: Avoid running `./viro` manually; use Go tests instead

## Architecture

### Core Package Structure

```
internal/
├── value/       - Value types (TypeInteger, TypeString, TypeBlock, etc.)
│                 All data is represented as Value with type tag + payload
├── eval/        - Core evaluator with type-based dispatch
├── stack/       - Unified stack for data and frames (index-based, not pointer-based)
├── frame/       - Frame and context system (local-by-default scoping)
├── native/      - Native function implementations and Registry
├── verror/      - Structured error system with categories (Syntax, Script, Math, etc.)
├── parse/       - Parser and dialect engine
└── repl/        - REPL implementation
```

### Key Design Principles

1. **Type-Based Dispatch**: All values have a Type field that determines evaluation behavior
   - Literals evaluate to themselves
   - Words evaluate to bound values
   - Functions execute with arguments
   - Blocks evaluate to themselves (deferred)
   - Parens evaluate their contents (immediate)

2. **Index-Based Access**: Stack and Frame use integer indices, NOT pointers
   - Prevents invalidation during stack expansion
   - All frame references use indices into stack

3. **Local-by-Default Scoping**: Function parameters and body words are automatically local
   - Words in function body create local variables
   - Global access requires explicit capture before function definition
   - Example: `x: 100; test: fn [] [x: 5 x]; test; x` → x still 100

4. **Structured Errors**: Category-based errors with diagnostic context
   - Categories: Throw(0), Note(100), Syntax(200), Script(300), Math(400), Access(500), Internal(900)
   - Each error includes: code, ID, args, near context, where context

5. **Left-to-Right Evaluation**: No operator precedence (REBOL-style)
   - `3 + 4 * 2` → 14 (not 11)
   - Use parens to group: `3 + (4 * 2)` → 11

### Value Type System

All data is represented using `value.Value` with discriminated union:

```go
type Value struct {
    Type    ValueType   // TypeNone, TypeInteger, TypeString, TypeBlock, etc.
    Payload interface{} // Type-specific data
}
```

**Constructor functions** (use these, never direct struct creation):
- `value.NoneVal()`, `value.LogicVal(bool)`, `value.IntVal(int64)`
- `value.StrVal(string)`, `value.BlockVal([]Value)`, `value.ParenVal([]Value)`
- `value.WordVal(symbol)`, `value.FuncVal(*FunctionValue)`

**Type assertions** (safe extraction):
- `val.AsInteger()`, `val.AsLogic()`, `val.AsString()`, `val.AsBlock()`, `val.AsFunction()`

### Native Function System

Native functions are stored in the **root frame** (frame index 0) and resolved via standard lexical scoping. This enables:
- User code to shadow natives with local bindings
- Consistent word resolution strategy (no special cases)
- Standard frame chain traversal for all lookups

Each native function:
1. Has a `FunctionValue` with parameter specs and implementation
2. Parameters marked as `Eval: true` are evaluated before call, `false` passed raw
3. Refinements supported: `--flag` (boolean) or `--option []` (value)
4. Implementation receives: `args []Value`, `refValues map[string]Value`, `eval Evaluator`

**Adding a new native**:
1. Define contract in `specs/*/contracts/*.md`
2. Write test in `test/contract/*_test.go`
3. Implement in `internal/native/*.go`
4. Add to appropriate `internal/native/register_*.go` file (Math, Series, Data, I/O, Control, or Help)
5. Call `registerAndBind(name, fn)` within the registration function

**Registration files** (organized by category):
- `register_math.go` - Math operations (+, -, *, /, comparison, logic, trigonometry)
- `register_series.go` - Series operations (first, last, append, insert, length?, etc.)
- `register_data.go` - Data manipulation (set, get, type?, object, clone, etc.)
- `register_io.go` - I/O operations (print, input, file operations)
- `register_control.go` - Control flow (if, when, loop, while, fn)
- `register_help.go` - Help and reflection (?, words, type-of, spec-of, debug, trace)

**Note**: `native.Registry` still exists for backward compatibility with the help system, but the evaluator uses frame-based lookup exclusively

### Action System (Feature 004: Dynamic Function Invocation)

**Actions** are polymorphic functions that dispatch to type-specific implementations based on the first argument's type. This enables REBOL-style series operations where `first [1 2 3]` and `first "hello"` both work correctly.

**Architecture**:
- **TypeRegistry**: Global map from `ValueType` → `*Frame` storing type frames
- **Type Frames**: Regular frames containing type-specific implementations, stored in TypeRegistry (not on stack)
  - Parent = 0 (root frame index)
  - Index = -1 (not in frameStore)
  - Contains type-specific function implementations
- **Actions**: Values of type `TypeAction` stored in root frame like functions
  - Contain: Name and ParamSpec only
  - Dispatch at runtime: action name + first arg type → type frame → function

**Dispatch Flow**:
1. Action invoked: `first [1 2 3]`
2. First argument evaluated → type is `TypeBlock`
3. Look up `TypeBlock` in `frame.TypeRegistry` → get block type frame
4. Look up `"first"` in block type frame → get block-specific first function
5. Invoke function with arguments

**Adding type-specific implementations**:
1. Create implementations in `internal/native/series_block.go`, `series_string.go`, etc.
2. Register into type frames using `RegisterActionImpl(TypeBlock, "first", funcImpl)`
3. Create action value using `CreateAction("first", paramSpec)`
4. Bind action to root frame in `register_series.go`

**Error handling**:
- `action-no-impl`: Type has no type frame or type frame doesn't have this action
- `arg-count`: Wrong number of arguments
- Type-specific errors: Handled by individual implementations

**Extensibility**:
- New types can register via `frame.RegisterTypeFrame(typ, frame)`
- Dispatch logic treats all types uniformly (no special cases)
- Type frames use standard frame mechanism (Words/Values arrays)

### Evaluator Interface

Two parallel interfaces exist due to import cycle constraints:

- `native.Evaluator`: returns `*verror.Error` (used by native implementations)
- `value.Evaluator`: returns `error` (used by FunctionValue.Native field)

Adapters bridge between them automatically in registry code.

### Error Handling

Use structured errors from `verror` package:

```go
verror.NewScriptError("no-value", [3]string{"x", "", ""}, near, where)
verror.NewMathError("div-zero", [3]string{}, near, where)
verror.NewSyntaxError("unclosed-block", [3]string{}, near, where)
```

Errors include:
- `Near []Value`: expressions around error location
- `Where []string`: call stack trace (function names)

## Important Patterns

### Testing Pattern

Always use table-driven tests:

```go
func TestNativeAdd(t *testing.T) {
    tests := []struct {
        name     string
        args     []value.Value
        want     value.Value
        wantErr  bool
    }{
        {"add integers", []value.Value{value.IntVal(3), value.IntVal(4)}, value.IntVal(7), false},
        {"add negative", []value.Value{value.IntVal(-5), value.IntVal(3)}, value.IntVal(-2), false},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := native.Add(tt.args)
            if (err != nil) != tt.wantErr {
                t.Errorf("Add() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if !got.Equals(tt.want) {
                t.Errorf("Add() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

### Frame Access Pattern

Always use index-based references:

```go
// CORRECT
type Frame struct {
    Parent int  // index into stack
}

// INCORRECT - DO NOT USE
type Frame struct {
    Parent *Frame  // pointer invalidates on stack expansion
}
```

### Series Operations

Blocks and strings implement series interface:
- `First()`, `Last()`, `At(index)`, `Length()` - access
- `Append(value)`, `Insert(value)`, `Remove()` - modification
- Series are mutable (in-place operations)

## Feature Development

### Current Status
- **Phase 1 (Core)**: Largely complete - basic evaluation, control flow, series, functions
- **Phase 2 (Deferred Features)**: In progress (branch `002-implement-deferred-features`)
  - High-precision decimals (IEEE 754 decimal128)
  - Parse dialect for pattern matching
  - Objects and paths for structured data
  - Sandboxed ports for file/network I/O
  - Observability (tracing, debugging, reflection)

### Specification-Driven Development

Each feature has comprehensive specs in `specs/*/`:
- `spec.md` - Feature specification with user stories and requirements
- `plan.md` - Implementation plan and timeline
- `data-model.md` - Entity definitions and relationships
- `contracts/*.md` - Detailed function contracts
- `tasks.md` - Task breakdown and progress

Always consult specs before implementing.

## Common Gotchas

1. **Operator precedence**: There is NONE. Everything is left-to-right: `2 + 3 * 4` = 20, not 14
2. **Blocks vs Parens**: `[1 + 2]` stores block, `(1 + 2)` evaluates to 3
3. **Local scoping**: Function words are LOCAL by default, don't modify globals accidentally
4. **Native shadowing**: You CAN shadow native functions with local bindings (e.g., define local `print` or use `--debug` refinement)
5. **Index-based refs**: Never store frame/stack pointers, always use indices
6. **Test coverage**: Every code change MUST have test coverage (enforced in Copilot instructions)
7. **No real network calls in tests**: Use mocked/stubbed servers on 127.0.0.1 only

## Version and Dependencies

- **Go**: 1.21+ required (uses generics)
- **Current version**: See README.md for release version
- **Key dependency**: `github.com/ericlagergren/decimal` for high-precision decimals
- **REPL**: `github.com/chzyer/readline` for command history

## Documentation Resources

- Main docs: `specs/001-implement-the-core/` and `specs/002-implement-deferred-features/`
- Quickstart: `specs/001-implement-the-core/quickstart.md`
- Data model: `specs/001-implement-the-core/data-model.md`
- Native contracts: `specs/*/contracts/`
- Design decisions: `specs/*/research.md`
- Copilot rules: `.github/copilot-instructions.md` (important patterns)

## Branch Strategy

- `main` - Stable releases
- `001-implement-the-core` - Phase 1 development (completed)
- `002-implement-deferred-features` - Current active development (Phase 2)

Always check which branch you're on before making changes. Feature branches follow pattern `NNN-feature-name` where NNN is the spec number.

## Zasady generowania kodu

- Nie pisz komentarzy w kodzie; wszystkie wyjaśnienia i dokumentacja powinny być poza blokami kodu.
