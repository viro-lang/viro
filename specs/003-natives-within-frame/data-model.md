# Data Model: Natives Within Frame

**Feature**: 003-natives-within-frame
**Date**: 2025-10-12
**Status**: Design
**Related**: [spec.md](./spec.md) | [plan.md](./plan.md) | [research.md](./research.md)

## Overview

This document defines the data structures, relationships, and lifecycle for storing native functions in the root frame instead of a global registry. The model unifies word resolution to use standard lexical scoping for all function types.

---

## Entities

### 1. Root Frame

**Definition**: The outermost lexical scope in the evaluator's frame chain, serving as the global environment where native functions are initially registered.

**Type**: `*frame.Frame`

**Attributes**:
| Attribute | Type | Description | Constraints |
|-----------|------|-------------|-------------|
| `Type` | `frame.FrameType` | Frame category | Must be `FrameClosure` |
| `Words` | `[]string` | Symbol names bound in this frame | Contains 70+ native function names |
| `Values` | `[]value.Value` | Bound values parallel to Words | Each Value wraps a `*FunctionValue` |
| `Parent` | `int` | Index of parent frame | Always `-1` (root has no parent) |
| `Index` | `int` | Position in frameStore | Always `0` (first frame) |
| `Name` | `string` | Diagnostic identifier | Always `"(top level)"` |
| `Manifest` | `*ObjectManifest` | Optional field metadata | Always `nil` for root frame |

**Lifecycle**:
1. **Creation**: During `NewEvaluator()` construction
   - Pre-allocated with capacity 80 via `NewFrameWithCapacity(FrameClosure, -1, 80)`
   - Assigned `Index = 0` and `Name = "(top level)"`

2. **Population**: Immediately after creation
   - 6 category registration functions called in sequence
   - Each calls `rootFrame.Bind(name, value.FuncVal(fn))`
   - Panic if any registration fails (duplicate name, nil function)

3. **Runtime**: Persists for entire evaluator lifetime
   - Never removed from frameStore
   - Marked as captured: `captured[0] = true`
   - Accessible via frame chain traversal from any inner scope
   - **Mutable**: User code can rebind/overwrite natives (per clarification)

**Relationships**:
- **Contains**: 70+ native function bindings (name → FunctionValue)
- **Parent of**: All top-level user-defined frames
- **Owned by**: Single Evaluator instance

**Invariants**:
- ✅ Must contain all native functions after `NewEvaluator()` returns
- ✅ Must be first frame in frameStore (Index = 0)
- ✅ Must have Parent = -1 (no parent)
- ✅ Must be marked as captured (never GC'd)

---

### 2. Native Function Binding

**Definition**: A word-to-value mapping in the root frame representing a built-in function implemented in Go.

**Structure**:
```
Word (string):    "+"  // Native function name
Value (value.Value):
  Type:    TypeFunction
  Payload: *FunctionValue {
    Name:   "+"
    Type:   FuncNative
    Params: [...ParamSpec]
    Infix:  true
    Doc:    &NativeDoc{...}
    Native: func(args, refs, eval) -> (Value, error)
  }
```

**Attributes**:
| Component | Type | Description | Source |
|-----------|------|-------------|--------|
| Word symbol | `string` | Function name (e.g., "+", "print", "fn") | Defined in registration code |
| Value wrapper | `value.Value` | Container for FunctionValue | Created by `value.FuncVal(fn)` |
| Function metadata | `*value.FunctionValue` | Function signature and implementation | Created during registration |

**Lifecycle**:
1. **Registration**: During evaluator construction
   - `FunctionValue` created with `value.NewNativeFunction(...)`
   - Optional fields set (`Infix`, `Doc`)
   - Wrapped in Value via `value.FuncVal(fn)`
   - Bound to root frame via `rootFrame.Bind(name, val)`

2. **Lookup**: During word resolution
   - User code references native name (e.g., `+ 3 4`)
   - Evaluator traverses frame chain: current → parent → ... → root
   - Root frame checked: `rootFrame.Get("+")` → Returns Value
   - Value unwrapped: `val.AsFunction()` → Returns FunctionValue
   - Function invoked via unified `invokeFunction()` path

3. **Shadowing**: User-defined binding overrides native
   - User defines local function/variable with native name
   - Bound to inner frame (function or closure frame)
   - Frame chain resolution finds inner binding first
   - Native remains in root frame, accessible if inner binding removed

**Relationships**:
- **Stored in**: Root Frame (Words/Values arrays)
- **References**: FunctionValue pointer (shared, not copied)
- **Shadowed by**: User-defined bindings in inner frames

**Invariants**:
- ✅ All native names unique within root frame (no duplicates)
- ✅ All Values must have Type = TypeFunction
- ✅ All FunctionValues must have Type = FuncNative
- ✅ Function metadata (Doc, Params) preserved after storage

---

### 3. Native Function Category

**Definition**: Logical grouping of related native functions for registration organization.

**Categories** (from research.md D1):
| Category | File | Functions | Examples |
|----------|------|-----------|----------|
| Math | `register_math.go` | 20 | `+`, `-`, `*`, `/`, `<`, `>`, `=`, `and`, `or`, `not`, `abs`, `sqrt` |
| Series | `register_series.go` | 5 | `first`, `last`, `append`, `insert`, `length?` |
| Data | `register_data.go` | 8 | `set`, `get`, `type?`, `object`, `clone` |
| I/O | `register_io.go` | 10 | `print`, `input`, `open`, `read`, `write`, `close` |
| Control | `register_control.go` | 5 | `if`, `when`, `loop`, `while`, `fn` |
| Help | `register_help.go` | 11 | `help`, `trace`, `debug`, `reflect`, `info` |

**Registration Function Signature**:
```go
func RegisterMathNatives(rootFrame *frame.Frame)
func RegisterSeriesNatives(rootFrame *frame.Frame)
func RegisterDataNatives(rootFrame *frame.Frame)
func RegisterIONatives(rootFrame *frame.Frame)
func RegisterControlNatives(rootFrame *frame.Frame)
func RegisterHelpNatives(rootFrame *frame.Frame)
```

**Lifecycle**:
- **Invocation**: Called sequentially during `NewEvaluator()`
- **Side Effect**: Populates root frame with category's natives
- **Error Handling**: Panics if registration fails

**Relationships**:
- **Calls**: `rootFrame.Bind()` for each native in category
- **Invoked by**: `NewEvaluator()` construction function

---

### 4. Word Lookup Resolution

**Definition**: The process of resolving a word symbol to its bound value by traversing the frame chain.

**Algorithm** (unified for all words):
```
function Lookup(symbol string) -> (Value, bool):
  frame := currentFrame
  while frame != nil:
    value, found := frame.Get(symbol)
    if found:
      return value, true
    if frame.Parent == -1:
      break
    frame := getFrameByIndex(frame.Parent)
  return NoneVal(), false
```

**Resolution Order** (innermost to outermost):
1. **Current frame**: Function parameters, local variables
2. **Parent frame**: Captured closure variables
3. **...** (traverse parent chain)
4. **Root frame**: Native functions, global user-defined names

**State Transitions**:
```
Word Symbol (string)
  ↓
[Check Current Frame]
  ↓ found → Return Value
  ↓ not found
[Check Parent Frame]
  ↓ found → Return Value
  ↓ not found
[Continue up chain...]
  ↓
[Check Root Frame]
  ↓ found → Return Value (native or global)
  ↓ not found → ERROR (no-value)
```

**Properties**:
- **Lexical scoping**: Resolution follows frame parent chain
- **Shadowing**: Inner bindings hide outer bindings with same name
- **Deterministic**: Same word always resolves to same binding in same context
- **No special cases**: Natives resolved via same path as user-defined words

**Changes from Current Implementation**:
| Aspect | Before (Registry) | After (Frame) |
|--------|-------------------|---------------|
| Native check | `native.Lookup()` called first | No special check, follows frame chain |
| Resolution order | Registry → Frame chain | Frame chain only |
| Code paths | 2 paths (native vs user) | 1 unified path |
| Shadowing natives | Impossible (registry wins) | Allowed (frame chain wins) |

---

## Relationships Diagram

```
┌─────────────────────────────────────────────────────────────┐
│ Evaluator                                                   │
│ ┌─────────────────────────────────────────────────────────┐ │
│ │ frameStore: []*Frame                                    │ │
│ │   [0] Root Frame (Native Functions)                     │ │
│ │       Words:  ["+", "-", "*", "print", "fn", ...]       │ │
│ │       Values: [FuncVal, FuncVal, FuncVal, ...]          │ │
│ │       Parent: -1                                        │ │
│ │   [1] User Frame 1                                      │ │
│ │       Parent: 0  (points to root)                       │ │
│ │   [2] User Frame 2                                      │ │
│ │       Parent: 1  (points to frame 1)                    │ │
│ └─────────────────────────────────────────────────────────┘ │
│                                                              │
│ Lookup("print") → Traverse: Frame 2 → Frame 1 → Root (found)│
│ Lookup("x")     → Traverse: Frame 2 (found local var)       │
└──────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│ Native Function Binding (in Root Frame)                     │
│ ┌──────────────┐    Wraps    ┌────────────────────────────┐ │
│ │ Value        │ ───────────> │ FunctionValue              │ │
│ │ Type: Func   │             │ Name:   "+"                │ │
│ │ Payload: ptr │             │ Type:   FuncNative         │ │
│ └──────────────┘             │ Params: [left, right]      │ │
│                               │ Infix:  true               │ │
│                               │ Doc:    &NativeDoc{...}    │ │
│                               │ Native: func(...) {...}    │ │
│                               └────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
```

---

## State Machines

### Evaluator Construction State Machine

```
[Uninitialized]
      ↓
  NewEvaluator() called
      ↓
[Creating Root Frame]
  - NewFrameWithCapacity(FrameClosure, -1, 80)
  - Set Name = "(top level)"
  - Set Index = 0
      ↓
[Creating Evaluator Struct]
  - Allocate Stack, Frames, frameStore
  - Add root frame to frameStore[0]
  - Mark root as captured
      ↓
[Registering Natives - Math]
  - RegisterMathNatives(rootFrame)
  - Bind 20 math functions
      ↓ (panic if error)
[Registering Natives - Series]
  - RegisterSeriesNatives(rootFrame)
  - Bind 5 series functions
      ↓ (panic if error)
[Registering Natives - Data]
  - RegisterDataNatives(rootFrame)
  - Bind 8 data functions
      ↓ (panic if error)
[Registering Natives - I/O]
  - RegisterIONatives(rootFrame)
  - Bind 10 I/O functions
      ↓ (panic if error)
[Registering Natives - Control]
  - RegisterControlNatives(rootFrame)
  - Bind 5 control functions
      ↓ (panic if error)
[Registering Natives - Help]
  - RegisterHelpNatives(rootFrame)
  - Bind 11 help/debug functions
      ↓ (panic if error)
[Initialized]
  - Root frame contains 70+ natives
  - Evaluator ready for use
      ↓
  return evaluator
```

**Error States**:
- Panic → Program terminates with message
- No partial initialization (fail-fast)

---

### Word Lookup State Machine

```
[Word Symbol Input] "print"
      ↓
[Get Current Frame]
      ↓
[Search Current Frame]
  - Linear scan Words array
      ↓
  Found? ──yes──> [Return Value] → END
      │ no
      ↓
[Has Parent Frame?]
      │ no (Parent == -1)
      ↓
  [ERROR: no-value] → END
      │ yes
      ↓
[Get Parent Frame]
  - frame = getFrameByIndex(frame.Parent)
      ↓
[Search Parent Frame]
  - Linear scan Words array
      ↓
  Found? ──yes──> [Return Value] → END
      │ no
      ↓
[Repeat up chain...]
      ↓
[Reach Root Frame]
      ↓
[Search Root Frame]
  - Check 70+ native names
      ↓
  Found? ──yes──> [Return Native FunctionValue] → END
      │ no
      ↓
[ERROR: no-value] → END
```

---

## Data Validation Rules

### Root Frame Validation

**Rule RF-1**: Root frame must be first in frameStore
```go
assert(rootFrame.Index == 0)
assert(evaluator.frameStore[0] == rootFrame)
```

**Rule RF-2**: Root frame must have no parent
```go
assert(rootFrame.Parent == -1)
```

**Rule RF-3**: Root frame must contain all native functions
```go
expectedNatives := []string{"+", "-", "*", "/", "print", "fn", ...} // 70+ names
for _, name := range expectedNatives {
    val, found := rootFrame.Get(name)
    assert(found, "native %s not found", name)
    assert(val.Type == TypeFunction, "native %s not a function", name)
}
```

**Rule RF-4**: Native bindings must wrap FunctionValue
```go
val, _ := rootFrame.Get("+")
fn, ok := val.AsFunction()
assert(ok, "native + not unwrappable")
assert(fn.Type == FuncNative, "native + not FuncNative type")
```

---

### Native Registration Validation

**Rule NR-1**: No duplicate native names
```go
seen := make(map[string]bool)
for _, word := range rootFrame.Words {
    assert(!seen[word], "duplicate native: %s", word)
    seen[word] = true
}
```

**Rule NR-2**: All FunctionValues must be non-nil
```go
for i, word := range rootFrame.Words {
    val := rootFrame.Values[i]
    fn, ok := val.AsFunction()
    assert(ok && fn != nil, "native %s has nil function", word)
}
```

**Rule NR-3**: Native function metadata preserved
```go
// Before storage
fnBefore := value.NewNativeFunction("+", params, impl)
fnBefore.Doc = &NativeDoc{...}

// After storage + retrieval
rootFrame.Bind("+", value.FuncVal(fnBefore))
val, _ := rootFrame.Get("+")
fnAfter, _ := val.AsFunction()

assert(fnAfter == fnBefore, "function pointer changed")
assert(fnAfter.Doc == fnBefore.Doc, "metadata lost")
```

---

### Word Lookup Validation

**Rule WL-1**: Shadowing respects lexical scoping
```go
// Setup: Native "print" in root, user "print" in inner frame
rootFrame.Bind("print", value.FuncVal(nativePrint))
innerFrame.Bind("print", value.FuncVal(userPrint))

// Lookup from inner scope
val, _ := evaluator.Lookup("print")  // Starts at innerFrame
fn, _ := val.AsFunction()
assert(fn == userPrint, "inner binding should shadow native")
```

**Rule WL-2**: Natives accessible from nested scopes
```go
// Setup: Native in root, no shadow
rootFrame.Bind("+", value.FuncVal(nativePlus))
// innerFrame has no "+" binding

// Lookup from inner scope
val, _ := evaluator.Lookup("+")  // Traverses to root
fn, _ := val.AsFunction()
assert(fn == nativePlus, "native should be found via traversal")
```

**Rule WL-3**: Closure captures value, not name
```go
// Setup: Closure created when native visible, later shadowed
closureFn := createClosureThatReferences("+")  // Captures nativePlus
innerScope.Bind("+", value.FuncVal(userPlus))   // Shadow native

// Invoke closure
result := closureFn.Invoke()
assert(result uses nativePlus, "closure should see captured value")
```

---

## Performance Characteristics

### Space Complexity

**Root Frame**:
- 70 natives × (16 bytes string + 24 bytes Value) ≈ **2.8 KB**
- Negligible compared to evaluator stack (1024 slots = 24 KB)

**Frame Chain**:
- No change from current implementation
- Each frame: ~40 bytes overhead + bindings

---

### Time Complexity

**Evaluator Construction**:
- Root frame allocation: O(1) with pre-capacity
- Native registration: O(N) where N = 70 natives
  - Each Bind: O(1) amortized (pre-allocated capacity)
  - Total: **~700 µs** (10 µs per native)

**Word Lookup**:
- Best case (local variable): O(1) → O(k) where k = frame size (~5-20)
- Average case (parent frame): O(d × k) where d = depth (typically 2-3)
- Worst case (native/global): O(D × K) where D = max depth, K = root frame size
  - Root frame: 70+ bindings → linear scan acceptable (cache-friendly)

**Comparison to Registry**:
| Operation | Registry (Before) | Frame (After) | Winner |
|-----------|-------------------|---------------|--------|
| Native lookup | O(1) map | O(D × K) traversal | Registry faster |
| User word lookup | O(1) map miss + O(D × k) | O(d × k) | **Frame faster** |
| Construction | O(N) registry populate | O(N) frame populate | Tie |

**Net Effect**: Frame-based lookup is **faster for typical code** because:
1. Most lookups are user-defined words (not natives)
2. Eliminates registry check overhead for every word
3. Root frame lookup is rare (locals/parents resolved first)

---

## Migration Path

### Phase 1: Add Frame Registration (No Breaking Changes)

**Changes**:
- Create 6 `register_*.go` files
- Keep `native.Registry` populated (backward compat)
- Add `RegisterAllNatives(rootFrame)` function
- Call during `NewEvaluator()`

**Verification**:
- Root frame contains all natives
- Registry still contains all natives
- All existing tests pass

---

### Phase 2: Switch Evaluator to Frame Lookups

**Changes**:
- Remove `native.Lookup()` calls from `evaluator.go` (4 locations)
- Use `e.Lookup()` (frame chain traversal) instead
- Update word resolution to unified path

**Verification**:
- All existing tests pass (no behavior change)
- New shadowing tests pass (previously impossible)

---

### Phase 3: Remove Registry

**Changes**:
- Stop populating `native.Registry` in registration functions
- Delete `var Registry` declaration
- Delete `func Lookup()` function

**Verification**:
- No compilation errors (no remaining registry references)
- All tests pass

---

## Testing Strategy

### Unit Tests

**Root Frame Initialization** (`test/contract/evaluator_test.go`):
```go
func TestNewEvaluatorInitializesNatives(t *testing.T) {
    e := eval.NewEvaluator()
    rootFrame := e.GetFrameByIndex(0)

    // Verify all natives present
    nativeNames := []string{"+", "-", "*", "/", "print", "fn", /* ... */}
    for _, name := range nativeNames {
        val, found := rootFrame.Get(name)
        require.True(t, found, "native %s not found", name)
        require.Equal(t, value.TypeFunction, val.Type)
    }
}
```

**Shadowing Behavior** (`test/contract/native_scoping_test.go`):
```go
func TestNativeShadowing(t *testing.T) {
    e := eval.NewEvaluator()

    // Define local "print" function
    code := `print: fn [msg] [msg]; print "test"`
    result, err := e.Do_Blk(parse(code))

    require.NoError(t, err)
    require.Equal(t, "test", result.AsString())
    // Local print shadows native, no stdout output
}
```

---

## Constraints & Assumptions

**Constraints**:
1. Root frame must be index 0 in frameStore
2. Root frame never removed/GC'd (marked captured)
3. Native registration must complete or panic (no partial state)
4. Frame Bind operation must be O(1) amortized

**Assumptions**:
1. Native function count remains ~70 (acceptable for linear scan)
2. User code typically has <20 bindings per frame
3. Frame chain depth typically <5 levels
4. Evaluator construction performance <1ms acceptable

**Non-Goals**:
- ❌ Optimizing root frame lookup (70 items is fine for linear scan)
- ❌ Supporting dynamic native registration after construction
- ❌ Versioning or hot-reloading of natives

---

**Data Model Status**: ✅ Complete | **Next**: Generate contracts/ API specifications
