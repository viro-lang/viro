# Feature Specification: Deferred Language Capabilities

**Feature Branch**: `002-implement-deferred-features`  
**Created**: 202- **- **FR-014**: System MUST report parse failures via Syntax error (200) including near/where context and the rule that failed.
- **FR-015**: System MUST provide tracing controls `trace --on`, `trace --off`, and `trace?`, emitting structured events (timestamp, value, word, duration) through a configurable sink that defaults to a rotating log file in the current working directory.
- **FR-016**: System MUST expose debugger commands via a `debug` native (subcommands `breakpoint`, `remove`, `step`, `continue`, `stack`, `locals`) that can be invoked interactively or programmatically.
- **FR-017**: System MUST add reflection natives `type-of`, `spec-of`, `body-of`, `words-of`, and `values-of` for introspecting functions, objects, and frames without mutating them.
- **FR-018**: System MUST extend error handling so that file and network failures attach relevant metadata (path, host) to the error context.2**: System MUST extend series natives with `copy`, `copy --part`, `find`, `find --last`, `remove`, `remove --part`, `skip`, `take`, `sort`, and `reverse`, all operating on blocks and strings with refinement parity.
- **FR-013**: System MUST implement the `parse` dialect supporting core combinators (`some`, `any`, `opt`, `not`, `into`, `ahead`, `set`, `copy`) and custom rule blocks, returning boolean success and capture data.
- **FR-014**: System MUST report parse failures via Syntax error (200) including near/where context and the rule that failed.0-08  
**Status**: Draft  
**Input**: User description: "Implement deferred features that were not included in the first version. Review documentation in specs/001-implement-the-core to identify deferred features."

## Clarifications

### Session 2025-10-08

- Q: What is the TLS certificate verification policy for HTTP ports? → A: Allow user to disable verification per request via CLI flag; otherwise enforce by default.
- Q: How do users configure the sandbox root for file operations? → A: Single CLI parameter at startup sets sandbox root; fallback to current working directory when omitted.
- Q: Where do `trace --on` events write by default? → A: Default sink is **stderr** for immediate visibility. Users may redirect to log file via `--trace-file` CLI flag.
- Q: Jaki jest domyślny timeout operacji sieciowych portów? → A: Brak wbudowanego limitu – poleganie na limitach systemowych lub jawnej konfiguracji.

### Session 2025-10-09

- Q: Jaki jest maksymalny limit głębokości zagnieżdżania obiektów i ścieżek (np. `a.b.c.d.e...`)? → A: Brak limitu (do wyczerpania stosu)
- Q: Czy HTTP porty automatycznie podążają za przekierowaniami (HTTP 301/302), czy zwracają odpowiedź redirect do skryptu? → A: Automatycznie podążaj (max 10 redirects)
- Q: Jak system obsługuje próbę otwarcia tego samego portu pliku przez dwa równoczesne skrypty? → A: Brak lockingu - zależy od OS
- Q: Jakie jest zachowanie HTTP portu gdy odpowiedź ma Content-Type inny niż application/json (np. text/html, application/xml)? → A: Zwróć raw string w body, status/headers bez zmian
- Q: Jaki jest maksymalny rozmiar pliku trace log przed rotacją (lumberjack)? → A: 50 MB per file

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Precise Calculations with Decimal Numbers (Priority: P1)

Power users can express financial and scientific calculations that depend on decimal precision and extended math functions without leaving the REPL.

**Why this priority**: Lack of decimal arithmetic blocks real-world scenarios (billing, rates, measurements). Delivering precise math first unlocks immediate business value while staying close to the original interpreter roadmap.

**Independent Test**: Start the REPL, evaluate expressions such as `rate: 12.5%`, `total: 19.99 * 3`, `round --places total 2`, `pow 1.05 12`, and confirm results within tolerance across integer–decimal mixes and advanced math functions.

**Acceptance Scenarios**:

1. **Given** the REPL is running, **When** a user enters `subtotal: 19.99 * 3`, **Then** the system promotes operands to decimal, computes `59.97`, and preserves $\pm 1 \times 10^{-12}$ precision.
2. **Given** decimal literals with exponent notation, **When** a user evaluates `1.2e-3 + 0.0008`, **Then** the system returns `0.002` without rounding drift.
3. **Given** advanced math functions are available, **When** a user enters `pow 1.05 12`, **Then** the system returns `1.79585632602213` and tagging the result as decimal.
4. **Given** rounding helpers accept precision hints, **When** a user evaluates `round --places 123.4567 2`, **Then** the system returns `123.46`.
5. **Given** invalid combinations should fail fast, **When** a user calls `sqrt -4`, **Then** the system raises Math error (400) describing domain restrictions and leaves the REPL usable.

---

### User Story 2 - Persist and Exchange Data via Ports (Priority: P2)

Script authors can read/write files and interact with remote endpoints through a unified port abstraction, enabling automation workflows directly from Viro.

**Why this priority**: Without file and basic network I/O the interpreter cannot automate real tasks. Ports introduce persistent storage and remote access while reusing deferred file/network capabilities.

**Independent Test**: From a clean workspace, run scripts that create configuration files, append logs, fetch JSON via HTTP GET, and stream data through an open port; verify data integrity and error handling without crashing the REPL.

**Acceptance Scenarios**:

1. **Given** write permissions to `/tmp`, **When** a user runs `save %/tmp/report.txt data`, **Then** the system creates the file, writes block contents as text, and returns the saved path.
2. **Given** a file exists, **When** a user evaluates `read %/tmp/report.txt`, **Then** the system returns the file contents as string, preserving newline characters.
3. **Given** an HTTP endpoint responding with JSON, **When** a user evaluates `http --get https://example.test/data`, **Then** the system returns block data with parsed headers and body.
4. **Given** ports support streaming, **When** a user executes `port: open tcp://localhost:4000` followed by `write port "PING"` and `read port`, **Then** the system transmits bytes, receives a response, and leaves the port open until `close port`.
5. **Given** restricted paths, **When** a user attempts `remove %/etc/passwd`, **Then** the system raises Access error (500) explaining the prohibition and leaves existing files untouched.

---

### User Story 3 - Structure Data with Objects and Paths (Priority: P3)

Developers can create object-like contexts and navigate nested data using path expressions to keep large programs maintainable.

**Why this priority**: Large scripts require structured data. Objects were explicitly deferred; introducing them now enables better data organization without global namespace pollution.

**Independent Test**: Instantiate objects and access nested fields (`user.address.city`). Verify that path assignment updates underlying values.

**Acceptance Scenarios**:

1. **Given** the object constructor, **When** a user evaluates `person: object [name: "Ana" address: object [city: "Porto" zip: 4000]]`, **Then** the system produces nested frames with isolated bindings.
2. **Given** path evaluation semantics, **When** a user reads `person.address.city`, **Then** the system returns `"Porto"` by traversing nested contexts.
3. **Given** path assignment rules, **When** a user executes `person.address.city: "Lisboa"`, **Then** the stored value updates and subsequent reads reflect the change.
4. **Given** nested object access, **When** a user creates deeply nested structures, **Then** the system maintains proper frame isolation and parent chain traversal without imposing an artificial nesting depth limit (constrained only by available stack space).

---

### User Story 4 - Transform Data with Advanced Series and Parse Dialect (Priority: P4)

Data engineers can reshape complex strings and blocks using high-level series utilities and a declarative `parse` dialect that follows left-to-right evaluation lineage expectations.

**Why this priority**: Powerful data manipulation is essential for configuration, ETL, and DSL scenarios. Parse and additional series functions were deferred; implementing them next unlocks high-leverage workflows.

**Independent Test**: Use `parse` to validate structured input, leverage `copy --part`, `find`, `take`, `remove`, and `sort` across blocks/strings, and confirm deterministic results for success and failure cases.

**Acceptance Scenarios**:

1. **Given** a CSV-like string, **When** a user runs `parse data [some digit "," some digit]`, **Then** the system returns `true` when the pattern matches and exposes captured pieces via rules.
2. **Given** `find` supports refinement logic, **When** a user evaluates `find --last [1 2 3 2 1] 2`, **Then** the system returns the series positioned at the last occurrence.
3. **Given** `copy --part` handles strings, **When** a user executes `copy --part "abcdef" 3`, **Then** the system returns `"abc"` without mutating the source.
4. **Given** parse failures provide diagnostics, **When** a user runs `parse "abc" [some digit]`, **Then** the system raises Syntax error (200) with index of first mismatch.
5. **Given** nested rules, **When** a user executes `parse blocks [some [into rule]]` on nested blocks, **Then** recursion terminates with correct boolean result and stack safety.

---

### User Story 5 - Observe, Debug, and Reflect on Programs (Priority: P5)

Maintainers can trace evaluation, inspect runtime state, and introspect program structure to diagnose issues and understand program behavior.

**Why this priority**: Productionizing the language requires observability and introspection. These features were explicitly deferred, but they close the feedback loop for diagnosing issues and understanding program execution.

**Independent Test**: Run scripts while enabling tracing, set breakpoints, inspect stack frames, and use reflection to examine functions and objects. Confirm visibility and stability.

**Acceptance Scenarios**:

1. **Given** tracing controls, **When** a user executes `trace --on` followed by `square 5`, **Then** the system emits structured events for each evaluation step and `trace --off` stops emission.
2. **Given** breakpoint support, **When** a user sets `debug --breakpoint 'square`, **Then** invoking `square` enters an interactive prompt with options `step`, `continue`, and `stack`.
3. **Given** reflection helpers, **When** a user evaluates `spec-of :square`, **Then** the system returns the function's parameter block and refinements.
4. **Given** tracing overhead guardrails, **When** tracing remains inactive, **Then** steady-state evaluation performance matches baseline within 5%.

### Edge Cases

- Decimal arithmetic with extremely large exponents (e.g., $10^{\pm 308}$) must clamp or raise Math error (400) before overflow.
- Mixed integer/decimal division by zero must continue to raise Math error without introducing `NaN` to the value space.
- File I/O operations encountering permission-denied or non-existent paths must raise Access error (500) without leaking partial data.
- Network ports must offer optional timeout configuration; when unset the interpreter relies on operating system limits and must surface hangs via interrupt guidance.
- Parse dialect must guard against infinite loops in user-defined rules (e.g., `some []`) through iteration caps or detection.
- Debugging and tracing must redact sensitive data flagged as protected (e.g., secrets) before emitting logs.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST introduce a `decimal!` value type supporting **exactly 34 decimal digits of precision** (IEEE 754 decimal128 semantics) and preserving scale metadata for round-trip formatting. Operations that would exceed 34-digit precision MUST raise Math error (400) with code `decimal-overflow`.
- **FR-002**: System MUST parse decimal literals with optional sign, fractional component, and exponent (`1.23e-4`) and store them as `decimal!` values.
- **FR-003**: System MUST promote integer operands to `decimal!` during mixed arithmetic and deliver results with **round-to-nearest-even (bankers' rounding) applied automatically to intermediate calculations**. Final result precision follows the operand with greater scale. Explicit `round`, `ceil`, `floor`, `truncate` natives override automatic rounding behavior via their refinements.
- **FR-004**: System MUST provide math natives `pow`, `sqrt`, `exp`, `log`, `log-10`, `sin`, `cos`, `tan`, `asin`, `acos`, and `atan` accepting `decimal!` and `integer!` values with domain validation.
- **FR-005**: System MUST expose rounding natives `round`, `ceil`, `floor`, and `truncate`, each accepting refinements to control precision (number of decimal places) and rounding mode (`half-up`, `half-even`).
- **FR-006**: System MUST provide file natives `read`, `save`, `append`, `remove`, `rename`, and `make-dir` that operate on absolute or sandboxed relative paths with UTF-8 filenames, where the sandbox root is supplied via a startup CLI parameter and defaults to the current working directory if unspecified. File locking behavior defers to operating system defaults; concurrent access from multiple scripts follows OS-level file locking policies.
- **FR-007**: System MUST implement a unified `port!` abstraction with `open`, `close`, `read`, `write`, and `query` operations that works for files, TCP sockets, and HTTP resources, exposing optional timeout refinements while defaulting to operating-system behavior when unset.
- **FR-008**: System MUST support HTTP convenience natives `http --get`, `http --post`, and `http --head`, returning structured response blocks containing status, headers, and body. HTTP client automatically follows redirects (301, 302, 303, 307, 308) up to a maximum of 10 hops before raising a network error. Response body is returned as a raw string regardless of Content-Type; parsing (e.g., JSON) is the caller's responsibility.
- **FR-009**: System MUST introduce `object!` construction that captures word/value pairs into a dedicated frame supporting nested objects.
- **FR-010**: System MUST evaluate path expressions (`user.address.city`, `array.3`) across objects, blocks, and future maps, using dot notation (`.`) for both field access and series indexing.
- **FR-011**: System MUST allow path assignment to mutate terminal targets when permissible (e.g., words, object fields, series elements) while preventing structural violations (e.g., assigning into immutable data). Parser distinguishes tokens by first character(s): numbers start with digit or `-`+digit (`19.99`, `-3.14`), refinements start with `--` (`--option`), words/paths start with letter (`config.timeout`).
- **FR-012**: System MUST extend series natives with `copy`, `copy --part`, `find`, `find --last`, `remove`, `remove --part`, `skip`, `take`, `sort`, and `reverse`, all operating on blocks and strings with refinement parity.
- **FR-013**: System MUST implement the `parse` dialect supporting core combinators (`some`, `any`, `opt`, `not`, `into`, `ahead`, `set`, `copy`) and custom rule blocks, returning boolean success and capture data.
- **FR-014**: System MUST report parse failures via Syntax error (200) including near/where context and the rule that failed.
- **FR-015**: System MUST provide tracing controls `trace --on`, `trace --off`, and `trace?`, emitting structured events (timestamp, value, word, duration) through a configurable sink that defaults to **stderr** for immediate observability. Users may redirect trace output to a rotating log file via the `--trace-file` CLI flag (rotation at 50 MB per file, retaining up to 5 backup files with compression). Trace format is line-delimited JSON for machine parsing.
- **FR-016**: System MUST expose debugger commands via a `debug` native (subcommands `breakpoint`, `remove`, `step`, `continue`, `stack`, `locals`) that can be invoked interactively or programmatically.
- **FR-017**: System MUST add reflection natives `type-of`, `spec-of`, `body-of`, `words-of`, and `values-of` for introspecting functions, objects, and frames without mutating them.
- **FR-018**: System MUST extend error handling so that file and network failures attach relevant metadata (path, host) to the error context.
- **FR-019**: System MUST preserve backwards compatibility: scripts created for Feature 001 continue to run without modification when none of the new capabilities are used.
- **FR-020**: System MUST enforce TLS certificate validation for HTTPS ports by default and allow users to disable verification via two mechanisms: (1) Global CLI flag `--allow-insecure-tls` applies to all HTTP operations in the session, (2) Per-port refinement `--insecure` passed to `open` or `http` natives overrides default for that specific connection. Both mechanisms MUST emit warning to stderr when used.

### Key Entities *(include if feature involves data)*

- **DecimalValue (`decimal!`)**: Represents high-precision numeric values with mantissa, exponent, and rounding context metadata.
- **Port**: Abstraction over files, sockets, and HTTP connections storing scheme, state (`open`, `closed`), buffers, and capability flags.
- **ObjectInstance (`object!`)**: Structured frame capturing word/value pairs with optional parent prototype, supporting path traversal.
- **PathExpression**: Sequence of steps (word, index, refinement) evaluated against nested data, supporting both read and write semantics.
- **ParsePattern**: Declarative rule graph describing pattern combinators, capture targets, and failure hints.
- **TraceSession**: Runtime container for trace configuration (filters, sinks), buffered events, and performance impact tracking.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Decimal arithmetic maintains absolute error below $1 \times 10^{-12}$ for operations within $\pm 10^{12}$ and executes single-operation expressions in under 2 milliseconds on reference hardware (Apple M1/M2).
- **SC-002**: File read/write throughput achieves at least 50 MB/s for sequential operations on local SSDs, and HTTP GET requests complete within 2 seconds for 95% of calls to a LAN endpoint.
- **SC-003**: Parse dialect validates a corpus of 50 representative patterns with 0 false positives and 0 false negatives, completing each parse under 250 milliseconds for inputs up to 1 MB.
- **SC-004**: Debug tracing overhead remains below 5% CPU impact when disabled and below 25% when enabled with default sampling, while breakpoint interaction latency stays under 150 milliseconds.
- **SC-005**: Backward compatibility regression suite (all Feature 001 contract and integration tests) continues to pass with zero changes.

## Assumptions

1. Decimal numbers will use IEEE 754 decimal128 semantics implemented with a high-precision arithmetic core optimized for interactive workloads.
2. File system access remains sandboxed to a root directory defined by a startup CLI parameter, defaulting to the current working directory when the parameter is absent to avoid accidental system writes.
3. HTTP support targets HTTPS over TCP with optional proxy configuration; other protocols (FTP, WebSocket) remain out of scope for this phase.
4. Parse dialect will prioritize compatibility with established syntax patterns where practical; unsupported constructs (e.g., `thru`, `reject`) may be deferred to later phases.
