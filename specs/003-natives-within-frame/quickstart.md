# Quickstart: Testing Native Shadowing in Viro

**Feature**: 003-natives-within-frame
**Date**: 2025-10-12
**Audience**: Viro developers testing native shadowing behavior

## Overview

This guide demonstrates how to test the new native shadowing behavior enabled by moving natives into the root frame. After implementation, you'll be able to define local variables and functions with names matching native functions, and use refinement parameters that previously caused conflicts.

---

## Prerequisites

- Viro interpreter built from feature branch `003-natives-within-frame`
- Familiarity with Viro syntax and REPL usage

---

## Quick Tests

### Test 1: Refinement Parameter with Native Name

**Previously**: Caused collision error
**Now**: Works without conflict

```viro
; Define function with --debug refinement
greet: fn [name --debug] [
    if debug [
        print ["Debug: greeting" name]
    ]
    print ["Hello," name]
]

; Call without refinement
greet "Alice"
; Output: Hello, Alice

; Call with refinement
greet "Bob" --debug
; Output:
; Debug: greeting Bob
; Hello, Bob
```

**Expected Result**: No error, both calls work correctly

---

### Test 2: Local Variable Shadowing Native

**Previously**: Impossible (native always found first)
**Now**: Local variable shadows native in its scope

```viro
; Define local "type" variable (shadows native type?)
test: fn [] [
    type: "custom"
    print ["Local type is:" type]
]

test
; Output: Local type is: custom

; Native still accessible at top level
print ["Native type? function:" (type? type?)]
; Output: Native type? function: function!
```

**Expected Result**: Local `type` variable doesn't conflict with native `type?` function

---

### Test 3: User-Defined Function Shadowing Native

**Previously**: Impossible (native registry checked first)
**Now**: User function shadows native in local scope

```viro
; Define custom print that wraps native
print: fn [msg] [
    ; Note: Native print still accessible in root frame
    ; but we can't call it from here after shadowing
    ; (Feature design: shadowing is complete)
    append msg " [custom]"
    ; This would be the native: native-print msg
    ; For this example, we just return the modified message
    msg
]

result: print "Hello"
print result
; Since print is shadowed everywhere after first definition,
; this demonstrates shadowing behavior

; Better example: Shadow in local scope only
wrapper: fn [] [
    print: "not a function, just a string"
    print  ; Returns the string
]

wrapper
; Output: not a function, just a string

; Native print still works at top level
print "Native print works here"
; Output: Native print works here
```

**Expected Result**: Shadowing works, native remains accessible in outer scopes

---

### Test 4: Nested Scope Shadowing

**Previously**: Not tested (shadowing wasn't possible)
**Now**: Demonstrates lexical scoping rules

```viro
; Root level: native "+" available

level1: fn [] [
    +: fn [a b] [a * b]  ; Shadow "+" with multiplication
    print ["Level 1 + 3 4 =" (+ 3 4)]

    level2: fn [] [
        +: fn [a b] [a - b]  ; Shadow again with subtraction
        print ["Level 2 + 3 4 =" (+ 3 4)]
    ]

    level2
    print ["Back to level 1: + 3 4 =" (+ 3 4)]
]

level1
; Output:
; Level 1 + 3 4 = 12  (multiplication)
; Level 2 + 3 4 = -1  (subtraction)
; Back to level 1: + 3 4 = 12  (multiplication)

; Native still works at top level
print ["Native: + 3 4 =" (+ 3 4)]
; Output: Native: + 3 4 = 7  (addition)
```

**Expected Result**: Each scope sees its own binding, innermost wins

---

### Test 5: Closure Capture with Shadowing

**Previously**: Not applicable (no shadowing)
**Now**: Closure captures value, not name

```viro
; Create closure capturing native "+"
captured-fn: fn [] [
    adder: :+  ; Get-word captures current value of "+"
    fn [a b] [adder a b]
]

my-add: captured-fn

; Now shadow "+" at top level
+: "shadowed"

; Closure still uses captured native "+"
print ["Captured function: my-add 10 20 =" (my-add 10 20)]
; Output: Captured function: my-add 10 20 = 30

; But direct reference sees shadow
print ["Direct reference: +" +]
; Output: Direct reference: + shadowed
```

**Expected Result**: Closure uses captured native, not shadowed value

---

## REPL Session Examples

### Session 1: Interactive Shadowing

```viro
>> ; Start with native print
>> print "Hello"
Hello

>> ; Shadow print with custom function
>> print: fn [msg] [append msg " !!!"]

>> ; Now print behaves differently (returns string, doesn't output)
>> print "Hello"
== "Hello !!!"

>> ; Native is gone in this scope
>> ; To get it back, restart REPL or capture it first with :print
```

---

### Session 2: Testing Refinement Collisions

```viro
>> ; Define function using --trace refinement
>> process: fn [data --trace --debug] [
    if trace [print "Tracing enabled"]
    if debug [print "Debug mode on"]
    print ["Processing:" data]
]

>> ; Call with various refinements
>> process "test"
Processing: test

>> process "test" --trace
Tracing enabled
Processing: test

>> process "test" --trace --debug
Tracing enabled
Debug mode on
Processing: test

>> ; No collisions with native trace or debug functions!
```

---

## Automated Test Script

Save as `test-shadowing.viro`:

```viro
; Test Suite: Native Shadowing

print "=== Test 1: Refinement Parameters ==="
fn1: fn [x --debug] [
    if debug [print "Debug mode"]
    x
]
assert (fn1 5) = 5
assert (fn1 5 --debug) = 5
print "PASS: Refinement parameters work"

print "\n=== Test 2: Local Variable Shadowing ==="
fn2: fn [] [
    type: "local"
    type
]
assert (fn2) = "local"
assert (type? 42) = "integer!"
print "PASS: Local variable shadows native"

print "\n=== Test 3: Nested Shadowing ==="
fn3: fn [] [
    +: fn [a b] [a * b]
    inner: fn [] [
        +: fn [a b] [a - b]
        + 10 5
    ]
    [inner, (+ 2 3)]
]
result: fn3
assert (result/1) = 5   ; Inner: 10 - 5
assert (result/2) = 6   ; Outer: 2 * 3
assert (+ 2 3) = 5      ; Native: 2 + 3
print "PASS: Nested shadowing respects scope"

print "\n=== All Tests Passed ==="
```

Run: `./viro test-shadowing.viro`

---

## Common Issues & Solutions

### Issue 1: "I shadowed a native and can't get it back"

**Problem**: After shadowing a native in the REPL, you lose access to it.

**Solution**: Capture natives before shadowing:
```viro
; Capture native before shadowing
native-print: :print
native-add: :+

; Now shadow freely
print: fn [msg] [native-print ["Custom:" msg]]
+: "shadowed"

; Still have access to natives
native-print "Direct access works"
native-add 1 2  ; == 3
```

---

### Issue 2: "My closure doesn't see the shadowed value"

**Expected Behavior**: Closures capture values at creation time, not names.

**Example**:
```viro
x: 10
closure: fn [] [x]  ; Captures value 10

x: 20  ; Rebind x

closure  ; Still returns 10, not 20
```

**Explanation**: Closure captured the **value** bound to `x` when created, not the name `x` itself.

---

### Issue 3: "Performance seems slower after shadowing natives"

**Explanation**: Adding more bindings to frames increases lookup time linearly. However:
- Root frame now has ~70 natives (linear scan)
- Most lookups resolve in local/parent frames (not root)
- Performance impact is negligible for typical code

**Mitigation**: Keep local frames small (<20 bindings recommended)

---

## Verification Checklist

After testing, verify:

- [ ] Functions with `--debug`, `--trace`, `--type` refinements work without errors
- [ ] Local variables can use native names (`type`, `print`, `loop`, etc.)
- [ ] User functions can shadow natives in local scopes
- [ ] Native functions remain accessible from outer scopes after shadowing
- [ ] Nested shadowing follows lexical scoping rules (innermost wins)
- [ ] Closures capture values, not names (immune to rebinding)
- [ ] Existing Viro code without shadowing works identically
- [ ] No performance degradation in typical use cases
- [ ] All existing test suites pass

---

## Further Reading

- [spec.md](./spec.md) - Full feature specification
- [data-model.md](./data-model.md) - Technical data model
- [contracts/](./contracts/) - API contracts
- CLAUDE.md - Viro language design principles

---

## Troubleshooting

### Get Help

If you encounter issues:

1. Check the [spec.md](./spec.md) for expected behavior
2. Review [contracts/lookup.md](./contracts/lookup.md) for lookup semantics
3. Run existing test suite: `go test ./test/contract/...`
4. File an issue on the Viro repository

### Debug Mode

Enable tracing to see word resolution:

```viro
trace --on
+ 3 4  ; See how "+" is resolved
trace --off
```

(Note: Requires observability features from branch 002-implement-deferred-features)

---

**Quickstart Status**: ✅ Complete | **Last Updated**: 2025-10-12
