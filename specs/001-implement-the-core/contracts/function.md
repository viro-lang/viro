# Contract: Function Definition

**Category**: Function Operations  
**Function**: `fn`  
**Purpose**: Create user-defined functions with parameters and body

---

## Native: `fn`

**Signature**: `fn [params] [body]`

**Parameters**:
- `params`: Block containing parameter specifications
- `body`: Block containing function body code

**Return**: Function value (user-defined function)

**Behavior**: 
1. Extract parameter list from first block
2. Validate parameter specifications
3. Capture body block
4. Create FunctionValue with type=User, params, body
5. Return function value

**Parameter Specification Format**:
- Simple parameter: `[n]` → parameter named `n`, any type
- Multiple parameters: `[a b c]` → three parameters
- **Refinements (Phase 1)**: Bash-style optional parameters
  - `--flag` → boolean flag (true if present, false otherwise)
  - `--option []` → value refinement accepting any type
- Phase 2+: Type constraints, docstrings
  - `--option [type]` → value refinement with type constraint
  - `--flag ["description"]` → flag with documentation
  - `--option [type "description"]` → value refinement with type and docs

**Refinement Syntax Rules**:
1. Refinements start with `--` prefix in parameter list
2. Refinements without `[]` are boolean flags
3. Refinements with `[]` accept a value (stored in variable named after refinement, without `--`)
4. At call site, refinements can appear anywhere among arguments (order independent)
5. Flag refinements set local variable to `true` (present) or `false` (absent)
6. Value refinements set local variable to provided value or `none` (if not provided)
7. Refinement names (without `--`) must be unique and not conflict with positional parameters

**Examples**:
```rebol
; Simple function
square: fn [n] [n * n]
square 5        → 25

; Multiple parameters
add: fn [a b] [a + b]
add 3 4         → 7

; No parameters
greet: fn [] [print "hello"]
greet           → prints "hello", returns none

; Local variables (local by default)
calculate: fn [x] [
    temp: x * 2   ; temp is local
    temp + 1      ; returns 11 when x=5
]
calculate 5     → 11

; Local isolation (does not modify global)
counter: 0
increment: fn [] [
    counter: counter + 1  ; creates LOCAL counter (does not modify global)
    counter               ; returns 1
]
increment       → 1
counter         → 0 (global unchanged)

; Refinements - flag only
verbose-print: fn [msg --verbose] [
    either verbose [
        print ["[INFO]" msg]
    ] [
        print msg
    ]
]
verbose-print "hello"           → prints "hello"
verbose-print "hello" --verbose → prints "[INFO] hello"

; Refinements - value refinement
greet: fn [name --title []] [
    either title [
        print [title name]
    ] [
        print name
    ]
]
greet "Alice"                 → prints "Alice"
greet "Bob" --title "Dr."     → prints "Dr. Bob"

; Mixed refinements and arguments
process: fn [data --deep --limit []] [
    ; data: positional arg
    ; deep: boolean flag
    ; limit: value refinement (integer)
    result: either deep [deep-copy data] [data]
    either limit [
        take result limit
    ] [
        result
    ]
]
process [1 2 3]                      → [1 2 3]
process [1 2 3] --limit 2            → [1 2]
process [1 2 3] --deep               → deep copy
process [1 2 3] --deep --limit 2     → deep copy, first 2
process [1 2 3] --limit 2 --deep     → same (order doesn't matter)
```

**Type Rules**:
- First argument must be Block (parameter list)
- Second argument must be Block (body)
- Positional parameters must be Word type (symbols)
- Refinements must be Word type starting with `--` prefix
- Refinement metadata (if present) must be Block containing type or docstring
- Parameter names must be unique (including refinement names without `--` prefix)

**Validation Rules**:
1. First block contains only words (parameter names) or refinement-words (--name)
2. Refinements must start with `--` prefix
3. If refinement followed by block `[]`, it's a value refinement; otherwise it's a flag
4. No duplicate parameter names (positional or refinement, comparing without `--`)
5. Body block can contain any valid expressions
6. Both parameter and body blocks must be provided (not none)

**Test Cases**:
1. `fn [n] [n * n]` creates function, calling with 5 returns 25
2. `fn [a b] [a + b]` creates function, calling with 3 and 4 returns 7
3. `fn [] [42]` creates function with no params, calling returns 42
4. `fn [x y z] [x + y + z]` creates function with 3 params
5. `fn "invalid" [42]` errors (not block)
6. `fn [42] [body]` errors (parameter not word)
7. `fn [x x] [x]` errors (duplicate parameter names)
8. `fn [x] 42` errors (body not block)
9. Local isolation: `x: 100  test: fn [] [x: 5]  test  x` → global x still 100
10. Local variables: `fn [n] [temp: n * 2  temp]` → temp is local, not accessible outside
11. Flag refinement: `fn [msg --verbose] [...]` with `call "hi" --verbose` → verbose=true
12. Value refinement: `fn [name --title []] [...]` with `call "Alice" --title "Dr."` → title="Dr."
13. Missing refinement: `fn [name --title []] [...]` with `call "Bob"` → title=none
14. Mixed order: `fn [a b --flag] [...]` can be called `call 1 --flag 2` or `call --flag 1 2`
15. Duplicate names: `fn [title --title []] [...]` errors (title used twice)

---

## Function Execution Semantics

**Call Process**:
1. Evaluator recognizes Function type during dispatch
2. Scans arguments to separate positional args from refinements
3. Collects positional arguments (evaluates expressions, stopping at `--` tokens)
4. Collects refinements: flag refinements set to true, value refinements collect next value
5. Validates positional argument count matches required parameter count
6. Creates new Frame for function execution
7. Binds positional parameters to positional argument values
8. Binds refinements: flags to true/false, value refinements to provided value or none
9. Evaluates body block in frame context
10. Returns last value from body evaluation
11. Destroys frame

**Refinement Argument Collection**:
- When `--name` token encountered during argument collection:
  - If `name` is flag refinement: set `name` local variable to `true`
  - If `name` is value refinement: evaluate next expression, set `name` to result
  - If `name` not defined in function: error "Unknown refinement: --name"
- Refinements not provided at call site:
  - Flag refinements default to `false`
  - Value refinements default to `none`

**Frame Creation**:
- Frame type: FrameFunctionArgs
- Words: parameter names + refinement names (without `--` prefix)
- Values: positional argument values + refinement values (true/false/value/none)
- Parent: current evaluation frame (lexical scoping for closures)
- **Local-by-default**: All words used in body are local unless captured from parent

**Example Execution**:
```rebol
square: fn [n] [n * n]
square 5

; Execution steps:
; 1. Eval "square" → FunctionValue
; 2. Recognize function type
; 3. Collect argument: eval "5" → Integer(5)
; 4. Create frame: Words=[n], Values=[5], Parent=current_frame
; 5. Eval body [n * n] in new frame:
;    - Eval "n" → lookup in frame → 5
;    - Eval "*" → NativeMultiply
;    - Eval "n" → 5
;    - Call NativeMultiply([5, 5]) → 25
; 6. Return 25
; 7. Destroy frame

; Refinement example:
greet: fn [name --formal --title []] [
    either formal [
        either title [print [title name]] [print name]
    ] [
        print ["Hi" name]
    ]
]
greet "Alice" --formal --title "Dr."

; Execution steps:
; 1. Eval "greet" → FunctionValue(params=[name, --formal, --title[]])
; 2. Scan arguments: "Alice", --formal, --title, "Dr."
; 3. Collect positional: ["Alice"]
; 4. Collect refinements: formal=true (flag), title="Dr." (value from next token)
; 5. Validate: 1 positional arg matches 1 required param
; 6. Create frame: Words=[name, formal, title], Values=["Alice", true, "Dr."]
; 7. Eval body with bindings: name="Alice", formal=true, title="Dr."
; 8. Result: prints "Dr. Alice"
```

**Closure Semantics** (Phase 1):
- Functions capture lexical environment (parent frame reference)
- **Local-by-default scoping**: All words in function body are local unless captured from outer scope
- Parameters are automatically local
- Assignment (`set` or set-word) creates local variables
- Functions do NOT access or modify global variables by default

**Local-by-Default Example**:
```rebol
x: 100                  ; global x
test: fn [] [
    x: 5                ; creates LOCAL x (does not modify global)
    x                   ; returns 5
]
test                    ; returns 5
x                       ; still 100 (global unchanged)
```

**Closure Example** (capturing outer scope):
```rebol
make-adder: fn [x] [
    ; x is local to make-adder
    fn [y] [
        ; Inner function captures x from outer scope
        ; y is local to inner function
        x + y           ; x from closure, y from parameter
    ]
]
add5: make-adder 5      ; captures x=5
add5 10                 ; returns 15 (closure captures x=5)
```

---

## Error Cases

**Definition Errors** (at `fn` call):
1. First argument not block → Script error (300): "Fn expects block for parameters"
2. Second argument not block → Script error: "Fn expects block for body"
3. Parameter not word → Script error: "Parameter must be word, got {type}"
4. Refinement not starting with `--` → Script error: "Invalid refinement syntax: {name}"
5. Duplicate parameter names → Script error: "Duplicate parameter name: {name}"
6. Refinement name conflicts with positional param → Script error: "Refinement name conflicts: {name}"
7. Empty parameter list → Valid (no-argument function)

**Execution Errors** (at function call):
1. Positional argument count mismatch → Script error: "Expected {N} arguments, got {M}"
2. Unknown refinement → Script error: "Unknown refinement: --{name}"
3. Value refinement without value → Script error: "Refinement --{name} requires a value"
4. Body evaluation error → Propagate error with function context in "where"

**Example Error Messages**:
```rebol
square: fn [n] [n * n]
square          → Error: Expected 1 arguments, got 0
square 1 2      → Error: Expected 1 arguments, got 2

bad: fn [42] [x]
                → Error: Parameter must be word, got integer!

test: fn [x --x []] [x]
                → Error: Refinement name conflicts: x

greet: fn [name --title []] [name]
greet "Alice" --unknown
                → Error: Unknown refinement: --unknown

greet "Bob" --title
                → Error: Refinement --title requires a value
```

---

## Stack Frame Layout

**Function Call Frame** (pushed onto stack):
```
[index]     [content]
frameBase   Return value slot (initially none)
+1          Prior frame pointer (index of calling frame)
+2          Function metadata (FunctionValue reference)
+3          Positional argument 1 value
+4          Positional argument 2 value
...         Additional positional arguments
+3+N        Refinement values (flags: true/false, options: value/none)
+3+N+R      Local variables (created via set during body evaluation)
```

**Frame Access**:
- Positional parameters: mapped to stack slots frameBase+3 onwards
- Refinements: mapped to slots after positional args (frameBase+3+posArgCount)
- Local variables: allocated on stack as `set` is called
- Return: last body evaluation result placed in frameBase slot

**Example Frame Layout**:
```rebol
greet: fn [name --formal --title []] [...]
greet "Alice" --formal --title "Dr."

Frame layout:
[0]  Return slot (none initially)
[1]  Prior frame pointer
[2]  FunctionValue reference
[3]  "Alice" (positional: name)
[4]  true (refinement: formal)
[5]  "Dr." (refinement: title)
[6+] Local variables (if any created during execution)
```

---

## Integration with Evaluator

**Type Dispatch**:
```go
func (e *Evaluator) EvalValue(v Value) (Value, error) {
    switch v.Type {
    case TypeFunction:
        return e.CallFunction(v)
    // ... other types
    }
}

func (e *Evaluator) CallFunction(fn Value) (Value, error) {
    funcVal := fn.Payload.(*FunctionValue)
    
    // Collect arguments
    args := e.CollectArguments(funcVal.Params)
    
    // Validate count
    if len(args) != len(funcVal.Params) {
        return nil, ArgCountError(...)
    }
    
    // Create frame
    frame := e.stack.NewFrame(funcVal, args)
    
    // Bind parameters
    for i, param := range funcVal.Params {
        frame.Set(param.Name, args[i])
    }
    
    // Evaluate body
    result, err := e.EvalBlock(funcVal.Body, frame)
    
    // Cleanup
    e.stack.DestroyFrame()
    
    return result, err
}
```

---

## Test Strategy

**Unit Tests** (func native):
1. Parameter extraction from block
2. Body capture
3. FunctionValue construction
4. Error cases (invalid parameters, duplicate names)

**Integration Tests** (function calls):
1. Simple function definition and call
2. Multiple parameters
3. No parameters
4. Nested function calls
5. Recursive functions (up to depth 100 per spec)
6. Closures (capture outer scope)
7. Error propagation from body
8. Argument count validation

**Example Test**:
```go
func TestFunctionDefinitionAndCall(t *testing.T) {
    tests := []struct {
        name   string
        def    string       // function definition
        call   string       // function call
        want   Value
    }{
        {"simple", "square: fn [n] [n * n]", "square 5", IntVal(25)},
        {"multi-param", "add: fn [a b] [a + b]", "add 3 4", IntVal(7)},
        {"no-param", "f: fn [] [42]", "f", IntVal(42)},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            eval := NewEvaluator()
            eval.Eval(tt.def)   // define function
            result, err := eval.Eval(tt.call)  // call function
            
            assert.NoError(t, err)
            assert.Equal(t, tt.want, result)
        })
    }
}
```

---

## Implementation Checklist

- [ ] Parse parameter block (extract word names)
- [ ] Validate parameters (words only, unique names)
- [ ] Create FunctionValue structure
- [ ] Register function in native dispatcher
- [ ] Argument collection during call
- [ ] Argument count validation
- [ ] Frame creation with parameter binding
- [ ] Body evaluation in frame context
- [ ] Return value handling
- [ ] Frame cleanup
- [ ] All test cases pass

**Dependencies**:
- Value system (FunctionValue, Block, Word types)
- Frame system (frame creation, binding, parent chain)
- Stack system (frame layout, push/pop)
- Evaluator (block evaluation, type dispatch)
- Error system (Script error for validation)

---

## Advanced Features (Future Phases)

**Type Constraints** (deferred):
```rebol
; Phase 2+: typed parameters
typed-add: fn [a [integer!] b [integer!]] [a + b]
typed-add 3 4       → 7
typed-add "x" 4     → Error: Type mismatch for parameter 'a'
```

**Optional Parameters** (deferred):
```rebol
; Phase 2+: optional with defaults
greet: fn [name [string!] /title [string!]] [
    either title [
        print [title name]
    ] [
        print name
    ]
]
greet "Alice"              → prints "Alice"
greet/title "Alice" "Dr."  → prints "Dr. Alice"
```

**Refinements** (deferred):
```rebol
; Phase 2+: function refinements
copy/deep data      ; copy with /deep refinement
```

**Local Variables** (Phase 1 - local by default):
```rebol
; All words in function body are local by default
calc: fn [x] [
    temp: x * 2       ; temp is LOCAL (created on first assignment)
    result: temp + 1  ; result is LOCAL
    result            ; return local result
]
calc 5                ; returns 11
temp                  ; Error: No value for word 'temp' (temp is local to calc)

; Compare with global scope
global-val: 100
test: fn [] [
    global-val: 50    ; creates LOCAL global-val (does NOT modify global)
    global-val        ; returns 50
]
test                  ; returns 50
global-val            ; still 100 (global unchanged)
```

**Rationale**: Local-by-default prevents accidental modification of global state, a common source of bugs in REBOL. Functions are isolated by default, making them safer and more predictable.

**Current Phase 1 Scope**:
- Simple parameters (word list, no types)
- Body block (any expressions)
- Lexical scoping (closure support)
- Argument count validation
- Frame-based execution
