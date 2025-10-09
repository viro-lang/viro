# Native Function & Protocol Contracts: Deferred Capabilities

**Feature**: Deferred Language Capabilities (002)  
**Date**: 2025-10-08  
**Purpose**: Define contract specifications for new native functions, dialects, and protocols introduced in Phase 002.

---

## Contract Organization

1. **math-decimal.md** – High-precision decimal arithmetic and advanced math natives
   - `decimal` constructor and promotion rules
   - `pow`, `sqrt`, `exp`, `log`, `log-10`
   - Trigonometric natives (`sin`, `cos`, `tan`, `asin`, `acos`, `atan`)
   - Rounding helpers (`round`, `ceil`, `floor`, `truncate`)

2. **ports.md** – Unified port protocol for filesystem, TCP, and HTTP
   - `open`, `close`, `read`, `write`, `query`, `wait`
   - Sandbox resolution, TLS toggles, timeout handling

3. **objects.md** – Object construction and path semantics
   - `object`, `context`, `path` evaluation, mutation rules

4. **parse.md** – Parse dialect contracts
   - Grammar for rules, copy/set semantics, failure reporting

5. **trace-debug.md** – Observability primitives
   - `trace --on`, `trace --off`, `trace?`
   - `debug` command set
   - Trace session structure and filtering

6. **reflection.md** – Reflection and introspection natives
   - `type-of`, `spec-of`, `body-of`, `words-of`, `values-of`

Each contract adheres to the same structure used in Phase 001: signature, parameters, return, behavior, type rules, examples, error cases, and test coverage expectations.

---

## Common Principles

- **Consistency with Core**: New natives integrate with existing evaluator dispatch, stack, and frame systems without bypassing index-based safety.
- **Secure Defaults**: Sandbox, TLS verification, and whitelist checks default to safe behavior; opt-outs require explicit refinements.
- **Observability**: Trace and debug contracts ensure all new operations emit meaningful events for diagnostics.
- **Spec Alignment**: Contracts reference Functional Requirements FR-004 through FR-023 and associated clarification decisions.

---

## Reading Order

1. Start with `math-decimal.md` to understand numeric foundations and promotion rules.
2. Review `objects.md` to learn namespace and path mechanics.
3. Study `ports.md` for I/O surfaces and sandbox guarantees.
4. Consult `parse.md` to grasp dialect semantics before implementing evaluator hooks.
5. Finish with `trace-debug.md` and `reflection.md` for operational tooling.

---

## Related Artifacts

- `data-model.md`: Entity structures supporting these contracts.
- `research.md`: Phase 0 investigations informing the contracts.
- `plan.md`: Implementation sequencing anchored by these contracts.
- `quickstart.md`: User-facing examples demonstrating contract-compliant behavior.

Use this README as a map—each category file should be treated as a contract appendix to the core language specification.
