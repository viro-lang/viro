# Contract: Mutation & Collection

## Scope

Ensures parse actions that mutate or collect data behave deterministically and respect copy-on-write.

## Test Cases

1. **`change`**: Replace every `foo` with `bar` and confirm other references retain original data.
2. **`insert`/`remove`**: Insert separators during parsing and remove matched segments, verifying cursor advancement.
3. **`collect`/`keep`**: Build result blocks from matched tokens, ensuring nested collects merge properly.
4. **`set` + mutation**: Use `set word word!` followed by `insert` using captured values.
5. **Immutable input**: Attempt to mutate string literals or read-only binaries and assert Access error (500).

## Fixtures

- Strings with repeated tokens, blocks representing template ASTs, shared-series scenarios to test copy-on-write.

## File Mapping

Implemented in `test/contract/parse_mutation_test.go` and referenced by higher-level integration tests.