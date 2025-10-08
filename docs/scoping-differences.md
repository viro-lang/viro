# Scoping Differences from REBOL

**Viro vs REBOL** - Understanding variable scoping behavior

---

## Key Difference: Local-By-Default

### REBOL: Global-By-Default

In REBOL, variables are **global by default** unless explicitly made local:

```rebol
x: 10

test: func [] [
    x: 20    ; Modifies GLOBAL x
]

test
print x      ; Shows 20 (global was modified)
```

To make variables local in REBOL, you must declare them:

```rebol
x: 10

test: func [/local x] [  ; Explicit /local refinement
    x: 20                ; Now LOCAL to function
]

test
print x      ; Shows 10 (global unchanged)
```

### Viro: Local-By-Default

In Viro, variables are **local by default**:

```
>> x: 10
10
>> test: fn [] [x: 20]
function[test]
>> test
20
>> x
10
```

The function creates a **new local variable** `x` that shadows the global one.

---

## Rationale for Local-By-Default

### 1. Predictability

**Problem with global-by-default**: Surprising action-at-a-distance

```rebol
; REBOL: Function unexpectedly modifies global state
x: 10
helper: func [] [x: 20]  ; Oops! Changes global x
process: func [] [
    ; ... 100 lines of code ...
    helper
    ; ... x is now 20, not 10! ...
]
```

**Solution in Viro**: Local by default prevents accidents

```
x: 10
helper: fn [] [x: 20]    # Creates local x
process: fn [] [
    # ... 100 lines of code ...
    helper
    # ... global x still 10 ...
]
```

### 2. Encapsulation

Functions should not leak implementation details:

```
>> counter: 0
0
>> increment: fn [] [
..   temp: counter       # Local variable
..   temp: (+ temp 1)
..   temp                # Return value
.. ]
function[increment]
>> increment
1
>> temp                  # Error: temp doesn't exist outside function
** Script Error: Word 'temp' is not defined
```

### 3. Familiarity

Most modern languages use local-by-default:
- Python: Local unless `global` keyword
- JavaScript: Local unless `var` at top level
- Go: Local within function
- Rust: Local by default

---

## How to Access Global Variables

### Option 1: Pass as Arguments

```
>> global-x: 10
10
>> modify-x: fn [x] [(+ x 5)]
function[modify-x]
>> new-value: (modify-x global-x)
15
>> global-x
10
```

### Option 2: Use `set` and `get` Natives

**`get`** - Fetch global variable:

```
>> global-x: 10
10
>> reader: fn [] [(get 'global-x)]
function[reader]
>> reader
10
```

**`set`** - Modify global variable:

```
>> global-x: 10
10
>> writer: fn [] [(set 'global-x 20)]
function[writer]
>> writer
20
>> global-x
20
```

**Note**: Use quoted word `'global-x` (lit-word) to pass the word itself, not its value.

### Option 3: Return New Value

```
>> counter: 0
0
>> increment: fn [] [(+ (get 'counter) 1)]
function[increment]
>> counter: (increment)
1
>> counter: (increment)
2
```

---

## Lexical Scoping (Closures)

Functions capture their **lexical environment**:

```
>> make-counter: fn [] [
..   count: 0
..   fn [] [
..     count: (+ count 1)
..     count
..   ]
.. ]
function[make-counter]
>> counter1: (make-counter)
function[unnamed]
>> counter2: (make-counter)
function[unnamed]
>> counter1
1
>> counter1
2
>> counter2
1
>> counter1
3
```

Each invocation of `make-counter` creates a **new closure** with its own `count`.

---

## Comparison Table

| Aspect               | REBOL                        | Viro                          |
|----------------------|------------------------------|-------------------------------|
| Default scope        | Global                       | Local                         |
| Make local           | `/local` refinement          | Automatic                     |
| Access global        | Direct access                | `get`/`set` natives          |
| Shadowing            | Requires `/local`            | Automatic                     |
| Closure support      | Yes                          | Yes                           |
| Predictability       | Low (global modifications)   | High (isolated functions)     |
| Encapsulation        | Requires discipline          | Enforced by default           |

---

## Examples

### Example 1: Function with Local Variables

**REBOL**:
```rebol
x: 10
y: 20

add-and-square: func [a b /local temp] [
    temp: a + b
    temp * temp
]

result: add-and-square x y
; temp is undefined outside function
```

**Viro**:
```
>> x: 10
10
>> y: 20
20
>> add-and-square: fn [a b] [
..   temp: (+ a b)
..   (* temp temp)
.. ]
function[add-and-square]
>> result: (add-and-square x y)
900
>> temp
** Script Error: Word 'temp' is not defined
```

### Example 2: Counter with Closure

**REBOL**:
```rebol
make-counter: func [/local count] [
    count: 0
    func [] [
        count: count + 1
        count
    ]
]

c1: make-counter
c1  ; 1
c1  ; 2
```

**Viro**:
```
>> make-counter: fn [] [
..   count: 0
..   fn [] [
..     count: (+ count 1)
..     count
..   ]
.. ]
function[make-counter]
>> c1: (make-counter)
function[unnamed]
>> c1
1
>> c1
2
```

### Example 3: Modifying Global State

**REBOL** (easy but dangerous):
```rebol
total: 0

add-to-total: func [n] [
    total: total + n  ; Modifies global
]

add-to-total 5
print total  ; 5
```

**Viro** (explicit):
```
>> total: 0
0
>> add-to-total: fn [n] [
..   new-total: (+ (get 'total) n)
..   set 'total new-total
..   new-total
.. ]
function[add-to-total]
>> add-to-total 5
5
>> total
5
```

Or better, avoid global mutation:

```
>> total: 0
0
>> add-to-total: fn [current n] [(+ current n)]
function[add-to-total]
>> total: (add-to-total total 5)
5
>> total: (add-to-total total 10)
15
```

---

## Migration Guide: REBOL → Viro

### Pattern 1: Simple Function

**REBOL**:
```rebol
square: func [n] [n * n]
```

**Viro**:
```
square: fn [n] [(* n n)]
```

*No change needed - simple functions work the same.*

### Pattern 2: Function with Local Variables

**REBOL**:
```rebol
complex-calc: func [a b /local temp1 temp2] [
    temp1: a * a
    temp2: b * b
    temp1 + temp2
]
```

**Viro**:
```
complex-calc: fn [a b] [
    temp1: (* a a)
    temp2: (* b b)
    (+ temp1 temp2)
]
```

*Remove `/local` - variables are automatically local.*

### Pattern 3: Modifying Global

**REBOL**:
```rebol
counter: 0

increment: func [] [
    counter: counter + 1
]
```

**Viro** (Option A - explicit global access):
```
counter: 0

increment: fn [] [
    new: (+ (get 'counter) 1)
    set 'counter new
    new
]
```

**Viro** (Option B - return new value):
```
counter: 0

increment: fn [n] [(+ n 1)]

# Usage
counter: (increment counter)
```

### Pattern 4: Closure

**REBOL**:
```rebol
make-account: func [initial /local balance] [
    balance: initial
    func [amount] [
        balance: balance + amount
        balance
    ]
]
```

**Viro**:
```
make-account: fn [initial] [
    balance: initial
    fn [amount] [
        balance: (+ balance amount)
        balance
    ]
]
```

*Closures work similarly - just remove `/local`.*

---

## Best Practices for Viro

### 1. Prefer Pure Functions

Functions that don't modify external state are easier to reason about:

```
>> add: fn [a b] [(+ a b)]
function[add]
>> result: (add 3 4)
7
```

### 2. Pass State as Arguments

Instead of reading globals, pass values in:

```
>> process: fn [data config] [
..   # Use data and config
..   # Return new value
.. ]
```

### 3. Return New Values

Instead of modifying state, return updated values:

```
>> update-counter: fn [count] [(+ count 1)]
function[update-counter]
>> counter: 0
0
>> counter: (update-counter counter)
1
```

### 4. Use Closures for Encapsulation

When you need mutable state, encapsulate it in a closure:

```
>> make-stack: fn [] [
..   items: []
..   fn [op val] [
..     # Push/pop operations on items
..   ]
.. ]
```

### 5. Document Global Access

If a function uses `get`/`set`, document it clearly:

```
>> # WARNING: Modifies global 'config' variable
>> update-config: fn [key val] [
..   # ... set global config ...
.. ]
```

---

## Why This Design Choice?

### Benefits of Local-By-Default

1. **Fewer Bugs**: No accidental global modifications
2. **Better Testing**: Pure functions are easier to test
3. **Clearer Intent**: Global access is explicit via `get`/`set`
4. **Modern Conventions**: Matches expectations from other languages
5. **Safer Refactoring**: Can change function internals without breaking callers

### Tradeoffs

1. **More Verbose**: Need `get`/`set` for global access
2. **Different from REBOL**: Existing REBOL code needs adaptation
3. **Learning Curve**: REBOL users must adjust mental model

**Philosophy**: Safety and predictability over convenience.

---

## Summary

| Feature                  | REBOL         | Viro            |
|--------------------------|---------------|-----------------|
| Variable scope default   | Global        | Local           |
| Explicit local           | `/local`      | Not needed      |
| Explicit global          | Not needed    | `get`/`set`     |
| Shadowing behavior       | Opt-in        | Automatic       |
| Closure support          | Yes           | Yes             |
| Design philosophy        | Convenience   | Safety          |

**Key Takeaway**: In Viro, functions create isolated scopes by default. Use `get`/`set` for explicit global access when needed.

---

For more information:
- REPL Usage: [`docs/repl-usage.md`](./repl-usage.md)
- Architecture: [`docs/interpreter.md`](./interpreter.md)
- Quickstart: [`specs/001-implement-the-core/quickstart.md`](../specs/001-implement-the-core/quickstart.md)
