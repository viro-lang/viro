# Plan 033: Return Function for Early Exit from Functions

## Feature Summary

Add `return` native function to provide early exit from user-defined functions in Viro. This function allows returning a value from anywhere within a function body, bypassing remaining code.

**This Plan:**
- Basic `return` with optional value
- Support for user-defined functions
- Support for top-level scripts and REPL (early exit)
- Transparent block support (`do`, `reduce`, `compose` propagate return signals)
- Full test coverage

**Future Considerations:**
- Interaction with `break`/`continue` in nested scenarios

## Research Findings

### Current Function Architecture

User-defined functions execute via `executeFunction()` (`internal/eval/evaluator.go` line 868):

```go
func (e *Evaluator) executeFunction(fn *value.FunctionValue, posArgs []core.Value, refinements map[string]core.Value) (core.Value, error) {
    // Create function frame
    frame := frame.NewFrameWithCapacity(frame.FrameFunctionArgs, parent, len(fn.Params))
    e.PushFrameContext(frame)
    defer e.popFrame()
    
    // Bind parameters
    e.bindFunctionParameters(frame, fn, posArgs, refinements)
    
    // Execute function body
    result, err := e.DoBlock(fn.Body.Elements, fn.Body.Locations())
    if err != nil {
        return value.NewNoneVal(), err
    }
    
    return result, nil  // Last expression value returned
}
```

**Key characteristics:**
- Functions return the value of the last expression in the body
- `DoBlock` executes the function body
- Errors propagate through the return chain
- Functions create a frame boundary

### Error Propagation Flow

1. Native functions return `(core.Value, error)`
2. `DoBlock` propagates errors through the evaluation chain
3. Function boundaries catch and handle return signals
4. Top-level handlers convert uncaught signals to user errors

### Error Annotation Behavior

From `internal/eval/evaluator.go`, the `annotateError` function only operates on `*verror.Error` types:

```go
func (e *Evaluator) annotateError(err error, ...) error {
    if verr, ok := err.(*verror.Error); ok {  // Type assertion
        // ... annotation logic (adds "near", "where", location)
    }
    return err  // Other error types pass through unchanged
}
```

This means custom error types propagate without annotation overhead - perfect for control flow signals!

## Architecture Overview

### Control Flow Mechanism: Custom Error Type with Value Payload

**Decision:** Use a custom error type that carries the return value directly, rather than storing it separately.

**Rationale:**
- **No mutable state** - Value travels with the signal (better architecture)
- **Better performance** - Direct value access, better data locality
- **Simpler implementation** - One component vs three (no Evaluator field + signal + extraction)
- **Type safety** - Compiler-enforced value extraction
- **No corruption risk** - Each signal carries its own value (no shared state)
- **Cleaner mental model** - "Return creates a signal carrying a value"

**Why Different from Break/Continue?**

Break and continue don't carry values, so using `verror.Error` works fine. Return fundamentally differs - it carries a value, making a custom type more appropriate. The slight inconsistency is worth the architectural benefits.

**How It Works:**

1. **Return Signal Type** (custom error with value):
   ```go
   type ReturnSignal struct {
       value core.Value
   }
   
   func (r *ReturnSignal) Error() string {
       return "return signal"
   }
   
   func (r *ReturnSignal) Value() core.Value {
       return r.value
   }
   ```

2. **Return Native Creates Signal:**
   ```go
   func Return(args []core.Value, ...) (core.Value, error) {
       returnVal := value.NewNoneVal()
       if len(args) == 1 {
           returnVal = args[0]
       }
       return value.NewNoneVal(), NewReturnSignal(returnVal)
   }
   ```

3. **Function Handling:**
   ```go
   result, err = e.DoBlock(fn.Body.Elements, fn.Body.Locations())
   if err != nil {
       if returnSig, ok := err.(*ReturnSignal); ok {
           return returnSig.Value(), nil  // Extract value directly
       }
       return value.NewNoneVal(), err  // Propagate other errors
   }
   return result, nil  // Normal completion
   ```

4. **Top-Level Handling** (return at script/REPL level):
   - If `*ReturnSignal` reaches top level → extract value and return it normally
   - Allows early exit from scripts and REPL input
   - Handling in: REPL, API, cmd entry points

5. **Transparent Block Behavior:**
   - `do`, `reduce`, `compose` propagate return signals unchanged (no special handling)
   - Return exits the enclosing function, not the transparent block
   - Error annotation skips `*ReturnSignal` (only processes `*verror.Error`)

### Why This Works

**Normal function return:**
```viro
fn: function [x] [
    print "start"
    return x + 10
    print "never"  ; Skipped
]
fn 5  ; Returns 15
```
- `return` creates `*ReturnSignal` carrying value 15
- Function catches signal → extracts and returns value ✓

**Return in nested function:**
```viro
outer: function [] [
    inner: function [] [return 42]
    x: inner  ; x = 42
    return x + 10  ; Returns 52
]
outer  ; Returns 52
```
- Inner function catches its own return → returns 42
- Outer function sees normal value (not error) ✓

**Return at top level (script/REPL):**
```viro
; In script
print "Starting"
when (some-condition) [return 1]  ; Early exit
print "Never reached"
; Script returns 1

; In REPL
>> return 42
42  ; Returns and displays 42
```
- return creates `*ReturnSignal` carrying value
- Propagates to top-level
- Top-level extracts value and returns it
- Allows early exit from scripts ✓

**Return through transparent blocks:**
```viro
fn: function [x] [
    do [
        when (> x 10) [return x]
        print "x is small"
    ]
    x + 100
]
fn 20  ; Returns 20 (not 120)
```
- do calls DoBlock → gets `*ReturnSignal` carrying value 20
- do propagates error (no special handling for custom error types)
- Function catches signal and extracts value ✓

**Return in loop within function:**
```viro
fn: function [] [
    loop 3 [
        print "iteration"
        return 42
    ]
    print "never"
]
fn  ; Returns 42
```
- Loop body executes return → creates `*ReturnSignal`
- Loop's `isLoopControlSignal` only checks for break/continue (ignores `*ReturnSignal`)
- Loop propagates error (doesn't catch)
- Function catches return signal and extracts value ✓

## Implementation Roadmap

### Step 1: Create ReturnSignal Type (No Error Infrastructure Needed!)

Since `return` works at both function and top level, we don't need error IDs or messages for "return outside function".

**File:** `internal/eval/return_signal.go` (NEW FILE)

Create the custom error type that carries the return value:

```go
package eval

import "github.com/marcin-radoszewski/viro/internal/core"

type ReturnSignal struct {
	value core.Value
}

func NewReturnSignal(val core.Value) *ReturnSignal {
	return &ReturnSignal{value: val}
}

func (r *ReturnSignal) Error() string {
	return "return signal"
}

func (r *ReturnSignal) Value() core.Value {
	return r.value
}
```

**Purpose:** 
- Custom error type that implements the `error` interface
- Carries the return value directly (no separate storage needed)
- Bypasses error annotation (not a `*verror.Error`)

**Validation:** Build succeeds, type is available for use.

### Step 2: Write Test Cases First (TDD)

**File:** `test/contract/return_test.go` (NEW FILE)

Create comprehensive test suite covering:

1. **Basic Return Tests:**
   - Return with value (integer, string, block, object)
   - Return with no value (should return none)
   - Return on first line of function
   - Return on last line of function
   - Return with expression evaluation

2. **Early Exit Tests:**
   - Return skips remaining function code
   - Return with counter to verify early exit
   - Return in conditional branches

3. **Nested Function Tests:**
   - Return in nested function (only inner exits)
   - Outer function continues after inner return
   - Multiple nested levels

4. **Transparent Block Tests:**
   - Return in do block inside function (exits function)
   - Return in reduce inside function (exits function)
   - Return in compose inside function (exits function)
   - Return in when/unless inside function (exits function)

5. **Loop Interaction Tests:**
   - Return in loop body inside function (exits function, not loop)
   - Return in while body inside function (exits function, not loop)
   - Return in foreach body inside function (exits function, not loop)
   - Verify loop doesn't catch return signal

6. **Top-Level Return Tests:**
   - Return at top level in script (early exit)
   - Return in REPL (value extraction)
   - Return in top-level loop (exits script, not loop)
   - Return in top-level conditional

7. **Value Propagation Tests:**
   - Return complex values (blocks, objects)
   - Return result of function call
   - Return series operations

Use table-driven format:

```go
tests := []struct {
    name     string
    input    string
    expected core.Value
    wantErr  bool
}{
    {
        name:     "return with integer value",
        input:    "fn: function [x] [return x + 10]\nfn 5",
        expected: value.NewIntVal(15),
        wantErr:  false,
    },
    {
        name:     "return exits early",
        input:    "fn: function [] [x: 0\nreturn 42\nx: 100]\nfn",
        expected: value.NewIntVal(42),
        wantErr:  false,
    },
    {
        name:     "return with no value returns none",
        input:    "fn: function [] [return]\nfn",
        expected: value.NewNoneVal(),
        wantErr:  false,
    },
    // ... more tests
}
```

**Validation:** All tests should FAIL (function doesn't exist yet).

### Step 3: Implement Return Native (No Helper Needed!)

With the custom error type approach, we don't need a helper function. We can use direct type assertion.

### Step 4: Implement Return Native

**File:** `internal/native/control.go`

Add after the `Continue` function (around line 312):

```go
func Return(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) > 1 {
		return value.NewNoneVal(), arityError("return", 1, len(args))
	}
	
	returnVal := value.NewNoneVal()
	if len(args) == 1 {
		returnVal = args[0]
	}
	
	return value.NewNoneVal(), eval.NewReturnSignal(returnVal)
}
```

**File:** `internal/eval/evaluator.go`

Add factory method:

```go
func (e *Evaluator) NewReturnSignal(val core.Value) error {
	return NewReturnSignal(val)
}
```

**File:** `internal/core/core.go`

Add interface method:

```go
type Evaluator interface {
    // ... existing methods
    NewReturnSignal(val core.Value) error
}
```

**File:** `internal/native/register.go`

Register the function:

```go
Register("return", Return)
```

**Validation:** 
- Build succeeds
- Return tests now fail differently (throw error, not "no value for word: return")

### Step 5: Modify Function Execution to Catch Return

**File:** `internal/eval/evaluator.go`

Modify `executeFunction` (lines 868-891):

**Current code:**
```go
func (e *Evaluator) executeFunction(fn *value.FunctionValue, posArgs []core.Value, refinements map[string]core.Value) (core.Value, error) {
	parent := fn.Parent
	if parent == -1 {
		parent = 0
	}

	frame := frame.NewFrameWithCapacity(frame.FrameFunctionArgs, parent, len(fn.Params))
	frame.Name = functionDisplayName(fn)
	e.PushFrameContext(frame)
	defer e.popFrame()

	e.bindFunctionParameters(frame, fn, posArgs, refinements)

	if fn.Body == nil {
		return value.NewNoneVal(), verror.NewInternalError("function body missing", [3]string{})
	}

	result, err := e.DoBlock(fn.Body.Elements, fn.Body.Locations())
	if err != nil {
		return value.NewNoneVal(), err
	}

	return result, nil
}
```

**Replace with:**
```go
func (e *Evaluator) executeFunction(fn *value.FunctionValue, posArgs []core.Value, refinements map[string]core.Value) (core.Value, error) {
	parent := fn.Parent
	if parent == -1 {
		parent = 0
	}

	frame := frame.NewFrameWithCapacity(frame.FrameFunctionArgs, parent, len(fn.Params))
	frame.Name = functionDisplayName(fn)
	e.PushFrameContext(frame)
	defer e.popFrame()

	e.bindFunctionParameters(frame, fn, posArgs, refinements)

	if fn.Body == nil {
		return value.NewNoneVal(), verror.NewInternalError("function body missing", [3]string{})
	}

	result, err := e.DoBlock(fn.Body.Elements, fn.Body.Locations())
	if err != nil {
		if returnSig, ok := err.(*ReturnSignal); ok {
			return returnSig.Value(), nil
		}
		return value.NewNoneVal(), err
	}

	return result, nil
}
```

**Validation:**
- Basic return tests should PASS
- Early exit tests should PASS
- Value propagation tests should PASS

### Step 6: Ensure Loops Don't Catch Return Signals

**File:** `internal/native/control.go`

Verify that the loop control signal handler ONLY catches break/continue, not return.

The `isLoopControlSignal` helper (added in plan 031) only checks for `*verror.Error` types with specific IDs:

**Current code (from plan 031):**
```go
func isLoopControlSignal(err error) (isControl bool, signalType string) {
	if err == nil {
		return false, ""
	}
	verr, ok := err.(*verror.Error)  // Type assertion to *verror.Error
	if !ok {
		return false, ""  // Returns false for other error types!
	}
	// ... only checks ErrIDBreak and ErrIDContinue
}
```

**Why this works:**
- `*ReturnSignal` is NOT a `*verror.Error`
- Type assertion `verr, ok := err.(*verror.Error)` fails for `*ReturnSignal`
- Function returns `false, ""` immediately
- Loop doesn't catch the signal, propagates it upward ✓

**Validation:** No changes needed - the existing code already ignores `*ReturnSignal`.

**Test verification:**
- Return in loop inside function → exits function (not caught by loop) ✓

### Step 7: Add Top-Level Return Handling

Extract value from return signals at top level (allows early exit from scripts/REPL).

**File:** `internal/api/api.go`

Modify error handling after DoBlock (around line 250):

**Add to existing conversion block (after break/continue):**
```go
result, err := evaluator.DoBlock(values, locations)
if err != nil {
	// Check for return signal (allow top-level return)
	if returnSig, ok := err.(*eval.ReturnSignal); ok {
		return returnSig.Value(), nil  // Extract value and return normally
	}
	
	// Check for break/continue (these ARE errors at top level)
	if verr, ok := err.(*verror.Error); ok {
		if verr.Category == verror.ErrThrow {
			if verr.ID == verror.ErrIDBreak {
				return value.NewNoneVal(), verror.NewScriptError(
					verror.ErrIDBreakOutsideLoop,
					[3]string{},
				)
			}
			if verr.ID == verror.ErrIDContinue {
				return value.NewNoneVal(), verror.NewScriptError(
					verror.ErrIDContinueOutsideLoop,
					[3]string{},
				)
			}
		}
	}
	return result, err
}
return result, nil
```

**Note:** Return signals are handled normally (value extraction), while break/continue are still errors at top level.

**File:** `internal/repl/repl.go`

Add similar handling after DoBlock (around line 342).

**File:** `cmd/viro/run.go` or script execution entry point

Add similar handling where script results are processed.

**Validation:**
- Top-level return tests should PASS
- Return value is correctly extracted and returned
- REPL displays return value
- Scripts exit with return value

### Step 8: Comprehensive Testing

Run full test suite:

```bash
go test ./test/contract/return_test.go -v
go test ./test/contract/control_test.go -v  # Ensure no regressions
go test ./test/contract/loop_control_test.go -v  # Ensure return doesn't interfere
go test ./... -v  # Full test suite
```

**Expected Results:**
- All return_test.go tests PASS
- No regressions in existing control flow tests
- No regressions in loop control tests
- All integration tests PASS

## Integration Points

### 1. Native Function Registry

**Location:** `internal/native/register.go`

**Integration:** Add registration:
- `Register("return", Return)`

**Impact:** Minimal - follows existing pattern for native functions

### 2. Function Execution Flow

**Location:** `internal/eval/evaluator.go` - `executeFunction()`

**Integration:** Add return signal checking after `DoBlock` call:
```go
if err != nil {
    if returnSig, ok := err.(*ReturnSignal); ok {
        return returnSig.Value(), nil
    }
    // Propagate other errors
}
```

**Impact:** Localized to function execution, no changes to DoBlock or evaluator core

### 3. Error Propagation Chain

**Flow:**
1. Native `return` function creates `*ReturnSignal` carrying the value
2. DoBlock propagates error (signal bypasses annotation since it's not `*verror.Error`)
3. Function catches signal and extracts value OR signal propagates upward
4. Top-level extracts value from return signal (allows early script exit)

**Impact:** No changes to core error propagation mechanism

### 4. Interaction with Loop Control

**Behavior:**
```viro
; Return propagates through loops
fn: function [] [
    loop 3 [
        print "once"
        return 42
    ]
    print "never"
]
fn  ; Returns 42, prints "once" only

; Break only exits loop, not function
fn: function [] [
    loop 3 [break]
    return 99
]
fn  ; Returns 99
```

**Implementation:** Loops ignore return signals (only catch break/continue)

### 5. Transparent Blocks Remain Transparent

**Location:** No changes needed in `do`, `reduce`, `compose`

**Behavior:**
```viro
fn: function [x] [
    do [
        when (> x 10) [return x]
    ]
    x + 100
]
fn 20  ; Returns 20 (return exits function, not do block)
```

**Implementation:** Transparent blocks propagate all errors including return signals

## Testing Strategy

### Test Organization

**Primary File:** `test/contract/return_test.go`

Follow existing patterns from `test/contract/control_test.go` and `test/contract/loop_control_test.go`:
- Table-driven tests
- Clear test names describing behavior
- Expected vs actual value comparison
- Error case validation

### Test Categories

#### 1. Basic Functionality Tests

```go
func TestReturn_Basic(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected core.Value
		wantErr  bool
	}{
		{
			name:     "return integer value",
			input:    "fn: function [x] [return x + 10]\nfn 5",
			expected: value.NewIntVal(15),
			wantErr:  false,
		},
		{
			name:     "return no value returns none",
			input:    "fn: function [] [return]\nfn",
			expected: value.NewNoneVal(),
			wantErr:  false,
		},
		{
			name:     "return string value",
			input:    "fn: function [] [return \"hello\"]\nfn",
			expected: value.NewStrVal("hello"),
			wantErr:  false,
		},
		// More tests...
	}
	// Test execution loop...
}
```

#### 2. Early Exit Tests

```go
func TestReturn_EarlyExit(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected core.Value
		wantErr  bool
	}{
		{
			name:     "return skips remaining code",
			input:    "fn: function [] [x: 0\nreturn 42\nx: 100]\nfn",
			expected: value.NewIntVal(42),
			wantErr:  false,
		},
		{
			name:     "return in conditional branch",
			input:    "fn: function [x] [when (> x 10) [return 1]\nreturn 0]\nfn 20",
			expected: value.NewIntVal(1),
			wantErr:  false,
		},
		// More tests...
	}
	// Test execution loop...
}
```

#### 3. Nested Function Tests

```go
func TestReturn_Nested(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected core.Value
		wantErr  bool
	}{
		{
			name: "return in inner function only",
			input: `
				outer: function [] [
					inner: function [] [return 42]
					x: inner
					return x + 10
				]
				outer
			`,
			expected: value.NewIntVal(52),
			wantErr:  false,
		},
		// More tests...
	}
	// Test execution loop...
}
```

#### 4. Transparent Block Tests

```go
func TestReturn_TransparentBlocks(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected core.Value
		wantErr  bool
	}{
		{
			name:     "return in do block exits function",
			input:    "fn: function [x] [do [return x]\nx + 100]\nfn 5",
			expected: value.NewIntVal(5),
			wantErr:  false,
		},
		{
			name:     "return in when exits function",
			input:    "fn: function [x] [when (> x 10) [return x]\nx + 100]\nfn 20",
			expected: value.NewIntVal(20),
			wantErr:  false,
		},
		// More tests...
	}
	// Test execution loop...
}
```

#### 5. Loop Interaction Tests

```go
func TestReturn_LoopInteraction(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected core.Value
		wantErr  bool
	}{
		{
			name:     "return in loop exits function",
			input:    "fn: function [] [x: 0\nloop 3 [x: x + 1\nreturn 42]\nx + 100]\nfn",
			expected: value.NewIntVal(42),
			wantErr:  false,
		},
		{
			name:     "break in loop, return after",
			input:    "fn: function [] [loop 3 [break]\nreturn 99]\nfn",
			expected: value.NewIntVal(99),
			wantErr:  false,
		},
		// More tests...
	}
	// Test execution loop...
}
```

#### 6. Top-Level Return Tests

```go
func TestReturn_TopLevel(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected core.Value
		wantErr  bool
	}{
		{
			name:     "return at top level",
			input:    "return 42",
			expected: value.NewIntVal(42),
			wantErr:  false,
		},
		{
			name:     "return in top-level loop exits script",
			input:    "x: 0\nloop 3 [x: x + 1\nreturn x]\nx + 100",
			expected: value.NewIntVal(1),
			wantErr:  false,
		},
		{
			name:     "return in top-level conditional",
			input:    "x: 20\nwhen (> x 10) [return 99]\nx + 1",
			expected: value.NewIntVal(99),
			wantErr:  false,
		},
		{
			name:     "return with no value at top level",
			input:    "x: 5\nreturn\nx: 10",
			expected: value.NewNoneVal(),
			wantErr:  false,
		},
		// More tests...
	}
	// Test execution loop...
}
```

#### 7. Value Propagation Tests

```go
func TestReturn_ValuePropagation(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected core.Value
		wantErr  bool
	}{
		{
			name:     "return block value",
			input:    "fn: function [] [return [1 2 3]]\nfn",
			expected: value.NewBlockVal([]core.Value{value.NewIntVal(1), value.NewIntVal(2), value.NewIntVal(3)}),
			wantErr:  false,
		},
		{
			name:     "return function call result",
			input:    "inner: function [] [42]\nouter: function [] [return inner]\nouter",
			expected: value.NewIntVal(42),
			wantErr:  false,
		},
		// More tests...
	}
	// Test execution loop...
}
```

### Test Coverage Goals

- **Line Coverage:** >95% of new code
- **Branch Coverage:** 100% of control flow paths
- **Edge Cases:** All documented scenarios tested
- **Error Cases:** All error conditions verified

### Validation Checklist

After all steps:
- [ ] All return_test.go tests pass
- [ ] No regressions in existing tests
- [ ] Return works in user-defined functions
- [ ] Return with value propagates correctly
- [ ] Return with no value returns none
- [ ] Early exit behavior verified
- [ ] Nested functions behave correctly
- [ ] Transparent blocks (do/when/unless) work correctly
- [ ] Return propagates through loops to function boundary
- [ ] Top-level return works in scripts (early exit)
- [ ] Top-level return works in REPL (value extraction)

## Potential Challenges and Mitigations

### Challenge 1: Value Storage During Propagation

**Issue:** Return value needs to persist while error signal propagates to function boundary.

**Solution:** Custom error type carries the value directly.

**Benefits:**
- No mutable state in Evaluator
- Value travels with the signal (better locality)
- No risk of value corruption from nested returns
- Type-safe value extraction
- Simpler implementation (one component vs three)

**Alternatives Considered:**
- Store in Evaluator field - adds mutable state, potential bugs
- Serialize to error Args - complex for large values, performance overhead

**Decision:** Custom error type is superior in every way.

### Challenge 2: Return vs Break/Continue in Loops

**Issue:** Should loops catch return signals like they catch break/continue?

**Answer:** No. Return should propagate through loops to the function boundary.

**Rationale:**
- Return is a function-level operation
- Break/continue are loop-level operations
- Consistent with mainstream languages (JavaScript, Python, etc.)
- Clear mental model: return exits functions, break exits loops

**Implementation:**
- `isLoopControlSignal` only checks for break/continue
- Return signals propagate uncaught through loops
- Function catches return signal ✓

### Challenge 3: Return in Top-Level Code

**Issue:** What happens when return is called outside any function?

**Answer:** Works normally - returns the value and exits the script/REPL input.

**Implementation:**
- Top-level handlers (API, REPL, cmd) catch `*ReturnSignal`
- Extract the value with `returnSig.Value()`
- Return it as the result of the script/REPL input

**Benefit:** Allows early exit from scripts, similar to Ruby, Bash, and other scripting languages.

**Example:**
```viro
; In REPL
>> return 42
42  ; Returns and displays 42

; In script (early exit)
print "Starting"
when (error-condition) [
    print "Error occurred"
    return 1  ; Exit script with value 1
]
print "Continuing"
return 0  ; Normal exit
```

### Challenge 4: Interaction with Existing Control Flow

**Issue:** How does return interact with break/continue in complex scenarios?

**Answer:** They are independent control flow mechanisms:

**Scenario 1: Return in loop**
```viro
fn: function [] [
    loop 3 [
        return 42  ; Exits function, not loop
    ]
]
```
Result: Return propagates through loop to function ✓

**Scenario 2: Break then return**
```viro
fn: function [] [
    loop 3 [break]  ; Exits loop
    return 99       ; Exits function
]
```
Result: Both work independently ✓

**Scenario 3: Return and break together**
```viro
fn: function [x] [
    loop 3 [
        when (> x 10) [return x]
        break
    ]
    return 0
]
```
Result: If x > 10, return exits function; else break exits loop, then return 0 ✓

**Mitigation:**
- Clear documentation of precedence
- Comprehensive tests for interaction scenarios
- Each control flow mechanism has distinct scope

### Challenge 5: None vs Explicit Return

**Issue:** Distinguish between `return` (no value) and `return none`?

**Answer:** Both return none value - no distinction needed.

**Behavior:**
```viro
fn1: function [] [return]       ; Returns none
fn2: function [] [return none]  ; Returns none
fn3: function [] [42]           ; Last expression: returns 42
fn4: function [] []             ; Empty body: returns none
```

**Mitigation:**
- `return` with 0 args → none
- `return` with 1 arg → that value
- Simple and consistent ✓

## Viro Guidelines Reference

### Coding Standards Followed

1. **No comments in code** - All documentation in plan and package docs
2. **Constructor functions** - Use `value.NewNoneVal()`, `value.NewIntVal()`, etc.
3. **Error handling** - Use `verror.NewScriptError()`, `verror.NewError()` with category/ID/args
4. **Table-driven tests** - All tests follow `[]struct{name, input, expected, wantErr}` pattern
5. **TDD approach** - Write tests first, implement to make them pass

### Viro Naming Conventions

- Native function names: lowercase (`return`)
- Error IDs: kebab-case (`return-outside-function`)
- Category constants: PascalCase (`ErrThrow`, `ErrScript`)

### Architecture Alignment

- **Value system:** Returns `core.Value` from natives
- **Error categories:** Uses custom error type for signals, `ErrScript` (300) for user errors
- **Frame system:** No new frames, uses existing function frame management
- **Evaluator integration:** Minimal - only signal type and factory method

## Complexity Analysis

### Original Approach (Evaluator Field + Error for Top-Level)

**Components:**
1. Error ID constants (`ErrIDReturn`, `ErrIDReturnOutsideFunction`)
2. Error message templates (2 messages)
3. Evaluator field (`returnValue core.Value`)
4. Setter method (`SetReturnValue()`)
5. Helper function (`isReturnSignal()`)
6. Top-level error conversion logic

**Issues:**
- Mutable state in Evaluator
- Value separated from signal
- Risk of corruption from nested returns
- Error conversion at top level
- More moving parts = more complexity

### Custom Error Type Approach with Top-Level Support (This Plan)

**Components:**
1. Custom type (`ReturnSignal` struct)
2. Factory method (`NewReturnSignal()`)
3. Top-level value extraction (simple!)

**Benefits:**
- **50% fewer components** (3 vs 6)
- **No error infrastructure** for return
- No mutable state
- Value locality
- Type safety
- Top-level return is a feature, not an error!
- Simpler mental model

**Complexity Reduction:**
```
Original:  Signal + Storage + Extraction + Error Conversion = 4 mechanisms
New:       Signal with Value + Extraction = 2 mechanisms (50% reduction)
```

**Even Simpler:** By making top-level return valid, we eliminate the need for error IDs, messages, and conversion logic!

## Summary

This plan implements `return` for early exit from functions using a **custom error type approach** that:

1. **Leverages existing infrastructure** - Uses error propagation mechanism
2. **Maintains separation of concerns** - Functions handle return, loops propagate it
3. **Provides clear semantics** - Return exits functions, not loops or blocks
4. **Superior architecture** - No mutable state, value travels with signal
5. **Better performance** - Direct value access, better locality, no field lookups
6. **Type safe** - Compiler-enforced value extraction
7. **Handles edge cases** - Nested functions, transparent blocks, loop interaction all work correctly
8. **Follows TDD** - Tests written first, implementation makes them pass
9. **Aligns with Viro** - Coding standards, naming conventions, minimal evaluator changes

**Architecture Comparison:**
- Break/continue: Use `verror.Error` (don't carry values)
- Return: Uses custom `*ReturnSignal` (carries a value)
- Trade-off: Slight inconsistency for major architectural benefits

**Key Behaviors:**
- `fn: function [] [return 42]` → returns 42 ✓
- `fn: function [] [do [return 42]]` → returns 42 (do is transparent) ✓
- `fn: function [] [loop 3 [return 42]]` → returns 42 (propagates through loop) ✓
- `return 42` (top-level) → returns 42 (early script exit) ✓
- REPL: `>> return 42` → displays 42 ✓

**Next Step:** Begin implementation at Step 1 (Error Infrastructure).
