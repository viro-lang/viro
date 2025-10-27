# Plan 014: LLM-Friendly Step-by-Step Debugger for Viro

## Problem Statement

When debugging Viro programs, LLM agents need a way to:
1. Execute code step-by-step
2. Inspect frame state (local variables, parent chain)
3. Inspect the call stack
4. Examine values at each step
5. Control execution flow (step, continue, finish)

Current state:
- Basic debugger exists (`debug.go`) with breakpoints and stepping modes
- Trace system exists (`trace.go`) for observability
- No interactive step-by-step execution for debugging
- No easy way to inspect frame contents during execution
- Debugger is command-based but not interactive

## Research Summary

### Existing Infrastructure

**Evaluator Architecture** (`internal/eval/evaluator.go`):
- `EvaluateExpression()` - evaluates single expressions from a block
- `DoBlock()` - evaluates entire blocks
- Frame chain via `Frames` slice with parent indices
- Call stack via `callStack` string array
- Support for tracing via `trace.GlobalTraceSession`
- Support for breakpoints via `debug.GlobalDebugger.HasBreakpoint()`

**Debug System** (`internal/debug/debug.go`):
- Three modes: `DebugModeOff`, `DebugModeActive`, `DebugModeStepping`
- Breakpoint management by word name
- Stepping control (`EnableStepping`, `DisableStepping`)
- No frame inspection capabilities

**Trace System** (`internal/trace/trace.go`):
- Structured event emission (JSON format)
- Filtering by word patterns
- Records: timestamp, word, value, duration
- File output with rotation

**Frame System** (`internal/frame/frame.go`):
- Parallel arrays: `Words` and `Values`
- Parent index for lexical scoping
- Types: FunctionArgs, Closure, Object, TypeFrame
- `GetAll()` returns all bindings as `[]core.Binding`

**REPL** (`internal/repl/repl.go`):
- Existing debug prompt: `[debug] >> `
- Multi-line input support
- History management
- Error recovery

### Design Options Evaluated

#### Option 1: Modify Evaluator for Step Mode
**Approach**: Add hooks to `EvaluateExpression()` to pause after each step
- ✅ Minimal changes to existing code
- ✅ Works with all expression types
- ❌ Requires evaluator state machine
- ❌ Complex to handle nested blocks

#### Option 2: Interpreter Pattern with Visitor
**Approach**: Create separate step-by-step interpreter
- ✅ Clean separation of concerns
- ✅ Easy to add step control
- ❌ Duplicates evaluation logic
- ❌ Large refactoring effort

#### Option 3: Enhanced Debug Native with REPL Integration ⭐ RECOMMENDED
**Approach**: Extend existing `debug` native with interactive step mode
- ✅ Leverages existing REPL infrastructure
- ✅ Uses existing debug/trace systems
- ✅ Minimal changes to evaluator
- ✅ LLM can use native Viro commands
- ✅ Works in both interactive and scripted modes

### Recommended Approach

**Hybrid approach combining:**
1. **Step callback in evaluator** - Minimal hook for step notifications
2. **Enhanced debug native** - Commands for step control and inspection
3. **REPL integration** - Interactive debugging session
4. **Programmatic API** - For LLM agent use

## Implementation Plan

### Phase 1: Core Step Execution (2-3 hours)

**Goal**: Add ability to execute code one expression at a time with state inspection

**Changes**:

1. **Extend Debug Session** (`internal/debug/debug.go`):
   ```go
   type StepState struct {
       Paused        bool
       CurrentBlock  []core.Value
       CurrentPos    int
       StepCallback  func(StepInfo)
   }
   
   type StepInfo struct {
       Position      int
       Expression    core.Value
       FrameIndex    int
       CallDepth     int
   }
   ```

2. **Add Step Hook to Evaluator** (`internal/eval/evaluator.go`):
   ```go
   // In EvaluateExpression, before evaluation:
   if debug.GlobalDebugger != nil && debug.GlobalDebugger.IsStepping() {
       debug.GlobalDebugger.OnStep(StepInfo{
           Position:   position,
           Expression: block[position],
           FrameIndex: e.currentFrameIndex(),
           CallDepth:  len(e.callStack),
       })
   }
   ```

3. **Add Frame Inspection** (`internal/debug/debug.go`):
   ```go
   func (d *Debugger) GetCurrentFrame(eval core.Evaluator) core.Frame
   func (d *Debugger) GetFrameBindings(frame core.Frame) []core.Binding
   func (d *Debugger) GetCallStack(eval core.Evaluator) []string
   ```

**Tests**:
- Test step callback is invoked for each expression
- Test frame inspection returns correct bindings
- Test call stack depth tracking

### Phase 2: Debug Command Extensions (2-3 hours)

**Goal**: Add commands for step-by-step debugging

**New Native Functions**:

1. **`debug-step`** - Execute one expression and pause
   ```viro
   debug --on
   debug-step  ; executes next expression
   ```

2. **`debug-inspect`** - Show current execution state
   ```viro
   debug-inspect  
   ; Returns object with:
   ; { position: 5, expression: "x + 1", frame: {...}, stack: [...] }
   ```

3. **`debug-frame`** - Get frame variables with optional level
   ```viro
   debug-frame      ; current frame
   debug-frame 1    ; parent frame
   debug-frame -1   ; all frames
   ```

4. **`debug-eval`** - Evaluate expression in current frame context
   ```viro
   debug-eval [x + 1]  ; evaluate in current frame
   ```

**Enhanced `debug` Native Refinements**:
- `debug --step` - Single step
- `debug --continue` - Run until next breakpoint
- `debug --finish` - Run until function returns
- `debug --inspect` - Show current state
- `debug --frame [level]` - Show frame at level

**Tests**:
- Test each debug command in isolation
- Test stepping through nested blocks
- Test frame inspection at different call depths
- Test debug-eval with local variables

### Phase 3: Interactive Debug REPL (2-3 hours)

**Goal**: Provide interactive debugging session within REPL

**Changes**:

1. **Debug REPL Mode** (`internal/repl/repl.go`):
   ```go
   func (r *REPL) EnterDebugMode(code string) error
   func (r *REPL) ExitDebugMode()
   func (r *REPL) IsInDebugMode() bool
   ```

2. **Debug Commands**:
   - `n` or `next` - step to next expression
   - `c` or `continue` - continue execution
   - `l` or `locals` - show local variables
   - `s` or `stack` - show call stack
   - `f` or `frame N` - show frame at level N
   - `p expr` or `print expr` - evaluate expression
   - `q` or `quit` - exit debug mode

3. **Visual Feedback**:
   ```
   [debug:3] >> n
   => x: 10
   [debug:4] >> l
   { x: 10, y: 20 }
   [debug:4] >> p x + y
   30
   [debug:4] >> c
   => 30
   >>
   ```

**Tests**:
- Test entering/exiting debug mode
- Test debug commands in interactive session
- Test visual feedback for each command
- Test error handling in debug mode

### Phase 4: LLM-Friendly API (1-2 hours)

**Goal**: Make debugging easy for LLM agents via simple commands

**Programmatic Interface**:

1. **Batch Debug Script**:
   ```viro
   ; Script to debug factorial
   debug --on
   fact: fn [n] [
       if (= n 0) [1] [
           (* n (fact (- n 1)))
       ]
   ]
   
   ; Start debugging
   debug-step
   fact 3
   debug-inspect  ; { position: 0, expression: "fact 3", ... }
   debug-step
   debug-frame    ; { n: 3 }
   debug-step
   debug-continue
   ```

2. **JSON Output Mode** (for LLM parsing):
   ```viro
   debug --inspect --json
   ; Returns: {"position":5,"frame":{"n":3},"stack":["(top level)","fact"]}
   ```

3. **Helper Functions**:
   ```viro
   debug-trace-to [word]    ; step until word is called
   debug-run-until [expr]   ; step until expression is true
   ```

**Documentation**:
- Create `docs/debugging-guide.md` with examples
- Add LLM-specific examples
- Document all debug commands

**Tests**:
- Test batch debugging scripts
- Test JSON output format
- Test helper functions
- Integration tests with real programs

## File Changes Summary

### New Files
- `internal/debug/step.go` - Step execution logic
- `internal/debug/inspect.go` - Frame/stack inspection
- `internal/native/debug.go` - Debug native functions
- `docs/debugging-guide.md` - User documentation
- `test/contract/debug_step_test.go` - Contract tests

### Modified Files
- `internal/debug/debug.go` - Add step state and callbacks
- `internal/eval/evaluator.go` - Add step hooks
- `internal/repl/repl.go` - Add debug REPL mode
- `internal/native/register_control.go` - Register debug natives

## Success Criteria

1. **Functional**:
   - ✅ Can step through code one expression at a time
   - ✅ Can inspect frame variables at each step
   - ✅ Can inspect call stack
   - ✅ Can evaluate expressions in debug context
   - ✅ Can set breakpoints and continue to them

2. **Usability**:
   - ✅ Simple commands for LLM agents
   - ✅ Clear visual feedback in REPL
   - ✅ JSON output for programmatic use
   - ✅ Comprehensive documentation

3. **Testing**:
   - ✅ All debug commands have contract tests
   - ✅ Integration tests with real programs
   - ✅ Edge cases (empty blocks, errors, nested calls)

## Example Usage

### For LLMs (Automated Debugging)
```viro
; Debug factorial to find bug
debug --on
fact: fn [n] [
    if (= n 0) [1] [
        (* n (fact (- n 1)))  ; Bug: should be (- n 1)
    ]
]

; Step through execution
debug-step
result: fact 3
debug-inspect  ; => { position: 0, expression: "fact 3" }
debug-step
debug-frame    ; => { n: 3 }
debug-step
debug-step
debug-frame    ; => { n: 2 }
debug-continue
```

### For Humans (Interactive Debugging)
```
>> debug --on
>> fact: fn [n] [if (= n 0) [1] [(* n (fact (- n 1)))]]
>> fact 3
[debug:0] >> n
=> n: 3
[debug:1] >> l
{ n: 3 }
[debug:1] >> p n
3
[debug:1] >> n
=> if (= n 0) [1] [(* n (fact (- n 1)))]
[debug:2] >> c
6
>>
```

## Non-Goals

- Visual debugger UI (terminal only)
- Time-travel debugging
- Conditional breakpoints (use debug-run-until instead)
- Remote debugging
- Performance profiling (use trace for that)

## Dependencies

- Existing debug system (`internal/debug`)
- Existing trace system (`internal/trace`)
- REPL infrastructure (`internal/repl`)
- Evaluator hooks (`internal/eval`)

## Estimated Effort

- Phase 1: 2-3 hours
- Phase 2: 2-3 hours
- Phase 3: 2-3 hours
- Phase 4: 1-2 hours
- Testing: 2-3 hours
- Documentation: 1-2 hours

**Total: 10-16 hours**

## Risk Assessment

**Low Risk**:
- Minimal changes to evaluator core
- Builds on existing debug/trace infrastructure
- No breaking changes to existing code
- Can be incrementally adopted

**Mitigation**:
- Phase implementation allows early testing
- Comprehensive contract tests
- Keep original evaluation path unchanged
- Feature flag for debug mode

## Future Enhancements

- Watchpoints (break when variable changes)
- Conditional breakpoints
- Reverse stepping (with execution history)
- Performance profiling integration
- DAP (Debug Adapter Protocol) support
- Visual call graph
