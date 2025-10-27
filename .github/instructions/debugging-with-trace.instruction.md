# Debugging Viro Code with Enhanced Trace System

This document provides specific instructions for LLM agents (GitHub Copilot, Claude, GPT-4, etc.) debugging Viro code using the enhanced trace system. This is a **quick reference guide** - use this when debugging Viro programs.

## Quick Reference

### Essential Trace Commands

```viro
; Enable comprehensive tracing (recommended for debugging)
trace --on --verbose --include-args --step-level 1

; Your buggy code here
; ...

; Disable tracing
trace --off

; Check if tracing is enabled
trace?
```

### Common Trace Configurations

| Configuration | Use Case | Command |
|--------------|----------|---------|
| Full debugging | Find logic bugs, track variables | `trace --on --verbose --include-args --step-level 1` |
| Function debugging | Verify arguments and returns | `trace --on --include-args --step-level 0` |
| Performance analysis | Find slow functions | `trace --on --step-level 0` |
| Recursion debugging | Detect infinite loops | `trace --on --max-depth 10 --include-args` |
| Focused debugging | Specific function only | `trace --on --only [function-name] --verbose` |

## Complete Debugging Workflow

When debugging Viro code, follow this 5-step workflow:

### Step 1: Enable Enhanced Tracing

Enable tracing **before** running the problematic code:

```viro
trace --on --verbose --include-args --step-level 1
```

### Step 2: Execute Code

Run the code that exhibits the bug:

```viro
; Example: factorial function
fact: fn [n] [
    if (= n 0) [1] [
        (* n (fact (- n 1)))
    ]
]
result: fact 5
```

### Step 3: Disable Tracing

Stop tracing to prevent additional output:

```viro
trace --off
```

### Step 4: Parse and Analyze Trace Output

Trace output is written to stderr as line-delimited JSON. Parse it to analyze execution:

**Python example:**
```python
import json

# Read trace output (from file or captured stderr)
events = []
with open('trace.json') as f:
    for line in f:
        events.append(json.loads(line))

# Analyze events
for event in events:
    print(f"Step {event['step']}: {event.get('event_type', 'unknown')} - {event['word']} = {event['value']}")
```

**JavaScript example:**
```javascript
const fs = require('fs');

// Read trace output
const content = fs.readFileSync('trace.json', 'utf-8');
const events = content.trim().split('\n').map(line => JSON.parse(line));

// Analyze events
events.forEach(event => {
    console.log(`Step ${event.step}: ${event.event_type || 'unknown'} - ${event.word} = ${event.value}`);
});
```

### Step 5: Identify Bug and Fix

Based on trace analysis:

1. **Check depth progression** - Does it increase without returning? → Infinite recursion
2. **Examine frame state** - Are variables changing as expected? → Logic error
3. **Verify function arguments** - Are correct values being passed? → Argument order issue
4. **Compare durations** - Which functions take longest? → Performance bottleneck
5. **Track conditional branches** - Which `if` branches execute? → Conditional logic error

## Trace Output Format Specification

Each trace event is a JSON object with the following fields:

### Core Fields (Always Present)

- **`timestamp`** (string): ISO 8601 timestamp
  - Example: `"2024-01-15T10:30:00.123Z"`
  
- **`word`** (string): Word being evaluated
  - Example: `"fact"`, `"+"`, `"x"`
  
- **`value`** (string): Result value as string
  - Example: `"42"`, `"true"`, `"function[fact]"`
  
- **`duration`** (integer): Evaluation time in nanoseconds
  - Example: `1234567` (1.23ms)
  - Note: 0 for instant operations

### Enhanced Fields (Optional, use `omitempty`)

- **`event_type`** (string): Event classification
  - Values: `"eval"`, `"setword"`, `"call"`, `"return"`, `"block-enter"`, `"block-exit"`
  
- **`step`** (integer): Monotonic execution step counter
  - Example: `1`, `2`, `3`, ...
  - Helps track execution order
  
- **`depth`** (integer): Call stack depth (0 = top level)
  - Example: `0` (top-level), `1` (first call), `2` (nested call)
  - Critical for detecting infinite recursion
  
- **`position`** (integer): Position in current block
  - Example: `0`, `1`, `2` for expressions in a block
  
- **`expression`** (string): Mold of expression being evaluated
  - Example: `"(+ 1 2)"`, `"[1 2 3]"`
  
- **`args`** (object): Function arguments (with `--include-args`)
  - Example: `{"n": "5", "x": "10"}`
  - Keys are parameter names, values are string representations
  
- **`frame`** (object): Local variables (with `--verbose`)
  - Example: `{"x": "42", "fact": "function[fact]"}`
  - Shows all visible variables in current scope
  
- **`parent_expr`** (string): Parent context expression
  - Example: `"(if (= n 0) [1] [...])"` - Shows enclosing expression
  
- **`error`** (string): Error message if evaluation failed
  - Example: `"Division by zero"`, `"Undefined word: foo"`

### Event Types Explained

#### `"eval"` - General Expression Evaluation
Emitted for literals, word lookups, path evaluations, parenthesized expressions.

```json
{"timestamp":"...","event_type":"eval","word":"x","value":"42","step":10,"depth":1}
```

#### `"setword"` - Variable Assignment
Emitted when a set-word assigns a value (e.g., `x: 42`).

```json
{"timestamp":"...","event_type":"setword","word":"x","value":"42","step":1,"depth":0,"frame":{"x":"42"}}
```

#### `"call"` - Function Call Entry
Emitted when entering a function. Use with `--include-args` to see arguments.

```json
{"timestamp":"...","event_type":"call","word":"add","step":5,"depth":2,"args":{"a":"10","b":"20"}}
```

#### `"return"` - Function Call Exit
Emitted when exiting a function. Shows the return value.

```json
{"timestamp":"...","event_type":"return","word":"add","value":"30","step":6,"depth":2}
```

#### `"block-enter"` - Block Evaluation Start
Emitted when starting to evaluate a block `[...]`.

```json
{"timestamp":"...","event_type":"block-enter","step":3,"depth":1}
```

#### `"block-exit"` - Block Evaluation End
Emitted when finishing block evaluation. Shows the block's result.

```json
{"timestamp":"...","event_type":"block-exit","value":"42","step":8,"depth":1}
```

## Common Debugging Patterns

### Pattern 1: Detecting Infinite Recursion

**Problem**: Function never returns, program hangs.

**Trace Command**:
```viro
trace --on --step-level 1 --max-depth 20 --include-args
```

**What to Look For**:
1. Filter events for `event_type: "call"` with the recursive function name
2. Check if `depth` keeps increasing: 1 → 2 → 3 → 4...
3. Look for absence of `"return"` events
4. Verify base case is being evaluated with `--include-args`

**Python Analysis**:
```python
import json

events = [json.loads(line) for line in open('trace.json')]

# Find recursive calls
calls = [e for e in events if e.get('event_type') == 'call']
returns = [e for e in events if e.get('event_type') == 'return']

# Check call/return balance
if len(calls) > len(returns) + 5:
    print(f"⚠️  Unbalanced calls: {len(calls)} calls, {len(returns)} returns")
    print("Likely infinite recursion!")
    
    # Show depth progression
    depths = [e['depth'] for e in calls if 'depth' in e]
    print(f"Depth progression: {depths[:20]}")
    
    # Show arguments of last few calls
    for call in calls[-5:]:
        print(f"  {call['word']} with args: {call.get('args', {})}")
```

**Example Bug**:
```viro
; Missing base case
countdown: fn [n] [
    print n
    countdown (- n 1)  ; No stopping condition!
]
```

**Trace Output**:
```json
{"event_type":"call","word":"countdown","step":1,"depth":1,"args":{"n":"5"}}
{"event_type":"call","word":"countdown","step":5,"depth":2,"args":{"n":"4"}}
{"event_type":"call","word":"countdown","step":9,"depth":3,"args":{"n":"3"}}
...
{"event_type":"call","word":"countdown","step":81,"depth":21,"args":{"n":"-15"}}
```

**Fix**: Add base case to stop recursion.

---

### Pattern 2: Tracking Variable Mutations

**Problem**: Variable has unexpected value, need to track all changes.

**Trace Command**:
```viro
trace --on --verbose --step-level 1
```

**What to Look For**:
1. Filter events for `event_type: "setword"` with the variable name
2. Examine `value` field at each assignment
3. Check `frame` field to see other variables at that moment
4. Verify progression matches expectations

**Python Analysis**:
```python
import json

events = [json.loads(line) for line in open('trace.json')]

# Track variable changes
var_name = 'x'
changes = [e for e in events if e.get('event_type') == 'setword' and e['word'] == var_name]

print(f"Variable '{var_name}' changes:")
for change in changes:
    step = change['step']
    value = change['value']
    frame = change.get('frame', {})
    print(f"  Step {step}: {var_name} = {value}")
    print(f"    Frame: {frame}")
```

**Example Bug**:
```viro
; Incorrect discount calculation
x: 100
discount: 0.20
x: (* x discount)  ; Bug: calculates 20% OF price, not 80% (price after discount)
print x  ; Expected 80, got 20
```

**Trace Output**:
```json
{"event_type":"setword","word":"x","value":"100","step":1,"frame":{"x":"100"}}
{"event_type":"setword","word":"discount","value":"0.20","step":2,"frame":{"x":"100","discount":"0.20"}}
{"event_type":"call","word":"*","value":"20.0","step":3,"depth":1}
{"event_type":"setword","word":"x","value":"20.0","step":4,"frame":{"x":"20.0","discount":"0.20"}}
```

**Analysis**: x changes from 100 → 20.0 instead of expected 80.0.

**Fix**: Use `(* x (- 1 discount))` to calculate discounted price.

---

### Pattern 3: Verifying Function Arguments

**Problem**: Function produces wrong results, suspect incorrect arguments.

**Trace Command**:
```viro
trace --on --include-args --step-level 0
```

**What to Look For**:
1. Filter events for `event_type: "call"` with function name
2. Examine `args` field - check parameter names and values
3. Compare actual arguments to expected arguments
4. Look for argument order issues

**Python Analysis**:
```python
import json

events = [json.loads(line) for line in open('trace.json')]

# Find function calls
func_name = 'compound-interest'
calls = [e for e in events if e.get('event_type') == 'call' and e['word'] == func_name]

print(f"Calls to '{func_name}':")
for call in calls:
    args = call.get('args', {})
    print(f"  Step {call['step']}: {func_name}({args})")
    
    # Validate arguments
    if 'principal' in args and float(args['principal']) > 10:
        if 'rate' in args and float(args['rate']) > 1:
            print("    ⚠️  Warning: rate > 1 looks wrong (should be decimal like 0.05)")
```

**Example Bug**:
```viro
; Function signature: compound-interest [principal rate years]
; Incorrect call: swapped rate and years
result: compound-interest 1000 10 0.05  ; Bug: rate=10, years=0.05
```

**Trace Output**:
```json
{"event_type":"call","word":"compound-interest","step":1,"depth":1,"args":{"principal":"1000","rate":"10","years":"0.05"}}
```

**Analysis**: Arguments are swapped - rate should be 0.05, years should be 10.

**Fix**: Correct argument order: `compound-interest 1000 0.05 10`.

---

### Pattern 4: Analyzing Conditional Logic

**Problem**: Wrong branch executes in conditional statements.

**Trace Command**:
```viro
trace --on --step-level 1 --include-args
```

**What to Look For**:
1. Trace `if` function calls
2. Check condition evaluation (the expression after `if`)
3. Identify which block executes (then vs. else)
4. Look for boundary value issues (`>` vs. `>=`)

**Python Analysis**:
```python
import json

events = [json.loads(line) for line in open('trace.json')]

# Find conditional evaluations
conditionals = []
for i, event in enumerate(events):
    if event.get('word') == 'if' and event.get('event_type') == 'call':
        # Look ahead for the condition result
        condition_result = None
        for j in range(i+1, min(i+5, len(events))):
            if events[j].get('event_type') in ['call', 'eval']:
                condition_result = events[j].get('value')
                break
        conditionals.append({
            'step': event['step'],
            'condition': condition_result,
            'depth': event.get('depth')
        })

print("Conditional evaluations:")
for cond in conditionals:
    print(f"  Step {cond['step']}: if condition = {cond['condition']}")
```

**Example Bug**:
```viro
; Grade function with boundary issues
grade: fn [score] [
    if (> score 90) ["A"] [  ; Bug: 90 is not > 90
        if (> score 80) ["B"] ["C"]
    ]
]
result: grade 90  ; Expected "A", got "B"
```

**Trace Output**:
```json
{"event_type":"call","word":"grade","step":1,"depth":1,"args":{"score":"90"}}
{"event_type":"call","word":"if","step":2,"depth":2}
{"event_type":"call","word":">","value":"false","step":3,"depth":3}
{"event_type":"block-enter","step":4,"depth":3}
{"event_type":"call","word":"if","step":5,"depth":3}
{"event_type":"call","word":">","value":"true","step":6,"depth":4}
{"event_type":"eval","value":"B","step":7,"depth":4}
```

**Analysis**: First condition `(> 90 90)` is false, falls through to second condition.

**Fix**: Use `>=` instead of `>` for inclusive boundaries.

---

### Pattern 5: Performance Profiling

**Problem**: Code is slow, need to identify bottlenecks.

**Trace Command**:
```viro
trace --on --step-level 0  ; Calls only for performance
```

**What to Look For**:
1. Aggregate `duration` field by function name
2. Identify functions with highest total time
3. Check call frequency (many small calls vs. few large calls)
4. Calculate average duration per call

**Python Analysis**:
```python
import json
from collections import defaultdict

events = [json.loads(line) for line in open('trace.json')]

# Aggregate timing data
stats = defaultdict(lambda: {'count': 0, 'total_ns': 0})

for event in events:
    if event.get('event_type') == 'call':
        word = event['word']
        duration = event.get('duration', 0)
        stats[word]['count'] += 1
        stats[word]['total_ns'] += duration

# Sort by total time
sorted_stats = sorted(stats.items(), key=lambda x: x[1]['total_ns'], reverse=True)

print("Performance Profile:")
print(f"{'Function':<20} {'Total (ms)':<12} {'Calls':<10} {'Avg (ms)':<12}")
print("-" * 60)

for word, s in sorted_stats[:10]:
    total_ms = s['total_ns'] / 1_000_000
    count = s['count']
    avg_ms = total_ms / count if count > 0 else 0
    print(f"{word:<20} {total_ms:>10.2f}   {count:>8}   {avg_ms:>10.4f}")

# Identify hotspots
hotspots = [(word, s) for word, s in sorted_stats if s['total_ns'] > 100_000_000]  # > 100ms
if hotspots:
    print("\n⚠️  Performance Hotspots (>100ms):")
    for word, s in hotspots:
        print(f"  {word}: {s['total_ns']/1_000_000:.2f}ms")
```

**Example Analysis Output**:
```
Performance Profile:
Function             Total (ms)   Calls      Avg (ms)    
------------------------------------------------------------
process-data              5234.57        1      5234.5700
loop                      5234.56        1      5234.5600
length                    3000.45     2000         1.5002
append                    1975.31     1000         1.9753
at                         891.23     1000         0.8912
transform                  456.79     1000         0.4568

⚠️  Performance Hotspots (>100ms):
  process-data: 5234.57ms
  loop: 5234.56ms
  length: 3000.45ms
  append: 1975.31ms
```

**Analysis**: `length` called 2000 times, `append` takes 1.98ms per call - likely O(n²) complexity.

**Fix**: Cache length, use better data structures, or eliminate redundant calls.

---

## Parsing Trace Output

### Python Parsing Template

Complete script to parse and analyze trace output:

```python
import json
import sys
from collections import defaultdict, Counter

def parse_trace(trace_file):
    """Parse line-delimited JSON trace file."""
    events = []
    with open(trace_file, 'r') as f:
        for line in f:
            line = line.strip()
            if line:
                try:
                    events.append(json.loads(line))
                except json.JSONDecodeError as e:
                    print(f"Warning: Failed to parse line: {line[:50]}...", file=sys.stderr)
    return events

def analyze_calls(events):
    """Analyze function calls and returns."""
    calls = [e for e in events if e.get('event_type') == 'call']
    returns = [e for e in events if e.get('event_type') == 'return']
    
    print(f"Total calls: {len(calls)}")
    print(f"Total returns: {len(returns)}")
    
    if len(calls) > len(returns):
        print(f"⚠️  Unbalanced: {len(calls) - len(returns)} calls without returns")
        print("   Possible infinite recursion or program interrupted")

def analyze_depth(events):
    """Analyze call stack depth."""
    depths = [e.get('depth', 0) for e in events if 'depth' in e]
    if depths:
        print(f"Max depth: {max(depths)}")
        print(f"Avg depth: {sum(depths)/len(depths):.2f}")

def analyze_timing(events):
    """Aggregate timing by function."""
    stats = defaultdict(lambda: {'count': 0, 'total_ns': 0})
    
    for e in events:
        if e.get('event_type') == 'call':
            word = e['word']
            duration = e.get('duration', 0)
            stats[word]['count'] += 1
            stats[word]['total_ns'] += duration
    
    sorted_stats = sorted(stats.items(), key=lambda x: x[1]['total_ns'], reverse=True)
    
    print("\nTiming Analysis:")
    print(f"{'Function':<20} {'Total (ms)':<12} {'Calls':<10} {'Avg (ms)'}")
    for word, s in sorted_stats[:10]:
        total_ms = s['total_ns'] / 1_000_000
        count = s['count']
        avg_ms = total_ms / count if count > 0 else 0
        print(f"{word:<20} {total_ms:>10.2f}   {count:>8}   {avg_ms:>10.4f}")

def track_variable(events, var_name):
    """Track changes to a specific variable."""
    changes = [e for e in events if e.get('event_type') == 'setword' and e['word'] == var_name]
    
    if not changes:
        print(f"No changes found for variable: {var_name}")
        return
    
    print(f"\nVariable '{var_name}' changes:")
    for change in changes:
        print(f"  Step {change['step']}: {var_name} = {change['value']}")
        if 'frame' in change:
            print(f"    Frame: {change['frame']}")

def main():
    if len(sys.argv) < 2:
        print("Usage: python analyze_trace.py <trace.json> [variable_name]")
        sys.exit(1)
    
    trace_file = sys.argv[1]
    events = parse_trace(trace_file)
    
    print(f"Parsed {len(events)} trace events\n")
    
    analyze_calls(events)
    analyze_depth(events)
    analyze_timing(events)
    
    if len(sys.argv) > 2:
        var_name = sys.argv[2]
        track_variable(events, var_name)

if __name__ == "__main__":
    main()
```

**Usage**:
```bash
python analyze_trace.py trace.json
python analyze_trace.py trace.json x  # Track variable 'x'
```

---

### JavaScript Parsing Template

```javascript
const fs = require('fs');

function parseTrace(traceFile) {
    const content = fs.readFileSync(traceFile, 'utf-8');
    return content.trim().split('\n')
        .filter(line => line.trim())
        .map(line => {
            try {
                return JSON.parse(line);
            } catch (e) {
                console.error(`Warning: Failed to parse line: ${line.slice(0, 50)}...`);
                return null;
            }
        })
        .filter(e => e !== null);
}

function buildCallTree(events) {
    const callStack = [];
    const callTree = [];
    
    for (const event of events) {
        if (event.event_type === 'call') {
            const call = {
                word: event.word,
                args: event.args || {},
                depth: event.depth,
                step: event.step,
                children: [],
                returnValue: null
            };
            
            if (callStack.length > 0 && callStack[callStack.length - 1].depth === event.depth - 1) {
                callStack[callStack.length - 1].children.push(call);
            } else if (event.depth === 0 || callStack.length === 0) {
                callTree.push(call);
            }
            
            callStack.push(call);
        } else if (event.event_type === 'return' && callStack.length > 0) {
            const call = callStack.pop();
            if (call && call.word === event.word) {
                call.returnValue = event.value;
            }
        }
    }
    
    return callTree;
}

function printCallTree(node, indent = 0) {
    const spaces = '  '.repeat(indent);
    const argsStr = Object.entries(node.args)
        .map(([k, v]) => `${k}:${v}`)
        .join(', ');
    console.log(`${spaces}${node.word}(${argsStr}) → ${node.returnValue}`);
    
    for (const child of node.children) {
        printCallTree(child, indent + 1);
    }
}

function analyzePerformance(events) {
    const stats = {};
    
    for (const event of events) {
        if (event.event_type === 'call') {
            const word = event.word;
            if (!stats[word]) {
                stats[word] = { count: 0, totalNs: 0 };
            }
            stats[word].count++;
            stats[word].totalNs += event.duration || 0;
        }
    }
    
    const sorted = Object.entries(stats)
        .sort((a, b) => b[1].totalNs - a[1].totalNs);
    
    console.log('\nPerformance Analysis:');
    console.log('Function'.padEnd(20) + 'Total (ms)'.padStart(12) + 'Calls'.padStart(10));
    
    for (const [word, s] of sorted.slice(0, 10)) {
        const totalMs = (s.totalNs / 1_000_000).toFixed(2);
        console.log(word.padEnd(20) + totalMs.padStart(12) + s.count.toString().padStart(10));
    }
}

// Main
if (process.argv.length < 3) {
    console.error('Usage: node analyze_trace.js <trace.json>');
    process.exit(1);
}

const traceFile = process.argv[2];
const events = parseTrace(traceFile);

console.log(`Parsed ${events.length} trace events\n`);

const callTree = buildCallTree(events);
console.log('Call Tree:');
callTree.forEach(node => printCallTree(node));

analyzePerformance(events);
```

**Usage**:
```bash
node analyze_trace.js trace.json
```

---

## Best Practices for LLM Debugging

When using the trace system as an LLM agent, follow these best practices:

### 1. Always Enable Trace Before Running Code

```viro
; ✅ Good: Enable first
trace --on --verbose --include-args --step-level 1
fact: fn [n] [...]
result: fact 5
trace --off

; ❌ Bad: Code runs before tracing enabled
fact: fn [n] [...]
trace --on
result: fact 5  ; This is traced
```

### 2. Use Appropriate Trace Level

Start minimal, add detail as needed:

```viro
; Step 1: Start with calls only (fast)
trace --on --step-level 0

; Step 2: Add arguments if needed
trace --on --include-args --step-level 0

; Step 3: Full debugging if still unclear
trace --on --verbose --include-args --step-level 1
```

### 3. Write Parsing Scripts for Analysis

Don't manually read trace output - write scripts to analyze it:

```python
# ✅ Good: Automated analysis
events = parse_trace('trace.json')
calls = [e for e in events if e.get('event_type') == 'call']
print(f"Found {len(calls)} function calls")

# ❌ Bad: Manual inspection
# "Let me read through all 10,000 trace lines..."
```

### 4. Focus on Relevant Events

Use filters to reduce noise:

```python
# Focus on specific function
fact_calls = [e for e in events if e['word'] == 'fact' and e.get('event_type') == 'call']

# Focus on specific depth
top_level = [e for e in events if e.get('depth', 0) <= 2]

# Focus on slow operations
slow_ops = [e for e in events if e.get('duration', 0) > 1_000_000]  # > 1ms
```

### 5. Compare Execution Patterns

Compare traces from working vs. buggy code:

```python
# Parse both traces
events_good = parse_trace('trace_good.json')
events_buggy = parse_trace('trace_buggy.json')

# Compare call patterns
calls_good = [e['word'] for e in events_good if e.get('event_type') == 'call']
calls_buggy = [e['word'] for e in events_buggy if e.get('event_type') == 'call']

# Find differences
if calls_good != calls_buggy:
    print("⚠️  Different function call patterns!")
    print(f"Good version: {calls_good}")
    print(f"Buggy version: {calls_buggy}")
```

### 6. Verify Fixes with Re-tracing

After fixing a bug, re-trace to verify:

```viro
; Original buggy code - trace it
trace --on --verbose --include-args
buggy-result: buggy-fn 5
trace --off

; Save trace as 'trace_before.json'

; Apply fix

; Trace again
trace --on --verbose --include-args
fixed-result: fixed-fn 5
trace --off

; Save trace as 'trace_after.json'
; Compare the two traces
```

### 7. Use Max Depth for Recursive Functions

Prevent trace explosion:

```viro
; ✅ Good: Limit depth
trace --on --max-depth 10

; ❌ Bad: No limit on recursive function
trace --on
; fibonacci 50 generates millions of trace events
```

### 8. Document Findings

When analyzing traces, document your findings:

```python
# ✅ Good: Clear analysis with evidence
print("=== Bug Analysis ===")
print(f"Issue: Infinite recursion in 'countdown' function")
print(f"Evidence: {len(calls)} calls, {len(returns)} returns")
print(f"Max depth reached: {max(depths)}")
print(f"Missing base case at step {first_deep_call_step}")
print(f"Recommendation: Add 'if (<= n 0) [return]' before recursive call")
```

---

## Troubleshooting Guide

### Problem: No Trace Output

**Symptoms**: Tracing enabled but no events in output.

**Checks**:
1. Verify trace is enabled: `trace?` should return `true`
2. Check if code actually executed after `trace --on`
3. Ensure stderr is being captured (trace writes to stderr by default)
4. Check if filters are too restrictive (e.g., `--only [non-existent-function]`)

**Solution**:
```viro
; Verify tracing is on
trace --on
is-on: trace?
print is-on  ; Should print "true"

; Run simple code to test
x: 42
print x

trace --off
```

---

### Problem: Trace Output Too Large

**Symptoms**: Trace file is gigabytes, parsing is slow.

**Checks**:
1. Are you tracing recursive functions without `--max-depth`?
2. Is `--step-level 1` necessary or would `--step-level 0` suffice?
3. Is `--verbose` needed or can you debug without frame state?
4. Can you use `--only` to focus on specific functions?

**Solution**:
```viro
; Before: Unlimited tracing
trace --on --verbose --step-level 1
fibonacci 30  ; Generates millions of events

; After: Limited tracing
trace --on --step-level 0 --max-depth 5 --only [fibonacci]
fibonacci 30  ; Manageable output
```

---

### Problem: Missing Frame State

**Symptoms**: `frame` field is empty or not present.

**Checks**:
1. Is `--verbose` flag enabled?
2. Frame state only captured at certain points (function boundaries)
3. Is the variable actually in scope at that point?

**Solution**:
```viro
; Ensure --verbose is used
trace --on --verbose --include-args --step-level 1
```

---

### Problem: Events Out of Order

**Symptoms**: Step numbers seem non-sequential.

**Explanation**: 
- Step counter is global and monotonic
- Multiple threads or concurrent operations may interleave
- This is expected behavior

**Solution**:
- Use `depth` field to understand call hierarchy
- Sort events by `step` if reading out of order
- Use `timestamp` for precise ordering

```python
# Sort by step
events = sorted(events, key=lambda e: e.get('step', 0))

# Sort by timestamp
events = sorted(events, key=lambda e: e['timestamp'])
```

---

### Problem: Can't Parse JSON Output

**Symptoms**: JSON parse errors.

**Checks**:
1. Is the output line-delimited JSON (one object per line)?
2. Are there non-JSON lines mixed in (e.g., print statements)?
3. Is the file complete or was it truncated?

**Solution**:
```python
# Robust parsing with error handling
events = []
with open('trace.json') as f:
    for i, line in enumerate(f, 1):
        line = line.strip()
        if not line:
            continue
        try:
            events.append(json.loads(line))
        except json.JSONDecodeError as e:
            print(f"Line {i}: Failed to parse: {line[:50]}...", file=sys.stderr)
            print(f"  Error: {e}", file=sys.stderr)
```

---

## Real Debugging Session Examples

### Example 1: Fixing Stack Overflow

**Initial Bug Report**: Program crashes with stack overflow.

**Step 1: Enable tracing**
```viro
trace --on --max-depth 30 --include-args
power: fn [base exp] [
    if (= exp 0) [1] [
        (* base (power base exp))  ; Bug here
    ]
]
result: power 2 10
trace --off
```

**Step 2: Analyze trace**
```python
events = parse_trace('trace.json')
calls = [e for e in events if e['word'] == 'power' and e.get('event_type') == 'call']
print(f"Total power calls: {len(calls)}")
for call in calls[:10]:
    print(f"  Depth {call['depth']}: args={call['args']}")
```

**Output**:
```
Total power calls: 30
  Depth 1: args={'base': '2', 'exp': '10'}
  Depth 2: args={'base': '2', 'exp': '10'}
  Depth 3: args={'base': '2', 'exp': '10'}
  ...
  Depth 30: args={'base': '2', 'exp': '10'}
```

**Analysis**: Exponent never decrements! Bug is in recursive call.

**Fix**:
```viro
power: fn [base exp] [
    if (= exp 0) [1] [
        (* base (power base (- exp 1)))  ; Fixed: decrement exp
    ]
]
```

---

### Example 2: Performance Degradation

**Initial Bug Report**: Function was fast, now it's slow after refactoring.

**Step 1: Trace both versions**
```bash
# Fast version
echo 'trace --on --file "trace_fast.json" | process-list data | trace --off' | viro

# Slow version  
echo 'trace --on --file "trace_slow.json" | process-list data | trace --off' | viro
```

**Step 2: Compare performance**
```python
def get_timing_stats(trace_file):
    events = parse_trace(trace_file)
    stats = {}
    for e in events:
        if e.get('event_type') == 'call':
            word = e['word']
            stats[word] = stats.get(word, 0) + e.get('duration', 0)
    return stats

fast = get_timing_stats('trace_fast.json')
slow = get_timing_stats('trace_slow.json')

print("Performance Comparison:")
for word in slow:
    if word in fast:
        ratio = slow[word] / fast[word] if fast[word] > 0 else float('inf')
        if ratio > 2:
            print(f"  {word}: {ratio:.1f}x slower")
```

**Output**:
```
Performance Comparison:
  validate-item: 15.3x slower
  lookup-cache: 8.7x slower
```

**Analysis**: New validation and cache lookup are bottlenecks.

**Fix**: Optimize or cache validation results.

---

## Summary

This guide provides LLM agents with everything needed to effectively debug Viro code using the enhanced trace system:

1. **Quick commands** for common scenarios
2. **Complete workflow** from enabling trace to fixing bugs
3. **Detailed format specification** for parsing JSON output
4. **5 common debugging patterns** with examples
5. **Parsing templates** in Python and JavaScript
6. **8 best practices** for efficient debugging
7. **Troubleshooting** for common issues
8. **Real examples** of debugging sessions

## Key Takeaways

- Always enable trace BEFORE running code
- Start with minimal tracing (`--step-level 0`), add detail as needed
- Write scripts to parse and analyze trace output
- Use `depth` field to detect infinite recursion
- Use `args` field to verify function arguments
- Use `frame` field to track variable changes
- Use `duration` field to find performance bottlenecks
- Compare traces from working vs. buggy code to find differences

## See Also

- [Debugging Guide](../../docs/debugging-guide.md) - Complete user-facing documentation
- [Debugging Examples](../../docs/debugging-examples.md) - Practical scenarios with analysis
- [AGENTS.md](../../AGENTS.md) - General agent guidelines for Viro
