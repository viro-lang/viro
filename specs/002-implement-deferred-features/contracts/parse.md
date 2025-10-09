# Contract: Parse Dialect

**Feature**: Deferred Language Capabilities (002)  
**Functional Requirements**: FR-018, FR-019  
**Applies To**: `parse` native and supporting rule words (`set`, `copy`, `into`, `some`, `any`, `opt`, `not`, `ahead`, `fail`, `reject`)

---

## Signature
```
parse input rules --case --part length --trace
```

### Parameters
- `input`: `string!` or `block!`.
- `rules`: `block!` representing parse grammar.
- `--case`: case-sensitive string comparisons (default case-insensitive for ASCII letters).
- `--part length`: limit parsing to first `length` characters/values.
- `--trace`: enable verbose trace events for debugging; writes to trace sink with category `parse`.

### Return
- `logic!` – `true` if entire (or limited) input consumed per grammar, `false` otherwise. When parse fails, optional named `fail` rules can throw with structured Syntax error.

---

## Rule Semantics

### Literals
- `"abc"`, `123`, `'word` – match exact literal (respect `/case`).

### Word Rules
- `word` – fetch rule from caller frame (must be function! returning rule block or value) or treat as set word target for `set`/`copy`.

### `set word rule`
- Evaluates `rule`; on success assigns matched value to `word` in caller frame. Uses deep copy for blocks/strings.

### `copy word rule`
- Like `set` but copies substring/series instead of evaluated value; `word` receives same type as input slice.

### `into block`
- Recursively parses nested block using block as new rule set.

### Quantifiers
- `some rule`: one or more occurrences.
- `any rule`: zero or more occurrences.
- `opt rule`: zero or one occurrence.
- `rule | rule`: alternation (first-match wins).
- `rule rule`: sequence.

### Control
- `not rule`: succeeds if `rule` fails (no input consumed).
- `ahead rule`: lookahead; evaluates `rule` without consuming input.
- `fail`: aborts parse with Syntax error (category 200) using nearest/where context.
- `reject`: aborts parse returning `false` without error.

### Blocks & Parens
- `[ ... ]`: grouping; evaluated lazily.
- `( expr )`: evaluate expression; if returns rule block, splice into rules; if returns `logic!`, treat as condition (false fails rule).

### Named Rules
- `rule: [...]` inside `rules` assigns block to word for reuse (local grammar definition).

---

## Behavior & Algorithm

1. Initialize `ParseState` with input series, index = 1, captures map.
2. Process rules sequentially; each rule returns success flag and new index.
3. On success, continue; on failure, backtrack according to quantifiers.
4. After rules complete, success if index > length (or equals `--part` limit).
5. `--trace` emits events on rule entry/exit with indentation indicating recursion depth.

---

## Error Handling

- Invalid rule type (e.g., `integer!` where rule expected) → Script error (`invalid-parse-rule`).
- Infinite loop detection: track `(rule pointer, index)` combinations; if repeated without consuming input, raise Syntax error (`parse-stalled`).
- `set`/`copy` on protected words (locked frame) → Access error (`protected-word`).
- `--case` only allowed for string input; using with block raises Script error.

---

## Examples

```viro
parse "abc123" [some letter some digit]             ; => true
parse "abc" [copy word some letter end]             ; word: "abc"
parse [1 2 3] [some integer 3]                      ; => true
parse --case "ABC" ["ABC"]                          ; => true (case sensitive)
parse data [some [digit | letter] fail "Bad data"] ; raises error when encountering other chars
```

---

## Testing Expectations

- Contract tests verifying each combinator, capture, and control flow.
- Integration tests linking parse output to object construction and decimal conversion.
- `--trace` outputs validated via trace sink inspection (ensuring start/stop events).

---

## Performance Considerations

- Use iterative approach with explicit stack (no recursion) to avoid Go recursion limits.
- Memoize named rules when possible to reduce re-evaluation overhead.
- Provide benchmark cases for large input (10k tokens) to guard against regressions.
