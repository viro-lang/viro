# Operator Precedence Reference

**Viro Interpreter** - Complete operator precedence guide

---

## Precedence Levels

Operators are evaluated in order from highest to lowest precedence. Operators at the same level are evaluated left-to-right.

### Level 1: Function Calls (Highest)

**Syntax**: `function-name arg1 arg2 ...`

**Examples**:
```
square 5        → (square 5) → 25
add 3 4         → (add 3 4) → 7
max 10 20       → (max 10 20) → 20
```

**Binding**: Function name followed by expected number of arguments

---

### Level 2: Unary Negation

**Operator**: `-` (prefix)

**Examples**:
```
-5              → -5
-x              → (- x)
-(3 + 4)        → -(7) → -7
```

**Note**: Negation binds tighter than binary operators but looser than function calls

---

### Level 3: Multiplication and Division

**Operators**: `*`, `/`

**Associativity**: Left-to-right

**Examples**:
```
6 * 7           → 42
20 / 4          → 5
6 * 7 / 2       → (/ (* 6 7) 2) → 21
10 / 2 * 3      → (* (/ 10 2) 3) → 15
```

**With parens**:
```
10 / (2 * 3)    → (/ 10 (* 2 3)) → 1
```

---

### Level 4: Addition and Subtraction

**Operators**: `+`, `-`

**Associativity**: Left-to-right

**Examples**:
```
3 + 4           → 7
10 - 3          → 7
3 + 4 - 2       → (- (+ 3 4) 2) → 5
```

**With higher precedence**:
```
3 + 4 * 2       → (+ 3 (* 4 2)) → 11
10 - 6 / 2      → (- 10 (/ 6 2)) → 7
```

---

### Level 5: Comparison Operators

**Operators**: `<`, `>`, `<=`, `>=`

**Associativity**: Left-to-right (though chaining not recommended)

**Examples**:
```
5 > 3           → true
10 <= 10        → true
x < y           → (< x y)
```

**With arithmetic**:
```
3 + 4 < 10      → (< (+ 3 4) 10) → true
x * 2 >= 10     → (>= (* x 2) 10)
```

---

### Level 6: Equality Operators

**Operators**: `=`, `<>`

**Associativity**: Left-to-right

**Examples**:
```
5 = 5           → true
5 <> 3          → true
x = y           → (= x y)
```

**With comparisons**:
```
3 < 5 = true    → (= (< 3 5) true) → true
```

---

### Level 7: Logical Operators (Lowest)

**Operators**: `and`, `or`

**Associativity**: Left-to-right

**Examples**:
```
true and false  → false
true or false   → true
x and y         → (and x y)
```

**With other operators**:
```
5 > 3 and 10 < 20       → (and (> 5 3) (< 10 20)) → true
x = 0 or y = 0          → (or (= x 0) (= y 0))
```

**Note**: `not` is a function, not an operator, so it has function call precedence (Level 1)

---

## Complete Precedence Table

| Level | Operators         | Associativity | Example              | Result        |
|-------|-------------------|---------------|----------------------|---------------|
| 1     | Function calls    | Left-to-right | `square 5`           | `25`          |
| 2     | `-` (unary)       | Right-to-left | `-5`                 | `-5`          |
| 3     | `*`, `/`          | Left-to-right | `6 * 7`              | `42`          |
| 4     | `+`, `-` (binary) | Left-to-right | `3 + 4`              | `7`           |
| 5     | `<`, `>`, `<=`, `>=` | Left-to-right | `5 > 3`           | `true`        |
| 6     | `=`, `<>`         | Left-to-right | `5 = 5`              | `true`        |
| 7     | `and`, `or`       | Left-to-right | `true and false`     | `false`       |

---

## Parentheses Override

Parentheses `(...)` have highest priority and force evaluation order:

```
3 + 4 * 2       → 11            # * before +
(3 + 4) * 2     → 14            # Parens first

10 / 2 * 3      → 15            # Left-to-right
10 / (2 * 3)    → 1             # Parens first

x < 10 and y > 0                # and is lowest
(x < 10) and (y > 0)            # Same (explicit grouping)
```

---

## Examples by Complexity

### Simple Arithmetic

```
3 + 4           → (+ 3 4) → 7
5 * 6           → (* 5 6) → 30
10 - 2          → (- 10 2) → 8
20 / 4          → (/ 20 4) → 5
```

### Mixed Operators

```
3 + 4 * 2       → (+ 3 (* 4 2)) → 11
10 - 6 / 2      → (- 10 (/ 6 2)) → 7
2 * 3 + 4 * 5   → (+ (* 2 3) (* 4 5)) → 26
```

### With Comparisons

```
5 > 3           → (> 5 3) → true
3 + 4 > 5       → (> (+ 3 4) 5) → true
x * 2 < 10      → (< (* x 2) 10)
```

### With Logic

```
5 > 3 and 10 < 20       → (and (> 5 3) (< 10 20)) → true
x = 0 or y = 0          → (or (= x 0) (= y 0))
not (x > 10) and y < 5  → (and (not (> x 10)) (< y 5))
```

### Complex Expressions

```
1 + 2 * 3 - 4 + 5 / 2
→ (+ (- (+ 1 (* 2 3)) 4) (/ 5 2))
→ (+ (- (+ 1 6) 4) 2)
→ (+ (- 7 4) 2)
→ (+ 3 2)
→ 5
```

```
(3 + 4) * (5 - 2)
→ (* (+ 3 4) (- 5 2))
→ (* 7 3)
→ 21
```

---

## Comparison with Other Languages

### C/C++/Java/JavaScript

Viro's precedence mostly matches these languages:

| Viro Level | C/Java Equivalent    |
|------------|----------------------|
| 1          | Function calls `()`  |
| 2          | Unary `-`            |
| 3          | `*`, `/`             |
| 4          | `+`, `-`             |
| 5          | `<`, `>`, `<=`, `>=` |
| 6          | `==`, `!=`           |
| 7          | `&&`, `||`           |

**Difference**: Viro uses `=` for equality (not `==`) and `<>` for inequality (not `!=`).

### Python

Similar precedence, with these symbol differences:
- Viro: `and`, `or` (words) = Python: `and`, `or` (words) ✓
- Viro: `=` (equality) ≠ Python: `==` (equality)
- Viro: `<>` (not equal) ≠ Python: `!=` (not equal)

### REBOL

Viro **differs significantly** from REBOL:

**REBOL**: Uses strict left-to-right evaluation with *no operator precedence*
```rebol
3 + 4 * 2  ; In REBOL → ((3 + 4) * 2) → 14
```

**Viro**: Uses mathematical precedence (like most languages)
```
3 + 4 * 2  ; In Viro → (3 + (4 * 2)) → 11
```

---

## Function Calls in Expressions

Functions have highest precedence, so they consume arguments before operators:

```
square: fn [n] [(* n n)]

square 3 + 4
→ (+ (square 3) 4)
→ (+ 9 4)
→ 13

square (3 + 4)
→ (square (+ 3 4))
→ (square 7)
→ 49
```

**Multiple function calls**:
```
add: fn [a b] [(+ a b)]

add 3 4 * 2
→ (* (add 3 4) 2)
→ (* 7 2)
→ 14

add (3 * 2) (4 * 2)
→ (add (* 3 2) (* 4 2))
→ (add 6 8)
→ 14
```

---

## Common Mistakes

### Mistake 1: Assuming REBOL Evaluation

**Wrong assumption** (REBOL style):
```
3 + 4 * 2  ; Expecting 14
```

**Actual result**:
```
3 + 4 * 2  → 11  # Mathematical precedence
```

**Fix**: Use parens for REBOL-style evaluation:
```
(3 + 4) * 2  → 14
```

### Mistake 2: Function Call Scope

**Wrong**:
```
square 3 + 4  ; Expecting square(3+4) = 49
→ 13          # Actually: (square 3) + 4
```

**Right**:
```
square (3 + 4)  → 49
```

### Mistake 3: Comparison Chains

**Unclear**:
```
x < y < z  ; Valid but unclear
```

**Better**:
```
(x < y) and (y < z)  ; Explicit intent
```

### Mistake 4: Operator vs Function

`not` is a **function**, not an operator:

```
not true and false
→ (and (not true) false)
→ false

# not (true and false)  # Would need explicit parens
```

---

## Best Practices

### 1. Use Parens for Clarity

Even when not required, parens improve readability:

```
# Works but less clear
x * 2 + y * 3

# Better
(x * 2) + (y * 3)
```

### 2. Break Complex Expressions

```
# Hard to read
result: x * 2 + y * 3 - z / 4

# Easier
a: (x * 2)
b: (y * 3)
c: (z / 4)
result: (a + b - c)
```

### 3. Function Calls Need Parens in Expressions

```
# Always wrap function calls in expressions
total: (add x y) + (multiply a b)
```

### 4. Test Incrementally

```
# Build up complex expressions piece by piece
>> x: 5
>> y: 10
>> x * 2
10
>> y * 3
30
>> (x * 2) + (y * 3)
40
```

---

## Quick Reference Card

```
Highest Precedence (evaluates first)
  ↓
  function-name args
  -x (unary negation)
  * /
  + -
  < > <= >=
  = <>
  and or
  ↑
Lowest Precedence (evaluates last)

Parens (...) override everything
```

---

## Testing Precedence

Use the REPL to verify precedence:

```
>> 3 + 4 * 2
11

>> (3 + 4) * 2
14

>> 5 > 3 and 10 < 20
true

>> square: fn [n] [(* n n)]
>> square 3 + 4
13

>> square (3 + 4)
49
```

---

For more examples, see [`docs/repl-usage.md`](./repl-usage.md).
