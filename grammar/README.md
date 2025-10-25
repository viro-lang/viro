# Viro PEG Grammar

This directory contains the PEG (Parsing Expression Grammar) definition for the Viro programming language, implemented using the [Pigeon parser generator](https://github.com/mna/pigeon).

## Overview

The Viro parser is a **single-stage parser** that directly converts source text into flat sequences of `core.Value` objects. Unlike traditional two-stage parsers (tokenizer → parser), the PEG grammar handles both lexical and syntactic analysis in one pass.

### Key Design Principles

1. **Parse as-is, don't transform**: The parser reads syntax literally and produces flat value sequences. The evaluator handles all semantic interpretation and transformation.

2. **Flat sequences only**: The parser never builds nested structures (except for blocks/parens which are themselves value types). For example:
   - Input: `3 + 4 * 2`
   - Output: `[IntVal(3), WordVal("+"), IntVal(4), WordVal("*"), IntVal(2)]`

3. **No semantic interpretation**: The parser doesn't know that `+` is an operator or that `true` is a boolean. It just produces the appropriate value types based on syntax.

## Grammar File Structure

The `viro.peg` file consists of:

1. **Initializer block** (lines 1-57): Go code including imports and helper functions
2. **Grammar rules** (lines 59+): PEG parsing rules with semantic actions

### Top-Level Rule

```peg
Input <- _ vals:Value* _ EOF {
    return flattenValues(vals), nil
}
```

Parses zero or more values, ignoring whitespace and comments, and returns them as a flat slice.

### Value Types

The grammar recognizes these Viro value types:

- **Literals**:
  - `Integer`: `-?[0-9]+` (no decimal point or exponent)
  - `Decimal`: Numbers with `.` or scientific notation (`1.5`, `2e10`, `3.14e-2`)
  - `String`: `"text"`

- **Word Variants**:
  - `Word`: `identifier` (bare words like `print`, `true`, `+`)
  - `SetWord`: `identifier:` (assignment target)
  - `GetWord`: `:identifier` (value retrieval)
  - `LitWord`: `'identifier` (literal/quoted word)

- **Compound Types**:
  - `Path`: `word.segment.segment` where first segment must be a word, following segments can be words or integers
  - `Block`: `[value1 value2 ...]` (flat sequence in block container)
  - `Paren`: `(value1 value2 ...)` (flat sequence in paren container)
  - `Datatype`: `word!` (type literals like `integer!`, `string!`)

### Special Syntax

- **Comments**: `;` to end of line, consumed by whitespace rule
- **Paths**: First element must be a word; subsequent elements can be words or integers
  - Valid: `user.name`, `data.0`, `obj.field.subfield`
  - Invalid: `1.field`, `42.name` (can't start with number)

### Precedence and Choice

PEG ordered choice (`/`) tries alternatives left-to-right:

```peg
Value <- _ val:(Decimal / Integer / String / Block / Paren / 
                SetWord / GetWord / LitWord / Path / 
                Datatype / Word) _ {
    return val, nil
}
```

**Order matters**: `Decimal` must come before `Integer` because integers are a subset of decimal syntax. `SetWord`, `GetWord`, `LitWord`, and `Path` must come before `Word` because they're word prefixes/variants.

## Building the Parser

### Prerequisites

Install Pigeon:
```bash
go install github.com/mna/pigeon@latest
```

### Generate Parser

```bash
# Using Makefile (recommended)
make grammar

# Or directly
pigeon -o internal/parse/peg/parser.go grammar/viro.peg
```

The generated parser is placed in `internal/parse/peg/parser.go` and should be committed to version control.

### Build Process Integration

The Makefile includes grammar generation as part of the build:

```makefile
.PHONY: grammar
grammar:
	pigeon -o internal/parse/peg/parser.go grammar/viro.peg

.PHONY: build
build: grammar
	go build -o viro ./cmd/viro
```

## Modifying the Grammar

### Adding New Value Types

1. Add the value type to the `Value` rule choice
2. Define the grammar rule for the new type
3. Implement semantic action to construct the appropriate `core.Value`
4. Update tests

Example: Adding a new literal type

```peg
Value <- _ val:(NewType / Decimal / Integer / ...) _ {
    return val, nil
}

NewType <- "@" [a-z]+ {
    return value.NewTypeVal(string(c.text[1:])), nil
}
```

### Semantic Actions

Semantic actions are Go code blocks that construct values:

```peg
Integer <- '-'? [0-9]+ !DecimalStart {
    return value.IntVal(toInt(string(c.text))), nil
}
```

- `c.text` contains the matched text as `[]byte`
- Return `(core.Value, error)` or `(any, error)`
- Helper functions in initializer block handle conversions

### Testing Grammar Changes

After modifying the grammar:

1. Regenerate: `make grammar`
2. Run tests: `go test ./internal/parse/...`
3. Run integration tests: `go test ./test/integration/...`
4. Verify contract tests: `go test ./test/contract/...`

## Parser API

The parser is accessed through wrapper functions in `internal/parse/parse.go`:

```go
// Parse converts source text to flat value sequence
func Parse(input string) ([]core.Value, error)

// ParseEval is an alias for Parse (maintained for compatibility)
func ParseEval(input string) ([]core.Value, error)
```

Error handling wraps Pigeon errors in Viro's `verror.SyntaxError` type with source context.

## Helper Functions

The initializer block provides these helpers:

- `toInt(s string) int64`: Convert string to integer
- `toDecimal(s string) (*decimal.Big, int16)`: Convert string to decimal with scale
- `calculateScale(s string) int16`: Calculate decimal scale from string representation
- `buildPath(segments []value.PathSegment) core.Value`: Construct path value
- `toSlice(v any) []any`: Type assertion helper
- `flattenValues(v any) []core.Value`: Flatten parsed value arrays

## Examples

### Input → Output Mapping

```
"hello"          → [StrVal("hello")]
42               → [IntVal(42)]
3.14             → [DecimalVal(3.14, scale=2)]
x: 10            → [SetWordVal("x"), IntVal(10)]
:x               → [GetWordVal("x")]
'literal         → [LitWordVal("literal")]
[1 2 3]          → [BlockVal([IntVal(1), IntVal(2), IntVal(3)])]
user.name        → [PathVal(segments=[Word("user"), Word("name")])]
data.0           → [PathVal(segments=[Word("data"), Index(0)])]
print 2 + 3      → [WordVal("print"), IntVal(2), WordVal("+"), IntVal(3)]
```

Note: All outputs are flat sequences. Blocks and parens are value containers, but their elements are also flat.

## Further Reading

- [Pigeon Documentation](https://pkg.go.dev/github.com/mna/pigeon)
- [PEG Basics](https://en.wikipedia.org/wiki/Parsing_expression_grammar)
- [Viro Language Specs](../specs/)
- [Parser Migration Plan](../plans/004_peg_parser.md)
