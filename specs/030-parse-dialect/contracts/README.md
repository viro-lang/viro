# Parse Dialect Contracts

## Overview

All interpreter work must follow TDD. This bundle enumerates the contract suites that must exist under `test/contract/` before implementing each capability. The suites cover strings, blocks, control flow, mutation, and charsets to guarantee Rebol parity.

## Planned Contract Suites

| File | Focus | Notes |
|------|-------|-------|
| `test/contract/parse_string_basic_test.go` | Literal matches, alternation, quantifiers. | Drives Tasks 4–5. |
| `test/contract/parse_block_semantics_test.go` | Datatype tests, recursion, `into`. | Drives Task 7. |
| `test/contract/parse_capture_test.go` | `copy`, `set`, `word:` marks, paren evaluation. | Drives Task 6. |
| `test/contract/parse_charset_test.go` | `charset`, `--case`, `--all`, `to`/`thru`. | Drives Task 8. |
| `test/contract/parse_control_test.go` | `fail`, `reject`, `accept`, `if`, `while`, `not`. | Drives Task 9. |
| `test/contract/parse_mutation_test.go` | `insert`, `remove`, `change`, `collect`, `keep`. | Drives Task 10. |

## Coverage Matrix

| Scenario | String | Block | Mutation | Control |
|----------|--------|-------|----------|---------|
| Literal + alternation | ✅ | ✅ | n/a | n/a |
| Quantifiers (`any`, `some`, `opt`) | ✅ | ✅ | n/a | ✅ |
| Navigation (`to`, `thru`, `--part`) | ✅ | ✅ | n/a | n/a |
| Captures (`copy`, `set`, `word:`) | ✅ | ✅ | ✅ | ✅ |
| Charsets (`charset`, `--case`) | ✅ | n/a | n/a | n/a |
| Control words | ✅ | ✅ | n/a | ✅ |
| Mutation actions | ✅ | ✅ | ✅ | n/a |

## Exit Criteria

- All contract files compile and fail before implementation, then pass after the corresponding task.
- Each suite documents fixtures and expectations inline (no comments in Go sources; use table-driven tests and descriptive names).
- Integration tests referencing these contracts run via `go test ./test/contract/...` and `go test ./test/integration/...` prior to merging.