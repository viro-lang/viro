# Operator Evaluation Reference

**Viro Interpreter** - Left-to-right evaluation guide

---

## Left-to-Right Evaluation

Viro uses **left-to-right evaluation** with **no operator precedence**. This means operators are evaluated in the order they appear, from left to right.

### Basic Principle

```
3 + 4 * 2       → ((3 + 4) * 2) → (7 * 2) → 14
```

**NOT**:
```
3 + 4 * 2       → (3 + (4 * 2)) → (3 + 8) → 11  # This is mathematical precedence
```

---

## Examples

### Simple Arithmetic

All operators are evaluated left-to-right as they appear:

```
3 + 4           → (+ 3 4) → 7
5 * 6           → (* 5 6) → 30
10 - 2          → (- 10 2) → 8
20 / 4          → (/ 20 4) → 5
```

### Multiple Operators

```
3 + 4 * 2       → (+ 3 4) = 7, then (* 7 2) = 14
10 - 6 / 2      → (- 10 6) = 4, then (/ 4 2) = 2
2 * 3 + 4 * 5   → (* 2 3) = 6, then (+ 6 4) = 10, then (* 10 5) = 50
```

### With Comparisons

```
5 > 3           → (> 5 3) → true
3 + 4 > 5       → (+ 3 4) = 7, then (> 7 5) → true
x * 2 < 10      → (* x 2), then (< result 10)
```

### With Logic

```
5 > 3 and 10 < 20       → (> 5 3) = true, then (and true 10) (evaluates second), then (< result 20)
x = 0 or y = 0          → (= x 0), then (or result y), then (= result 0)
```

---

## Parentheses Control Order

Use parentheses `(...)` to force specific evaluation order:

```
3 + 4 * 2       → 14            # Left-to-right: (3 + 4) * 2
3 + (4 * 2)     → 11            # Parens first: 3 + 8

10 - 6 / 2      → 2             # Left-to-right: (10 - 6) / 2
10 - (6 / 2)    → 7             # Parens first: 10 - 3

2 + 3 * 4       → 20            # Left-to-right: (2 + 3) * 4
2 + (3 * 4)     → 14            # Parens first: 2 + 12
```

---

## Comparison with Other Languages

### Traditional Languages (C, Java, JavaScript, Python)

Most languages use **mathematical operator precedence**:

```javascript
// JavaScript
3 + 4 * 2       // → 11 (multiplication first)
```

**Viro is different**: Uses left-to-right evaluation

```
3 + 4 * 2       ; In Viro → 14 (left-to-right)
```

Viro uses strict left-to-right evaluation with no operator precedence.

---

## Function Calls in Expressions

Functions consume their arguments before operators continue:

```
square: fn [n] [(* n n)]

square 3 + 4
→ (square 3) + 4
→ 9 + 4
→ 13

(square 3) + 4
→ 9 + 4
→ 13

square (3 + 4)
→ square 7
→ 49
```

---

## Common Patterns

### Pattern 1: Mathematical Expressions

To get traditional mathematical precedence behavior, use parentheses:

```
# Viro left-to-right:
3 + 4 * 2       → 14

# To get mathematical precedence (11):
3 + (4 * 2)     → 11
```

### Pattern 2: Chained Operations

```
# All left-to-right
a: 10 + 20 * 2 - 5
; → ((10 + 20) * 2) - 5
; → (30 * 2) - 5
; → 60 - 5
; → 55

# With explicit grouping for clarity
a: ((10 + 20) * 2) - 5
```

### Pattern 3: Boolean Logic

```
# Left-to-right evaluation
x > 10 and y < 20 or z = 30
; → ((x > 10) and y) < 20 or z = 30  # Can be confusing!

# Better: use explicit parentheses
(x > 10) and (y < 20) or (z = 30)
; or
((x > 10) and (y < 20)) or (z = 30)
```

---

## Best Practices

### 1. Use Parentheses Liberally

Even when not required, parentheses improve readability and prevent confusion:

```
# Can be confusing
x * 2 + y * 3

# Much clearer
(x * 2) + (y * 3)
```

### 2. Break Complex Expressions

```
# Hard to understand
result: x * 2 + y * 3 - z / 4

# Much clearer
a: (x * 2)
b: (y * 3)
c: (z / 4)
result: (a + b) - c
```

### 3. Know the Evaluation Model

Remember: Viro evaluates **left-to-right**, not by mathematical precedence.

```
# This may surprise you if you expect math precedence:
10 + 5 * 2      → 30, not 20

# Use parens for math precedence:
10 + (5 * 2)    → 20
```

### 4. Test Incrementally

```
# Build up complex expressions piece by piece
>> x: 5
>> y: 10
>> x + y
15
>> (x + y) * 2
30
>> x + y * 2    # Note: different result!
30
```

---

## Quick Reference

```
Left-to-Right Evaluation (no precedence)
  ↓
  Evaluate operators as they appear
  (operator arg1 arg2)
  ↓
  Continue with next operator
  ↑
Parentheses (...) force specific order
```

**Key Rule**: Operations are evaluated in the order they appear, left to right.

---

## Testing Evaluation

Use the REPL to verify left-to-right evaluation:

```
>> 3 + 4 * 2
14

>> 3 + (4 * 2)
11

>> 10 - 6 / 2
2

>> 10 - (6 / 2)
7
```

---

For more examples, see [`docs/repl-usage.md`](./repl-usage.md).
