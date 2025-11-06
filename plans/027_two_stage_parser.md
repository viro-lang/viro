# 027 Two-Stage Parser Refactoring

## Status: COMPLETED ✅

All 6 phases of the two-stage parser implementation have been completed successfully. This plan document is preserved for historical reference.

## Historical Context

This plan documents the migration from a grammar-based parser to Viro's current two-stage parser architecture (tokenizer + semantic parser).

**Previous approach:** Grammar-based parser that mixed lexical and semantic analysis

**Current approach:** Two-stage parser with clear separation:
- **Stage 1 (Tokenizer)**: Split input into tokens (space-separated literals, brackets, strings)
- **Stage 2 (Parser)**: Classify literals and build value structures

This refactoring enabled:

- ✅ Metaprogramming: Exposed `tokenize`, `parse`, `load`, `classify` functions to Viro
- ✅ Simplicity: Each stage has single responsibility
- ✅ Maintainability: Plain Go code, no external dependencies
- ✅ Extensibility: Support for custom dialects and user-defined syntax

Reference: `docs/two-stage-parser.md` for complete architecture description.

## Completion Summary

All goals achieved:

1. ✅ Replaced old parser with two-stage tokenizer + parser
2. ✅ Maintained 100% compatibility with existing behavior (all 2053 tests passing)
3. ✅ Excellent parsing performance (158ns-875ns tokenization, 5.6μs-222μs parsing)
4. ✅ Exposed parsing primitives to Viro code for metaprogramming
5. ✅ Removed external dependencies

---

## Original Implementation Plan

The sections below document the original implementation approach for historical reference.

## Implementation Plan

### Phase 1: Tokenizer Implementation

**File**: `internal/tokenize/tokenizer.go`

Create tokenizer with TDD approach:

**Step 1.1: Define Token types**

```go
type TokenType int

const (
    TokenLiteral    TokenType = iota
    TokenString
    TokenLParen
    TokenRParen
    TokenLBracket
    TokenRBracket
    TokenEOF
)

type Token struct {
    Type   TokenType
    Value  string
    Line   int
    Column int
}

type Tokenizer struct {
    input   string
    pos     int
    line    int
    column  int
}
```

**Step 1.2: Write tokenizer tests** (`internal/tokenize/tokenizer_test.go`)

Test cases:

- Empty input → [EOF]
- Single literal → [Literal("abc"), EOF]
- Multiple literals → [Literal("abc"), Literal("def"), EOF]
- Whitespace separation (space, tab, newline)
- Brackets: `[`, `]`, `(`, `)`
- Strings: `"hello"`, `"with \"quotes\""`, `"with\nnewline"`
- Comments: `; comment\nabc` → [Literal("abc"), EOF]
- Mixed: `abc [1 "test"] def` → proper token sequence
- Error: unclosed string `"hello`
- Error: invalid escape `"hello\x"`
- Position tracking: verify line/column for each token

**Step 1.3: Implement tokenizer**

Core methods:

```go
func NewTokenizer(input string) *Tokenizer
func (t *Tokenizer) NextToken() (Token, error)
func (t *Tokenizer) Tokenize() ([]Token, error)

// Internal helpers
func (t *Tokenizer) skipWhitespace()
func (t *Tokenizer) skipComment()
func (t *Tokenizer) readString() (string, error)
func (t *Tokenizer) readLiteral() string
func (t *Tokenizer) peek() byte
func (t *Tokenizer) advance()
```

Implementation logic (see `docs/two-stage-parser.md` for detailed algorithm):

1. Skip whitespace and comments
2. Check for EOF
3. Check for brackets/parens → emit single-char token
4. Check for string (`"`) → read until closing quote with escape handling
5. Otherwise → read literal (non-whitespace, non-bracket chars)

**Step 1.4: Verify tokenizer tests pass**

### Phase 2: Parser Implementation

**File**: `internal/parse/semantic_parser.go`

**Step 2.1: Define Parser structure**

```go
type Parser struct {
    tokens []Token
    pos    int
}

func NewParser(tokens []Token) *Parser
func (p *Parser) Parse() ([]core.Value, error)

// Internal helpers
func (p *Parser) parseValue() (core.Value, error)
func (p *Parser) parseBlock() (core.Value, error)
func (p *Parser) parseParen() (core.Value, error)
func (p *Parser) parseObject() (core.Value, error)
func (p *Parser) classifyLiteral(text string) (core.Value, error)
func (p *Parser) parsePath(text string) ([]value.PathSegment, error)
```

**Step 2.2: Write parser tests** (`internal/parse/semantic_parser_test.go`)

Literal classification tests:

- Integers: `"5"` → IntegerVal(5), `"-10"` → IntegerVal(-10)
- Decimals: `"3.14"` → DecimalVal(3.14), `"1e10"` → DecimalVal(1e10)
- Set-words: `"abc:"` → SetWordVal("abc")
- Get-words: `":abc"` → GetWordVal("abc")
- Lit-words: `"'abc"` → LitWordVal("abc")
- Words: `"abc"` → WordVal("abc"), `"--flag"` → WordVal("--flag")
- Datatypes: `"integer!"` → DatatypeVal("integer!")
- Paths: `"obj.field"` → PathVal([Word("obj"), Word("field")])
- Index paths: `"data.1"` → PathVal([Word("data"), Index(1)])
- Set-paths: `"obj.x:"` → SetPathVal([Word("obj"), Word("x")])
- Get-paths: `":obj.x"` → GetPathVal([Word("obj"), Word("x")])

Structure building tests:

- Empty block: `[]` → BlockVal([])
- Simple block: `[1 2 3]` → BlockVal([IntVal(1), IntVal(2), IntVal(3)])
- Nested blocks: `[[1] [2]]` → BlockVal([BlockVal([IntVal(1)]), BlockVal([IntVal(2)])])
- Parens: `(1 + 2)` → ParenVal([IntVal(1), WordVal("+"), IntVal(2)])
- Mixed: `abc: [1 2]` → [SetWordVal("abc"), BlockVal([IntVal(1), IntVal(2)])]

Error tests:

- Unclosed block: `[1 2` → error with position
- Unexpected closing bracket: `]` → error
- Invalid path: `obj.` → error (empty segment)

**Step 2.3: Implement classifyLiteral function**

Algorithm (see `docs/two-stage-parser.md` for complete logic):

```go
func (p *Parser) classifyLiteral(text string) (core.Value, error) {
    // 1. Check for lit-word (starts with ')
    if strings.HasPrefix(text, "'") {
        return value.NewLitWordVal(text[1:]), nil
    }

    // 2. Check for get-word or get-path (starts with :)
    if strings.HasPrefix(text, ":") {
        base := text[1:]
        if strings.Contains(base, ".") {
            segments, err := p.parsePath(base)
            if err != nil { return core.Value{}, err }
            return value.GetPathVal(value.NewGetPath(segments, value.NewNoneVal())), nil
        }
        return value.NewGetWordVal(base), nil
    }

    // 3. Check for set-word or set-path (ends with :)
    if strings.HasSuffix(text, ":") {
        base := text[:len(text)-1]
        if strings.Contains(base, ".") {
            segments, err := p.parsePath(base)
            if err != nil { return core.Value{}, err }
            return value.SetPathVal(value.NewSetPath(segments, value.NewNoneVal())), nil
        }
        return value.NewSetWordVal(base), nil
    }

    // 4. Check for path (contains .)
    if strings.Contains(text, ".") {
        segments, err := p.parsePath(text)
        if err != nil { return core.Value{}, err }
        return value.PathVal(value.NewPath(segments, value.NewNoneVal())), nil
    }

    // 5. Check for integer
    if matched, _ := regexp.MatchString(`^-?[0-9]+$`, text); matched {
        n, _ := strconv.ParseInt(text, 10, 64)
        return value.NewIntVal(n), nil
    }

    // 6. Check for decimal
    if matched, _ := regexp.MatchString(`^-?[0-9]+\.[0-9]+([eE][+-]?[0-9]+)?$`, text); matched {
        // Parse decimal with ericlagergren/decimal
        d := new(decimal.Big)
        d.SetString(text)
        scale := calculateScale(text)
        return value.DecimalVal(d, scale), nil
    }
    if matched, _ := regexp.MatchString(`^-?[0-9]+[eE][+-]?[0-9]+$`, text); matched {
        d := new(decimal.Big)
        d.SetString(text)
        scale := calculateScale(text)
        return value.DecimalVal(d, scale), nil
    }

    // 7. Check for datatype (ends with !)
    if strings.HasSuffix(text, "!") {
        return value.NewDatatypeVal(text), nil
    }

    // 8. Default: word (includes refinements like --flag)
    return value.NewWordVal(text), nil
}
```

**Step 2.4: Implement parsePath function**

```go
func (p *Parser) parsePath(text string) ([]value.PathSegment, error) {
    if text == "" {
        return nil, fmt.Errorf("empty path")
    }

    parts := strings.Split(text, ".")
    segments := make([]value.PathSegment, 0, len(parts))

    for _, part := range parts {
        if part == "" {
            return nil, fmt.Errorf("empty path segment")
        }

        // Try parsing as integer index
        if n, err := strconv.ParseInt(part, 10, 64); err == nil {
            segments = append(segments, value.PathSegment{
                Type:  value.PathSegmentIndex,
                Value: n,
            })
        } else {
            segments = append(segments, value.PathSegment{
                Type:  value.PathSegmentWord,
                Value: part,
            })
        }
    }

    return segments, nil
}
```

**Step 2.5: Implement structure building (blocks and parens)**

```go
func (p *Parser) parseValue() (core.Value, error) {
    token := p.tokens[p.pos]
    p.pos++

    switch token.Type {
    case TokenLiteral:
        return p.classifyLiteral(token.Value)

    case TokenString:
        return value.NewStrVal(token.Value), nil

    case TokenLBracket:
        values, err := p.parseUntil(TokenRBracket, "block")
        if err != nil {
            return core.Value{}, err
        }
        return value.NewBlockVal(values), nil

    case TokenLParen:
        values, err := p.parseUntil(TokenRParen, "paren")
        if err != nil {
            return core.Value{}, err
        }
        return value.NewParenVal(values), nil

    case TokenRBracket, TokenRParen:
        return core.Value{}, fmt.Errorf("unexpected closing bracket at line %d, column %d", token.Line, token.Column)

    case TokenEOF:
        return core.Value{}, fmt.Errorf("unexpected EOF")

    default:
        return core.Value{}, fmt.Errorf("unknown token type")
    }
}

// Generic helper: parse values until closing token is found
func (p *Parser) parseUntil(closingType TokenType, structName string) ([]core.Value, error) {
    values := []core.Value{}

    for p.pos < len(p.tokens) {
        token := p.tokens[p.pos]

        if token.Type == closingType {
            p.pos++
            return values, nil
        }

        if token.Type == TokenEOF {
            return nil, fmt.Errorf("unclosed %s", structName)
        }

        val, err := p.parseValue()
        if err != nil {
            return nil, err
        }
        values = append(values, val)
    }

    return nil, fmt.Errorf("unclosed %s", structName)
}
```

**Step 2.6: Implement top-level Parse function**

```go
func (p *Parser) Parse() ([]core.Value, error) {
    values := []core.Value{}

    for p.pos < len(p.tokens) {
        token := p.tokens[p.pos]

        if token.Type == TokenEOF {
            break
        }

        val, err := p.parseValue()
        if err != nil {
            return nil, err
        }
        values = append(values, val)
    }

    return values, nil
}
```

**Step 2.7: Verify parser tests pass**

### Phase 3: Integration and Migration

**Step 3.1: Update parse.Parse() to use new implementation**

Modify `internal/parse/parse.go`:

```go
func Parse(input string) ([]core.Value, error) {
    // Tokenize
    tokenizer := tokenize.NewTokenizer(input)
    tokens, err := tokenizer.Tokenize()
    if err != nil {
        // Convert to verror
        return nil, verror.NewSyntaxError(verror.ErrIDInvalidSyntax, [3]string{err.Error(), "", ""})
    }

    // Parse
    parser := NewParser(tokens)
    values, err := parser.Parse()
    if err != nil {
        // Convert to verror with proper error ID
        errID := verror.ErrIDInvalidSyntax
        if strings.Contains(err.Error(), "unclosed block") {
            errID = verror.ErrIDUnclosedBlock
        } else if strings.Contains(err.Error(), "unclosed paren") {
            errID = verror.ErrIDUnclosedParen
        }

        vErr := verror.NewSyntaxError(errID, [3]string{err.Error(), "", ""})
        if input != "" {
            vErr.SetNear(input)
        }
        return nil, vErr
    }

    return values, nil
}
```

**Step 3.2: Run full test suite**

```bash
go test ./...
```

Expected: All existing tests pass without changes.

**Step 3.3: Remove old dependencies**

✅ COMPLETED in Phase 5:
1. ✅ Old grammar files removed
2. ✅ Generated parser code removed  
3. ✅ External tools removed from build process
4. ✅ Makefile updated to remove grammar generation
5. ✅ Build simplified - no code generation needed

**Step 3.4: Verify clean build**

```bash
make clean
make build
go test ./...
```

### Phase 4: Expose Parsing to Viro

**Step 4.1: Add tokenize native function**

File: `internal/native/parse.go`

```go
// Tokenize source code into tokens
// Usage: tokenize "abc: 5"
// Returns: [token! token! token!]
func nativeTokenize(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
    input, ok := value.AsString(args[0])
    if !ok {
        return core.Value{}, verror.NewScriptError("type-mismatch", [3]string{"tokenize", "string!", value.TypeName(args[0].Type)})
    }

    tokenizer := tokenize.NewTokenizer(input)
    tokens, err := tokenizer.Tokenize()
    if err != nil {
        return core.Value{}, verror.NewSyntaxError(verror.ErrIDInvalidSyntax, [3]string{err.Error(), "", ""})
    }

    // Convert tokens to Viro values (token! datatype)
    result := make([]core.Value, 0, len(tokens))
    for _, tok := range tokens {
        // Create token object with fields: type, value, line, column
        tokenObj := createTokenObject(tok)
        result = append(result, tokenObj)
    }

    return value.NewBlockVal(result), nil
}
```

**Step 4.2: Add parse native function**

```go
// Parse tokens into values
// Usage: parse [token! token! token!]
// Returns: [value! value!]
func nativeParse(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
    tokensBlock, ok := value.AsBlock(args[0])
    if !ok {
        return core.Value{}, verror.NewScriptError("type-mismatch", [3]string{"parse", "block!", value.TypeName(args[0].Type)})
    }

    // Convert token objects to Token structs
    tokens, err := convertToTokens(tokensBlock)
    if err != nil {
        return core.Value{}, err
    }

    // Parse
    parser := parse.NewParser(tokens)
    values, err := parser.Parse()
    if err != nil {
        return core.Value{}, verror.NewSyntaxError(verror.ErrIDInvalidSyntax, [3]string{err.Error(), "", ""})
    }

    return value.NewBlockVal(values), nil
}
```

**Step 4.3: Add load native function**

```go
// Load source code (tokenize + parse)
// Usage: load "abc: 5"
// Returns: [set-word! integer!]
func nativeLoad(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
    input, ok := value.AsString(args[0])
    if !ok {
        return core.Value{}, verror.NewScriptError("type-mismatch", [3]string{"load", "string!", value.TypeName(args[0].Type)})
    }

    values, err := parse.Parse(input)
    if err != nil {
        return core.Value{}, err
    }

    return value.NewBlockVal(values), nil
}
```

**Step 4.4: Add classify native function**

```go
// Classify a literal string into its Viro value type
// Usage: classify "abc:"
// Returns: set-word!
func nativeClassify(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
    input, ok := value.AsString(args[0])
    if !ok {
        return core.Value{}, verror.NewScriptError("type-mismatch", [3]string{"classify", "string!", value.TypeName(args[0].Type)})
    }

    // Create dummy parser to use classifyLiteral
    parser := parse.NewParser([]tokenize.Token{})
    val, err := parser.ClassifyLiteral(input)
    if err != nil {
        return core.Value{}, verror.NewSyntaxError(verror.ErrIDInvalidSyntax, [3]string{err.Error(), "", ""})
    }

    return val, nil
}
```

**Step 4.5: Register native functions**

Add to `internal/native/register_io.go`:

```go
func RegisterIONatives(reg *Registry) {
    // ... existing registrations

    // Parser functions
    registerAndBind("tokenize", createNativeFunction(
        "tokenize",
        []ParameterSpec{{Name: "source", Eval: true}},
        nil,
        nativeTokenize,
    ))

    registerAndBind("parse", createNativeFunction(
        "parse",
        []ParameterSpec{{Name: "tokens", Eval: true}},
        nil,
        nativeParse,
    ))

    registerAndBind("load", createNativeFunction(
        "load",
        []ParameterSpec{{Name: "source", Eval: true}},
        nil,
        nativeLoad,
    ))

    registerAndBind("classify", createNativeFunction(
        "classify",
        []ParameterSpec{{Name: "literal", Eval: true}},
        nil,
        nativeClassify,
    ))
}
```

**Step 4.6: Write user-facing tests**

File: `test/contract/parser_test.go`

Test cases:

```viro
; Tokenize
tokens: tokenize "abc: 5"
assert: length? tokens = 3

; Load
values: load "abc: 5"
assert: length? values = 2
assert: type-of first values = set-word!
assert: type-of second values = integer!

; Classify
assert: type-of classify "abc:" = set-word!
assert: type-of classify "123" = integer!
assert: type-of classify "--flag" = word!
```

**Step 4.7: Write contracts**

File: `specs/002-implement-deferred-features/contracts/parser.md`

Document:

- `tokenize` signature, behavior, examples
- `parse` signature, behavior, examples
- `load` signature, behavior, examples
- `classify` signature, behavior, examples

### Phase 5: Documentation and Cleanup (COMPLETED ✅)

**Status**: Phase 5 completed. Old parser fully removed and documentation updated.

**Completed tasks:**

1. **Old parser removal verified:**
   - ✅ Old grammar files removed in earlier phases
   - ✅ Generated parser code removed
   - ✅ External dependencies never in go.mod (clean)
   - ✅ Makefile grammar target removed

2. **Documentation updates:**
   - ✅ `grammar/README.md` - Replaced with two-stage parser architecture doc
   - ✅ `AGENTS.md` - Removed grammar generation command references
   - ✅ `plans/027_two_stage_parser.md` - Marked Phase 5 complete
   - ✅ `docs/parser-performance.md` - Removed old parser references
   - ✅ `docs/two-stage-parser.md` - Removed old parser comparisons
   - ✅ `CLAUDE.md` - Updated build commands and architecture
   - ✅ `docs/viro_architecture_knowledge.md` - Updated parser section
   - ✅ `docs/viro_core_knowledge_rag.md` - Updated parser section

3. **Verification:**
   - ✅ All 2053 tests pass
   - ✅ Build works without grammar generation: `make build`
   - ✅ No old parser imports found in codebase
   - ✅ Documentation accurately reflects two-stage architecture

---

### Phase 6: Performance Validation (COMPLETED ✅)

**Status**: Phase 6 completed. Performance validation successful.

**Completed tasks:**

1. **Created comprehensive benchmarks** (`internal/parse/parse_bench_test.go`):
   - ✅ Tokenization benchmarks (simple, medium, complex, long strings)
   - ✅ Semantic parsing benchmarks (various code patterns)
   - ✅ End-to-end benchmarks (real-world scripts)

2. **Ran performance tests:**
   - ✅ Tokenization: 158ns-875ns (sub-microsecond)
   - ✅ Parsing: 5.6μs-222μs for complete scripts
   - ✅ Real-world script (17 lines): ~222μs total
   - ✅ Throughput: ~76,500 lines/sec, ~45M tokens/sec

3. **Documented results** (`docs/parser-performance.md`):
   - ✅ Detailed performance breakdown
   - ✅ Throughput calculations
   - ✅ Memory characteristics analysis
   - ✅ Conclusions: Performance exceeds expectations

**Result**: Parser performance is excellent. No optimization needed.

---

### Original Performance Plan (Historical Reference)

**Step 6.1: Create performance benchmarks**

## Testing Strategy

### Unit Tests

1. **Tokenizer**: Test each token type, whitespace handling, strings, comments, errors
2. **Parser**: Test literal classification, path parsing, structure building, errors
3. **Integration**: Test parse.Parse() with representative inputs

### Integration Tests

1. Run full contract test suite
2. Run REPL smoke tests
3. Run performance benchmarks

### Regression Prevention (COMPLETED ✅)

✅ **Verification completed:**

1. ✅ All 2053 existing tests passed with new parser
2. ✅ New parser produces identical outputs to old implementation
3. ✅ No behavioral changes detected
4. ✅ REPL functionality verified unchanged

## Error Handling

Maintain or improve error quality:

**Tokenizer errors:**

- Unclosed string: include line, column, and context
- Invalid escape: include position and invalid sequence

**Parser errors:**

- Unclosed bracket: include position of opening bracket
- Unexpected closing bracket: include position
- Invalid path: include problematic segment

Convert all errors to appropriate `verror` types:

- `ErrIDInvalidSyntax` - general syntax errors
- `ErrIDUnclosedBlock` - unclosed `[`
- `ErrIDUnclosedParen` - unclosed `(`

## Performance Targets

- **Parsing speed**: At least as fast as PEG, ideally 10-30% faster
- **Memory usage**: No increase in allocations
- **Build time**: Significantly faster (no grammar generation)

## Risks and Mitigations

**Risk**: Subtle behavioral differences from PEG parser
**Mitigation**: Comprehensive comparison testing, regression test suite

**Risk**: Performance regression
**Mitigation**: Benchmark before/after, optimize if needed

**Risk**: Error message quality degradation
**Mitigation**: Compare error outputs, ensure position tracking is accurate

**Risk**: Breaking existing code
**Mitigation**: Feature flag to toggle between parsers during transition

## Acceptance Criteria

- [x] All tokenizer unit tests pass
- [x] All parser unit tests pass
- [x] Full contract test suite passes (no regressions)
- [x] PEG code and dependencies removed (Phase 5 complete)
- [x] Build works without grammar generation
- [x] Native functions (`tokenize`, `parse`, `load-string`, `classify`) implemented and tested
- [x] Documentation updated (contract: specs/002-implement-deferred-features/contracts/parser-natives.md)
- [ ] Performance benchmarks show no regression (Phase 6)
- [x] Error messages maintain or improve quality

## Effort Estimate

- **Tokenizer implementation**: Medium (2-3 hours)
- **Parser implementation**: Large (4-5 hours)
- **Integration**: Small (1 hour)
- **Native functions**: Medium (2-3 hours)
- **Testing and validation**: Medium (2-3 hours)
- **Documentation**: Small (1 hour)

**Total**: 12-16 hours

## Implementation Order

1. Tokenizer (TDD: tests → implementation → verify)
2. Parser (TDD: tests → implementation → verify)
3. Integration (replace PEG, verify no regressions)
4. Native functions (expose to Viro)
5. Documentation and cleanup

## Success Metrics

- Zero test regressions
- Parser performance equal or better than PEG
- Clean build without grammar generation
- Parsing functions usable from Viro code
- Improved maintainability (less code, clearer separation)

## Future Enhancements

Once two-stage parser is stable:

1. **Binary literals**: Add `#{FF00}` support
2. **Unicode escapes**: Support `\u{1F600}` in strings
3. **Custom literals**: Allow user-defined literal types
4. **Parser hooks**: Enable syntax extensions from Viro
5. **Better diagnostics**: Add source ranges to all values

## References

- Architecture: `docs/two-stage-parser.md`
- Current PEG grammar: `grammar/viro.peg`
- Parser entry point: `internal/parse/parse.go`
- Execution model: `docs/execution-model.md`
