# Plan 015: Enhanced Trace System - Quick Summary

## What This Plan Does

Extends Viro's existing trace system to provide comprehensive, LLM-friendly debugging output.

## Key Features

### 1. Expression-Level Tracing
- Trace every expression (not just function calls)
- Track literals, setwords, paths, parens, blocks, etc.

### 2. Rich Debug Information
```json
{
  "event_type": "call",
  "step": 42,
  "depth": 2,
  "word": "fact",
  "expression": "fact 3",
  "args": {"n": "3"},
  "frame": {"n": "3"},
  "value": "6",
  "duration": 1500000
}
```

### 3. Simple API for LLMs
```viro
trace --on --verbose --include-args --step-level 1
; Run your code
trace --off
```

### 4. LLM-Specific Documentation
- `.github/instructions/debugging-with-trace.instruction.md`
- Complete workflow guide
- Parsing examples (Python & JavaScript)
- Common debugging patterns
- Troubleshooting guide

## LLM Debugging Workflow

1. **Enable**: `trace --on --verbose --include-args --step-level 1`
2. **Run**: Execute Viro code
3. **Disable**: `trace --off`
4. **Parse**: Read JSON from stderr
5. **Analyze**: Find bugs using step-by-step trace

## Implementation Effort

- **Total**: 8-12 hours
- **Phase 1**: Core infrastructure (1-2h)
- **Phase 2**: Expression tracing (3-4h)
- **Phase 3**: Native commands (2-3h)
- **Phase 4**: Documentation (1-2h)
- **Phase 5**: LLM instructions (1h)

## Why This Approach?

✅ **Non-interactive** - Perfect for LLM batch processing  
✅ **Extends existing** - Builds on current trace system  
✅ **Backward compatible** - No breaking changes  
✅ **Structured output** - Easy JSON parsing  
✅ **Comprehensive** - Captures everything LLMs need  

## Files Created/Modified

### New Files
- `docs/debugging-guide.md`
- `docs/debugging-examples.md`
- `.github/instructions/debugging-with-trace.instruction.md` ⭐
- `test/contract/trace_enhanced_test.go`

### Modified Files
- `internal/trace/trace.go`
- `internal/eval/evaluator.go`
- `internal/native/control.go`
- `internal/native/register_control.go`

## Next Steps

1. Start with Phase 1 (data structures)
2. Implement Phase 2 (tracing hooks)
3. Add Phase 3 (native commands)
4. Write Phase 4 (documentation)
5. Create Phase 5 (LLM instructions)
6. Test thoroughly with real debugging scenarios

See full plan: `015_enhanced_trace_for_llm_debugging.md`
