# Contract: Root Frame Native Registration

**Feature**: 003-natives-within-frame
**Component**: `internal/native.Register*Natives()`
**Date**: 2025-10-12

## Purpose

Defines the contract for native function registration functions that populate the root frame during evaluator construction. Each category registration function is responsible for binding its native functions to the root frame with proper validation.

---

## Function Signatures

```go
func RegisterMathNatives(rootFrame *frame.Frame)
func RegisterSeriesNatives(rootFrame *frame.Frame)
func RegisterDataNatives(rootFrame *frame.Frame)
func RegisterIONatives(rootFrame *frame.Frame)
func RegisterControlNatives(rootFrame *frame.Frame)
func RegisterHelpNatives(rootFrame *frame.Frame)
```

**Package**: `internal/native`
**Visibility**: Public (exported)
**Returns**: None (panics on error)

---

## Contract (Applies to All Registration Functions)

### Pre-conditions

**Pre-1**: `rootFrame` parameter is non-nil

**Pre-2**: `rootFrame` has sufficient capacity (pre-allocated to avoid reallocations)

**Pre-3**: `rootFrame` does not already contain names from this category (no prior registration)

### Post-conditions

**Post-1**: All natives for this category bound to `rootFrame`

**Post-2**: All bindings satisfy:
- Key (word) = native function name (string)
- Value = `value.FuncVal(functionValue)` wrapping `*value.FunctionValue`
- FunctionValue.Type = `FuncNative`
- FunctionValue metadata complete (Name, Params, Doc, Native implementation)

**Post-3**: No duplicate names within category

**Post-4**: Function order preserved (for diagnostics/debugging)

### Invariants

**Inv-1**: Each category registration is idempotent (calling twice with same frame produces same result - second call would detect duplicates and panic)

**Inv-2**: Registration functions are pure (no global state modification except frame parameter)

**Inv-3**: All FunctionValue instances created during registration persist for evaluator lifetime (no GC)

### Failure Modes

**FM-1**: **Nil root frame parameter**
- **Trigger**: `rootFrame == nil`
- **Behavior**: Panic with message `"RegisterXNatives: rootFrame is nil"`
- **Recoverable**: No (fatal error)

**FM-2**: **Duplicate native name**
- **Trigger**: Name already exists in `rootFrame`
- **Behavior**: Panic with message `"native registration failed: duplicate name {name}"`
- **Recoverable**: No (fatal error)

**FM-3**: **Nil FunctionValue**
- **Trigger**: FunctionValue construction fails or returns nil
- **Behavior**: Panic with message `"native registration failed: {name} is nil"`
- **Recoverable**: No (fatal error)

---

## Category-Specific Contracts

### RegisterMathNatives

**Natives**: 20 functions
- **Group 1**: `+`, `-`, `*`, `/` (arithmetic)
- **Group 2**: `<`, `>`, `<=`, `>=`, `=`, `<>` (comparison)
- **Group 3**: `and`, `or`, `not` (logic)
- **Group 4**: `abs`, `min`, `max`, `sqrt`, `power`, `floor`, `ceiling` (advanced math)

**Special Properties**:
- Infix operators: `+`, `-`, `*`, `/`, `<`, `>`, `<=`, `>=`, `=`, `<>`, `and`, `or`
  - Must set `FunctionValue.Infix = true`
- All functions evaluated args (no raw parameters)

---

### RegisterSeriesNatives

**Natives**: 5 functions
- `first`, `last`, `append`, `insert`, `length?`

**Special Properties**:
- Series operations need evaluator access (some natives use `eval.Do_Blk`)
- `length?` has convention suffix `?` indicating predicate/query

---

### RegisterDataNatives

**Natives**: 8 functions
- **Group 6**: `set`, `get`, `type?` (data access)
- **Group 7**: `object`, `clone`, `copy`, `find`, `index?` (object operations)

**Special Properties**:
- `set`, `get`, `object` require evaluator access
- Object operations interact with frame system

---

### RegisterIONatives

**Natives**: 10 functions
- **Group 8**: `print`, `input` (console I/O)
- **Group 9**: `open`, `read`, `write`, `close`, `exists?`, `delete`, `rename`, `dir` (port operations)

**Special Properties**:
- `print` requires evaluator access (evaluates arguments)
- Port operations interact with sandboxed port system

---

### RegisterControlNatives

**Natives**: 5 functions
- **Group 10**: `if`, `when`, `loop`, `while` (control flow)
- **Group 11**: `fn` (function definition)

**Special Properties**:
- All control natives require evaluator access (evaluate conditionals/bodies)
- `fn` is special-case (function constructor)

---

### RegisterHelpNatives

**Natives**: 11 functions
- **Group 12**: `help`, `info` (help system)
- **Group 13**: `trace`, `debug`, `breakpoint`, `step`, `reflect`, `words`, `values`, `stats`, `gc` (observability)

**Special Properties**:
- Help functions interact with native documentation system
- Debug functions interact with global debugger/tracer state

---

## Implementation Pattern

### Standard Registration Pattern

```go
func RegisterMathNatives(rootFrame *frame.Frame) {
    if rootFrame == nil {
        panic("RegisterMathNatives: rootFrame is nil")
    }

    // Define natives for this category
    natives := []struct {
        name string
        fn   *value.FunctionValue
    }{
        {
            name: "+",
            fn: value.NewNativeFunction(
                "+",
                []value.ParamSpec{
                    value.NewParamSpec("left", true),  // evaluated
                    value.NewParamSpec("right", true),
                },
                func(args []value.Value, refValues map[string]value.Value, eval value.Evaluator) (value.Value, error) {
                    result, err := Add(args)
                    if err != nil {
                        return value.NoneVal(), err
                    }
                    return result, nil
                },
            ),
        },
        // ... more natives
    }

    // Post-process metadata
    for i := range natives {
        fn := natives[i].fn
        if fn == nil {
            panic(fmt.Sprintf("native registration failed: %s is nil", natives[i].name))
        }

        // Set infix flag if applicable
        if natives[i].name == "+" || natives[i].name == "-" /* ... */ {
            fn.Infix = true
        }

        // Attach documentation
        fn.Doc = &NativeDoc{
            Category:    "Math",
            Summary:     "Adds two numbers",
            Description: "...",
            Parameters:  []ParamDoc{ /* ... */ },
            Returns:     "...",
            Examples:    []string{"3 + 4  ; => 7"},
        }
    }

    // Bind to root frame with validation
    for _, n := range natives {
        // Check for duplicates
        if _, exists := rootFrame.Get(n.name); exists {
            panic(fmt.Sprintf("native registration failed: duplicate name %s", n.name))
        }

        // Bind native to frame
        rootFrame.Bind(n.name, value.FuncVal(n.fn))
    }
}
```

---

## Usage Examples

### Example 1: Successful Registration

```go
// Create root frame
rootFrame := frame.NewFrameWithCapacity(frame.FrameClosure, -1, 80)

// Register math natives
native.RegisterMathNatives(rootFrame)

// Verify binding
plusVal, found := rootFrame.Get("+")
// found == true
// plusVal.Type == TypeFunction

plusFn, _ := plusVal.AsFunction()
// plusFn.Name == "+"
// plusFn.Type == FuncNative
// plusFn.Infix == true
```

### Example 2: Duplicate Name Detection

```go
rootFrame := frame.NewFrameWithCapacity(frame.FrameClosure, -1, 80)

// Register math natives (includes "+")
native.RegisterMathNatives(rootFrame)

// Manually bind duplicate (simulates error)
rootFrame.Bind("+", value.FuncVal(otherFn))

// Register series natives (no conflict)
native.RegisterSeriesNatives(rootFrame)  // Success

// Try to register math again
native.RegisterMathNatives(rootFrame)  // Panics: "duplicate name +"
```

---

## Testing Strategy

### Test 1: Category Registration

```go
func TestRegisterMathNatives(t *testing.T) {
    rootFrame := frame.NewFrameWithCapacity(frame.FrameClosure, -1, 80)

    native.RegisterMathNatives(rootFrame)

    mathNatives := []string{"+", "-", "*", "/", "<", ">", "and", "or", "not"}
    for _, name := range mathNatives {
        val, found := rootFrame.Get(name)
        require.True(t, found, "native %s not registered", name)

        fn, ok := val.AsFunction()
        require.True(t, ok, "native %s not a function", name)
        require.Equal(t, value.FuncNative, fn.Type)
    }
}
```

### Test 2: Metadata Preservation

```go
func TestRegisterMathNativesPreservesMetadata(t *testing.T) {
    rootFrame := frame.NewFrameWithCapacity(frame.FrameClosure, -1, 80)
    native.RegisterMathNatives(rootFrame)

    plusVal, _ := rootFrame.Get("+")
    plusFn, _ := plusVal.AsFunction()

    require.Equal(t, "+", plusFn.Name)
    require.True(t, plusFn.Infix)
    require.NotNil(t, plusFn.Doc)
    require.Equal(t, "Math", plusFn.Doc.Category)
    require.Contains(t, plusFn.Doc.Summary, "add")
}
```

### Test 3: Nil Frame Panic

```go
func TestRegisterMathNativesPanicsOnNilFrame(t *testing.T) {
    require.Panics(t, func() {
        native.RegisterMathNatives(nil)
    })
}
```

### Test 4: Duplicate Detection

```go
func TestRegisterMathNativesPanicsOnDuplicate(t *testing.T) {
    rootFrame := frame.NewFrameWithCapacity(frame.FrameClosure, -1, 80)
    native.RegisterMathNatives(rootFrame)

    // Manually create duplicate
    rootFrame.Bind("+", value.FuncVal(someFn))

    require.Panics(t, func() {
        native.RegisterMathNatives(rootFrame)
    }, "Should panic on duplicate name")
}
```

---

## Dependencies

**Depends On**:
- `internal/frame.Frame.Bind()` - Frame binding operation
- `internal/frame.Frame.Get()` - Duplicate detection
- `internal/value.NewNativeFunction()` - FunctionValue construction
- `internal/value.FuncVal()` - Value wrapping

**Depended On By**:
- `internal/eval.NewEvaluator()` - Evaluator construction

---

## Backward Compatibility

**Breaking Changes**: None

**Migration Notes**:
- Old registry-based registration code removed (internal change)
- Public API unchanged (evaluator construction still `NewEvaluator()`)
- Native function implementations unchanged

---

## Performance

**Registration Performance**:
- Each `Bind()` call: ~10 µs (amortized O(1) with pre-capacity)
- Category registration: 5-20 functions × 10 µs = 50-200 µs each
- Total registration: ~700 µs for all 70 natives

**Optimization**:
- Pre-allocate root frame capacity (80 slots) to avoid slice reallocations
- Registration functions called sequentially (no parallelization benefit)

---

## References

- [spec.md](../spec.md) - FR-001, FR-005, FR-008, FR-010
- [data-model.md](../data-model.md) - Native Function Binding entity
- [research.md](../research.md) - D1: Native registration file organization

---

**Contract Status**: ✅ Complete | **Version**: 1.0 | **Approved**: Pending implementation
