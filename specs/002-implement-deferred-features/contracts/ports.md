# Contract: Ports & External I/O

**Feature**: Deferred Language Capabilities (002)  
**Functional Requirements**: FR-006, FR-007, FR-008, FR-009, FR-010  
**Applies To**: `open`, `close`, `read`, `write`, `save`, `load`, `query`, `wait`

---

## 1. `open`

### Signature
```
open target --binary --lines --seek --timeout duration --insecure
```

### Parameters
- `target`: `file!`, `url!`, `tcp://host:port` (represented as `string!` or `url!` value), or port spec block.
- `--binary`: return binary! content for reads (default text).
- `--lines`: treat text read/write as line-delimited.
- `--seek`: allow relative positioning; unsupported for HTTP.
- `--timeout duration`: optional `integer!` milliseconds or `decimal!` seconds to override OS default.
- `--insecure`: only valid for HTTPS targets; disables certificate verification (FR-007 clarification #1).

### Return
- `port!` handle encapsulating driver, spec, and timeout settings.

### Behavior
1. Resolve sandbox root for file targets, ensuring normalized path remains within root (FR-006, clarif #2).
2. Parse target into `PortSpec` (scheme, path, mode, options).
3. Select driver:
   - File driver: standard library file operations.
   - TCP driver: `net.Dial` with optional TLS upgrade.
   - HTTP driver: builds client with shared connection pool; reuses TLS state keyed by flag combos.
4. Apply timeout if provided using `context.WithTimeout` or `net.Dialer.Timeout`.
5. Store `--insecure` flag only for HTTPS; raise Script error otherwise.

### Error Cases
- Sandbox violation → Access error (500) with `attempted-path` metadata.
- Unsupported scheme → Script error (300) `unknown-port-scheme`.
- TLS handshake failure when `--insecure` not set → Access error `tls-verification-failed`.

### Tests
- Open file within sandbox succeeds; outside fails.
- Open HTTPS with invalid cert → fails unless `--insecure` used.
- Open TCP with custom timeout respects duration.

---

## 2. `close`

### Signature
```
close port
```

### Behavior
- Invokes driver `Close()`, sets state to `PortClosed`.
- Idempotent: closing already closed port returns `none!`.

### Error Cases
- Non-port argument → Script error.

---

## 3. `read`

### Signature
```
read source --binary --lines --part length
```

### Parameters
- `source`: `file!`, `url!`, `port!`.
- `--binary`: force binary return.
- `--lines`: return block of lines.
- `--part length`: `integer!` specifying bytes/chars.

### Behavior
1. If `source` literal path or url, opens temporary port, reads, closes.
2. Respect sandbox and TLS rules as in `open`.
3. For HTTP, default method GET; reuse shared client.
4. `--part` for HTTP uses `Range` header; for file uses `io.Reader` limit.
5. `--lines` splits on `\n` and returns `block!` of `string!`.

### Error Cases
- Timeout while reading → IO error (600) `timeout` with duration metadata.
- `--lines` used with binary data → Script error.

---

## 4. `write`

### Signature
```
write target data --append --binary --lines --timeout duration --insecure
```

### Parameters
- `target`: same as `open`.
- `data`: `string!`, `binary!`, `block!` (when `--lines`).

### Behavior
1. For file targets, create directories as needed under sandbox root.
2. `--append` appends to file; disallowed for HTTP/TCP.
3. `--lines` expects block of strings; joins with `\n`.
4. HTTP defaults to PUT; `binary` writes bytes.

### Error Cases
- Attempt to `--append` HTTP → Script error.
- Sandbox escape attempts → Access error.

---

## 5. `save` & `load`

Convenience wrappers around `write`/`read` with serialization:

- `save %file value`: serializes value using loadable format (block notation) and writes text file.
- `load %file`: reads file and parses block into Viro values.
- Honor sandbox, binary, and timeout rules from underlying operations.

### Error Cases
- Parse failure during `load` → Syntax error (200).
- Value containing unsupported types (future) → Script error.

---

## 6. `query`

### Signature
```
query port
```

### Behavior
- Returns object! with metadata depending on scheme:
  - File: size, modified time, permissions.
  - HTTP: status, headers, content-length.
  - TCP: local/remote addresses, state.

### Error Cases
- Port closed → Access error (`port-closed`).

---

## 7. `wait`

### Signature
```
wait port-or-block
```

### Behavior
- For single port: blocks until readable/writable/timeout using driver-specific polling.
- For block of ports: returns first ready port; supports `none!` for timeout.
- Honors per-port timeout; if none set, uses OS defaults.

### Error Cases
- Mixed types inside block → Script error.

---

## Security & Sandbox Guarantees

- Sandbox root configured via CLI flag `--sandbox-root` (clarification #2).
- All file paths normalized with `filepath.Clean` and validated using `strings.HasPrefix(resolved, root)` after resolving symlinks.
- HTTP requests restricted to schemes `http` or `https`; other schemes rejected.
- `--insecure` limited to HTTPS; trace event emitted when used.

---

## Observability

- Trace events: `port-open`, `port-read`, `port-write`, `port-close`, `port-error` with metadata (scheme, path, bytes).
- Debugger `ports` command enumerates open ports, states, timeouts.

---

## Testing Expectations

- Contract tests for sandbox, TLS, timeout, append rules.
- Integration scenarios: SC-004 (HTTP fetch), SC-005 (file import/export), SC-006 (sanboxed denial).
- Benchmarks measure `read`/`write` throughput with various buffer sizes.
