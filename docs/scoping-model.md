# Viro Scoping Model

**Understanding Viro's Local-By-Default Variable Scoping**

---

## Core Principle: Local-By-Default

Viro uses **local-by-default** scoping. Variables declared or assigned within a function are automatically local to that function.

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

## Why Local-By-Default?

### 1. Predictability

Functions don't accidentally modify external state:

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

Functions don't leak implementation details:

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

### 3. Industry Standard

Most modern languages use local-by-default:
- Python: Local unless `global` keyword
- JavaScript: Local with `let`/`const`
- Go: Local within function
- Rust: Local by default

---

## Accessing Global Variables

When you need to access or modify global state, Viro provides explicit mechanisms.

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

### Option 2: Use `get` and `set` Natives

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

## Scoping Rules Summary

| Aspect               | Behavior                      |
|----------------------|-------------------------------|
| Default scope        | Local                         |
| Function parameters  | Local                         |
| Variables in body    | Local (automatic shadowing)   |
| Access global        | `get`/`set` natives          |
| Closure support      | Yes (captures lexical scope)  |
| Shadowing            | Automatic                     |
| Encapsulation        | Enforced by default           |

---

## Examples

### Example 1: Function with Local Variables

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

### Example 3: Explicit Global Modification

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

## Comparison with Global-By-Default Languages

Some languages default to global scope unless you explicitly declare variables as local. This can lead to surprising behavior:

**Global-by-default (conceptual example)**:
```
x: 10
test: function [] [
    x: 20    # Modifies GLOBAL x
]
test
print x      # Shows 20 (global was modified)
```

**Viro's local-by-default**:
```
>> x: 10
10
>> test: fn [] [x: 20]  # Creates LOCAL x
function[test]
>> test
20
>> x
10                       # Global unchanged
```

---

## Best Practices

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

## Design Benefits

### Advantages

1. **Fewer Bugs**: No accidental global modifications
2. **Better Testing**: Pure functions are easier to test
3. **Clearer Intent**: Global access is explicit via `get`/`set`
4. **Modern Conventions**: Matches expectations from contemporary languages
5. **Safer Refactoring**: Can change function internals without breaking callers

### Tradeoffs

1. **More Verbose**: Need `get`/`set` for global access when truly necessary
2. **Explicit Global Access**: Requires intention when working with shared state

**Philosophy**: Safety and predictability over convenience.

---

## Summary

**Key Takeaway**: In Viro, functions create isolated scopes by default. Use `get`/`set` for explicit global access when needed.

| Feature                  | Viro Behavior   |
|--------------------------|-----------------|
| Variable scope default   | Local           |
| Explicit local           | Not needed      |
| Explicit global          | `get`/`set`     |
| Shadowing behavior       | Automatic       |
| Closure support          | Yes             |
| Design philosophy        | Safety first    |

---

For more information:
- REPL Usage: [`docs/repl-usage.md`](./repl-usage.md)
- Operator Evaluation: [`docs/operator-precedence.md`](./operator-precedence.md)
- Quickstart: [`specs/001-implement-the-core/quickstart.md`](../specs/001-implement-the-core/quickstart.md)
