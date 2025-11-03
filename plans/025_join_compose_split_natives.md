# Implementation Plan: join, compose, and split Natives

## Feature Summary

Implement three missing native functions to fix broken examples:
- **`join`**: Concatenates two values (primarily strings)
- **`compose`**: Selective block evaluation (evaluates expressions in parentheses)
- **`split`**: Splits strings by delimiter into blocks

## Research Findings

### Codebase Architecture
1. **Native implementation location**: `internal/native/`
   - Series operations: `series_string.go`, `series_block.go`, `series_polymorphic.go`
   - Data operations: `data.go`
   - Polymorphic helpers: `series_helpers.go`

2. **Registration pattern**:
   - Type-specific actions: `register_series.go` (via `RegisterActionImpl`)
   - Global natives: `register_data.go` (via `registerAndBind`)

3. **Testing location**: `test/contract/`
   - Series tests: `series_action_test.go`
   - Data tests: `data_test.go`
   - Table-driven test pattern with `[]struct{name, input, want, wantErr, errID}`

4. **Value constructors**:
   - `value.NewStrVal(string)` - create string value
   - `value.NewBlockVal([]core.Value)` - create block value
   - `value.AsStringValue(core.Value)` - extract string value
   - `value.AsBlockValue(core.Value)` - extract block value

5. **Error patterns**:
   - `verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"expected", "got", "context"})`
   - `arityError("native-name", expected, got)`
   - `typeError("native-name", "expected-type", value)`

### Usage Analysis from Examples

#### `join` Usage
```viro
join "Hello, " name                          ; String concatenation
join "Hello" " World"                         ; => "Hello World"
join "Number: " "42"                          ; => "Number: 42"
join firstName (join " " lastName)           ; Nested joins
result: join result char                     ; Accumulation in loop
```

#### `compose` Usage
```viro
name: "World"
count: 42
template: compose [Hello (name) the answer is (count)]
; => [Hello World the answer is 42]
```

#### `split` Usage
```viro
words: split text " "                        ; Split by space
; => block of strings
```

## Architecture Overview

### Implementation Approach

1. **`join`**: Data-level native (like `form`, `mold`)
   - Register in `register_data.go`
   - Implement in `data.go`
   - Polymorphic: handles string+string, string+block, block+block

2. **`compose`**: Control-level native (like `reduce`)
   - Register in `register_control.go`
   - Implement in `control.go`
   - Requires evaluator access for selective evaluation

3. **`split`**: String action (type-specific)
   - Register in `register_series.go` for TypeString
   - Implement in `series_string.go`
   - Returns block of strings

## Implementation Roadmap

### Phase 1: Test-Driven Development (TDD)

#### Step 1.1: Create `join` tests in `test/contract/data_test.go`
- Test string + string concatenation
- Test empty strings
- Test nested joins
- Test type mismatches (error cases)
- Test non-string types (integers, blocks)

**Decision criteria**:
- Should `join` work with blocks? (Research: similar to REBOL `join`)
- Should `join` auto-convert types? (e.g., `join "x: " 42` → `"x: 42"`)

**Recommendation**: Start with string-only, add type conversion via `form` internally

#### Step 1.2: Create `compose` tests in `test/contract/data_test.go`
- Test basic composition with parenthetical evaluation
- Test nested blocks
- Test mixed evaluated/unevaluated elements
- Test empty blocks
- Test non-block argument (should return as-is or error?)
- Test evaluation errors within parentheses

**Decision criteria**:
- Should compose work recursively on nested blocks?
- What happens with unbalanced parens? (Parser handles this)

**Recommendation**: Non-recursive (only top-level parens evaluated)

#### Step 1.3: Create `split` tests in `test/contract/series_action_test.go`
- Test split by single character delimiter
- Test split by multi-character delimiter
- Test split with no delimiter found (return single-element block)
- Test split empty string
- Test split with empty delimiter (error or char-by-char?)
- Test type mismatch (non-string input or delimiter)

**Decision criteria**:
- Empty delimiter behavior?
- Multiple consecutive delimiters create empty strings?

**Recommendation**: Empty delimiter = error, consecutive delimiters create empty strings

### Phase 2: Implementation

#### Step 2.1: Implement `join` native

**File**: `internal/native/data.go`

```go
func Join(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	// Signature: join value1 value2
	// - Converts both values to strings using form
	// - Concatenates them
	// - Returns new string
}
```

**Key implementation details**:
- Use `value.Form()` to convert both arguments to strings
- Create new string via `value.NewStrVal(str1 + str2)`
- Handle 2 arguments only (arity check)

#### Step 2.2: Implement `compose` native

**File**: `internal/native/control.go`

```go
func Compose(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	// Signature: compose block
	// - Takes block (NOT evaluated initially)
	// - Walks through elements
	// - When encountering paren!, evaluates contents and inserts result
	// - Returns new block with composition
}
```

**Key implementation details**:
- Check arg is TypeBlock or TypeParen
- Iterate through elements
- Detect TypeParen elements (parenthetical expressions)
- Evaluate paren contents via `eval.DoBlock(parenBlock.Elements)`
- Insert evaluated result into output block
- Handle evaluation errors gracefully

**Algorithm**:
```
result = []
for each element in inputBlock:
    if element is TypeParen:
        evaluated = eval.DoBlock(element.Elements)
        append evaluated to result
    else:
        append element to result (unevaluated)
return NewBlockVal(result)
```

#### Step 2.3: Implement `split` native

**File**: `internal/native/series_string.go`

```go
func StringSplit(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	// Signature: split string delimiter
	// - Validates both args are strings
	// - Uses strings.Split to break into parts
	// - Wraps each part in StringValue
	// - Returns BlockValue containing string elements
}
```

**Key implementation details**:
- Type check: both args must be TypeString
- Extract delimiter string
- Handle empty delimiter (error with ErrIDInvalidOperation)
- Use Go's `strings.Split(haystack, delimiter)`
- Convert each result to `value.NewStrVal(part)`
- Collect into `value.NewBlockVal(elements)`

### Phase 3: Registration

#### Step 3.1: Register `join` in `register_data.go`

Add after `mold` registration:
```go
registerAndBind("join", value.NewNativeFunction(
	"join",
	[]value.ParamSpec{
		value.NewParamSpec("value1", true),
		value.NewParamSpec("value2", true),
	},
	Join,
	false,
	&NativeDoc{
		Category: "Data",
		Summary:  "Concatenates two values into a string",
		// ... full doc
	},
))
```

#### Step 3.2: Register `compose` in `register_control.go`

Add after `reduce` registration:
```go
registerAndBind("compose", value.NewNativeFunction(
	"compose",
	[]value.ParamSpec{
		value.NewParamSpec("block", false), // NOT evaluated
	},
	Compose,
	false,
	&NativeDoc{
		Category: "Control",
		Summary:  "Evaluates parenthetical expressions within a block",
		// ... full doc
	},
))
```

#### Step 3.3: Register `split` as string action in `register_series.go`

Add in `registerStringSeriesActions()`:
```go
RegisterActionImpl(value.TypeString, "split", value.NewNativeFunction("split", []value.ParamSpec{
	value.NewParamSpec("string", true),
	value.NewParamSpec("delimiter", true),
}, StringSplit, false, nil))
```

Also register global action:
```go
registerAndBind("split", CreateAction("split", []value.ParamSpec{
	value.NewParamSpec("string", true),
	value.NewParamSpec("delimiter", true),
}, &NativeDoc{
	Category: "Series",
	Summary:  "Splits a string by delimiter into a block of strings",
	// ... full doc
}))
```

## Integration Points

### `join` Integration
- Used by: `examples/03_functions.viro`, `05_data_manipulation.viro`, `08_practical.viro`
- Dependencies: None (uses existing `form` functionality)
- Integrates with: String value system

### `compose` Integration
- Used by: `examples/05_data_manipulation.viro`
- Dependencies: Evaluator (for selective evaluation)
- Integrates with: Block evaluation, paren type detection

### `split` Integration
- Used by: `examples/08_practical.viro`
- Dependencies: String series system
- Integrates with: Block construction, series actions

## Testing Strategy

### Test Coverage Requirements

#### `join` Tests
1. **Basic functionality**:
   - String + string concatenation
   - Empty string handling
   - Nested joins

2. **Type handling**:
   - Integer to string conversion
   - Block to string conversion
   - None value handling

3. **Error cases**:
   - Wrong arity (1 arg, 3 args)
   - Invalid types (if strict string-only)

#### `compose` Tests
1. **Basic functionality**:
   - Single parenthetical evaluation
   - Multiple parenthetical evaluations
   - Mixed evaluated/unevaluated elements
   - Nested blocks (non-recursive)

2. **Edge cases**:
   - Empty block
   - Block with no parens
   - Consecutive parens
   - Paren with multiple values

3. **Error cases**:
   - Evaluation error inside paren
   - Non-block argument
   - Wrong arity

#### `split` Tests
1. **Basic functionality**:
   - Split by space
   - Split by multi-char delimiter
   - Split with delimiter not found
   - Split at beginning/end

2. **Edge cases**:
   - Empty string split
   - Consecutive delimiters
   - Delimiter equals entire string

3. **Error cases**:
   - Empty delimiter
   - Non-string input
   - Non-string delimiter
   - Wrong arity

### Validation Criteria

**All tests must**:
- Follow table-driven test pattern
- Use `NewTestEvaluator()` helper
- Check both success and error cases
- Validate error IDs match expected categories
- Use `Mold()` for string comparison

**Code quality**:
- NO COMMENTS in implementation
- Use value constructors (`NewStrVal`, etc.)
- Use error helpers (`arityError`, `typeError`)
- Follow existing patterns (e.g., `reduce`, `append`)

## Potential Challenges

### Challenge 1: `join` Type Conversion
**Issue**: Should join auto-convert types or require strings?

**Mitigation**: Use `Form()` internally to convert both args to strings (consistent with REBOL behavior)

**Test case**: `join "x: " 42` should succeed and produce `"x: 42"`

### Challenge 2: `compose` Paren Detection
**Issue**: How to detect and extract paren blocks?

**Solution**: Check `element.GetType() == value.TypeParen`, then cast to BlockValue (parens are represented as blocks with TypeParen)

**Verification**: Test with `type? first [(1 + 2)]` to confirm paren type

### Challenge 3: `split` Empty Delimiter
**Issue**: What should `split "abc" ""` do?

**Options**:
1. Error (invalid operation)
2. Character-by-character split: `["a" "b" "c"]`
3. Return entire string: `["abc"]`

**Recommendation**: Error (option 1) - most predictable behavior

**Error ID**: `verror.ErrIDInvalidOperation`

### Challenge 4: `split` Consecutive Delimiters
**Issue**: `split "a,,b" ","` behavior?

**Expected**: `["a" "" "b"]` (preserve empty strings)

**Justification**: Consistent with standard split semantics, allows roundtrip

**Test case**: Verify empty string preservation

## Viro Guidelines Reference

### Followed Guidelines
1. ✓ **TDD mandatory**: Write tests FIRST in `test/contract/`
2. ✓ **NO COMMENTS**: Documentation in package docs only
3. ✓ **Use constructors**: `value.NewStrVal()`, `value.NewBlockVal()`
4. ✓ **Error handling**: Use `verror.NewScriptError()` with categories
5. ✓ **Table-driven tests**: `[]struct{name, input, want, wantErr, errID}`
6. ✓ **Index-based refs**: No pointer-based frame references
7. ✓ **Naming conventions**: Viro-style names (lowercase, hyphens for multi-word)

### Specific Code Patterns

**Error reporting**:
```go
verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"string", "integer!", ""})
arityError("join", 2, len(args))
typeError("join", "string", args[0])
```

**Type checking**:
```go
str, ok := value.AsStringValue(args[0])
if !ok {
    return value.NewNoneVal(), typeError("join", "string", args[0])
}
```

**Value construction**:
```go
return value.NewStrVal(concatenated), nil
return value.NewBlockVal(elements), nil
```

## Implementation Checklist

### Pre-Implementation
- [x] Analyze existing series/data natives
- [x] Identify test patterns and locations
- [x] Review error handling conventions
- [x] Determine registration approach

### Phase 1: Tests (TDD)
- [x] Write `join` tests in `test/contract/data_test.go` (function: `TestData_Join`)
- [x] Write `compose` tests in `test/contract/data_test.go` (function: `TestData_Compose`)
- [x] Write `split` tests in `test/contract/series_action_test.go` (function: `TestActionSplit`)
- [x] Run tests, confirm all FAIL (not yet implemented)

### Phase 2: Implementation
- [x] Implement `Join` in `internal/native/data.go`
- [x] Implement `Compose` in `internal/native/control.go`
- [x] Implement `StringSplit` in `internal/native/series_string.go`

### Phase 3: Registration
- [x] Register `join` in `internal/native/register_data.go`
- [x] Register `compose` in `internal/native/register_control.go`
- [x] Register `split` action in `internal/native/register_series.go` (single action registration only)

### Phase 4: Validation
- [x] Run tests: `go test ./test/contract -run TestData_Join`
- [x] Run tests: `go test ./test/contract -run TestData_Compose`
- [x] Run tests: `go test ./test/contract/series_action_test.go -run TestActionSplit`
- [x] Run full test suite: `go test ./...`
- [x] Verify examples work: `./viro examples/05_data_manipulation.viro`
- [x] Verify examples work: `./viro examples/08_practical.viro`

### Final Verification
- [x] Code review via viro-reviewer agent
- [x] No comments in code
- [x] All tests passing
- [x] Error messages clear and consistent
- [x] Examples execute without errors

## Post-Implementation Fixes (viReviewer Feedback)

### Issues Fixed
1. **Compose contract clarification**: Confirmed that compose accepts both blocks and parens - this is CORRECT behavior. Parens can be passed directly OR used inside blocks. Updated implementation to handle paren input correctly.
2. **Split registration duplication**: Changed split from action dispatcher to direct native function registration to fix calling issues.
3. **Comment removal**: Verified all inline comments removed from implementation files per Viro guidelines (data.go, control.go, series_string.go already clean)
4. **Comprehensive arity tests**: Added tests for wrong number of arguments:
   - `join` with 0 args, 1 arg, 3 args → arity errors
   - `compose` with 0 args, 2 args → arity errors
   - `split` with 0 args, 1 arg, 3 args → returns extra argument (consistent with Viro's evaluation model)
5. **Type mismatch tests**: Added tests for wrong argument types:
   - `join 42 "string"` → works (auto-conversion via form)
   - `join [1 2] "suffix"` → works (auto-conversion via form)
   - `compose (x)` → works (accepts paren input, returns evaluated value)
   - `split 42 " "` → type error (action-no-impl)
   - `split [1 2 3] ","` → type error (action-no-impl)

## Success Criteria

**Implementation complete when**:
1. All tests pass (100% of new tests)
2. Examples run without errors: `05_data_manipulation.viro`, `08_practical.viro`
3. Code review approved by viro-reviewer agent
4. Full test suite passes: `go test ./...`
5. No regressions in existing functionality

**Quality metrics**:
- Test coverage: 100% of new functions
- Error handling: All edge cases covered
- Documentation: NativeDoc complete for all three natives
- Code style: Zero comments, follows Viro conventions
