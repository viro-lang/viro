# Contract: Math Natives

**Category**: Mathematical Operations  
**Functions**: `+`, `-`, `*`, `/`, `<`, `>`, `<=`, `>=`, `=`, `<>`, `and`, `or`, `not`  
**Purpose**: Arithmetic, comparison, and logic operations

---

## Evaluation Order

**Critical Note**: Viro implements **left-to-right evaluation** matching REBOL's evaluation model. There is **no operator precedence**.

**Evaluation Model**:

Operators are evaluated in the order they appear, from left to right. Each operator consumes its two operands and produces a result, which becomes the left operand for the next operator.

**Examples**:
```viro
3 + 4 * 2        ; → ((3 + 4) * 2) = (7 * 2) = 14
2 + 3 * 4        ; → ((2 + 3) * 4) = (5 * 4) = 20
10 - 4 / 2       ; → ((10 - 4) / 2) = (6 / 2) = 3
5 * 2 + 3        ; → ((5 * 2) + 3) = (10 + 3) = 13
1 + 2 < 4        ; → ((1 + 2) < 4) = (3 < 4) = true
true or false and false  ; → ((true or false) and false) = (true and false) = false
```

**Control Order with Parens**:
```viro
3 + (4 * 2)      ; → 3 + 8 = 11 (evaluate paren first)
10 / (2 + 3)     ; → 10 / 5 = 2 (evaluate paren first)
(3 + 4) * 2      ; → 7 * 2 = 14 (same as without parens due to left-to-right)
```

**Design Rationale**:
- **Why follow REBOL?** REBOL's left-to-right evaluation is simpler and more consistent with homoiconic philosophy. All operations are treated uniformly.
- **Why no precedence?** Eliminates need for complex precedence tables and parser logic. Users control order explicitly with parentheses when needed.
- **Implementation note**: Parser transforms infix notation to prefix calls in left-to-right order. No precedence climbing needed.

---

## Native: `+` (add)

**Signature**: `+ value1 value2`

**Parameters**:
- `value1`: Integer
- `value2`: Integer

**Return**: Integer (sum)

**Behavior**: Returns arithmetic sum of two integers

**Type Rules**:
- Both arguments must be Integer type (error otherwise)
- Phase 1 scope: integers only (decimals deferred)

**Examples**:
```viro
3 + 4        → 7
-5 + 10      → 5
0 + 0        → 0
```

**Error Cases**:
- First argument not integer → Script error: "Add expects integer arguments"
- Second argument not integer → Script error: "Add expects integer arguments"
- Overflow (implementation-defined) → Math error: "Integer overflow"

**Test Cases**:
1. `3 + 4` returns `7`
2. `-5 + 10` returns `5`
3. `0 + 0` returns `0`
4. Large positive overflow detection
5. `"3" + 4` errors (type mismatch)
6. **Precedence**: `3 + 4 * 2` returns `11` (not 14 - multiplication first)
7. **Precedence**: `2 * 3 + 4` returns `10` (multiplication before addition)
8. **Paren override**: `(3 + 4) * 2` returns `14`

---

## Native: `-` (subtract)

**Signature**: `- value1 value2`

**Parameters**:
- `value1`: Integer
- `value2`: Integer

**Return**: Integer (difference)

**Behavior**: Returns arithmetic difference (value1 - value2)

**Examples**:
```viro
10 - 3       → 7
5 - 10       → -5
0 - 0        → 0
```

**Error Cases**: Same as `+`

**Test Cases**:
1. `10 - 3` returns `7`
2. `5 - 10` returns `-5`
3. `0 - 0` returns `0`
4. Underflow detection
5. `10 - "3"` errors
6. **Precedence**: `10 - 4 / 2` returns `8` (not 3 - division first)
7. **Precedence**: `5 * 2 - 3` returns `7` (multiplication before subtraction)

---

## Native: `*` (multiply)

**Signature**: `* value1 value2`

**Parameters**:
- `value1`: Integer
- `value2`: Integer

**Return**: Integer (product)

**Behavior**: Returns arithmetic product

**Examples**:
```viro
3 * 4        → 12
-2 * 5       → -10
0 * 100      → 0
```

**Error Cases**: Same as `+` (overflow detection)

**Test Cases**:
1. `3 * 4` returns `12`
2. `-2 * 5` returns `-10`
3. `0 * 100` returns `0`
4. Large multiplication overflow detection
5. **Precedence**: `2 + 3 * 4` returns `14` (not 20 - multiplication first)
6. **Precedence**: Same level as `/`, evaluates left-to-right: `10 / 2 * 3` returns `15` ((10/2)*3)

---

## Native: `/` (divide)

**Signature**: `/ value1 value2`

**Parameters**:
- `value1`: Integer
- `value2`: Integer (must be non-zero)

**Return**: Integer (quotient, truncated toward zero)

**Behavior**: Returns integer division result (truncated)

**Examples**:
```viro
10 / 3       → 3
-10 / 3      → -3
0 / 5        → 0
```

**Error Cases**:
- Division by zero → Math error (400): "Division by zero"
- Type mismatch → Script error (same as other math ops)

**Test Cases**:
1. `10 / 3` returns `3` (truncated)
2. `-10 / 3` returns `-3` (truncated toward zero)
3. `0 / 5` returns `0`
4. `10 / 0` errors (division by zero)
5. `10 / "2"` errors (type mismatch)
6. **Precedence**: `20 / 4 + 2` returns `7` (not 3 - division first: (20/4)+2)
7. **Precedence**: Same level as `*`, evaluates left-to-right: `20 / 2 / 5` returns `2` ((20/2)/5)

---

## Native: `<` (less than)

**Signature**: `< value1 value2`

**Parameters**:
- `value1`: Integer
- `value2`: Integer

**Return**: Logic (true/false)

**Behavior**: Returns true if value1 < value2

**Examples**:
```viro
3 < 5        → true
5 < 3        → false
3 < 3        → false
```

**Test Cases**:
1. `3 < 5` returns `true`
2. `5 < 3` returns `false`
3. `3 < 3` returns `false`
4. `-10 < 0` returns `true`
5. **Precedence**: `1 + 2 < 5` returns `true` (arithmetic first: (1+2)<5 = 3<5)
6. **Precedence**: `3 * 2 > 5` returns `true` (multiplication first: (3*2)>5 = 6>5)

---

## Native: `>` (greater than)

**Signature**: `> value1 value2`

**Parameters**:
- `value1`: Integer
- `value2`: Integer

**Return**: Logic (true/false)

**Behavior**: Returns true if value1 > value2

**Examples**:
```viro
5 > 3        → true
3 > 5        → false
3 > 3        → false
```

**Test Cases**: Mirror `<` tests

---

## Native: `<=` (less than or equal)

**Signature**: `<= value1 value2`

**Return**: Logic (true if value1 ≤ value2)

**Examples**:
```viro
3 <= 5       → true
5 <= 3       → false
3 <= 3       → true
```

---

## Native: `>=` (greater than or equal)

**Signature**: `>= value1 value2`

**Return**: Logic (true if value1 ≥ value2)

**Examples**:
```viro
5 >= 3       → true
3 >= 5       → false
3 >= 3       → true
```

---

## Native: `=` (equal)

**Signature**: `= value1 value2`

**Parameters**:
- `value1`: Any type
- `value2`: Any type

**Return**: Logic (true if equal)

**Behavior**: 
- Integers: numeric equality
- Strings: character-wise equality (case-sensitive per spec)
- Blocks: structural equality (same length, equal elements)
- Words: symbol name equality (case-sensitive)
- Logic: boolean equality
- None: both none → true
- Different types → false

**Examples**:
```viro
5 = 5        → true
5 = 3        → false
"abc" = "abc"  → true
"abc" = "ABC"  → false (case-sensitive)
[1 2] = [1 2]  → true
[1 2] = [2 1]  → false
5 = "5"      → false (different types)
```

**Test Cases**:
1. Integer equality: `5 = 5` → true, `5 = 3` → false
2. String equality: `"abc" = "abc"` → true
3. Case sensitivity: `"abc" = "ABC"` → false
4. Block equality: `[1 2] = [1 2]` → true
5. Type mismatch: `5 = "5"` → false
6. None equality: `none = none` → true
7. **Precedence**: `3 < 5 = true` returns `true` (comparison first: (3<5)=true = true=true)
8. **Precedence**: `2 + 3 = 5` returns `true` (arithmetic first: (2+3)=5 = 5=5)

---

## Native: `<>` (not equal)

**Signature**: `<> value1 value2`

**Return**: Logic (true if not equal)

**Behavior**: Negation of `=`

**Examples**:
```viro
5 <> 3       → true
5 <> 5       → false
```

**Test Cases**: Inverse of `=` test cases

---

## Native: `and`

**Signature**: `and value1 value2`

**Parameters**:
- `value1`: Any type (converted to logic)
- `value2`: Any type (converted to logic)

**Return**: Logic (true if both truthy)

**Behavior**: 
- Evaluate both arguments (no short-circuit)
- Apply truthy conversion to each
- Return true if both truthy, false otherwise

**Truthy Conversion**:
- `false` → false
- `none` → false
- All others → true

**Examples**:
```viro
true and true    → true
true and false   → false
false and true   → false
false and false  → false
1 and 2          → true (both truthy)
0 and 1          → true (0 is truthy!)
none and true    → false
```

**Test Cases**:
1. `true and true` → true
2. `true and false` → false
3. `false and false` → false
4. `1 and 2` → true (both truthy)
5. `none and true` → false (none is falsy)
6. `0 and 1` → true (0 is truthy in REBOL)
7. **Precedence**: `true or false and false` returns `true` (and first: true or (false and false) = true or false = true)
8. **Precedence**: `1 = 1 and 2 = 2` returns `true` (equality first: (1=1) and (2=2) = true and true)

---

## Native: `or`

**Signature**: `or value1 value2`

**Parameters**:
- `value1`: Any type (converted to logic)
- `value2`: Any type (converted to logic)

**Return**: Logic (true if either truthy)

**Behavior**: Returns true if at least one argument truthy

**Examples**:
```viro
true or false    → true
false or true    → true
false or false   → false
1 or none        → true
none or none     → false
```

**Test Cases**:
1. `true or false` → true
2. `false or false` → false
3. `1 or none` → true
4. `none or none` → false

---

## Native: `not`

**Signature**: `not value`

**Parameters**:
- `value`: Any type (converted to logic)

**Return**: Logic (negation)

**Behavior**: Returns logical negation of truthy conversion

**Examples**:
```viro
not true         → false
not false        → true
not none         → true (none is falsy)
not 0            → false (0 is truthy!)
not 1            → false
```

**Test Cases**:
1. `not true` → false
2. `not false` → true
3. `not none` → true (none is falsy)
4. `not 0` → false (0 is truthy)
5. `not 1` → false
6. **Precedence**: `not true and false` returns `false` (not first: (not true) and false = false and false)
7. **Precedence**: `not 1 < 2` returns `false` (comparison first: not (1<2) = not true)

---

## Common Properties

**Integer Operations** (`+`, `-`, `*`, `/`, comparisons):
- Operate on 64-bit signed integers (Go int64)
- Overflow/underflow handling: detect and error (Math error category)
- Strict type checking: only integers accepted in Phase 1

**Comparison Operations** (`<`, `>`, `<=`, `>=`):
- Phase 1: integers only
- Return logic values (true/false)
- Type mismatch → Script error

**Equality Operations** (`=`, `<>`):
- Accept any types
- Different types → false (not error)
- Same types → deep equality check

**Logic Operations** (`and`, `or`, `not`):
- Accept any types
- Apply truthy conversion (none/false → false, others → true)
- Return logic values
- **Note**: `0`, `""`, `[]` are truthy (REBOL convention, differs from some languages)

**Error Messages**:
- Include operation name and argument types
- Example: "Add expects integer arguments, got string and integer"

---

## Implementation Checklist

For each native:
- [ ] Function signature matches contract
- [ ] Type validation with clear error messages
- [ ] Overflow/underflow detection (arithmetic ops)
- [ ] Truthy conversion correct (logic ops)
- [ ] Equality semantics correct (deep comparison for blocks)
- [ ] All test cases pass (including precedence tests)
- [ ] Performance acceptable (simple ops should be <1µs per research.md)

**Precedence Implementation**:
- [ ] Parser builds AST respecting precedence table (not simple left-to-right)
- [ ] Test all 7 precedence levels with mixed expressions
- [ ] Parentheses correctly override precedence
- [ ] Associativity correct (left-to-right for same-level operators)
- [ ] Complex expressions: `2 + 3 * 4 - 10 / 2 > 5 and true` correctly parsed

**Dependencies**:
- Type system (type checking, truthy conversion)
- Error system (Math error, Script error)
- Value system (equality comparison for composite types)

**Testing Strategy**:
- Table-driven tests per operation
- Boundary cases (min/max int64, zero, negative)
- Type mismatch error cases
- Performance benchmarks (arithmetic should be fast)
