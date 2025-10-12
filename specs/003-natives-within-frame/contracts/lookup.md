# Contract: Unified Word Lookup Resolution

**Feature**: 003-natives-within-frame
**Component**: `internal/eval.Evaluator.Lookup()`
**Date**: 2025-10-12

## Purpose

Defines the contract for unified word resolution that traverses the frame chain from innermost to outermost scope, eliminating special-case native registry lookups. This contract ensures consistent lexical scoping for all word types.

---

## Function Signature

```go
func (e *Evaluator) Lookup(symbol string) (value.Value, bool)
```

**Package**: `internal/eval`
**Receiver**: `*Evaluator`
**Visibility**: Public (exported)
**Returns**:
- `value.Value` - The bound value if found, `NoneVal()` if not found
- `bool` - `true` if symbol found in frame chain, `false` otherwise

---

## Contract

### Pre-conditions

**Pre-1**: Evaluator is fully initialized (via `NewEvaluator()`)

**Pre-2**: `symbol` is a non-empty string

**Pre-3**: Evaluator has at least one frame (root frame)

### Post-conditions

**Post-1**: If symbol found in any frame from current to root:
- Returns `(value, true)` where value is the bound value
- Value is from the **innermost** frame containing the symbol (shadowing)

**Post-2**: If symbol not found in any frame:
- Returns `(NoneVal(), false)`

**Post-3**: Frame chain traversal order preserved:
1. Current frame (top of `Frames` stack)
2. Current frame's parent
3. Parent's parent
4. ... (continue until root)
5. Root frame (Index = 0, Parent = -1)

**Post-4**: No side effects (pure query operation)

### Invariants

**Inv-1**: Lookup result is deterministic - same symbol in same frame context always returns same value

**Inv-2**: Frame chain traversal terminates (root frame has Parent = -1)

**Inv-3**: **No special-case logic for natives** - all words follow same resolution path

**Inv-4**: Shadowing semantics: inner scope bindings hide outer scope bindings with same name

### Failure Modes

**FM-1**: **Invalid frame chain** (broken parent link)
- **Trigger**: Parent index references non-existent or invalid frame
- **Behavior**: Returns `(NoneVal(), false)` or panics (defensive check)
- **Recoverable**: No (internal error)

**FM-2**: **Empty frame stack**
- **Trigger**: `len(e.Frames) == 0` (should never happen)
- **Behavior**: Returns `(NoneVal(), false)`
- **Recoverable**: No (evaluator invariant violation)

---

## Implementation Details

### Algorithm (Frame Chain Traversal)

```go
func (e *Evaluator) Lookup(symbol string) (value.Value, bool) {
    // Start from current frame
    frame := e.currentFrame()

    // Traverse chain until symbol found or root exhausted
    for frame != nil {
        // Check this frame
        if val, ok := frame.Get(symbol); ok {
            return val, true  // Found in this frame
        }

        // Move to parent
        if frame.Parent == -1 {
            break  // Reached root with no parent
        }
        frame = e.getFrameByIndex(frame.Parent)
    }

    // Not found in any frame
    return value.NoneVal(), false
}
```

**Complexity**:
- **Time**: O(D × K) where D = chain depth, K = average frame size
  - Typical: D ≤ 5, K ≤ 20 → ~100 comparisons worst case
- **Space**: O(1) - no allocations

---

### Resolution Examples

#### Example 1: Native Function Lookup

**Setup**:
```
Frame Chain:
  [2] Function Frame: words=["x", "y"], parent=1
  [1] Closure Frame:  words=["z"], parent=0
  [0] Root Frame:     words=["+", "-", "print", ...], parent=-1
```

**Lookup**: `e.Lookup("+")`

**Trace**:
1. Check Frame[2]: "+" not in ["x", "y"] → continue
2. Check Frame[1]: "+" not in ["z"] → continue
3. Check Frame[0]: "+" in root frame → **FOUND**
4. Return `(FuncVal(nativePlus), true)`

**Result**: Native function found via root frame (no special path)

---

#### Example 2: Local Variable Shadowing Native

**Setup**:
```
Frame Chain:
  [2] Function Frame: words=["print", "x"], parent=1  // User-defined "print"
  [1] Closure Frame:  words=[], parent=0
  [0] Root Frame:     words=["print", "+", ...], parent=-1  // Native "print"
```

**Lookup**: `e.Lookup("print")`

**Trace**:
1. Check Frame[2]: "print" in ["print", "x"] → **FOUND**
2. Return `(user-defined-print-value, true)`

**Result**: User-defined binding shadows native (innermost wins)

---

#### Example 3: Word Not Found

**Setup**:
```
Frame Chain:
  [2] Function Frame: words=["x"], parent=1
  [1] Closure Frame:  words=["y"], parent=0
  [0] Root Frame:     words=["+", "-", "print"], parent=-1
```

**Lookup**: `e.Lookup("undefined")`

**Trace**:
1. Check Frame[2]: "undefined" not in ["x"] → continue
2. Check Frame[1]: "undefined" not in ["y"] → continue
3. Check Frame[0]: "undefined" not in ["+", "-", "print"] → continue
4. Frame[0].Parent == -1 → exit loop
5. Return `(NoneVal(), false)`

**Result**: Word not found → evaluator will raise `no-value` error

---

## Usage in Evaluator

### Current Usage (Registry + Frame)

**BEFORE** (`evaluator.go:522-525`):
```go
// Check native registry first
if nativeFn, found := native.Lookup(wordStr); found {
    return e.invokeFunction(nativeFn, seq, idx, lastResult)
}

// Check user-defined functions
if resolved, found := e.Lookup(wordStr); found && resolved.Type == value.TypeFunction {
    fn, _ := resolved.AsFunction()
    return e.invokeFunction(fn, seq, idx, lastResult)
}
```

**Issues**:
- Two separate lookups (registry + frame)
- Native always wins (no shadowing possible)
- Duplicate code paths

---

### Unified Usage (Frame Only)

**AFTER** (proposed):
```go
// Unified lookup (natives and user-defined)
if resolved, found := e.Lookup(wordStr); found {
    if resolved.Type == value.TypeFunction {
        fn, _ := resolved.AsFunction()
        return e.invokeFunction(fn, seq, idx, lastResult)
    }
    // Word is not a function, evaluate normally
    return resolved, nil
}

// Not found → error
return value.NoneVal(), verror.NewScriptError(verror.ErrIDNoValue, [3]string{wordStr, "", ""})
```

**Benefits**:
- ✅ Single lookup path
- ✅ Shadowing supported (frame chain ordering)
- ✅ Simpler code (no special cases)
- ✅ Eliminates registry overhead

---

## Modified Call Sites

**Location**: `internal/eval/evaluator.go`

**Affected Functions**:

1. **`evalWord()`** (line ~745)
   - **REMOVE**: `native.Lookup()` call
   - **KEEP**: `e.Lookup()` call
   - **CHANGE**: Remove special-case for natives returning word

2. **`evalGetWord()`** (line ~828)
   - **REMOVE**: `native.Lookup()` check
   - **KEEP**: `e.Lookup()` call

3. **`evaluateWithFunctionCall()`** (line ~512)
   - **REMOVE**: `native.Lookup()` call (line 522-525)
   - **MODIFY**: Use `e.Lookup()` for all words
   - **LOGIC**: If lookup returns function → invoke, else evaluate normally

**Removed Code**:
```go
// DELETE THIS:
if nativeFn, found := native.Lookup(wordStr); found {
    return e.invokeFunction(nativeFn, seq, idx, lastResult)
}
```

**Replacement**:
```go
// Unified lookup already exists below:
if resolved, found := e.Lookup(wordStr); found && resolved.Type == value.TypeFunction {
    fn, _ := resolved.AsFunction()
    return e.invokeFunction(fn, seq, idx, lastResult)
}
```

---

## Testing Strategy

### Test 1: Native Lookup via Frame Chain

```go
func TestLookupFindsNativeInRootFrame(t *testing.T) {
    e := eval.NewEvaluator()

    // Lookup native from top level
    plusVal, found := e.Lookup("+")

    require.True(t, found, "+ should be found")
    require.Equal(t, value.TypeFunction, plusVal.Type)

    plusFn, ok := plusVal.AsFunction()
    require.True(t, ok)
    require.Equal(t, "+", plusFn.Name)
    require.Equal(t, value.FuncNative, plusFn.Type)
}
```

### Test 2: Shadowing Native with Local Binding

```go
func TestLookupShadowsNativeWithLocalBinding(t *testing.T) {
    e := eval.NewEvaluator()

    // Create local frame with "print" variable
    localFrame := frame.NewFrame(frame.FrameFunctionArgs, 0)  // Parent = root
    localFrame.Bind("print", value.IntVal(42))
    e.PushFrameContext(localFrame)

    // Lookup "print" from local scope
    val, found := e.Lookup("print")

    require.True(t, found)
    require.Equal(t, value.TypeInteger, val.Type, "Should find local, not native")
    require.Equal(t, int64(42), val.Payload)
}
```

### Test 3: Multi-Level Shadowing

```go
func TestLookupMultiLevelShadowing(t *testing.T) {
    e := eval.NewEvaluator()

    // Level 1: Shadow native "+" with string
    frame1 := frame.NewFrame(frame.FrameFunctionArgs, 0)
    frame1.Bind("+", value.StrVal("level1"))
    e.PushFrameContext(frame1)

    val1, _ := e.Lookup("+")
    require.Equal(t, "level1", val1.AsString())

    // Level 2: Shadow frame1's "+" with integer
    frame2 := frame.NewFrame(frame.FrameFunctionArgs, 1)
    frame2.Bind("+", value.IntVal(99))
    e.PushFrameContext(frame2)

    val2, _ := e.Lookup("+")
    require.Equal(t, int64(99), val2.Payload)

    // Pop frame2 → frame1's "+" visible again
    e.PopFrameContext()
    val3, _ := e.Lookup("+")
    require.Equal(t, "level1", val3.AsString())
}
```

### Test 4: Word Not Found

```go
func TestLookupReturnsNotFoundForUndefinedWord(t *testing.T) {
    e := eval.NewEvaluator()

    val, found := e.Lookup("nonexistent-word-xyz")

    require.False(t, found, "Undefined word should not be found")
    require.Equal(t, value.TypeNone, val.Type)
}
```

### Test 5: Closure Capture Semantics

```go
func TestLookupClosureCapturesNativeValue(t *testing.T) {
    e := eval.NewEvaluator()

    // Create closure capturing "+" native
    closureCode := parse("fn [] [:+]")  // Get-word captures value
    closureFn, _ := e.Do_Blk(closureCode)

    // Shadow "+" native in global scope
    rootFrame := e.GetFrameByIndex(0)
    rootFrame.Bind("+", value.StrVal("shadowed"))

    // Invoke closure → should see captured native, not shadow
    result, _ := closureFn.AsFunction().Invoke(e, []value.Value{})

    fn, ok := result.AsFunction()
    require.True(t, ok, "Closure should return original native function")
    require.Equal(t, "+", fn.Name)
    require.Equal(t, value.FuncNative, fn.Type, "Should be native, not shadowed string")
}
```

---

## Performance Comparison

### Benchmark Setup

```go
func BenchmarkLookupNative(b *testing.B) {
    e := eval.NewEvaluator()
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, _ = e.Lookup("+")
    }
}

func BenchmarkLookupUserDefined(b *testing.B) {
    e := eval.NewEvaluator()
    rootFrame := e.GetFrameByIndex(0)
    rootFrame.Bind("myvar", value.IntVal(42))
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, _ = e.Lookup("myvar")
    }
}

func BenchmarkLookupNested(b *testing.B) {
    e := eval.NewEvaluator()
    // Create 3-level deep frame chain
    frame1 := frame.NewFrame(frame.FrameFunctionArgs, 0)
    frame1.Bind("x", value.IntVal(1))
    e.PushFrameContext(frame1)

    frame2 := frame.NewFrame(frame.FrameFunctionArgs, 1)
    frame2.Bind("y", value.IntVal(2))
    e.PushFrameContext(frame2)

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, _ = e.Lookup("x")  // Requires parent traversal
    }
}
```

**Expected Results**:
- Native lookup (root): ~50 ns (70 comparisons worst case)
- Local variable: ~10 ns (1-5 comparisons)
- Nested lookup: ~30 ns (2-3 frame checks)

---

## Dependencies

**Depends On**:
- `frame.Frame.Get()` - Single-frame lookup
- `Evaluator.currentFrame()` - Current scope access
- `Evaluator.getFrameByIndex()` - Parent frame access

**Depended On By**:
- `evalWord()` - Word evaluation
- `evalGetWord()` - Get-word evaluation
- `evaluateWithFunctionCall()` - Function dispatch
- `evalSetWord()` - Set-word validation (check if word exists)

---

## Backward Compatibility

**Breaking Changes**: None for existing functionality

**New Behavior**:
- ✅ Native shadowing now possible (previously prevented)
- ✅ Refinement parameters can use native names without collision

**Migration**:
- Existing code without shadowing → **identical behavior**
- Code that previously failed due to name collision → **now works**

---

## References

- [spec.md](../spec.md) - FR-003, FR-004, FR-006, User Story 3
- [data-model.md](../data-model.md) - Word Lookup Resolution entity
- [research.md](../research.md) - Q2: Performance baseline for word lookups

---

**Contract Status**: ✅ Complete | **Version**: 1.0 | **Approved**: Pending implementation
