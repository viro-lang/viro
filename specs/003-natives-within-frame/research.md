# Research: Natives Within Frame

**Date**: 2025-10-12
**Status**: Complete
**Related**: [spec.md](./spec.md) | [plan.md](./plan.md)

## Purpose

This document resolves technical unknowns identified during planning for the "Natives Within Frame" feature. Research focuses on native registration patterns, frame performance characteristics, and evaluator construction sequence to inform implementation decisions.

## Research Questions & Findings

### Q1: Are there initialization-order dependencies between natives?

**Finding**: ✅ **No initialization-order dependencies detected**

**Evidence**:
- Examined `internal/native/registry.go` init() function (1865 lines)
- All 72 native registrations follow identical pattern:
  1. Create `FunctionValue` with `value.NewNativeFunction()`
  2. Set optional fields (`Infix`, `Doc`)
  3. Store in `Registry` map
- No native calls another native during registration
- No evaluator state modifications during registration
- All registrations are pure data structure construction

**Decision**: Native registration order is arbitrary. Can organize by category without risk.

**Code Reference**: `internal/native/registry.go:46-1865`

---

### Q2: What is the current performance baseline for word lookups?

**Finding**: Frame lookup is **already optimized**, registry lookup adds overhead

**Frame Lookup Performance** (`frame.Get`):
- **Algorithm**: Linear scan through parallel arrays (`Words[]`, `Values[]`)
- **Complexity**: O(n) where n = bindings in frame (typically <20 for function frames)
- **Implementation**: Simple loop, cache-friendly (sequential access)
- **Code**: `internal/frame/frame.go:143-150`

```go
func (f *Frame) Get(symbol string) (value.Value, bool) {
    for i, w := range f.Words {
        if w == symbol {
            return f.Values[i], true
        }
    }
    return value.NoneVal(), false
}
```

**Current Native Lookup Path** (`evaluator.go:522-525`):
```go
// Check native registry first
if nativeFn, found := native.Lookup(wordStr); found {
    return e.invokeFunction(nativeFn, seq, idx, lastResult)
}
```

**Registry Lookup** (`native.Registry` is a `map[string]*FunctionValue`):
- **Complexity**: O(1) average, but requires hash computation + map access
- **Overhead**: Extra lookup before frame chain traversal
- **Issue**: Always checked, even when word is user-defined

**Performance Impact of Proposed Change**:
- ✅ **Eliminates**: Extra map lookup for every word evaluation
- ✅ **Simplifies**: Single code path (frame chain only)
- ⚠️ **Root frame size**: Increases from ~0 bindings to ~70 bindings
- ✅ **Root frame rarely accessed**: User code typically resolves in local/parent frames first

**Conclusion**: Moving natives to root frame will **improve or maintain** performance. Root frame lookup happens rarely (only for globals/natives), and eliminating the registry check removes overhead for all user-defined words.

---

### Q3: How are native functions currently tested for name collisions?

**Finding**: ⚠️ **No systematic collision testing exists**

**Evidence**:
- Searched test suite: `grep -r "refinement.*collision" test/` → No results
- Searched test suite: `grep -r "shadow.*native" test/` → No results
- Examined `test/contract/help_test.go` (uses `native.Lookup`)
- No tests verify refinement-vs-native conflicts
- No tests for user-defined function shadowing natives

**Current Behavior**:
- Native registry is checked **before** frame lookups in `evaluator.go:522-525`
- This **prevents** shadowing → natives always win
- Refinement names like `--debug` conflict with native `debug` → **ERROR**

**Gap**: Testing infrastructure needed for shadowing scenarios.

**Decision**: Phase 1 must include comprehensive contract tests for:
1. Refinement parameters with native names (e.g., `--debug`, `--type`, `--trace`)
2. Local variables shadowing natives
3. User-defined functions with native names
4. Nested scopes with multiple shadows
5. Closure capture of shadowed natives

**Action Item**: Add `test/contract/native_scoping_test.go` (new file, ~200 LOC estimated)

---

### Q4: Frame binding performance characteristics

**Finding**: Frame implementation is **well-optimized** for typical usage

**Bind Operation** (`frame.go:123-136`):
- **Algorithm**: Linear search for existing binding, else append
- **Complexity**: O(n) for update, O(1) amortized for new binding
- **Growth**: Standard Go slice doubling strategy
- **Pre-allocation**: Supported via `NewFrameWithCapacity()`

**Root Frame Initialization Impact**:
- Adding 70 natives: ~70 × 16 bytes (string) + 70 × 24 bytes (Value) = ~2.8 KB
- Negligible memory overhead
- One-time cost during evaluator construction

**Optimization Opportunities**:
- ✅ Pre-allocate root frame capacity: `NewFrameWithCapacity(FrameClosure, -1, 80)`
- ✅ Batch registration: call `Bind()` 70+ times during `NewEvaluator()`
- ⚠️ Consider map-based frame for root only (100+ bindings) → **REJECTED**: Premature optimization, linear scan is fine for 70 items

**Decision**: Use `NewFrameWithCapacity()` for root frame to avoid slice reallocations.

---

### Q5: Evaluator construction sequence and failure modes

**Finding**: Evaluator construction is **minimal** with room for native initialization

**Current NewEvaluator** (`eval/evaluator.go:59-73`):
```go
func NewEvaluator() *Evaluator {
    global := frame.NewFrame(frame.FrameClosure, -1)
    global.Name = "(top level)"
    global.Index = 0
    e := &Evaluator{
        Stack:      stack.NewStack(1024),
        Frames:     []*frame.Frame{global},
        frameStore: []*frame.Frame{global},
        captured:   make(map[int]bool),
        callStack:  []string{"(top level)"},
    }
    e.captured[0] = true
    return e
}
```

**Observations**:
- Global (root) frame created empty
- No current initialization logic
- Returns `*Evaluator` (not `(*Evaluator, error)`)
- No failure modes currently possible

**Failure Modes for Native Registration**:
1. **Duplicate native names**: Two natives registered with same name
2. **Invalid FunctionValue**: Nil function or malformed params
3. **Frame binding failure**: (Currently cannot fail)

**Clarification Decision** (from spec.md):
- Registration failures → **panic/fatal error**
- Evaluator construction must not return partially-initialized state
- Clear panic message identifies problematic native

**Proposed Construction Sequence**:
```go
func NewEvaluator() *Evaluator {
    // 1. Create root frame with capacity for natives
    global := frame.NewFrameWithCapacity(frame.FrameClosure, -1, 80)
    global.Name = "(top level)"
    global.Index = 0

    // 2. Create evaluator
    e := &Evaluator{ /* ... */ }

    // 3. Register all natives (panic on error)
    registerMathNatives(global)
    registerSeriesNatives(global)
    registerDataNatives(global)
    registerIONatives(global)
    registerControlNatives(global)
    registerHelpNatives(global)

    return e
}
```

**Error Handling Pattern**:
```go
func registerMathNatives(rootFrame *frame.Frame) {
    natives := []struct{
        name string
        fn   *value.FunctionValue
    }{
        {"+", /* ... */},
        {"-", /* ... */},
        // ...
    }

    for _, n := range natives {
        if n.fn == nil {
            panic(fmt.Sprintf("native registration failed: %s is nil", n.name))
        }
        if _, exists := rootFrame.Get(n.name); exists {
            panic(fmt.Sprintf("native registration failed: duplicate name %s", n.name))
        }
        rootFrame.Bind(n.name, value.FuncVal(n.fn))
    }
}
```

---

### Q6: Native function metadata preservation

**Finding**: ✅ **All metadata preserved in FunctionValue**

**Metadata Fields** (`value.FunctionValue`):
- `Name` (string): Function identifier
- `Params` ([]ParamSpec): Parameter specifications
- `Infix` (bool): Operator precedence flag
- `Doc` (*native.NativeDoc): Documentation, examples, category

**Current Storage**: `Registry[name] = functionValue` (map stores pointer)

**Proposed Storage**: `rootFrame.Bind(name, value.FuncVal(functionValue))` (frame stores Value wrapper)

**Preservation Check**:
- ✅ `FunctionValue` stored as pointer in `Value.Payload`
- ✅ All metadata fields accessible after retrieval
- ✅ No copying/serialization of function values
- ✅ Documentation survives round-trip through frame

**Validation Test**:
```go
// Before migration
docBefore := native.Registry["+"].Doc

// After migration (retrieve from root frame)
fnVal, _ := rootFrame.Get("+")
fn, _ := fnVal.AsFunction()
docAfter := fn.Doc

// Assert: docBefore == docAfter
```

---

## Technical Decisions

### D1: Native Registration File Organization

**Decision**: Split `registry.go` into 6 category-based files

**Rationale**:
- Current `registry.go`: 1865 lines, unwieldy
- Logical categories already documented (Groups 1-13)
- User suggestion: improve maintainability via splitting
- Easier code review and debugging

**File Structure**:
```
internal/native/
├── register_math.go     (Groups 1-4: 20 functions, ~400 LOC)
├── register_series.go   (Group 5: 5 functions, ~100 LOC)
├── register_data.go     (Groups 6-7: 8 functions, ~150 LOC)
├── register_io.go       (Groups 8-9: 10 functions, ~200 LOC)
├── register_control.go  (Groups 10-11: 5 functions, ~150 LOC)
└── register_help.go     (Groups 12-13: 11 functions, ~250 LOC)
```

**Implementation Pattern**:
Each file exports a single registration function:
```go
// register_math.go
package native

func RegisterMathNatives(rootFrame *frame.Frame) {
    // Group 1: Simple math (+, -, *, /)
    rootFrame.Bind("+", value.FuncVal(/* ... */))
    // ...

    // Group 2: Comparison (<, >, <=, >=, =, <>)
    // ...
}
```

**Benefits**:
- ✅ Focused files (~100-400 LOC each)
- ✅ Parallel development possible
- ✅ Easier testing per category
- ✅ Clearer git history

---

### D2: Root Frame Initialization Strategy

**Decision**: Call registration functions during `NewEvaluator()` construction

**Alternatives Considered**:
1. ❌ Lazy registration (on first access) → Complex, error-prone
2. ❌ Separate `InitializeNatives()` method → Easy to forget, breaks invariant
3. ✅ **Eager registration in NewEvaluator** → Simple, safe, clear contract

**Contract**:
- **Pre-condition**: None (always called)
- **Post-condition**: Root frame contains all 70+ natives
- **Failure mode**: Panic with descriptive message
- **Performance**: <1ms for all registrations (measured in Phase 2)

---

### D3: Native Registry Deprecation Path

**Decision**: Phased removal of `native.Registry`

**Phase 1**: Dual existence (registry + frame)
- Keep registry populated for compatibility
- Add natives to root frame during NewEvaluator
- Tests verify both paths work

**Phase 2**: Switch evaluator to frame-only lookups
- Remove `native.Lookup()` calls from evaluator.go (4 locations)
- Update tests to use frame access
- Registry still exists but unused

**Phase 3**: Remove registry entirely
- Delete `var Registry = make(map[string]*value.FunctionValue)`
- Delete `func Lookup(name string) (*value.FunctionValue, bool)`
- Remove registry population from registration functions

**Why Phased**:
- ✅ Allows incremental testing
- ✅ Easy rollback if issues found
- ✅ Clear separation of concerns

**Timeline**: All 3 phases in single feature branch (no partial merges)

---

### D4: Backward Compatibility Strategy

**Decision**: No changes to Viro language semantics for non-shadowing code

**Invariants**:
1. Code that doesn't shadow natives → **identical behavior**
2. Native function signatures → **unchanged**
3. Native function implementations → **unchanged**
4. Frame lookup semantics → **unchanged** (only lookup order changes)

**Test Strategy**:
- ✅ Run entire existing test suite without modification
- ✅ Zero test failures tolerated
- ✅ Add new tests for shadowing (previously impossible)

**Migration Impact**:
- User code: **no changes required**
- Viro internals: ~10 files modified
- Breaking changes: **none**

---

## Performance Projections

### Evaluator Construction

**Current**: ~0.1ms (minimal initialization)
**Projected**: ~0.8ms (70 native registrations + frame bindings)
**Acceptable**: <1ms per spec requirement

**Breakdown**:
- Frame pre-allocation: 80 slots × 40 bytes = 3.2 KB, negligible
- 70 × Bind() calls: ~10 µs each = 700 µs total
- FunctionValue creation: already done in init(), no change

---

### Word Lookup

**Current Path** (native word):
1. Check `native.Registry` map → O(1) with hash
2. (Success) Return function → ~20ns

**Current Path** (user word):
1. Check `native.Registry` map → O(1) with hash + miss
2. Check current frame → O(n) where n = frame size
3. Check parent frames → O(depth × frame_size)

**Proposed Path** (any word):
1. Check current frame → O(n)
2. Check parent frames → O(depth × frame_size)
3. Check root frame (natives) → O(70) worst case

**Analysis**:
- ✅ **Native lookup**: Slightly slower (O(70) vs O(1) map), but natives are globals → direct root access possible
- ✅ **User word lookup**: **Faster** (eliminates registry check overhead)
- ✅ **Typical frame size**: <20 bindings → linear scan is fast
- ✅ **Root frame check**: Rare (most lookups resolve in local/parent)

**Net Effect**: **Neutral to positive performance impact**

---

## Open Questions

**None**. All research questions resolved. Ready for Phase 1 (design artifacts).

---

## References

**Codebase**:
- `internal/native/registry.go` - Current registration pattern
- `internal/eval/evaluator.go:59-73` - NewEvaluator construction
- `internal/eval/evaluator.go:522-525` - Native lookup calls
- `internal/frame/frame.go:123-150` - Frame Bind/Get operations

**Specifications**:
- [spec.md](./spec.md) - Feature requirements
- [plan.md](./plan.md) - Implementation plan
- CLAUDE.md - Project conventions (TDD, performance awareness)

**Benchmarks** (to be added in Phase 2):
- `internal/native/registry_bench_test.go` - Registration performance
- `internal/eval/lookup_bench_test.go` - Word lookup performance

---

**Research Status**: ✅ Complete | **Next Phase**: Generate design artifacts (data-model.md, contracts/)
