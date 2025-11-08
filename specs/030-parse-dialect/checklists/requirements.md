# Requirements Checklist

- [ ] `parse` native registered in `internal/native/register_io.go` with help coverage.
- [ ] Legacy tokenizer helper exported as `parse-values` (or agreed name) with docs.
- [ ] Contract suites exist for strings, blocks, control flow, mutation, and charsets before interpreter changes.
- [ ] Engine implements `copy`, `set`, `word:` marks, and paren evaluation with lexical scoping.
- [ ] `bitset!/charset` value type implemented with `charset` native and documented constructors.
- [ ] Control words map to deterministic engine codes and `verror` entries.
- [ ] Mutation actions (`insert`, `remove`, `change`, `collect`, `keep`) respect copy-on-write and sandbox rules.
- [ ] Quickstart + architecture docs published alongside release notes.