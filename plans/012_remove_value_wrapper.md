# Plan 012: Remove Value Wrapper Using Pure Interface Implementations

## Overview

Remove the `value.Value` struct wrapper by implementing the `core.Value` interface directly on each value type. This eliminates one layer of indirection while maintaining type safety and polymorphism through Go's interface system.

## Problem Statement

Current architecture has performance overhead:
1. **Double Indirection**: `core.Value` interface wraps `value.Value` struct which wraps `Payload any`
2. **Unnecessary Wrapper**: The `value.Value` struct adds no functionality, just wraps the payload
3. **Memory Overhead**: Each value allocates 3 objects (interface, struct, payload)
4. **Cache Inefficiency**: Scattered allocations harm CPU cache performance

Current structure:
```
core.Value (interface) → value.Value (struct) → Payload (any)
                          - Type: core.ValueType
                          - Payload: any
```

## Goals

1. Eliminate `value.Value` struct wrapper completely
2. Have each value type implement `core.Value` interface directly
3. Reduce memory allocations and indirection
4. Maintain type safety and polymorphism through interfaces
5. Use idiomatic Go patterns (like `time.Duration`)

## Recommended Approach: Pure Interface Implementations (Option 4)

Each value type implements the `core.Value` interface directly. For primitive types, use type definitions (not aliases) to create new types that support method attachment.

### Key Insight: Type Definitions vs Type Aliases

Go allows creating new types based on primitives:
```go
type IntValue int64      // Type definition - NEW type, can add methods
type IntAlias = int64    // Type alias - SAME as int64, cannot add methods
```

This is a well-established Go pattern used throughout the standard library:
- `time.Duration` (based on `int64`)
- `http.HandlerFunc` (based on function type)
- Many others

### Proposed Type Hierarchy

**Primitive-Based Value Types** (new type definitions):
```go
// In internal/value/ package
type IntValue int64
type LogicValue bool
type WordValue string
type SetWordValue string
type GetWordValue string
type LitWordValue string
type DatatypeValue string
type NoneValue struct{}  // Empty struct for sentinel value
```

**Struct-Based Value Types** (already exist, just add interface methods):
```go
// Already in internal/value/ package
type StringValue struct { ... }
type BlockValue struct { ... }
type FunctionValue struct { ... }
type DecimalValue struct { ... }
type ObjectInstance struct { ... }
type PortValue struct { ... }
type PathValue struct { ... }
type BinaryValue struct { ... }
```

### Interface Implementation Pattern

Each type implements the `core.Value` interface:

```go
// IntValue implements core.Value
func (i IntValue) GetType() core.ValueType {
    return value.TypeInteger
}

func (i IntValue) String() string {
    return strconv.FormatInt(int64(i), 10)
}

func (i IntValue) Mold() string {
    return i.String()
}

func (i IntValue) Form() string {
    return i.String()
}

func (i IntValue) Equals(other core.Value) bool {
    if oi, ok := other.(IntValue); ok {
        return i == oi
    }
    return false
}

// GetPayload can be removed or return the value itself
func (i IntValue) GetPayload() any {
    return int64(i)
}
```

Similar implementations for all other types.

### Type Mapping

| Viro Type | Go Type | Implementation |
|-----------|---------|----------------|
| None | `value.NoneValue` | Empty struct |
| Logic | `value.LogicValue` | Based on `bool` |
| Integer | `value.IntValue` | Based on `int64` |
| String | `*value.StringValue` | Pointer to struct |
| Word | `value.WordValue` | Based on `string` |
| SetWord | `value.SetWordValue` | Based on `string` |
| GetWord | `value.GetWordValue` | Based on `string` |
| LitWord | `value.LitWordValue` | Based on `string` |
| Block | `*value.BlockValue` | Pointer to struct |
| Paren | `*value.ParenValue` | Wrapper around `*BlockValue` |
| Function | `*value.FunctionValue` | Pointer to struct |
| Decimal | `*value.DecimalValue` | Pointer to struct |
| Object | `*value.ObjectInstance` | Pointer to struct |
| Port | `*value.PortValue` | Pointer to struct |
| Path | `*value.PathValue` | Pointer to struct |
| Datatype | `value.DatatypeValue` | Based on `string` |
| Binary | `*value.BinaryValue` | Pointer to struct |

### Handling Block vs Paren

Since Block and Paren share the same structure but different types, we have two options:

**Option A: Type field in BlockValue**
```go
type BlockValue struct {
    Elements []core.Value
    Index    int
    typ      core.ValueType  // TypeBlock or TypeParen
}

func (b *BlockValue) GetType() core.ValueType {
    return b.typ
}
```

**Option B: Separate wrapper types** (cleaner)
```go
type BlockValue struct {
    Elements []core.Value
    Index    int
}

type BlockVal struct{ *BlockValue }
type ParenVal struct{ *BlockValue }

func (b *BlockVal) GetType() core.ValueType { return TypeBlock }
func (p *ParenVal) GetType() core.ValueType { return TypeParen }
```

Recommendation: **Option A** (simpler, less boilerplate)

## Implementation Steps

### Phase 1: Create New Value Types (Parallel Implementation)

1. **Create primitive-based types** in `internal/value/primitives.go`:
   ```go
   type IntValue int64
   type LogicValue bool
   type NoneValue struct{}
   type WordValue string
   type SetWordValue string
   type GetWordValue string
   type LitWordValue string
   type DatatypeValue string
   ```

2. **Implement `core.Value` interface** for each primitive type:
   - `GetType() ValueType`
   - `String() string`
   - `Mold() string`
   - `Form() string`
   - `Equals(other Value) bool`
   - `GetPayload() any` (optional, for compatibility)

3. **Add interface methods** to existing struct types:
   - Update `StringValue`, `BlockValue`, `FunctionValue`, etc.
   - Ensure they implement `core.Value` interface

4. **Create new constructor functions** (keep old ones for now):
   ```go
   func IntVal(i int64) core.Value {
       return IntValue(i)
   }
   
   func StrVal(s string) core.Value {
       return NewStringValue(s)  // returns *StringValue
   }
   ```

5. **Create new type assertion helpers**:
   ```go
   func AsInteger(v core.Value) (int64, bool) {
       if iv, ok := v.(IntValue); ok {
           return int64(iv), true
       }
       return 0, false
   }
   ```

### Phase 2: Update Core Package

1. **Simplify `core.Value` interface** (optional):
   - Consider removing `GetPayload()` if not needed
   - Keep minimal interface surface

2. **Update `NativeFunc` signature** (no change needed):
   ```go
   type NativeFunc func(args []Value, refValues map[string]Value, eval Evaluator) (Value, error)
   ```
   This already uses the interface!

### Phase 3: Migrate Codebase

1. **Update native functions** in `internal/native/`:
   - Replace `value.AsInteger()` with type assertion pattern
   - Replace `value.IntVal()` with new constructors
   - Test each native function individually

2. **Update evaluator** in `internal/eval/`:
   - No major changes needed (already uses `core.Value`)
   - Update any direct `value.Value` references

3. **Update frame package** in `internal/frame/`:
   - Storage already uses `[]core.Value`
   - No changes needed

4. **Update tests**:
   - Update test helpers to use new constructors
   - Ensure all contract tests pass

### Phase 4: Remove Old Value Struct

1. **Delete `value.Value` struct** from `internal/value/value.go`
2. **Remove old constructor functions** (if any remain)
3. **Clean up imports** across codebase

### Phase 5: Optimization and Cleanup

1. **Run benchmarks**:
   - `BenchmarkMathAdd`
   - `BenchmarkEvalSimpleExpression`
   - `BenchmarkEvalComplexExpression`
   
2. **Profile memory allocations**:
   ```bash
   go test -bench=. -benchmem -memprofile=mem.out
   go tool pprof mem.out
   ```

3. **Verify improvements**:
   - Less memory per value allocation
   - Fewer allocations overall
   - Better performance

4. **Final cleanup**:
   - Remove `GetPayload()` if not used
   - Update documentation
   - Final test sweep

## Migration Example

### Before (Current System)

```go
// internal/value/value.go
type Value struct {
    Type    core.ValueType
    Payload any
}

func IntVal(i int64) Value {
    return Value{Type: TypeInteger, Payload: i}
}

func AsInteger(v core.Value) (int64, bool) {
    if v.GetType() != TypeInteger {
        return 0, false
    }
    i, ok := v.GetPayload().(int64)
    return i, ok
}

// internal/native/math.go
func Add(args []core.Value, ...) (core.Value, error) {
    a, ok := value.AsInteger(args[0])
    if !ok {
        return value.NoneVal(), mathTypeError("add", args[0])
    }
    b, ok := value.AsInteger(args[1])
    if !ok {
        return value.NoneVal(), mathTypeError("add", args[1])
    }
    result, overflow := addInt64(a, b)
    if overflow {
        return value.NoneVal(), overflowError("add")
    }
    return value.IntVal(result), nil
}
```

### After (Option 4)

```go
// internal/value/primitives.go
type IntValue int64

func (i IntValue) GetType() core.ValueType {
    return TypeInteger
}

func (i IntValue) String() string {
    return strconv.FormatInt(int64(i), 10)
}

func (i IntValue) Mold() string {
    return i.String()
}

func (i IntValue) Form() string {
    return i.String()
}

func (i IntValue) Equals(other core.Value) bool {
    if oi, ok := other.(IntValue); ok {
        return i == oi
    }
    return false
}

// Constructor returns interface
func IntVal(i int64) core.Value {
    return IntValue(i)
}

// Type assertion helper
func AsInteger(v core.Value) (int64, bool) {
    if iv, ok := v.(IntValue); ok {
        return int64(iv), true
    }
    return 0, false
}

// internal/native/math.go (minimal changes!)
func Add(args []core.Value, ...) (core.Value, error) {
    a, ok := value.AsInteger(args[0])  // Same call!
    if !ok {
        return value.NoneVal(), mathTypeError("add", args[0])
    }
    b, ok := value.AsInteger(args[1])
    if !ok {
        return value.NoneVal(), mathTypeError("add", args[1])
    }
    result, overflow := addInt64(a, b)
    if overflow {
        return value.NoneVal(), overflowError("add")
    }
    return value.IntVal(result), nil  // Same call!
}
```

**Key Observation**: Native function code barely changes! The API stays almost identical.

## Memory Layout Comparison

### Current System (per integer value)
```
core.Value interface:     16 bytes (pointer + type descriptor)
  └─> value.Value struct: 16 bytes (Type field + Payload field)
       └─> int64 (boxed): 24 bytes (interface wrapper + 8-byte value)
Total: ~56 bytes (3 allocations)
```

### Option 4 (per integer value)
```
core.Value interface:    16 bytes (pointer + type descriptor)
  └─> IntValue:           8 bytes (direct value, boxed in interface)
Total: 24 bytes (1 allocation)
```

**Improvement**: ~60% reduction in memory usage!

### For Pointer Types (e.g., StringValue)

**Current**:
```
core.Value interface:     16 bytes
  └─> value.Value struct: 16 bytes
       └─> *StringValue:   8 bytes (pointer)
            └─> StringValue: (on heap)
Total: 40 bytes + heap
```

**Option 4**:
```
core.Value interface:  16 bytes
  └─> *StringValue:     8 bytes (pointer)
       └─> StringValue: (on heap)
Total: 24 bytes + heap
```

**Improvement**: ~40% reduction in wrapper overhead!

## Performance Implications

### Advantages

1. **Eliminates one indirection layer**: No more `value.Value` struct wrapper
2. **Fewer allocations**: Primitives box directly into interface (1 allocation vs 2+)
3. **Better cache locality**: Smaller memory footprint per value
4. **Idiomatic Go**: Uses language features as designed
5. **Type safety**: Interface ensures all types implement required methods
6. **Fast type assertions**: Go optimizes interface type assertions/switches

### Potential Concerns

1. **Interface boxing still exists**: But this is unavoidable in Go without losing polymorphism
2. **More boilerplate**: Each type needs interface methods (but code generators could help)
3. **Type assertion overhead**: Minimal in practice (Go optimizes these well)

### Benchmark Expectations

Based on memory analysis:
- **Math operations**: 20-30% faster (less indirection, fewer allocations)
- **Memory usage**: 40-60% reduction in value wrapper overhead
- **Allocation rate**: 50%+ reduction (eliminate Value struct allocation)

## API Compatibility

### Constructor Functions (Minimal Change)
```go
// Before and After have same signature!
func IntVal(i int64) core.Value
func StrVal(s string) core.Value
func BlockVal(elements []core.Value) core.Value
```

### Accessor Functions (Minimal Change)
```go
// Before and After have same signature!
func AsInteger(v core.Value) (int64, bool)
func AsString(v core.Value) (*StringValue, bool)
func AsBlock(v core.Value) (*BlockValue, bool)
```

### Type Checking (Minimal Change)
```go
// Before and After have same call
if v.GetType() == value.TypeInteger { ... }
```

**Key Point**: Most calling code doesn't need to change!

## Risk Mitigation

### Risks

1. **Boilerplate code**: Each type needs interface methods
2. **Block/Paren handling**: Need to distinguish two types using same struct
3. **Test coverage**: Must ensure all types properly implement interface
4. **Migration errors**: Possible to miss some usages during migration

### Mitigations

1. **Code generation**: Could generate interface methods from templates
2. **Compiler checks**: Go compiler will catch missing interface methods
3. **Incremental migration**: Keep old and new systems in parallel initially
4. **Comprehensive testing**: All contract tests must pass
5. **Type switch exhaustiveness**: Use tools to ensure all cases covered

### Rollback Plan

- Git branch strategy allows easy reversion
- Parallel implementation in Phase 1 allows comparison
- Benchmarks provide objective go/no-go decision point

## Success Criteria

1. **Performance**: ≥20% improvement in math benchmarks
2. **Memory**: ≥40% reduction in value wrapper overhead
3. **Allocations**: ≥50% reduction in allocation count
4. **Tests**: All contract tests pass without modification
5. **Code quality**: Code remains readable and maintainable

## Timeline Estimate

- Phase 1 (Create new types): 6-8 hours
- Phase 2 (Update core): 1-2 hours
- Phase 3 (Migrate codebase): 8-12 hours
- Phase 4 (Remove old struct): 2-3 hours
- Phase 5 (Optimization): 3-4 hours

**Total**: ~20-30 hours of focused work

## Real-World Go Examples

This pattern is used extensively in the Go standard library:

### time.Duration
```go
type Duration int64

func (d Duration) String() string { ... }
func (d Duration) Seconds() float64 { ... }
```

### http.HandlerFunc
```go
type HandlerFunc func(ResponseWriter, *Request)

func (f HandlerFunc) ServeHTTP(w ResponseWriter, r *Request) {
    f(w, r)
}
```

### fs.FileMode
```go
type FileMode uint32

func (m FileMode) String() string { ... }
func (m FileMode) IsDir() bool { ... }
```

## Open Questions

1. **Should we remove `GetPayload()`?**
   - Likely yes, since each type IS the payload
   - Keep during transition for compatibility
   - Remove in cleanup phase

2. **Block vs Paren: Option A or B?**
   - Recommendation: Option A (type field in BlockValue)
   - Simpler, less code duplication
   - Can revisit if needed

3. **Should constructors be in `core` or `value` package?**
   - Current: constructors in `value` package
   - Recommendation: keep in `value` package
   - `core` package defines interface only

## References

- Current value system: `internal/value/value.go`
- Core interface: `internal/core/core.go`
- Native functions: `internal/native/*.go`
- Benchmarks: `internal/native/math_bench_test.go`, `test/integration/eval_bench_test.go`
- Go interfaces guide: https://go.dev/tour/methods/9
- Type definitions: https://go.dev/ref/spec#Type_definitions

## Next Steps

1. Create experimental branch: `git checkout -b feature/remove-value-wrapper`
2. **Phase 1**: Implement all new types in `internal/value/primitives.go`
3. **Phase 1**: Add interface methods to existing struct types
4. **Phase 1**: Write unit tests for each new type
5. **Checkpoint**: Run benchmarks, verify improvements
6. If improvements confirmed: proceed with Phase 2-5
7. If not: investigate and optimize, or reconsider approach
