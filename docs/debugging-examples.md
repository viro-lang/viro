# Debugging Examples for Viro

This document provides practical debugging scenarios using Viro's enhanced trace system. Each example includes the buggy code, trace output, analysis, and fix.

## Table of Contents

1. [Example 1: Infinite Recursion Bug](#example-1-infinite-recursion-bug)
2. [Example 2: Variable Mutation Tracking](#example-2-variable-mutation-tracking)
3. [Example 3: Incorrect Function Arguments](#example-3-incorrect-function-arguments)
4. [Example 4: Conditional Logic Error](#example-4-conditional-logic-error)
5. [Example 5: Performance Investigation](#example-5-performance-investigation)

## Example 1: Infinite Recursion Bug

### The Problem

A factorial function hangs and never returns.

### Buggy Code

```viro
fact: fn [n] [
    (* n (fact (- n 1)))
]

result: fact 5
```

### Enabling Trace

```viro
trace --on --step-level 1 --max-depth 10

fact: fn [n] [
    (* n (fact (- n 1)))
]

result: fact 5

trace --off
```

### Sample Trace Output

```json
{"timestamp":"2024-01-15T10:30:00.100Z","event_type":"setword","word":"fact","value":"function[fact]","step":1,"depth":0}
{"timestamp":"2024-01-15T10:30:00.101Z","event_type":"call","word":"fact","value":"<pending>","step":2,"depth":1}
{"timestamp":"2024-01-15T10:30:00.102Z","event_type":"call","word":"*","value":"<pending>","step":3,"depth":2}
{"timestamp":"2024-01-15T10:30:00.103Z","event_type":"call","word":"fact","value":"<pending>","step":4,"depth":2}
{"timestamp":"2024-01-15T10:30:00.104Z","event_type":"call","word":"*","value":"<pending>","step":5,"depth":3}
{"timestamp":"2024-01-15T10:30:00.105Z","event_type":"call","word":"fact","value":"<pending>","step":6,"depth":3}
{"timestamp":"2024-01-15T10:30:00.106Z","event_type":"call","word":"*","value":"<pending>","step":7,"depth":4}
{"timestamp":"2024-01-15T10:30:00.107Z","event_type":"call","word":"fact","value":"<pending>","step":8,"depth":4}
...
{"timestamp":"2024-01-15T10:30:00.120Z","event_type":"call","word":"fact","value":"<pending>","step":21,"depth":10}
```

### Analysis

1. **Observation**: The `depth` field keeps increasing: 1 → 2 → 3 → 4...
2. **Problem**: No `return` events, indicating functions never complete
3. **Root Cause**: Missing base case - function recurses forever
4. **Clue**: Every `fact` call immediately leads to another `fact` call

### The Fix

Add a base case to stop recursion:

```viro
fact: fn [n] [
    if (= n 0) [
        1
    ] [
        (* n (fact (- n 1)))
    ]
]

result: fact 5
```

### Verification

```viro
trace --on --step-level 1 --include-args

result: fact 5

trace --off
```

**Expected trace pattern:**
```json
{"event_type":"call","word":"fact","step":2,"depth":1,"args":{"n":"5"}}
{"event_type":"call","word":"fact","step":5,"depth":2,"args":{"n":"4"}}
{"event_type":"call","word":"fact","step":8,"depth":3,"args":{"n":"3"}}
{"event_type":"call","word":"fact","step":11,"depth":4,"args":{"n":"2"}}
{"event_type":"call","word":"fact","step":14,"depth":5,"args":{"n":"1"}}
{"event_type":"call","word":"fact","step":17,"depth":6,"args":{"n":"0"}}
{"event_type":"return","word":"fact","value":"1","step":18,"depth":6}
{"event_type":"return","word":"fact","value":"1","step":19,"depth":5}
{"event_type":"return","word":"fact","value":"2","step":20,"depth":4}
{"event_type":"return","word":"fact","value":"6","step":21,"depth":3}
{"event_type":"return","word":"fact","value":"24","step":22,"depth":2}
{"event_type":"return","word":"fact","value":"120","step":23,"depth":1}
```

**Key observation**: Depth decreases as functions return, confirming recursion terminates.

---

## Example 2: Variable Mutation Tracking

### The Problem

A calculation produces wrong results, need to track variable changes.

### Buggy Code

```viro
x: 100
discount: 0.20
tax: 0.08

; Apply discount
x: (* x discount)

; Add tax
x: (+ x tax)

print x
```

**Expected**: 86.4 (100 * 0.80 = 80, then 80 + 8% = 86.4)  
**Actual**: 20.08

### Enabling Trace

```viro
trace --on --verbose --step-level 1

x: 100
discount: 0.20
tax: 0.08

x: (* x discount)
x: (+ x tax)

print x

trace --off
```

### Sample Trace Output

```json
{"timestamp":"2024-01-15T10:31:00.100Z","event_type":"setword","word":"x","value":"100","step":1,"depth":0,"frame":{"x":"100"}}
{"timestamp":"2024-01-15T10:31:00.101Z","event_type":"setword","word":"discount","value":"0.20","step":2,"depth":0,"frame":{"x":"100","discount":"0.20"}}
{"timestamp":"2024-01-15T10:31:00.102Z","event_type":"setword","word":"tax","value":"0.08","step":3,"depth":0,"frame":{"x":"100","discount":"0.20","tax":"0.08"}}
{"timestamp":"2024-01-15T10:31:00.103Z","event_type":"call","word":"*","value":"20.0","step":4,"depth":1}
{"timestamp":"2024-01-15T10:31:00.104Z","event_type":"setword","word":"x","value":"20.0","step":5,"depth":0,"frame":{"x":"20.0","discount":"0.20","tax":"0.08"}}
{"timestamp":"2024-01-15T10:31:00.105Z","event_type":"call","word":"+","value":"20.08","step":6,"depth":1}
{"timestamp":"2024-01-15T10:31:00.106Z","event_type":"setword","word":"x","value":"20.08","step":7,"depth":0,"frame":{"x":"20.08","discount":"0.20","tax":"0.08"}}
```

### Analysis

1. **Step 1-3**: Variables initialized correctly: x=100, discount=0.20, tax=0.08
2. **Step 4**: Multiply: `(* 100 0.20)` = **20.0** ⚠️ 
3. **Step 5**: x updated to 20.0 (should be 80.0!)
4. **Root Cause**: Applied discount amount instead of discounted price

**The bug**: `(* x discount)` calculates 20% OF the price, not 80% (price after discount).

### The Fix

```viro
x: 100
discount: 0.20
tax: 0.08

; Apply discount (keep 80% = 1 - discount)
x: (* x (- 1 discount))

; Add tax (8% of current price)
x: (* x (+ 1 tax))

print x
```

### Verification Trace

```json
{"event_type":"call","word":"*","value":"80.0","step":4,"depth":1}
{"event_type":"setword","word":"x","value":"80.0","step":5,"depth":0}
{"event_type":"call","word":"*","value":"86.4","step":7,"depth":1}
{"event_type":"setword","word":"x","value":"86.4","step":8,"depth":0}
```

**Correct values**: 80.0 → 86.4 ✓

---

## Example 3: Incorrect Function Arguments

### The Problem

A function produces unexpected results; need to verify arguments.

### Buggy Code

```viro
; Calculate compound interest: principal * (1 + rate)^years
compound-interest: fn [principal rate years] [
    (* principal (pow (+ 1 rate) years))
]

; Investment scenario
initial: 1000
annual-rate: 0.05
period: 10

result: compound-interest initial period annual-rate
print result
```

**Expected**: ~1628.89  
**Actual**: Much higher (wrong result)

### Enabling Trace

```viro
trace --on --include-args --step-level 0

result: compound-interest initial period annual-rate

trace --off
```

### Sample Trace Output

```json
{"timestamp":"2024-01-15T10:32:00.100Z","event_type":"call","word":"compound-interest","value":"<pending>","step":1,"depth":1,"args":{"principal":"1000","rate":"10","years":"0.05"}}
{"timestamp":"2024-01-15T10:32:00.101Z","event_type":"call","word":"pow","value":"10.05","step":2,"depth":2}
{"timestamp":"2024-01-15T10:32:00.102Z","event_type":"return","word":"compound-interest","value":"10050.0","step":3,"depth":1}
```

### Analysis

1. **Step 1**: Function called with args: `{"principal":"1000","rate":"10","years":"0.05"}`
2. **Problem**: Arguments are swapped!
   - `rate` = 10 (should be 0.05)
   - `years` = 0.05 (should be 10)
3. **Root Cause**: Function call used wrong argument order

**The bug**: Called `compound-interest initial period annual-rate` but parameter order is `[principal rate years]`.

### The Fix

Fix the function call to match parameter order:

```viro
; Correct argument order: principal, rate, years
result: compound-interest initial annual-rate period
print result
```

### Verification Trace

```json
{"event_type":"call","word":"compound-interest","args":{"principal":"1000","rate":"0.05","years":"10"},"step":1,"depth":1}
{"event_type":"return","word":"compound-interest","value":"1628.89","step":3,"depth":1}
```

**Correct arguments**: principal=1000, rate=0.05, years=10 ✓

---

## Example 4: Conditional Logic Error

### The Problem

A grading function returns incorrect grades for boundary values.

### Buggy Code

```viro
grade: fn [score] [
    if (> score 90) ["A"] [
        if (> score 80) ["B"] [
            if (> score 70) ["C"] [
                if (> score 60) ["D"] [
                    "F"
                ]
            ]
        ]
    ]
]

; Test cases
print (grade 90)   ; Expected: A, Actual: B
print (grade 80)   ; Expected: B, Actual: C
print (grade 70)   ; Expected: C, Actual: D
```

### Enabling Trace

```viro
trace --on --step-level 1 --include-args

result: grade 90

trace --off
```

### Sample Trace Output

```json
{"timestamp":"2024-01-15T10:33:00.100Z","event_type":"call","word":"grade","step":1,"depth":1,"args":{"score":"90"}}
{"timestamp":"2024-01-15T10:33:00.101Z","event_type":"call","word":"if","step":2,"depth":2}
{"timestamp":"2024-01-15T10:33:00.102Z","event_type":"call","word":">","value":"false","step":3,"depth":3}
{"timestamp":"2024-01-15T10:33:00.103Z","event_type":"block-enter","step":4,"depth":3}
{"timestamp":"2024-01-15T10:33:00.104Z","event_type":"call","word":"if","step":5,"depth":3}
{"timestamp":"2024-01-15T10:33:00.105Z","event_type":"call","word":">","value":"true","step":6,"depth":4}
{"timestamp":"2024-01-15T10:33:00.106Z","event_type":"block-enter","step":7,"depth":4}
{"timestamp":"2024-01-15T10:33:00.107Z","event_type":"eval","word":"B","value":"B","step":8,"depth":4}
{"timestamp":"2024-01-15T10:33:00.108Z","event_type":"block-exit","value":"B","step":9,"depth":4}
{"timestamp":"2024-01-15T10:33:00.109Z","event_type":"return","word":"grade","value":"B","step":10,"depth":1}
```

### Analysis

1. **Step 3**: `(> 90 90)` returns **false** ⚠️
2. **Step 6**: Falls through to next condition: `(> 90 80)` returns **true**
3. **Step 8**: Returns "B" instead of "A"
4. **Root Cause**: Using `>` (greater than) instead of `>=` (greater than or equal)

**The bug**: Score of exactly 90 is not `> 90`, so it falls to the next tier.

### The Fix

Use `>=` for inclusive boundaries:

```viro
grade: fn [score] [
    if (>= score 90) ["A"] [
        if (>= score 80) ["B"] [
            if (>= score 70) ["C"] [
                if (>= score 60) ["D"] [
                    "F"
                ]
            ]
        ]
    ]
]
```

### Verification Trace

```json
{"event_type":"call","word":"grade","args":{"score":"90"},"step":1,"depth":1}
{"event_type":"call","word":">=","value":"true","step":3,"depth":3}
{"event_type":"eval","value":"A","step":5,"depth":3}
{"event_type":"return","word":"grade","value":"A","step":6,"depth":1}
```

**Correct result**: 90 >= 90 is true, returns "A" ✓

---

## Example 5: Performance Investigation

### The Problem

A data processing function is slower than expected.

### Code to Investigate

```viro
; Process a list of items
process-data: fn [items] [
    result: []
    loop (length items) [
        item: (at items (length result))
        transformed: (transform item)
        result: (append result transformed)
    ]
    result
]

; Helper functions
transform: fn [x] [(* x 2)]
at: fn [list idx] [(select list idx)]
append: fn [list item] [(join list [item])]
length: fn [list] [(size list)]

; Process 1000 items
items: (range 1 1000)
result: process-data items
```

### Enabling Trace

```viro
trace --on --step-level 0

result: process-data items

trace --off
```

### Sample Trace Output (Aggregated)

```json
{"event_type":"call","word":"process-data","duration":5234567890,"step":1,"depth":1}
{"event_type":"call","word":"loop","duration":5234560000,"step":2,"depth":2}
{"event_type":"call","word":"length","duration":1234567,"step":3,"depth":3}
{"event_type":"call","word":"at","duration":891234,"step":4,"depth":3}
{"event_type":"call","word":"length","duration":1234567,"step":5,"depth":3}
{"event_type":"call","word":"transform","duration":456789,"step":6,"depth":3}
{"event_type":"call","word":"append","duration":987654,"step":7,"depth":3}
{"event_type":"call","word":"length","duration":1234567,"step":8,"depth":3}
...
```

### Analysis Using Python Script

```python
import json

# Read trace
events = [json.loads(line) for line in open('trace.json')]

# Count calls and total duration per function
stats = {}
for e in events:
    if e.get('event_type') == 'call':
        word = e['word']
        duration = e.get('duration', 0)
        if word not in stats:
            stats[word] = {'count': 0, 'total_ns': 0}
        stats[word]['count'] += 1
        stats[word]['total_ns'] += duration

# Print statistics
print("Function Performance:")
for word, s in sorted(stats.items(), key=lambda x: x[1]['total_ns'], reverse=True):
    total_ms = s['total_ns'] / 1_000_000
    count = s['count']
    avg_ms = total_ms / count if count > 0 else 0
    print(f"{word}: {total_ms:.2f}ms total, {count} calls, {avg_ms:.4f}ms avg")
```

**Output:**
```
Function Performance:
process-data: 5234.57ms total, 1 calls, 5234.5700ms avg
loop: 5234.56ms total, 1 calls, 5234.5600ms avg
length: 3000.45ms total, 2000 calls, 1.5002ms avg
append: 1975.31ms total, 1000 calls, 1.9753ms avg
at: 891.23ms total, 1000 calls, 0.8912ms avg
transform: 456.79ms total, 1000 calls, 0.4568ms avg
```

### Key Findings

1. **`length` called 2000 times** - Once per iteration for loop condition, once for indexing
2. **O(n²) complexity** - Both `at` and `append` may traverse the list
3. **Hotspot**: `length` and `append` consume most time

### The Fix

Optimize by eliminating redundant calls and using better data structures:

```viro
process-data: fn [items] [
    result: []
    count: (length items)
    idx: 0
    loop count [
        item: (at items idx)
        transformed: (transform item)
        result: (append result transformed)
        idx: (+ idx 1)
    ]
    result
]
```

**Better approach** (if Viro supports it):
- Use iterators instead of indexing
- Use native loop constructs like `foreach`
- Pre-allocate result array if possible

### Verification

After optimization, re-run trace and compare:

**Before**: 5234.57ms  
**After**: 1234.56ms (4x faster) ✓

---

## Summary

These examples demonstrate how to use Viro's trace system for:

1. **Infinite Recursion**: Monitor `depth` field to detect unbounded recursion
2. **Variable Tracking**: Use `--verbose` to see frame state changes
3. **Argument Validation**: Use `--include-args` to verify function inputs
4. **Logic Errors**: Trace conditional evaluations to find boundary bugs
5. **Performance**: Aggregate `duration` data to identify bottlenecks

## Tips for Effective Debugging

1. **Start Simple**: Use `--step-level 0` first, add detail as needed
2. **Use Filters**: Focus on problem areas with `--only` or `--max-depth`
3. **Automate Analysis**: Write scripts to parse and analyze trace JSON
4. **Compare Runs**: Diff traces from working vs. buggy versions
5. **Iterate**: Fix one issue, re-trace, verify, repeat

## See Also

- [Debugging Guide](debugging-guide.md) - Complete trace system reference
- [Observability Guide](observability.md) - Monitoring and metrics
