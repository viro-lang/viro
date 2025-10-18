# Contract: Data Natives

**Category**: Data Operations  
**Functions**: `set`, `get`, `type?`  
**Purpose**: Variable manipulation and type inspection

---

## Native: `set`

**Signature**: `set word value`

**Parameters**:
- `word`: Word (symbol to bind)
- `value`: Any type (value to bind)

**Return**: The value that was set

**Behavior**: 
1. Evaluate second argument to get value
2. Bind word to value in current frame
3. Return the value

**Type Rules**:
- First argument must be Word type (not evaluated word, the word itself)
- Second argument can be any type (evaluated)

**Usage Pattern**:
```viro
set 'x 10        ; binds x to 10, returns 10
; Note: 'x is lit-word (quoted), evaluates to word without lookup

; Equivalent to:
x: 10            ; set-word syntax (more common)
```

**Examples**:
```viro
set 'x 10        → 10 (x is now bound to 10)
set 'y 3 + 4     → 7 (y is now bound to 7)
set 'name "Alice"  → "Alice" (name bound to string)
```

**Error Cases**:
- First argument not word → Script error (300): "Set expects word argument"
- Word is set-word instead of word → typically works (implementation choice), but lit-word syntax preferred

**Test Cases**:
1. `set 'x 10` returns `10`, then `x` evaluates to `10`
2. `set 'y 3 + 4` returns `7`, then `y` evaluates to `7`
3. `set 'name "Alice"` returns `"Alice"`, then `name` evaluates to `"Alice"`
4. `set 42 10` errors (not a word)
5. Multiple sets: `set 'x 10  set 'x 20` → `x` is now `20` (overwrites)

**Note**: In REBOL, `set` is less common than set-word syntax (`x: 10`). The native exists for programmatic word binding (e.g., `set word-from-variable value`).

---

## Native: `get`

**Signature**: `get word`

**Parameters**:
- `word`: Word (symbol to look up)

**Return**: Value bound to word

**Behavior**: 
1. Get word symbol
2. Look up in current frame (and parent frames)
3. Return bound value

**Type Rules**:
- Argument must be Word type (not evaluated)
- Word must be bound (error if unbound)

**Usage Pattern**:
```viro
x: 10
get 'x           ; returns 10
; Equivalent to:
:x               ; get-word syntax (more common)
```

**Examples**:
```viro
x: 10
get 'x           → 10

name: "Alice"
get 'name        → "Alice"

get 'undefined   → Error: no value for word
```

**Error Cases**:
- Argument not word → Script error: "Get expects word argument"
- Word unbound → Script error (300): "No value for word: {symbol}"

**Test Cases**:
1. `x: 10  get 'x` returns `10`
2. `name: "Alice"  get 'name` returns `"Alice"`
3. `get 'undefined-word` errors (no value)
4. `get 42` errors (not word)

**Note**: Like `set`, `get` is less common than get-word syntax (`:x`). The native exists for programmatic word lookup.

---

## Native: `type?`

**Signature**: `type? value`

**Parameters**:
- `value`: Any type (evaluated)

**Return**: Word representing type name

**Behavior**: Returns word identifying value's type

**Type Mapping**:
- Integer → `'integer!`
- String → `'string!`
- Word → `'word!`
- Set-word → `'set-word!`
- Get-word → `'get-word!`
- Lit-word → `'lit-word!`
- Block → `'block!`
- Function → `'function!`
- Logic → `'logic!`
- None → `'none!`

**Examples**:
```viro
type? 42         → integer!
type? "hello"    → string!
type? [1 2 3]    → block!
type? true       → logic!
type? none       → none!
```

**Test Cases**:
1. `type? 42` returns word `integer!`
2. `type? "hello"` returns word `string!`
3. `type? [1 2 3]` returns word `block!`
4. `type? true` returns word `logic!`
5. `type? false` returns word `logic!`
6. `type? none` returns word `none!`
7. `type? :x` returns type of x's value
8. `type? 'x` returns `word!` (lit-word evaluates to word)

**Usage**:
```viro
; Type checking in code
value: 42
if type? value = 'integer! [
    print "It's an integer"
] [
    print "It's something else"
]
```

**Error Cases**: None (accepts any value)

**Implementation Note**:
- Return type is Word (not string)
- Type names follow REBOL convention: lowercase with `!` suffix
- Consistent naming: `integer!`, `string!`, `block!`, `word!`, `logic!`, `none!`, `function!`, `set-word!`, `get-word!`, `lit-word!`

---

## Native: `form`

**Signature**: `form value`

**Parameters**:
- `value`: Any type (evaluated)

**Return**: String representation for display

**Behavior**:
Returns human-readable string format. For blocks, omits outer brackets. For strings, omits quotes. Does not evaluate block contents.

**Type Rules**:
- Argument can be any type (evaluated)

**Usage Pattern**:
```viro
form [1 2 3]     ; returns "1 2 3" (no brackets)
form "hello"     ; returns "hello" (no quotes)
form 42          ; returns "42"
```

**Examples**:
```viro
form [1 2 3]     → "1 2 3"
form "hello"     → "hello"
form 42          → "42"
form true        → "true"
form [a b c]     → "a b c"
```

**Error Cases**: None (accepts any value)

**Test Cases**:
1. `form [1 2 3]` returns string `"1 2 3"`
2. `form "hello"` returns string `"hello"`
3. `form 42` returns string `"42"`
4. `form true` returns string `"true"`
5. `form none` returns string `"none"`
6. `form 'word` returns string `"word"`

---

## Native: `mold`

**Signature**: `mold value`

**Parameters**:
- `value`: Any type (evaluated)

**Return**: String representation for serialization

**Behavior**:
Returns REBOL-readable string format. For blocks, includes outer brackets. For strings, includes quotes. Does not evaluate block contents.

**Type Rules**:
- Argument can be any type (evaluated)

**Usage Pattern**:
```viro
mold [1 2 3]     ; returns "[1 2 3]" (with brackets)
mold "hello"     ; returns "\"hello\"" (with quotes)
mold 42          ; returns "42"
```

**Examples**:
```viro
mold [1 2 3]     → "[1 2 3]"
mold "hello"     → "\"hello\""
mold 42          → "42"
mold true        → "true"
mold [a b c]     → "[a b c]"
```

**Error Cases**: None (accepts any value)

**Test Cases**:
1. `mold [1 2 3]` returns string `"[1 2 3]"`
2. `mold "hello"` returns string `"\"hello\""`
3. `mold 42` returns string `"42"`
4. `mold true` returns string `"true"`
5. `mold none` returns string `"none"`
6. `mold 'word` returns string `"word"`

---

## Native: `reduce`

**Signature**: `reduce block`

**Parameters**:
- `block`: Block (not evaluated)

**Return**: Block containing evaluation results

**Behavior**:
1. Takes a block as input
2. Evaluates each element in the block
3. Returns a new block containing the results of evaluation
4. Preserves the order of elements

**Type Rules**:
- Argument must be Block type (not evaluated)
- Block elements can be any type (will be evaluated)
- Returns Block type

**Usage Pattern**:
```viro
reduce [1 2 3]           ; returns [1 2 3] (literals)
reduce [1 + 2, 3 * 4]    ; returns [3, 12] (expressions)
reduce [x, y + 1]        ; returns [value-of-x, value-of-y-plus-1]
```

**Examples**:
```viro
reduce [1 2 3]           → [1 2 3]
reduce [1 + 2, 3 * 4]    → [3, 12]
reduce []                → []
reduce [true, false]     → [true, false]
reduce ["hello", "world"] → ["hello", "world"]
```

**Error Cases**:
- Argument not block → Script error (300): "Reduce expects block argument"
- Evaluation errors in block elements → propagate the first error encountered

**Test Cases**:
1. `reduce [1 2 3]` returns block `[1 2 3]`
2. `reduce [1 + 2, 3 * 4]` returns block `[3, 12]`
3. `reduce []` returns block `[]`
4. `reduce [true, false]` returns block `[true, false]`
5. `reduce 42` errors (not a block)
6. `reduce [1, undefined-word]` errors (evaluation fails)

**Implementation Note**:
- Each element is evaluated individually in the current evaluation context
- Evaluation errors are propagated immediately (first error stops processing)
- Empty blocks return empty blocks
- Nested blocks are evaluated as single elements

---

## Common Properties

**Word Operations** (`set`, `get`):
- Operate on word symbols, not word values
- Require lit-word syntax (quoted): `'x` not `x`
- Modify or query current frame context
- Frame lookup follows parent chain (lexical scoping)

**Type Inspection** (`type?`):
- Universal operation (works on any value)
- Returns word value (not string)
- Type names are REBOL datatype words (with `!` suffix)

**Frame Context**:
- `set` and `get` operate in current evaluation frame
- Frame hierarchy: function frame → parent frame → ... → global frame
- Word lookup searches frame chain until found or error

**Error Messages**:
- Include operation name and expected types
- For `get` with unbound word: include word symbol in message

---

## Implementation Checklist

For each native:
- [ ] Function signature matches contract
- [ ] Type validation (set/get expect word)
- [ ] Frame context operations (set creates or updates binding, get searches frame chain)
- [ ] Return correct value
- [ ] All test cases pass
- [ ] Error messages clear and include context

**Dependencies**:
- Value system (Word type, all value types for type?)
- Frame system (bind, get, set operations)
- Error system (Script error for undefined words)
- Type system (type constants and names)

**Testing Strategy**:
- Table-driven tests for type? (one test per value type)
- Set/get round-trip tests (set then get same word)
- Error cases (unbound word, wrong argument types)
- Frame context tests (local vs parent frame lookup)

**Advanced Scenarios** (to test in integration):
```viro
; Set/get with function arguments
square: fn [n] [
    set 'temp n * n
    get 'temp
]
square 5  → 25

; Type checking in control flow
check-type: fn [val] [
    if type? val = 'integer! [
        print "integer"
    ] [
        print "not integer"
    ]
]
```

**Future Extensions** (out of Phase 1 scope):
- `set` with block of words (parallel assignment)
- `get` with path (object field access)
- Additional type? variants (integer?, string?, block? predicates)
- `to` type conversion natives
