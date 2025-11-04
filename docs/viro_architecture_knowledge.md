# Viro Language & Interpreter Architecture Knowledge Base

**Generated**: 2025-11-04  
**Purpose**: Comprehensive understanding of Viro programming language and interpreter architecture for agent reference

---

## Language Overview

### Core Design Principles

**Viro** is a homoiconic programming language interpreter implemented in Go with several distinctive features:

1. **Homoiconic Code-as-Data**: Code and data share the same representation (like Lisp)
2. **Left-to-Right Evaluation**: No operator precedence - expressions evaluate strictly left to right
3. **Local-by-Default Scoping**: Variables are local unless explicitly captured as closures
4. **Type-Based Dispatch**: Runtime type system drives evaluation semantics
5. **Bash-Style Refinements**: Command-line style flags (`--flag`, `--option value`)

### Key Language Constructs

#### Blocks vs Parens
- **Blocks `[...]`**: Deferred evaluation (returns block itself)
- **Parens `(...)`**: Immediate evaluation (evaluates contents, returns last result)

#### Word Types
- **Word** (e.g., `x`): Evaluates to bound value
- **Set-Word** (e.g., `x:`): Assignment operator
- **Get-Word** (e.g., `:x`): Fetch without evaluation  
- **Lit-Word** (e.g., `'x`): Returns word itself (quoted)

#### Examples
```viro
x: 10                    ; Set x to 10
x                        ; Evaluates to 10
[x 1 + 2]                ; Block: returns [x 1 + 2] (deferred)
(x 1 + 2)                ; Paren: evaluates to (10 1 + 2) → 13 (immediate)
3 + 4 * 2                ; Left-to-right: ((3 + 4) * 2) = 14 (no precedence)
3 + (4 * 2)              ; Parens override: (3 + 8) = 11
```

---

## Interpreter Architecture

### Value System

All data in Viro is represented as **tagged union values** with type discrimination:

```go
type ValueType uint8

const (
    TypeNone     // Represents absence of value (nil/null)
    TypeLogic    // Boolean true/false
    TypeInteger  // 64-bit signed integer
    TypeString   // UTF-8 character sequence
    TypeWord     // Symbol identifier
    TypeSetWord  // Assignment symbol (x: value)
    TypeGetWord  // Fetch symbol (evaluates without evaluation)
    TypeLitWord  // Quoted symbol (returns word itself)
    TypeBlock    // Series of values (deferred evaluation)
    TypeParen    // Series of values (immediate evaluation)
    TypeFunction // Executable function (native or user-defined)
    TypeDecimal  // IEEE 754 decimal128 high-precision decimal
    TypeObject   // Object instance with frame-based fields
    TypePort     // I/O port abstraction
    TypePath     // Path expression
    TypeGetPath  // Get-path expression
    TypeSetPath  // Set-path expression
    TypeDatatype // Datatype literal (e.g., object!, integer!)
    TypeBinary   // Raw byte sequence
)
```

### Core Components

#### 1. Evaluator (`internal/eval/evaluator.go`)
- **Type-based dispatch**: Routes evaluation based on value type
- **Expression evaluation**: Handles infix operators with left-to-right precedence
- **Function invocation**: Manages native and user-defined function calls
- **Frame management**: Creates/destroys execution frames

#### 2. Stack System (`internal/stack/stack.go`)
- **Unified storage**: Single stack for data values and frame metadata
- **Index-based access**: Uses integer indices instead of pointers (prevents stack expansion bugs)
- **Frame layout**:
  ```
  [frameBase]     Return value slot
  [frameBase+1]   Prior frame pointer  
  [frameBase+2]   Function metadata
  [frameBase+3]   Argument 1
  [frameBase+4]   Argument 2
  ...             Additional arguments
  [frameBase+3+N] Local variables (grows as needed)
  ```

#### 3. Frame System (`internal/frame/frame.go`)
- **Variable storage**: Maps symbols to values using parallel arrays
- **Scope management**: Parent-child relationships for lexical scoping
- **Frame types**:
  - `FrameFunctionArgs`: Function call frames (destroyed on return)
  - `FrameClosure`: Captured lexical environment (persists)
  - `FrameObject`: Object field storage (future)

#### 4. Parser (`grammar/viro.peg`)
- **PEG grammar**: Parsing Expression Grammar defines syntax
- **Left-to-right parsing**: Supports immediate evaluation semantics
- **Path expressions**: Dot notation for object/block navigation

### Type-Based Dispatch

The evaluator uses type tags to determine evaluation behavior:

```go
switch element.GetType() {
case value.TypeInteger, value.TypeString, value.TypeLogic,
     value.TypeNone, value.TypeDecimal, value.TypeObject,
     value.TypePort, value.TypeDatatype, value.TypeBlock,
     value.TypeFunction, value.TypeBinary:
    // Return value as-is (literals)
    return position + 1, element, nil

case value.TypeParen:
    // Evaluate block contents immediately
    result, err := e.DoBlock(block.Elements)
    return position + 1, result, err

case value.TypeWord:
    // Look up symbol, evaluate result
    return e.evaluateWord(block, element, position, traceStart, shouldTraceExpr)

case value.TypeSetWord:
    // Evaluate next expression, bind to symbol
    return e.evaluateSetWord(block, element, position, traceStart, shouldTraceExpr)

// ... additional cases
}
```

---

## Native Functions

### Implementation Pattern

Native functions follow a consistent contract:

```go
func FunctionName(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
    // Validate argument count
    if len(args) != expected {
        return value.NewNoneVal(), arityError("function", expected, len(args))
    }
    
    // Type validation
    if args[0].GetType() != value.TypeExpected {
        return value.NewNoneVal(), typeError("function", "expected-type", args[0])
    }
    
    // Implementation
    result, err := implementFunction(args, refValues, eval)
    return result, err
}
```

### Key Native Functions

#### Control Flow
- **when**: Conditional execution with truthy/falsy evaluation
- **if**: Binary conditional (both branches required)
- **loop**: Fixed-count iteration
- **while**: Conditional iteration with re-evaluation support

#### Data Manipulation
- **set/get**: Variable binding and lookup
- **type?**: Type introspection
- **form/mold**: String representation (human-readable vs serialization)
- **join/rejoin**: String concatenation
- **reduce/compose**: Block evaluation and transformation

#### Object System
- **object/context**: Create object instances
- **make**: Object creation with prototypes
- **select**: Field lookup with defaults
- **put**: Field assignment

### Function Arguments & Refinements

Viro supports two types of function parameters:

1. **Positional parameters**: Regular arguments
2. **Refinements**: Bash-style flags and options

```viro
; Function definition with refinements
greet: fn [name --formal --title []] [
    ; name: positional parameter
    ; --formal: boolean refinement (true/false)
    ; --title: value refinement (accepts string)
    
    if formal [
        if title [print [title name ", good day!"]]
        [print ["Good day," name]]
    ] [
        when title [print ["Hi" title name "!"]]
        [print ["Hi" name "!"]]
    ]
]

; Usage examples
greet "Alice"                        ; Hi Alice !
greet "Bob" --formal                  ; Good day, Bob  
greet "Carol" --title "Dr."           ; Hi Dr. Carol !
greet "Dave" --formal --title "Prof." ; Prof. Dave, good day!
```

---

## Error Handling

### Structured Error System (`internal/verror/`)

Errors have structured categories, codes, and context:

```go
type ErrorCategory uint16

const (
    ErrThrow    ErrorCategory = 0    // Loop control errors
    ErrNote     ErrorCategory = 100  // Warnings
    ErrSyntax   ErrorCategory = 200  // Parsing errors
    ErrScript   ErrorCategory = 300  // Runtime errors
    ErrMath     ErrorCategory = 400  // Arithmetic errors  
    ErrAccess   ErrorCategory = 500  // I/O/security errors
    ErrInternal ErrorCategory = 900  // System errors
)

type Error struct {
    Category ErrorCategory
    Code     int           // Specific error code within category
    ID       string        // Error identifier (e.g., "no-value", "div-zero")
    Args     [3]string     // Arguments for message interpolation
    Near     []Value       // Context around error location
    Where    []string      // Call stack trace
    Message  string        // Formatted error message
}
```

---

## Extension Guidelines

### Adding New Native Functions

1. **Define function signature** in appropriate `native/*.go` file
2. **Write contract tests** in `test/contract/` 
3. **Register in bootstrap** (`internal/bootstrap/bootstrap.go`)
4. **Update grammar** if new syntax needed

### Adding New Value Types

1. **Define type constant** in `internal/value/types.go`
2. **Implement value methods** in appropriate file
3. **Update evaluator dispatch** in `internal/eval/evaluator.go`
4. **Update parser grammar** if needed

### Function Parameter Validation

```go
// Example validation patterns:

// Arity check
if len(args) != expected {
    return value.NewNoneVal(), arityError("function", expected, len(args))
}

// Type validation  
if args[0].GetType() != value.TypeExpected {
    return value.NewNoneVal(), typeError("function", "expected-type", args[0])
}

// Value extraction with validation
count, ok := value.AsIntValue(args[0])
if !ok {
    return value.NewNoneVal(), typeError("loop", "integer for count", args[0])
}
```

### Frame and Stack Safety

**Critical**: Always use index-based references instead of pointers:

```go
// CORRECT: Index-based
type Frame struct {
    Parent int  // index into stack, not *Frame
}

// INCORRECT: Pointer-based (invalidates on stack expansion)  
type Frame struct {
    Parent *Frame  // DO NOT USE
}
```

---

## Key Design Patterns

### 1. Constructor Functions

Always use constructor functions for type safety:

```go
func NewIntVal(i int64) Value {
    return Value{Type: TypeInteger, Payload: i}
}

func NewBlockVal(elems []Value) Value {
    return Value{Type: TypeBlock, Payload: &BlockValue{Elements: elems}}
}
```

### 2. Safe Type Assertions

```go
func AsInteger(v Value) (int64, bool) {
    if v.Type != TypeInteger {
        return 0, false
    }
    return v.Payload.(int64), true
}
```

### 3. Frame Binding Pattern

```go
// Binding variables in current frame
currentFrame := e.currentFrame()
currentFrame.Bind(wordStr, result)

// Looking up variables  
resolved, found := e.Lookup(wordStr)
```

---

## Performance Characteristics

### Stack-Based Execution
- **O(1)** function calls via frame allocation
- **O(1)** variable access via index-based frames
- **O(n)** block evaluation (linear in block size)

### Memory Management
- **Copy-on-write** for immutable values
- **Reference counting** for frame cleanup
- **Stack expansion** handled transparently

### Type Dispatch
- **O(1)** type tag lookup
- **O(1)** native function calls
- **O(n)** user function calls (due to block evaluation)

---

## Testing Strategy

### Contract Tests (`test/contract/`)
- **Native function validation**: Each function has comprehensive contract tests
- **Type safety testing**: Verify correct error handling for invalid inputs
- **Integration testing**: Cross-function interaction tests

### Integration Tests (`test/integration/`)
- **End-to-end scenarios**: Complete program execution tests
- **Error handling**: Invalid program behavior tests
- **Performance benchmarking**: Execution time and memory usage

---

## Future Extensions

### Planned Features (Feature 002)
- **High-precision decimals**: Financial/scientific calculations
- **Sandboxed I/O**: File and network operations with security
- **Parse dialect**: Pattern matching and data transformation
- **Enhanced observability**: Advanced tracing and debugging

### Architectural Improvements
- **JIT compilation**: Performance optimization for hot paths
- **Parallel execution**: Multi-core support for independent computations  
- **Module system**: Code organization and reuse
- **Foreign function interface**: Integration with other languages

---

## Debugging & Development

### Built-in Tracing
```viro
trace --on --verbose --include-args --step-level 1
; Your code here
trace --off
```

### REPL Usage
- **Interactive evaluation**: Test expressions and functions
- **Error context**: Detailed error messages with stack traces
- **Variable inspection**: Examine current frame state

### CLI Modes
- **Script execution**: `viro script.viro`
- **Expression evaluation**: `viro -c "expression"`  
- **Syntax checking**: `viro --check script.viro`

---

This knowledge base provides the foundation for understanding, extending, and debugging the Viro interpreter and language implementation.