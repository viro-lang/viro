# Plan 015: Enhanced Trace System for LLM-Friendly Debugging

## Problem Statement

LLM agents debugging Viro code need comprehensive execution traces that include:

- Every expression evaluated
- Frame state (local variables) at each step
- Call stack depth and context
- Function arguments and return values
- Clear expression flow through nested blocks

**Current trace limitations:**

- Only traces function calls, not individual expressions
- No frame state in trace output
- No call stack depth tracking
- Limited context about expression evaluation
- Arguments/results not consistently traced

**LLM workflow for debugging:**

1. Write/modify Viro code
2. Run with enhanced tracing
3. Parse JSON trace output
4. Analyze execution flow to find bugs
5. Fix and iterate

## Research Summary

### Current Trace Infrastructure

**TraceEvent Structure** (`internal/trace/trace.go:44-49`):

```go
type TraceEvent struct {
    Timestamp time.Time `json:"timestamp"`
    Value     string    `json:"value"`     // Result value
    Word      string    `json:"word"`      // Word being evaluated
    Duration  int64     `json:"duration"`  // Nanoseconds
}
```

**Existing Trace Emission Points**:

- `evaluator.go:459-467` - Function calls only (when word resolves to function)
- `evaluator.go:440-447` - Breakpoint hits
- Various port operations (open, read, write, close)
- Object operations (create, field read/write)

**What's Missing:**

- ❌ Expression-level tracing (literals, set-words, paths, parens, etc.)
- ❌ Frame state not captured
- ❌ Call stack depth not tracked
- ❌ Function arguments not logged
- ❌ Block entry/exit not traced
- ❌ No execution step counter

### Design Considerations

**Performance Impact:**

- Trace only when enabled (already has this check)
- Minimize allocations in hot path
- Make frame serialization optional (can be expensive)
- Use sync.Pool for event objects if needed

**Output Format:**

- Keep line-delimited JSON for streaming
- Add optional "verbose" mode for frame state
- Structured format easy for LLMs to parse
- Human-readable when printed

**Backward Compatibility:**

- Keep existing `TraceEvent` structure
- Add new fields as optional
- New fields are omitted from JSON when empty (using `omitempty`)
- Existing trace consumers unaffected

## Proposed Solution

### Phase 1: Enhanced TraceEvent Structure (1-2 hours)

**Extend TraceEvent** to include debugging information:

```go
// internal/trace/trace.go
type TraceEvent struct {
    // Existing fields
    Timestamp time.Time `json:"timestamp"`
    Value     string    `json:"value"`
    Word      string    `json:"word"`
    Duration  int64     `json:"duration"`

    // New fields for debugging (all optional via omitempty)
    EventType  string            `json:"event_type,omitempty"`   // "eval", "call", "return", "block-enter", "block-exit"
    Step       int64             `json:"step,omitempty"`         // Execution step counter
    Depth      int               `json:"depth,omitempty"`        // Call stack depth
    Position   int               `json:"position,omitempty"`     // Position in current block
    Expression string            `json:"expression,omitempty"`   // Mold of expression being evaluated
    Args       map[string]string `json:"args,omitempty"`         // Function arguments (name -> value)
    Frame      map[string]string `json:"frame,omitempty"`        // Local variables (only in verbose mode)
    ParentExpr string            `json:"parent_expr,omitempty"`  // Context expression
    Error      string            `json:"error,omitempty"`        // Error message if evaluation failed
}
```

**Add TraceFilters options**:

```go
// internal/trace/trace.go
type TraceFilters struct {
    IncludeWords []string
    ExcludeWords []string
    MinDuration  time.Duration

    // New options for debugging
    Verbose      bool  // Include frame state
    StepLevel    int   // 0=calls only, 1=expressions, 2=all
    IncludeArgs  bool  // Include function arguments
    MaxDepth     int   // Only trace up to this call depth (0=unlimited)
}
```

**Add step counter to TraceSession**:

```go
// internal/trace/trace.go
type TraceSession struct {
    mu       sync.Mutex
    enabled  bool
    sink     io.Writer
    logger   *lumberjack.Logger
    filters  TraceFilters
    stepCounter int64  // Monotonic step counter
}

func (ts *TraceSession) NextStep() int64 {
    ts.mu.Lock()
    defer ts.mu.Unlock()
    ts.stepCounter++
    return ts.stepCounter
}

func (ts *TraceSession) ResetStepCounter() {
    ts.mu.Lock()
    defer ts.mu.Unlock()
    ts.stepCounter = 0
}
```

**Tests:**

- Test TraceEvent JSON serialization with new fields
- Test backward compatibility (old fields still work)
- Test omitempty behavior (empty fields not in JSON)
- Test step counter increment

### Phase 2: Expression-Level Tracing (3-4 hours)

Add trace emission for ALL expression types in evaluator. See detailed code in implementation section.

**Key changes:**

- Add tracing to `evaluateElement()` for all expression types
- Add tracing to `invokeFunctionExpression()` for calls/returns
- Add tracing to `DoBlock()` for block enter/exit
- Add helper methods: `captureFrameState()`, `emitTraceResult()`, `captureFunctionArgs()`

**Tests:**

- Test tracing for each expression type (literal, setword, word, paren, block, path, etc.)
- Test step counter increments correctly
- Test depth tracking through nested calls
- Test frame state capture in verbose mode
- Test function arguments capture
- Test block enter/exit events

### Phase 3: Enhanced Trace Native Commands (2-3 hours)

**Add new refinements to `trace` native**:

```viro
trace --on --verbose --step-level 1 --include-args --max-depth 5
```

**New refinements:**

- `--verbose` - Include frame state in trace events
- `--step-level N` - Control granularity (0=calls, 1=expressions, 2=all)
- `--include-args` - Include function arguments
- `--max-depth N` - Limit trace depth

**Tests:**

- Test --verbose flag enables frame state
- Test --step-level controls granularity
- Test --include-args captures function arguments
- Test --max-depth limits trace depth
- Test step counter reset on trace --on

### Phase 4: Documentation and Examples (1-2 hours)

**Create debugging guide** (`docs/debugging-guide.md`):

- Quick start examples
- Trace output format specification
- Event types description
- Trace levels explanation
- Common debugging patterns
- Performance notes

**Create examples** (`docs/debugging-examples.md`):

- Finding recursive function bugs
- Tracking variable changes
- Example trace outputs with analysis

**Tests:**

- Verify documentation examples produce expected output
- Test parsing examples work with actual trace output

### Phase 5: LLM Instructions File (1 hour)

**Create `.github/instructions/debugging-with-trace.instruction.md`**

This file will be automatically used by GitHub Copilot and other LLM tools when debugging Viro code.

Content includes:

- Quick reference for trace commands
- Complete debugging workflow (5 steps)
- JSON format specification with all fields explained
- Common debugging patterns (5 patterns with examples)
- Parsing examples in Python and JavaScript
- Best practices for LLM debugging (8 practices)
- Troubleshooting guide
- Real debugging session examples

Key sections:

1. **Quick Reference** - Common trace commands
2. **Debugging Workflow** - 5-step process
3. **Trace Format** - JSON field descriptions
4. **Common Patterns** - Infinite recursion, variable tracking, performance, etc.
5. **Parsing Code** - Python and JavaScript examples
6. **Best Practices** - 8 specific practices for LLMs
7. **Troubleshooting** - Common issues and solutions
8. **Real Examples** - Actual debugging sessions with analysis

**Update `AGENTS.md`** to reference the new debugging instructions file in a "Debugging" section.

## File Changes Summary

### New Files

1. **`docs/debugging-guide.md`** - User-facing debugging guide
2. **`docs/debugging-examples.md`** - Example debugging scenarios
3. **`.github/instructions/debugging-with-trace.instruction.md`** - LLM-specific instructions
4. **`test/contract/trace_enhanced_test.go`** - Contract tests

### Modified Files

1. **`internal/trace/trace.go`**
   - Add fields to `TraceEvent`: EventType, Step, Depth, Position, Expression, Args, Frame, ParentExpr, Error
   - Add fields to `TraceFilters`: Verbose, StepLevel, IncludeArgs, MaxDepth
   - Add `stepCounter` to `TraceSession`
   - Add methods: `NextStep()`, `ResetStepCounter()`, `GetVerbose()`, `GetIncludeArgs()`, `ShouldTraceExpression()`, `ShouldTraceAtDepth()`

2. **`internal/eval/evaluator.go`**
   - Add trace emission to `evaluateElement()` for all expression types
   - Add trace emission to `invokeFunctionExpression()` for calls/returns
   - Add trace emission to `DoBlock()` for block enter/exit
   - Add methods: `captureFrameState()`, `emitTraceResult()`, `captureFunctionArgs()`

3. **`internal/native/control.go`**
   - Add refinement handling in `Trace()`: --verbose, --step-level, --include-args, --max-depth
   - Add step counter reset on trace --on

4. **`internal/native/register_control.go`**
   - Update `trace` native ParamSpec to include new refinements

## Implementation Phases

### Phase 1: Core Infrastructure (1-2 hours)

- Extend TraceEvent structure
- Add TraceFilters options
- Add step counter to TraceSession
- Write tests for new structures

### Phase 2: Expression Tracing (3-4 hours)

- Add trace hooks to evaluateElement()
- Add trace hooks to invokeFunctionExpression()
- Add trace hooks to DoBlock()
- Implement helper methods
- Write tests for all expression types

### Phase 3: Native Commands (2-3 hours)

- Add new refinements to trace native
- Update registration
- Add helper methods to TraceSession
- Write tests for new refinements

### Phase 4: Documentation (1-2 hours)

- Write debugging guide
- Write examples
- Test all examples

### Phase 5: LLM Instructions (1 hour)

- Write comprehensive LLM instructions
- Include parsing examples
- Add troubleshooting guide
- Test with real trace output

**Total Estimated Effort: 8-12 hours**

## Success Criteria

### Functional Requirements

- ✅ Trace every expression evaluation (when step-level >= 1)
- ✅ Capture frame state in verbose mode
- ✅ Track call stack depth
- ✅ Log function arguments
- ✅ Emit structured JSON events
- ✅ Support filtering by word, depth, level
- ✅ Backward compatible with existing traces

### Performance Requirements

- ✅ Minimal overhead when tracing disabled (existing check)
- ✅ Reasonable overhead in non-verbose mode (<10% slowdown)
- ✅ Verbose mode acceptable for debugging (may be slower)

### Usability Requirements

- ✅ Simple API for LLMs: `trace --on --verbose`
- ✅ Parseable JSON output
- ✅ Clear documentation with examples
- ✅ Works with existing REPL

## Example Usage

### Basic Tracing (LLM Workflow)

```viro
; Step 1: Enable enhanced tracing
trace --on --verbose --include-args --step-level 1

; Step 2: Run code
fact: fn [n] [
    if (= n 0) [1] [
        (* n (fact (- n 1)))
    ]
]
result: fact 3

; Step 3: Disable tracing
trace --off
```

**Trace Output** (stderr or file):

```json
{"timestamp":"2024-01-15T10:30:00.123Z","event_type":"setword","step":1,"depth":0,"word":"fact","value":"function[fact]"}
{"timestamp":"2024-01-15T10:30:00.124Z","event_type":"call","step":2,"depth":1,"word":"fact","args":{"n":"3"},"frame":{"n":"3"}}
{"timestamp":"2024-01-15T10:30:00.125Z","event_type":"call","step":3,"depth":2,"word":"if"}
{"timestamp":"2024-01-15T10:30:00.126Z","event_type":"call","step":4,"depth":3,"word":"=","value":"false"}
...
```

**LLM Analysis**:

- Parse JSON lines
- Build execution tree from depth field
- Track variable changes via frame field
- Identify performance bottlenecks via duration
- Find bugs by comparing expected vs actual values

### Focused Tracing

```viro
; Only trace specific function
trace --on --only [calculate-interest] --include-args
; Run complex program
trace --off
```

### Performance Monitoring

```viro
; Calls only, no frame state
trace --on --step-level 0
; Run program
trace --off
```

## Testing Strategy

### Unit Tests (`test/contract/trace_enhanced_test.go`)

- Test TraceEvent serialization with new fields
- Test expression-level tracing for all types
- Test frame state capture in verbose mode
- Test all new refinements
- Test step counter behavior

### Integration Tests

- Test full program traces
- Test recursive function tracing
- Test error tracing
- Verify trace output structure

## Risk Assessment

### Low Risk

- Extends existing trace system
- Backward compatible (new fields are optional)
- Feature-flagged (only when trace enabled)
- No changes to core evaluation logic (just hooks)

### Mitigation

- Comprehensive testing for all expression types
- Performance benchmarks for verbose mode
- Clear documentation for LLM consumers
- Gradual rollout (phase by phase)

## Future Enhancements

- Compressed binary trace format for large programs
- Trace replay/visualization tools
- Statistical analysis of trace data
- Integration with profiling tools
- DAP (Debug Adapter Protocol) support
- Trace diff between runs
- Automatic bug detection heuristics

## Dependencies

- Existing trace system (`internal/trace`)
- Existing evaluator (`internal/eval`)
- JSON serialization (`encoding/json`)
- No new external dependencies

## Backward Compatibility

- ✅ Existing TraceEvent fields unchanged
- ✅ New fields use `omitempty` tag
- ✅ Existing trace consumers unaffected
- ✅ Default behavior unchanged (calls only)
- ✅ Opt-in via new refinements

## Notes for Implementation

1. **Start with Phase 1** - Get data structures right first
2. **Test thoroughly** - Trace is critical for debugging
3. **Optimize later** - Correctness over performance initially
4. **Document as you go** - Keep examples up to date
5. **Get LLM feedback** - Test with actual LLM debugging workflows
