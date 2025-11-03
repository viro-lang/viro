# Implementation Plan: Binary Literal Support

## Feature Summary

Add comprehensive binary literal support to the Viro language, enabling hexadecimal notation for raw byte sequences. This feature will allow developers to write binary data directly in source code using the syntax `#{DEADBEEF}` with optional spacing for readability (`#{DE AD BE EF}`).

**Key deliverables:**
- Parser support for binary literal syntax `#{...}` with hexadecimal content
- Proper hexadecimal display formatting in `mold` and `form` operations
- Full round-trip capability (parse → mold → parse preserves value)
- Enable existing skipped test for binary series operations

---

## Research Findings

### Current State Analysis

**BinaryValue Implementation** (`internal/value/binary.go`):
- Complete type exists with full series operations (append, first, last, skip, back, etc.)
- Constructor: `NewBinaryValue(data []byte) *BinaryValue` (internal)
- Public constructor: `NewBinaryVal(data []byte) core.Value` (primitives.go:339)
- **CRITICAL BUG**: Current `Mold()` (lines 31-36) does `"#{" + string(b.data) + "}"` which incorrectly treats raw bytes as ASCII characters
- `Form()` (lines 38-40) just delegates to `Mold()` - no truncation logic

**Type System** (`internal/value/types.go`):
- `TypeBinary` defined as value type (line 33)
- Fully integrated: `TypeToString()` returns "binary!" (lines 75-76)
- Recognized as series type in `IsSeries()` (line 88)
- No changes needed to type system

**Parser Status** (`grammar/viro.peg`):
- No binary literal rule exists
- `HexDigit` rule defined (line 106: `[0-9A-Fa-f]`) but currently unused
- `Value` rule (lines 80-84) needs Binary option added
- Grammar is PEG-based, generates `internal/parse/peg/parser.go` via pigeon tool

**Blocked Tests** (`test/contract/series_test.go`):
- Test "back binary skip" (lines 1611-1619) creates binary series with multiple append operations
- Currently skipped with message: "Binary literals not implemented in parser yet" (lines 1624-1627)
- Test validates series navigation (next, back) on binary values
- Once parser works, this test should pass without modification

**Usage Context**:
- Binary data already used in port operations (read with `--binary` flag)
- File I/O can read/write binary data programmatically
- Missing: ability to write binary literals directly in source code

### Comparable Implementations

**String literals** (grammar/viro.peg:100-104):
```
String <- '"' [^"]* '"' {
    text := string(c.text)
    content := text[1:len(text)-1]
    return value.NewStrVal(content), nil
}
```
Pattern: delimiter parsing, content extraction, constructor call

**Integer literals** (grammar/viro.peg:86-88):
```
Integer <- '-'? [0-9]+ !DecimalStart {
    return value.NewIntVal(toInt(string(c.text))), nil
}
```
Pattern: regex match, helper function conversion, constructor call

**Helper functions** (grammar/viro.peg:13-34):
- Defined in preamble section between `{...}` brackets
- Access to `string(c.text)` for matched content
- Can import stdlib packages (strconv, strings, fmt already imported)

---

## Architecture Overview

### Three-Phase Implementation

**Phase 1: Fix BinaryValue Display** (testable independently)
- Update `Mold()` to generate proper hexadecimal output
- Update `Form()` to truncate long binaries for human readability
- Can test with programmatically created binaries before parser works

**Phase 2: Add Parser Support** (depends on Phase 1 format)
- Add hex parsing helper function to grammar preamble
- Create Binary literal grammar rule
- Update Value rule to include Binary option
- Regenerate parser with `make grammar`

**Phase 3: Integration Testing** (depends on Phases 1 & 2)
- Remove skip from existing series test
- Add comprehensive parser tests
- Add round-trip tests (parse → mold → parse)
- Validate full feature integration

---

## Implementation Roadmap

### Phase 1: BinaryValue Display Methods

**File**: `internal/value/binary.go`

**Task 1.1: Update Mold() Method**

Current broken implementation (lines 31-36):
```go
func (b *BinaryValue) Mold() string {
    if len(b.data) == 0 {
        return "#{}"
    }
    return "#{" + string(b.data) + "}"  // BUG: treats bytes as ASCII
}
```

**Implementation approach:**
1. Handle empty binary: return `"#{}"`
2. Convert each byte to 2-character uppercase hex string
3. Insert space after every 2 hex characters (every byte) for readability
4. Format: `#{DE AD BE EF}` matching requirement examples

**Decision points for coder agent:**
- **Hex encoding approach**: Use `fmt.Sprintf("%02X", byte)` for per-byte control OR `encoding/hex.EncodeToString()` for full string then add spaces
- **Space insertion**: Build string incrementally with spaces OR encode full hex then insert spaces every 2 chars
- **Recommendation**: Incremental approach for better control - loop bytes, format each, add space between

**Expected output examples:**
- `[]byte{}` → `"#{}"`
- `[]byte{0xFF}` → `"#{FF}"`
- `[]byte{0xDE, 0xAD}` → `"#{DE AD}"`
- `[]byte{0xDE, 0xAD, 0xBE, 0xEF}` → `"#{DE AD BE EF}"`

**Imports needed:**
- Add `encoding/hex` OR rely on `fmt` (already imported line 5)
- Prefer `fmt.Sprintf("%02X", byte)` to avoid additional import

**Task 1.2: Update Form() Method**

Current implementation (lines 38-40):
```go
func (b *BinaryValue) Form() string {
    return b.Mold()
}
```

**Implementation approach:**
1. For short binaries (≤ 64 bytes), use same output as Mold()
2. For long binaries, truncate display with ellipsis and byte count
3. Format: `#{DE AD BE EF ... (256 bytes)}`

**Decision points:**
- **Truncation threshold**: 64 bytes recommended (128 hex chars + 63 spaces ≈ 191 display chars)
- **Truncation format**: Show first N bytes then "... (total bytes)"
- **How many bytes to show**: 8-16 bytes gives good preview without overwhelming

**Pattern from other types:**
- StringValue: Mold adds quotes, Form doesn't (string.go:27-32)
- BlockValue: Both show full content (block.go:44-64)  
- DecimalValue: Both identical (decimal.go:44-58)

**Recommendation**: Binary more like string (Mold for code, Form for display), so differentiate them

**Expected output examples:**
- 64 bytes or less: same as Mold()
- 128 bytes: `#{DE AD BE EF CA FE BA BE ... (128 bytes)}`

**Task 1.3: Create Tests**

**File**: `test/contract/binary_test.go` (new file) OR add to `data_test.go`

**Test structure** (table-driven pattern):
```go
tests := []struct{
    name string
    data []byte
    wantMold string
    wantForm string
}{
    {name: "empty", data: []byte{}, wantMold: "#{}", wantForm: "#{}"},
    {name: "single byte", data: []byte{0xFF}, wantMold: "#{FF}", wantForm: "#{FF}"},
    // ... more cases
}
```

**Test cases needed:**
- Empty binary
- Single byte (various values: 0x00, 0xFF, 0x42)
- Multiple bytes (2, 4, 8, 16 bytes)
- Exactly 64 bytes (boundary case)
- 65 bytes (just over threshold)
- 128+ bytes (verify truncation)
- All zeros, all 0xFF (edge patterns)

**Validation checkpoints:**
- [ ] All Mold() tests pass
- [ ] All Form() tests pass
- [ ] Truncation works correctly
- [ ] No buffer overflows or panics
- [ ] Output format matches requirements exactly

---

### Phase 2: Parser Support

**File**: `grammar/viro.peg`

**Task 2.1: Add Hex Parsing Helper**

Add to preamble section (after existing helpers, around line 60):

**Function signature:**
```go
func parseHex(hexStr string) ([]byte, error)
```

**Implementation approach:**
1. Strip all whitespace from input (spaces, tabs)
2. Validate even number of hex digits (must be pairs)
3. Validate all characters are valid hex (0-9, A-F, a-f)
4. Parse pairs into bytes
5. Return byte slice or error

**Decision points:**
- **Whitespace stripping**: Use `strings.ReplaceAll()` multiple times OR `strings.Map()` OR regex
- **Recommendation**: `strings.Map()` with unicode.IsSpace for clean approach
- **Hex parsing**: Manual parsing (iterate pairs) OR `encoding/hex.DecodeString()`
- **Recommendation**: `encoding/hex.DecodeString()` after whitespace removal - simpler, stdlib-tested
- **Error handling**: Return descriptive errors that PEG can use

**Error cases:**
- Odd number hex digits: "binary literal requires even number of hex digits"
- Invalid characters: "invalid hex character in binary literal"
- Return nil byte slice on error so PEG parser can fail cleanly

**Pseudo-code:**
```go
func parseHex(hexStr string) ([]byte, error) {
    // Remove all whitespace
    cleaned := strings.Map(func(r rune) rune {
        if unicode.IsSpace(r) { return -1 }
        return r
    }, hexStr)
    
    // Handle empty
    if cleaned == "" { return []byte{}, nil }
    
    // Validate even length
    if len(cleaned) % 2 != 0 {
        return nil, fmt.Errorf("binary literal requires even number of hex digits")
    }
    
    // Decode hex
    bytes, err := hex.DecodeString(cleaned)
    if err != nil {
        return nil, fmt.Errorf("invalid hex character in binary literal")
    }
    
    return bytes, nil
}
```

**Imports needed in preamble:**
- Add `encoding/hex` to existing imports (line 8 area)
- `unicode` may be needed for IsSpace (or use simpler approach)

**Task 2.2: Add Binary Grammar Rule**

Add rule after String/Decimal/Integer rules (around line 105):

**Grammar rule:**
```peg
Binary <- '#' '{' content:HexContent '}' {
    hexStr := string(content.([]byte))
    bytes, err := parseHex(hexStr)
    if err != nil {
        return nil, err
    }
    return value.NewBinaryVal(bytes), nil
}

HexContent <- [0-9A-Fa-f \t\n\r]* {
    return c.text, nil
}
```

**Decision points:**
- **Content capture**: Capture everything between braces, let parseHex validate
- **Whitespace in pattern**: Allow space, tab, newline, carriage return in content
- **Alternative approach**: Strictly parse hex pairs with whitespace between - MORE COMPLEX
- **Recommendation**: Capture all, validate in helper - simpler grammar, better errors

**Placement considerations:**
- Must come BEFORE Word rule to avoid `#` being part of word
- Actually, `#` not in WordChars (line 246: `[\p{L}_?+*/<>=-]`), so order less critical
- Recommend: place with other literal types (Integer, Decimal, String area)

**Task 2.3: Update Value Rule**

Current Value rule (lines 80-84):
```peg
Value <- _ val:(Decimal / Integer / String / Block / Paren /
                Path / SetPath / SetWord / GetPath / GetWord / LitWord /
                Datatype / Word) _ {
    return val, nil
}
```

**Change needed:**
Add `Binary` to the ordered choice, position before `Word`:

```peg
Value <- _ val:(Decimal / Integer / Binary / String / Block / Paren /
                Path / SetPath / SetWord / GetPath / GetWord / LitWord /
                Datatype / Word) _ {
    return val, nil
}
```

**Ordering rationale:**
- Binary starts with `#`, unique prefix
- Place with other literals (Integer, Decimal, String) for logical grouping
- Before complex types (Path, Word) following existing pattern

**Task 2.4: Generate Parser**

**Command**: `make grammar`

This runs: `pigeon -o internal/parse/peg/parser.go grammar/viro.peg`

**Validation:**
- Check for parser generation errors
- Verify `internal/parse/peg/parser.go` updated (timestamp)
- DO NOT manually edit parser.go - all changes via grammar

**If parser generation fails:**
- Review grammar syntax (PEG rules are strict)
- Check helper function compiles (Go code in actions)
- Common issues: unmatched braces, missing return in action code

**Task 2.5: Parser Tests**

**File**: `internal/parse/parse_test.go` OR `test/integration/parse_binary_test.go`

**Test categories:**

**Success cases:**
```go
{name: "empty binary", input: `#{}`, want: value.NewBinaryVal([]byte{})},
{name: "single byte uppercase", input: `#{FF}`, want: value.NewBinaryVal([]byte{0xFF})},
{name: "single byte lowercase", input: `#{ff}`, want: value.NewBinaryVal([]byte{0xFF})},
{name: "multiple bytes", input: `#{DEADBEEF}`, want: value.NewBinaryVal([]byte{0xDE, 0xAD, 0xBE, 0xEF})},
{name: "with spaces", input: `#{DE AD BE EF}`, want: value.NewBinaryVal([]byte{0xDE, 0xAD, 0xBE, 0xEF})},
{name: "mixed case", input: `#{DeAdBeEf}`, want: value.NewBinaryVal([]byte{0xDE, 0xAD, 0xBE, 0xEF})},
{name: "with newlines", input: "#{DE AD\nBE EF}", want: value.NewBinaryVal([]byte{0xDE, 0xAD, 0xBE, 0xEF})},
{name: "many bytes", input: `#{00010203}`, want: value.NewBinaryVal([]byte{0x00, 0x01, 0x02, 0x03})},
```

**Error cases:**
```go
{name: "odd digits", input: `#{FFF}`, wantErr: true},
{name: "invalid char", input: `#{GGGG}`, wantErr: true},
{name: "unclosed", input: `#{FF`, wantErr: true},
{name: "no opening brace", input: `#FF}`, wantErr: true},
```

**Integration with evaluation:**
```go
{name: "binary in expression", input: `length? #{DEADBEEF}`, want: value.NewIntVal(4)},
{name: "binary in block", input: `[#{FF} #{00}]`, want: value.NewBlockVal([...])}
```

**Validation checkpoints:**
- [ ] All valid hex patterns parse correctly
- [ ] Case insensitivity works (a-f and A-F both accepted)
- [ ] Whitespace handling (spaces, tabs, newlines)
- [ ] Empty binary parses
- [ ] Error cases fail with descriptive messages
- [ ] Binary integrates with other value types

---

### Phase 3: Integration Testing

**Task 3.1: Enable Existing Test**

**File**: `test/contract/series_test.go`

**Change** (lines 1624-1627):
Remove the skip check:
```go
// DELETE THESE LINES:
if strings.Contains(tt.input, "#{") || strings.Contains(tt.input, "append #{}") {
    t.Skip("Binary literals not implemented in parser yet - cannot construct binary series for testing")
    return
}
```

**Expected result:**
Test "back binary skip" should now pass without modification.

**Test validates:**
- Binary literal parsing: `#{}`
- Append operations on binary: `append bin 1`
- Series navigation: `next next bin`, `back moved`
- Value extraction: `first backData`

**Task 3.2: Round-Trip Tests**

**File**: `test/contract/binary_test.go` or `test/integration/binary_roundtrip_test.go`

**Round-trip pattern:**
```go
tests := []struct{
    name string
    input string  // binary literal
    want []byte   // expected byte content
}{
    {name: "empty", input: "#{}", want: []byte{}},
    {name: "single", input: "#{42}", want: []byte{0x42}},
    {name: "deadbeef", input: "#{DEADBEEF}", want: []byte{0xDE, 0xAD, 0xBE, 0xEF}},
}

for _, tt := range tests {
    // Parse literal
    parsed := parse(tt.input)
    binary := parsed.(*value.BinaryValue)
    
    // Check bytes
    assertEqual(binary.Bytes(), tt.want)
    
    // Mold it
    molded := binary.Mold()
    
    // Parse molded output
    reparsed := parse(molded)
    rebinary := reparsed.(*value.BinaryValue)
    
    // Should equal original
    assertEqual(rebinary.Bytes(), tt.want)
}
```

**Validates:**
- Parser produces correct byte values
- Mold produces parseable output
- Round-trip preserves data exactly
- No data corruption through format conversion

**Task 3.3: Comprehensive Feature Tests**

**File**: `test/contract/binary_test.go`

**Feature coverage:**

**Literal parsing:**
- All valid hex patterns (0-9, A-F, a-f)
- Various byte sequences (empty, single, multiple, many)
- Whitespace variations (none, spaces, tabs, newlines, mixed)

**Display formatting:**
- Mold output format (hex with spaces)
- Form output format (truncation for large binaries)
- Edge cases (empty, single byte, boundary sizes)

**Series operations:**
- Create binary with literals
- Append, insert, remove operations
- Navigation: first, last, next, back, skip
- Copy, take operations
- Integration with series natives

**Type integration:**
- Type checking: `type? #{FF}` → `binary!`
- Equality: `#{FF} = #{FF}` → `true`
- Inequality: `#{FF} = #{00}` → `false`
- In blocks: `[1 #{FF} "text"]`
- In function calls: `length? #{DEADBEEF}`

**I/O integration:**
- Write binary to file
- Read binary from file
- Port operations with binary data

**Validation checkpoints:**
- [ ] All parser tests pass
- [ ] Round-trip tests pass  
- [ ] Series operations work correctly
- [ ] Type system integration complete
- [ ] I/O operations handle binaries
- [ ] No regressions in existing tests

---

## Integration Points

### Parser to Value System

**Flow**: Grammar rule → parseHex() → value.NewBinaryVal() → BinaryValue

**Connection**:
- Grammar captures hex string content
- parseHex() converts to []byte
- NewBinaryVal() wraps in BinaryValue
- Returns as core.Value to evaluator

**Validation**: Parser tests verify this chain works correctly.

### Value Display to Parser

**Flow**: BinaryValue → Mold() → hex string → parseable literal

**Connection**:
- Mold() formats bytes as `#{XX XX XX}`
- Output must be valid grammar for Binary rule
- Round-trip tests verify parse(mold(x)) == x

**Validation**: Round-trip tests ensure compatibility.

### Series Operations

**Already Complete**: BinaryValue implements full Series interface

**Operations supported:**
- Length: `length? bin`
- Access: `first bin`, `last bin`
- Navigation: `next bin`, `back bin`, `skip bin 2`
- Modification: `append bin 255`, `insert bin 0`
- Slicing: `copy/part bin 4`, `take bin 2`

**No changes needed** - just enable with literals.

### Type System

**Already Complete**: TypeBinary fully integrated

**Type operations:**
- Type check: `type? #{FF}` returns "binary!"
- Type comparison: `binary! = type? #{}`
- Series check: `series? #{FF}` returns true

**No changes needed** - just works with literals.

---

## Testing Strategy

### Test Organization

**Unit Tests**:
- `test/contract/binary_test.go` - Binary-specific functionality
- Tests for Mold(), Form(), parsing, round-trips

**Integration Tests**:
- `test/integration/binary_integration_test.go` - Feature integration
- Tests combining binaries with other features
- I/O operations, series operations, type checking

**Parser Tests**:
- `internal/parse/parse_test.go` - Parser-level tests
- Or `test/integration/parse_binary_test.go` for higher-level

**Existing Tests**:
- `test/contract/series_test.go` - Enable skipped test
- Validates series operations work with binaries

### Test Coverage Requirements

**Must test:**
- ✓ Empty binary: `#{}`
- ✓ Single byte: various values
- ✓ Multiple bytes: various patterns
- ✓ Large binaries: 64+ bytes
- ✓ All hex digits: 0-9, A-F, a-f
- ✓ Whitespace: spaces, tabs, newlines
- ✓ Case insensitivity: `#{ff}` == `#{FF}`
- ✓ Mold format: uppercase hex with spaces
- ✓ Form truncation: large binaries
- ✓ Round-trip: parse → mold → parse
- ✓ Error cases: odd digits, invalid chars, unclosed
- ✓ Series operations: all series natives
- ✓ Type integration: type checking, equality
- ✓ I/O integration: file read/write

### Test Execution Order

**Phase 1 tests** (run first):
```bash
go test -v ./test/contract -run TestBinaryMold
go test -v ./test/contract -run TestBinaryForm
```

**Phase 2 tests** (after parser):
```bash
go test -v ./internal/parse -run TestParseBinary
go test -v ./test/integration -run TestBinaryParsing
```

**Phase 3 tests** (full integration):
```bash
go test -v ./test/contract -run TestBackBinarySkip  # Previously skipped
go test -v ./test/contract -run TestBinary          # All binary tests
go test -v ./test/integration -run TestBinary       # Integration tests
```

**Full test suite**:
```bash
make test                    # All tests
make test-summary           # Summary view
go test -json ./... | jq    # Structured output for analysis
```

---

## Potential Challenges & Mitigations

### Challenge 1: Parser Rule Ambiguity

**Issue**: `#` character could potentially conflict with future syntax extensions

**Mitigation**:
- `#` not in WordChars set ([\p{L}_?+*/<>=-]), so no current conflict
- Binary rule requires `#{` prefix (two characters), very specific
- Place Binary rule before Word rule in ordered choice for precedence
- Well-defined grammar rules prevent ambiguity

**Validation**: Parser tests with binaries adjacent to words

### Challenge 2: Hex Encoding Performance

**Issue**: Converting large binaries to hex strings could be slow

**Analysis**:
- Mold() called for display, not frequently in tight loops
- Form() can truncate for large binaries, limiting work
- encoding/hex.EncodeToString() is optimized stdlib code

**Mitigation**:
- Use stdlib hex encoding (fast, tested)
- Form() truncates at 64 bytes (limits display work)
- Accept that molding very large binaries may be slower
- Could add benchmarks if needed: `go test -bench=BinaryMold`

**When to optimize**: Only if profiling shows actual bottleneck

### Challenge 3: Memory for Large Binaries

**Issue**: Displaying large binary as hex string uses 2x+ memory

**Analysis**:
- String requires 2 chars per byte + spaces = ~3x data size
- 1MB binary → ~3MB string for mold
- Temporary allocation, GC will collect

**Mitigation**:
- Form() truncates to reduce memory for display
- Mold() only used when full representation needed (save, debugging)
- Not storing molded string, just generating on demand
- If concern grows, could add streaming hex encoder

**Threshold**: 1MB binary → 3MB string is acceptable for modern systems

### Challenge 4: Grammar Generation Errors

**Issue**: Syntax errors in PEG grammar can be cryptic

**Prevention**:
- Test helper function separately before adding to grammar
- Use simple pattern for Binary rule (proven pattern from String)
- Validate helper compiles: write in separate .go file first
- Review existing rules for pattern reference

**Recovery**:
- If `make grammar` fails, read pigeon error message carefully
- Check for: unmatched braces, missing semicolons, invalid Go in actions
- Test helper function in isolation
- Compare to working rules (String, Integer patterns)

**Validation**: Build grammar incrementally, test each addition

### Challenge 5: Round-Trip Format Consistency

**Issue**: Mold output must be parseable by grammar

**Requirement**: parse(mold(x)) must equal x

**Approach**:
- Mold outputs `#{XX XX XX}` format (uppercase hex, spaces)
- Grammar accepts hex (any case) with optional whitespace
- Grammar is MORE permissive than Mold output (good!)
- Mold output is ONE valid format, not the only format

**Validation**:
- Round-trip tests verify parse(mold(x)) == x
- Parser tests verify multiple formats accepted
- Consistency: Mold always produces same format for same bytes

**Risk mitigation**: Write round-trip tests FIRST in Phase 3

---

## Viro Guidelines Reference

### Code Style Compliance

**NO COMMENTS in code** (AGENTS.md):
- All implementation code must have zero inline comments
- Documentation belongs in package docs or markdown files
- This plan serves as documentation; code should be self-explanatory

**Constructor usage** (AGENTS.md):
- Always use `value.NewBinaryVal()` for public API
- Use `value.NewBinaryValue()` for internal construction  
- Never create struct literals: `&BinaryValue{...}` is forbidden

**Error handling** (AGENTS.md):
- Use `verror.NewScriptError()` for script-level errors
- Use `verror.NewMathError()` for math errors (not applicable here)
- Parser errors can use fmt.Errorf() for PEG error reporting

**Import organization** (AGENTS.md):
- Group: stdlib → external → internal
- Alphabetical within groups
- For binary.go: `encoding/hex`, `fmt` (stdlib) → `github.com/.../core` (internal)

**Naming conventions** (AGENTS.md):
- Use Viro-style names: `length?`, `type?` (with ? suffix for predicates)
- Function names follow Go conventions: `parseHex`, `Mold`, `Form`
- No Hungarian notation or type prefixes

### Testing Requirements

**TDD mandatory** (AGENTS.md):
- Write tests FIRST before implementation
- Every code change MUST have test coverage
- Prefer automated tests over manual execution

**Table-driven tests** (AGENTS.md):
- Always use `tests := []struct{name, args, want, wantErr}`
- Clear test names describing scenario
- Comprehensive coverage of success and error paths

**Test execution** (AGENTS.md):
- Use `make test` for full suite
- Use `go test -v ./path -run TestName` for specific tests
- Use `make test-summary` for summary view

### Development Workflow

**Use viro-coder agent** (AGENTS.md):
- When editing ANY Viro interpreter code, MUST use viro-coder agent
- Never edit code directly in main conversation
- Agent has specialized Viro architecture expertise

**Code review process** (AGENTS.md):
- After viro-coder finishes, MUST use viro-reviewer agent
- Main agent decides whether to apply suggestions
- If changes applied, use viro-coder again (never edit directly)

**Build commands** (AGENTS.md):
- Generate grammar: `make grammar` or `pigeon -o internal/parse/peg/parser.go grammar/viro.peg`
- Build: `make build` (includes grammar generation)
- Test: `make test` or `go test ./...`

---

## Implementation Checklist

### Phase 1: BinaryValue Display

- [ ] Add hex encoding to binary.go (imports if needed)
- [ ] Implement BinaryValue.Mold() with proper hex formatting
- [ ] Implement BinaryValue.Form() with truncation logic
- [ ] Create test file: test/contract/binary_test.go
- [ ] Write Mold() test cases (empty, single, multiple, many bytes)
- [ ] Write Form() test cases (short, long, boundary cases)
- [ ] Run tests: `go test -v ./test/contract -run TestBinary`
- [ ] Verify all Mold/Form tests pass
- [ ] Manual verification: create binary in REPL, check display

### Phase 2: Parser Support

- [ ] Add encoding/hex import to grammar preamble
- [ ] Implement parseHex() helper function in grammar preamble
- [ ] Add HexContent grammar rule
- [ ] Add Binary grammar rule (using parseHex helper)
- [ ] Update Value rule to include Binary option
- [ ] Run `make grammar` to regenerate parser
- [ ] Check for grammar generation errors
- [ ] Verify parser.go timestamp updated
- [ ] Write parser test cases (valid hex patterns)
- [ ] Write parser error test cases (invalid patterns)
- [ ] Run parser tests: `go test -v ./internal/parse -run TestParseBinary`
- [ ] Verify all parser tests pass
- [ ] Manual verification: `./viro -c "#{DEADBEEF}"`

### Phase 3: Integration Testing

- [ ] Remove skip from series_test.go (lines 1624-1627)
- [ ] Run: `go test -v ./test/contract -run TestBackBinarySkip`
- [ ] Verify previously skipped test now passes
- [ ] Write round-trip test cases
- [ ] Run round-trip tests
- [ ] Verify round-trip preserves data
- [ ] Write comprehensive integration tests
- [ ] Test binary literals in blocks, functions, expressions
- [ ] Test series operations with literals
- [ ] Test I/O operations with binaries
- [ ] Run full test suite: `make test`
- [ ] Verify no regressions in existing tests
- [ ] Run test summary: `make test-summary`
- [ ] Manual integration testing in REPL

### Final Validation

- [ ] All unit tests pass
- [ ] All integration tests pass
- [ ] No test regressions
- [ ] Round-trip works: parse → mold → parse
- [ ] Manual REPL testing confirms functionality
- [ ] Code follows Viro style guidelines (no comments, constructors, imports)
- [ ] Documentation updated (if needed)
- [ ] Example scripts work with binary literals
- [ ] Performance acceptable (no obvious bottlenecks)

---

## Success Criteria

### Functional Requirements

✓ **Binary literal parsing**:
- Empty binary `#{}` parses to zero-length byte slice
- Single byte `#{FF}` parses to `[]byte{0xFF}`
- Multiple bytes `#{DEADBEEF}` parses correctly
- Whitespace allowed: `#{DE AD BE EF}` accepted
- Case insensitive: `#{ff}` == `#{FF}` == `#{Ff}`

✓ **Display formatting**:
- Mold outputs uppercase hex: `#{DE AD BE EF}`
- Mold includes spaces between bytes for readability
- Form truncates large binaries with "... (N bytes)"
- Empty binary displays as `#{}`

✓ **Round-trip capability**:
- parse(mold(binary)) equals original binary
- No data loss through format conversion
- Byte order preserved exactly

✓ **Integration**:
- Binary literals work in all contexts (blocks, functions, expressions)
- Series operations fully functional with literals
- Type system recognizes binary type
- I/O operations handle binary data

### Quality Requirements

✓ **Code quality**:
- Zero inline comments (docs in package/markdown only)
- Constructor functions used exclusively
- Imports organized correctly
- Error handling follows Viro patterns
- Naming follows conventions

✓ **Test coverage**:
- All success paths tested
- All error paths tested
- Edge cases covered (empty, large, boundary)
- Round-trip tests pass
- No regression in existing tests

✓ **Performance**:
- Hex encoding reasonably fast
- Large binaries don't cause hangs
- Memory usage acceptable
- Form() truncation prevents excessive output

✓ **User experience**:
- Clear error messages for invalid literals
- Readable display format with spaces
- Truncated display for large binaries
- Consistent behavior across features

---

## Next Steps After Completion

### Documentation

- Update language documentation with binary literal syntax
- Add examples to docs/examples
- Document hex format requirements
- Add binary type to type system documentation

### Potential Enhancements (Future)

- Base64 encoding/decoding natives
- Bitwise operations on binaries (and, or, xor, shift)
- Binary concatenation operations
- Binary search/pattern matching
- Compression/decompression natives
- Checksum/hash natives (MD5, SHA, CRC)

### Related Features

- URL encoding/decoding (works with binaries)
- HTTP body handling (binary request/response)
- Crypto operations (encryption, signing)
- Image/media file handling

---

## Summary

This plan provides comprehensive guidance for implementing binary literal support in Viro through three distinct phases: fixing display methods, adding parser support, and integration testing. The phased approach allows independent testing at each stage and minimizes integration issues.

**Key implementation decisions:**
- Use uppercase hex in Mold() output with spaces: `#{DE AD BE EF}`
- Truncate Form() at 64 bytes for readability
- Parser accepts any case hex with flexible whitespace
- parseHex() helper handles validation and conversion
- Round-trip capability ensures format consistency

**Critical path:**
1. Fix Mold/Form → enables testing display without parser
2. Add parser → enables literal syntax
3. Integration tests → validates complete feature

**Validation strategy:**
- Unit tests for each component
- Parser tests for syntax coverage
- Round-trip tests for format compatibility
- Integration tests for feature completeness
- Existing test enablement proves series operations work

The coder agent should follow this plan sequentially, validating each phase before proceeding to the next, and using the decision frameworks provided to make appropriate implementation choices within Viro's architectural constraints.
