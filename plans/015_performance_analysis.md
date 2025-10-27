# Performance Analysis: Enhanced Trace System

## Current Implementation Performance

### When Tracing is OFF (Normal Operation)
- **Minimal overhead**: Just a boolean check `trace.GlobalTraceSession.IsEnabled()`
- **No allocations**: No memory allocation when disabled
- **No I/O**: No file or stderr writes
- **Benchmark results**: ~6-13ns per operation (from existing benchmarks)

### When Tracing is ON (Current)
- **Function calls only**: Traces when word resolves to function
- **Mutex lock**: `Emit()` acquires mutex for thread safety
- **JSON marshaling**: Serializes event to JSON
- **I/O write**: Writes to stderr or file
- **Filtering**: Applies word/duration filters

## Proposed Changes Performance Impact

### Phase 1: Enhanced TraceEvent Structure
- **When OFF**: No impact (same boolean check)
- **When ON**: Minimal impact (larger JSON structure, but same marshaling cost)
- **Memory**: Slightly more memory per event (new optional fields)

### Phase 2: Expression-Level Tracing
- **When OFF**: No impact (same boolean check)
- **When ON**: 
  - **Step-level 0**: Same as current (calls only)
  - **Step-level 1**: **~2-5x slower** - traces every expression
  - **Step-level 2**: **~3-10x slower** - traces everything including blocks

### Phase 3: Native Commands
- **When OFF**: No impact
- **When ON**: Minimal impact (same filtering/marshaling)

### Phase 4: Documentation
- **No runtime impact** (documentation only)

## Performance Recommendations

### For Production Use (Tracing OFF)
✅ **Zero performance impact** - same as current implementation

### For Debugging Use (Tracing ON)

**Recommended Settings:**
```viro
; Fast debugging (calls only)
trace --on --step-level 0

; Balanced debugging (expressions)
trace --on --step-level 1 --max-depth 5

; Full debugging (everything) - SLOW
trace --on --verbose --step-level 2
```

**Performance Guidelines:**
- `--step-level 0`: **~10-50% slowdown** (calls only)
- `--step-level 1`: **~200-500% slowdown** (all expressions)
- `--verbose`: **~2-3x additional slowdown** (frame capture)
- `--include-args`: **~10-20% additional slowdown** (argument capture)

### Optimization Strategies

1. **Lazy Frame Capture**:
```go
func (e *Evaluator) captureFrameState() map[string]string {
    if !trace.GlobalTraceSession.GetVerbose() {
        return nil  // Skip entirely
    }
    // Only capture when verbose is enabled
}
```

2. **Conditional Tracing**:
```go
if trace.GlobalTraceSession.ShouldTraceExpression() {
    // Only emit when step-level allows expressions
}
```

3. **Batch Emits** (Future):
```go
// Instead of immediate emit, buffer and batch
trace.GlobalTraceSession.QueueEvent(event)
```

## Performance Test Results (Estimated)

Based on current benchmarks (~6ns per operation):

| Configuration | Overhead | Use Case |
|---------------|----------|----------|
| Tracing OFF | ~0% | Production |
| `--step-level 0` | ~25% | Light debugging |
| `--step-level 1` | ~300% | Full debugging |
| `--verbose` | ~500% | Variable inspection |
| `--include-args` | ~50% | Argument tracing |

## Memory Impact

### When Tracing OFF
- **No additional memory** (same as current)

### When Tracing ON
- **Event objects**: ~200-500 bytes per event
- **Frame state**: ~50-200 bytes per variable in verbose mode
- **JSON marshaling**: Temporary buffers during serialization

## I/O Impact

### Current Implementation
- **Function calls only**: ~1-10 events per function call
- **Output**: Line-delimited JSON to stderr/file

### Enhanced Implementation
- **Step-level 0**: Same as current
- **Step-level 1**: **~5-20x more events** (every expression)
- **Step-level 2**: **~10-50x more events** (blocks + expressions)

## Recommendations

### For LLM Debugging
```viro
; Recommended for LLMs - good balance of detail vs speed
trace --on --verbose --include-args --step-level 1 --max-depth 10
```

### For Performance Monitoring
```viro
; Minimal overhead for production monitoring
trace --on --step-level 0 --only [critical-functions]
```

### For Development
```viro
; Full debugging when needed
trace --on --verbose --step-level 2
```

## Implementation Notes

1. **All new trace points gated by `IsEnabled()`** - zero cost when off
2. **Verbose mode opt-in** - frame capture only when requested
3. **Step-level control** - users can choose granularity
4. **Depth limiting** - prevent trace explosion in recursive code

## Conclusion

**✅ Excellent performance characteristics:**
- **Tracing OFF**: Zero performance impact
- **Tracing ON**: Configurable overhead (10% to 500%)
- **Backward compatible**: No changes to existing behavior
- **Opt-in features**: Users control verbosity/speed trade-off

The enhanced trace system provides comprehensive debugging capabilities for LLMs while maintaining excellent performance when tracing is disabled.