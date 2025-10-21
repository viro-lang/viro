# Parser Package Test Coverage Improvement Plan

## Current Coverage Analysis

**Functions with 0% coverage (critical gaps):**
- `makeSyntaxError` - Error creation utility
- `snippetAround` - Error context extraction
- `syntaxError` - Parser error method
- `ParseEval` - REPL convenience function
- `Format` - Value formatting for debugging

**Functions with low coverage:**
- `tokenize`: 62.3% - Missing complex tokenization cases
- `parsePrimary`: 63.6% - Missing error paths and edge cases
- `peek`: 66.7% - Missing EOF handling

## Comprehensive Test Plan

### Phase 1: Error Handling & Edge Cases (Target: +15-20% coverage)

**1. Syntax Error Testing**
- Test `makeSyntaxError` and `snippetAround` functions
- Test various error positions and contexts
- Test error message formatting
- Test snippet extraction at string boundaries

**2. Parser Error Cases**
- Test `syntaxError` method calls
- Test malformed input handling
- Test unexpected EOF scenarios
- Test invalid token sequences

**3. Tokenization Error Cases**
- Unclosed string literals
- Invalid number formats
- Malformed paths (dot without following segment)
- Invalid character sequences
- Unicode edge cases

### Phase 2: Tokenization Coverage (Target: +10-15% coverage)

**4. Complex Number Parsing**
- Scientific notation edge cases (e+, e-, E+, E-)
- Very large/small exponents
- Invalid exponent formats
- Numbers at string boundaries

**5. Path Tokenization**
- Paths starting with numbers (1.field, 42.0.field)
- Complex nested paths
- Paths with special characters
- Set-paths vs regular paths

**6. Word Variants**
- Refinement words (--flag, --option)
- Get-words (:word)
- Lit-words ('word)
- Datatype literals (integer!, string!)
- Special keywords (true, false, none)

**7. Operator Tokenization**
- Multi-character operators (<=, >=, <>)
- Operator precedence edge cases
- Operators in different contexts

### Phase 3: Parsing Coverage (Target: +10-15% coverage)

**8. Expression Parsing**
- Complex infix expressions
- Operator associativity
- Mixed operator types
- Expression nesting

**9. Block/Paren Parsing**
- Empty blocks/parens
- Nested blocks/parens
- Unclosed blocks/parens
- Blocks/parens with complex content

**10. Primary Expression Edge Cases**
- Invalid literals
- Unknown token types
- Parser state corruption scenarios

### Phase 4: Utility Functions (Target: +5-10% coverage)

**11. Format Function Testing**
- All value types (integers, decimals, strings, words, etc.)
- Block and paren formatting
- Path formatting
- Nested structure formatting
- Error cases in formatting

**12. ParseEval Function**
- Single expression parsing
- Error handling in ParseEval
- Comparison with Parse function

### Phase 5: Integration & Boundary Testing (Target: +5-10% coverage)

**13. Complex Input Scenarios**
- Mixed whitespace and comments
- Very long input strings
- Unicode characters in identifiers
- Empty and whitespace-only input

**14. Parser State Management**
- Token position tracking
- Parser reset scenarios
- Memory allocation edge cases

## Implementation Strategy

### Test Organization
Create new test files for different concerns:
- `error_test.go` - Error handling and edge cases
- `tokenize_test.go` - Tokenization specifics
- `parse_edge_test.go` - Parser edge cases
- `format_test.go` - Format function testing

### Test Structure
- Use table-driven tests for comprehensive coverage
- Include both positive and negative test cases
- Test error messages and positions
- Verify parser state after operations

### Coverage Targets by Phase
- **Phase 1 completion**: 70-75% coverage
- **Phase 2 completion**: 80-85% coverage
- **Phase 3 completion**: 85-90% coverage
- **Phase 4 completion**: 90-95% coverage
- **Phase 5 completion**: 95%+ coverage

### Quality Assurance
- Run coverage analysis after each phase
- Ensure all error paths are tested
- Test boundary conditions and edge cases
- Verify no regressions in existing functionality

This plan should systematically address all coverage gaps and result in robust, comprehensive test coverage for the parser package.