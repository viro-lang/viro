# Two-Stage Parser Architecture

This document describes Viro's two-stage parser design that separates tokenization from semantic parsing.

## Design Philosophy

The parser follows a **strict separation of concerns**:

1. **Tokenizer (Lexical Analysis)**: Language-agnostic, recognizes only syntactic structure
2. **Parser (Semantic Analysis)**: Viro-specific, interprets token meaning and creates Value structures

This separation enables:

- Extensibility: Parser logic can be exposed to Viro code for metaprogramming
- Simplicity: Each stage has a single, focused responsibility
- Dialect support: Same tokenizer, different parsers for different dialects
- Testability: Each stage can be tested independently

## Stage 1: Tokenizer

The tokenizer performs **pure lexical analysis** without any knowledge of Viro semantics.

### Token Types

```go
type TokenType int

const (
    TokenLiteral    TokenType = iota  // Any word-like sequence
    TokenString                       // Quoted string with escapes
    TokenLParen                       // (
    TokenRParen                       // )
    TokenLBracket                     // [
    TokenRBracket                     // ]
    TokenEOF                          // End of input
)

type Token struct {
    Type   TokenType
    Value  string     // Raw text (for Literal/String)
    Line   int        // Line number (for error reporting)
    Column int        // Column number (for error reporting)
}
```

### Tokenization Rules

1. **Whitespace**: Space, tab, newline, carriage return - used to separate tokens
2. **Comments**: `;` to end of line - completely ignored
3. **Strings**: `"..."` with escape sequence support (`\"`, `\\`, `\n`, etc.)
4. **Brackets**: `[`, `]`, `(`, `)` - emitted as single-character tokens
5. **Literals**: Everything else - any sequence of non-whitespace, non-bracket characters

### Examples

```
Input:  abc: 5
Tokens: [Literal("abc:"), Literal("5"), EOF]

Input:  [1 + 2]
Tokens: [LBracket, Literal("1"), Literal("+"), Literal("2"), RBracket, EOF]

Input:  "hello world"
Tokens: [String("hello world"), EOF]

Input:  --flag abc
Tokens: [Literal("--flag"), Literal("abc"), EOF]

Input:  obj.field:
Tokens: [Literal("obj.field:"), EOF]

Input:  ; comment
        abc
Tokens: [Literal("abc"), EOF]
```

### String Handling

Strings are the **only** construct that requires special tokenizer logic:

```
Input:  "say \"hello\""
Token:  String("say \"hello\"")  ; Escaped quotes preserved in value

Input:  "line 1\nline 2"
Token:  String("line 1\nline 2")  ; Escape sequences processed
```

**Escape sequences supported:**

- `\"` - Quote
- `\\` - Backslash
- `\n` - Newline
- `\t` - Tab
- `\r` - Carriage return

### Error Conditions

1. **Unclosed string**: `"hello` → Error: unclosed string literal
2. **Invalid escape**: `"hello\x"` → Error: invalid escape sequence

## Stage 2: Parser

The parser performs **semantic analysis**, transforming tokens into Viro Value structures.

### Parser Responsibilities

1. Classify literals into specific Viro types (integer, set-word, word, etc.)
2. Build structural values (blocks, parens) from bracket pairs
3. Handle paths (dot-separated sequences)
4. Track position for error reporting

### Literal Classification

The parser examines each `TokenLiteral` and classifies it:

```
Literal "5"        → IntegerVal(5)
Literal "3.14"     → DecimalVal(3.14)
Literal "abc"      → WordVal("abc")
Literal "abc:"     → SetWordVal("abc")
Literal ":abc"     → GetWordVal("abc")
Literal "'abc"     → LitWordVal("abc")
Literal "--flag"   → WordVal("--flag")  ; Just a word!
Literal "integer!" → DatatypeVal("integer!")
Literal "obj.x"    → PathVal([Word("obj"), Word("x")])
Literal "obj.x:"   → SetPathVal([Word("obj"), Word("x")])
Literal ":obj.x"   → GetPathVal([Word("obj"), Word("x")])
```

### Classification Algorithm

```
function classifyLiteral(text string) Value:
    ; Literal words (starts with ')
    if text starts with "'":
        return LitWordVal(text[1:])

    ; Get-words and get-paths (starts with :)
    if text starts with ":":
        if text contains ".":
            segments = parsePathSegments(text[1:])
            return GetPathVal(segments)
        else:
            return GetWordVal(text[1:])

    ; Set-words and set-paths (ends with :)
    if text ends with ":":
        base = text[:-1]
        if base contains ".":
            segments = parsePathSegments(base)
            return SetPathVal(segments)
        else:
            return SetWordVal(base)

    ; Paths (contains .)
    if text contains ".":
        segments = parsePathSegments(text)
        return PathVal(segments)

    ; Numbers
    if text matches integer pattern:
        return IntegerVal(parseInt(text))

    if text matches decimal pattern:
        return DecimalVal(parseDecimal(text))

    ; Datatypes (ends with !)
    if text ends with "!":
        return DatatypeVal(text)

    ; Default: regular word (includes refinements like --flag)
    return WordVal(text)
```

### Path Segment Parsing

Paths are dot-separated sequences where each segment can be:

- Word segment: `obj.field`
- Index segment: `block.1` (numeric)

```
function parsePathSegments(text string) []PathSegment:
    parts = split(text, ".")
    segments = []

    for each part in parts:
        if part matches integer pattern:
            segments.append(PathSegment{Type: Index, Value: parseInt(part)})
        else:
            segments.append(PathSegment{Type: Word, Value: part})

    return segments

Examples:
    "obj.field"     → [Word("obj"), Word("field")]
    "data.1"        → [Word("data"), Index(1)]
    "obj.x.y"       → [Word("obj"), Word("x"), Word("y")]
    "matrix.1.2"    → [Word("matrix"), Index(1), Index(2)]
```

### Block and Paren Parsing

Brackets create recursive structures. The key is having **one main function** that handles all value types:

```
; Top-level: parse all values until EOF
function Parse() []Value:
    values = []
    while position < len(tokens) and tokens[position].Type != TokenEOF:
        value = parseValue()
        values.append(value)
    return values

; Core function: parse a single value (THE BIG SWITCH)
function parseValue() Value:
    token = tokens[position]
    position++

    switch token.Type:
        case TokenLiteral:
            return classifyLiteral(token.Value)

        case TokenString:
            return StringVal(token.Value)

        case TokenLBracket:
            values = parseUntil(TokenRBracket, "block")
            return BlockVal(values)

        case TokenLParen:
            values = parseUntil(TokenRParen, "paren")
            return ParenVal(values)

        case TokenRBracket, TokenRParen:
            error("unexpected closing bracket")

        case TokenEOF:
            error("unexpected EOF")

; Generic helper: parse until closing token is found
function parseUntil(closingType TokenType, structName string) []Value:
    values = []

    while position < len(tokens):
        token = tokens[position]

        if token.Type == closingType:
            position++
            return values

        if token.Type == TokenEOF:
            error("unclosed " + structName)

        ; Recursive call to parseValue - no duplication!
        value = parseValue()
        values.append(value)

    error("unclosed " + structName)
```

**Key insight**: `parseValue()` is the **only** function with the big switch. `parseUntil()` is a simple helper that eliminates duplication between block and paren parsing.

### Error Reporting

Both stages track position for precise error messages:

**Tokenizer errors:**

```
Error: Unclosed string literal at line 5, column 10
    "hello world
    ^
```

**Parser errors:**

```
Error: Unclosed block at line 3, column 5
    data: [1 2 3
          ^
```

## Key Differences from PEG Parser

### Current PEG Approach

```
SetWord ← word:WordChars ':' !':' {
    return value.NewSetWordVal(word.(string)), nil
}

Path ← first:WordChars rest:('.' PathElement)+ !':' {
    ; Complex path construction logic...
    return value.PathVal(path), nil
}
```

Problems:

- Grammar mixes tokenization (WordChars) and semantics (SetWord)
- Path logic buried in grammar rules
- Hard to expose to user code
- Requires PEG parser generator dependency

### Two-Stage Approach

```
Tokenizer:
    "abc:" → Literal("abc:")

Parser:
    Literal("abc:") → classifyLiteral("abc:")
                   → ends with ':'
                   → SetWordVal("abc")
```

Benefits:

- Clear separation: tokenizer sees "abc:", parser interprets it
- Parser logic is plain Go code (no grammar DSL)
- Easy to expose classification function to Viro
- No external dependencies

## Extensibility: Viro-Level Access

The two-stage design enables metaprogramming:

### Native Functions for Parsing

```viro
; Tokenize source code
tokens: tokenize "abc: 5"
; Returns: [token! token! token!]

; Parse tokens into values
values: parse tokens
; Returns: [set-word! integer!]

; Convenience wrapper
values: load "abc: 5"
; Returns: [set-word! integer!]

; Custom classification
classify "abc:"    ; → set-word!
classify "123"     ; → integer!
classify "--flag"  ; → word!
```

### Use Cases

1. **Code generation**: Build Viro code as strings, load dynamically
2. **DSLs**: Create domain-specific languages with custom parsing
3. **Macros**: Transform code before evaluation
4. **Analysis**: Inspect code structure without evaluation

### Example: Custom Dialect

```viro
; Parse a custom query dialect
query-code: "SELECT name WHERE age > 18"
tokens: tokenize query-code

; Custom parser for this dialect
parsed: parse-query-dialect tokens
; Returns: [
;   operation: 'select
;   fields: ["name"]
;   condition: [age > 18]
; ]

; Execute query
results: execute-query parsed database
```

## Implementation Strategy

### Phase 1: Tokenizer Implementation

1. Create `internal/tokenize/tokenizer.go`
2. Implement token stream with position tracking
3. Handle strings with escape sequences
4. Write comprehensive tokenizer tests

### Phase 2: Parser Implementation

1. Create `internal/parse/parser.go` (replace PEG)
2. Implement literal classification logic
3. Implement bracket/paren/brace handling
4. Implement path parsing
5. Write comprehensive parser tests

### Phase 3: Integration

1. Update `parse.Parse()` to use tokenizer + parser
2. Remove PEG dependency and generated code
3. Update Makefile to remove grammar generation
4. Verify all existing tests pass

### Phase 4: Expose to Viro

1. Add `tokenize` native function
2. Add `parse` native function
3. Add `load` native function
4. Add `classify` native function
5. Write user-facing tests

## Testing Strategy

### Tokenizer Tests

Test each token type in isolation:

- Literals separated by various whitespace
- Strings with escape sequences
- Brackets and parens
- Comments
- Mixed tokens
- Error cases (unclosed strings)

### Parser Tests

Test classification logic:

- Integer literals (positive, negative, zero)
- Decimal literals (with/without exponent)
- Set-words, get-words, lit-words
- Paths (word paths, index paths, mixed)
- Set-paths, get-paths
- Datatypes
- Regular words (including refinements)

Test structure building:

- Empty blocks
- Nested blocks
- Parens with immediate values
- Mixed structures
- Error cases (unclosed brackets)

### Integration Tests

Use existing test suite to verify:

- All contract tests pass
- REPL functionality unchanged
- Error messages maintain quality

## Performance Considerations

The two-stage parser should be **at least as fast** as the PEG parser:

1. **Tokenizer**: Single linear pass, no backtracking
2. **Parser**: Single linear pass over tokens
3. **Total**: Two linear passes vs PEG's recursive descent with backtracking

Expected improvement: 10-30% faster parsing due to elimination of PEG overhead.

## Migration Path

1. Implement new parser in parallel (no disruption)
2. Add feature flag to switch between parsers
3. Verify new parser with full test suite
4. Default to new parser
5. Remove PEG code after stability period

## Future Enhancements

Once the two-stage parser is stable:

1. **Token metadata**: Add source ranges for better error messages
2. **Binary literals**: Add `#{FF00}` support in tokenizer
3. **Unicode escapes**: Support `\u{...}` in strings
4. **Custom literals**: Allow user-defined literal types
5. **Syntax extensions**: Enable user-defined syntax via parser hooks

## Conclusion

The two-stage parser design:

- Simplifies the codebase (removes PEG dependency)
- Enables metaprogramming (exposes parsing to user code)
- Improves maintainability (clear separation of concerns)
- Supports future extensibility (custom dialects, macros)

This aligns perfectly with Viro's homoiconic design philosophy where code is data and the boundaries between language and user code blur.
