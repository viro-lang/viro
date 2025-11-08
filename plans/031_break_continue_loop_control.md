# Plan 031: Break and Continue Loop Control

## Feature Summary

Add `break` and `continue` native functions to provide loop control flow in Viro. These functions allow early exit from loops and skipping to the next iteration respectively.

**Phase 1 (This Plan):**
- Basic single-level `break` and `continue`
- Support for `loop`, `while`, and `foreach`
- Function boundary enforcement (break/continue cannot cross function calls)
- Transparent block support (`do`, `reduce`, `compose` propagate control signals)
- Error detection for usage outside loops
- Full test coverage

**Phase 2 (Future):**
- Multi-level support with `--levels N` refinement
- Example: `break --levels 2` exits 2 nested loops

## Research Findings

### Current Loop Architecture

All three loop constructs use `eval.DoBlock()` to execute loop bodies:

1. **Loop** (`internal/native/control.go` lines 89-157):
   - Fixed iteration count with optional `--with-index` refinement
   - Pattern: `for i := 0; i < count; i++ { result, err = eval.DoBlock(...) }`

2. **While** (`internal/native/control.go` lines 159-222):
   - Conditional loop with block re-evaluation
   - Pattern: `for ToTruthy(condition) { result, err = eval.DoBlock(...) }`

3. **Foreach** (`internal/native/control.go` lines 519-631):
   - Series iteration with variable binding
   - Pattern: `for i := startIndex; i < length; { result, err = eval.DoBlock(...) }`

All loops:
- Return `(core.Value, error)` from DoBlock
- Check `if err != nil` after DoBlock call
- Currently propagate all errors without special handling

### Error Category System

From `internal/verror/categories.go`:
- **ErrThrow (Category 0)**: "Loop control: break outside loop, etc."
  - Specifically designed for control flow mechanisms
  - NOT a user-facing error category
  - Perfect for internal control flow signals

- **ErrScript (Category 300)**: "Script: undefined words, type mismatches, invalid operations"
  - User-facing runtime errors
  - Use for "break outside loop" errors shown to users

### Error Propagation Flow

1. Native functions return `(core.Value, error)`
2. `DoBlock` propagates errors through the evaluation chain
3. Evaluator's `annotateError()` adds context (near, where, location)
4. Top-level handlers (REPL, API, cmd) display errors to users

## Architecture Overview

### Control Flow Mechanism: Error-Based Signaling

**Decision:** Use special error instances (category `ErrThrow`) to signal break/continue.

**Rationale:**
- Leverages existing error propagation through DoBlock
- ErrThrow category specifically designed for this purpose
- Minimal changes to evaluator core
- Clean separation: loops catch and handle, others propagate
- No need for evaluator state tracking

**How It Works:**

1. **Control Flow Signals** (internal, not errors):
   - `break` returns: `verror.NewError(ErrThrow, "break", [3]string{})`
   - `continue` returns: `verror.NewError(ErrThrow, "continue", [3]string{})`
   - These propagate through DoBlock like errors
   - Loops catch and handle them specially

2. **Loop Handling:**
   ```go
   result, err = eval.DoBlock(block.Elements, block.Locations())
   if err != nil {
       isControl, signalType := isLoopControlSignal(err)
       if isControl {
           if signalType == "break" {
               return value.NewNoneVal(), nil  // Exit loop
           }
           // signalType == "continue"
           continue  // Next iteration
       }
       return value.NewNoneVal(), err  // Propagate other errors
   }
   ```

3. **Top-Level Conversion** (uncaught signals):
   - If ErrThrow/"break" reaches top level → convert to ErrScript/"break-outside-loop"
   - If ErrThrow/"continue" reaches top level → convert to ErrScript/"continue-outside-loop"
   - Conversion happens in: REPL, API, cmd entry points

4. **Nested Loop Behavior:**
   - Each loop catches control signals independently
   - Break/continue affects only the innermost loop
   - Signals are consumed (not re-thrown) when caught

### Why This Works

**Nested loops:**
```viro
loop 3 [
    loop 2 [break]  ; Inner break
    print "outer"    ; Executes 3 times
]
```
- Inner loop catches break → exits → returns none to outer
- Outer loop sees no error → continues normally ✓

**Break through function calls should NOT work:**
```viro
fn: function [] [break]
loop 3 [fn]  ; Error: break outside loop
```
- fn's DoBlock evaluates break → returns ErrThrow/"break"
- Function catches control signal at function boundary → converts to ErrScript/"break-outside-loop"
- Error propagates to loop as script error (not caught)
- User sees: "break called outside of loop" ✓

**Break outside loop:**
```viro
break  ; Error!
```
- break returns ErrThrow/"break"
- Propagates to top-level
- Converted to ErrScript/"break-outside-loop"
- Displayed to user ✓

**Reduce/compose/do transparency:**
```viro
loop 3 [
    do [break]  ; Works - break exits loop
]
```
- do calls DoBlock → gets ErrThrow/"break"
- do propagates error (no special handling)
- Loop catches it ✓

## Implementation Roadmap

### Step 1: Error Infrastructure (No Code Changes Yet)

**File:** `internal/verror/categories.go`

Add error ID constants after line 131:

```go
// Loop control error IDs (ErrThrow category)
ErrIDBreak    = "break"     // Internal control flow signal
ErrIDContinue = "continue"  // Internal control flow signal

// Loop control error cases (ErrScript category)
ErrIDBreakOutsideLoop    = "break-outside-loop"
ErrIDContinueOutsideLoop = "continue-outside-loop"
```

**File:** `internal/verror/error.go`

Add message templates after line 185:

```go
// Loop control messages
ErrIDBreak:               "break",  // Internal only, not shown
ErrIDContinue:            "continue",  // Internal only, not shown
ErrIDBreakOutsideLoop:    "break called outside of loop",
ErrIDContinueOutsideLoop: "continue called outside of loop",
```

**Validation:** Build should succeed, no functional changes.

### Step 2: Write Test Cases First (TDD)

**File:** `test/contract/loop_control_test.go` (NEW FILE)

Create comprehensive test suite covering:

1. **Basic Break Tests:**
   - Break in loop (verify early exit, return none)
   - Break in while (verify early exit, return none)
   - Break in foreach (verify early exit, return none)
   - Break on first iteration
   - Break with counter variable (verify iterations count)

2. **Basic Continue Tests:**
   - Continue in loop (verify iteration skipped)
   - Continue in while (verify condition re-evaluated)
   - Continue in foreach (verify next element processed)
   - Continue with counter variable (verify skip behavior)

3. **Nested Loop Tests:**
   - Break in nested loop (only inner exits)
   - Continue in nested loop (only inner affected)
   - Outer loop continues after inner break

4. **Function Boundary Tests:**
   - Break in function called from loop → error (break doesn't cross function boundary)
   - Continue in function called from loop → error (continue doesn't cross function boundary)
   - Break/continue in do block inside loop → works (do is transparent, not a function)

5. **Error Cases:**
   - Break outside any loop → error
   - Continue outside any loop → error

6. **Edge Cases:**
   - Break/continue with --with-index refinement
   - Break in do block inside loop (propagates to loop - do is transparent)
   - Continue in do block inside loop (propagates to loop - do is transparent)
   - Break in user function called from loop (error - function boundary blocks it)

Use table-driven format like existing `test/contract/control_test.go`:

```go
tests := []struct {
    name     string
    input    string
    expected core.Value
    wantErr  bool
}{
    {
        name:     "break exits loop immediately",
        input:    "loop 3 [break]",
        expected: value.NewNoneVal(),
        wantErr:  false,
    },
    // ... more tests
}
```

**Validation:** All tests should FAIL (functions don't exist yet).

### Step 3: Implement Loop Control Helper

**File:** `internal/native/control.go`

Add helper function after the `ToTruthy` function (around line 315):

```go
func isLoopControlSignal(err error) (isControl bool, signalType string) {
	if err == nil {
		return false, ""
	}
	verr, ok := err.(*verror.Error)
	if !ok {
		return false, ""
	}
	if verr.Category != verror.ErrThrow {
		return false, ""
	}
	if verr.ID == verror.ErrIDBreak {
		return true, "break"
	}
	if verr.ID == verror.ErrIDContinue {
		return true, "continue"
	}
	return false, ""
}
```

**Purpose:** Detect if an error is a loop control signal.

**Validation:** Helper compiles, no behavior changes yet.

### Step 4: Implement Break Native

**File:** `internal/native/control.go`

Add after the `Foreach` function (around line 631):

```go
func Break(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 0 {
		return value.NewNoneVal(), arityError("break", 0, len(args))
	}
	return value.NewNoneVal(), verror.NewError(verror.ErrThrow, verror.ErrIDBreak, [3]string{})
}
```

**File:** `internal/native/register.go`

Register the function in the appropriate section:

```go
Register("break", Break)
```

**Validation:** 
- Build succeeds
- Break tests now fail differently (throw error, not "no value for word: break")

### Step 5: Implement Continue Native

**File:** `internal/native/control.go`

Add after the `Break` function:

```go
func Continue(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 0 {
		return value.NewNoneVal(), arityError("continue", 0, len(args))
	}
	return value.NewNoneVal(), verror.NewError(verror.ErrThrow, verror.ErrIDContinue, [3]string{})
}
```

**File:** `internal/native/register.go`

Register the function:

```go
Register("continue", Continue)
```

**Validation:**
- Build succeeds
- Continue tests fail differently (throw error propagates uncaught)

### Step 6: Modify Loop Native

**File:** `internal/native/control.go`

Modify the `Loop` function (lines 89-157). Replace the iteration loop (lines 145-154):

**Current code:**
```go
for i := 0; i < int(count); i++ {
	if hasIndexRef && indexVal.GetType() != value.TypeNone {
		currentFrame.Bind(indexWord, value.NewIntVal(int64(i)))
	}

	result, err = eval.DoBlock(block.Elements, block.Locations())
	if err != nil {
		return value.NewNoneVal(), err
	}
}
```

**Replace with:**
```go
for i := 0; i < int(count); i++ {
	if hasIndexRef && indexVal.GetType() != value.TypeNone {
		currentFrame.Bind(indexWord, value.NewIntVal(int64(i)))
	}

	result, err = eval.DoBlock(block.Elements, block.Locations())
	if err != nil {
		isControl, signalType := isLoopControlSignal(err)
		if isControl {
			if signalType == "break" {
				return value.NewNoneVal(), nil
			}
			continue
		}
		return value.NewNoneVal(), err
	}
}
```

**Validation:**
- Loop break tests should PASS
- Loop continue tests should PASS

### Step 7: Modify While Native

**File:** `internal/native/control.go`

Modify the `While` function (lines 159-222). Two locations need changes:

**Location 1: Condition is a block (lines 190-207)**

Current inner loop:
```go
for {
	conditionResult, err := eval.DoBlock(conditionBlock.Elements, conditionBlock.Locations())
	if err != nil {
		return value.NewNoneVal(), err
	}

	if !ToTruthy(conditionResult) {
		break
	}

	result, err = eval.DoBlock(bodyBlock.Elements, bodyBlock.Locations())
	if err != nil {
		return value.NewNoneVal(), err
	}
}
```

Replace with:
```go
for {
	conditionResult, err := eval.DoBlock(conditionBlock.Elements, conditionBlock.Locations())
	if err != nil {
		return value.NewNoneVal(), err
	}

	if !ToTruthy(conditionResult) {
		break
	}

	result, err = eval.DoBlock(bodyBlock.Elements, bodyBlock.Locations())
	if err != nil {
		isControl, signalType := isLoopControlSignal(err)
		if isControl {
			if signalType == "break" {
				return value.NewNoneVal(), nil
			}
			continue
		}
		return value.NewNoneVal(), err
	}
}
```

**Location 2: Condition is not a block (lines 208-219)**

Current loop:
```go
for ToTruthy(condition) {
	var err error
	result, err = eval.DoBlock(bodyBlock.Elements, bodyBlock.Locations())
	if err != nil {
		return value.NewNoneVal(), err
	}
}
```

Replace with:
```go
for ToTruthy(condition) {
	var err error
	result, err = eval.DoBlock(bodyBlock.Elements, bodyBlock.Locations())
	if err != nil {
		isControl, signalType := isLoopControlSignal(err)
		if isControl {
			if signalType == "break" {
				return value.NewNoneVal(), nil
			}
			continue
		}
		return value.NewNoneVal(), err
	}
}
```

**Validation:**
- While break tests should PASS
- While continue tests should PASS

### Step 8: Modify Foreach Native

**File:** `internal/native/control.go`

Modify the `Foreach` function (lines 519-631). Replace the iteration loop (lines 607-628):

**Current code:**
```go
for i := startIndex; i < length; {
	for j := 0; j < numVars; j++ {
		if i < length {
			element := series.ElementAt(i)
			currentFrame.Bind(varNames[j], element)
			i++
		} else {
			currentFrame.Bind(varNames[j], value.NewNoneVal())
		}
	}

	if hasIndexRef && indexVal.GetType() != value.TypeNone {
		currentFrame.Bind(indexWord, value.NewIntVal(int64(iteration)))
	}

	result, err = eval.DoBlock(bodyBlock.Elements, bodyBlock.Locations())

	if err != nil {
		return value.NewNoneVal(), err
	}
	iteration++
}
```

**Replace with:**
```go
for i := startIndex; i < length; {
	for j := 0; j < numVars; j++ {
		if i < length {
			element := series.ElementAt(i)
			currentFrame.Bind(varNames[j], element)
			i++
		} else {
			currentFrame.Bind(varNames[j], value.NewNoneVal())
		}
	}

	if hasIndexRef && indexVal.GetType() != value.TypeNone {
		currentFrame.Bind(indexWord, value.NewIntVal(int64(iteration)))
	}

	result, err = eval.DoBlock(bodyBlock.Elements, bodyBlock.Locations())

	if err != nil {
		isControl, signalType := isLoopControlSignal(err)
		if isControl {
			if signalType == "break" {
				return value.NewNoneVal(), nil
			}
			iteration++
			continue
		}
		return value.NewNoneVal(), err
	}
	iteration++
}
```

**Note:** For foreach, when continue is triggered, we still increment the iteration counter before continuing.

**Validation:**
- Foreach break tests should PASS
- Foreach continue tests should PASS

### Step 8.5: Add Function Boundary Error Conversion

Convert loop control signals at function boundaries to prevent break/continue from crossing function calls.

**File:** `internal/eval/evaluator.go`

Modify the `callUserDefinedFunction` function (around line 626). Replace the error handling after `executeFunction`:

**Current code:**
```go
func (e *Evaluator) callUserDefinedFunction(fn *value.FunctionValue, posArgs []core.Value, refValues map[string]core.Value, name string, position int, traceStart time.Time) (core.Value, error) {
	result, err := e.executeFunction(fn, posArgs, refValues)
	if err != nil {
		if e.traceEnabled {
			e.emitTraceResult("return", name, name, value.NewNoneVal(), position, traceStart, err)
		}
		return value.NewNoneVal(), err
	}
	return result, nil
}
```

**Replace with:**
```go
func (e *Evaluator) callUserDefinedFunction(fn *value.FunctionValue, posArgs []core.Value, refValues map[string]core.Value, name string, position int, traceStart time.Time) (core.Value, error) {
	result, err := e.executeFunction(fn, posArgs, refValues)
	if err != nil {
		if verr, ok := err.(*verror.Error); ok {
			if verr.Category == verror.ErrThrow {
				if verr.ID == verror.ErrIDBreak {
					convertedErr := verror.NewScriptError(
						verror.ErrIDBreakOutsideLoop,
						[3]string{},
					)
					if e.traceEnabled {
						e.emitTraceResult("return", name, name, value.NewNoneVal(), position, traceStart, convertedErr)
					}
					return value.NewNoneVal(), convertedErr
				}
				if verr.ID == verror.ErrIDContinue {
					convertedErr := verror.NewScriptError(
						verror.ErrIDContinueOutsideLoop,
						[3]string{},
					)
					if e.traceEnabled {
						e.emitTraceResult("return", name, name, value.NewNoneVal(), position, traceStart, convertedErr)
					}
					return value.NewNoneVal(), convertedErr
				}
			}
		}
		if e.traceEnabled {
			e.emitTraceResult("return", name, name, value.NewNoneVal(), position, traceStart, err)
		}
		return value.NewNoneVal(), err
	}
	return result, nil
}
```

**Important:** This conversion happens ONLY in `callUserDefinedFunction`, NOT in `callNativeFunction`. This ensures that:
- User-defined functions create a boundary that blocks break/continue
- Native functions like `do`, `reduce`, and `compose` remain transparent and propagate control signals unchanged

**Validation:**
- User function with break/continue returns "break-outside-loop" or "continue-outside-loop" error
- Loop calling function with break/continue sees script error (not control signal)
- `do [break]` inside loop still works (native function, no conversion)

### Step 9: Add Top-Level Error Conversion

Convert uncaught control flow signals to user-facing errors at entry points.

**File:** `internal/api/api.go`

Modify around line 250. Replace:

```go
result, err := evaluator.DoBlock(values, locations)
if err != nil {
	return result, err
}
return result, nil
```

With:

```go
result, err := evaluator.DoBlock(values, locations)
if err != nil {
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

**File:** `internal/repl/repl.go`

Modify around line 342. Similar conversion after:

```go
result, err := r.evaluator.DoBlock(values, locations)
```

Add conversion check before error handling.

**File:** `cmd/viro/main.go` or equivalent script execution entry point

Add similar conversion at the point where script DoBlock results are handled.

**Validation:**
- Error case tests should PASS
- Break/continue outside loop produces clear error message

### Step 10: Comprehensive Testing

Run full test suite:

```bash
go test ./test/contract/loop_control_test.go -v
go test ./test/contract/control_test.go -v  # Ensure no regressions
go test ./... -v  # Full test suite
```

**Expected Results:**
- All loop_control_test.go tests PASS
- No regressions in existing control flow tests
- All integration tests PASS

## Integration Points

### 1. Native Function Registry

**Location:** `internal/native/register.go`

**Integration:** Add two new registrations:
- `Register("break", Break)`
- `Register("continue", Continue)`

**Impact:** Minimal - follows existing pattern for native functions

### 2. Loop Evaluation Flow

**Location:** All three loop natives in `internal/native/control.go`

**Integration:** Add error checking after `DoBlock` calls:
```go
if err != nil {
    isControl, signalType := isLoopControlSignal(err)
    if isControl {
        // Handle break/continue
    }
    // Propagate other errors
}
```

**Impact:** Localized to loop implementations, no evaluator changes

### 3. Error Propagation Chain

**Flow:**
1. Native function returns error
2. DoBlock propagates error
3. Loop catches and handles OR propagates upward
4. Top-level converts uncaught control signals to script errors

**Impact:** No changes to core error propagation mechanism

### 4. Function Boundaries

**Location:** Function call handling in evaluator

**Integration:** Functions must convert loop control signals to script errors at their boundary

**Behavior:**
```viro
; Transparent blocks (do, reduce, compose) - NO conversion needed
loop 3 [
    do [break]  ; do propagates ErrThrow/"break" to loop - works ✓
]

; User-defined functions - conversion at boundary required
fn: function [] [break]
loop 3 [fn]  ; fn converts to ErrScript/"break-outside-loop" - error ✓
```

**Implementation:** Add conversion in function return handling (see Step 8.5)

## Testing Strategy

### Test Organization

**Primary File:** `test/contract/loop_control_test.go`

Follow existing patterns from `test/contract/control_test.go`:
- Table-driven tests
- Clear test names describing behavior
- Expected vs actual value comparison
- Error case validation

### Test Categories

#### 1. Basic Functionality Tests

```go
func TestLoopControl_BreakBasic(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected core.Value
		wantErr  bool
	}{
		{
			name:     "break in loop returns none",
			input:    "loop 3 [break]",
			expected: value.NewNoneVal(),
			wantErr:  false,
		},
		{
			name:     "break exits early - counter check",
			input:    "x: 0\nloop 10 [x: x + 1\nwhen (= x 3) [break]]\nx",
			expected: value.NewIntVal(3),
			wantErr:  false,
		},
		// More tests...
	}
	// Test execution loop...
}
```

#### 2. Nested Loop Tests

```go
func TestLoopControl_Nested(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected core.Value
		wantErr  bool
	}{
		{
			name: "break in inner loop only",
			input: `
				outer: 0
				inner: 0
				loop 3 [
					outer: outer + 1
					loop 3 [
						inner: inner + 1
						break
					]
				]
				outer
			`,
			expected: value.NewIntVal(3),  // Outer completes
			wantErr:  false,
		},
		// More tests...
	}
	// Test execution loop...
}
```

#### 3. Error Case Tests

```go
func TestLoopControl_Errors(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		errID   string  // Expected error ID
	}{
		{
			name:    "break outside loop",
			input:   "break",
			wantErr: true,
			errID:   verror.ErrIDBreakOutsideLoop,
		},
		{
			name:    "continue outside loop",
			input:   "continue",
			wantErr: true,
			errID:   verror.ErrIDContinueOutsideLoop,
		},
		{
			name:    "break in function called from loop - boundary blocks it",
			input:   "fn: function [] [break]\nloop 3 [fn]",
			wantErr: true,
			errID:   verror.ErrIDBreakOutsideLoop,
		},
		{
			name:    "continue in function called from loop - boundary blocks it",
			input:   "fn: function [] [continue]\nloop 3 [fn]",
			wantErr: true,
			errID:   verror.ErrIDContinueOutsideLoop,
		},
		// More tests...
	}
	// Test execution with error ID verification...
}
```

#### 3b. Transparent Block Tests (do/reduce/compose)

```go
func TestLoopControl_TransparentBlocks(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected core.Value
		wantErr  bool
	}{
		{
			name:     "break in do block - works (no boundary)",
			input:    "x: 0\nloop 10 [x: x + 1\ndo [when (= x 3) [break]]]\nx",
			expected: value.NewIntVal(3),
			wantErr:  false,
		},
		{
			name:     "continue in do block - works (no boundary)",
			input:    "x: 0\nloop 3 [x: x + 1\ndo [continue]\nx: x + 100]\nx",
			expected: value.NewIntVal(3),  // No +100 additions
			wantErr:  false,
		},
		// More tests...
	}
	// Test execution loop...
}
```

#### 4. Integration Tests

Tests for:
- Break/continue with --with-index refinement
- Function calls containing break/continue (should error - boundary exists)
- Reduce/compose/do transparency (should work - no boundary)
- Complex nested scenarios
- Counter and accumulator patterns
- Difference between function boundaries and transparent blocks

### Test Coverage Goals

- **Line Coverage:** >95% of new code
- **Branch Coverage:** 100% of control flow paths
- **Edge Cases:** All documented scenarios tested
- **Error Cases:** All error conditions verified

### Validation Checklist

After all steps:
- [ ] All loop_control_test.go tests pass
- [ ] No regressions in existing tests
- [ ] Break works in loop, while, foreach
- [ ] Continue works in loop, while, foreach
- [ ] Nested loops behave correctly
- [ ] Error messages are clear and accurate
- [ ] Break/continue in functions return error (don't cross function boundary)
- [ ] Reduce/compose/do propagate control signals (transparent, no boundary)

## Potential Challenges and Mitigations

### Challenge 1: Error Annotation Overhead

**Issue:** ErrThrow errors might get annotated with "near" and "where" context unnecessarily.

**Mitigation:** 
- Accept minor overhead - it's bounded and rare
- ErrThrow errors are short-lived (caught by loops)
- Only uncaught signals are annotated (rare error case)
- Alternative: Skip annotation for ErrThrow category (requires evaluator changes)

**Decision:** Accept overhead for Phase 1 simplicity.

### Challenge 2: Foreach Variable Binding State

**Issue:** When continue is triggered in foreach, variables are already bound for current iteration.

**Mitigation:**
- Variables will have values from current iteration when continue is called
- This is expected behavior (user can use values before calling continue)
- Document this behavior clearly

**Example:**
```viro
foreach [1 2 3] 'x [
    print x        ; Prints value
    continue       ; Skips rest, x is already bound
    print "never"
]
```

### Challenge 3: While Condition Re-evaluation

**Issue:** With while, continue should re-evaluate the condition before next iteration.

**Mitigation:**
- Current while structure already does this
- Continue just skips to next loop iteration
- Loop naturally re-evaluates condition ✓

**Verification Test:**
```viro
x: 0
while [x < 3] [
    x: x + 1
    continue
    print "never"
]
; x should be 3, condition evaluated each iteration
```

### Challenge 4: Function Boundaries

**Issue:** Break/continue should not cross function call boundaries.

**Mitigation:**
- Function call handling converts ErrThrow signals to ErrScript errors
- Clear error messages: "break called outside of loop"
- Documentation explaining break/continue don't work across function boundaries
- Test cases demonstrating the distinction between functions and transparent blocks

**Example - Functions (boundary exists):**
```viro
fn: function [] [break]
loop 3 [fn]  ; Error: break called outside of loop
```

**Example - Transparent blocks (no boundary):**
```viro
loop 3 [
    do [break]  ; Works: break exits loop
]
```

## Phase 2 Considerations: Multi-Level Support

### Design for --levels Refinement

**Syntax:**
```viro
loop 3 [
    loop 3 [
        break --levels 2  ; Break out of both loops
    ]
]
```

**Implementation Approach:**

1. **Error Payload:** Use Args[0] to carry level count
   ```go
   verror.NewError(ErrThrow, ErrIDBreak, [3]string{"2", "", ""})
   ```

2. **Native Function Changes:**
   ```go
   func Break(args []Value, refValues map[string]Value, eval Evaluator) (Value, error) {
       levels := int64(1)
       if levelsVal, ok := refValues["levels"]; ok && levelsVal.GetType() != value.TypeInteger {
           levels, _ = value.AsIntValue(levelsVal)
           if levels < 1 {
               return value.NewNoneVal(), verror.NewScriptError(
                   verror.ErrIDInvalidOperation,
                   [3]string{"break --levels must be >= 1", "", ""},
               )
           }
       }
       return value.NewNoneVal(), verror.NewError(
           verror.ErrThrow,
           verror.ErrIDBreak,
           [3]string{fmt.Sprintf("%d", levels), "", ""},
       )
   }
   ```

3. **Loop Handling Changes:**
   ```go
   isControl, signalType := isLoopControlSignal(err)
   if isControl {
       levels := extractLevels(err)  // Parse Args[0]
       if levels > 1 {
           // Re-throw with decremented count
           verr, _ := err.(*verror.Error)
           return value.NewNoneVal(), verror.NewError(
               verror.ErrThrow,
               verr.ID,
               [3]string{fmt.Sprintf("%d", levels-1), "", ""},
           )
       }
       // levels == 1, handle normally
       if signalType == "break" {
           return value.NewNoneVal(), nil
       }
       continue
   }
   ```

4. **Helper Function:**
   ```go
   func extractLevels(err error) int64 {
       verr, ok := err.(*verror.Error)
       if !ok || verr.Args[0] == "" {
           return 1
       }
       levels, parseErr := strconv.ParseInt(verr.Args[0], 10, 64)
       if parseErr != nil {
           return 1
       }
       return levels
   }
   ```

### Additional Validation for Phase 2

- Error if --levels exceeds actual loop depth
- Clear error message: "break --levels 3 but only 2 loops active"
- Requires loop depth tracking (count active loops)

### Phase 2 Test Cases

```go
{
    name: "break --levels 2 in nested loop",
    input: `
        x: 0
        loop 3 [
            loop 3 [
                x: x + 1
                when (= x 2) [break --levels 2]
            ]
            x: x + 100  ; Never executes
        ]
        x
    `,
    expected: value.NewIntVal(2),
    wantErr:  false,
},
```

## Viro Guidelines Reference

### Coding Standards Followed

1. **No comments in code** - All documentation in plan and package docs
2. **Constructor functions** - Use `value.NewNoneVal()`, `value.NewIntVal()`, `value.NewError()`
3. **Error handling** - Use `verror.NewScriptError()`, `verror.NewError()` with category/ID/args
4. **Table-driven tests** - All tests follow `[]struct{name, input, expected, wantErr}` pattern
5. **TDD approach** - Write tests first, implement to make them pass

### Viro Naming Conventions

- Native function names: lowercase, hyphenated if needed (N/A for single words)
- Error IDs: kebab-case (`break-outside-loop`)
- Category constants: PascalCase (`ErrThrow`, `ErrScript`)
- Query functions: suffix with `?` (N/A for this feature)
- Modification functions: suffix with `!` (N/A for this feature)

### Architecture Alignment

- **Value system:** Returns `core.Value` from natives
- **Type-based dispatch:** Not applicable (no type-specific behavior)
- **Error categories:** Uses existing ErrThrow (0) and ErrScript (300)
- **Frame system:** No new frames, uses existing loop frame management
- **Evaluator integration:** Minimal - only top-level conversion added

## Summary

This plan implements break and continue loop control using an error-based signaling approach that:

1. **Leverages existing infrastructure** - Uses error propagation, no evaluator state changes
2. **Maintains separation of concerns** - Loops handle control flow, functions create boundaries, transparent blocks propagate
3. **Provides clear semantics** - Break exits, continue skips, both affect innermost loop only
4. **Respects function boundaries** - Break/continue cannot cross function call boundaries, matching behavior of mainstream languages
5. **Preserves transparent blocks** - `do`, `reduce`, `compose` remain transparent and propagate control signals
6. **Handles edge cases** - Nested loops, function boundaries, transparent blocks all work correctly
7. **Enables future extension** - Phase 2 multi-level support is straightforward
8. **Follows TDD** - Tests written first, implementation makes them pass
9. **Aligns with Viro** - Coding standards, naming conventions, error handling all consistent

**Key Behaviors:**
- `loop 3 [break]` → exits loop ✓
- `loop 3 [do [break]]` → exits loop (do is transparent) ✓
- `fn: function [] [break]\nloop 3 [fn]` → error: "break called outside of loop" ✓

**Next Step:** Begin implementation at Step 1 (Error Infrastructure).
