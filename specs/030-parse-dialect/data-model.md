# Data Model: Parse Dialect

## SeriesCursor

| Field | Description |
|-------|-------------|
| `series value.Series` | Underlying string or block value. |
| `index int` | Current 1-based position. |
| `end int` | Length of the series. |
| `marks []cursorMark` | Stack of saved positions for backtracking and `word:` marks. |

Responsibilities:
- Provide `Advance(n)`, `Rewind(pos)`, `Slice(from,to)` helpers.
- Enforce copy-on-write when mutation actions target shared series.
- Normalize behavior across `string!` and `block!` via interface methods (`ElementAt`, `DatatypeAt`).

## RuleNode

Represents a normalized rule element consisting of:
- `Kind` (Literal, Datatype, WordInvoke, Group, Action, Control, Charset).
- `Value` (literal value, datatype symbol, referenced word, child block).
- `Options` (min/max repeats, case sensitivity, refinements).

Rule blocks are lazily interpreted; we only materialize nodes when necessary (e.g., to precompute charset bitsets).

## MatchState & BacktrackFrame

- `MatchState` tracks current cursor, parent frame, bound words, captures, and refinement flags.
- `BacktrackFrame` stores {cursor index, rule pointer, repeat counters, capture snapshot}. Frames push on entry to a repeating rule and pop when success/failure resolves.

## CaptureContext

- `Words map[string]value.Value` – values assigned via `set` or `copy`.
- `Marks map[string]int` – indices recorded via `word:` syntax.
- `Collector []value.Value` – values appended via `collect`/`keep`.
- Provides `Snapshot()` / `Restore()` to support backtracking.

## Charset / Bitset Value

- Stored as a dedicated `bitset!` type containing a bitmap of Unicode scalar ranges.
- Built via `charset` native that consumes rule blocks (`["aeiou" not digit]`).
- Supports `Union`, `Intersect`, `Contains(rune)` APIs reused by `to`, `thru`, `any`, `some`.

## Native Boundary

`parse` native signature (conceptual):
```
parse input rules --case --all --any --part limit
```
- Validates argument types, wraps input in `SeriesCursor`, configures refinements, and delegates to `dialect.Engine`. On failure returns `false` or raises Syntax error with `near`/`where` metadata.

## Diagnostics

- Extend `verror` with IDs: `parse-invalid-rule`, `parse-control-flow`, `parse-depth-limit`, `parse-immutable-target`.
- Error payload includes `rule`, `index`, `near`, `where`, `message` fields for CLI display and tooling.