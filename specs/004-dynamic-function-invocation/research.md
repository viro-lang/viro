# Research: Dynamic Function Invocation (Action Types)

**Feature**: 004-dynamic-function-invocation
**Date**: 2025-10-13

## Research Questions

This document captures technical decisions and their rationale for implementing the action dispatch system in Viro.

---

## 1. Action Value Type Implementation

### Decision
Implement `action!` as a new value type with the following structure:

```go
type ActionValue struct {
    Name      string        // Action name (e.g., "first", "append")
    ParamSpec []Parameter   // Parameter specification
    // Refinements handled by individual type implementations
}

// Global TypeRegistry (initialized at startup)
var TypeRegistry = map[ValueType]*Frame{
    TypeBlock:   &blockFrame,   // Direct pointer to block type frame
    TypeString:  &stringFrame,  // Direct pointer to string type frame
    // ... etc.
}

// Type frames have:
// - Parent = 0 (index to root frame on stack)
// - Index = -1 (not in frameStore)
```

### Rationale
- **Simple Action Value**: ActionValue only stores Name and ParamSpec. It doesn't maintain type-to-frame mappings, keeping action values lightweight (~50 bytes).
- **Global TypeRegistry**: A single global map from ValueType → *Frame serves all actions. Each type (block!, string!, etc.) has one frame containing all its type-specific implementations.
- **Direct Pointer Storage**: Type frames are stored as direct pointers in TypeRegistry, not on the stack. This provides:
  - Faster dispatch (one fewer indirection - no frameStore.Get() call)
  - Cleaner separation (execution frames on stack, type metadata in registry)
  - Cleaner stack traces (type frames have Index=-1, don't appear in execution traces)
  - Safe pointers (type frames are permanent, created at startup, never moved)
- **Index-Based Parent References**: Type frames use Parent=0 (integer index) to reference root frame on stack, maintaining index-based architecture for parent chain traversal.
- **Type Frame Organization**: Each type owns its operations. For example, the block frame contains "first", "last", "append", etc. If a type doesn't support an operation, that function simply isn't in its frame.
- **Consistent Dispatch**: All actions use the same dispatch mechanism: Action.Name + ArgType → TypeRegistry[ArgType] → TypeFrame.Get(Action.Name)
- **Name Field**: Stored for better error messages when dispatch fails.

### Alternatives Considered
- **Per-Action TypeFrames Map**: Each action maintains its own map of supported types to frame indices. Rejected because it duplicates type-to-frame mappings across actions, increases memory per action, and creates multiple sources of truth. The global TypeRegistry approach is simpler and more consistent with the "type frames" concept (each type owns its operations).
- **Embedding Function Directly**: Storing the function implementation directly in the action value would require type-checking logic within the action value. Rejected because it duplicates the evaluator's existing type-based dispatch mechanism.
- **Multiple Dispatch (Multi-Methods)**: Dispatching on multiple argument types would be more powerful but significantly more complex. Rejected as out of scope per spec.md (Assumptions: "Actions always dispatch based on the first argument's type").

---

## 2. Type Frame Organization

### Decision
Create type frames eagerly at interpreter startup, one per native value type:

- **Initialization**: During `cmd/viro/main.go` startup, after stack/root frame creation
- **Frame Creation**: One frame per type (block!, string!, integer!, etc.)
- **Frame Parent**: Each type frame's parent is the root frame (Parent = 0, index) to inherit global bindings
- **Frame Storage**: Type frames stored directly in TypeRegistry (not on stack), each has Index = -1
- **Lifetime**: Type frames persist for entire interpreter session (never garbage collected)

**Frame Creation Order**:
1. Create root frame (index 0 on stack) with core natives (if, fn, etc.)
2. Create type frames for each ValueType:
   - Block frame (Index=-1, not in frameStore)
   - String frame (Index=-1, not in frameStore)
   - Integer frame (Index=-1, not in frameStore)
   - ... etc.
3. Store type frames in TypeRegistry (map[ValueType]*Frame)
4. Populate type frames with type-specific function implementations
5. Create action values (Name + ParamSpec only)
6. Bind action values into root frame

### Rationale
- **Eager Initialization**: All type frames created at startup ensures predictable performance (no lazy initialization overhead during execution) and simplifies debugging (frames always exist).
- **Parent Chain via Index**: Type frames use Parent=0 (integer index) to reference root frame on stack, maintaining index-based architecture while living in TypeRegistry.
- **Direct Storage in TypeRegistry**: Storing type frames in TypeRegistry (not on stack) provides cleaner separation between execution context (stack) and type metadata (registry).
- **Static Structure**: Type frames never modified after initialization, eliminating concurrency concerns and making the system easier to reason about.
- **Standard Frame Mechanism**: Reusing existing Frame structure (Words/Values arrays) requires no new data structures or special cases in the frame system itself.
- **Stack Trace Cleanliness**: Type frames have Index=-1 (not in frameStore), so they don't appear in execution stack traces.

### Alternatives Considered
- **Type Frames on Stack**: Store type frames in frameStore at indices 1, 2, 3, etc. Rejected because:
  - Type frames aren't execution context (never pushed/popped)
  - Wastes stack indices that appear in traces but aren't meaningful
  - Requires extra indirection during dispatch (TypeRegistry → index → frameStore.Get() → frame)
  - Mixes execution state with type metadata
- **Lazy Frame Creation**: Create type frames on first use of an action for that type. Rejected because it adds runtime complexity, makes initialization order unpredictable, and complicates testing.
- **Dynamic Frame Modification**: Allow adding functions to type frames at runtime. Rejected because it's out of scope (spec.md: "Dynamic modification of type frames at runtime" is explicitly excluded) and adds unnecessary complexity.

---

## 3. Dispatch Mechanism

### Decision
Implement dispatch in the evaluator as follows:

1. **Action Evaluation**: When evaluator encounters an `action!` value during evaluation:
   - Check if action is being invoked (i.e., there are arguments to consume)
   - If yes, proceed to step 2; if no, evaluate to itself (actions are first-class values)

2. **First Argument Type Check**:
   - Evaluate the first argument (if ParamSpec indicates it should be evaluated)
   - Get the value's type using `.Type` field
   - Look up the type in the global `TypeRegistry` map

3. **Frame Lookup**:
   - If type found in TypeRegistry: retrieve frame pointer directly
   - If type not found: generate script error "Action 'name' not defined for type T!" (type has no type frame)

4. **Function Resolution**:
   - Look up action name in the type frame using `frame.Get(action.Name)`
   - This gives us the type-specific FunctionValue
   - If not found: generate script error "Action 'name' not defined for type T!" (type frame exists but doesn't implement this action)

5. **Function Invocation**:
   - Invoke the resolved function with all arguments (including the first)
   - Let the function implementation handle subsequent argument validation
   - Pass refinements through unchanged

### Rationale
- **Evaluator Integration**: Dispatch happens in the evaluator's main switch statement on value types, consistent with how functions, words, and other special types are handled.
- **First Argument Only**: Dispatching solely on first argument type keeps the system simple and predictable, matching Viro's behavior.
- **Global TypeRegistry with Direct Pointers**: Using a single global map with direct frame pointers simplifies dispatch logic and provides fast lookup. One map access yields the frame directly (no frameStore.Get() indirection).
- **Two-Stage Lookup**: First check if type has a frame (TypeRegistry lookup), then check if frame has the action (frame.Get). This provides clear error messages for both "type not supported" and "action not implemented for type" scenarios.
- **Argument Delegation**: Type-specific implementations validate their own subsequent arguments, avoiding duplicate validation logic in the dispatcher.
- **Refinement Passthrough**: Refinements are passed unchanged to type implementations, letting each implementation decide how to handle them.

### Alternatives Considered
- **Pre-evaluation of All Arguments**: Evaluate all arguments before dispatch to ensure they're valid. Rejected because some action parameters might be unevaluated (e.g., `append/only block value` where `/only` needs the block unevaluated), and this would break that pattern.
- **Dispatcher Validates All Arguments**: Have dispatch logic validate all argument types. Rejected because it duplicates logic and makes adding new type-specific arguments harder (would require updating the dispatcher).
- **Method Caching**: Cache (action, type) → function resolutions to avoid repeated lookups. Deferred as premature optimization; measure performance first, then optimize if needed.

---

## 4. Native Function Migration Strategy

### Decision
Migrate existing series operations to actions incrementally:

**Phase 1: Core Series Actions** (must complete before merging)
- `first` - Get first element of series
- `last` - Get last element of series
- `append` - Add element to end of series
- `insert` - Add element at position in series
- `length?` - Get series length

**Phase 2: Extended Series Actions** (can be separate PR)
- `head`, `tail`, `next`, `back` - Navigation
- `at`, `skip` - Positioning
- `copy`, `clear`, `remove` - Manipulation

**Migration Process per Function**:
1. Create type-specific implementations in new files:
   - `internal/native/series_block.go` - Block-specific series functions
   - `internal/native/series_string.go` - String-specific series functions
2. Update `internal/native/register_series.go`:
   - Create action value with type frames map
   - Bind action to root frame (replacing old native binding)
3. Update existing contract tests to verify action dispatch
4. Add new contract tests for unsupported type errors
5. Remove old native implementations once action works

**Backward Compatibility**:
- From user perspective: no change (same function names, same behavior)
- From code perspective: functions become actions, but evaluation semantics identical
- Old test suite continues to pass during migration

### Rationale
- **Incremental Delivery**: Migrating in phases reduces risk and allows for early validation of the dispatch system.
- **Core Series First**: Functions in Phase 1 are most commonly used and provide sufficient test coverage for the dispatch mechanism.
- **File Organization**: Grouping type-specific implementations by type (series_block.go, series_string.go) makes the codebase easier to navigate.
- **No Breaking Changes**: Users see no difference in behavior, ensuring smooth transition.
- **Test Coverage**: Existing contract tests ensure no regressions during migration.

### Alternatives Considered
- **Big Bang Migration**: Convert all natives to actions at once. Rejected because it's high-risk and harder to debug if issues arise.
- **Separate Action Namespace**: Create new action names (e.g., `action-first`, `action-append`) and deprecate old natives. Rejected because it breaks existing code and creates user confusion.
- **Manual Registration**: Require developers to manually create type frames and register functions. Rejected because it's error-prone; prefer a registration helper function that automates type frame setup and function registration.

---

## 5. Error Handling and Diagnostics

### Decision
Implement two categories of action-related errors:

**Script Errors** (user-facing):
```go
// When action called on unsupported type
verror.NewScriptError("action-no-impl",
    [3]string{actionName, typeName, ""},
    near, where)
// Message: "Action 'first' not defined for type integer!"

// When action called with wrong arity (existing error, reused)
verror.NewScriptError("wrong-arity",
    [3]string{actionName, expected, got},
    near, where)
```

**Internal Errors** (implementation bugs):
```go
// When type frame lookup succeeds but function not in frame
verror.NewInternalError("action-frame-corrupt",
    [3]string{actionName, typeName, ""},
    near, where)
// Indicates: type registered but implementation missing (bug)
```

**Error Context**:
- `Near`: Include action call expression and first argument
- `Where`: Include call stack showing where action was invoked
- Error code range: 300-399 for script errors (action-no-impl = 341), 900+ for internal

### Rationale
- **Clear User Errors**: "action-no-impl" error clearly indicates what action was called and what type doesn't support it, guiding users to fix their code.
- **Debugging Aid**: Internal errors help developers identify bugs in type frame registration (e.g., forgot to add `first` to block frame).
- **Consistent Error System**: Reuses Viro's existing `verror` package categories, maintaining consistency with other interpreter errors.
- **Rich Context**: `Near` and `Where` fields provide enough context for users to locate and fix issues.

### Alternatives Considered
- **Generic "Type Mismatch" Error**: Reuse existing type error instead of action-specific error. Rejected because it doesn't convey that actions support the type, just not this particular action.
- **Warning Instead of Error**: When type unsupported, return `none` with a warning. Rejected because silent failures are harder to debug; explicit errors better follow Viro semantics philosophy.
- **Suggest Alternative**: Include "did you mean?" suggestions in error message. Deferred as nice-to-have; implement basic errors first, enhance later if needed.

---

## 6. Performance Considerations

### Decision
Accept dispatch overhead initially; measure and optimize later if needed.

**Expected Overhead**:
- Type frame map lookup: O(1) hash map access
- Frame chain traversal: O(1) to O(depth) depending on shadowing
- Function resolution: O(1) word lookup in type frame
- Total overhead: Estimated 2-5x slower than direct native call

**Measurement Plan**:
1. Add benchmark tests comparing:
   - Direct native call: `native.First(block)`
   - Action dispatch: `evaluator.Eval(firstAction, block)`
2. Run benchmarks on representative operations (first, append, length?)
3. Document overhead in research.md results section
4. If overhead > 10x, investigate optimizations

**Future Optimizations** (only if needed):
- Inline caching: Cache last (action, type) → function resolution
- Devirtualization: For hot paths, directly call implementation if type known
- JIT compilation: Generate specialized code for common type combinations

### Rationale
- **Premature Optimization Avoided**: Don't optimize until performance problem is proven.
- **Interpreter Context**: Viro is an interpreter, not a JIT; some overhead is acceptable for flexibility gains.
- **Measurement First**: Benchmarks provide objective data for optimization decisions.
- **Incremental Approach**: Implement core functionality, measure, then optimize hot paths.

### Alternatives Considered
- **Inline Caching from Start**: Implement caching immediately. Rejected because it adds complexity before we know if it's needed, and makes debugging harder during initial development.
- **Direct Call Optimization**: Have evaluator special-case common actions to skip dispatch. Rejected because it defeats the purpose of polymorphism and creates maintenance burden (special cases).
- **Compile-Time Devirtualization**: Analyze code to determine types statically. Rejected as out of scope; Viro is dynamically typed, static analysis would require major language changes.

---

## Technology Stack Summary

**Languages & Frameworks**:
- Go 1.21+ (existing)
- Go standard library only (no new dependencies)

**Core Packages Modified**:
- `internal/value` - Add action! type
- `internal/eval` - Add dispatch logic
- `internal/frame` - Add type frame initialization
- `internal/native` - Refactor registration

**Testing Approach**:
- Table-driven contract tests (existing pattern)
- Benchmark tests for performance measurement (new)
- Error case coverage for all action scenarios

**Best Practices Applied**:
- Index-based references (no pointers to frames)
- Immutable type frames (initialized once at startup)
- Contract-first testing (TDD)
- Incremental migration (phased rollout)

---

## Open Issues

None identified. All technical decisions resolved through research and alignment with existing Viro architecture.

---

## References

- Viro Architecture: `CLAUDE.md`
- Feature Spec: `specs/004-dynamic-function-invocation/spec.md`
- REBOL Actions: [REBOL/Core User Guide, Chapter 9: Functions](http://www.rebol.com/docs/core23/rebolcore-9.html) (reference for polymorphic function patterns)
- Go interface dispatch: Standard library's `reflect` package (for comparison, though not used in implementation)
