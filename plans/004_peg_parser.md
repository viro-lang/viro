# Migration Plan: Custom Parser to Pigeon PEG Generator

**Project**: Viro Interpreter  
**Created**: 2025-01-25  
**Status**: Planning Phase  
**Target**: Replace hand-written parser in `internal/parse/parse.go` with Pigeon-generated parser

---

## Executive Summary

Migrate from the current 705-line hand-written two-stage parser (tokenization + parsing) to a single-stage PEG grammar using Pigeon. This eliminates the tokenization layer entirely - PEG rules parse text directly to flat sequences of Viro Value objects.

**Benefits**:
- **Simplicity**: Single-stage parsing (Text → flat Values) vs two-stage (Text → Tokens → Values)
- **Maintainability**: Declarative PEG grammar vs ~500 lines of imperative code
- **Correctness**: Formal grammar reduces edge-case bugs
- **Extensibility**: Adding syntax is simpler (add grammar rules vs manual token handling)
- **Documentation**: PEG grammar serves as executable specification
- **Clean separation**: Parser reads syntax as-is; evaluator handles transformation

**Risks**:
- Breaking changes to error messages (mitigation: comprehensive error message tests)
- Performance characteristics may differ (mitigation: benchmarking and optimization)

---

## Current Parser Analysis

### Architecture

**Current two-stage process** (to be simplified):
1. **Tokenization** (`tokenize()` at line 113): Text → Tokens (~300 lines)
2. **Parsing** (`parseExpression()`, `parsePrimary()` at lines 457, 481): Tokens → Values (~200 lines)

**New PEG approach** (single-stage):
- **Direct parsing**: Text → flat Value sequences in one pass
- Parser reads syntax literally: `3 + 4 * 2` → `[IntVal(3), WordVal("+"), IntVal(4), WordVal("*"), IntVal(2)]`
- NO transformation in parser (evaluator handles that)
- PEG rules handle both lexical and syntactic parsing
- Semantic actions construct Value objects directly
- **Simpler architecture**: ~500 lines eliminated, replaced by declarative grammar

### Key Features to Preserve

1. **Parse syntax as-is** (no transformation!)
   - Parser reads `3 + 4 * 2` as flat sequence: `[3, +, 4, *, 2]`
   - Parser does NOT build nested structures
   - Evaluator handles any transformation during evaluation
   - Parser's job: text → flat Value sequence

2. **Value types and special syntax**:
   - Set-words: `variable:`
   - Get-words: `:variable`
   - Lit-words: `'literal`
   - Paths: `user.address.city` (word-based), `things.1` (word + integer index)
     - First element MUST be a word
     - Following elements can be words OR integers
     - Examples: `user.name`, `data.0`, `obj.field.subfield`
     - NOT valid: `1.field`, `42.0.name` (can't start with number)
   - Refinements: `--flag`, `--option value`
   - Datatypes: `integer!`, `string!`
   - Comments: `; comment to EOL`

3. **Scientific notation**: `1.5e+2`, `2.5E-3`, `6.022e23`

4. **Block vs Paren distinction**:
   - Blocks `[...]`: Creates BlockValue containing flat sequence
   - Parens `(...)`: Creates ParenValue containing flat sequence
   - Parser just creates correct Value type with flat contents

5. **Error handling**: Rich error messages with source context

### Test Coverage

- **92.7% coverage** achieved (from 53.5%)
- 5 test files: `error_test.go`, `tokenize_test.go`, `parse_edge_test.go`, `format_test.go`, `integration_test.go`
- Comprehensive test suite for regression detection

---

## Pigeon PEG Parser Overview

### What is Pigeon?

- **PEG (Parsing Expression Grammar)** parser generator for Go
- Inspired by PEG.js, produces pure Go code
- Supports semantic actions in Go for Value construction
- UTF-8 Unicode support built-in

### Installation

```bash
go install github.com/mna/pigeon@latest
```

### Basic Workflow

1. Write PEG grammar in `.peg` file with Go semantic actions
2. Generate parser: `pigeon -o parser.go grammar.peg`
3. Import and use generated parser in Go code

### Key PEG Concepts

- **Ordered choice**: `/` tries alternatives left-to-right
- **Sequences**: Space-separated items must all match
- **Predicates**: `&expr` (positive lookahead), `!expr` (negative lookahead)
- **Repetition**: `*` (zero or more), `+` (one or more), `?` (optional)
- **Semantic actions**: Go code blocks `{ return ... }` to build Values

---

## Migration Strategy

### Phase 1: PEG Grammar Design (2-3 days)

**Goal**: Write complete PEG grammar that parses text to flat Value sequences

**Tasks**:

1. **Create `grammar/viro.peg` file structure**
   - Initializer block with imports and helper functions
   - Lexical rules (whitespace, comments)
   - Value rules (all Viro value types)

2. **Define grammar rules** (direct text-to-flat-values parsing):
   ```peg
   // Top-level rule: parse input into flat sequence of values
   Input <- _ vals:Value* _ EOF {
     return vals, nil
   }
   
   // Value: any Viro value (read as-is, no transformation!)
   Value <- _ val:(Decimal / Integer / String / Block / Paren / 
                   SetWord / GetWord / LitWord / Path / 
                   Datatype / Word) _ {
     return val, nil
   }
   
   // Literals - parse directly to Value objects
   Integer <- num:$('-'? [0-9]+ !('.' / [eE])) {
     return value.IntVal(toInt(num)), nil
   }
   
   Decimal <- num:$('-'? [0-9]+ ('.' [0-9]+)? ([eE] [+-]? [0-9]+)?) {
     d, scale := toDecimal(num)
     return value.DecimalVal(d, scale), nil
   }
   
   String <- '"' chars:$([^"])* '"' {
     return value.StrVal(chars), nil
   }
   
   // Word variants - parse directly to appropriate Value types
   SetWord <- word:WordChars ':' ![:] {
     return value.SetWordVal(word), nil
   }
   
   GetWord <- ':' word:WordChars {
     return value.GetWordVal(word), nil
   }
   
   LitWord <- '\'' word:WordChars {
     return value.LitWordVal(word), nil
   }
   
   Word <- word:WordChars {
     return value.WordVal(word), nil
   }
   
   WordChars <- $([a-zA-Z_?!] [a-zA-Z0-9_?!-]*)
   
   // Path - first segment MUST be a word, rest can be word or integer
   Path <- segments:PathSegments ':'? {
     return buildPath(segments), nil
   }
   
   PathSegments <- first:WordChars 
                   rest:('.' (WordChars / [0-9]+))+ {
     return buildPathSegments(first, rest), nil
   }
   
   // Blocks and parens - parse contents as flat sequences
   Block <- '[' _ vals:Value* _ ']' {
     return value.BlockVal(vals), nil
   }
   
   Paren <- '(' _ vals:Value* _ ')' {
     return value.ParenVal(vals), nil
   }
   
   
   // Datatype literals
   Datatype <- word:WordChars '!' {
     return value.DatatypeVal(word + "!"), nil
   }
   
   // Whitespace and comments (consumed, not returned)
   _ <- ([ \t\n\r] / Comment)*
   Comment <- ';' [^\n]* '\n'?
   EOF <- !.
   ```

3. **Implement semantic actions** (in PEG initializer block):
   ```go
   {
   package parse
   
   import (
       "strconv"
       "strings"
       "github.com/ericlagergren/decimal"
       "github.com/marcin-radoszewski/viro/internal/core"
       "github.com/marcin-radoszewski/viro/internal/value"
   )
   
   func toInt(s string) int64 {
       n, _ := strconv.ParseInt(s, 10, 64)
       return n
   }
   
   func toDecimal(s string) (*decimal.Big, int16) {
       d := new(decimal.Big)
       d.SetString(s)
       scale := calculateScale(s)
       return d, scale
   }
   
   func calculateScale(s string) int16 {
       if idx := strings.Index(s, "."); idx >= 0 {
           endIdx := len(s)
           if eIdx := strings.IndexAny(s, "eE"); eIdx > idx {
               endIdx = eIdx
           }
           return int16(endIdx - idx - 1)
       }
       return 0
   }
   
   func buildPath(segments any) core.Value {
       segs := segments.([]value.PathSegment)
       return value.PathVal(value.NewPath(segs, value.NoneVal()))
   }
   
   func buildPathSegments(first, rest any) []value.PathSegment {
       // Build path segments: first MUST be word, rest can be word or integer
       var segments []value.PathSegment
       
       // First segment is always a word
       segments = append(segments, value.PathSegment{
           Type: value.PathSegmentWord,
           Value: first.(string),
       })
       
       // Rest segments can be word or integer
       for _, item := range toSlice(rest) {
           parts := toSlice(item)
           segment := parts[1] // Skip the '.' part
           
           // Try to parse as integer
           if num, err := strconv.ParseInt(segment.(string), 10, 64); err == nil {
               segments = append(segments, value.PathSegment{
                   Type: value.PathSegmentIndex,
                   Value: num,
               })
           } else {
               // It's a word
               segments = append(segments, value.PathSegment{
                   Type: value.PathSegmentWord,
                   Value: segment.(string),
               })
           }
       }
       
       return segments
   }
   
   func toSlice(v any) []any {
       if v == nil { return nil }
       return v.([]any)
   }
   }
   ```

4. **Key principle**: Parser reads as-is!
   - PEG rules combine lexical + syntactic parsing
   - Semantic actions build Value objects (NO transformation)
   - Result: Text → flat []core.Value sequence
   - Example: `3 + 4` → `[IntVal(3), WordVal("+"), IntVal(4)]` (flat!)
   - Example: `true` → `WordVal("true")` (NOT `LogicVal(true)`)
   - Evaluator handles any transformation and semantic interpretation

**Deliverables**:
- `grammar/viro.peg` with complete single-pass grammar
- Documented grammar rules showing direct text-to-flat-values parsing
- Helper functions in initializer block (conversion only, no transformation)
- No separate tokenization stage
- No transformation (evaluator's responsibility)

---

### Phase 2: Parser Generation & Integration (2 days)

**Goal**: Generate parser from grammar and integrate into codebase

**Tasks**:

1. **Create build process**
   ```makefile
   # Add to Makefile
   .PHONY: grammar
   grammar:
   	pigeon -o internal/parse/generated_parser.go grammar/viro.peg
   
   .PHONY: build
   build: grammar
   	go build -o viro ./cmd/viro
   ```

2. **Generate initial parser**
   ```bash
   pigeon -o internal/parse/generated_parser.go grammar/viro.peg
   ```

3. **Create adapter layer** in `internal/parse/parse.go`:
   ```go
   // Parse wraps the generated parser
   func Parse(input string) ([]core.Value, error) {
       result, err := ParseReader("", strings.NewReader(input))
       if err != nil {
           return nil, err
       }
       return result.([]core.Value), nil
   }
   ```

4. **Preserve public API**:
   - Keep `Parse()` function signature unchanged
   - Keep `Format()` function for debugging
   - Maintain `ParseEval()` for REPL compatibility

5. **Add grammar to version control**:
   ```bash
   git add grammar/viro.peg
   git add internal/parse/generated_parser.go
   ```

**Deliverables**:
- Generated parser at `internal/parse/generated_parser.go`
- Thin adapter layer preserving existing API
- Updated Makefile with grammar generation

---

### Phase 3: Testing & Validation (3-4 days)

**Goal**: Ensure new parser passes all existing tests and matches behavior

**Tasks**:

1. **Run existing test suite**
   ```bash
   go test ./internal/parse/... -v
   go test ./test/contract/... -v
   go test ./test/integration/... -v
   ```

2. **Fix grammar issues revealed by tests**
   - Iterate: Fix grammar → regenerate → retest
   - Track deviations from expected behavior
   - Document any intentional changes

3. **Benchmark performance**
   ```bash
   go test -bench=. ./internal/parse/...
   go test -bench=. ./test/integration/...
   ```
   - Compare with baseline from old parser
   - Identify performance regressions
   - Optimize grammar if needed (memoization flags)

4. **Error message validation**
   - Review error output quality
   - Enhance error messages if needed
   - Add source position tracking

5. **Add grammar-specific tests**:
   - Test edge cases specific to PEG (backtracking, greedy matching)
   - Verify flat Value sequences (no nesting!)
   - Test: `Parse("3 + 4 * 2")` should produce `[IntVal(3), WordVal("+"), IntVal(4), WordVal("*"), IntVal(2)]`
   - Parser does NOT build nested structures

**Acceptance Criteria**:
- ✅ All existing 92.7% test coverage passing
- ✅ No performance regression > 20%
- ✅ Error messages maintain quality
- ✅ All contracts validated (see `specs/001-*/contracts/`)

**Deliverables**:
- Updated grammar with all fixes
- Performance benchmark report
- Test results documenting 100% pass rate

---

### Phase 4: Code Cleanup & Documentation (1-2 days)

**Goal**: Remove old parser code, update docs, prepare for merge

**Tasks**:

1. **Remove old parser**:
   ```bash
   rm internal/parse/parse.go
   ```

2. **Reorganize parse package**:
   ```
   internal/parse/
   ├── generated_parser.go  (Pigeon output - single-stage parser)
   ├── parse.go             (Thin adapter: Parse() wrapper)
   └── format.go            (Format function moved here)
   ```
   
   **Key difference**: Much simpler! No tokenization, no transformation, no old code.

3. **Update documentation**:
   - Add `grammar/README.md` explaining PEG grammar
   - Update `docs/` with grammar reference
   - Add grammar modification guide
   - Document build process with grammar generation

4. **Update AGENTS.md and CLAUDE.md**:
   ```markdown
   ## Build & Test Commands
   - Generate grammar: `make grammar` or `pigeon -o internal/parse/generated_parser.go grammar/viro.peg`
   - Build: `make build` (includes grammar generation)
   - Test: `go test ./...`
   ```

5. **Add grammar to CI/CD** (`.github/workflows/opencode.yml`):
   ```yaml
   - name: Install Pigeon
     run: go install github.com/mna/pigeon@latest
   
   - name: Generate Parser
     run: make grammar
   
   - name: Verify Grammar Up-to-Date
     run: git diff --exit-code internal/parse/generated_parser.go
   ```

**Deliverables**:
- Clean package structure
- Updated documentation
- CI/CD pipeline with grammar generation
- Migration complete!

---

### Phase 5: Post-Migration Enhancements (Optional, 1-2 days)

**Goal**: Leverage PEG for future improvements

**Potential Enhancements**:

1. **Better error recovery**: Use Pigeon's error recovery mechanisms
2. **Syntax extensions**: Add new language features via grammar rules
3. **Grammar testing**: Property-based testing of grammar rules
4. **Grammar optimization**: Memoization, cut operators for performance
5. **Source position metadata**: Richer error reporting

---

## Risk Mitigation

### Risk 1: Performance Regression

**Mitigation**:
- Benchmark before/after migration
- Use Pigeon optimization flags (`-optimize-grammar`)
- Profile hot paths if needed
- Acceptable threshold: <20% performance loss (tradeoff for maintainability)

### Risk 2: Breaking Changes

**Mitigation**:
- Comprehensive test suite (92.7% coverage) catches regressions
- Git history preserves old parser if rollback needed
- Run full integration test suite
- Validate all contracts in `specs/001-*/contracts/`
- When tests fail: describe issue to user and ask whether to fix parser or tests

### Risk 3: Error Message Quality

**Mitigation**:
- Test error messages explicitly
- Enhance Pigeon error reporting with custom error types
- Add source position tracking

### Risk 4: Grammar Bugs

**Mitigation**:
- Iterative development: test early, test often
- Start with minimal grammar, expand incrementally
- Peer review grammar rules

---

## Timeline Estimate

| Phase | Duration | Dependencies |
|-------|----------|--------------|
| Phase 1: Grammar Design | 2-3 days | None |
| Phase 2: Integration | 2 days | Phase 1 |
| Phase 3: Testing | 3-4 days | Phase 2 |
| Phase 4: Cleanup | 1-2 days | Phase 3 |
| **Total** | **8-11 days** | Sequential |

**Optional Phase 5**: 1-2 days (can be done later)

---

## Success Criteria

✅ **Migration Complete When**:
1. All existing tests pass (92.7% coverage maintained)
2. Performance within acceptable range (<20% regression)
3. Error messages maintain quality and context
4. Grammar is documented and understandable
5. Build process includes grammar generation
6. CI/CD pipeline validates grammar
7. Parser produces flat Value sequences (no transformation)

✅ **Quality Gates**:
- [ ] All contract tests pass (`test/contract/`)
- [ ] All integration tests pass (`test/integration/`)
- [ ] Benchmark performance acceptable
- [ ] Code review approved
- [ ] Documentation updated
- [ ] Parser outputs verified as flat sequences

---

## Rollback Plan

If migration fails or has critical issues:

1. **Immediate rollback**:
   ```bash
   git checkout internal/parse/parse.go
   rm internal/parse/generated_parser.go
   rm internal/parse/format.go  # If it was created
   ```

2. **Revert grammar changes**:
   ```bash
   git rm -rf grammar/
   git checkout Makefile
   ```

3. **Restore CI/CD**:
   ```bash
   git checkout .github/workflows/
   ```

4. **Validate rollback**:
   ```bash
   make test
   ```

**Note**: Since old parser is removed (not archived), git rollback restores it.

---

## Next Steps

1. **Review this plan** with stakeholders
2. **Create feature branch**: `git checkout -b feature/pigeon-parser-migration`
3. **Start Phase 1**: Begin PEG grammar design
4. **Iterate**: Test-driven development for each grammar rule
5. **Complete phases sequentially** with validation at each stage

---

## References

- **Pigeon Documentation**: https://pkg.go.dev/github.com/mna/pigeon
- **Pigeon GitHub**: https://github.com/mna/pigeon
- **PEG Tutorial**: https://github.com/mna/pigeon/wiki
- **Current Parser**: `internal/parse/parse.go` (705 lines)
- **Test Coverage Report**: `plans/001_parser_refactor.md`
- **Grammar Contracts**: `specs/001-implement-the-core/contracts/`

---

**End of Migration Plan**

---

## Instructions for LLM Implementing This Plan

### When Tests Fail After Migration

If any tests fail after implementing the new parser, follow this procedure:

1. **Analyze the failure**:
   - Identify which test is failing
   - Understand what the test expects
   - Understand what the new parser produces
   - Determine the root cause

2. **Report to user**:
   ```
   Test failure detected in <test_name>:
   
   **Expected behavior**: <what test expects>
   **Actual behavior**: <what parser produces>
   
   **Example**:
   Input: <test input>
   Test expects: <expected output>
   Parser produces: <actual output>
   
   **Analysis**: <your understanding of the issue>
   
   This could be either:
   a) The new parser needs to be fixed to match the expected behavior
   b) The test expectations need to be updated to match the new (correct) parser behavior
   
   Should I fix the grammar/parser or update the test?
   ```

3. **Wait for user decision** - Do NOT automatically fix tests or parser

4. **Implement the fix** according to user's decision:
   - If "fix parser": Update grammar, regenerate, retest
   - If "fix test": Update test expectations, document why

5. **Verify the fix** and continue to next failing test

### Key Principles

- **Never silently fix tests** - Always ask the user first
- **Provide concrete examples** - Show actual vs expected output
- **Explain your reasoning** - Help user make informed decision
- **One issue at a time** - Don't batch multiple unrelated fixes
- **Document changes** - Keep track of what was changed and why

---
