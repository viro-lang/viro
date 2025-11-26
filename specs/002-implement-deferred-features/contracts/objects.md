# Contract: Objects & Path Semantics

**Feature**: Deferred Language Capabilities (002)  
**Functional Requirements**: FR-014, FR-015, FR-016, FR-017  
**Applies To**: `object`, `context`, path evaluation, assignment, `select`, `put`

**Path Notation**: This contract uses unified dot notation (`.`) for all path traversal: object field access (`object.field`), series indexing (`block.3`, `array.(index)`), and nested paths (`data.items.2.name`). The parser distinguishes tokens by examining the first character(s): numbers start with a digit or `-` followed immediately by a digit (e.g., `19.99`, `-3.14`), refinements start with `--` (e.g., `--option`), and words/paths start with a letter (e.g., `config.timeout`, `items.5`).

---

## 1. `object`

### Signature
```
object spec
```

### Parameters
- `spec`: `block!` describing fields and optional initial values.
  - Syntax for object creation `make object!`:
    - `word` for field declaration (initialized to `none`).
    - `word: value` for explicit initialization (evaluated immediately in new object context).
    - `word: [block]` retains block (no evaluation) when preceded by `'` (quote) in spec.

### Return
- `object!` (ObjectInstance) referencing new frame with declared fields bound.

### Behavior
1. Create new Frame with parent = caller frame (for access to outer bindings during initialization).
2. Iterate spec block:
   - Collect words; ensure unique within object.
   - For `word: value`, evaluate `value` in object frame, assign result.
   - For `set-word` whose value is block preceded by `quote`, store unevaluated block.
3. Freeze manifest: record field names and optional type hints (future extension).

### Error Cases
- Duplicate word in spec → Script error (`object-field-duplicate`).
- Initialization expression throws error → propagate.

### Tests
- Basic object creation, nested objects, quoted blocks.

---

## 2. `context`

### Signature
```
context spec
```

### Behavior
- Alias for `object spec` but discards parent frame (isolated scope).

### Error Cases
- Same as `object`.

---

## 3. Path Evaluation

### Supported Forms
- `object.field`
- `object.sub.path`
- `block.3`
- `block.(index)` (paren evaluation)
- `word/:field` (path through set-word referencing object)

### Behavior
1. Evaluate first element (word/path value) to produce base value.
2. For each subsequent segment:
    - If base is `object!` and segment `word!`, lookup field; if missing, check parent objects recursively; if still missing, return `none`.
    - If base `block!` or `string!` and segment `integer!`, perform 1-based index access; out-of-range → return `none`.
    - If segment is paren, evaluate expression each time.
3. Prior to final resolution, capture penultimate base and segment metadata for assignment operations.
4. For `object.field: value`, perform type validation if hint exists, update frame slot in place.

### Mutation Rules
- Only final segment assignable.
- Attempt to assign into literal value (e.g., `1.2: 3`) → Script error (`immutable-target`).
- When base is `block!`, assignment updates underlying slice (copy-on-write if block shared across frames per constitution principle).

### Error Cases
- Path evaluation encountering `none!` mid-chain → Script error (`none-path`).
- Path applied to unsupported type → Script error (`path-type-mismatch`).

### Tests
- Access nested object fields, with parent prototypes.
- Mutation through path updates underlying frame.
- Block indexing and set-path operations behave correctly.

---

## 4. `select`

### Signature
```
select series selector --default value
```

### Behavior
- For `object!`: treat as association list; returns field value or `--default` when missing.
- For `block!`: returns value when matching key (alternating key/value pattern).
- For `path!`: evaluate path to retrieve value.

### Error Cases
- Non-object/block/path → Script error.

---

## 5. `put`

### Signature
```
put target key value
```

### Parameters
- `target`: `object!` or `block!` to modify
- `key`: field name (for objects) or key (for blocks)
- `value`: value to assign

### Behavior
- **For objects**: Sets or updates object field after validation.
- **For blocks**: Treats block as association list of alternating key/value pairs:
  - Searches from current block index (respecting cursor position)
  - Updates existing key/value pair if key found
  - Appends new key/value pair if key not found (unless value is `none`)
  - Removes key/value pair if value is `none`
  - Handles odd-length blocks gracefully
  - Keys matched using same logic as `select` (word-like symbol equality or general `Equals`)
- Returns the assigned value (or `none` for removal)

### Error Cases
- Invalid target type → Script error (`put target`)
- Invalid field/key type for objects → Script error (`put field`)

---

## Observability

- Trace events: `object-create`, `object-field-read`, `object-field-write` with metadata (object id, field name).
- Debugger `locals` command lists object fields in current frame.

---

## Testing Expectations

- Contract tests for object creation, parent lookup, path assignment, select/put behavior.
- Integration tests verifying user stories (nested path mutation, parse output into objects).
