# Viro Language Homoiconic Programming Language - Core Architecture Summary

Viro is a homoiconic programming language with unique features that make it distinct from traditional languages. Here's the comprehensive understanding for agent reference:

## Key Language Characteristics

### Homoiconic Nature (Code = Data)
- Code and data share identical representation
- Blocks `[...]` represent both code and data structures
- Same AST structures for both program logic and data manipulation

### Evaluation Model
- **Left-to-Right Evaluation**: No operator precedence, strict left-to-right parsing
- **3 + 4 * 2** evaluates as **((3 + 4) * 2) = 14** (not 11 like traditional math)
- **3 + (4 * 2)** = **11** (parentheses override left-to-right)

### Blocks vs Parens
- **Blocks `[...]`**: Deferred evaluation, return themselves unchanged
- **Parens `(...)`**: Immediate evaluation, evaluate contents and return result

### Word Types System
- **Word**: `x` - evaluates to bound value
- **Set-Word**: `x:` - assignment operator  
- **Get-Word**: `:x` - fetch without evaluation
- **Lit-Word**: `'x` - returns word itself (quoted)

## Interpreter Architecture

### Core Components
1. **Evaluator** (`internal/eval/evaluator.go`): Type-based dispatch system
2. **Stack** (`internal/stack/stack.go`): Unified data/frame storage
3. **Frame** (`internal/frame/frame.go`): Variable scoping and binding
4. **Value System** (`internal/value/`): Tagged union types
5. **Tokenizer** (`internal/tokenize/`): Lexical analysis producing tokens
6. **Parser** (`internal/parse/`): Semantic analysis producing values

### Type-Based Dispatch
The evaluator uses type tags to determine evaluation behavior:
```go
switch element.GetType() {
case value.TypeInteger, value.TypeString, value.TypeLogic:
    // Return literal as-is
case value.TypeBlock:
    // Return block unchanged (deferred)
case value.TypeParen:
    // Evaluate block contents (immediate)
case value.TypeWord:
    // Look up symbol and evaluate
case value.TypeSetWord:
    // Evaluate next expr, bind to symbol
}
```

### Stack Architecture
- **Index-based references** (not pointers) for safety
- **Unified storage**: Single stack for values and frame metadata
- **Frame layout**: Return slot, prior frame, function metadata, arguments, locals

### Local-by-Default Scoping
- Variables are local unless explicitly captured as closures
- Function parameters create local bindings
- Prevents accidental global variable modification

## Native Functions System

### Implementation Pattern
```go
func FunctionName(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
    // Validation
    if len(args) != expected { return arityError(...) }
    if args[0].GetType() != value.TypeExpected { return typeError(...) }
    
    // Implementation
    result, err := implementFunction(args, refValues, eval)
    return result, err
}
```

### Key Functions
- **Control Flow**: `when`, `if`, `loop`, `while` 
- **Data Manipulation**: `set`, `get`, `type?`, `form`, `mold`, `join`
- **Object System**: `object`, `context`, `make`, `select`, `put`

### Refinements (Bash-style flags)
```viro
greet: fn [name --formal --title []] [
    ; name: positional parameter
    ; --formal: boolean refinement  
    ; --title: value refinement
    
    if formal [print ["Good day," name]]
    [print ["Hi" name "!"]]
]

; Usage
greet "Alice"              ; Hi Alice !
greet "Bob" --formal       ; Good day, Bob
greet "Carol" --title "Dr." ; Hi Dr. Carol !
```

## Extension Guidelines

### Adding New Native Functions
1. **Define signature** in `native/*.go`
2. **Write contract tests** in `test/contract/`
3. **Register in bootstrap** (`internal/bootstrap/bootstrap.go`)
4. **Update grammar** if new syntax needed

### Adding New Value Types  
1. **Define type constant** in `internal/value/types.go`
2. **Implement value methods** 
3. **Update evaluator dispatch** in `internal/eval/evaluator.go`
4. **Update parser grammar** if needed

### Critical Safety Rules
- **Always use index-based references** (not pointers) for frames/stack
- **Use constructor functions** for type safety
- **Validate types before extraction** in all operations

## Examples

### Basic Operations
```viro
x: 10                    ; Set x to 10
[x 1 + 2]                ; Block: returns [x 1 + 2] (deferred)
(x 1 + 2)                ; Paren: evaluates to 13 (immediate)
3 + 4 * 2                ; Left-to-right: 14
```

### Function Definition
```viro
square: fn [n] [n * n]   ; User-defined function
square 5                 ; Returns 25
```

### Object System
```viro
person: object [name: "Alice" age: 30]
select person "name"     ; Returns "Alice"
put person "age" 31      ; Updates age to 31
```

### Path Expressions with Eval Segments
```viro
; Static path access
obj.field              ; Direct field access
block.3                ; Direct index access

; Dynamic path access using eval segments
fields: ["name" "age" "email"]
obj.(fields.1)         ; Evaluates to obj.age
data: [10 20 30]
data.(1 + 1)           ; Evaluates to data.2

; Assignment with eval segments
key: "status"
obj.(key): "active"    ; Sets obj.status to "active"

; Why not .(expr).field?
; Leading eval segments are invalid syntax because:
; 1. Ambiguous: could be decimal or path
; 2. Unnecessary: use variables instead
;    Instead of:  .(get-obj).field
;    Write:       obj: get-obj  obj.field
```

**Note**: Eval segments are materialized once per path traversal and cached within that traversal. Each call to traverse a path evaluates eval segments fresh - there is no persistent caching across different path operations.
- **O(1)** function calls via frame allocation
- **O(1)** variable access via index-based frames  
- **O(n)** block evaluation (linear in size)
- **Stack expansion** handled transparently

This understanding provides the foundation for extending, debugging, and developing the Viro interpreter and language features.