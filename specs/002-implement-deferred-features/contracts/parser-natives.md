# Contract: Parser Natives (Two-Stage Parser)

**Feature**: Deferred Language Capabilities (002)  
**Functional Requirements**: FR-023 (Metaprogramming Support)  
**Applies To**: `tokenize`, `parse`, `load-string`, `classify`

---

## 1. `tokenize`

### Signature
```
tokenize source
```

### Parameters
- `source`: `string!` - The Viro source code to tokenize.

### Return
- `block!` - A block of token objects.

### Behavior
- Tokenizes a Viro source code string into token objects (Stage 1 of two-stage parser).
- Each token object has fields:
  - `type`: `word!` - One of: `literal`, `string`, `lparen`, `rparen`, `lbracket`, `rbracket`, `eof`
  - `value`: `string!` - The literal text of the token
  - `line`: `integer!` - Line number (1-based)
  - `column`: `integer!` - Column number (1-based)
- Handles whitespace separation (space, tab, newline, comma)
- Handles comments (`;` to end of line)
- Handles string literals with escape sequences
- Returns all tokens including EOF marker

### Error Cases
- Unclosed string → Syntax error (`unterminated-string`)
- Invalid escape sequence → Syntax error (`invalid-escape`)

### Tests
- Empty input → `[eof-token]`
- Single literal → `[literal-token eof-token]`
- Multiple literals with whitespace
- Brackets and parentheses
- String literals with escapes
- Comments ignored
- Position tracking verified

---

## 2. `parse`

### Signature
```
parse tokens
```

### Parameters
- `tokens`: `block!` - A block of token objects (from `tokenize`)

### Return
- `block!` - A block of classified Viro values.

### Behavior
- Parses token objects into Viro values (Stage 2 of two-stage parser).
- Processes structural tokens (`[`, `]`, `(`, `)`) into blocks and parens
- Delegates literal classification to internal `ClassifyLiteral` function
- Builds nested structures recursively
- Returns top-level values (may contain nested blocks/parens)

### Error Cases
- Non-block input → Script error (`type-mismatch`)
- Invalid token object structure → Script error (`invalid-token`)
- Unclosed block → Syntax error (`unclosed-block`)
- Unclosed paren → Syntax error (`unclosed-paren`)
- Unexpected closing bracket → Syntax error (`unexpected-closing`)

### Tests
- Empty token list → `[]`
- Single literal token → `[classified-value]`
- Block tokens → nested block structure
- Paren tokens → paren structure
- Mixed literals and structures
- Error cases verified

---

## 3. `load-string`

### Signature
```
load-string source
```

### Parameters
- `source`: `string!` - The Viro source code to parse.

### Return
- `block!` - A block of parsed Viro values.

### Behavior
- Combines `tokenize` and `parse` in one step (convenience function).
- Equivalent to: `parse tokenize source`
- Returns parsed values ready for evaluation or manipulation.
- This is the primary function for loading Viro code from strings.

### Error Cases
- Non-string input → Script error (`type-mismatch`)
- Any tokenization or parsing error → propagated

### Tests
- Simple literals → classified values
- Blocks and parens → nested structures
- Multiple values → flat block of values
- Empty string → `[]`
- Roundtrip test: `load-string source` should parse correctly

---

## 4. `classify`

### Signature
```
classify literal
```

### Parameters
- `literal`: `string!` - A literal string to classify.

### Return
- `any!` - The classified Viro value (integer!, word!, set-word!, etc.)

### Behavior
- Classifies a single literal string into its appropriate Viro type.
- Uses same logic as parser's `ClassifyLiteral` method.
- Supported classifications:
  - Integers: `"42"` → `42` (integer!)
  - Decimals: `"3.14"` → `3.14` (decimal!)
  - Strings: Already classified by tokenizer
  - Set-words: `"abc:"` → `abc:` (set-word!)
  - Get-words: `":abc"` → `:abc` (get-word!)
  - Lit-words: `"'abc"` → `'abc` (lit-word!)
  - Datatypes: `"integer!"` → `integer!` (datatype!)
  - Paths: `"obj.field"` → path with segments
  - Set-paths: `"obj.x:"` → set-path
  - Get-paths: `":obj.x"` → get-path
  - Words: `"abc"` → `abc` (word!)

### Error Cases
- Non-string input → Script error (`type-mismatch`)
- Invalid literal format → Syntax error (`invalid-literal`)

### Tests
- Each value type verified
- Edge cases (negative numbers, scientific notation, complex paths)
- Invalid formats raise errors

---

## Integration & Usage

### Example: Metaprogramming
```viro
; Tokenize source
tokens: tokenize "x: 42"
; => [token{type: literal, value: "x:"} token{type: literal, value: "42"} token{type: eof}]

; Parse tokens
values: parse tokens
; => [set-word! integer!]

; Direct load
code: load-string "x: 42"
; => [set-word! integer!]

; Classify single literal
classify "abc:"  ; => abc: (set-word!)
classify "42"    ; => 42 (integer!)
```

### Example: Custom Dialect Implementation
```viro
; Build custom parser for SQL-like syntax
sql: load-string "SELECT name FROM users WHERE age > 18"
; Process tokens to implement custom dialect behavior
```

### Example: Code Generation
```viro
; Generate code dynamically
expr: compose [x: (calculate-value) + 10]
code-string: form expr
parsed: load-string code-string
do parsed  ; Execute generated code
```

---

## Performance Considerations

- Tokenization is O(n) where n is source length
- Parsing is O(n) where n is token count
- Token objects allocated once during tokenization
- No backtracking or lookahead required
- Suitable for large code files (tested up to 10k tokens)

---

## Testing Expectations

- Contract tests for each native function (`test/contract/parser_test.go`)
- Unit tests for tokenizer (`internal/tokenize/tokenizer_test.go`)
- Unit tests for parser (`internal/parse/semantic_parser_test.go`)
- Integration tests verifying roundtrip: `load-string source` produces correct values
- Integration test verifying equivalence: `parse tokenize source` = `load-string source`
- Error handling verified for all error cases
- Position tracking verified (line/column numbers)

---

## Dependencies

- `internal/tokenize/tokenizer.go` - Stage 1 tokenizer
- `internal/parse/semantic_parser.go` - Stage 2 parser with `ClassifyLiteral`
- `internal/native/parse.go` - Native function implementations
- `internal/value` - Value constructors and type system

---

## Migration Notes

- These natives expose the two-stage parser implementation
- The old PEG parser is fully replaced
- All existing Viro code continues to work unchanged
- New metaprogramming capabilities enabled through these primitives
- `load` native (if exists) should be aliased or replaced with `load-string`
