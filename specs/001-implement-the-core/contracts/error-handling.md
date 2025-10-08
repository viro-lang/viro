# Contract: Error Handling

**Category**: Error System  
**Purpose**: Structured error representation and categorization

---

## Error Structure

**Type**: `Error` struct

**Fields**:
```go
type Error struct {
    Category ErrorCategory  // Error class (0-900)
    Code     int           // Specific error code
    ID       string        // Symbolic identifier
    Args     [3]string     // Message interpolation arguments
    Near     []Value       // Context: expressions around error
    Where    []string      // Context: call stack trace
    Message  string        // Formatted error message
}
```

---

## Error Categories

### Throw (0)

**Purpose**: Loop control flow errors

**Error IDs**:
- `break-outside-loop`: Break statement outside loop context
- `continue-outside-loop`: Continue statement outside loop context

**Examples**:
```viro
break               → Error: Break outside loop context
continue            → Error: Continue outside loop context

; Valid usage (future phase):
loop 5 [
    if x > 10 [break]
]
```

**Phase 1 Status**: Break/continue not implemented (deferred), but error category defined

---

### Note (100)

**Purpose**: Non-fatal warnings

**Usage**: Future use for deprecation warnings, performance hints

**Phase 1 Status**: Category defined, no specific warnings in Phase 1

---

### Syntax (200)

**Purpose**: Source code parsing errors

**Error IDs**:
- `unclosed-block`: Block opened with `[` but never closed
- `unclosed-string`: String opened with `"` but never closed
- `invalid-literal`: Malformed literal value (e.g., `12x34abc`)
- `unexpected-token`: Token in invalid position

**Examples**:
```viro
[1 2 3              → Syntax error (200): Unclosed block at end of input

"hello              → Syntax error (200): Unclosed string at end of input

12x34abc            → Syntax error (200): Invalid literal format

) 1 2 3             → Syntax error (200): Unexpected token ')'
```

**Error Context**:
- Near: tokens around error position
- Where: empty (parsing occurs before evaluation)

**Test Cases**:
1. `[1 2 3` errors with "unclosed-block"
2. `"hello` errors with "unclosed-string"
3. `12xabc` errors with "invalid-literal"

---

### Script (300)

**Purpose**: Runtime evaluation errors

**Error IDs**:
- `no-value`: Word has no bound value
- `type-mismatch`: Operation received wrong argument type
- `arg-count`: Function called with wrong number of arguments
- `invalid-operation`: Operation not valid for given arguments
- `not-defined`: Attempting to use undefined identifier

**Examples**:
```viro
undefined-word      → Script error (300): No value for word: undefined-word

3 + "string"        → Script error (300): Type mismatch for '+': cannot add integer and string

square: fn [n] [n * n]
square              → Script error (300): Expected 1 arguments, got 0
square 1 2 3        → Script error (300): Expected 1 arguments, got 3

first 42            → Script error (300): First expects series argument, got integer
```

**Error Context**:
- Near: expression that caused error and surrounding context
- Where: function call stack at error point

**Message Interpolation**:
- Args[0]: operation or word name
- Args[1]: received type or value
- Args[2]: expected type or value

**Test Cases**:
1. Undefined word → "no-value" with word name
2. Type mismatch → "type-mismatch" with operation and types
3. Wrong arg count → "arg-count" with expected and actual counts
4. Invalid operation → "invalid-operation" with details

---

### Math (400)

**Purpose**: Arithmetic operation errors

**Error IDs**:
- `div-zero`: Division by zero
- `overflow`: Integer overflow in arithmetic
- `underflow`: Integer underflow in arithmetic
- `invalid-math`: Invalid mathematical operation

**Examples**:
```viro
10 / 0              → Math error (400): Division by zero

9223372036854775807 + 1  → Math error (400): Integer overflow in add

-9223372036854775808 - 1  → Math error (400): Integer underflow in subtract
```

**Error Context**:
- Near: math expression that caused error
- Where: function call stack (if in function)

**Test Cases**:
1. `10 / 0` errors with "div-zero"
2. `max-int + 1` errors with "overflow"
3. `min-int - 1` errors with "underflow"

---

### Access (500)

**Purpose**: I/O, file access, and security errors

**Error IDs**:
- `io-error`: General I/O failure
- `read-error`: Cannot read from source
- `write-error`: Cannot write to destination
- `permission-denied`: Insufficient permissions
- `not-found`: File or resource not found

**Examples**:
```viro
; Phase 1: input read errors
input               → Access error (500): IO error reading input (if stdin fails)

; Future phases: file operations
read %/root/secret  → Access error (500): Permission denied
write %readonly.txt → Access error (500): Write error: file is read-only
```

**Phase 1 Status**: Only `input` native can trigger Access errors (stdin read failures)

**Test Cases**:
1. Simulate stdin EOF → `input` errors with "io-error"
2. Simulate stdin read failure → `input` errors with "read-error"

---

### Internal (900)

**Purpose**: Interpreter internal errors (bugs)

**Error IDs**:
- `stack-overflow`: Stack depth exceeded limit
- `out-of-memory`: Memory allocation failed
- `internal-error`: Unrecoverable internal error
- `not-implemented`: Feature not yet implemented

**Examples**:
```viro
; Recursive function without base case
infinite: fn [] [infinite]
infinite            → Internal error (900): Stack overflow (after ~100 levels)

; Theoretical memory exhaustion
loop 1000000000 [append data [1 2 3]]
                    → Internal error (900): Out of memory

; Future feature
parse "text" [rules]  → Internal error (900): Parse not implemented in Phase 1
```

**Error Context**:
- Near: expression being evaluated when error occurred
- Where: full call stack at error point

**Test Cases**:
1. Deep recursion → "stack-overflow" after depth limit
2. (Memory errors hard to test reliably)
3. Unimplemented features → "not-implemented" with feature name

---

## Error Construction

### Factory Functions

```go
func NewThrowError(id string, args [3]string, near []Value, where []string) *Error {
    return &Error{
        Category: ErrThrow,
        Code:     0 + hashID(id),
        ID:       id,
        Args:     args,
        Near:     near,
        Where:    where,
        Message:  formatMessage(id, args),
    }
}

func NewSyntaxError(id string, args [3]string, near []Value) *Error {
    return &Error{
        Category: ErrSyntax,
        Code:     200 + hashID(id),
        ID:       id,
        Args:     args,
        Near:     near,
        Where:    []string{},  // no call stack during parsing
        Message:  formatMessage(id, args),
    }
}

func NewScriptError(id string, args [3]string, near []Value, where []string) *Error {
    return &Error{
        Category: ErrScript,
        Code:     300 + hashID(id),
        ID:       id,
        Args:     args,
        Near:     near,
        Where:    where,
        Message:  formatMessage(id, args),
    }
}

func NewMathError(id string, args [3]string, near []Value, where []string) *Error {
    return &Error{
        Category: ErrMath,
        Code:     400 + hashID(id),
        ID:       id,
        Args:     args,
        Near:     near,
        Where:    where,
        Message:  formatMessage(id, args),
    }
}

func NewAccessError(id string, args [3]string, near []Value, where []string) *Error {
    return &Error{
        Category: ErrAccess,
        Code:     500 + hashID(id),
        ID:       id,
        Args:     args,
        Near:     near,
        Where:    where,
        Message:  formatMessage(id, args),
    }
}

func NewInternalError(id string, args [3]string, near []Value, where []string) *Error {
    return &Error{
        Category: ErrInternal,
        Code:     900 + hashID(id),
        ID:       id,
        Args:     args,
        Near:     near,
        Where:    where,
        Message:  formatMessage(id, args),
    }
}
```

### Message Formatting

```go
func formatMessage(id string, args [3]string) string {
    templates := map[string]string{
        "no-value":           "No value for word: %1",
        "type-mismatch":      "Type mismatch for '%1': expected %2, got %3",
        "arg-count":          "Expected %1 arguments, got %2",
        "div-zero":           "Division by zero",
        "overflow":           "Integer overflow in %1",
        "unclosed-block":     "Unclosed block at %1",
        "unclosed-string":    "Unclosed string at %1",
        "io-error":           "IO error: %1",
        "stack-overflow":     "Stack overflow (depth limit exceeded)",
        "not-implemented":    "Feature not implemented: %1",
    }
    
    template := templates[id]
    if template == "" {
        return fmt.Sprintf("Error: %s", id)
    }
    
    // Replace %1, %2, %3 with args
    msg := template
    msg = strings.Replace(msg, "%1", args[0], -1)
    msg = strings.Replace(msg, "%2", args[1], -1)
    msg = strings.Replace(msg, "%3", args[2], -1)
    
    return msg
}
```

---

## Error Display

### REPL Error Format

```
>> undefined-word
Script error (300): No value for word: undefined-word
Near: undefined-word
Where: (top level)

>> square: fn [n] [n * n]
=> [function]
>> square
Script error (300): Expected 1 arguments, got 0
Near: square
Where: (top level)

>> bad: fn [] [x + y]
=> [function]
>> bad
Script error (300): No value for word: x
Near: x + y
Where: bad → (top level)
```

### Error.Error() Method

Implements Go error interface:

```go
func (e *Error) Error() string {
    var parts []string
    
    // Category and message
    parts = append(parts, fmt.Sprintf("%s error (%d): %s",
        e.Category.String(), e.Code, e.Message))
    
    // Near context (if available)
    if len(e.Near) > 0 {
        nearStr := formatValues(e.Near)
        parts = append(parts, fmt.Sprintf("Near: %s", nearStr))
    }
    
    // Where context (if available)
    if len(e.Where) > 0 {
        whereStr := strings.Join(e.Where, " → ")
        parts = append(parts, fmt.Sprintf("Where: %s", whereStr))
    }
    
    return strings.Join(parts, "\n")
}
```

---

## Context Capture

### Near Context

Capture expressions around error location:

```go
func (ctx *EvalContext) CaptureNear() []Value {
    if ctx.Block == nil || ctx.Index < 0 {
        return []Value{}
    }
    
    start := max(0, ctx.Index-3)
    end := min(len(ctx.Block.Elements), ctx.Index+4)
    
    return ctx.Block.Elements[start:end]
}
```

### Where Context

Capture call stack:

```go
func (s *Stack) CaptureWhere() []string {
    var calls []string
    
    frameIdx := s.CurrentFrame
    for frameIdx != -1 {
        frame := s.GetFrame(frameIdx)
        if frame.Function != nil {
            calls = append(calls, frame.Function.Name)
        }
        frameIdx = frame.PriorFrame
    }
    
    if len(calls) == 0 {
        return []string{"(top level)"}
    }
    
    return calls
}
```

---

## Error Propagation

### Native Functions

Natives return errors via Go error return:

```go
func NativeDivide(args []Value) (Value, error) {
    a, _ := AsInteger(args[0])
    b, _ := AsInteger(args[1])
    
    if b == 0 {
        return NoneVal(), NewMathError("div-zero", 
            [3]string{"", "", ""}, nil, nil)
    }
    
    return IntVal(a / b), nil
}
```

### Evaluator

Evaluator propagates errors up call stack:

```go
func (e *Evaluator) EvalBlock(block *BlockValue) (Value, error) {
    var result Value = NoneVal()
    
    for i := 0; i < len(block.Elements); i++ {
        val, err := e.EvalValue(block.Elements[i])
        if err != nil {
            // Attach context if not already present
            if evalErr, ok := err.(*Error); ok {
                if len(evalErr.Near) == 0 {
                    evalErr.Near = e.ctx.CaptureNear()
                }
                if len(evalErr.Where) == 0 {
                    evalErr.Where = e.stack.CaptureWhere()
                }
            }
            return NoneVal(), err  // propagate
        }
        result = val
    }
    
    return result, nil
}
```

### REPL

REPL catches errors and displays without crashing:

```go
func (repl *REPL) EvalAndPrint(input string) {
    result, err := repl.eval.Eval(input)
    
    if err != nil {
        fmt.Fprintf(os.Stderr, "%s\n", err.Error())
        return  // continue REPL loop
    }
    
    fmt.Printf("=> %s\n", result.String())
}
```

---

## Testing Strategy

### Error Construction Tests

```go
func TestErrorConstruction(t *testing.T) {
    err := NewScriptError("no-value", [3]string{"x", "", ""}, nil, nil)
    
    assert.Equal(t, ErrScript, err.Category)
    assert.Equal(t, "no-value", err.ID)
    assert.Contains(t, err.Message, "No value for word: x")
}
```

### Error Propagation Tests

```go
func TestErrorPropagation(t *testing.T) {
    eval := NewEvaluator()
    
    _, err := eval.Eval("undefined-word")
    
    assert.Error(t, err)
    evalErr := err.(*Error)
    assert.Equal(t, ErrScript, evalErr.Category)
    assert.Equal(t, "no-value", evalErr.ID)
}
```

### Near/Where Context Tests

```go
func TestErrorContext(t *testing.T) {
    eval := NewEvaluator()
    
    _, err := eval.Eval("square: fn [n] [1 / 0]  square 5")
    
    evalErr := err.(*Error)
    assert.Equal(t, ErrMath, evalErr.Category)
    assert.NotEmpty(t, evalErr.Near)  // should capture "1 / 0"
    assert.NotEmpty(t, evalErr.Where)  // should show "square → (top level)"
}
```

---

## Implementation Checklist

- [ ] Error struct definition
- [ ] ErrorCategory constants (0-900)
- [ ] Factory functions for each category
- [ ] Message interpolation (formatMessage)
- [ ] Error.Error() method (Go error interface)
- [ ] Near context capture
- [ ] Where context capture (call stack)
- [ ] Error propagation in evaluator
- [ ] REPL error display
- [ ] All error test cases pass

**Dependencies**:
- Value system (for Near context)
- Stack system (for Where context)
- Frame system (for function names in call stack)
- Go error interface

---

## Success Criteria

Per spec **SC-003**: Error messages include sufficient context that users can diagnose and fix issues in under 2 minutes for common errors.

**Required Context**:
- Error category and specific message
- Near: expressions around error location
- Where: function call stack trace
- Clear, English-only messages (per spec clarification)

**Common Error Examples**:
1. Undefined word → shows word name and location
2. Type mismatch → shows operation, expected type, received type
3. Arg count → shows expected and actual counts
4. Division by zero → clear message with expression
5. Function error → shows function name in call stack
