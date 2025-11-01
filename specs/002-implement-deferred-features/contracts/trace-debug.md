# Contract: Trace & Debug Primitives

**Feature**: Deferred Language Capabilities (002)  
**Functional Requirements**: FR-020, FR-021  
**Applies To**: `trace --on`, `trace --off`, `trace?`, `debug`, `debug --breakpoint`, `debug --remove`, `debug --step`, `debug --locals`, `debug --stack`

---

## Trace Controls

### `trace --on`

#### Signature
```
trace --on [--only words] [--exclude words] [--file path] [--append]
```

#### Parameters
- `--only words`: block of words to include (whitelist). Empty block → include all.
- `--exclude words`: block of words to suppress.
- `--file path`: optional path to override default sink file (`viro-trace.log`). Resolved relative to sandbox root.
- `--append`: append to existing file instead of rotate.

#### Behavior
1. Enable TraceSession if not already active.
2. Configure filters per refinements.
3. Initialize sink:
    - Default: `lumberjack.Logger` rotating at 50 MB, 5 backups.
   - `--file`: create/append at specified path (ensuring sandbox compliance).
4. Emit `trace-control` event noting configuration (clarification #3).

#### Error Cases
- Path outside sandbox → Access error.
- `--only` and `--exclude` containing non-word entries → Script error.

### `trace --off`

- Disables TraceSession, flushes sink, emits `trace-control` event with `enabled=false`.

### `trace?`

- Returns object! summarizing trace state (enabled flag, active filters, sink path, sequence id).

---

## Debugger

debug command
### `debug`

#### Signatures
```
debug --on
debug --off
debug --continue
debug --step
debug --next
debug --finish
debug --breakpoint word [index]
debug --remove id
debug --locals
debug --stack
```

#### Commands
- `--on`: enable debugger (if disabled).
- `--off`: disable debugger and clear breakpoints.
- `--continue`: resume execution until next breakpoint.
- `--step`: step into next evaluation.
- `--next`: step over function calls.
- `--finish`: run until current function returns.
- `--breakpoint word [index]`: add breakpoint.
- `--remove id`: remove breakpoint by ID.
- `--locals`: return object! of local bindings (current frame).
- `--stack`: return block! of call stack entries.

#### Behavior
1. Debugger state stored per REPL session.
2. When active, evaluator checks breakpoints before executing function bodies or words.
3. `locals` deep-copies values to protect runtime state.
4. `stack` provides `word`, `location` for each frame.
5. `continue`, `step`, `next`, `finish` adjust Debugger mode and release control.

#### Error Cases
- Attempt to add breakpoint when debugger disabled → Script error.
- Breakpoint on unknown word → Script error (`unknown-symbol`).
- Removing nonexistent breakpoint → Script error (`no-such-breakpoint`).

### Breakpoint Semantics
- Breakpoints keyed by word name and optional location.
- Conditional breakpoints (future extension) accepted as block after word: `debug --breakpoint 'square [:arg > 10]` (evaluated in debugger context returning logic!).

---

## Trace Event Schema

```json
{
  "timestamp": "2025-10-08T12:34:56.123Z",
  "word": "square",
  "value": "25",
  "duration": 2300000,
  "event_type": "call",
  "step": 42,
  "depth": 1,
  "position": 5,
  "expression": "square 5",
  "args": {"n": "5"},
  "frame": {"n": "5"},
  "parent_expr": "result: square 5",
  "error": ""
}
```

- JSON lines stored in sink file.
- Sequence ID increments monotonically.
- Duration in nanoseconds (integer).

### Event Types
- `eval`: general expression evaluation (literals, words, paths).
- `call`: function call entry.
- `return`: function call exit.
- `block-enter`: entering block evaluation.
- `block-exit`: exiting block evaluation.
- `setword`: set-word assignment.

---

## Testing Expectations

- Contract tests verifying enabling/disabling trace, filter application, sandbox path enforcement.
- Debugger tests ensuring breakpoints fire, locals/stack data accurate, stepping semantics correct.
- Integration tests for success criteria SC-008 (trace usage) and SC-009 (debug session).

---

## Observability & Safety

- Trace sink path resolves via sandbox; failure emits Access error with attempted path.
- Concurrent trace writes protected by mutex.
- Debugger pausing uses channel-based coordination to avoid goroutine leaks.
- When debugger active, REPL prompt changes to `(debug)` indicator.
