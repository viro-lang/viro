# Contract: String Parsing

## Scope

Covers literal matching, alternation, repetition, navigation, charsets, captures, and whitespace/case refinements when parsing `string!` values.

## Test Cases

1. **Literal sequences**: `parse "abc" ["a" "b" "c"]` succeeds; mismatch returns `false` with error pointing at failing index.
2. **Alternation**: `parse data [["GET" | "POST"] space some letter]` toggles between alternatives.
3. **Quantifiers**: `parse digits [3 5 digit]`, `parse digits [some digit]`, `parse digits [opt sign some digit]`.
4. **Navigation**: `parse value [to "@" copy local skip copy domain to end]`, verifying `--all` vs default whitespace handling.
5. **Charsets**: `digit: charset ["0-9"]`, `letter: charset ["a-z" | "A-Z"]`, `parse` rules using `--case` and `not`.
6. **Captures**: `copy` into words, `word:` marks capturing cursor locations, paren evaluation returning booleans.
7. **Error paths**: invalid rule shapes, empty charsets, integer repeat bounds with `min > max` raising `parse-invalid-rule`.

## Fixtures

- Strings: email addresses, HTML snippets, CSV lines, query strings.
- Charsets: digits, letters, whitespace, punctuation.

## File Mapping

Implement in `test/contract/parse_string_basic_test.go` and `test/contract/parse_charset_test.go` with table-driven cases.