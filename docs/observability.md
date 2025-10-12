# Observability Guide: Trace & Debug

**Viro Interpreter** - Program transparency and diagnostic tools

---

## Overview

Viro provides comprehensive observability features to help you understand, debug, and optimize your programs:

- **Tracing**: Record evaluation steps with structured JSON event logs
- **Debugging**: Set breakpoints, step through execution, inspect program state
- **Reflection**: Examine types, function specifications, and object contents

These tools are essential for understanding program behavior, diagnosing issues, and verifying correctness.

---

## Tracing

### What is Tracing?

Tracing records each evaluation step as a JSON event, including:
- Function calls and returns
- Native operations
- Evaluation timing
- Error occurrences

Trace events are written to rotating log files for later analysis.

### Enabling Trace

**Basic usage**:
```viro
trace --on
; Your code here
calculate-interest 100 0.05 3
trace --off
```

**With filters** (only trace specific functions):
```viro
trace --on --only ['calculate-interest 'compound-growth]
; Only calls to calculate-interest and compound-growth are logged
```

**Excluding functions**:
```viro
trace --on --exclude ['print 'type?]
; Everything except print and type? calls are logged
```

### Trace Output

**Default location**: `viro-trace.log` in the current directory

**Custom file**:
```viro
trace --on --file %logs/my-trace.log
```

**Note**: File paths are resolved relative to the sandbox root (see Ports Guide).

**Append mode** (don't rotate, append to existing file):
```viro
trace --on --file %logs/trace.log --append
```

### Trace File Format

Events are written as JSON lines:

```json
{"id":1,"timestamp":"2025-10-11T14:23:45.123Z","word":"calculate-interest","category":"eval","duration":"2.3ms","metadata":{"args":[100,0.05,3]}}
{"id":2,"timestamp":"2025-10-11T14:23:45.125Z","word":"*","category":"native","duration":"0.1ms","metadata":{"args":[100,0.05]}}
{"id":3,"timestamp":"2025-10-11T14:23:45.126Z","word":"calculate-interest","category":"eval","duration":"3.1ms","metadata":{"result":115.76}}
```

**Event fields**:
- `id`: Sequential event number
- `timestamp`: ISO 8601 timestamp
- `word`: Function or operation name
- `category`: Event type (eval, native, error, trace, debug)
- `duration`: Execution time (when available)
- `metadata`: Additional context (arguments, results, errors)

### Querying Trace Status

```viro
status: trace?
print status
; ==> object! with fields: enabled, filters, file, sequence-id
```

### File Rotation

**Default behavior**:
- Maximum file size: 50 MB
- Backup files: 5 (viro-trace.log.1, viro-trace.log.2, ...)
- Oldest backups automatically deleted

When the current log reaches 50 MB, it rotates to `.1`, previous `.1` becomes `.2`, etc.

### Performance Impact

- **Trace disabled**: <5% overhead (instrumentation checks only)
- **Trace enabled**: <25% overhead (includes JSON serialization and I/O)

For performance-critical code, disable tracing or use `--only` filters.

### Common Trace Patterns

**Debug production issues**:
```viro
trace --on --file %debug/issue-123.log
; Reproduce the issue
trace --off
```

**Profile slow operations**:
```viro
trace --on --only ['slow-function]
slow-function large-dataset
trace --off
; Analyze duration field in trace events
```

**Monitor specific subsystems**:
```viro
trace --on --only ['http-request 'parse-csv 'validate-data]
; Trace only data pipeline functions
```

---

## Debugging

### What is Debugging?

The debugger provides interactive control over program execution:
- Set breakpoints on functions
- Step through code line-by-line
- Inspect local variables and call stack
- Continue execution until next breakpoint

### Enabling Debug Mode

```viro
debug --on
```

The REPL prompt changes to indicate debug mode (implementation pending).

### Setting Breakpoints

**On function entry**:
```viro
debug --breakpoint 'calculate-interest
```

**With location** (future enhancement):
```viro
debug --breakpoint 'process-data 5  ; Break at index 5 in function body
```

**List breakpoints**:
Breakpoints are assigned IDs when created. Store the ID if you need to remove them:
```viro
bp-id: debug --breakpoint 'my-function
; Later...
debug --remove bp-id
```

### Stepping Through Code

When execution pauses at a breakpoint:

**Continue** (run until next breakpoint):
```viro
debug --continue
```

**Step into** (enter function calls):
```viro
debug --step
```

**Step over** (execute function without entering):
```viro
debug --next
```

**Step out** (finish current function and return):
```viro
debug --finish
```

### Inspecting State

**Local variables** (current frame):
```viro
locals: debug --locals
print locals
; ==> object! with word/value pairs from current frame
```

**Call stack**:
```viro
stack: debug --stack
print stack
; ==> block! of stack frames [word location word location ...]
```

### Disabling Debug Mode

```viro
debug --off
```

This clears all breakpoints and returns to normal execution.

### Debug Workflow Example

```viro
; 1. Enable debugger and set breakpoint
debug --on
debug --breakpoint 'calculate-tax

; 2. Run code - execution pauses at breakpoint
result: calculate-tax 1000 0.2

; 3. Inspect state (REPL is interactive at breakpoint)
>> debug --locals
; Shows: {amount: 1000, rate: 0.2, ...}

>> debug --stack
; Shows call stack

; 4. Step through execution
>> debug --step
; Executes one evaluation step

; 5. Continue or finish
>> debug --continue
; Runs until completion or next breakpoint

; 6. Disable when done
debug --off
```

### Conditional Breakpoints (Future)

Planned syntax:
```viro
debug --breakpoint 'process-order [:total > 1000]
; Only break when total > 1000
```

### Limitations

- **T153/T154 not implemented**: Breakpoint integration in evaluator and REPL debug mode prompt are pending
- Breakpoints currently work for function entry only (not arbitrary code locations)
- Debugger state is per-REPL session (not persisted)

---

## Reflection

### Type Inspection

**Get value type**:
```viro
print type-of 42           ; ==> integer!
print type-of "hello"      ; ==> string!
print type-of [1 2 3]      ; ==> block!
print type-of :my-function ; ==> function!
```

### Function Inspection

**Function specification** (parameters):
```viro
add-tax: fn [amount rate] [(* amount (+ 1 rate))]
print spec-of :add-tax
; ==> [amount rate]
```

**Function body**:
```viro
print body-of :add-tax
; ==> [(* amount (+ 1 rate))]
```

**Note**: `spec-of` and `body-of` return **deep copies** to protect runtime state. Modifying returned values doesn't affect the original function.

### Object Inspection

**Get field names**:
```viro
invoice: object [id: 42 customer: "Acme" total: decimal "199.99"]
print words-of invoice
; ==> [id customer total]
```

**Get field values**:
```viro
print values-of invoice
; ==> [42 "Acme" 199.99]
```

**Order guarantee**: `words-of` and `values-of` return fields in the same order.

### Source Formatting

**Reconstruct source** (with formatting):
```viro
print source :my-function
; ==> Formatted function definition
```

### Safety Guarantees

All reflection functions return **immutable snapshots**:
- `spec-of`, `body-of`: Deep copies of function data
- `words-of`, `values-of`: Copies of object fields
- `source`: Reconstructed representation

Reflection is read-only and cannot modify live program state.

---

## Combining Trace and Debug

### Trace-First Debugging

1. Run with tracing to identify problematic function
2. Analyze trace log to narrow down issue
3. Set breakpoint on suspect function
4. Use debugger to step through and inspect state

**Example**:
```viro
; Step 1: Enable trace with filters
trace --on --only ['process-order 'calculate-discount]
; Run problematic code
result: process-orders data
trace --off

; Step 2: Analyze trace file (external tool or manual inspection)
; Identify that calculate-discount is returning wrong value

; Step 3: Use debugger
debug --on
debug --breakpoint 'calculate-discount
result: calculate-discount 1000 0.15
; Inspect locals at breakpoint to find bug
debug --locals
debug --off
```

### Performance Debugging

```viro
; Trace to measure timing
trace --on
slow-operation
trace --off

; Analyze trace events to find bottlenecks
; Set breakpoints on slow functions
debug --on
debug --breakpoint 'bottleneck-function
slow-operation
; Step through to understand why it's slow
```

---

## Best Practices

### Tracing

**Do**:
- Use `--only` filters for focused tracing (reduces overhead and log size)
- Disable tracing in production unless debugging specific issues
- Rotate trace files regularly (automatic with default settings)
- Include trace logs when reporting bugs

**Don't**:
- Leave tracing on during performance benchmarks (unless measuring trace overhead)
- Trace high-frequency functions without filters (generates huge logs)
- Store trace files outside sandbox root (security risk)

### Debugging

**Do**:
- Set breakpoints on function entry points first
- Use `debug --locals` to inspect state before stepping
- Clear breakpoints when done (`debug --off`)
- Test small functions interactively before debugging larger programs

**Don't**:
- Set too many breakpoints (makes stepping tedious)
- Forget to disable debug mode (affects performance)
- Modify returned values from `debug --locals` expecting changes (they're copies)

### Reflection

**Do**:
- Use `type-of` to verify value types during development
- Use `spec-of` to document function signatures
- Use `words-of`/`values-of` for object introspection and validation
- Combine reflection with debugging for comprehensive inspection

**Don't**:
- Rely on reflection in performance-critical code (creates copies)
- Assume reflection mutates original values (all returns are snapshots)
- Use reflection as a substitute for proper documentation

---

## Troubleshooting

### Trace file not created

**Check**:
- File path is within sandbox root
- Sandbox root configured correctly (`--sandbox-root` flag)
- Write permissions on target directory

**Error**: `Access error: sandbox-violation`
- Trace file path must be under sandbox root
- Use `trace --on --file %relative/path.log` (relative to sandbox)

### Breakpoint not firing

**Check**:
- Debugger enabled (`debug --on`)
- Function name spelled correctly (case-sensitive)
- Function actually called during execution

**Note**: T153 (breakpoint integration) is not fully implemented. Some breakpoint scenarios may not work yet.

### High trace overhead

**Solutions**:
- Use `--only` filters to reduce event volume
- Exclude high-frequency functions (`--exclude`)
- Consider disabling trace for performance-critical sections

### Debug prompt not showing

**Note**: T154 (REPL debug mode prompt) is pending implementation. The debugger functionality exists but the prompt indicator is not yet visible.

---

## Related Documentation

- **REPL Usage**: `docs/repl-usage.md` - Interactive programming guide
- **Ports Guide**: `docs/ports-guide.md` - Sandbox configuration for trace files
- **Contracts**: `specs/002-implement-deferred-features/contracts/trace-debug.md` - Technical specifications
- **Contracts**: `specs/002-implement-deferred-features/contracts/reflection.md` - Reflection specifications

---

## Implementation Status

**Feature 002 - User Story 5**: Mostly complete

**Implemented**:
- ✅ `trace --on/--off` with filters and file configuration
- ✅ `trace?` status query
- ✅ Trace event JSON serialization
- ✅ File rotation (lumberjack integration)
- ✅ `debug --on/--off/--breakpoint/--remove`
- ✅ `debug --step/--next/--finish/--continue`
- ✅ `debug --locals/--stack`
- ✅ All reflection natives (`type-of`, `spec-of`, `body-of`, `words-of`, `values-of`, `source`)

**Pending**:
- ⚠️ T153: Breakpoint checks in evaluator dispatch
- ⚠️ T154: REPL debug mode prompt indicator

See `specs/002-implement-deferred-features/tasks.md` for details.

---

Enjoy transparent, debuggable Viro programs! 🔍
