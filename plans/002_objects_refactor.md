# Plan: Object-Owned Frames Architecture Refactor

STATUS: Implemented

## Executive Summary

**Problem**: ObjectInstance currently uses FrameIndex to reference frames stored in evaluator.frameStore, creating tight coupling between objects and evaluators. This violates object self-containment.

**Solution**: Move object frames from evaluator storage to direct ownership by ObjectInstance, while preserving all frame-based semantics (prototypes, dynamic fields, scoping).

**Impact**: Major architectural improvement enabling self-contained objects, eliminating evaluator coupling, and simplifying object lifecycle management.

## Current Architecture Analysis

### Frame Storage Model

```go
// Current: Unified frame stack in evaluator
type Evaluator struct {
    frameStore []core.Frame  // ALL frames: functions, objects, blocks
}

type ObjectInstance struct {
    FrameIndex int  // Reference to evaluator.frameStore[idx]
}
```

### Problems

- **Tight coupling**: Objects only valid within specific evaluator instance
- **Pointer invalidation**: Frame stack expansion invalidates frame references
- **Complex lifecycle**: Objects depend on evaluator frame management
- **Limited portability**: Objects can't be serialized or passed between evaluators

### Benefits of Current System

- **Unified semantics**: All frames use same storage/lookup mechanism
- **Consistent scoping**: Objects participate in lexical scoping like functions
- **Simple implementation**: Single frame store for all contexts

## Proposed Architecture

### Object-Owned Frames Model

```go
// Proposed: Separate execution frames from object frames
type Evaluator struct {
    frameStore []core.Frame  // Only execution contexts (functions, blocks)
}

type ObjectInstance struct {
    Frame core.Frame         // Direct frame ownership
    ParentProto *ObjectInstance
    Manifest ObjectManifest
}
```

### Key Changes

- **Object frames owned by objects**: No evaluator dependency for field access
- **Evaluator focuses on execution**: frameStore contains only function/block contexts
- **Self-contained objects**: Complete state within object, serializable/portable
- **Preserved semantics**: Prototypes, dynamic fields, scoping behavior unchanged

## Implementation Plan

### Phase 1: Dual System Setup (Week 1-2)

**Goal**: Add owned frames alongside existing system, maintain full compatibility

#### Tasks

1. **Add Frame field to ObjectInstance**

   ```go
   type ObjectInstance struct {
       FrameIndex  int             // Keep for compatibility
       Frame       core.Frame      // New: owned frame
       ParentProto *ObjectInstance
       Manifest    ObjectManifest
   }
   ```

2. **Update object creation**
   - Modify `NewObject()` to create owned frame
   - Initialize owned frame with field bindings
   - Keep FrameIndex assignment for backward compatibility

3. **Add owned frame access methods**

   ```go
   func (obj *ObjectInstance) GetField(name string) (core.Value, bool) {
       return obj.Frame.Get(name)
   }

   func (obj *ObjectInstance) SetField(name string, val core.Value) {
       obj.Frame.Bind(name, val)
   }
   ```

4. **Update prototype chain logic**
   - Implement `GetFieldWithProto()` using owned frames
   - Maintain backward compatibility with FrameIndex-based lookup

#### Testing

- All existing tests pass
- New owned frame methods work correctly
- Prototype chains function with owned frames

### Phase 2: Migration to Owned Frames (Week 3-4)

**Goal**: Gradually migrate all object operations to use owned frames

#### Tasks

1. **Update native functions**
   - `Select`: Use `obj.GetFieldWithProto()` instead of evaluator lookup
   - `Put`: Use `obj.SetField()` instead of frame index access
   - `Make`: Update object creation to use owned frames

2. **Update evaluator object operations**
   - Remove object frame registration from evaluator
   - Update object field access in evaluator methods
   - Modify object creation in `Object` and `Make` natives

3. **Update string formatting**
   - Implement proper mold formatting in `ObjectInstance.String()`
   - Remove evaluator dependency from object formatting
   - Update `FormatValueAsString` to use owned frames

#### Testing

- All object operations work with owned frames
- Mold formatting produces correct output
- Performance benchmarks show no regression

### Phase 3: Cleanup and Optimization (Week 5-6)

**Goal**: Remove legacy FrameIndex system, optimize performance

#### Tasks

1. **Remove FrameIndex field**
   - Delete FrameIndex from ObjectInstance
   - Update all references to use owned frames
   - Remove evaluator object frame management

2. **Update evaluator architecture**
   - Simplify evaluator to focus on execution contexts
   - Remove object frame storage from frameStore
   - Optimize frame management for execution-only frames

3. **Performance optimization**
   - Profile and optimize owned frame access patterns
   - Ensure prototype chain lookups are efficient
   - Memory usage analysis and optimization

4. **Documentation updates**
   - Update CLAUDE.md with new architecture
   - Update data-model.md and research.md
   - Document object portability features

#### Testing

- Full test suite passes with new architecture
- Performance benchmarks meet or exceed current levels
- Memory usage is comparable or improved
- Integration tests validate object portability

## Migration Strategy

### Backward Compatibility

- **Phase 1-2**: Full backward compatibility maintained
- **Phase 3**: Breaking change - FrameIndex removal
- **Version bump**: Major version increment for Phase 3

### Rollback Plan

- **Phase 1-2**: Can rollback by removing Frame field and owned frame methods
- **Phase 3**: More complex - would need to restore FrameIndex system

### Risk Mitigation

- **Comprehensive testing**: Each phase has full test coverage
- **Gradual migration**: No big-bang changes
- **Feature flags**: Ability to enable/disable owned frames during transition

## Benefits

### Architectural

- **Self-contained objects**: No evaluator coupling
- **Serializable objects**: Can save/load object state
- **Portable objects**: Work across evaluator instances
- **Cleaner lifecycle**: Objects manage their own memory

### Performance

- **Direct access**: No evaluator lookup overhead for field access
- **Reduced coupling**: Less indirection in object operations
- **Better cache locality**: Object data co-located

### Developer Experience

- **Easier testing**: Objects can be created/tested independently
- **Better error messages**: No "invalid frame index" errors

## Risks and Challenges

### Implementation Complexity

- **Prototype chains**: Ensuring proper inheritance with owned frames
- **Scoping integration**: Objects still need to participate in lexical scoping
- **Memory management**: Proper cleanup of owned frames

### Performance Concerns

- **Memory overhead**: Each object carries its frame
- **Lookup performance**: Prototype chain traversal vs. frame index lookup
- **Cache efficiency**: Frame data spread across objects vs. contiguous array

### Testing Coverage

- **Edge cases**: Complex prototype chains, circular references
- **Concurrency**: If objects are shared between evaluators
- **Serialization**: Save/load object state correctly

## Success Criteria

### Functional

- ✅ All existing tests pass
- ✅ Object operations work identically
- ✅ Mold formatting produces correct output
- ✅ Prototype chains function properly
- ✅ Objects can be serialized/deserialized

### Performance

- ✅ No performance regression in benchmarks
- ✅ Memory usage comparable or improved
- ✅ Field access at least as fast as current system

### Architectural

- ✅ Objects are self-contained (no evaluator dependency)
- ✅ Clean separation of execution vs. object frames
- ✅ Simplified evaluator architecture
- ✅ Enhanced object portability

## Timeline and Resources

### Timeline

- **Phase 1**: 2 weeks (setup dual system)
- **Phase 2**: 2 weeks (migration to owned frames)
- **Phase 3**: 2 weeks (cleanup and optimization)
- **Total**: 6 weeks for complete implementation

### Resources Needed

- **Development**: 1 senior developer for architecture, 1 developer for implementation
- **Testing**: Comprehensive test coverage for all object operations
- **Documentation**: Update all architectural documentation
- **Performance**: Benchmarking and optimization work

## Conclusion

This refactor addresses a fundamental architectural issue: object coupling to evaluators. By moving frames into object ownership, we gain self-contained, portable objects while preserving all frame-based semantics.

The phased approach minimizes risk while delivering significant architectural improvements. The result will be a cleaner, more maintainable object system that better supports Viro's design goals.

**Recommendation**: Proceed with implementation. The benefits outweigh the complexity, and the phased approach manages risk effectively.
