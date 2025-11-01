# Plan 024: Implement Interactive Debugging

## Problem Statement

The current debugging system in Viro has basic infrastructure but lacks actual interactive debugging capabilities. While debug commands exist and breakpoints can be set, there's no way to:

1. **Step through code interactively** - The evaluator doesn't pause during stepping
2. **Inspect frame state** - `--locals` and `--stack` return empty/placeholder data  
3. **Use interactive debug REPL** - No actual debug session with commands like `next`, `continue`, `locals`
4. **Control execution flow** - Step commands exist but don't actually pause execution

This makes debugging Viro programs difficult and not user-friendly.

## Current State Analysis

### What's Working ✅
- Basic debugger infrastructure (`internal/debug/debug.go`)
- Debug native commands with proper signatures
- Breakpoint management (set/remove by word/ID)
- Debug mode states (off/active/stepping)
- REPL debug prompt support (`[debug] >> `)
- Trace system for observability

### What's Missing ❌
- **Step execution**: Evaluator doesn't pause after expressions when stepping
- **Frame inspection**: `--locals` returns empty object, `--stack` returns empty block
- **Interactive REPL**: No actual debug session - commands just set flags
- **Execution control**: No way to resume/step/continue from breakpoints
- **State inspection**: No way to examine variables during execution

## Implementation Plan

### Phase 1: Core Step Execution (3-4 hours)

**Goal**: Make the evaluator actually pause during step-by-step execution

**Changes**:

1. **Extend Debugger with Step State** (`internal/debug/debug.go`):
   ```go
   type StepState struct {
       Paused      bool
       WaitChan    chan struct{}  // Channel to pause/resume execution
       CurrentExpr core.Value     // Current expression being evaluated
       CurrentPos  int            // Position in current block
       FrameIndex  int            // Current frame index
   }
   
   func (d *Debugger) PauseExecution(expr core.Value, pos, frameIdx int)
   func (d *Debugger) ResumeExecution()
   func (d *Debugger) IsPaused() bool
   ```

2. **Add Step Hook to Evaluator** (`internal/eval/evaluator.go`):
   ```go
   // In EvaluateExpression, before evaluating each expression:
   if debug.GlobalDebugger != nil && debug.GlobalDebugger.ShouldPause() {
       debug.GlobalDebugger.PauseExecution(block[position], position, e.currentFrameIndex())
   }
   ```

3. **Update Debug Native** (`internal/native/control.go`):
   - Make `--step` actually trigger stepping instead of just setting flags
   - Add coordination between debug commands and evaluator pausing

**Tests**:
- Test evaluator pauses at each expression when stepping
- Test resume functionality works correctly
- Test stepping through nested blocks

### Phase 2: Frame Inspection Implementation (2-3 hours)

**Goal**: Make `--locals` and `--stack` return actual data

**Changes**:

1. **Add Frame Inspection to Debugger** (`internal/debug/debug.go`):
   ```go
   func (d *Debugger) GetFrameLocals(eval core.Evaluator, frameIdx int) map[string]core.Value
   func (d *Debugger) GetCallStack(eval core.Evaluator) []string
   ```

2. **Update Debug Native** (`internal/native/control.go`):
   - `--locals`: Return object with current frame bindings
   - `--stack`: Return block with call stack entries

**Tests**:
- Test locals returns correct variables for current frame
- Test stack returns proper call hierarchy
- Test inspection at different call depths

### Phase 3: Interactive Debug REPL (3-4 hours)

**Goal**: Create actual interactive debugging session

**Changes**:

1. **Add Debug REPL Mode** (`internal/repl/repl.go`):
   ```go
   type DebugSession struct {
       active    bool
       evaluator core.Evaluator
       debugger  *debug.Debugger
   }
   
   func (r *REPL) EnterDebugMode() error
   func (r *REPL) ExitDebugMode()
   func (r *REPL) HandleDebugCommand(cmd string) (bool, error)  // true = continue in debug mode
   ```

2. **Debug Commands**:
   - `n`/`next`: Step to next expression
   - `c`/`continue`: Continue until breakpoint
   - `l`/`locals`: Show local variables
   - `s`/`stack`: Show call stack
   - `p expr`/`print expr`: Evaluate expression
   - `q`/`quit`: Exit debug mode

3. **Enhanced Prompt**:
   ```
   [debug:5@fact] >> n
   => result: (* n (fact (- n 1)))
   [debug:6@fact] >> l
   { n: 2, result: none }
   ```

**Tests**:
- Test entering/exiting debug mode
- Test all debug commands work correctly
- Test prompt shows correct position/function

### Phase 4: Integration and Polish (2-3 hours)

**Goal**: Tie everything together and add finishing touches

**Changes**:

1. **Evaluator-Debugger Integration**:
   - Proper coordination between stepping and breakpoints
   - Handle nested function calls correctly
   - Maintain execution state across steps

2. **Add Debug-Eval Command**:
   - Evaluate expressions in current frame context
   - Useful for inspecting complex expressions

3. **Error Handling**:
   - Proper error recovery in debug mode
   - Clear error messages for invalid debug commands

**Tests**:
- Integration tests with real Viro programs
- Test error cases and recovery
- Test complex debugging scenarios

## File Changes Summary

### Modified Files
- `internal/debug/debug.go` - Add step state, frame inspection, pausing logic
- `internal/eval/evaluator.go` - Add step hooks and pausing coordination
- `internal/repl/repl.go` - Add interactive debug REPL mode
- `internal/native/control.go` - Implement actual --locals/--stack functionality

### New Files
- `test/contract/debug_interactive_test.go` - Contract tests for interactive debugging

## Success Criteria

1. **Functional**:
   - ✅ Can step through code one expression at a time
   - ✅ Can inspect local variables at each step
   - ✅ Can view call stack during execution
   - ✅ Can set breakpoints and continue to them
   - ✅ Interactive debug REPL works with all commands

2. **Usability**:
   - ✅ Clear visual feedback in REPL
   - ✅ Simple commands for debugging
   - ✅ Proper error handling and recovery

3. **Testing**:
   - ✅ All debug functionality has contract tests
   - ✅ Integration tests with real programs
   - ✅ Edge cases handled properly

## Example Usage

### Interactive Debugging Session
```
>> debug --on
>> fact: fn [n] [if (= n 0) [1] [(* n (fact (- n 1)))]]
>> debug --breakpoint 'fact
>> fact 3
Breakpoint hit at 'fact
[debug:0@fact] >> l
{ n: 3 }
[debug:0@fact] >> n
=> if (= n 0) [1] [(* n (fact (- n 1)))]
[debug:1@fact] >> n
=> (= n 0)
[debug:2@fact] >> p n
3
[debug:2@fact] >> c
6
>>
```

### Programmatic Debugging
```viro
debug --on
debug --breakpoint 'fact
result: fact 3  ; Will pause at breakpoint
debug --locals  ; { n: 3 }
debug --step
debug --locals  ; { n: 3, result: none }
debug --continue
```

## Dependencies

- Existing debug infrastructure (`internal/debug`)
- Evaluator frame system (`internal/eval`, `internal/frame`)
- REPL infrastructure (`internal/repl`)
- Native function system (`internal/native`)

## Estimated Effort

- Phase 1: 3-4 hours (step execution)
- Phase 2: 2-3 hours (frame inspection) 
- Phase 3: 3-4 hours (interactive REPL)
- Phase 4: 2-3 hours (integration and polish)
- Testing: 3-4 hours

**Total: 13-18 hours**

## Risk Assessment

**Medium Risk**:
- Changes to evaluator execution flow could introduce bugs
- Complex coordination between evaluator, debugger, and REPL
- Need careful handling of concurrent execution states

**Mitigation**:
- Incremental implementation with thorough testing at each phase
- Keep existing evaluation path unchanged when debugging disabled
- Comprehensive contract and integration tests
- Feature can be disabled if issues arise

## Future Enhancements

- Conditional breakpoints
- Watchpoints (break on variable change)
- Reverse debugging (limited history)
- Debug script recording/playback
- DAP (Debug Adapter Protocol) support
