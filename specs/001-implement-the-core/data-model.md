# Data Model: Viro Core Interpreter

**Feature**: Viro Core Language and REPL  
**Date**: 2025-01-07  
**Source**: Extracted from spec.md Key Entities section

## Overview

The Viro interpreter data model implements Viro's type-based evaluation system with eight core entities organized in three layers:

1. **Value Layer**: Type-tagged values representing all data
2. **Execution Layer**: Stack and frames for evaluation state
3. **Error Layer**: Structured error representation

All entities designed for index-based access per constitution Principle IV (Stack and Frame Safety).

---

## Entity Definitions

### 1. Value

**Purpose**: Universal data representation with type discrimination and payload storage.

**Structure**:
```go
type ValueType uint8

const (
    TypeNone ValueType = iota
    TypeLogic
    TypeInteger
    TypeString
    TypeWord
    TypeSetWord
    TypeGetWord
    TypeLitWord
    TypeBlock
    TypeParen
    TypeFunction
)

type Value struct {
    Type    ValueType
    Payload interface{}  // type-specific data
}
```

**Fields**:
- `Type`: Discriminator identifying value type (0-10 for initial types)
- `Payload`: Type-specific data (integer, string, word symbol, block reference, paren reference, function reference)

**Type-Specific Payloads**:
- `TypeNone`: nil payload (represents absence of value)
- `TypeLogic`: bool (true/false)
- `TypeInteger`: int64 (64-bit signed integers)
- `TypeString`: *StringValue (character sequence, see String entity)
- `TypeWord`: string (symbol name, case-sensitive)
- `TypeSetWord`: string (symbol name for assignment)
- `TypeGetWord`: string (symbol name for fetch)
- `TypeLitWord`: string (quoted symbol name)
- `TypeBlock`: *BlockValue (series of values, deferred evaluation)
- `TypeParen`: *BlockValue (series of values, immediate evaluation)
- `TypeFunction`: *FunctionValue (native or user-defined)

**Validation Rules**:
- Type must be valid constant from TypeNone to TypeFunction
- Payload must match type (enforced by constructor functions)
- Words are case-sensitive: `x` and `X` are distinct symbols

**State Transitions**:
- Values are generally immutable (create new value for modifications)
- Exception: Series (blocks, strings) support in-place modification

**Relationships**:
- Block values contain child Value references
- Function values reference Block for body
- Word values resolve to other Values via Frame binding

---

### 2. Block

**Purpose**: Ordered sequence of values, fundamental composite type for code and data. **Blocks evaluate to themselves** (deferred evaluation).

**Structure**:
```go
type BlockValue struct {
    Elements []Value  // ordered value sequence
    Index    int      // current series position (for iteration)
}
```

**Fields**:
- `Elements`: Slice of Value objects in order
- `Index`: Current position for series operations (0-based)

**Operations**:
- Access: `First()`, `Last()`, `At(index)`, `Length()`
- Modification: `Append(value)`, `Insert(value)`, `Remove()`
- Iteration: Advance index, check bounds
- Evaluation: **Blocks evaluate to themselves** (return same block without evaluating contents)

**Validation Rules**:
- Empty blocks allowed (zero elements)
- Index must be 0 ≤ index ≤ Length()
- Elements are heterogeneous (mixed types allowed)

**State Transitions**:
- Created with initial elements or empty
- Modified via append/insert/remove operations
- Index advances during iteration or evaluation
- Supports copy-on-write semantics where practical

**Evaluation Semantics**:
- Syntax: `[element1 element2 ...]`
- When a block is evaluated, it returns itself **without evaluating its contents**
- Example: `x: [1 + 2]` stores block `[1 + 2]`, not value `3`
- To evaluate block contents: use `do` function or pass to control flow functions
- Blocks as data: `[red green blue]`, `[1 2 3 4 5]`
- Blocks as code: passed to `when`, `if`, `loop`, `fn` for deferred evaluation

**Relationships**:
- Elements are Value objects (any type)
- Function bodies are Blocks
- Evaluation contexts hold current Block being evaluated

---

### 3. Paren

**Purpose**: Ordered sequence of values with **immediate evaluation**. Parens are evaluated when encountered, returning the result of evaluating their contents.

**Structure**:
```go
// Parens use the same BlockValue structure as blocks
// The distinction is in the Type tag (TypeParen vs TypeBlock)
```

**Fields**:
- Same as Block: `Elements` and `Index`
- Distinguished by `Type == TypeParen` in Value wrapper

**Operations**:
- Same series operations as Block (Access, Modification, Iteration)
- **Different evaluation behavior**: evaluates contents immediately

**Validation Rules**:
- Same as Block (empty allowed, heterogeneous elements, index bounds)

**Evaluation Semantics**:
- Syntax: `(element1 element2 ...)`
- When a paren is evaluated, it **evaluates its contents** and returns the last result
- Example: `x: (1 + 2)` stores value `3`, not block `(1 + 2)`
- Useful for: immediate calculation, nested expressions, dynamic values
- Example: `print ["Result:" (1 + 2)]` prints "Result: 3"
- Example: `when (x > 10) [print "large"]` evaluates condition immediately

**Comparison with Block**:
| Feature | Block `[...]` | Paren `(...)` |
|---------|--------------|---------------|
| Evaluation | Deferred (returns self) | Immediate (evaluates contents) |
| Usage | Data structures, code storage | Calculations, conditions |
| Example | `[1 + 2]` → `[1 + 2]` | `(1 + 2)` → `3` |
| Common use | Function bodies, lists | Inline expressions, dynamic access |

**Future Use Cases** (post-Phase 1):
- Path with dynamic index: `array.(index + 1)` 
- Computed object fields: `obj.(fieldname)`

**Relationships**:
- Shares BlockValue structure with Block
- Distinguished only by Type tag
- Both can contain any Value types

---

### 4. Word

**Purpose**: Symbolic identifier referencing values in evaluation context.

**Structure**:
```go
type WordValue struct {
    Symbol  string  // identifier name (case-sensitive)
    Binding *Frame  // bound frame (nil if unbound)
}
```

**Fields**:
- `Symbol`: String identifier (case-sensitive, e.g., "x", "X", "my-var")
- `Binding`: Reference to Frame containing value (index-based internally)

**Word Types** (distinguished by Value.Type):
- **Word** (TypeWord): Evaluates to bound value
- **Set-Word** (TypeSetWord): Assigns next evaluation result
- **Get-Word** (TypeGetWord): Fetches value without evaluation
- **Lit-Word** (TypeLitWord): Returns word itself (quoted)

**Validation Rules**:
- Symbol must be non-empty string
- Symbol is case-sensitive (per spec clarification)
- Valid symbol characters: letters, digits, hyphens, special chars (no spaces)
- Binding can be nil (unbound word triggers error on evaluation)

**State Transitions**:
- Created unbound during parsing
- Bound to Frame during context creation or binding operations
- Binding persists for word lifetime

**Evaluation Behavior**:
- Word: Look up Symbol in Binding frame, evaluate result
- Set-Word: Evaluate next expression, store in Binding frame at Symbol
- Get-Word: Look up Symbol in Binding frame, return without evaluation
- Lit-Word: Return word value itself

**Relationships**:
- Bound to Frame (many words → one frame)
- Resolves to any Value type via binding

---

### 5. Function

**Purpose**: Executable entity with formal parameters and body block.

**Structure**:
```go
type FunctionType uint8

const (
    FuncNative FunctionType = iota
    FuncUser
)

type FunctionValue struct {
    Type   FunctionType
    Name   string        // function name (for debugging/errors)
    Params []ParamSpec   // formal parameter specifications
    Body   *BlockValue   // body block (nil for natives)
    Native func([]Value) (Value, error)  // native implementation (nil for user functions)
}

type ParamSpec struct {
    Name       string      // parameter name (without -- prefix for refinements)
    Type       ValueType   // expected type (TypeNone = any)
    Optional   bool        // whether parameter is optional
    Refinement bool        // whether this is a refinement (--flag or --option)
    TakesValue bool        // for refinements: true if accepts value, false if boolean flag
}
```

**Parameter Types**:
- **Positional parameters**: Regular arguments (e.g., `name`, `value`)
- **Flag refinements**: Boolean switches (e.g., `--verbose`, `--deep`)
- **Value refinements**: Options accepting values (e.g., `--title []`, `--count [integer]`)

**Refinement Syntax in Definition**:
- `--flag` → boolean refinement (true if present, false otherwise)
- `--option []` → value refinement accepting any type
- `--option [type]` → value refinement accepting specific type (Phase 2+)
- `--option ["docstring"]` → flag with documentation (Phase 2+)
- `--option [type "docstring"]` → value refinement with type and docs (Phase 2+)

**Fields**:
- `Type`: Native (built-in) or User (func-defined)
- `Name`: Function identifier for error messages and debugging
- `Params`: Ordered list of parameter specifications
- `Body`: Block containing function body (for user functions)
- `Native`: Go function pointer (for native functions)

**Validation Rules**:
- Name should be non-empty for error reporting
- Native functions must have Native field set, Body nil
- User functions must have Body set, Native nil
- Parameter names must be unique within function
- Parameter types validated during argument evaluation

**State Transitions**:
- Native functions created during interpreter initialization
- User functions created via `fn` native
- Functions are immutable after creation (per functional programming principles)

**Execution Flow**:
1. Caller pushes arguments onto stack
2. Function creates new Frame for local variables and parameters
3. Arguments bound to parameter names in Frame (parameters are local)
4. **All words used in body are bound to local Frame by default** - words do NOT access global context unless explicitly captured before function definition
5. Body block evaluated in Frame context (user functions) or Native called (native functions)
6. Return value placed in designated stack slot
7. Frame destroyed, stack unwound

**Local-by-Default Semantics**:
- Parameter names are automatically local to function
- Any word used in function body creates a local variable if not already bound
- Functions do NOT modify global variables by accident
- Explicit global access requires capturing global word value before function definition (closures)

**Example** (local isolation):
```viro
x: 100                    ; global x
test: fn [] [x: 5  x]     ; x is LOCAL to test, not global
test                      ; returns 5
x                         ; still 100 (global x unchanged)
```

**Example** (refinements):
```viro
; Function with flag and value refinements
greet: fn [name --formal --title []] [
    ; name: positional parameter
    ; formal: boolean flag (true if --formal passed, false otherwise)
    ; title: value refinement (holds value if --title VALUE passed, none otherwise)
    
    if formal [
        if title [
            print [title name ", good day!"]
        ] [
            print ["Good day," name]
        ]
    ] [
        when title [
            print ["Hi" title name "!"]
        ]
        when (not title) [
            print ["Hi" name "!"]
        ]
    ]
]

; Usage:
greet "Alice"                          ; Hi Alice ! (formal=false, title=none)
greet "Bob" --formal                   ; Good day, Bob (formal=true, title=none)
greet "Carol" --title "Dr."            ; Hi Dr. Carol ! (formal=false, title="Dr.")
greet "Dave" --formal --title "Prof."  ; Prof. Dave , good day! (both set)
greet "Eve" --title "Ms." --formal     ; Ms. Eve , good day! (order doesn't matter)
```

**Relationships**:
- Body is BlockValue (for user functions)
- Execution creates Frame for parameter binding
- Native registry maps names to FunctionValue objects

---

### 6. Frame

**Purpose**: Variable storage container mapping symbols to values. Provides local-by-default scoping for function execution.

**Structure**:
```go
type FrameType uint8

const (
    FrameFunctionArgs FrameType = iota
    FrameClosure
    // FrameObject, FrameModule deferred to later phases
)

type Frame struct {
    Type      FrameType
    Words     []string   // symbol names (parallel to Values)
    Values    []Value    // bound values (parallel to Words)
    Parent    int        // index of parent frame (-1 if none)
}
```

**Fields**:
- `Type`: Frame category (function args, closure, object, module)
- `Words`: List of word symbols defined in this frame (local variables and parameters)
- `Values`: List of values bound to corresponding word symbols (parallel array)
- `Parent`: Index of enclosing frame for lexical scoping (-1 for global)

**Operations**:
- `Bind(word)`: Associate word with this frame (makes it local)
- `Get(symbol)`: Retrieve value for symbol (local frame only, does NOT search parent by default)
- `Set(symbol, value)`: Store value for symbol (creates local if needed)
- `HasWord(symbol)`: Check if symbol defined in this frame

**Local-by-Default Scoping**:
- When function executes, all words in body are bound to local Frame
- Parameters are automatically added to Frame as local variables
- Assignment creates local variable if word not already in Frame
- Global words are NOT accessed unless explicitly captured via closure
- This prevents accidental modification of global state

**Example**:
```viro
counter: 0                          ; global counter
increment: fn [n] [                 ; n is local parameter
    temp: n + 1                     ; temp is local variable (created on assignment)
    counter: counter + 1            ; counter is LOCAL (shadows global)
    temp                            ; return local temp
]
increment 5                         ; returns 6, global counter still 0
```

**Validation Rules**:
- Words and Values arrays must have same length
- Symbol names are case-sensitive
- Parent must be valid frame index or -1
- No duplicate symbols within single frame

**Frame Types**:
- **FunctionArgs**: Created for function calls, destroyed on return
- **Closure**: Captures lexical environment, persists beyond function return
- **Object/Module**: Deferred to later phase per spec Out of Scope

**State Transitions**:
- Created when function called (FunctionArgs) or closure captured
- Modified via Set operations (add new word or update existing)
- Destroyed when function returns (FunctionArgs) or garbage collected (Closure)

**Relationships**:
- Contains Values of any type
- Referenced by Word bindings (via index)
- Linked to Parent frame forming scope chain
- Stored on Stack during function execution

---

### 7. Stack

**Purpose**: Unified storage for data values and function call frames.

**Structure**:
```go
type Stack struct {
    Data         []Value  // unified storage for values and frame metadata
    Top          int      // index of next available slot
    CurrentFrame int      // index of current function frame (-1 if top level)
}
```

**Fields**:
- `Data`: Slice holding all values and frame structures
- `Top`: Next available stack slot (0-based index)
- `CurrentFrame`: Index where current frame begins (-1 for top-level evaluation)

**Operations**:
- `Push(value)`: Add value to stack top, increment Top
- `Pop()`: Remove value from stack top, decrement Top
- `Get(index)`: Retrieve value at absolute index (safe across expansions)
- `Set(index, value)`: Update value at absolute index
- `NewFrame(function, argCount)`: Allocate frame structure on stack
- `DestroyFrame()`: Unwind frame, restore previous frame as current

**Stack Frame Layout** (when function active):
```
[index]     [content]
frameBase   Return value slot (initially none)
+1          Prior frame pointer (index or -1)
+2          Function metadata (FunctionValue)
+3          Argument 1
+4          Argument 2
...         Additional arguments
+3+N        Local variables (grows as needed)
```

**Validation Rules**:
- Top always points to next available slot (0 ≤ Top ≤ Capacity)
- CurrentFrame points to valid frame base or -1
- Frame base must have proper layout (return slot, prior frame, function, args)
- Index-based access only (no direct slice pointers per constitution Principle IV)

**State Transitions**:
- Initially empty (Top=0, CurrentFrame=-1)
- Values pushed/popped during evaluation
- Frames allocated on function call, destroyed on return
- Automatic expansion when capacity reached (via Go slice append)

**Expansion Behavior**:
- Triggered when Top >= len(Data)
- Go slice semantics: doubles capacity up to 1024 elements, then ~1.25x growth
- Index-based access ensures existing frame pointers remain valid
- Expansion transparent to caller (per success criteria SC-009)

**Relationships**:
- Stores Value objects
- Stores Frame metadata in designated slots
- Referenced by Frame.Parent indices
- Managed by Evaluator during execution

---

### 8. Error

**Purpose**: Structured error representation with category, context, and diagnostic information.

**Structure**:
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
    Code     int           // specific error code within category
    ID       string        // error identifier (e.g., "no-value", "div-zero")
    Args     [3]string     // arguments for message interpolation
    Near     []Value       // expressions around error location (3 before, 3 after)
    Where    []string      // call stack trace (function names)
    Message  string        // formatted error message
}
```

**Fields**:
- `Category`: Error class (0-900 range per constitution Principle V)
- `Code`: Specific error within category (e.g., 301 = no value for word)
- `ID`: Symbolic error name for programmatic handling
- `Args`: Up to 3 string arguments for message interpolation
- `Near`: Value sequence showing error location (window around current expression)
- `Where`: Call stack showing function invocation chain
- `Message`: Human-readable error description in English (per spec clarification)

**Error Categories**:
- **Throw (0)**: `break` outside loop, `continue` without loop context
- **Note (100)**: Non-fatal warnings (future use)
- **Syntax (200)**: Invalid source syntax, unclosed blocks, malformed literals
- **Script (300)**: Undefined words, type mismatches, invalid operations
- **Math (400)**: Division by zero, overflow, invalid math operations
- **Access (500)**: File access errors (future), permission errors
- **Internal (900)**: Stack overflow, out of memory, interpreter bugs

**Validation Rules**:
- Category must be one of defined constants
- Code should be within category range (e.g., 300-399 for Script)
- ID should be kebab-case identifier
- Args used for %1, %2, %3 placeholders in message
- Near should show relevant expressions (up to 7 values)
- Where should show function call chain (most recent first)

**Message Interpolation Examples**:
- ID="no-value", Args=["x", "", ""] → "No value for word: x"
- ID="div-zero", Args=["", "", ""] → "Division by zero"
- ID="type-mismatch", Args=["+", "string", "integer"] → "Type mismatch for '+': cannot add string and integer"

**State Transitions**:
- Created when error condition detected
- Captured Near context from current evaluator position
- Captured Where context from current stack frame chain
- Formatted message generated once on creation
- Propagated up call stack until caught or displayed in REPL

**Relationships**:
- Near contains Value objects from evaluation context
- Where derived from Frame parent chain in Stack
- Displayed by REPL error handler
- Implements Go error interface for language compatibility

---

### 9. Series (Abstract)

**Purpose**: Abstract interface for sequential data types (blocks, strings).

**Interface**:
```go
type Series interface {
    First() Value      // First element
    Last() Value       // Last element
    At(index int) Value    // Element at position
    Length() int       // Number of elements
    Append(value Value)    // Add to end
    Insert(value Value)    // Insert at current position
    Remove()           // Remove at current position
    Index() int        // Current position
    SetIndex(int)      // Move position
}
```

**Implementations**:
- **BlockValue**: Elements are Values, operations work on Value array
- **StringValue**: Elements are runes (characters), operations work on []rune

**Operations** (common across implementations):
- `First()`: Returns element at index 0 (error if empty)
- `Last()`: Returns element at index Length()-1 (error if empty)
- `At(index)`: Returns element at position (bounds-checked)
- `Length()`: Returns count of elements
- `Append(value)`: Adds element to end, updates length
- `Insert(value)`: Adds element at current Index, shifts remaining elements
- `Remove()`: Removes element at current Index, shifts remaining elements
- `Index()`: Returns current position (for iteration)
- `SetIndex(pos)`: Sets current position (bounds-checked)

**Validation Rules**:
- Empty series allowed (Length() = 0)
- Index must be 0 ≤ index < Length()
- First/Last operations error on empty series
- Mutation operations modify series in-place (series values are mutable)

**Usage Pattern** (code examples):
```viro
data: [1 2 3]       ; BlockValue with 3 elements
first data          ; returns 1
last data           ; returns 3
append data 4       ; modifies to [1 2 3 4]
length? data        ; returns 4

str: "hello"        ; StringValue with 5 characters
first str           ; returns 'h'
append str " world" ; modifies to "hello world"
```

**Relationships**:
- Blocks implement Series interface
- Strings implement Series interface
- Native functions (first, last, append, etc.) accept any Series
- Type system validates Series operations at runtime

---

## Entity Relationships

```
Value
  ├─ contains → Block (TypeBlock)
  │   └─ contains → []Value (recursive)
  ├─ contains → Word (TypeWord/SetWord/GetWord/LitWord)
  │   └─ references → Frame (binding)
  ├─ contains → Function (TypeFunction)
  │   └─ contains → Block (body)
  └─ contains → primitives (TypeNone, TypeLogic, TypeInteger, TypeString)

Stack
  ├─ stores → []Value (unified data)
  └─ stores → Frame metadata (in designated slots)
      └─ references → prior Frame (via index)

Frame
  ├─ stores → []string (word symbols)
  ├─ stores → []Value (bound values)
  └─ references → parent Frame (via index)

Error
  ├─ captures → []Value (Near context)
  └─ captures → []string (Where context from Frame chain)

Series (interface)
  ├─ implemented by → Block
  └─ implemented by → String
```

**Key Design Properties**:
- **Index-based references**: Frames and Stack use integer indices, not pointers (constitution Principle IV)
- **Type-tagged values**: Single Value type with discriminated union (constitution Principle III)
- **Immutable primitives**: Integers, logic, none are immutable; series are mutable
- **Structured errors**: Category, code, context per constitution Principle V
- **Lexical scoping**: Frame parent chain supports closure semantics

---

## Data Flow Example

Code example: `square: fn [n] [n * n]` then `square 5`

1. **Parse**: Creates Block `[square: fn [n] [n * n]]`
   - Value 1: SetWord("square")
   - Value 2: Word("fn")
   - Value 3: Block `[n]` (parameter list)
   - Value 4: Block `[n * n]` (body)

2. **Eval SetWord "square"**:
   - Evaluates next expression (fn call)
   - Stores result in current Frame at "square"

3. **Eval Word "fn"**:
   - Resolves to native Function
   - Calls native with args: Block `[n]`, Block `[n * n]`
   - Returns new FunctionValue (user function)

4. **Store Function**:
   - Frame.Set("square", FunctionValue)
   - Word "square" now bound to function

5. **Parse and Eval**: `square 5`
   - Block: [Word("square"), Integer(5)]
   - Eval Word "square" → FunctionValue
   - Type dispatch recognizes Function type
   - Collect argument: Integer(5)

6. **Function Call**:
   - Stack.NewFrame(square_func, 1)
   - Frame layout: [return_slot, prior_frame, function, arg_n=5]
   - Bind word "n" to Frame (local parameter)
   - **All words in body are local by default** - word "n" resolves to local Frame, not global
   - Eval body Block `[n * n]`
     - Eval Word "n" → Integer(5) (from local Frame)
     - Eval Word "*" → NativeMultiply function
     - Eval Word "n" → Integer(5) (from local Frame)
     - Call NativeMultiply([5, 5]) → Integer(25)
   - Store result in return_slot
   - Stack.DestroyFrame()
   - Return Integer(25)

7. **REPL Display**: `25`

**Data Model Entities Used**:
- Value (word, integer, block, function types)
- Block (parameter list, body, argument collection)
- Word (symbol resolution via Frame binding)
- Function (user-defined with parameters and body)
- Frame (argument binding for "n")
- Stack (frame allocation, argument passing, return value)

---

## Implementation Notes

### Value Construction

Use constructor functions to ensure type/payload consistency:

```go
func NoneVal() Value { return Value{Type: TypeNone} }
func LogicVal(b bool) Value { return Value{Type: TypeLogic, Payload: b} }
func IntVal(i int64) Value { return Value{Type: TypeInteger, Payload: i} }
func StrVal(s string) Value { return Value{Type: TypeString, Payload: NewStringValue(s)} }
func WordVal(sym string) Value { return Value{Type: TypeWord, Payload: sym} }
func BlockVal(elems []Value) Value { return Value{Type: TypeBlock, Payload: &BlockValue{Elements: elems}} }
func FuncVal(f *FunctionValue) Value { return Value{Type: TypeFunction, Payload: f} }
```

### Type Checking

Use type assertions safely with validation:

```go
func AsInteger(v Value) (int64, bool) {
    if v.Type != TypeInteger {
        return 0, false
    }
    return v.Payload.(int64), true
}

func AsBlock(v Value) (*BlockValue, bool) {
    if v.Type != TypeBlock {
        return nil, false
    }
    return v.Payload.(*BlockValue), true
}
```

### Frame Access Pattern

Always use index-based access for frame references:

```go
// CORRECT: Index-based
type Frame struct {
    Parent int  // index into stack or frame registry
}
func (f *Frame) GetParent(stack *Stack) *Frame {
    if f.Parent == -1 {
        return nil
    }
    return stack.GetFrame(f.Parent)
}

// INCORRECT: Pointer-based (violates constitution Principle IV)
type Frame struct {
    Parent *Frame  // DO NOT USE - invalidates on stack expansion
}
```

### Error Construction

Use factory functions for structured errors:

```go
func NewScriptError(id string, args [3]string, near []Value, where []string) *Error {
    return &Error{
        Category: ErrScript,
        Code:     300 + hashID(id),  // derive code from ID
        ID:       id,
        Args:     args,
        Near:     near,
        Where:    where,
        Message:  formatMessage(id, args),
    }
}

// Usage:
err := NewScriptError("no-value", [3]string{"x", "", ""}, ctx.NearContext(), stack.CaptureCallStack())
```

---

## Testing Strategy

Each entity requires three test categories:

1. **Construction Tests**: Verify valid creation with constructor functions
2. **Operation Tests**: Verify each method/operation produces correct results
3. **Validation Tests**: Verify invalid inputs produce appropriate errors

**Example for Block**:
```go
func TestBlockConstruction(t *testing.T) { /* create valid blocks */ }
func TestBlockFirst(t *testing.T) { /* test First() on various blocks */ }
func TestBlockAppend(t *testing.T) { /* test Append() modifications */ }
func TestBlockEmptyError(t *testing.T) { /* First() on empty block errors */ }
```

Contract tests (test/contract/) will validate native functions using these entities.

---

**Next Artifact**: contracts/ directory with native function contracts
