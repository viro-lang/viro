# Quickstart: Viro Interpreter

**Feature**: Viro Core Language and REPL  
**Date**: 2025-01-07  
**Purpose**: Build, run, and test the Viro interpreter

---

## Prerequisites

- **Go**: Version 1.21 or later (for generics and improved error handling)
- **Git**: For cloning repository
- **Terminal**: macOS Terminal, iTerm2, or any Unix shell

**Verify Go Installation**:
```bash
go version
# Expected: go version go1.21.0 or later
```

---

## Quick Start

### 1. Clone and Build

```bash
# Clone repository
git clone <repository-url>
cd viro

# Build the interpreter
go build -o viro ./cmd/viro

# Verify build
./viro --version
# Expected: Viro v1.0.0 (or current version)
```

### 2. Run REPL

```bash
./viro
```

**Expected Output**:
```
Viro v1.0.0 - homoiconic interpreter
Type 'quit' or 'exit' to exit, Ctrl+C to interrupt evaluation
>> 
```

### 3. Try Basic Expressions

```viro
>> 42
=> 42

>> 3 + 4
=> 7

>> print "Hello, Viro!"
Hello, Viro!

>> x: 10
=> 10

>> x * 2
=> 20

>> quit
```

---

## REPL Features

### Expression Evaluation

```viro
>> 5 + 3
=> 8

>> 10 / 2
=> 5

>> 3 < 5
=> true
```

**Operator Precedence** (multiplication/division before addition/subtraction):

```viro
>> 3 + 4 * 2
=> 11                ; multiplication first: 3 + (4*2)

>> 2 + 3 * 4
=> 14                ; not 20: 2 + (3*4)

>> 10 - 4 / 2
=> 8                 ; division first: 10 - (4/2)

>> (3 + 4) * 2
=> 14                ; parentheses override: (3+4) * 2

>> 1 + 2 < 5
=> true              ; arithmetic first: (1+2) < 5
```

### Variables

```viro
>> name: "Alice"
=> "Alice"

>> name
=> "Alice"

>> age: 30
=> 30

>> age + 5
=> 35
```

### Blocks vs Parens

Blocks `[...]` evaluate to themselves (deferred evaluation):

```viro
>> [1 2 3]
=> [1 2 3]

>> x: [1 + 2]
=> [1 + 2]

>> x
=> [1 + 2]          ; block stored as-is, not evaluated
```

Parens `(...)` evaluate their contents immediately:

```viro
>> (1 + 2)
=> 3

>> y: (1 + 2)
=> 3

>> y
=> 3                ; result of evaluation stored

>> print ["Result:" (10 * 2)]
Result: 20
```

**Use cases:**
- Blocks: data structures, code storage, deferred execution
- Parens: immediate calculations, nested expressions, dynamic values

### Control Flow

```viro
>> when true [print "yes"]
yes

>> when false [print "yes"]

>> if 1 < 2 ["less"] ["more"]
=> "less"

>> if false ["yes"] ["no"]
=> "no"

>> loop 3 [print "hi"]
hi
hi
hi
```

### Functions

```viro
>> square: fn [n] [n * n]
=> [function]

>> square 5
=> 25

>> add: fn [a b] [a + b]
=> [function]

>> add 10 20
=> 30
```

### Series Operations

```viro
>> data: [1 2 3]
=> [1 2 3]

>> first data
=> 1

>> last data
=> 3

>> append data 4
=> [1 2 3 4]

>> length? data
=> 4
```

### Multi-line Input

```viro
>> calculate: fn [x] [
...     temp: x * 2      ; temp is local to function
...     temp + 1
... ]
=> [function]

>> calculate 5
=> 11

>> temp
Script error (300): No value for word: temp
; temp is local to calculate, not accessible outside
```

### Local-by-Default Scoping

Functions use local-by-default scoping for safety:

```viro
>> counter: 0
=> 0

>> increment: fn [] [
...     counter: counter + 1    ; creates LOCAL counter (does not modify global)
...     counter
... ]
=> [function]

>> increment
=> 1

>> counter
=> 0
; Global counter unchanged - function created local variable

>> increment
=> 1
; Each call creates fresh local counter (not persistent)
```

### Bash-Style Refinements

Functions support bash-style refinements for optional parameters:

```viro
>> greet: fn [name --formal --title []] [
...     if formal [
...         if title [
...             print [title name ", good day!"]
...         ] [
...             print ["Good day," name]
...         ]
...     ] [
...         when title [
...             print ["Hi" title name "!"]
...         ]
...         when (not title) [
...             print ["Hi" name "!"]
...         ]
...     ]
... ]
=> [function]

>> greet "Alice"
Hi Alice !

>> greet "Bob" --formal
Good day, Bob

>> greet "Carol" --title "Dr."
Hi Dr. Carol !

>> greet "Dave" --formal --title "Prof."
Prof. Dave , good day!

>> greet "Eve" --title "Ms." --formal
Ms. Eve , good day!
; Refinement order doesn't matter
```

**Refinement Syntax**:
- `--flag` → Boolean flag (true if present, false otherwise)
- `--option []` → Value refinement (accepts any value, none if not provided)
- Refinements can be mixed with positional arguments in any order

**Note**: REPL shows `...` prompt for continuation when expression is incomplete.

### Command History

- **Up arrow**: Previous command
- **Down arrow**: Next command
- **History**: Persists across sessions (stored in `~/.viro_history`)

### Interrupt

- **Ctrl+C**: Interrupt current evaluation, return to prompt
- **Ctrl+D**: Exit REPL (same as `quit`)

---

## Testing

### Run All Tests

```bash
# Run all tests with verbose output
go test -v ./...

# Expected output:
# === RUN   TestNativeAdd
# --- PASS: TestNativeAdd (0.00s)
# ...
# PASS
# ok      github.com/viro/internal/native    0.123s
```

### Run Specific Test Category

```bash
# Contract tests only
go test -v ./test/contract/...

# Integration tests only
go test -v ./test/integration/...

# Specific package
go test -v ./internal/eval
```

### Run Tests with Coverage

```bash
# Generate coverage report
go test -coverprofile=coverage.out ./...

# View coverage
go tool cover -html=coverage.out
```

### Run Benchmarks

```bash
# Run all benchmarks
go test -bench=. -benchmem ./...

# Expected output:
# BenchmarkStackPush-8      50000000    25.3 ns/op    0 B/op    0 allocs/op
# BenchmarkNativeAdd-8      10000000   123.0 ns/op   16 B/op    1 allocs/op
# BenchmarkEvalSimple-8       100000 10000.0 ns/op  512 B/op   20 allocs/op
```

---

## Project Structure

```
viro/
├── cmd/
│   └── viro/              # CLI entry point
│       ├── main.go        # Program entry
│       └── repl.go        # REPL implementation
├── internal/
│   ├── value/             # Value types
│   ├── eval/              # Core evaluator
│   ├── stack/             # Stack management
│   ├── frame/             # Frame & context
│   ├── native/            # Native functions
│   ├── error/             # Error system
│   └── parse/             # Parser (future)
├── test/
│   ├── contract/          # Native function tests
│   ├── integration/       # End-to-end tests
│   └── fixtures/          # Test scripts
├── specs/
│   └── 001-implement-the-core/
│       ├── spec.md        # Feature specification
│       ├── plan.md        # Implementation plan
│       ├── data-model.md  # Entity definitions
│       ├── research.md    # Technical decisions
│       ├── contracts/     # Native contracts
│       └── quickstart.md  # This file
├── docs/
│   └── interpreter.md     # Design specification
├── go.mod                 # Go module definition
└── README.md              # Project overview
```

---

## Development Workflow

### TDD Cycle (Per Constitution)

1. **Write failing test**:
   ```go
   func TestNativeAdd(t *testing.T) {
       result, err := NativeAdd([]Value{IntVal(3), IntVal(4)})
       assert.NoError(t, err)
       assert.Equal(t, IntVal(7), result)
   }
   ```

2. **Run test** (should fail):
   ```bash
   go test -v ./internal/native -run TestNativeAdd
   # Expected: FAIL - function not implemented
   ```

3. **Implement minimal code**:
   ```go
   func NativeAdd(args []Value) (Value, error) {
       a := args[0].Payload.(int64)
       b := args[1].Payload.(int64)
       return IntVal(a + b), nil
   }
   ```

4. **Run test** (should pass):
   ```bash
   go test -v ./internal/native -run TestNativeAdd
   # Expected: PASS
   ```

5. **Refactor** with confidence (tests protect changes)

### Adding a New Native Function

1. **Define contract** in `specs/001-implement-the-core/contracts/`:
   - Signature, parameters, return, behavior, examples, tests

2. **Write contract test** in `test/contract/`:
   ```go
   func TestNativeMyFunc(t *testing.T) {
       // table-driven tests per contract
   }
   ```

3. **Implement native** in `internal/native/`:
   ```go
   func NativeMyFunc(args []Value, eval *Evaluator) (Value, error) {
       // implementation per contract
   }
   ```

4. **Register native** in `internal/native/registry.go`:
   ```go
   registry["my-func"] = NativeMyFunc
   ```

5. **Run tests**:
   ```bash
   go test -v ./test/contract -run TestNativeMyFunc
   ```

6. **Add integration tests** if needed

---

## Debugging

### Enable Verbose Output

```bash
# Set environment variable for debug output
export VIRO_DEBUG=1
./viro

# Shows evaluation steps, stack operations, frame creation
```

### Inspect Values

```viro
>> x: [1 2 3]
=> [1 2 3]

>> type? x
=> block!

>> length? x
=> 3

>> first x
=> 1
```

### Common Errors

**Undefined Word**:
```viro
>> undefined-var
Script error (300): No value for word: undefined-var
Near: undefined-var
Where: (top level)
```

**Type Mismatch**:
```viro
>> 3 + "string"
Script error (300): Type mismatch for '+': cannot add integer and string
Near: 3 + "string"
Where: (top level)
```

**Division by Zero**:
```viro
>> 10 / 0
Math error (400): Division by zero
Near: 10 / 0
Where: (top level)
```

**Wrong Argument Count**:
```viro
>> square: fn [n] [n * n]
=> [function]

>> square
Script error (300): Expected 1 arguments, got 0
Near: square
Where: (top level)
```

---

## Examples

### Factorial

```viro
>> factorial: fn [n] [
...     if n <= 1 [
...         1
...     ] [
...         n * factorial n - 1
...     ]
... ]
=> [function]

>> factorial 5
=> 120

>> factorial 10
=> 3628800
```

### Sum of Series

```viro
>> sum: fn [n] [
...     total: 0
...     loop n [total: total + n  n: n - 1]
...     total
... ]
=> [function]

>> sum 10
=> 55
```

### Interactive Input

```viro
>> greet: fn [] [
...     print "What is your name?"
...     name: input
...     print ["Hello" name]
... ]
=> [function]

>> greet
What is your name?
Alice
Hello Alice
```

### Data Processing

```viro
>> data: [10 20 30]
=> [10 20 30]

>> double-all: fn [series] [
...     result: []
...     loop length? series [
...         append result (first series) * 2
...         series: next series    ; future: series position
...     ]
...     result
... ]
=> [function]

>> double-all data
=> [20 40 60]
```

---

## Performance

### Benchmarking

```bash
# Benchmark specific function
go test -bench=BenchmarkNativeAdd -benchmem ./internal/native

# Compare before/after optimization
go test -bench=. -benchmem ./... > before.txt
# ... make changes ...
go test -bench=. -benchmem ./... > after.txt
benchstat before.txt after.txt
```

### Expected Performance

Per research.md and spec.md success criteria:

| Operation | Target | Notes |
|-----------|--------|-------|
| Stack push/pop | <100 ns/op | Index-based access |
| Type dispatch | <50 ns/op | Switch statement |
| Native function call | <1 µs/op | Includes type checking |
| Simple expression | <10 ms | e.g., `3 + 4` |
| Complex expression | <100 ms | e.g., nested function calls |

### Profiling

```bash
# CPU profile
go test -cpuprofile=cpu.prof -bench=. ./internal/eval
go tool pprof cpu.prof

# Memory profile
go test -memprofile=mem.prof -bench=. ./internal/eval
go tool pprof mem.prof
```

---

## Troubleshooting

### Build Fails

**Problem**: `go build` fails with module errors

**Solution**:
```bash
go mod tidy
go mod download
go build -o viro ./cmd/viro
```

### Tests Fail

**Problem**: Tests fail after changes

**Solution**:
```bash
# Run tests with verbose output to see exact failure
go test -v ./...

# Run specific failing test
go test -v ./internal/eval -run TestEvaluator

# Check test coverage
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out | grep -v "100.0%"
```

### REPL Hangs

**Problem**: REPL appears frozen

**Solution**:
- Press Ctrl+C to interrupt current evaluation
- Check for infinite loops in code
- Review recent changes for blocking operations

### History Not Working

**Problem**: Up arrow doesn't recall commands

**Solution**:
```bash
# Check history file permissions
ls -la ~/.viro_history

# If missing, create empty file
touch ~/.viro_history
chmod 600 ~/.viro_history
```

---

## Next Steps

1. **Explore Examples**: Try examples in `test/fixtures/examples/`
2. **Read Specification**: Review `specs/001-implement-the-core/spec.md`
3. **Study Contracts**: Read native function contracts in `specs/001-implement-the-core/contracts/`
4. **Contribute**: Follow TDD workflow to add features
5. **Optimize**: Profile and benchmark critical paths

---

## Resources

- **Specification**: `specs/001-implement-the-core/spec.md`
- **Implementation Plan**: `specs/001-implement-the-core/plan.md`
- **Data Model**: `specs/001-implement-the-core/data-model.md`
- **Native Contracts**: `specs/001-implement-the-core/contracts/`
- **Constitution**: `.specify/memory/constitution.md`
- **Design Doc**: `docs/interpreter.md`
- **REBOL Documentation**: http://www.rebol.com/docs.html

---

## Support

For issues or questions:
1. Check error message and Near/Where context
2. Review relevant contract in `specs/001-implement-the-core/contracts/`
3. Run tests to verify expected behavior
4. Check constitution for architectural guidance

**Remember**: TDD is non-negotiable per constitution. Always write tests before implementation.
