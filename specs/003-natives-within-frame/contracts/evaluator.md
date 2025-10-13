# Contract: Evaluator Construction with Native Registration

**Feature**: 003-natives-within-frame
**Component**: `internal/eval.NewEvaluator()`
**Date**: 2025-10-12

## Purpose

Defines the contract for evaluator construction, including root frame initialization and native function registration. This contract ensures all native functions are available before any user code executes.

---

## Function Signature

```go
func NewEvaluator() *Evaluator
```

**Package**: `internal/eval`
**Visibility**: Public (exported)
**Returns**: `*Evaluator` - Fully initialized evaluator instance

---

## Contract

### Pre-conditions

**NONE** - Function can be called without setup.

### Post-conditions

**PC-1**: Evaluator instance is fully initialized and ready for use

**PC-2**: Root frame exists at `frameStore[0]` with:
- `Index = 0`
- `Parent = -1`
- `Name = "(top level)"`
- `Type = FrameClosure`
- Pre-allocated capacity ≥ 80 bindings

**PC-3**: Root frame contains all native functions:
- Math natives (20 functions): `+`, `-`, `*`, `/`, `<`, `>`, `<=`, `>=`, `=`, `<>`, `and`, `or`, `not`, `abs`, `min`, `max`, `sqrt`, `power`, `floor`, `ceiling`
- Series natives (5 functions): `first`, `last`, `append`, `insert`, `length?`
- Data natives (8 functions): `set`, `get`, `type?`, `object`, `clone`, `copy`, `find`, `index?`
- I/O natives (10 functions): `print`, `input`, `open`, `read`, `write`, `close`, `exists?`, `delete`, `rename`, `dir`
- Control natives (5 functions): `if`, `when`, `loop`, `while`, `fn`
- Help natives (11 functions): `help`, `trace`, `debug`, `breakpoint`, `step`, `reflect`, `info`, `words`, `values`, `stats`, `gc`

**PC-4**: All native bindings satisfy:
- Value type = `TypeFunction`
- FunctionValue type = `FuncNative`
- Function metadata preserved (Name, Params, Doc, Infix)

**PC-5**: Root frame is marked as captured: `captured[0] = true`

**PC-6**: Stack initialized with capacity 1024

**PC-7**: Call stack initialized with `["(top level)"]`

### Invariants

**I-1**: Root frame never removed or GC'd (captured flag prevents)

**I-2**: No duplicate native names in root frame

**I-3**: Construction completes in <1ms (performance requirement)

### Failure Modes

**FM-1**: **Native registration duplicate name**
- **Trigger**: Two natives registered with same name
- **Behavior**: Panic with message `"native registration failed: duplicate name {name}"`
- **Recoverable**: No (fatal error)

**FM-2**: **Native registration nil function**
- **Trigger**: FunctionValue is nil during registration
- **Behavior**: Panic with message `"native registration failed: {name} is nil"`
- **Recoverable**: No (fatal error)

**FM-3**: **Frame binding failure**
- **Trigger**: (Currently impossible - Bind cannot fail)
- **Behavior**: Would panic if implemented with validation
- **Recoverable**: No (fatal error)

---

## Implementation Details

### Construction Sequence

1. **Create root frame**
   ```go
   global := frame.NewFrameWithCapacity(frame.FrameClosure, -1, 80)
   global.Name = "(top level)"
   global.Index = 0
   ```

2. **Create evaluator struct**
   ```go
   e := &Evaluator{
       Stack:      stack.NewStack(1024),
       Frames:     []*frame.Frame{global},
       frameStore: []*frame.Frame{global},
       captured:   make(map[int]bool),
       callStack:  []string{"(top level)"},
   }
   e.captured[0] = true
   ```

3. **Register natives by category** (sequential, panic-on-error)
   ```go
   native.RegisterMathNatives(global)      // Panics if error
   native.RegisterSeriesNatives(global)    // Panics if error
   native.RegisterDataNatives(global)      // Panics if error
   native.RegisterIONatives(global)        // Panics if error
   native.RegisterControlNatives(global)   // Panics if error
   native.RegisterHelpNatives(global)      // Panics if error
   ```

4. **Return initialized evaluator**
   ```go
   return e
   ```

### Registration Function Contract

Each `Register*Natives(rootFrame *frame.Frame)` function:

**Pre-conditions**:
- `rootFrame` is non-nil
- `rootFrame` is empty or contains non-conflicting names

**Post-conditions**:
- All category natives bound to `rootFrame`
- No duplicate names within category
- All FunctionValues non-nil with correct metadata

**Failure**:
- Panics with descriptive message if validation fails

---

## Usage Examples

### Example 1: Normal Construction

```go
// Create evaluator
e := eval.NewEvaluator()

// Root frame contains natives
rootFrame := e.GetFrameByIndex(0)
plusVal, found := rootFrame.Get("+")
// found == true
// plusVal.Type == TypeFunction

plusFn, _ := plusVal.AsFunction()
// plusFn.Name == "+"
// plusFn.Type == FuncNative
```

### Example 2: Accessing Native from User Code

```go
e := eval.NewEvaluator()

// User code references native
code := parse("+ 3 4")
result, err := e.Do_Blk(code)
// result == IntVal(7)
// err == nil
```

### Example 3: Construction Failure (Duplicate Native)

```go
// Hypothetical: if two registration functions bind same name
func RegisterMathNatives(rootFrame *frame.Frame) {
    rootFrame.Bind("+", value.FuncVal(plusFn))
    // ...
}

func RegisterSeriesNatives(rootFrame *frame.Frame) {
    rootFrame.Bind("+", value.FuncVal(seriesPlusFn))  // DUPLICATE!
    // Panics: "native registration failed: duplicate name +"
}

e := eval.NewEvaluator()  // Panics during construction
```

---

## Testing Strategy

### Test 1: Root Frame Initialization

```go
func TestNewEvaluatorInitializesRootFrame(t *testing.T) {
    e := eval.NewEvaluator()

    // Verify root frame exists
    rootFrame := e.GetFrameByIndex(0)
    require.NotNil(t, rootFrame)
    require.Equal(t, 0, rootFrame.Index)
    require.Equal(t, -1, rootFrame.Parent)
    require.Equal(t, "(top level)", rootFrame.Name)
}
```

### Test 2: All Natives Registered

```go
func TestNewEvaluatorRegistersAllNatives(t *testing.T) {
    e := eval.NewEvaluator()
    rootFrame := e.GetFrameByIndex(0)

    nativeNames := []string{
        "+", "-", "*", "/",          // Math
        "print", "input",             // I/O
        "if", "when", "loop", "fn",  // Control
        "first", "last", "append",   // Series
        "help", "trace", "debug",    // Help
        // ... (all 70+ natives)
    }

    for _, name := range nativeNames {
        val, found := rootFrame.Get(name)
        require.True(t, found, "native %s not found", name)
        require.Equal(t, value.TypeFunction, val.Type, "native %s wrong type", name)

        fn, ok := val.AsFunction()
        require.True(t, ok, "native %s not unwrappable", name)
        require.Equal(t, value.FuncNative, fn.Type, "native %s not FuncNative", name)
    }
}
```

### Test 3: Native Metadata Preserved

```go
func TestNewEvaluatorPreservesNativeMetadata(t *testing.T) {
    e := eval.NewEvaluator()
    rootFrame := e.GetFrameByIndex(0)

    // Check "+" has correct metadata
    plusVal, _ := rootFrame.Get("+")
    plusFn, _ := plusVal.AsFunction()

    require.Equal(t, "+", plusFn.Name)
    require.True(t, plusFn.Infix, "+ should be infix")
    require.NotNil(t, plusFn.Doc, "+ should have documentation")
    require.Equal(t, 2, len(plusFn.Params), "+ should have 2 parameters")
}
```

### Test 4: Construction Performance

```go
func BenchmarkNewEvaluator(b *testing.B) {
    for i := 0; i < b.N; i++ {
        _ = eval.NewEvaluator()
    }
    // Target: <1ms per construction (1,000,000 ns)
}
```

---

## Dependencies

**Depends On**:
- `internal/frame.NewFrameWithCapacity()` - Root frame creation
- `internal/native.Register*Natives()` - Category registration functions
- `internal/stack.NewStack()` - Stack initialization

**Depended On By**:
- `cmd/viro/main.go` - REPL initialization
- `internal/repl.NewRepl()` - REPL evaluator setup
- All test files - Test evaluator creation

---

## Backward Compatibility

**Breaking Changes**: None

**Migration Notes**:
- External code calling `NewEvaluator()` - No changes required
- Existing evaluator usage - Identical behavior for non-shadowing code
- Test suites - Should pass without modification

---

## References

- [spec.md](../spec.md) - FR-005, FR-010, SC-002
- [data-model.md](../data-model.md) - Evaluator Construction State Machine
- [research.md](../research.md) - Q5: Evaluator construction sequence

---

**Contract Status**: ✅ Complete | **Version**: 1.0 | **Approved**: Pending implementation
