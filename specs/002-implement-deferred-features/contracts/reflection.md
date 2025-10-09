# Contract: Reflection & Introspection

**Feature**: Deferred Language Capabilities (002)  
**Functional Requirements**: FR-022  
**Applies To**: `type-of`, `spec-of`, `body-of`, `words-of`, `values-of`, `source`

---

## 1. `type-of`

### Signature
```
type-of value
```

### Parameters
- `value`: any Viro value.

### Return
- `word!` naming the type (e.g., `integer!`, `decimal!`, `object!`, `port!`).

### Behavior
- Maps underlying `ValueType` to canonical word name.
- For path results (transient), returns type of resolved value.
- For errors, returns `error!`.

### Tests
- Validate each concrete type mapping, including new ones from Phase 002.

---

## 2. `spec-of`

### Signature
```
spec-of value
```

### Parameters
- `value`: `function!`, `native!`, `object!`.

### Return
- `block!` copy of specification:
  - Function: argument spec block (deep copy).
  - Native: generated spec block describing parameters and refinements.
  - Object: block of field definitions (word/value pairs with initializers where applicable).

### Behavior
- Produces immutable copy: modifications to returned block do not affect source entity.

### Error Cases
- Unsupported type → Script error (`spec-unsupported-type`).

---

## 3. `body-of`

### Signature
```
body-of value
```

### Parameters
- `value`: `function!`, `object!`.

### Return
- `block!` copy of body or initialization block.
  - Function: block representing function body (deep copy).
  - Object: block used during creation (if retained), otherwise synthesized from manifest.

### Behavior
- Returns deep copy to prevent mutation of source entity.

### Error Cases
- Value without accessible body (native) → Script error (`no-body`).

---

## 4. `words-of`

### Signature
```
words-of value
```

### Return
- `block!` of words representing public bindings:
  - Object: field names.
  - Frame (future) or context: words in frame.

### Error Cases
- Non-object/frame → Script error.

---

## 5. `values-of`

### Signature
```
values-of value
```

### Return
- `block!` of values corresponding to `words-of` order; copies produced for mutable types to avoid direct frame access.

### Error Cases
- Align with `words-of`.

---

## 6. `source`

### Signature
```
source value
```

### Behavior
- Returns formatted string representation of value suitable for `do`.
- For functions, pretty-prints spec + body.
- Honors new types: decimals printed preserving scale, objects shown with field assignments.

### Error Cases
- Value with no source representation → Script error (`source-unsupported`).

---

## Observability

- Trace events: `reflection` category with metadata (action, target-type, cost).
- For large results (>1 MB), trace event flagged `large=true` and operation truncated unless `/full` refinement (future) supplied.

---

## Testing Expectations

- Contract tests verifying type coverage, immutability of returned blocks, cross-consistency (`words-of` + `values-of`).
- Integration tests ensure reflection interacts with objects/ports as expected.
