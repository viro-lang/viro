# Debugging Guide for Viro

This guide covers the enhanced trace system designed for debugging Viro programs, especially when using LLM agents for automated debugging.

## Table of Contents

1. [Quick Start](#quick-start)
2. [Trace Output Format](#trace-output-format)
3. [Event Types](#event-types)
4. [Trace Levels](#trace-levels)
5. [Common Debugging Patterns](#common-debugging-patterns)
6. [Performance Notes](#performance-notes)
7. [Parsing Trace Output](#parsing-trace-output)

## Quick Start

### Basic Tracing

Enable tracing before running your code:

```viro
; Enable basic tracing (function calls only)
trace --on

; Your code here
fact: fn [n] [
    if (= n 0) [1] [
        (* n (fact (- n 1)))
    ]
]
result: fact 3

; Disable tracing
trace --off
```

### Enhanced Tracing (Recommended for Debugging)

For comprehensive debugging with frame state and arguments:

```viro
; Enable enhanced tracing
trace --on --verbose --include-args --step-level 1

; Your code here
result: fact 3

; Disable tracing
trace --off
```

### Trace to File

```viro
; Trace to file instead of stderr
trace --on --verbose --file "trace.json"

; Your code here
; ...

trace --off
```

## Trace Output Format

The trace system outputs line-delimited JSON events. Each line is a valid JSON object representing a single trace event.

### Basic Fields (Always Present)

- `timestamp` (string): ISO 8601 timestamp of the event
- `word` (string): The word being evaluated (if applicable)
- `value` (string): String representation of the evaluated value
- `duration` (integer): Nanoseconds spent evaluating (0 for instant operations)

### Enhanced Fields (Optional, Phase 1-3)

- `event_type` (string): Type of event (see Event Types below)
- `step` (integer): Monotonic execution step counter
- `depth` (integer): Call stack depth (0 = top level)
- `position` (integer): Position in current block
- `expression` (string): Mold of expression being evaluated
- `args` (object): Function arguments as name->value map (with `--include-args`)
- `frame` (object): Local variables as name->value map (with `--verbose`)
- `parent_expr` (string): Parent context expression
- `error` (string): Error message if evaluation failed

### Example Output

```json
{"timestamp":"2024-01-15T10:30:00.123Z","word":"fact","value":"function[fact]","duration":0,"event_type":"setword","step":1,"depth":0}
{"timestamp":"2024-01-15T10:30:00.124Z","word":"fact","value":"6","duration":1234567,"event_type":"call","step":2,"depth":1,"args":{"n":"3"},"frame":{"n":"3","fact":"function[fact]"}}
{"timestamp":"2024-01-15T10:30:00.125Z","word":"=","value":"false","duration":450,"event_type":"call","step":3,"depth":2}
{"timestamp":"2024-01-15T10:30:00.126Z","word":"*","value":"6","duration":890,"event_type":"call","step":4,"depth":2}
{"timestamp":"2024-01-15T10:30:00.127Z","word":"fact","value":"2","duration":987654,"event_type":"return","step":5,"depth":1}
```

## Event Types

The `event_type` field indicates what kind of event occurred:

### `"eval"`
General expression evaluation. Used for literals, words, paths, and other expressions.

**Example:**
```json
{"event_type":"eval","word":"x","value":"42","step":10,"depth":1}
```

### `"setword"`
Set-word assignment (e.g., `x: 42`).

**Example:**
```json
{"event_type":"setword","word":"x","value":"42","step":1,"depth":0}
```

### `"call"`
Function call entry. Includes function arguments when `--include-args` is used.

**Example:**
```json
{"event_type":"call","word":"add","value":"<pending>","step":5,"depth":2,"args":{"a":"10","b":"20"}}
```

### `"return"`
Function call exit. Shows the return value.

**Example:**
```json
{"event_type":"return","word":"add","value":"30","step":6,"depth":2}
```

### `"block-enter"`
Entering a block evaluation.

**Example:**
```json
{"event_type":"block-enter","step":3,"depth":1}
```

### `"block-exit"`
Exiting a block evaluation. Shows the block's result.

**Example:**
```json
{"event_type":"block-exit","value":"42","step":8,"depth":1}
```

## Trace Levels

Control trace granularity with `--step-level`:

### Level 0: Calls Only (Default)
Only traces function calls and returns. Minimal overhead, best for performance monitoring.

```viro
trace --on --step-level 0
```

**Output includes:**
- Function calls (`event_type: "call"`)
- Function returns (`event_type: "return"`)
- Port operations
- Object operations

### Level 1: Expressions
Traces all expressions including literals, set-words, words, paths, and function calls.

```viro
trace --on --step-level 1
```

**Output includes:**
- All from Level 0
- Expression evaluations (`event_type: "eval"`)
- Set-word assignments (`event_type: "setword"`)
- Block enter/exit (`event_type: "block-enter"`, `"block-exit"`)

### Level 2: All (Future)
Reserved for even more detailed tracing (e.g., internal operations).

## Common Debugging Patterns

### Pattern 1: Finding Infinite Recursion

Enable tracing and monitor the `depth` field:

```viro
trace --on --step-level 1

; Buggy recursive function
countdown: fn [n] [
    print n
    countdown (- n 1)  ; Missing base case!
]
countdown 5

trace --off
```

**Debugging approach:**
1. Parse trace output
2. Filter for `event_type: "call"` with `word: "countdown"`
3. Monitor `depth` field - it will keep increasing
4. Identify the missing base case

### Pattern 2: Tracking Variable Changes

Use `--verbose` to capture frame state at each step:

```viro
trace --on --verbose --step-level 1

x: 10
x: (+ x 5)
x: (* x 2)
print x

trace --off
```

**Debugging approach:**
1. Filter trace events for `event_type: "setword"` with `word: "x"`
2. Examine `value` field at each assignment
3. Verify expected progression: 10 → 15 → 30

### Pattern 3: Function Argument Inspection

Use `--include-args` to see what arguments are passed:

```viro
trace --on --include-args --step-level 0

add: fn [a b] [(+ a b)]
result: add 10 20

trace --off
```

**Debugging approach:**
1. Filter for `event_type: "call"` with `word: "add"`
2. Examine `args` field: `{"a":"10","b":"20"}`
3. Verify arguments match expectations

### Pattern 4: Performance Bottlenecks

Monitor `duration` field to identify slow operations:

```viro
trace --on --step-level 0

slow-operation: fn [n] [
    ; Some expensive computation
    loop n [
        ; ...
    ]
]

slow-operation 1000000

trace --off
```

**Debugging approach:**
1. Parse all events
2. Sort by `duration` (descending)
3. Identify functions with highest duration
4. Optimize hot paths

### Pattern 5: Focused Tracing

Use `--only` to trace specific functions:

```viro
; Only trace 'calculate-interest' and 'compound'
trace --on --only [calculate-interest compound] --include-args

; Run entire program
calculate-loan-payment 100000 0.05 30

trace --off
```

**Benefits:**
- Reduces trace noise
- Focuses on problem area
- Better performance

### Pattern 6: Depth-Limited Tracing

Use `--max-depth` to avoid deep recursion noise:

```viro
; Only trace top 5 call levels
trace --on --max-depth 5 --step-level 1

; Deep recursive function
fibonacci 20

trace --off
```

**Benefits:**
- Prevents trace explosion in deep recursion
- Focuses on top-level logic
- Better readability

## Performance Notes

### Overhead by Configuration

| Configuration | Overhead | Use Case |
|--------------|----------|----------|
| `--step-level 0` | ~1-5% | Production monitoring |
| `--step-level 1` | ~5-10% | General debugging |
| `--verbose` | ~10-30% | Deep debugging (frame capture) |
| `--include-args` | ~2-5% | Function debugging |

### Best Practices

1. **Start Minimal**: Begin with `--step-level 0`, add flags as needed
2. **Use Filters**: Use `--only` or `--exclude` to reduce noise
3. **Limit Depth**: Use `--max-depth` for recursive functions
4. **Disable When Done**: Always call `trace --off` to avoid overhead
5. **File Output**: Use `--file` for large traces to avoid stderr pollution

### Memory Considerations

- Each trace event allocates ~500 bytes (without frame state)
- With `--verbose`, each event can be 2-5 KB depending on frame size
- Use file output with rotation for long-running traces

## Parsing Trace Output

### Python Example

```python
import json
import sys

def analyze_trace(trace_file):
    events = []
    with open(trace_file) as f:
        for line in f:
            events.append(json.loads(line))
    
    # Find function calls
    calls = [e for e in events if e.get('event_type') == 'call']
    
    # Calculate total time per function
    timing = {}
    for call in calls:
        word = call['word']
        duration = call.get('duration', 0)
        timing[word] = timing.get(word, 0) + duration
    
    # Sort by total time
    sorted_timing = sorted(timing.items(), key=lambda x: x[1], reverse=True)
    
    print("Functions by total time:")
    for word, total_ns in sorted_timing[:10]:
        total_ms = total_ns / 1_000_000
        print(f"{word}: {total_ms:.2f}ms")

if __name__ == "__main__":
    analyze_trace(sys.argv[1])
```

### JavaScript Example

```javascript
const fs = require('fs');

function analyzeTrace(traceFile) {
    const content = fs.readFileSync(traceFile, 'utf-8');
    const events = content.trim().split('\n').map(line => JSON.parse(line));
    
    // Build call tree
    const callStack = [];
    const callTree = [];
    
    for (const event of events) {
        if (event.event_type === 'call') {
            const call = {
                word: event.word,
                args: event.args || {},
                depth: event.depth,
                step: event.step,
                children: []
            };
            
            if (callStack.length > 0) {
                callStack[callStack.length - 1].children.push(call);
            } else {
                callTree.push(call);
            }
            
            callStack.push(call);
        } else if (event.event_type === 'return') {
            if (callStack.length > 0) {
                const call = callStack.pop();
                call.returnValue = event.value;
            }
        }
    }
    
    return callTree;
}

function printCallTree(node, indent = 0) {
    const spaces = '  '.repeat(indent);
    const argsStr = JSON.stringify(node.args);
    console.log(`${spaces}${node.word}(${argsStr}) → ${node.returnValue}`);
    
    for (const child of node.children) {
        printCallTree(child, indent + 1);
    }
}

// Usage
const traceFile = process.argv[2];
const callTree = analyzeTrace(traceFile);
callTree.forEach(node => printCallTree(node));
```

## Troubleshooting

### No Trace Output

**Problem**: Running code after `trace --on` but seeing no output.

**Solutions:**
1. Verify trace is enabled: `trace?` should return `true`
2. Check stderr (default output)
3. Verify filters aren't too restrictive
4. Ensure `--step-level` is appropriate for what you're tracing

### Too Much Output

**Problem**: Trace output is overwhelming.

**Solutions:**
1. Use `--only [word1 word2]` to focus on specific functions
2. Use `--step-level 0` to only trace calls
3. Use `--max-depth N` to limit recursion depth
4. Remove `--verbose` if frame state isn't needed

### Missing Frame State

**Problem**: `frame` field is empty or missing.

**Solutions:**
1. Ensure `--verbose` flag is used
2. Frame state only captured at function boundaries
3. Check that variables are actually defined in scope

### Incorrect Step Ordering

**Problem**: Step counter seems out of order.

**Solutions:**
1. Step counter is global and monotonic
2. Concurrent operations may appear interleaved
3. Use `depth` field to understand call hierarchy

## Advanced Usage

### Combining with Debugger

Use trace for post-mortem analysis after hitting breakpoints:

```viro
; Set breakpoint
debug --breakpoint fact

; Enable tracing
trace --on --verbose --include-args

; Run code
fact 5

; Analyze trace to understand what led to breakpoint
trace --off
```

### Trace Diff Between Runs

Compare two trace files to find behavioral changes:

```bash
# Run 1 (working version)
echo 'trace --on --file "trace1.json" | fact 5 | trace --off' | viro

# Run 2 (buggy version) 
echo 'trace --on --file "trace2.json" | fact 5 | trace --off' | viro

# Compare
diff <(jq -r '.step + " " + .word + " " + .value' trace1.json) \
     <(jq -r '.step + " " + .word + " " + .value' trace2.json)
```

### Statistical Analysis

Aggregate trace data for insights:

```python
import json
from collections import Counter

def trace_statistics(trace_file):
    events = [json.loads(line) for line in open(trace_file)]
    
    # Call frequency
    calls = [e['word'] for e in events if e.get('event_type') == 'call']
    frequency = Counter(calls)
    
    print("Most called functions:")
    for word, count in frequency.most_common(10):
        print(f"  {word}: {count} calls")
    
    # Average call depth
    depths = [e['depth'] for e in events if 'depth' in e]
    if depths:
        avg_depth = sum(depths) / len(depths)
        max_depth = max(depths)
        print(f"\nAverage depth: {avg_depth:.1f}")
        print(f"Maximum depth: {max_depth}")
```

## See Also

- [Debugging Examples](debugging-examples.md) - Practical debugging scenarios
- [Observability Guide](observability.md) - Monitoring and instrumentation
- [REPL Usage](repl-usage.md) - Interactive development
