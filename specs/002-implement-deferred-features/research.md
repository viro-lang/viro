# Phase 0: Research & Technical Decisions

**Feature**: Deferred Language Capabilities (002)  
**Date**: 2025-10-08  
**Purpose**: Resolve outstanding technical unknowns before design and implementation planning.

## Research Tasks

### 1. Decimal Arithmetic Engine for `decimal!`

**Question**: Which Go library best provides IEEE 754 decimal128 precision with manageable performance for interactive REPL workloads?

**Options Evaluated**:
- **ericlagergren/decimal**: IEEE 754 compliant decimal floating-point supporting decimal32/64/128 with configurable contexts.
- **shopspring/decimal**: Arbitrary precision decimal using big.Int; popular but slower for frequent allocations.
- **cockroachdb/apd**: Arbitrary precision with PostgreSQL compatibility; heavier API surface.
- **Implement from scratch**: Highest effort, high risk of numerical bugs.

**Decision**: Adopt **github.com/ericlagergren/decimal**.

**Rationale**:
- Direct support for IEEE 754 decimal128 contexts, matching assumption #1.
- Optimised for high-throughput financial workloads; benchmarks show lower allocation counts vs shopspring.
- Provides rounding modes (half-even, half-up) and math library (pow, exp, trig) required by FR-004/FR-005.
- Thread-safe contexts allow per-evaluator configuration for precision/rounding refinements.
- Actively maintained, permissive BSD-3 license.

**Integration Notes**:
- Wrap `decimal.Big` inside `DecimalValue` to hide external API.
- Pre-create shared contexts for 34-digit precision and reuse to avoid allocations.
- Convert from/to string via context-aware `FMA`/`String` with scale preservation metadata stored alongside value.

**Alternatives Rejected**:
- `shopspring/decimal`: simpler API but lacks trig/log functions and is slower under GC pressure.
- `cockroachdb/apd`: excellent correctness guarantees but unnecessary complexity for REPL scope.
- Custom implementation: unnecessary risk; community-vetted library preferred.

---

### 2. TLS Verification Toggle for HTTP Ports

**Question**: How to enforce certificate validation by default while allowing per-request override controlled via CLI refinement?

**Options Evaluated**:
- Custom `http.Client` with `tls.Config{InsecureSkipVerify: true}` when requested.
- Global environment variable toggling TLS behaviour.
- Separate native for insecure requests.

**Decision**: Use dedicated HTTP client pool keyed by (`verifyTLS` flag, timeout).

**Rationale**:
- Default client uses strict verification. Optional refinement `--insecure` on `http --get|--post|--head` triggers reuse of client with `InsecureSkipVerify`.
- Per-request CLI refinement (e.g., `http --get url --insecure`) satisfies FR-023 without global side-effects.
- Client pool avoids repeated transport construction while respecting different timeout combinations.

**Implementation Notes**:
- Maintain map `map[clientKey]*http.Client` guarded by mutex.
- Each key encodes `verifyTLS bool`, `timeout time.Duration` (optional refinement), and potential proxy settings.
- Document security risks when `--insecure` used; emit warning to trace log.

**Alternatives Rejected**:
- Environment variable: lacks per-request granularity.
- Separate native: duplicates logic and confuses API.

---

### 3. Sandbox Root Configuration

**Question**: How should the interpreter enforce filesystem sandboxing with CLI-provided root fallback to CWD?

**Decision**: Accept optional CLI flag `--sandbox-root <path>` at startup.

**Rationale**:
- Aligns with clarification (A + fallback). When absent, default to `os.Getwd()`.
- Store canonical absolute root path; all file natives resolve user-provided paths relative to root.
- Enforce sandbox by cleaning paths (`filepath.Clean`), joining with root, and verifying the resulting absolute path has the root prefix (`strings.HasPrefix` using path separators) before performing operations.
- Provide helper `resolveSandboxPath(rel string, allowEscape bool)` to share logic across file natives.

**Security Considerations**:
- Deny attempts to escape via `..`, symlinks looping outside root (use `filepath.EvalSymlinks` on final path).
- For absolute paths supplied by users, ensure they still fall under sandbox root; otherwise raise Access error (500).

**Alternatives Rejected**:
- Hard-coded CWD: inflexible for embedding or automation scenarios.
- Runtime mutator native: complicates guarantees about file scope.

---

### 4. Unified `port!` Abstraction & Timeout Handling

**Question**: What structure supports file, TCP, and HTTP ports with optional timeouts defaulting to OS behaviour?

**Decision**: Define `Port` struct with pluggable driver interface.

**Design**:
```go
type PortDriver interface {
    Open(ctx context.Context, spec PortSpec) error
    Read([]byte) (int, error)
    Write([]byte) (int, error)
    Close() error
    Query() PortInfo
}

type Port struct {
    Scheme string
    Driver PortDriver
    Timeout *time.Duration // nil => OS defaults
}
```
- `PortSpec` encapsulates target (path/URL/address) and mode flags.
- Timeout refinements (`--timeout 5`) set `Port.Timeout`; absence leaves nil so drivers rely on OS/dialer defaults (per clarification C).
- Implement drivers: `fileDriver`, `tcpDriver`, `httpDriver` each using idiomatic Go packages (`os`, `net`, `net/http`).

**Rationale**:
- Keeps API uniform across schemes while respecting optional timeout requirement.
- Context-based open operations allow cancellation via REPL Ctrl+C.
- Query returns metadata (size, remote addr, headers) as block for script-level introspection.

**Alternatives Rejected**:
- Single switch statement for all operations: harder to extend for future protocols (UDP, serial).
- Mandatory timeout: contradicts clarification.

---

### 5. Trace Logging Sink and Rotation

**Question**: How to persist trace events to a log file without uncontrolled growth?

**Decision**: Use `gopkg.in/natefinch/lumberjack.v2` for log rotation.

**Rationale**:
- Provides size-based rotation with compression and retention limits; widely adopted in Go ecosystem.
- Minimal dependency, pure Go, no CGO.
- Default configuration: `viro-trace.log`, max 10 MB per file, keep 5 backups, compress older logs.
- Integrate with `log.Logger` dedicated to tracing; allow refinement `--trace-sink` to override path or choose stdout.

**Operational Notes**:
- Trace events structured as JSON lines for easy post-processing.
- When user disables tracing, logger flushes and closes file.

**Alternatives Rejected**:
- Rolling our own rotation: error-prone.
- Always writing to stdout: conflicts with requirement for default file sink.
- Syslog integration: portability concerns on Windows.

---

### 5. Trace Logging Sink and Rotation

## Summary of Resolved Clarifications

| Item | Resolution |
|------|------------|
| Decimal precision & math library | Use `github.com/ericlagergren/decimal` with shared decimal128 contexts |
| TLS verification override | Per-request `--insecure` refinement selects client with `InsecureSkipVerify` |
| Sandbox root configuration | Startup flag `--sandbox-root`; default to CWD with canonical path enforcement |
| Port abstraction & timeout | Driver-based `Port` with optional timeout refinements, OS defaults otherwise |
| Trace sink defaults | Lumberjack-rotated log file `viro-trace.log` in working directory |

All technical uncertainties required for Phase 1 artifacts have been resolved.