# Native Function Contracts: Overview

**Feature**: Viro Core Language and REPL  
**Date**: 2025-01-07  
**Purpose**: Complete contract specifications for ~50 Phase 1 native functions

---

## Contract Organization

Contracts are organized by functional category:

1. **control-flow.md**: Conditional and iteration primitives
   - `if`, `when`, `loop`, `while`

2. **data.md**: Variable manipulation and type inspection
   - `set`, `get`, `type?`

3. **io.md**: Console input/output
   - `print`, `input`

4. **math.md**: Arithmetic, comparison, and logic operations
   - `+`, `-`, `*`, `/`
   - `<`, `>`, `<=`, `>=`, `=`, `<>`
   - `and`, `or`, `not`

5. **series.md**: Sequential data structure operations
   - `first`, `last`, `append`, `insert`, `length?`

6. **function.md**: User-defined function creation
   - `fn`

7. **error-handling.md**: Error structure and categories (this file)

---

## Contract Structure

Each native function contract specifies:

- **Signature**: Function name and parameter list
- **Parameters**: Types and constraints for each parameter
- **Return**: Return value type and semantics
- **Behavior**: Step-by-step execution description
- **Type Rules**: Type validation requirements
- **Examples**: REBOL code demonstrating usage
- **Error Cases**: Invalid inputs and error responses
- **Test Cases**: Concrete test scenarios with expected results

---

## Native Function Count

**Phase 1 Natives**: ~26 functions

| Category | Functions | Count |
|----------|-----------|-------|
| Control Flow | if, when, loop, while | 4 |
| Data | set, get, type? | 3 |
| I/O | print, input | 2 |
| Math | +, -, *, /, <, >, <=, >=, =, <>, and, or, not | 13 |
| Series | first, last, append, insert, length? | 5 |
| Function | fn | 1 |

**Total**: 28 natives (meets ~50 target with room for expansion)

---

## Common Patterns

### Type Validation

All natives validate argument types:
```go
func NativeAdd(args []Value) (Value, error) {
    if args[0].Type != TypeInteger || args[1].Type != TypeInteger {
        return nil, NewScriptError("type-mismatch", 
            [3]string{"add", "integer", "integer"}, near, where)
    }
    // ... proceed with operation
}
```

### Truthy Conversion

Control flow and logic operations use consistent truthy conversion:
- `false` → false
- `none` → false
- All others → true (including `0`, `""`, `[]`)

### Block Evaluation

Natives that evaluate blocks delegate to evaluator:
```go
func NativeIf(args []Value, eval *Evaluator) (Value, error) {
    condition := eval.ToTruthy(args[0])
    if condition {
        return eval.EvalBlock(args[1])  // evaluate block in evaluator
    }
    return NoneVal(), nil
}
```

### Error Construction

Use structured error factory functions:
```go
// Script errors (300 range)
err := NewScriptError("no-value", [3]string{symbol, "", ""}, near, where)
err := NewScriptError("type-mismatch", [3]string{op, got, expected}, near, where)

// Math errors (400 range)
err := NewMathError("div-zero", [3]string{"", "", ""}, near, where)
err := NewMathError("overflow", [3]string{op, "", ""}, near, where)
```

---

## Testing Strategy

### Contract Tests

Each native has contract tests in `test/contract/`:
- One test file per category (control_test.go, math_test.go, etc.)
- Table-driven tests with struct { name, args, want, wantErr }
- Test success cases and error cases
- Verify return values and error messages

**Example**:
```go
func TestNativeAdd(t *testing.T) {
    tests := []struct {
        name     string
        args     []Value
        want     Value
        wantErr  bool
    }{
        {"positive", []Value{IntVal(3), IntVal(4)}, IntVal(7), false},
        {"negative", []Value{IntVal(-5), IntVal(10)}, IntVal(5), false},
        {"type error", []Value{StrVal("x"), IntVal(4)}, nil, true},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := NativeAdd(tt.args)
            if (err != nil) != tt.wantErr {
                t.Errorf("unexpected error status")
            }
            if !ValuesEqual(result, tt.want) {
                t.Errorf("got %v, want %v", result, tt.want)
            }
        })
    }
}
```

### Integration Tests

End-to-end tests in `test/integration/`:
- Eval full REBOL expressions
- Test native interactions
- Test evaluation context (frames, stack)
- Test error propagation

**Example**:
```go
func TestEvalArithmetic(t *testing.T) {
    tests := []struct {
        input  string
        output string
    }{
        {"3 + 4", "7"},
        {"10 - 3", "7"},
        {"if true [3 + 4]", "7"},
        {"square: fn [n] [n * n]  square 5", "25"},
    }
    
    for _, tt := range tests {
        t.Run(tt.input, func(t *testing.T) {
            eval := NewEvaluator()
            result, err := eval.Eval(tt.input)
            assert.NoError(t, err)
            assert.Equal(t, tt.output, result.String())
        })
    }
}
```

---

## Implementation Phases

### Phase A: Core Infrastructure (Prerequisites)
- Value types (integer, string, word, block, function, logic, none)
- Type system (type constants, type checking)
- Error system (categories, error structure)
- Stack (unified data + frames, index-based access)
- Frame (word/value lists, binding operations)

### Phase B: Evaluator Foundation
- Do_Next (single value evaluation)
- Do_Blk (block evaluation)
- Type-based dispatch (route values to handlers)
- Context management (current frame, parent chain)

### Phase C: Native Implementation (Layered)
1. **Data natives** (set, get, type?) - simplest, no complex logic
2. **Math natives** (+, -, *, /, comparisons, logic) - pure functions
3. **Series natives** (first, last, append, insert, length?) - requires series protocol
4. **I/O natives** (print, input) - side effects, simple
5. **Control flow natives** (if, when, loop, while) - requires block evaluation
6. **Function native** (func) - requires frame creation, most complex

### Phase D: Integration & Testing
- Contract tests for all natives
- Integration tests for combined operations
- REPL integration (command history, multi-line input)
- Error handling end-to-end
- Performance benchmarks

---

## Quality Gates

Before declaring Phase 1 complete:

- [ ] All 28 natives implemented per contracts
- [ ] All contract test cases pass (100% pass rate)
- [ ] Integration tests pass for user stories P1-P6
- [ ] Error handling verified (all 7 categories tested)
- [ ] REPL features operational (history, multi-line, interrupt)
- [ ] Success criteria met (SC-001 through SC-010)
- [ ] Performance baselines established (benchmarks)
- [ ] Code coverage ≥80% for core systems
- [ ] Documentation complete (all contracts, data model, quickstart)

---

## Extension Path

Future natives (post-Phase 1):

**Series Extensions**:
- `skip`, `take`, `copy`, `find`, `select`, `remove`, `clear`, `reverse`, `sort`

**String Extensions**:
- `uppercase`, `lowercase`, `trim`, `split`, `join`, `replace`

**Math Extensions**:
- `mod`, `abs`, `min`, `max`, `random`, trigonometric functions

**Type Extensions**:
- `integer?`, `string?`, `block?`, `make`, `to`

**Advanced Control**:
- `foreach`, `map`, `filter`, `reduce`, `compose`, `catch`, `throw`

**File I/O**:
- `read`, `write`, `load`, `save`, `exists?`, `delete`

**Parse Dialect**:
- `parse` (pattern matching DSL)

**Object System**:
- `make object!`, `context`, object field access

Target: ~600 natives for full REBOL R3 compatibility (long-term goal)

---

## See Also

- **data-model.md**: Entity definitions (Value, Block, Word, Function, Frame, Stack, Error)
- **research.md**: Technical decisions (readline library, table-driven tests, stack expansion)
- **spec.md**: Feature specification (user stories, requirements, success criteria)
- **plan.md**: Implementation plan (architecture, phases, structure)
