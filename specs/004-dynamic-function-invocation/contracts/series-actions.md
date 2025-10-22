# Contract: Series Actions

**Component**: Series Operations as Actions
**Version**: 1.0
**Date**: 2025-10-13

## Purpose

Defines contracts for series operations (first, last, append, insert, length?) implemented as polymorphic actions. Each operation dispatches to type-specific implementations for blocks and strings.

---

## Action: first

### Signature
```viro
first series
```

### Description
Returns the first element of a series (block or string).

### Parameters
| Name | Type | Evaluation | Description |
|------|------|------------|-------------|
| `series` | block!/string! | Evaluated | The series to get first element from |

### Returns
- **Block**: First value in the block
- **String**: First character as a string (length 1)

### Supported Types
- `block!` → Returns first value in block
- `string!` → Returns first character as string

### Error Conditions
| Error Code | Condition | Example |
|------------|-----------|---------|
| `action-no-impl` | Type not block or string | `first 42` |
| `out-of-range` | Series is empty | `first []` or `first ""` |
| `wrong-arity` | Wrong number of arguments | `first` or `first [1] [2]` |

### Examples

```viro
>> first [1 2 3]
== 1

>> first "hello"
== "h"

>> first []
** Script error: Series is empty

>> first 42
** Script error: Action 'first' not defined for type integer!
```

### Test Cases

| Input | Expected Output | Description |
|-------|-----------------|-------------|
| `first [1 2 3]` | `1` | Block with integers |
| `first ["a" "b"]` | `"a"` | Block with strings |
| `first [[1 2] [3 4]]` | `[1 2]` | Nested blocks |
| `first "hello"` | `"h"` | String |
| `first "a"` | `"a"` | Single-character string |
| `first []` | Error: out-of-range | Empty block |
| `first ""` | Error: out-of-range | Empty string |
| `first 42` | Error: action-no-impl | Integer (unsupported) |

---

## Action: last

### Signature
```viro
last series
```

### Description
Returns the last element of a series (block or string).

### Parameters
| Name | Type | Evaluation | Description |
|------|------|------------|-------------|
| `series` | block!/string! | Evaluated | The series to get last element from |

### Returns
- **Block**: Last value in the block
- **String**: Last character as a string (length 1)

### Supported Types
- `block!` → Returns last value in block
- `string!` → Returns last character as string

### Error Conditions
| Error Code | Condition | Example |
|------------|-----------|---------|
| `action-no-impl` | Type not block or string | `last 42` |
| `out-of-range` | Series is empty | `last []` or `last ""` |
| `wrong-arity` | Wrong number of arguments | `last` or `last [1] [2]` |

### Examples

```viro
>> last [1 2 3]
== 3

>> last "hello"
== "o"

>> last []
** Script error: Series is empty

>> last 42
** Script error: Action 'last' not defined for type integer!
```

### Test Cases

| Input | Expected Output | Description |
|-------|-----------------|-------------|
| `last [1 2 3]` | `3` | Block with integers |
| `last ["a" "b"]` | `"b"` | Block with strings |
| `last [[1 2] [3 4]]` | `[3 4]` | Nested blocks |
| `last "hello"` | `"o"` | String |
| `last "a"` | `"a"` | Single-character string |
| `last []` | Error: out-of-range | Empty block |
| `last ""` | Error: out-of-range | Empty string |
| `last 42` | Error: action-no-impl | Integer (unsupported) |

---

## Action: append

### Signature
```viro
append series value
```

### Description
Appends a value to the end of a series. Modifies the series in-place and returns the modified series.

### Parameters
| Name | Type | Evaluation | Description |
|------|------|------------|-------------|
| `series` | block!/string! | Evaluated | The series to append to (modified in-place) |
| `value` | any | Evaluated | The value to append |

### Returns
- **Block**: The modified block (same reference, modified in-place)
- **String**: The modified string (same reference, modified in-place)

### Supported Types
- `block!` → Appends value to block (any value type accepted)
- `string!` → Appends string to string (value must be string)

### Error Conditions
| Error Code | Condition | Example |
|------------|-----------|---------|
| `action-no-impl` | Type not block or string | `append 42 3` |
| `type-mismatch` | String append with non-string value | `append "hello" 42` |
| `wrong-arity` | Wrong number of arguments | `append [1]` or `append [1] 2 3` |

### Examples

```viro
>> b: [1 2]
>> append b 3
== [1 2 3]
>> b
== [1 2 3]

>> s: "hel"
>> append s "lo"
== "hello"
>> s
== "hello"

>> append "test" 42
** Script error: Type mismatch (expected string, got integer)

>> append 42 3
** Script error: Action 'append' not defined for type integer!
```

### Test Cases

| Input | Expected Output | Description |
|-------|-----------------|-------------|
| `b: [1 2]; append b 3; b` | `[1 2 3]` | Append integer to block |
| `b: []; append b "a"; b` | `["a"]` | Append to empty block |
| `b: [1]; append b [2 3]; b` | `[1 [2 3]]` | Append block to block (nested) |
| `s: "hel"; append s "lo"; s` | `"hello"` | Append string to string |
| `s: ""; append s "a"; s` | `"a"` | Append to empty string |
| `append "test" 42` | Error: type-mismatch | Non-string appended to string |
| `append 42 3` | Error: action-no-impl | Integer (unsupported) |

---

## Action: insert

### Signature
```viro
insert series value
```

### Description
Inserts a value at the beginning of a series. Modifies the series in-place and returns the modified series.

### Parameters
| Name | Type | Evaluation | Description |
|------|------|------------|-------------|
| `series` | block!/string! | Evaluated | The series to insert into (modified in-place) |
| `value` | any | Evaluated | The value to insert at the beginning |

### Returns
- **Block**: The modified block (same reference, modified in-place)
- **String**: The modified string (same reference, modified in-place)

### Supported Types
- `block!` → Inserts value at beginning of block (any value type accepted)
- `string!` → Inserts string at beginning of string (value must be string)

### Error Conditions
| Error Code | Condition | Example |
|------------|-----------|---------|
| `action-no-impl` | Type not block or string | `insert 42 3` |
| `type-mismatch` | String insert with non-string value | `insert "hello" 42` |
| `wrong-arity` | Wrong number of arguments | `insert [1]` or `insert [1] 2 3` |

### Examples

```viro
>> b: [2 3]
>> insert b 1
== [1 2 3]
>> b
== [1 2 3]

>> s: "orld"
>> insert s "W"
== "World"
>> s
== "World"

>> insert "test" 42
** Script error: Type mismatch (expected string, got integer)

>> insert 42 3
** Script error: Action 'insert' not defined for type integer!
```

### Test Cases

| Input | Expected Output | Description |
|-------|-----------------|-------------|
| `b: [2 3]; insert b 1; b` | `[1 2 3]` | Insert integer at beginning |
| `b: []; insert b "a"; b` | `["a"]` | Insert into empty block |
| `b: [3]; insert b [1 2]; b` | `[[1 2] 3]` | Insert block at beginning |
| `s: "orld"; insert s "W"; s` | `"World"` | Insert string at beginning |
| `s: ""; insert s "a"; s` | `"a"` | Insert into empty string |
| `insert "test" 42` | Error: type-mismatch | Non-string inserted into string |
| `insert 42 3` | Error: action-no-impl | Integer (unsupported) |

---

## Action: length?

### Signature
```viro
length? series
```

### Description
Returns the number of elements in a series.

### Parameters
| Name | Type | Evaluation | Description |
|------|------|------------|-------------|
| `series` | block!/string! | Evaluated | The series to get length of |

### Returns
- **Integer**: Number of elements in block or characters in string

### Supported Types
- `block!` → Returns number of values in block
- `string!` → Returns number of characters in string

### Error Conditions
| Error Code | Condition | Example |
|------------|-----------|---------|
| `action-no-impl` | Type not block or string | `length? 42` |
| `wrong-arity` | Wrong number of arguments | `length?` or `length? [1] [2]` |

### Examples

```viro
>> length? [1 2 3]
== 3

>> length? "hello"
== 5

>> length? []
== 0

>> length? ""
== 0

>> length? 42
** Script error: Action 'length?' not defined for type integer!
```

### Test Cases

| Input | Expected Output | Description |
|-------|-----------------|-------------|
| `length? [1 2 3]` | `3` | Block with 3 elements |
| `length? []` | `0` | Empty block |
| `length? [[1 2] [3 4]]` | `2` | Nested blocks (2 top-level elements) |
| `length? "hello"` | `5` | String with 5 characters |
| `length? ""` | `0` | Empty string |
| `length? "a"` | `1` | Single-character string |
| `length? 42` | Error: action-no-impl | Integer (unsupported) |

---

## Common Invariants

### Type-Specific Behavior
1. **Block operations**: Accept any value type for `value` parameter
2. **String operations**: Only accept string values for `value` parameter (type-mismatch error otherwise)
3. **In-place modification**: `append` and `insert` modify the series and return the same reference
4. **Immutable operations**: `first`, `last`, `length?` do not modify the series

### Error Handling
1. **Unsupported type**: Always generates `action-no-impl` error with clear message
2. **Empty series**: `first` and `last` generate `out-of-range` error
3. **Type mismatch**: String operations validate value type, generate `type-mismatch` error
4. **Arity errors**: Standard function arity validation applies before dispatch

### Dispatch Behavior
1. All series actions dispatch on first argument type (the series)
2. Dispatcher uses global TypeRegistry to find type frame, then looks up action name in that frame
3. Type frames contain implementations: Block frame → block-specific funcs, String frame → string-specific funcs
4. Subsequent arguments validated by type-specific implementation (not dispatcher)

---

## Migration Checklist

For each series operation migrated from native to action:

- [ ] Type-specific implementation created:
  - [ ] Block implementation in `internal/native/series_block.go`
  - [ ] String implementation in `internal/native/series_string.go`
- [ ] Type-specific functions registered in their respective type frames
- [ ] Action value created (Name + ParamSpec only)
- [ ] Action registered in root frame (replacing old native)
- [ ] Existing contract tests updated to verify dispatch
- [ ] New contract tests added for unsupported type errors
- [ ] Performance benchmarks run (compare native vs action)
- [ ] Documentation updated with action semantics

---

## Performance Requirements

**Dispatch Overhead** (vs direct native call):
- Target: < 5x slowdown for action dispatch
- Measured: Benchmark each operation (first, last, append, insert, length?)
- Acceptable: If overhead < 10x, proceed; if > 10x, investigate optimization

**Memory Usage**:
- Action values: ~50 bytes each (5 actions × 50 = 250 bytes)
- Type frames: ~500 bytes total (2 types × ~250 bytes), stored in TypeRegistry
- Global TypeRegistry: ~80-160 bytes (~10 types × 8-16 bytes per pointer)
- Total overhead: ~1KB (negligible)

**Scalability**:
- Adding new series type: O(1) - create type frame, register in TypeRegistry, add functions to frame
- Adding new series action: O(1) - create action, add implementation to relevant type frames
- Runtime dispatch: O(1) - constant time regardless of number of actions/types
