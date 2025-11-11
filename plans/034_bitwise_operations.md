# Plan 034: Bitwise Operations for Viro

## Feature Summary

Add comprehensive bitwise operations to Viro through a new `bit` object containing native functions for bit manipulation. This feature provides:

1. **Bitwise logical operations** (AND, OR, XOR, NOT) supporting both `integer!` and `binary!` types
2. **Bit shift operations** (`<<`, `>>`) as global infix operators
3. **Individual bit manipulation** (set/clear specific bits in integers)
4. **Bit counting** (count set bits)

All operations follow Viro's type system conventions and left-to-right evaluation model.

## Research Findings

### 1. Object Creation Pattern (from `system` object in bootstrap.go)

**Location:** `internal/bootstrap/bootstrap.go` lines 65-82

```go
func InjectSystemArgs(evaluator core.Evaluator, args []string) {
    // Create owned frame for object storage
    ownedFrame := frame.NewFrame(frame.FrameObject, -1)
    ownedFrame.Bind("args", argsBlock)
    
    // Create object with owned frame
    systemObj := value.NewObject(ownedFrame)
    
    // Bind object to root frame
    rootFrame := evaluator.GetFrameByIndex(0)
    rootFrame.Bind("system", systemObj)
}
```

**Key insights:**
- Objects use dedicated frames with `frame.FrameObject` type
- Parent frame index is `-1` (no parent)
- Object fields are bindings in the owned frame
- Objects are bound to root frame as normal values

### 2. Infix Function Registration (from register_math.go)

**Location:** `internal/native/register_math.go` lines 35-56

```go
registerAndBind("+", value.NewNativeFunction("+",
    []value.ParamSpec{
        value.NewParamSpec("left", true),
        value.NewParamSpec("right", true),
    },
    Add,
    true,  // ← INFIX FLAG
    &NativeDoc{...}
))
```

**Key insights:**
- Fifth parameter to `NewNativeFunction` is the infix flag (bool)
- `true` = function can be used in infix notation
- No parser changes needed - infix is a function property
- Works with left-to-right evaluation model

### 3. Native Function Type Dispatch Pattern (from math.go)

**Location:** `internal/native/math.go` lines 31-82

```go
func mathOp(op func(int64, int64) int64, opName string) core.NativeFunc {
    return func(args []core.Value, ...) (core.Value, error) {
        // Type checking via switch on GetType()
        switch args[0].GetType() {
        case value.TypeInteger:
            // Handle integer case
        case value.TypeDecimal:
            // Handle decimal case
        default:
            return value.NewNoneVal(), typeError(opName, "integer! decimal!", args[0])
        }
    }
}
```

**Key insights:**
- Use `GetType()` for type dispatch
- Return type errors for unsupported types
- Can handle multiple types in single function
- Use helper functions to reduce code duplication

### 4. Binary Manipulation Patterns (from binary.go)

**Location:** `internal/value/binary.go` lines 12-292

```go
type BinaryValue struct {
    data  []byte
    index int
}

func (b *BinaryValue) Bytes() []byte {
    return b.data
}

func NewBinaryValue(data []byte) *BinaryValue {
    return &BinaryValue{
        data:  data,
        index: 0,
    }
}
```

**Key insights:**
- Binary data stored as `[]byte`
- Use `Bytes()` to access underlying data
- Create new BinaryValue with `NewBinaryValue([]byte)`
- Binary series support index tracking

### 5. Test Organization Pattern (from binary_test.go, math_test.go)

**Location:** `test/contract/math_test.go` lines 1-100

```go
func TestArithmeticNatives(t *testing.T) {
    tests := []struct {
        name     string
        op       string
        args     []core.Value
        expected core.Value
        wantErr  bool
    }{
        {
            name:     "add positive integers",
            op:       "+",
            args:     []core.Value{value.NewIntVal(3), value.NewIntVal(4)},
            expected: value.NewIntVal(7),
            wantErr:  false,
        },
        // More tests...
    }
}
```

**Key insights:**
- Table-driven tests with clear naming
- Test both success and error cases
- Use value constructors in tests
- Organize by operation category

### 6. Tokenizer Analysis

**Location:** `internal/tokenize/tokenizer.go`

The tokenizer uses `readLiteral()` to parse symbols and already handles multi-character operators like `<=`, `>=`, `<>`. The `<<` and `>>` operators will be tokenized as literals (words) and work through the existing infix mechanism without tokenizer changes.

**Key insight:** No tokenizer changes needed - `<<` and `>>` are valid literal tokens

### 7. Operator Evaluation Model

**Location:** `docs/operator-precedence.md`

Viro uses **left-to-right evaluation with no operator precedence**. This means:
- `2 << 3 + 1` evaluates as `(2 << 3) + 1` = `16 + 1` = `17`
- All infix operators are evaluated left-to-right
- Parentheses force specific evaluation order

**Key insight:** Bitwise operators follow left-to-right evaluation, no special precedence

## Architecture Overview

### Design Principles

1. **Object-based organization:** Group related functions under `bit` namespace
2. **Type polymorphism:** Support both `integer!` and `binary!` where semantically appropriate
3. **No mixed-type operations:** Operands must be same type
4. **Type preservation:** Operations return same type as input
5. **Infix support:** Common operations available in natural infix syntax
6. **Go stdlib leverage:** Use `math/bits` for efficient bit counting

### Component Structure

```
bit object (in root frame)
├── bit.and     (infix, integer!/binary!)
├── bit.or      (infix, integer!/binary!)
├── bit.xor     (infix, integer!/binary!)
├── bit.not     (prefix, integer!/binary!)
├── bit.shl     (infix, integer!/binary!, also <<)
├── bit.shr     (infix, integer!/binary!, also >>)
├── bit.on      (prefix, integer! only)
├── bit.off     (prefix, integer! only)
└── bit.count   (prefix, integer!/binary!)

Global operators (aliases to bit.shl/bit.shr)
├── <<          (infix, bound to bit.shl)
└── >>          (infix, bound to bit.shr)
```

### Type Semantics

#### For `integer!` Operations

- Use Go's standard bitwise operations (`&`, `|`, `^`, `&^`, `<<`, `>>`)
- Two's complement representation (Go's `int64`)
- Right shift is **arithmetic** (sign-extending, not logical)
- Examples:
  - `2 bit.and 3` → `2` (binary: 10 & 11 = 10)
  - `-1 >> 1` → `-1` (sign extension: all 1's remain all 1's)
  - `bit.not 0` → `-1` (bitwise complement)

#### For `binary!` Operations

- Operate byte-by-byte on binary data
- **Alignment:** Bytes are compared from the right (least significant byte first)
  - `#{01 02} bit.and #{03}` compares byte `02` with `03`, byte `01` with implicit `00`
- **Different-length handling:** 
  - **`bit.and`:** Zero remaining bytes (shorter binary treated as zero-padded)
    - `#{FF FF} bit.and #{FF}` → `#{00 FF}` (left byte ANDed with 0)
  - **`bit.or`:** Copy remaining bytes from longer (X OR 0 = X)
    - `#{FF FF} bit.or #{FF}` → `#{FF FF}` (left byte copied)
  - **`bit.xor`:** Copy remaining bytes from longer (X XOR 0 = X)
    - `#{FF FF} bit.xor #{FF}` → `#{FF 00}` (left byte copied)
- **Shift semantics:** No new bytes created, overflow is lost
  - `#{01} << 1` → `#{02}` (within byte boundary)
  - `#{80} << 1` → `#{00}` (high bit lost, no carry to new byte)
  - `#{01} >> 1` → `#{00}` (underflow lost)
  - Shifts operate within series boundaries
- **NOT operation:** Flips all bits in all bytes
  - `bit.not #{FF 00}` → `#{00 FF}`

**Rationale for binary operations:**
- Right-alignment matches numeric interpretation (LSB first)
- Different operators have different padding semantics matching their logical behavior
- No automatic growth maintains series semantics
- Predictable memory usage

### Error Handling

- **Type mismatch:** `bit.and 3 "hello"` → type error
- **Mixed types:** `bit.and 3 #{FF}` → type error "operands must be same type"
- **Integer-only operations:** `bit.on #{FF} 2` → type error "expects integer!"
- **Out of bounds:** `bit.on 5 100` → no error, sets bit 100 (may result in large integer)

## Implementation Roadmap

### Step 1: Create Bitwise Native Functions Module

**File:** `internal/native/bitwise.go` (NEW FILE)

**Purpose:** Implement all bitwise operation functions with type dispatch

**Functions to implement:**

1. **BitAnd** - Bitwise AND
   - Integer: `a & b`
   - Binary: byte-by-byte AND from right, zero remaining bytes

2. **BitOr** - Bitwise OR
   - Integer: `a | b`
   - Binary: byte-by-byte OR from right, copy remaining from longer

3. **BitXor** - Bitwise XOR
   - Integer: `a ^ b`
   - Binary: byte-by-byte XOR from right, copy remaining from longer

4. **BitNot** - Bitwise NOT
   - Integer: `^a` (bitwise complement)
   - Binary: flip all bits in all bytes

5. **BitShl** - Shift left
   - Integer: `a << count`
   - Binary: shift bytes left, losing overflow

6. **BitShr** - Shift right
   - Integer: `a >> count` (arithmetic, preserves sign)
   - Binary: shift bytes right, losing underflow

7. **BitOn** - Set bit to 1 (integer only)
   - `bitOn(value, position)` → `value | (1 << position)`

8. **BitOff** - Clear bit to 0 (integer only)
   - `bitOff(value, position)` → `value &^ (1 << position)`

9. **BitCount** - Count set bits
   - Integer: `bits.OnesCount64(uint64(value))`
   - Binary: count 1-bits across all bytes

**Implementation pattern:**

```go
package native

import (
    "math/bits"
    "github.com/marcin-radoszewski/viro/internal/core"
    "github.com/marcin-radoszewski/viro/internal/value"
    "github.com/marcin-radoszewski/viro/internal/verror"
)

func BitAnd(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
    if len(args) != 2 {
        return value.NewNoneVal(), arityError("bit.and", 2, len(args))
    }
    
    leftType := args[0].GetType()
    rightType := args[1].GetType()
    
    if leftType != rightType {
        return value.NewNoneVal(), verror.NewScriptError(
            verror.ErrIDTypeMismatch,
            [3]string{"bit.and", "operands must be same type", ""},
        )
    }
    
    switch leftType {
    case value.TypeInteger:
        left, _ := value.AsIntValue(args[0])
        right, _ := value.AsIntValue(args[1])
        return value.NewIntVal(left & right), nil
        
    case value.TypeBinary:
        left, _ := value.AsBinaryValue(args[0])
        right, _ := value.AsBinaryValue(args[1])
        return binaryAnd(left, right), nil
        
    default:
        return value.NewNoneVal(), typeError("bit.and", "integer! binary!", args[0])
    }
}

func binaryAnd(left, right *value.BinaryValue) core.Value {
    leftBytes := left.Bytes()
    rightBytes := right.Bytes()
    
    maxLen := len(leftBytes)
    if len(rightBytes) > maxLen {
        maxLen = len(rightBytes)
    }
    
    result := make([]byte, maxLen)
    
    leftLen := len(leftBytes)
    rightLen := len(rightBytes)
    
    for i := 0; i < maxLen; i++ {
        leftIdx := leftLen - maxLen + i
        rightIdx := rightLen - maxLen + i
        
        leftByte := byte(0)
        if leftIdx >= 0 {
            leftByte = leftBytes[leftIdx]
        }
        
        rightByte := byte(0)
        if rightIdx >= 0 {
            rightByte = rightBytes[rightIdx]
        }
        
        result[i] = leftByte & rightByte
    }
    
    return value.NewBinaryValue(result)
}
```

**Helper functions:**
- `binaryAnd(left, right *value.BinaryValue) core.Value` - Right-aligned AND, zero-pad shorter
- `binaryOr(left, right *value.BinaryValue) core.Value` - Right-aligned OR, copy longer's bytes
- `binaryXor(left, right *value.BinaryValue) core.Value` - Right-aligned XOR, copy longer's bytes
- `binaryNot(b *value.BinaryValue) core.Value` - Flip all bits
- `binaryShl(b *value.BinaryValue, count int64) core.Value` - Shift left, lose overflow
- `binaryShr(b *value.BinaryValue, count int64) core.Value` - Shift right, lose underflow
- `countBinaryBits(b *value.BinaryValue) int64` - Count set bits across all bytes

**Validation:** File compiles, functions available for registration

### Step 2: Create Registration Function

**File:** `internal/native/register_bitwise.go` (NEW FILE)

**Purpose:** Register all bitwise functions and create the `bit` object

**Pattern:** Similar to `RegisterMathNatives`

```go
package native

import (
    "github.com/marcin-radoszewski/viro/internal/core"
    "github.com/marcin-radoszewski/viro/internal/frame"
    "github.com/marcin-radoszewski/viro/internal/value"
)

func RegisterBitwiseNatives(rootFrame core.Frame) {
    // Create owned frame for bit object
    bitFrame := frame.NewFrame(frame.FrameObject, -1)
    
    // Register functions to bit object frame
    bitFrame.Bind("and", value.NewFuncVal(value.NewNativeFunction(
        "bit.and",
        []value.ParamSpec{
            value.NewParamSpec("left", true),
            value.NewParamSpec("right", true),
        },
        BitAnd,
        true,  // infix
        &NativeDoc{
            Category: "Bitwise",
            Summary:  "Performs bitwise AND operation",
            Description: `Performs bitwise AND on two values of the same type.
For integers: standard bitwise AND using two's complement.
For binaries: byte-by-byte AND from right (LSB first), zeros remaining bytes from longer operand.`,
            Parameters: []ParamDoc{
                {Name: "left", Type: "integer! binary!", Description: "First operand", Optional: false},
                {Name: "right", Type: "integer! binary!", Description: "Second operand (must match left type)", Optional: false},
            },
            Returns:  "Same type as input",
            Examples: []string{"2 bit.and 3  ; => 2", "#{FF 00} bit.and #{0F FF}  ; => #{0F 00}", "#{FF FF} bit.and #{FF}  ; => #{00 FF}"},
            SeeAlso:  []string{"bit.or", "bit.xor", "bit.not"},
            Tags:     []string{"bitwise", "logic"},
        },
    )))
    
    // Register bit.or, bit.xor, bit.not, bit.on, bit.off, bit.count similarly
    // ...
    
    // Register shift functions
    shlFunc := value.NewFuncVal(value.NewNativeFunction(
        "bit.shl",
        []value.ParamSpec{
            value.NewParamSpec("value", true),
            value.NewParamSpec("count", true),
        },
        BitShl,
        true,  // infix
        &NativeDoc{...},
    ))
    bitFrame.Bind("shl", shlFunc)
    
    shrFunc := value.NewFuncVal(value.NewNativeFunction(
        "bit.shr",
        []value.ParamSpec{
            value.NewParamSpec("value", true),
            value.NewParamSpec("count", true),
        },
        BitShr,
        true,  // infix
        &NativeDoc{...},
    ))
    bitFrame.Bind("shr", shrFunc)
    
    // Create bit object
    bitObj := value.NewObject(bitFrame)
    
    // Bind bit object to root frame
    rootFrame.Bind("bit", bitObj)
    
    // Also bind << and >> as global operators (aliases)
    rootFrame.Bind("<<", shlFunc)
    rootFrame.Bind(">>", shrFunc)
}
```

**Validation:** 
- Bit object created successfully
- All functions accessible via `bit.and`, `bit.or`, etc.
- Global `<<` and `>>` operators work

### Step 3: Integrate Registration into Bootstrap

**File:** `internal/bootstrap/bootstrap.go`

**Change:** Add bitwise natives registration

**Location:** In `NewEvaluatorWithNatives` function (around line 54)

**Modification:**

```go
func NewEvaluatorWithNatives(stdout, stderr io.Writer, stdin io.Reader, quiet bool) *eval.Evaluator {
    evaluator := eval.NewEvaluator()
    
    // ... existing I/O setup ...
    
    rootFrame := evaluator.GetFrameByIndex(0)
    native.RegisterMathNatives(rootFrame)
    native.RegisterSeriesNatives(rootFrame)
    native.RegisterDataNatives(rootFrame)
    native.RegisterIONatives(rootFrame, evaluator)
    native.RegisterControlNatives(rootFrame)
    native.RegisterHelpNatives(rootFrame)
    native.RegisterBitwiseNatives(rootFrame)  // ← ADD THIS
    
    return evaluator
}
```

**Validation:** 
- Build succeeds
- `bit` object available in REPL
- Can execute `bit.and 2 3` and get `2`

### Step 4: Write Comprehensive Test Suite (TDD)

**File:** `test/contract/bitwise_test.go` (NEW FILE)

**Test categories:**

#### 4.1 Integer Bitwise Operations

```go
func TestBitwiseInteger_AND(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected core.Value
        wantErr  bool
    }{
        {
            name:     "and with all bits set",
            input:    "bit.and 255 15",
            expected: value.NewIntVal(15),
        },
        {
            name:     "and infix notation",
            input:    "6 bit.and 3",
            expected: value.NewIntVal(2),
        },
        {
            name:     "and with zero",
            input:    "bit.and 42 0",
            expected: value.NewIntVal(0),
        },
        {
            name:     "and with negative",
            input:    "bit.and -1 255",
            expected: value.NewIntVal(255),  // -1 has all bits set
        },
    }
    // ... test execution loop
}

func TestBitwiseInteger_OR(t *testing.T) {
    // Similar pattern for OR
}

func TestBitwiseInteger_XOR(t *testing.T) {
    // Similar pattern for XOR
}

func TestBitwiseInteger_NOT(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected core.Value
    }{
        {
            name:     "not zero",
            input:    "bit.not 0",
            expected: value.NewIntVal(-1),
        },
        {
            name:     "not negative one",
            input:    "bit.not -1",
            expected: value.NewIntVal(0),
        },
        {
            name:     "not positive",
            input:    "bit.not 5",
            expected: value.NewIntVal(-6),
        },
    }
    // ... test execution
}
```

#### 4.2 Binary Bitwise Operations

```go
func TestBitwiseBinary_AND(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected core.Value
    }{
        {
            name:     "and same length",
            input:    "bit.and #{FF 00} #{0F F0}",
            expected: value.NewBinaryValue([]byte{0x0F, 0x00}),
        },
        {
            name:     "and different length - left longer",
            input:    "bit.and #{FF 00} #{0F}",
            expected: value.NewBinaryValue([]byte{0x00, 0x00}),  // Left byte ANDed with 0
        },
        {
            name:     "and different length - right longer",
            input:    "bit.and #{0F} #{FF 00}",
            expected: value.NewBinaryValue([]byte{0x00, 0x0F}),  // Right byte ANDed with 0
        },
        {
            name:     "and infix",
            input:    "#{AA} bit.and #{55}",
            expected: value.NewBinaryValue([]byte{0x00}),
        },
        {
            name:     "and right-aligned comparison",
            input:    "bit.and #{01 02} #{03}",
            expected: value.NewBinaryValue([]byte{0x00, 0x02}),  // 01 & 00, 02 & 03
        },
    }
    // ... test execution
}

func TestBitwiseBinary_NOT(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected core.Value
    }{
        {
            name:     "not all ones",
            input:    "bit.not #{FF}",
            expected: value.NewBinaryValue([]byte{0x00}),
        },
        {
            name:     "not all zeros",
            input:    "bit.not #{00}",
            expected: value.NewBinaryValue([]byte{0xFF}),
        },
        {
            name:     "not multiple bytes",
            input:    "bit.not #{FF 00 AA}",
            expected: value.NewBinaryValue([]byte{0x00, 0xFF, 0x55}),
        },
    }
    // ... test execution
}
```

#### 4.3 Shift Operations

```go
func TestBitwiseInteger_Shifts(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected core.Value
    }{
        {
            name:     "left shift simple",
            input:    "1 << 2",
            expected: value.NewIntVal(4),
        },
        {
            name:     "left shift using bit.shl",
            input:    "3 bit.shl 4",
            expected: value.NewIntVal(48),
        },
        {
            name:     "right shift simple",
            input:    "8 >> 2",
            expected: value.NewIntVal(2),
        },
        {
            name:     "right shift negative (arithmetic)",
            input:    "-16 >> 2",
            expected: value.NewIntVal(-4),  // Sign extension
        },
        {
            name:     "right shift negative by one",
            input:    "-1 >> 1",
            expected: value.NewIntVal(-1),  // All 1's stays all 1's
        },
    }
    // ... test execution
}

func TestBitwiseBinary_Shifts(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected core.Value
    }{
        {
            name:     "left shift within byte",
            input:    "#{01} << 2",
            expected: value.NewBinaryValue([]byte{0x04}),
        },
        {
            name:     "left shift overflow lost",
            input:    "#{80} << 1",
            expected: value.NewBinaryValue([]byte{0x00}),
        },
        {
            name:     "right shift within byte",
            input:    "#{08} >> 2",
            expected: value.NewBinaryValue([]byte{0x02}),
        },
        {
            name:     "right shift underflow lost",
            input:    "#{01} >> 1",
            expected: value.NewBinaryValue([]byte{0x00}),
        },
        {
            name:     "left shift multi-byte",
            input:    "#{01 00} << 8",
            expected: value.NewBinaryValue([]byte{0x00, 0x01}),  // Shift across bytes
        },
    }
    // ... test execution
}
```

#### 4.4 Bit Manipulation (Integer Only)

```go
func TestBitwiseInteger_BitOn(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected core.Value
    }{
        {
            name:     "set bit 0",
            input:    "bit.on 0 0",
            expected: value.NewIntVal(1),
        },
        {
            name:     "set bit 3",
            input:    "bit.on 0 3",
            expected: value.NewIntVal(8),
        },
        {
            name:     "set already-set bit",
            input:    "bit.on 5 0",
            expected: value.NewIntVal(5),  // No change
        },
        {
            name:     "set high bit",
            input:    "bit.on 0 7",
            expected: value.NewIntVal(128),
        },
    }
    // ... test execution
}

func TestBitwiseInteger_BitOff(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected core.Value
    }{
        {
            name:     "clear bit 0",
            input:    "bit.off 1 0",
            expected: value.NewIntVal(0),
        },
        {
            name:     "clear bit 3",
            input:    "bit.off 15 3",
            expected: value.NewIntVal(7),
        },
        {
            name:     "clear already-clear bit",
            input:    "bit.off 4 0",
            expected: value.NewIntVal(4),  // No change
        },
    }
    // ... test execution
}
```

#### 4.5 Bit Counting

```go
func TestBitwiseInteger_Count(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected core.Value
    }{
        {
            name:     "count zero",
            input:    "bit.count 0",
            expected: value.NewIntVal(0),
        },
        {
            name:     "count single bit",
            input:    "bit.count 8",
            expected: value.NewIntVal(1),
        },
        {
            name:     "count multiple bits",
            input:    "bit.count 15",
            expected: value.NewIntVal(4),  // 0b1111
        },
        {
            name:     "count negative",
            input:    "bit.count -1",
            expected: value.NewIntVal(64),  // All bits set in int64
        },
    }
    // ... test execution
}

func TestBitwiseBinary_Count(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected core.Value
    }{
        {
            name:     "count all zeros",
            input:    "bit.count #{00}",
            expected: value.NewIntVal(0),
        },
        {
            name:     "count all ones",
            input:    "bit.count #{FF}",
            expected: value.NewIntVal(8),
        },
        {
            name:     "count multiple bytes",
            input:    "bit.count #{FF 00 0F}",
            expected: value.NewIntVal(12),  // 8 + 0 + 4
        },
    }
    // ... test execution
}
```

#### 4.6 Error Cases

```go
func TestBitwiseErrors(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        wantErr bool
        errMsg  string
    }{
        {
            name:    "mixed types integer and binary",
            input:   "bit.and 3 #{FF}",
            wantErr: true,
            errMsg:  "operands must be same type",
        },
        {
            name:    "wrong type for bit.and",
            input:   `bit.and "hello" "world"`,
            wantErr: true,
            errMsg:  "type mismatch",
        },
        {
            name:    "bit.on with binary",
            input:   "bit.on #{FF} 2",
            wantErr: true,
            errMsg:  "type mismatch",
        },
        {
            name:    "bit.off with binary",
            input:   "bit.off #{FF} 2",
            wantErr: true,
            errMsg:  "type mismatch",
        },
        {
            name:    "shift with wrong type",
            input:   `"hello" << 2`,
            wantErr: true,
            errMsg:  "type mismatch",
        },
    }
    // ... test execution
}
```

#### 4.7 Integration and Composition

```go
func TestBitwiseComposition(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected core.Value
    }{
        {
            name:     "left-to-right evaluation",
            input:    "2 << 3 + 1",  // (2 << 3) + 1 = 16 + 1
            expected: value.NewIntVal(17),
        },
        {
            name:     "parentheses control order",
            input:    "2 << (3 + 1)",  // 2 << 4 = 32
            expected: value.NewIntVal(32),
        },
        {
            name:     "multiple bitwise ops",
            input:    "15 bit.and 7 bit.or 8",  // ((15 & 7) | 8) = (7 | 8) = 15
            expected: value.NewIntVal(15),
        },
        {
            name:     "bit manipulation chain",
            input:    "x: 0\nx: bit.on x 0\nx: bit.on x 2\nx",
            expected: value.NewIntVal(5),  // bits 0 and 2 set
        },
    }
    // ... test execution
}
```

**Test execution helper:**

```go
func runBitwiseTest(t *testing.T, input string) (core.Value, error) {
    evaluator := testutil.NewTestEvaluator()
    result, err := evaluator.Run(input)
    return result, err
}
```

**Validation:**
- All tests written and documented
- Tests FAIL initially (functions not yet implemented)
- Coverage for all operations and edge cases

### Step 5: Implement Bitwise Functions

**File:** `internal/native/bitwise.go`

Implement all functions following the patterns established in Step 1.

**Key implementation details:**

1. **Type checking first:** Verify operand types match
2. **Use helper functions:** Reduce duplication for binary operations
3. **Document edge cases:** Comment on overflow/underflow behavior
4. **Error messages:** Use `verror.NewScriptError` for user errors

**Example implementation for shifts:**

```go
func BitShl(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
    if len(args) != 2 {
        return value.NewNoneVal(), arityError("bit.shl", 2, len(args))
    }
    
    // Second arg must be integer (shift count)
    countType := args[1].GetType()
    if countType != value.TypeInteger {
        return value.NewNoneVal(), typeError("bit.shl", "integer!", args[1])
    }
    count, _ := value.AsIntValue(args[1])
    
    // First arg can be integer or binary
    switch args[0].GetType() {
    case value.TypeInteger:
        val, _ := value.AsIntValue(args[0])
        if count < 0 {
            return value.NewNoneVal(), verror.NewScriptError(
                verror.ErrIDOutOfBounds,
                [3]string{"bit.shl", "shift count must be non-negative", ""},
            )
        }
        return value.NewIntVal(val << uint(count)), nil
        
    case value.TypeBinary:
        bin, _ := value.AsBinaryValue(args[0])
        return binaryShl(bin, count), nil
        
    default:
        return value.NewNoneVal(), typeError("bit.shl", "integer! binary!", args[0])
    }
}

func binaryShl(b *value.BinaryValue, count int64) core.Value {
    data := b.Bytes()
    if count <= 0 || len(data) == 0 {
        return value.NewBinaryValue(data)
    }
    
    bitCount := int(count)
    byteShift := bitCount / 8
    bitShift := bitCount % 8
    
    result := make([]byte, len(data))
    
    // Handle byte-level shift
    if byteShift >= len(data) {
        // Complete overflow
        return value.NewBinaryValue(result)  // All zeros
    }
    
    // Shift bytes
    for i := 0; i < len(data)-byteShift; i++ {
        result[i] = data[i+byteShift]
    }
    
    // Shift bits within bytes
    if bitShift > 0 {
        carry := byte(0)
        for i := len(result) - 1; i >= 0; i-- {
            newCarry := result[i] >> (8 - bitShift)
            result[i] = (result[i] << bitShift) | carry
            carry = newCarry
        }
    }
    
    return value.NewBinaryValue(result)
}
```

**Validation:**
- All bitwise tests PASS
- Edge cases handled correctly
- Error messages clear and helpful

### Step 6: Documentation and Examples

**File:** Update `internal/native/register_bitwise.go` with comprehensive NativeDoc

Ensure each function has:
- Clear summary
- Detailed description including type-specific behavior
- Parameter documentation
- Return value description
- Multiple examples showing both integer and binary usage
- Cross-references to related functions
- Tags for categorization

**Example documentation:**

```go
&NativeDoc{
    Category: "Bitwise",
    Summary:  "Shifts bits left",
    Description: `Shifts bits to the left by the specified count.

For integers: Standard left shift using Go's << operator.
For binaries: Shifts all bytes left within the series boundaries.
  - Overflow beyond the series length is lost.
  - No new bytes are created.
  - Result has same length as input.

Left shift by N positions is equivalent to multiplying by 2^N for integers.`,
    Parameters: []ParamDoc{
        {Name: "value", Type: "integer! binary!", Description: "Value to shift", Optional: false},
        {Name: "count", Type: "integer!", Description: "Number of bit positions to shift (must be non-negative)", Optional: false},
    },
    Returns:  "Same type as input value",
    Examples: []string{
        "1 << 3  ; => 8 (integer shift)",
        "5 bit.shl 2  ; => 20 (named form)",
        "#{01} << 2  ; => #{04} (binary shift)",
        "#{80} << 1  ; => #{00} (overflow lost)",
        "#{01 00} << 8  ; => #{00 01} (multi-byte shift)",
    },
    SeeAlso:  []string{"bit.shr", ">>", "*"},
    Tags:     []string{"bitwise", "shift", "binary"},
}
```

**Validation:**
- All functions documented
- Examples tested and verified
- Help system integration works

### Step 7: Final Testing and Validation

**Manual testing in REPL:**

```viro
; Integer operations
>> 2 bit.and 3
2
>> 6 bit.or 3
7
>> 5 bit.xor 3
6
>> bit.not 0
-1

; Shift operations
>> 1 << 3
8
>> 16 >> 2
4

; Binary operations
>> bit.and #{FF 00} #{0F F0}
#{0F00}
>> bit.and #{FF FF} #{FF}
#{00FF}  ; Left byte zeroed (AND with 0)
>> #{80} << 1
#{00}

; Bit manipulation
>> bit.on 0 3
8
>> bit.off 15 2
11

; Bit counting
>> bit.count 15
4
>> bit.count #{FF 00 0F}
12

; Composition
>> 2 << 3 + 1
17
>> bit.on 0 0
1
>> bit.on (bit.on 0 0) 2
5
```

**Automated test suite:**

```bash
go test ./test/contract/bitwise_test.go -v
```

Expected: All tests PASS

**Integration tests:**

Check that bitwise operations work correctly in:
- Function bodies
- Conditional expressions
- Loop conditions
- Path expressions
- Object field values

**Validation checklist:**
- [ ] All bitwise operations work for integers
- [ ] All bitwise operations work for binaries
- [ ] Shift operators `<<` and `>>` work as infix
- [ ] `bit.on` and `bit.off` work correctly
- [ ] `bit.count` works for both types
- [ ] Error messages are clear
- [ ] Type checking prevents invalid operations
- [ ] Left-to-right evaluation works correctly
- [ ] Documentation is complete
- [ ] No regressions in existing tests

## Integration Points

### 1. Bootstrap Integration

**Location:** `internal/bootstrap/bootstrap.go` - `NewEvaluatorWithNatives()`

**Change:** Add `native.RegisterBitwiseNatives(rootFrame)` call

**Impact:** Minimal - follows existing pattern for native registration

### 2. Object System Integration

**Location:** Uses existing `value.NewObject()` and frame system

**Integration:** Create `bit` object with owned frame, bind to root frame

**Impact:** None - leverages existing object infrastructure

### 3. Infix Function Support

**Location:** Parser already supports infix via function property

**Integration:** Set `infix: true` in `NewNativeFunction()` calls

**Impact:** None - existing mechanism, no parser changes

### 4. Type System Integration

**Location:** Uses existing type checking via `GetType()` and type assertion helpers

**Integration:** Type dispatch in each function implementation

**Impact:** None - follows established patterns

### 5. Binary Series Integration

**Location:** `internal/value/binary.go` - `BinaryValue` type

**Integration:** 
- Access bytes via `Bytes()` method
- Create results with `NewBinaryValue([]byte)`
- Preserve series semantics (no auto-growth)

**Impact:** None - read-only access to binary data

### 6. Error Handling Integration

**Location:** Uses `verror` package for error reporting

**Integration:**
- Type errors: `typeError()` helper
- Arity errors: `arityError()` helper
- Script errors: `verror.NewScriptError()` with appropriate error IDs

**Impact:** None - uses existing error infrastructure

## Testing Strategy

### Test Organization

**Primary file:** `test/contract/bitwise_test.go`

**Structure:** Follow existing contract test patterns
- Table-driven tests
- Clear test names
- Separate test functions per operation category
- Integration tests for composition

### Test Categories

1. **Integer Operations** (AND, OR, XOR, NOT)
   - Basic operations
   - Edge cases (zero, negative, all bits set)
   - Infix vs prefix notation

2. **Binary Operations** (AND, OR, XOR, NOT)
   - Same-length operands
   - Different-length operands
   - Edge cases (empty, single byte, many bytes)

3. **Shift Operations** (<<, >>)
   - Integer shifts (positive, negative, zero)
   - Binary shifts (overflow, underflow, multi-byte)
   - Both infix and named forms

4. **Bit Manipulation** (bit.on, bit.off)
   - Set/clear various bit positions
   - Operations on zero
   - Operations on values with bits already set/clear

5. **Bit Counting** (bit.count)
   - Integer counting (various values, negative)
   - Binary counting (multiple bytes)

6. **Error Cases**
   - Type mismatches
   - Mixed types
   - Wrong types for integer-only operations
   - Invalid shift counts

7. **Integration and Composition**
   - Left-to-right evaluation
   - Parentheses control
   - Chained operations
   - Use in expressions

### Coverage Goals

- **Line coverage:** >95% of bitwise.go
- **Branch coverage:** 100% of type dispatch paths
- **Edge cases:** All documented edge cases tested
- **Error paths:** All error conditions verified

### Test Execution

```bash
# Run bitwise tests only
go test ./test/contract/bitwise_test.go -v

# Run all contract tests (verify no regressions)
go test ./test/contract/... -v

# Run full test suite
go test ./... -v

# Check coverage
go test -coverprofile=coverage.out ./internal/native/
go tool cover -html=coverage.out
```

## Potential Challenges and Mitigations

### Challenge 1: Binary Shift Implementation Complexity

**Issue:** Bit shifts across byte boundaries are non-trivial

**Solution:**
- Break into byte-shift and bit-shift components
- Handle byte-level shift first (simple array manipulation)
- Handle bit-level shift with carry propagation
- Test incrementally with simple cases first

**Mitigation:**
- Well-tested helper functions
- Clear comments on algorithm
- Edge case coverage

### Challenge 2: Different-Length Binary Operations

**Issue:** Defining behavior for operations on different-length binary values

**Decision:** 
- Align bytes from the right (LSB first)
- **AND:** Zero remaining bytes (shorter is zero-padded)
- **OR/XOR:** Copy remaining bytes from longer (X OR/XOR 0 = X)

**Rationale:**
- Right-alignment matches numeric interpretation (LSB first)
- Different operators require different padding semantics:
  - AND: X AND 0 = 0 (must zero)
  - OR: X OR 0 = X (copy from longer)
  - XOR: X XOR 0 = X (copy from longer)
- Result length always matches longer operand
- No actual allocation for padding (virtual zero-padding)

**Mitigation:**
- Clear documentation of right-alignment behavior
- Comprehensive tests for different-length cases
- Examples showing byte-by-byte comparison from right

### Challenge 3: Negative Number Shifts

**Issue:** Right shift semantics for negative integers

**Decision:** Use arithmetic shift (sign-extending), not logical shift

**Rationale:**
- Matches Go's default behavior
- Consistent with two's complement representation
- Preserves sign of value

**Example:**
```viro
>> -16 >> 2
-4  ; Not 4611686018427387900 (logical shift would give huge positive)
```

**Mitigation:**
- Document arithmetic shift behavior
- Test negative shift cases
- Examples showing sign preservation

### Challenge 4: Binary Shift Overflow/Underflow

**Issue:** What happens when bits shift beyond series boundaries?

**Decision:** Lost, no wraparound, no new bytes created

**Rationale:**
- Consistent with series semantics (no auto-growth)
- Predictable memory usage
- Simple mental model

**Example:**
```viro
>> #{80} << 1
#{00}  ; High bit lost, not #{00 01}
```

**Mitigation:**
- Clear documentation
- Tests specifically for overflow/underflow
- Examples in docs

### Challenge 5: Bit.count on Negative Integers

**Issue:** How many bits are "set" in a negative number?

**Decision:** Count all set bits in two's complement representation (64 bits)

**Implementation:** Use `bits.OnesCount64(uint64(value))`

**Example:**
```viro
>> bit.count -1
64  ; All bits set in int64
>> bit.count -2
63  ; All but bit 0
```

**Mitigation:**
- Document that int64 representation is used
- Test various negative values
- Examples showing behavior

### Challenge 6: Parser Token Handling for `<<` and `>>`

**Issue:** Will tokenizer correctly handle these operators?

**Answer:** Yes, they tokenize as literals (words)

**Verification:**
- Tokenizer uses `readLiteral()` which accepts `<` and `>` in symbols
- `<<` and `>>` are valid literal tokens
- Infix mechanism handles them via function property

**Mitigation:**
- Manual REPL testing
- Integration tests with these operators

### Challenge 7: Maintaining Left-to-Right Evaluation

**Issue:** Ensuring bitwise operators follow Viro's evaluation model

**Solution:** No special handling needed

**Rationale:**
- Infix is a function property, not a parser concern
- All infix operators evaluate left-to-right automatically
- No operator precedence in Viro

**Example:**
```viro
>> 2 << 3 + 1
17  ; (2 << 3) + 1 = 16 + 1, not 2 << 4 = 32
```

**Mitigation:**
- Test left-to-right evaluation explicitly
- Document in operator docs
- Examples showing composition

## Viro Guidelines Reference

### Coding Standards Followed

1. **No comments in code** - All documentation in this plan and in NativeDoc
2. **Constructor functions** - Use `value.NewIntVal()`, `value.NewBinaryValue()`, etc.
3. **Error handling** - Use `verror.NewScriptError()` with category/ID/args
4. **Table-driven tests** - All tests follow `[]struct{name, input, expected, wantErr}` pattern
5. **TDD approach** - Write tests first in Step 4, implement in Step 5

### Viro Naming Conventions

- Native function names: lowercase with dots (`bit.and`, `bit.or`)
- Operators: symbols (`<<`, `>>`)
- Object names: lowercase (`bit`)
- Error IDs: kebab-case (existing IDs used: `type-mismatch`, `out-of-bounds`)

### Architecture Alignment

- **Value system:** Returns `core.Value` from all natives
- **Object system:** Uses frame-based object storage
- **Error categories:** Uses `ErrScript` for user errors, `ErrInternal` for bugs
- **Frame system:** Creates object frame with `frame.FrameObject` type
- **Type dispatch:** Uses `GetType()` and type assertion helpers
- **Infix support:** Uses function property, no parser changes

### Documentation Alignment

- **NativeDoc structure:** Category, Summary, Description, Parameters, Returns, Examples, SeeAlso, Tags
- **Examples:** Show both simple and complex usage
- **Cross-references:** Link related functions
- **Type annotations:** Use Viro type syntax (`integer!`, `binary!`)

## Summary

This plan implements comprehensive bitwise operations for Viro through a well-organized `bit` object namespace containing nine functions:

**Bitwise Logic:** `bit.and`, `bit.or`, `bit.xor`, `bit.not`
**Shifts:** `bit.shl`, `bit.shr` (also `<<`, `>>` globally)
**Bit Manipulation:** `bit.on`, `bit.off`
**Utility:** `bit.count`

**Key architectural decisions:**
1. Object-based organization for namespace clarity
2. Support both `integer!` and `binary!` where semantically appropriate
3. Type preservation (operations return same type as input)
4. No mixed-type operations
5. Arithmetic right shift for integers (sign-preserving)
6. Overflow/underflow lost for binary shifts (no auto-growth)
7. Different-length binary operations: operate on overlap, copy remainder

**Implementation approach:**
1. TDD: Write comprehensive tests first
2. Minimal integration: Leverage existing infrastructure
3. No parser changes: Use existing infix mechanism
4. Clear error messages: Help users understand type requirements
5. Comprehensive documentation: Examples for both types

**Validation:**
- All operations tested for both `integer!` and `binary!`
- Error cases covered
- Integration with Viro's evaluation model verified
- No regressions in existing functionality
- Complete documentation with examples

This feature provides powerful bit manipulation capabilities while maintaining Viro's design principles and evaluation semantics.
