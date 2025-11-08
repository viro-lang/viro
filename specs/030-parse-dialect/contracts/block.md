# Contract: Block Parsing

## Scope

Validates datatype tests, literal word matching, recursion, `into`, `set`, and `copy` semantics when operating on `block!` series.

## Test Cases

1. **Datatype guards**: `parse data [some integer!]`, `parse data [word! word! integer!]`.
2. **Literal words and quoting**: `'print matches literal word, `quote` prevents evaluation.
3. **Recursive rules**: nested rule words referencing themselves, ensuring recursion terminates and depth guard errors when exceeded.
4. **`into` semantics**: `parse data [some [into [word! integer!]]]` verifying nested block parsing.
5. **Captures**: `set word word!` binds references, `copy block into target` duplicates nested blocks without aliasing.
6. **Control flow**: `reject`/`accept` from inside nested rules, `if`/`while` evaluating parens to drive rule branching.

## Fixtures

- Blocks that mimic ASTs, configuration DSLs, and nested path expressions.

## File Mapping

Implemented in `test/contract/parse_block_semantics_test.go` with helper fixtures in `test/scripts/` if necessary.