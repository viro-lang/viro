# Contract: Decimal Arithmetic & Advanced Math Natives

**Feature**: Deferred Language Capabilities (002)  
**Functional Requirements**: FR-004, FR-005  
**Applies To**: `decimal`, `pow`, `sqrt`, `exp`, `log`, `log-10`, `sin`, `cos`, `tan`, `asin`, `acos`, `atan`, `round`, `ceil`, `floor`, `truncate`

---

## 1. `decimal`

### Signature
```
decimal value
```

### Parameters
- `value`: `integer!`, `decimal!`, `string!`
  - `string!` must represent a valid decimal literal (optional sign, digits, optional decimal point, optional exponent).

### Return
- `decimal!` – value converted to decimal128 using default context (precision 34, rounding half-even).

### Behavior
1. If `value` already `decimal!`, return copy preserving context/scale.
2. If `integer!`, create decimal with magnitude = integer, scale = 0.
3. If `string!`, parse into decimal with scale computed from literal digits.
4. Store context pointer on returned value for downstream operations.

### Type Rules
- Reject `none!`, `logic!`, `word!`, `block!`, `function!`, `object!`, etc.

### Error Cases
- Invalid string literal → Math error (`invalid-decimal literal`).
- Overflow/underflow outside decimal128 bounds → Math error (`out-of-range`).

### Tests
- `decimal 42` ⇒ decimal 42.0 (scale 0).
- `decimal "19.99"` preserves scale 2.
- `decimal "1e4000"` → Math error.

---

## 2. `pow`

### Signature
```
pow base exponent
```

### Parameters
- `base`: `integer!` or `decimal!`
- `exponent`: `integer!` or `decimal!`

### Return
- `decimal!`

### Behavior
1. Promote integers to decimal.
2. Use decimal context to compute `base^exponent`.
3. Precision equals max(base precision, exponent precision) + 10 (capped at 34).
4. Result scale derived from context rounding.

### Error Cases
- Negative base with fractional exponent → Math error (`complex-result-unsupported`).
- Zero^negative → Math error (`division-by-zero`).
- Overflow/underflow beyond decimal128 limits.

### Tests
- `pow 2 10` ⇒ 1024 as decimal.
- `pow 9 0.5` ⇒ 3.
- `pow -2 0.5` ⇒ Math error.

---

## 3. `sqrt`

### Signature
```
sqrt value
```

### Parameters
- `value`: `integer!` or `decimal!`

### Return
- `decimal!`

### Behavior
- Promote to decimal and compute square root with context precision 34.

### Error Cases
- Negative input → Math error (`sqrt-negative`).

---

## 4. `exp`

### Behavior
- Returns decimal e^value with context precision 34.
- Large positive input saturates to overflow error, large negative to underflow (returns subnormal but non-zero if representable).

### Error Cases
- Input magnitude > 6000 → Math error (`exp-overflow`).

---

## 5. `log`

### Behavior
- Natural logarithm of positive decimal.

### Error Cases
- Non-positive input → Math error (`log-domain`).

---

## 6. `log-10`

### Behavior
- Base-10 logarithm via decimal library; same domain/range rules as `log`.

---

## 7. Trigonometric Functions

### Signatures
```
sin value
cos value
tan value
asin value
acos value
atan value
```

### Parameters
- For `sin`, `cos`, `tan`: `integer!` or `decimal!` representing radians.
- For inverse functions: `decimal!` in range [-1, 1].

### Return
- `decimal!`

### Behavior
- Convert inputs to decimal radians.
- Use decimal library trig implementations (Taylor series with context precision).
- Normalize angles to [-π, π] before evaluation to control error.

### Error Cases
- `tan` for odd multiples of π/2 → Math error (`tan-singularity`).
- `asin`/`acos` input outside [-1,1] → Math error (`domain-error`).

---

## 8. Rounding Helpers

### Signatures
```
round value
round --places value places
round --mode value mode
ceil value
floor value
truncate value
```

### Parameters
- `value`: `decimal!` or `integer!`
- `places`: `integer!` ≥ 0 specifying decimal places (default 0).
- `mode`: word! one of `half-up`, `half-even`, `half-down`, `toward-zero`, `toward-infinity`, `toward-neg-infinity`.

### Return
- `decimal!`

### Behavior
1. Promote integers to decimal.
2. `round` without refinements rounds to nearest integer using context rounding mode.
3. `--places` adjusts scale to `places` by setting context precision accordingly.
4. `--mode` temporarily overrides rounding mode for operation only.
5. `ceil`, `floor`, `truncate` map to `--mode` variations.

### Error Cases
- Negative `places` → Script error (`invalid-argument`).
- Unsupported `mode` word! → Script error (`invalid-mode`).

### Tests
- `round 3.45` (default half-even) ⇒ 3.4.
- `round --places 3.456 2` ⇒ 3.46.
- `ceil -1.2` ⇒ -1.
- `truncate 3.99` ⇒ 3.

---

## Shared Validation & Error Semantics

- All math natives require numeric arguments; non-numeric → Script error (`type-mismatch`).
- Overflow/underflow from decimal library yields Math error (400 range) with near/where context.
- Operations return immutable decimal values; assignment uses copy-on-write semantics to avoid shared context mutation.

---

## Testing Expectations

- Contract tests cover standard cases, edge precision, rounding combinations, and domain errors.
- Integration scenarios include currency calculations, trigonometric use in geometry, and parsing + evaluation combos.
- Benchmarks added to `internal/native/math_bench_test.go` to guard against regressions (FR-005).
