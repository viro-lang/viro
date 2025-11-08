# Plan 032: Multi-Level Break and Continue with --levels Refinement

## Feature Summary

**Phase 2 Enhancement**: Add `--levels N` refinement to `break` and `continue` natives to enable multi-level loop control, allowing exit from or continuation of nested loops at any depth.

**Key Capabilities:**
- `break --levels 2` exits 2 nested loops
- `continue --levels 2` skips to the second outer loop's next iteration
- Default behavior: `--levels 1` (maintains Phase 1 compatibility)
- Error detection: levels < 1 rejected at call site
- Error detection: levels exceeding actual loop depth (optional validation)

**Phase 1 Foundation (Already Implemented):**
- Basic single-level break/continue using ErrThrow category signals
- Loop natives catch and handle control signals via `isLoopControlSignal()`
- Function boundaries convert signals to ErrScript errors
- Transparent blocks (do, reduce, compose) propagate signals unchanged
- Full test coverage for single-level behavior

**This Plan Scope:**
- Extend Break/Continue natives to accept `--levels` refinement
- Modify loop handlers to decrement and re-throw multi-level signals
- Add helper function to extract level count from error payload
- Comprehensive test coverage for multi-level scenarios
- Optional loop depth tracking for validation

## Research Findings

### Phase 1 Implementation Analysis

**Signal Mechanism (from `internal/native/control.go`):**
```go
// Phase 1: Single-level signals
verror.NewError(verror.ErrThrow, verror.ErrIDBreak, [3]string{})
verror.NewError(verror.ErrThrow, verror.ErrIDContinue, [3]string{})
```

**Current Loop Handling Pattern:**
```go
result, err = eval.DoBlock(block.Elements, block.Locations())
if err != nil {
    isControl, signalType := isLoopControlSignal(err)
    if isControl {
        if signalType == "break" {
            return value.NewNoneVal(), nil  // Exit loop
        }
        continue  // Continue to next iteration
    }
    return value.NewNoneVal(), err  // Propagate other errors
}
```

### Error Payload Structure

From `internal/verror/error.go`:
```go
type Error struct {
    Category ErrorCategory
    Code     int
    ID       string
    Args     [3]string  // Three-element array for message interpolation
    Near     string
    Where    []string
    Message  string
    File     string
    Line     int
    Column   int
}
```

**Key Insight:** `Args[0]` can carry the level count as a string:
- Args[0] = "1" → single-level (default, backward compatible)
- Args[0] = "2" → exit/continue 2 levels
- Args[0] = "3" → exit/continue 3 levels
- Args[0] = "" → treated as "1" for backward compatibility

### Refinement Patterns in Viro

From existing natives (trace, loop, foreach):
```go
// Pattern 1: Boolean refinement (--on, --off)
if val, ok := refValues["on"]; ok && ToTruthy(val) {
    // Handle --on
}

// Pattern 2: Integer refinement (--step-level N)
if stepLevelVal, ok := refValues["step-level"]; ok && stepLevelVal.GetType() != value.TypeNone {
    if stepLevelVal.GetType() != value.TypeInteger {
        return value.NewNoneVal(), verror.NewScriptError(
            verror.ErrIDTypeMismatch,
            [3]string{"--step-level requires integer", "", ""},
        )
    }
    stepLevel, _ := value.AsIntValue(stepLevelVal)
    // Validate and use stepLevel
}

// Pattern 3: Word refinement (--with-index 'i)
indexVal, hasIndexRef := refValues["with-index"]
if hasIndexRef && indexVal.GetType() != value.TypeNone {
    if !value.IsWord(indexVal.GetType()) {
        return value.NewNoneVal(), verror.NewScriptError(
            verror.ErrIDTypeMismatch,
            [3]string{"--with-index requires a word", "", ""},
        )
    }
    indexWord, _ := value.AsWordValue(indexVal)
}
```

**For --levels:** Follow Pattern 2 (integer refinement with validation)

### Loop Architecture Review

**Three loop constructs** all use the same DoBlock-based error propagation:

1. **Loop** (lines 89-164): Fixed iteration count
2. **While** (lines 166-243): Conditional with re-evaluated or static condition
3. **Foreach** (lines 560-680): Series iteration with variable binding

All three:
- Call `eval.DoBlock()` for body execution
- Check errors with `isLoopControlSignal(err)`
- Handle break (exit) and continue (next iteration) uniformly
- Propagate non-control errors unchanged

**Integration Point:** All three need identical multi-level handling logic.

## Architecture Overview

### Multi-Level Signal Mechanism

**Design Decision:** Use Args[0] to encode level count in control flow signals.

**Rationale:**
- Minimal changes to existing error infrastructure
- Backward compatible (empty Args[0] defaults to 1)
- No new error categories or types needed
- Leverages existing error propagation chain

**How It Works:**

#### 1. Signal Creation (in Break/Continue natives)

```go
func Break(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
    if len(args) != 0 {
        return value.NewNoneVal(), arityError("break", 0, len(args))
    }
    
    levels := int64(1)  // Default
    if levelsVal, ok := refValues["levels"]; ok && levelsVal.GetType() != value.TypeNone {
        if levelsVal.GetType() != value.TypeInteger {
            return value.NewNoneVal(), verror.NewScriptError(
                verror.ErrIDTypeMismatch,
                [3]string{"--levels requires integer", "", ""},
            )
        }
        levels, _ = value.AsIntValue(levelsVal)
        if levels < 1 {
            return value.NewNoneVal(), verror.NewScriptError(
                verror.ErrIDInvalidOperation,
                [3]string{"--levels must be >= 1", "", ""},
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

#### 2. Signal Handling (in Loop natives)

```go
result, err = eval.DoBlock(block.Elements, block.Locations())
if err != nil {
    isControl, signalType := isLoopControlSignal(err)
    if isControl {
        levels := extractLevels(err)  // Parse Args[0]
        
        if levels > 1 {
            // Re-throw with decremented count
            verr, _ := err.(*verror.Error)
            newErr := verror.NewError(
                verror.ErrThrow,
                verr.ID,
                [3]string{fmt.Sprintf("%d", levels-1), "", ""},
            )
            return value.NewNoneVal(), newErr
        }
        
        // levels == 1, consume the signal
        if signalType == "break" {
            return value.NewNoneVal(), nil
        }
        continue  // continue signal
    }
    return value.NewNoneVal(), err
}
```

#### 3. Level Extraction Helper

```go
func extractLevels(err error) int64 {
    verr, ok := err.(*verror.Error)
    if !ok || verr.Args[0] == "" {
        return 1  // Default for backward compatibility
    }
    
    levels, parseErr := strconv.ParseInt(verr.Args[0], 10, 64)
    if parseErr != nil {
        return 1  // Fallback on parse error
    }
    
    return levels
}
```

### Nested Loop Behavior Examples

**Example 1: Basic Multi-Level Break**
```viro
x: 0
loop 3 [          ; Outer loop
    loop 3 [      ; Inner loop
        x: x + 1
        when (= x 2) [break --levels 2]  ; Exit both loops
    ]
    x: x + 100    ; Never executes
]
x  ; Result: 2
```

**Flow:**
1. Inner loop: x becomes 2, break --levels 2 throws signal with Args[0]="2"
2. Inner loop catches signal, sees levels=2, decrements to 1, re-throws
3. Outer loop catches signal, sees levels=1, exits
4. Result: 2 ✓

**Example 2: Multi-Level Continue**
```viro
x: 0
loop 3 --with-index 'i [     ; Outer: i=0,1,2
    loop 3 --with-index 'j [  ; Inner: j=0,1,2
        x: x + 1
        when (= x 3) [continue --levels 2]  ; Skip to outer's next iteration
        x: x + 10
    ]
    x: x + 100
]
x
```

**Flow:**
1. Iterations: (i=0,j=0): x=1→11, (i=0,j=1): x=12→22, (i=0,j=2): x=23, continue --levels 2
2. Inner loop catches continue signal, levels=2, decrements to 1, re-throws
3. Outer loop catches continue signal, levels=1, skips to next outer iteration (i=1)
4. Continues with i=1, j=0...
5. Result depends on full execution ✓

**Example 3: Function Boundary Blocks Multi-Level**
```viro
loop 3 [
    loop 3 [
        f: fn [] [break --levels 2]
        f  ; Error: break called outside of loop
    ]
]
```

**Flow:**
1. f is called, evaluates break --levels 2 inside function body
2. Function boundary (callUserDefinedFunction) converts ErrThrow to ErrScript
3. Error propagates to inner loop as script error (not control signal)
4. Error propagates to outer loop, user sees error ✓

**Example 4: Transparent Blocks Preserve Multi-Level**
```viro
x: 0
loop 3 [
    loop 3 [
        x: x + 1
        do [when (= x 2) [break --levels 2]]  ; Works - do is transparent
    ]
    x: x + 100
]
x  ; Result: 2
```

**Flow:**
1. do evaluates block, break --levels 2 signal returns through do unchanged
2. Inner loop catches signal, levels=2, decrements, re-throws
3. Outer loop catches signal, levels=1, exits
4. Result: 2 ✓

### Why This Design Works

1. **Backward Compatible:** Empty Args[0] or "1" behaves like Phase 1 (single-level)
2. **Function Boundaries:** Conversion to ErrScript happens before level checking (multi-level signals can't cross functions)
3. **Transparent Blocks:** do/reduce/compose propagate errors unchanged (including Args), so multi-level works
4. **Decremental Propagation:** Each loop decrements the count, ensuring signal reaches correct depth
5. **Simple Implementation:** No evaluator changes, no new error categories, minimal modifications

## Implementation Roadmap

### Step 1: Add Level Extraction Helper

**File:** `internal/native/control.go`

Add helper function after `isLoopControlSignal` (around line 356):

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
	
	if levels < 1 {
		return 1
	}
	
	return levels
}
```

**Purpose:** Extract level count from error Args[0] with fallback to 1.

**Validation:** 
- Build succeeds
- Helper compiles
- No behavior changes (not yet called)

### Step 2: Write Multi-Level Test Cases First (TDD)

**File:** `test/contract/loop_control_test.go`

Add new test function after existing tests (around line 356):

```go
func TestLoopControl_MultiLevel(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected core.Value
		wantErr  bool
	}{
		{
			name: "break --levels 2 in nested loop",
			input: `
				x: 0
				loop 3 [
					loop 3 [
						x: x + 1
						when (= x 2) [break --levels 2]
					]
					x: x + 100
				]
				x
			`,
			expected: value.NewIntVal(2),
			wantErr:  false,
		},
		{
			name: "break --levels 3 in triple-nested loop",
			input: `
				x: 0
				loop 2 [
					loop 2 [
						loop 2 [
							x: x + 1
							when (= x 3) [break --levels 3]
						]
						x: x + 100
					]
					x: x + 1000
				]
				x
			`,
			expected: value.NewIntVal(3),
			wantErr:  false,
		},
		{
			name: "continue --levels 2 in nested loop",
			input: `
				x: 0
				loop 3 --with-index 'i [
					loop 3 --with-index 'j [
						x: x + 1
						when (and (= i 0) (= j 2)) [continue --levels 2]
						x: x + 10
					]
					x: x + 100
				]
				x
			`,
			expected: value.NewIntVal(323),
			wantErr:  false,
		},
		{
			name: "break --levels 1 is same as break",
			input: `
				x: 0
				loop 3 [
					loop 3 [
						x: x + 1
						when (= x 2) [break --levels 1]
					]
					x: x + 100
				]
				x
			`,
			expected: value.NewIntVal(302),
			wantErr:  false,
		},
		{
			name: "continue --levels 1 is same as continue",
			input: `
				x: 0
				loop 3 [
					x: x + 1
					continue --levels 1
					x: x + 100
				]
				x
			`,
			expected: value.NewIntVal(3),
			wantErr:  false,
		},
		{
			name: "break --levels in while loops",
			input: `
				x: 0
				while [x < 10] [
					while [x < 10] [
						x: x + 1
						when (= x 3) [break --levels 2]
					]
					x: x + 100
				]
				x
			`,
			expected: value.NewIntVal(3),
			wantErr:  false,
		},
		{
			name: "break --levels in foreach",
			input: `
				x: 0
				foreach [1 2 3] 'a [
					foreach [10 20 30] 'b [
						x: x + a + b
						when (= x 33) [break --levels 2]
					]
					x: x + 100
				]
				x
			`,
			expected: value.NewIntVal(33),
			wantErr:  false,
		},
		{
			name: "break --levels with transparent blocks",
			input: `
				x: 0
				loop 3 [
					loop 3 [
						x: x + 1
						do [when (= x 2) [break --levels 2]]
					]
					x: x + 100
				]
				x
			`,
			expected: value.NewIntVal(2),
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Evaluate(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if !result.Equals(tt.expected) {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}
```

Add error test cases:

```go
func TestLoopControl_MultiLevelErrors(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		errID   string
	}{
		{
			name:    "break --levels 0 is invalid",
			input:   "loop 3 [break --levels 0]",
			wantErr: true,
			errID:   verror.ErrIDInvalidOperation,
		},
		{
			name:    "break --levels -1 is invalid",
			input:   "loop 3 [break --levels -1]",
			wantErr: true,
			errID:   verror.ErrIDInvalidOperation,
		},
		{
			name:    "continue --levels 0 is invalid",
			input:   "loop 3 [continue --levels 0]",
			wantErr: true,
			errID:   verror.ErrIDInvalidOperation,
		},
		{
			name:    "break --levels requires integer",
			input:   "loop 3 [break --levels \"two\"]",
			wantErr: true,
			errID:   verror.ErrIDTypeMismatch,
		},
		{
			name:    "continue --levels requires integer",
			input:   "loop 3 [continue --levels \"two\"]",
			wantErr: true,
			errID:   verror.ErrIDTypeMismatch,
		},
		{
			name:    "break --levels 2 in function crosses boundary",
			input:   "loop 3 [loop 3 [f: fn [] [break --levels 2]\nf]]",
			wantErr: true,
			errID:   verror.ErrIDBreakOutsideLoop,
		},
		{
			name:    "continue --levels 2 in function crosses boundary",
			input:   "loop 3 [loop 3 [f: fn [] [continue --levels 2]\nf]]",
			wantErr: true,
			errID:   verror.ErrIDContinueOutsideLoop,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Evaluate(tt.input)

			if !tt.wantErr {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				return
			}

			if err == nil {
				t.Errorf("Expected error but got none, result: %v", result)
				return
			}

			if verr, ok := err.(*verror.Error); ok {
				if verr.ID != tt.errID {
					t.Errorf("Expected error ID %s, got %s", tt.errID, verr.ID)
				}
			} else {
				t.Errorf("Expected verror.Error, got %T", err)
			}
		})
	}
}
```

**Validation:**
- Tests compile
- All tests FAIL (--levels not implemented yet)
- Error messages indicate "no value for word: levels" or similar

### Step 3: Modify Break Native to Accept --levels

**File:** `internal/native/control.go`

Replace the `Break` function (lines 850-855) with:

```go
func Break(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 0 {
		return value.NewNoneVal(), arityError("break", 0, len(args))
	}
	
	levels := int64(1)
	if levelsVal, ok := refValues["levels"]; ok && levelsVal.GetType() != value.TypeNone {
		if levelsVal.GetType() != value.TypeInteger {
			return value.NewNoneVal(), verror.NewScriptError(
				verror.ErrIDTypeMismatch,
				[3]string{"--levels requires integer", "", ""},
			)
		}
		levels, _ = value.AsIntValue(levelsVal)
		if levels < 1 {
			return value.NewNoneVal(), verror.NewScriptError(
				verror.ErrIDInvalidOperation,
				[3]string{"--levels must be >= 1", "", ""},
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

**Note:** Add import for `fmt` if not already present.

**Validation:**
- Build succeeds
- Error validation tests PASS (--levels 0, --levels -1, type mismatch)
- Multi-level tests still FAIL (loops don't handle multi-level yet)

### Step 4: Modify Continue Native to Accept --levels

**File:** `internal/native/control.go`

Replace the `Continue` function (lines 857-862) with:

```go
func Continue(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 0 {
		return value.NewNoneVal(), arityError("continue", 0, len(args))
	}
	
	levels := int64(1)
	if levelsVal, ok := refValues["levels"]; ok && levelsVal.GetType() != value.TypeNone {
		if levelsVal.GetType() != value.TypeInteger {
			return value.NewNoneVal(), verror.NewScriptError(
				verror.ErrIDTypeMismatch,
				[3]string{"--levels requires integer", "", ""},
			)
		}
		levels, _ = value.AsIntValue(levelsVal)
		if levels < 1 {
			return value.NewNoneVal(), verror.NewScriptError(
				verror.ErrIDInvalidOperation,
				[3]string{"--levels must be >= 1", "", ""},
			)
		}
	}
	
	return value.NewNoneVal(), verror.NewError(
		verror.ErrThrow,
		verror.ErrIDContinue,
		[3]string{fmt.Sprintf("%d", levels), "", ""},
	)
}
```

**Validation:**
- Build succeeds
- Continue error validation tests PASS
- Multi-level tests still FAIL (loops don't handle multi-level yet)

### Step 5: Modify Loop Native for Multi-Level Handling

**File:** `internal/native/control.go`

Modify the `Loop` function error handling (lines 150-160). Replace:

```go
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
```

With:

```go
		result, err = eval.DoBlock(block.Elements, block.Locations())
		if err != nil {
			isControl, signalType := isLoopControlSignal(err)
			if isControl {
				levels := extractLevels(err)
				
				if levels > 1 {
					verr, _ := err.(*verror.Error)
					newErr := verror.NewError(
						verror.ErrThrow,
						verr.ID,
						[3]string{fmt.Sprintf("%d", levels-1), "", ""},
					)
					return value.NewNoneVal(), newErr
				}
				
				if signalType == "break" {
					return value.NewNoneVal(), nil
				}
				continue
			}
			return value.NewNoneVal(), err
		}
```

**Validation:**
- Build succeeds
- Loop multi-level tests START to PASS
- While and foreach multi-level tests still FAIL (not modified yet)

### Step 6: Modify While Native for Multi-Level Handling

**File:** `internal/native/control.go`

**Location 1:** While with block condition (lines 210-220)

Replace:

```go
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
```

With:

```go
			result, err = eval.DoBlock(bodyBlock.Elements, bodyBlock.Locations())
			if err != nil {
				isControl, signalType := isLoopControlSignal(err)
				if isControl {
					levels := extractLevels(err)
					
					if levels > 1 {
						verr, _ := err.(*verror.Error)
						newErr := verror.NewError(
							verror.ErrThrow,
							verr.ID,
							[3]string{fmt.Sprintf("%d", levels-1), "", ""},
						)
						return value.NewNoneVal(), newErr
					}
					
					if signalType == "break" {
						return value.NewNoneVal(), nil
					}
					continue
				}
				return value.NewNoneVal(), err
			}
```

**Location 2:** While with non-block condition (lines 228-238)

Replace:

```go
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
```

With:

```go
			result, err = eval.DoBlock(bodyBlock.Elements, bodyBlock.Locations())
			if err != nil {
				isControl, signalType := isLoopControlSignal(err)
				if isControl {
					levels := extractLevels(err)
					
					if levels > 1 {
						verr, _ := err.(*verror.Error)
						newErr := verror.NewError(
							verror.ErrThrow,
							verr.ID,
							[3]string{fmt.Sprintf("%d", levels-1), "", ""},
						)
						return value.NewNoneVal(), newErr
					}
					
					if signalType == "break" {
						return value.NewNoneVal(), nil
					}
					continue
				}
				return value.NewNoneVal(), err
			}
```

**Validation:**
- Build succeeds
- While multi-level tests PASS
- Foreach multi-level tests still FAIL

### Step 7: Modify Foreach Native for Multi-Level Handling

**File:** `internal/native/control.go`

Modify the foreach iteration loop (lines 663-675). Replace:

```go
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
```

With:

```go
		result, err = eval.DoBlock(bodyBlock.Elements, bodyBlock.Locations())

		if err != nil {
			isControl, signalType := isLoopControlSignal(err)
			if isControl {
				levels := extractLevels(err)
				
				if levels > 1 {
					verr, _ := err.(*verror.Error)
					newErr := verror.NewError(
						verror.ErrThrow,
						verr.ID,
						[3]string{fmt.Sprintf("%d", levels-1), "", ""},
					)
					return value.NewNoneVal(), newErr
				}
				
				if signalType == "break" {
					return value.NewNoneVal(), nil
				}
				iteration++
				continue
			}
			return value.NewNoneVal(), err
		}
```

**Validation:**
- Build succeeds
- ALL multi-level tests PASS
- Error tests PASS
- Phase 1 tests still PASS (backward compatibility verified)

### Step 8: Comprehensive Testing

Run full test suite:

```bash
# Multi-level specific tests
go test -v ./test/contract -run TestLoopControl_MultiLevel

# Error validation tests
go test -v ./test/contract -run TestLoopControl_MultiLevelErrors

# All loop control tests (Phase 1 + Phase 2)
go test -v ./test/contract -run TestLoopControl

# Regression check: all control flow tests
go test -v ./test/contract -run TestControl

# Full integration suite
go test ./...
```

**Expected Results:**
- All TestLoopControl_MultiLevel tests PASS
- All TestLoopControl_MultiLevelErrors tests PASS
- All Phase 1 tests still PASS (backward compatibility)
- No regressions in control flow or integration tests

### Step 9: Manual Validation (Optional)

Create manual test script `tmp/test_multilevel_break.viro`:

```viro
; Test 1: Basic multi-level break
print "Test 1: Multi-level break"
x: 0
loop 3 [
    loop 3 [
        x: x + 1
        when (= x 2) [break --levels 2]
    ]
    x: x + 100
]
print ["Expected: 2, Got:" x]

; Test 2: Multi-level continue
print "Test 2: Multi-level continue"
y: 0
loop 3 --with-index 'i [
    loop 3 --with-index 'j [
        y: y + 1
        when (and (= i 0) (= j 2)) [continue --levels 2]
        y: y + 10
    ]
    y: y + 100
]
print ["Expected: 323, Got:" y]

; Test 3: Triple-nested
print "Test 3: Triple-nested break"
z: 0
loop 2 [
    loop 2 [
        loop 2 [
            z: z + 1
            when (= z 3) [break --levels 3]
        ]
        z: z + 100
    ]
    z: z + 1000
]
print ["Expected: 3, Got:" z]
```

Run:
```bash
./viro tmp/test_multilevel_break.viro
```

Expected output:
```
Test 1: Multi-level break
Expected: 2, Got: 2
Test 2: Multi-level continue
Expected: 323, Got: 323
Test 3: Triple-nested break
Expected: 3, Got: 3
```

## Integration Points

### 1. Break and Continue Natives

**Location:** `internal/native/control.go`

**Changes:**
- Add `--levels` refinement parameter handling
- Validate levels >= 1 at call site
- Encode levels in Args[0] of ErrThrow signal

**Impact:** 
- Backward compatible (default levels=1)
- Clear error messages for invalid inputs
- No change to function signature

### 2. All Loop Constructs

**Affected Functions:**
- Loop (lines 89-164)
- While (lines 166-243, two locations)
- Foreach (lines 560-680)

**Changes:**
- Extract level count from error Args[0]
- If levels > 1: decrement and re-throw
- If levels == 1: consume signal (break/continue)

**Impact:**
- Identical pattern across all three loops
- No change to loop APIs or semantics
- Phase 1 behavior preserved (levels=1)

### 3. Error Payload Protocol

**Location:** `internal/verror/error.go`

**Usage:**
- Args[0] carries level count as string ("1", "2", "3", ...)
- Empty Args[0] defaults to "1" (backward compatible)
- No new error categories or IDs needed

**Impact:**
- Minimal - leverages existing error structure
- No breaking changes to error system
- Future-proof for additional metadata

### 4. Function Boundary Conversion

**Location:** `internal/eval/evaluator.go` (callUserDefinedFunction)

**No Changes Required:**
- Existing conversion logic handles multi-level signals correctly
- verror.ConvertLoopControlSignal() converts ANY ErrThrow/break or continue
- Level count irrelevant once signal crosses function boundary

**Behavior:**
```viro
f: fn [] [break --levels 2]
loop 3 [loop 3 [f]]  ; Error: break called outside of loop
```
- Function converts ErrThrow→ErrScript regardless of levels
- Loop sees script error, not control signal
- Multi-level can't cross function boundaries ✓

### 5. Transparent Blocks

**Location:** `internal/native/control.go` (Do, Reduce, Compose)

**No Changes Required:**
- Transparent blocks propagate errors unchanged
- Args[0] preserved during propagation
- Multi-level signals work through transparent blocks

**Behavior:**
```viro
loop 3 [
    loop 3 [
        do [break --levels 2]  ; Works - do propagates signal
    ]
]
```

## Testing Strategy

### Test Organization

**Primary File:** `test/contract/loop_control_test.go`

**New Test Functions:**
1. `TestLoopControl_MultiLevel` - Multi-level functionality tests
2. `TestLoopControl_MultiLevelErrors` - Error validation tests

**Follow Existing Patterns:**
- Table-driven tests
- Clear descriptive names
- Expected vs actual value comparison
- Error ID verification for error cases

### Test Categories

#### 1. Multi-Level Break Tests

**Coverage:**
- 2-level break in nested loop (verify outer exits)
- 3-level break in triple-nested loop (verify all exit)
- Break --levels 1 equivalent to break (backward compatibility)
- Break --levels in while loops
- Break --levels in foreach loops
- Break --levels with transparent blocks (do)

**Validation:**
- Counter variables verify correct exit point
- Result value confirms expected iterations
- No extra iterations after break

#### 2. Multi-Level Continue Tests

**Coverage:**
- 2-level continue in nested loop (verify outer continues)
- 3-level continue in triple-nested loop
- Continue --levels 1 equivalent to continue (backward compatibility)
- Continue with --with-index refinement (verify correct iteration)
- Continue in while loops (condition re-evaluation)
- Continue in foreach loops (next element)

**Validation:**
- Counter variables track skipped vs executed code
- Index variables verify correct continuation point
- Result values confirm expected accumulation

#### 3. Error Validation Tests

**Coverage:**
- `--levels 0` → error (invalid)
- `--levels -1` → error (invalid)
- `--levels "string"` → type mismatch error
- `break --levels 2` in function → boundary blocks it
- `continue --levels 2` in function → boundary blocks it

**Validation:**
- Error is raised (not nil)
- Error has correct error ID
- Error message is clear and actionable

#### 4. Edge Cases

**Coverage:**
- Mixed loop types (loop + while + foreach)
- Very deep nesting (5+ levels)
- Levels equal to actual depth (boundary condition)
- Transparent blocks preserving multi-level (do, compose, reduce)
- Combination with --with-index refinement

**Validation:**
- Complex scenarios work correctly
- No infinite loops or stack issues
- Performance acceptable for deep nesting

### Test Coverage Goals

- **Line Coverage:** >95% of modified code
- **Branch Coverage:** 100% of control flow paths
  - levels == 1 (consume signal)
  - levels > 1 (decrement and re-throw)
  - levels < 1 (validation error)
- **Error Cases:** All error conditions verified
- **Regression:** All Phase 1 tests still pass

### Validation Checklist

After implementation:
- [ ] All TestLoopControl_MultiLevel tests pass
- [ ] All TestLoopControl_MultiLevelErrors tests pass
- [ ] All Phase 1 tests still pass (backward compatibility)
- [ ] No regressions in control flow tests
- [ ] No regressions in integration tests
- [ ] Error messages are clear and helpful
- [ ] Function boundaries block multi-level signals
- [ ] Transparent blocks preserve multi-level signals
- [ ] Manual test script produces expected output
- [ ] Performance acceptable for deep nesting (10+ levels)

## Potential Challenges and Mitigations

### Challenge 1: String Conversion Overhead

**Issue:** Converting int64 to string and back for each loop level adds overhead.

**Mitigation:**
- Overhead is minimal (single sprintf and parseInt per loop level)
- Multi-level breaks are rare in practice (typically 2-3 levels max)
- Alternative: Use Args[1] for numeric representation, but adds complexity
- **Decision:** Accept minimal overhead for implementation simplicity

### Challenge 2: Loop Depth Validation

**Issue:** User could specify `--levels 10` but only be 2 loops deep.

**Current Behavior:**
- Signal propagates to top level with levels=8 remaining
- Converted to ErrScript/"break-outside-loop" at top level
- User sees error but message doesn't mention levels

**Mitigation Options:**

**Option A (Current Plan):** No special handling
- Pros: Simple, minimal code
- Cons: Error message doesn't mention exceeded depth

**Option B:** Track loop depth in evaluator
- Pros: Can provide "exceeded depth" error
- Cons: Requires evaluator changes, adds state tracking complexity

**Option C:** Include remaining levels in error message
- Modify ConvertLoopControlSignal to check Args[0]
- If levels > 1, message becomes "break --levels N exceeds actual loop depth"
- Pros: Better error message, no evaluator changes
- Cons: Slight complexity in conversion logic

**Decision:** Start with Option A (current plan). If user feedback indicates confusing errors, implement Option C in future enhancement.

### Challenge 3: Backward Compatibility

**Issue:** Phase 1 code might break with Phase 2 changes.

**Mitigation:**
- Default levels=1 maintains exact Phase 1 behavior
- Empty Args[0] defaults to "1" via extractLevels()
- All Phase 1 tests included in regression suite
- No changes to loop APIs or semantics

**Verification:**
- Run Phase 1 test suite (TestLoopControl_BreakBasic, etc.)
- All tests must PASS without modification
- Manual spot-check: simple break/continue still works

### Challenge 4: Error Annotation Overhead

**Issue:** Multi-level signals might accumulate Near/Where context as they propagate.

**Current Behavior:**
- ErrThrow signals are re-created at each level (not accumulated)
- Each re-throw creates fresh error with no Near/Where
- Only final conversion to ErrScript gets annotated

**Impact:**
- No accumulation overhead ✓
- Error context points to final (top-level) location
- May lose context of original break/continue location

**Mitigation:**
- Accept current behavior for Phase 2
- If detailed context needed, future enhancement could preserve original location
- Most users won't need deep stack trace for break/continue

**Decision:** No changes needed - current behavior is acceptable.

### Challenge 5: Performance with Deep Nesting

**Issue:** Very deep nesting (10+ levels) with high-level breaks might be slow.

**Analysis:**
- Each loop level: extractLevels() + ParseInt() + sprintf()
- For 10 levels: ~10 string operations total
- Negligible compared to DoBlock() evaluation overhead

**Mitigation:**
- Performance acceptable for reasonable nesting depths (< 10)
- If extreme nesting becomes common, optimize extractLevels() with caching
- Alternative: Use integer Args (requires verror changes)

**Decision:** Accept current design. Monitor user feedback for performance issues.

## Future Enhancements (Out of Scope)

### 1. Loop Depth Validation

**Feature:** Detect when `--levels N` exceeds actual loop depth.

**Implementation:**
- Add loop depth counter to evaluator state
- Increment on loop entry, decrement on loop exit
- Validate levels <= current depth in Break/Continue natives
- Error: "break --levels 5 exceeds actual loop depth (3)"

**Effort:** Medium (requires evaluator changes)

**Value:** Better error messages, prevents silent failures

### 2. Named Loop Labels

**Feature:** `break --loop outer` instead of `break --levels 2`.

**Syntax:**
```viro
outer: loop 3 [
    inner: loop 3 [
        break --loop 'outer  ; Exit outer loop by name
    ]
]
```

**Implementation:**
- Extend loop natives to accept optional label parameter
- Store labels in evaluator context with depth
- Resolve label to level count in Break/Continue
- Error if label not found

**Effort:** High (significant evaluator and native changes)

**Value:** More readable, less error-prone than counting levels

### 3. Loop Result Collection

**Feature:** Collect values from break/continue for debugging.

**Syntax:**
```viro
result: loop 10 [
    when (= some-condition true) [break --with-value x]
]
; result = x
```

**Implementation:**
- Extend Break/Continue to accept optional value
- Store value in Args[1] of signal
- Loop extracts and returns value on break

**Effort:** Medium (extends current Args protocol)

**Value:** Enables break with return value pattern

### 4. Optimized Level Encoding

**Feature:** Use integer Args instead of string conversion.

**Implementation:**
- Modify verror.Error.Args to support typed values (not just strings)
- Store levels as int64 directly
- Eliminate sprintf/ParseInt overhead

**Effort:** High (breaks verror API, requires migration)

**Value:** Minor performance gain, cleaner implementation

**Blocker:** Requires verror package refactoring

## Viro Guidelines Reference

### Coding Standards Followed

1. **No comments in code** - All documentation in this plan and package docs
2. **Constructor functions** - Use `value.NewNoneVal()`, `value.NewIntVal()`, `verror.NewError()`, etc.
3. **Error handling** - Use `verror.NewScriptError()` with category/ID/args
4. **Table-driven tests** - All tests follow `[]struct{name, input, expected, wantErr}` pattern
5. **TDD approach** - Write tests first (Step 2), implement to make them pass (Steps 3-7)
6. **Index-based refs** - Frame.Parent is int index, NOT pointer (N/A for this feature)

### Viro Naming Conventions

- **Native function names:** lowercase, hyphenated (`break`, `continue` - single words, no hyphen)
- **Refinements:** kebab-case with double-dash (`--levels`, `--with-index`)
- **Error IDs:** kebab-case (`break-outside-loop`, `invalid-operation`)
- **Category constants:** PascalCase (`ErrThrow`, `ErrScript`)
- **Query functions:** suffix with `?` (N/A for this feature)
- **Modification functions:** suffix with `!` (N/A for this feature)

### Architecture Alignment

- **Value system:** Returns `core.Value` from natives ✓
- **Type-based dispatch:** Not applicable (no type-specific behavior)
- **Error categories:** Uses existing ErrThrow (0) and ErrScript (300) ✓
- **Frame system:** No new frames, uses existing loop frame management ✓
- **Evaluator integration:** No evaluator changes required ✓
- **Minimal changes:** Extends existing natives, no new infrastructure ✓

### Error Message Guidelines

- **Clear and actionable:** "--levels must be >= 1" (tells user what to fix)
- **Type-specific:** "--levels requires integer" (explains expected type)
- **Contextual:** "break called outside of loop" (explains where it failed)
- **Consistent:** Follow existing error message patterns

## Summary

This plan implements multi-level loop control (`break --levels N`, `continue --levels N`) by extending the Phase 1 error-based signaling mechanism with a level count carried in the error Args[0] field.

**Key Design Decisions:**

1. **Args[0] Protocol:** Level count encoded as string ("1", "2", "3", ...)
   - Backward compatible (empty defaults to "1")
   - No new error infrastructure needed
   - Simple implementation

2. **Decremental Propagation:** Each loop decrements level and re-throws
   - Signal reaches correct depth automatically
   - No evaluator state tracking needed
   - Works with nested heterogeneous loops

3. **Refinement Validation:** Input validation at call site (Break/Continue natives)
   - levels >= 1 enforced
   - Type checking (must be integer)
   - Clear error messages

4. **Boundary Preservation:** Function boundaries and transparent blocks unchanged
   - Functions convert signals to errors (blocks multi-level)
   - Transparent blocks propagate unchanged (preserves multi-level)
   - No new special cases needed

5. **Backward Compatibility:** Default levels=1 matches Phase 1 exactly
   - All Phase 1 tests pass unchanged
   - No breaking changes to APIs
   - Smooth migration path

**Implementation Complexity:** Low
- 3 native function modifications (Break, Continue, extractLevels helper)
- 3 loop native modifications (Loop, While, Foreach)
- ~100 lines of code total
- No evaluator changes
- No error infrastructure changes

**Testing Strategy:**
- TDD approach (tests first)
- Comprehensive multi-level scenarios
- Error validation tests
- Backward compatibility verification
- Manual validation script

**Next Steps:**
1. Begin implementation at Step 1 (Add extractLevels helper)
2. Follow TDD: write tests (Step 2), then implement (Steps 3-7)
3. Validate with full test suite (Step 8)
4. Optional manual validation (Step 9)

**Success Criteria:**
- All multi-level tests pass
- All error validation tests pass
- All Phase 1 tests still pass (backward compatibility)
- No regressions in control flow or integration tests
- Clear error messages for invalid inputs
- Performance acceptable for reasonable nesting depths
