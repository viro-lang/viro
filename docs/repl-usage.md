# REPL Usage Guide

**Viro Interpreter** - Interactive homoiconic programming

---

## Starting the REPL

```bash
./viro
```

**Output**:
```
Viro 0.1.0
Type 'exit' or 'quit' to leave
>> 
```

---

## Basic Usage

### Literals

**Integers**:
```
>> 42
42
>> -17
-17
```

**Strings**:
```
>> "hello world"
"hello world"
>> "embedded \"quotes\" work"
"embedded \"quotes\" work"
```

**Logic Values**:
```
>> true
true
>> false
false
```

**None** (represents absence of value):
```
>> none

```
*Note: `none` displays as blank (suppressed in REPL)*

---

## Variables

### Assignment (Set-Word)

Use `:` after a word to assign:

```
>> x: 10
10
>> name: "Alice"
"Alice"
>> flag: true
true
```

### Lookup (Word)

Use a word without `:` to get its value:

```
>> x
10
>> name
"Alice"
```

### Undefined Words

```
>> undefined-var
** Script Error: Word 'undefined-var' is not defined
```

---

## Arithmetic

### Basic Operations

```
>> 3 + 4
7
>> 10 - 3
7
>> 6 * 7
42
>> 20 / 4
5
```

### Left-to-Right Evaluation

All operators are evaluated from left to right (no precedence):

```
>> 3 + 4 * 2
14
>> 3 + (4 * 2)
11
```

### Complex Expressions

```
>> 10 - 6 / 2
2
>> 10 - (6 / 2)
7
```

*Evaluation*: Operations are processed left-to-right. Use parentheses to control order.

---

## Comparisons

```
>> 5 > 3
true
>> 3 < 5
true
>> 5 = 5
true
>> 5 <> 3
true
>> 10 >= 10
true
>> 5 <= 4
false
```

---

## Logic Operations

### AND, OR, NOT

```
>> and true false
false
>> or true false
true
>> not false
true
>> not true
false
```

### Truthy Values

- `false` and `none` are falsy
- Everything else is truthy

```
>> and 42 true
true
>> or none false
false
```

---

## Blocks and Parens

### Blocks `[...]` - Deferred Evaluation

Blocks return themselves without evaluating contents:

```
>> [1 + 2]
[(+ 1 2)]
>> [x: 10  x * 2]
[(: x 10) (* x 2)]
```

*Use blocks for*: Code storage, data structures, function bodies

### Parens `(...)` - Immediate Evaluation

Parens evaluate their contents immediately:

```
>> (1 + 2)
3
>> x: 5  (x * 2)
10
```

---

## Series Operations

### Creating Series

```
>> data: [1 2 3 4 5]
[1 2 3 4 5]
>> text: "hello"
"hello"
```

### Accessing Elements

```
>> first data
1
>> last data
5
>> first text
"h"
```

### Modifying Series

```
>> append data 6
[1 2 3 4 5 6]
>> data
[1 2 3 4 5 6]

>> insert data 0
[0 1 2 3 4 5 6]
```

*Note: Series operations modify in-place*

### Length Query

```
>> length? data
7
>> length? "hello"
5
```

---

## Control Flow

### When - Single Branch

Execute block only if condition is true:

```
>> when true [42]
42
>> when false [42]
none
```

### If - Two Branches

```
>> if true [1] [2]
1
>> if false [1] [2]
2
```

### Loop - Count-Based

```
>> counter: 0
0
>> loop 5 [counter: (+ counter 1)]
none
>> counter
5
```

### While - Condition-Based

```
>> n: 0
0
>> while (< n 3) [n: (+ n 1)]
none
>> n
3
```

---

## Functions

### Defining Functions

Use `fn` with parameter list and body:

```
>> square: fn [n] [(* n n)]
function[square]
```

### Calling Functions

```
>> square 5
25
>> square 10
100
```

### Multiple Parameters

```
>> add: fn [a b] [(+ a b)]
function[add]
>> add 3 7
10
```

### No Parameters

```
>> greet: fn [] ["Hello!"]
function[greet]
>> greet
"Hello!"
```

### Nested Functions

```
>> outer: fn [x] [
..   inner: fn [y] [(* x y)]
..   inner 10
.. ]
function[outer]
>> outer 5
50
```

*Note: Inner functions have access to outer parameters (lexical scoping)*

---

## Type Queries

```
>> type? 42
integer!
>> type? "hello"
string!
>> type? true
logic!
>> type? [1 2 3]
block!
>> type? square
function!
```

---

## Multi-Line Input

The REPL automatically detects incomplete input:

```
>> square: fn [n] [
..   (* n n)
.. ]
function[square]
```

*Prompt changes to `..` for continuation lines*

### Supported Multi-Line Constructs

- Function definitions
- Nested blocks
- Control flow
- Complex expressions

---

## Command History

### Navigation

- **Up Arrow** `↑`: Previous command
- **Down Arrow** `↓`: Next command
- **Ctrl+R**: Search history

### Persistent History

History is saved to `~/.viro_history` and restored between sessions.

### Editing

- **Left/Right Arrows**: Move cursor
- **Home/End**: Start/end of line
- **Ctrl+A/E**: Start/end of line (macOS/Linux)
- **Backspace/Delete**: Remove characters

---

## Error Handling

### Error Display

```
>> 10 / 0
** Math Error: Division by zero
Where: (/ 10 0)
```

### Recovery

After an error, the REPL continues:

```
>> 10 / 0
** Math Error: Division by zero
Where: (/ 10 0)
>> 10 / 2
5
```

### Interrupting Evaluation

Press **Ctrl+C** during evaluation to interrupt:

```
>> loop 999999999 [42]
^C
** Evaluation interrupted
>> 
```

---

## Special Commands

### Exiting the REPL

```
>> quit
Goodbye!
```

Or:

```
>> exit
Goodbye!
```

Or press **Ctrl+D** (macOS/Linux).

---

## I/O Operations

### Print

Display value and return none:

```
>> print 42
42

>> print "Hello, world!"
Hello, world!

```

### Prin

Display value without trailing newline and return none:

```
>> prin "Hello"
Hello>>
>> prin 42
42>>
>> prin [1 2 3]
1 2 3>>
```

*Note: Since `prin` does not output a newline, the REPL prompt `>>` appears on the same line as the output.*

### Input

Read line from stdin:

```
>> name: input
Alice
>> name
"Alice"
```

---

## Example Sessions

### Fibonacci Calculator

```
>> fib: fn [n] [
..   if (<= n 1) [
..     n
..   ] [
..     (+ (fib (- n 1)) (fib (- n 2)))
..   ]
.. ]
function[fib]
>> fib 10
55
```

### List Processing

```
>> numbers: [1 2 3 4 5]
[1 2 3 4 5]
>> sum: 0
0
>> loop (length? numbers) [
..   sum: (+ sum (first numbers))
..   numbers: [...]  # Not implemented: series advancement
.. ]
```

### Factorial

```
>> fact: fn [n] [
..   if (= n 0) [
..     1
..   ] [
..     (* n (fact (- n 1)))
..   ]
.. ]
function[fact]
>> fact 5
120
```

### Data Aggregation

```
>> total: 0
0
>> count: 0
0
>> data: [10 20 30 40 50]
[10 20 30 40 50]
>> loop 5 [
..   total: (+ total (first data))
..   data: ...  # Series iteration not yet implemented
.. ]
```

---

## Tips and Tricks

### 1. Use Parens for Complex Expressions

When nesting function calls, use parens for clarity:

```
>> square: fn [n] [(* n n)]
function[square]
>> (+ (square 3) (square 4))
25
```

### 2. Break Long Expressions

Use multi-line for readability:

```
>> result: if (> x 10) [
..   (* x 2)
.. ] [
..   (+ x 10)
.. ]
```

### 3. Check Types When Debugging

```
>> mystery-value: ...
>> type? mystery-value
integer!
```

### 4. Use Variables for Intermediate Results

```
>> base: 5
5
>> height: 10
10
>> area: (* base height)
50
>> volume: (* area 3)
150
```

### 5. Test Small Pieces First

Before building complex functions, test components:

```
>> (* 3 4)
12
>> fn [n] [(* n n)]
function[unnamed]
>> square: fn [n] [(* n n)]
function[square]
```

---

## Common Pitfalls

### 1. Forgetting Parens for Function Calls in Expressions

**Wrong**:
```
>> + square 3 square 4
** Error: ...
```

**Right**:
```
>> (+ (square 3) (square 4))
25
```

### 2. Using Block Instead of Paren

**Wrong** (block doesn't evaluate):
```
>> x: [5 + 3]
[(+ 5 3)]
```

**Right** (paren evaluates):
```
>> x: (5 + 3)
8
```

### 3. Modifying Series Assumptions

Series operations modify in-place:

```
>> data: [1 2 3]
[1 2 3]
>> append data 4
[1 2 3 4]
>> data
[1 2 3 4]  # Original modified!
```

### 4. Scope Confusion

Variables are local to functions:

```
>> x: 10
10
>> test: fn [] [x: 20]
function[test]
>> test
20
>> x
10  # Outer x unchanged
```

---

## Limitations (v1.0)

### Not Yet Implemented

- **Series iteration**: No `foreach` or series position advancement
- **Objects**: No object/context system
- **Parse dialect**: No pattern matching DSL
- **File I/O**: No `read`/`write` operations
- **Refinements**: Limited function customization
- **Error throw/catch**: No user exception handling
- **Module system**: No `import`/`export`

### Language Characteristics

- **Scoping**: Local-by-default for safe, predictable behavior
- **Native count**: 29 core functions
- **Series model**: Simplified value-based series
- **Datatypes**: 10 core types

---

## Getting Help

### In REPL

```
>> type? 42          # Check type
integer!
```

### Documentation

- Architecture: `docs/interpreter.md`
- Quickstart: `specs/001-implement-the-core/quickstart.md`
- Contracts: `specs/001-implement-the-core/contracts/`

### Evaluation Reference

See "Operator Evaluation Reference" section in `docs/operator-precedence.md` for details on left-to-right evaluation.

---

Enjoy exploring Viro! 🚀
