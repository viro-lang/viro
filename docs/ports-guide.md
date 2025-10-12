# Ports Guide: Secure I/O and Sandbox Configuration

**Viro Interpreter** - File and network I/O with security controls

---

## Overview

Ports provide a unified interface for external I/O operations:

- **File I/O**: Read and write files within a sandboxed directory
- **HTTP/HTTPS**: Fetch and send data over the web with TLS verification
- **TCP**: Low-level socket communication with timeout controls

All port operations enforce security policies to prevent unauthorized access:
- **Sandbox enforcement** for file operations
- **TLS verification** for HTTPS (with opt-out for development)
- **Timeout controls** to prevent hanging operations

---

## Quick Start

### Reading Files

```viro
; Start Viro with sandbox root
./viro --sandbox-root ~/my-project

; Read file (relative to sandbox root)
content: read %data/input.txt
print content
```

### Writing Files

```viro
; Write data to file (within sandbox)
write %output/results.txt "Processing complete"

; Append to file
write --append %logs/activity.log "Operation succeeded"
```

### HTTP Requests

```viro
; Fetch data from web
response: read https://api.example.com/data
print response

; POST data
write https://api.example.com/submit [status: "ok" count: 42]
```

---

## Sandbox Configuration

### What is the Sandbox?

The sandbox is a **security boundary** that restricts file operations to a specific directory tree. This prevents:
- Reading sensitive files (e.g., `/etc/passwd`, `~/.ssh/keys`)
- Writing to system directories
- Traversing outside the project directory

### Configuring Sandbox Root

**Command-line flag**:
```bash
./viro --sandbox-root /Users/alex/my-project
```

**Default behavior** (if flag not provided):
- Sandbox root = current working directory
- Example: If you run `./viro` from `/Users/alex/project`, the sandbox root is `/Users/alex/project`

**Absolute paths**:
```bash
./viro --sandbox-root /var/viro-data
```

**Relative paths** (resolved relative to current directory):
```bash
./viro --sandbox-root ./data
```

**Home directory expansion**:
```bash
./viro --sandbox-root ~/viro-workspace
```

### Sandbox Resolution Rules

1. **All file paths are resolved relative to sandbox root**

   ```viro
   ; Sandbox root: /Users/alex/project
   read %data/input.txt
   ; Resolves to: /Users/alex/project/data/input.txt
   ```

2. **Parent directory traversal is validated**

   ```viro
   ; Sandbox root: /Users/alex/project
   read %../../../etc/passwd
   ; ERROR: Access error - sandbox-violation
   ```

3. **Symlink resolution is enforced**

   ```viro
   ; Sandbox root: /Users/alex/project
   ; If project/data/link points to /etc/passwd
   read %data/link
   ; ERROR: Access error - sandbox-violation
   ; (filepath.EvalSymlinks detects escape attempt)
   ```

4. **Absolute paths are rejected**

   ```viro
   read %/etc/passwd
   ; ERROR: Access error - paths must be relative
   ```

### Sandbox Security Guarantees

✅ **Guaranteed safe**:
- Reading files under sandbox root
- Writing files under sandbox root
- Creating subdirectories
- Following symlinks that stay within sandbox

❌ **Blocked attempts**:
- Accessing files outside sandbox (via `..` or absolute paths)
- Following symlinks to external locations
- Reading system files
- Writing to protected directories

### Sandbox Violation Errors

When a sandbox violation occurs:

```viro
>> read %../../../etc/passwd
** Access Error: sandbox-violation
Where: (read %../../../etc/passwd)
Metadata: attempted-path="/Users/alex/etc/passwd", sandbox-root="/Users/alex/project"
```

**Error details**:
- `sandbox-violation` error ID
- Attempted path (resolved)
- Configured sandbox root

---

## Port Operations

### Opening Ports

**File port**:
```viro
port: open %data/config.txt
content: read port
close port
```

**HTTP port**:
```viro
port: open https://api.example.com
response: read port
close port
```

**TCP port**:
```viro
port: open tcp://localhost:8080
write port "GET /status HTTP/1.0\r\n\r\n"
response: read port
close port
```

### Port Refinements

#### `--timeout duration`

Set maximum wait time for operations (milliseconds):

```viro
; Wait up to 5 seconds (5000ms)
port: open --timeout 5000 https://slow-api.example.com
```

**Default behavior** (no timeout specified):
- Uses OS default timeouts
- HTTP: Typically 30-60 seconds
- TCP: OS-dependent
- File: No timeout (local I/O assumed fast)

**Timeout errors**:
```viro
>> port: open --timeout 100 https://very-slow-server.com
** IO Error: timeout
Where: (open --timeout 100 https://very-slow-server.com)
Metadata: duration=100ms
```

#### `--binary`

Return data as binary instead of text:

```viro
; Read image file as binary
image-data: read --binary %images/logo.png
print type-of image-data  ; ==> binary!
```

#### `--lines`

Read or write line-delimited text:

```viro
; Read file as block of lines
lines: read --lines %data/log.txt
print length? lines       ; Number of lines
print first lines         ; First line

; Write block of lines
write --lines %output.txt ["Line 1" "Line 2" "Line 3"]
```

#### `--append`

Append to file instead of overwriting:

```viro
write %log.txt "First entry"
write --append %log.txt "Second entry"
; File contains both entries
```

**File-only refinement** (not valid for HTTP/TCP):
```viro
>> write --append https://example.com "data"
** Script Error: --append not supported for HTTP
```

#### `--insecure` (HTTPS only)

**⚠️ Security Warning**: Disables TLS certificate verification

**Use cases**:
- Development with self-signed certificates
- Testing against local HTTPS servers
- Temporary workaround for certificate issues

**Not recommended for production**

```viro
; Development/testing only
port: open --insecure https://self-signed.local
response: read port
close port
```

**Logged event**:
When `--insecure` is used, a warning is written to stderr:
```
WARNING: TLS verification disabled for https://self-signed.local
```

**Invalid usage**:
```viro
; Error: --insecure only valid for HTTPS
>> open --insecure http://example.com
** Script Error: --insecure only valid for HTTPS targets
```

### Reading Data

**Read entire file**:
```viro
content: read %data/input.txt
```

**Read partial data** (first N bytes):
```viro
header: read --part %large-file.bin 1024
; Reads first 1024 bytes
```

**Read from port**:
```viro
port: open %data.txt
chunk: read --part port 512
close port
```

### Writing Data

**Write string**:
```viro
write %output.txt "Hello, world!"
```

**Write binary data**:
```viro
write --binary %image.png binary-data
```

**Write lines**:
```viro
write --lines %report.txt ["Header" "Data 1" "Data 2" "Footer"]
```

**Append to file**:
```viro
write --append %log.txt "New log entry"
```

### Convenience Wrappers

#### `save` - Serialize and Write

```viro
data: [id: 42 name: "Product" price: decimal "19.99"]
save %data/product.viro data
```

**Behavior**:
- Serializes Viro values to loadable text format
- Writes to file (enforces sandbox)

#### `load` - Read and Parse

```viro
data: load %data/product.viro
print data.name  ; Access deserialized data
```

**Behavior**:
- Reads file as text
- Parses into Viro values
- Returns block or single value

**Error handling**:
```viro
>> data: load %malformed.viro
** Syntax Error: unclosed-block
Where: (load %malformed.viro)
```

### Querying Port Metadata

```viro
port: open %data/file.txt
info: query port
print info
; ==> object! with fields: size, modified, permissions
close port
```

**HTTP example**:
```viro
port: open https://api.example.com/data
info: query port
print info
; ==> object! with fields: status, headers, content-length
```

**TCP example**:
```viro
port: open tcp://localhost:8080
info: query port
print info
; ==> object! with fields: local-address, remote-address, state
```

### Waiting for Port Readiness

**Wait for single port**:
```viro
port: open tcp://localhost:9000
wait port
; Blocks until port is readable/writable
```

**Wait for multiple ports**:
```viro
port1: open tcp://server1:8080
port2: open tcp://server2:8080
ready: wait [port1 port2]
print ready  ; First port that becomes ready
```

### Closing Ports

**Explicit close**:
```viro
port: open %data.txt
content: read port
close port
```

**Idempotent** (safe to call multiple times):
```viro
close port
close port  ; No error, returns none
```

**Auto-close on REPL exit**:
- Ports are automatically closed when REPL session ends
- Recommended to close explicitly for resource management

---

## HTTP/HTTPS Details

### TLS Verification (Default: Enabled)

**Secure by default**:
```viro
; Requires valid TLS certificate
response: read https://api.example.com
```

**Certificate validation checks**:
- Certificate chain trusted by system CA store
- Hostname matches certificate CN or SAN
- Certificate not expired

**Development override**:
```viro
; Disable verification (logged to stderr)
response: read --insecure https://self-signed.local
```

### HTTP Methods

**GET** (default for `read`):
```viro
response: read https://api.example.com/data
```

**POST** (default for `write` with data):
```viro
write https://api.example.com/submit [key: "value"]
```

**PUT** (explicit):
```viro
; PUT request (implementation may vary)
write https://api.example.com/resource data
```

**HEAD** (query only):
```viro
port: open https://example.com/file.zip
info: query port  ; HEAD request to get metadata
close port
```

### Redirect Following

**Automatic** (up to 10 redirects):
```viro
; Follows 301, 302, 303, 307, 308 redirects
response: read https://short.link/abc
; Automatically follows to final destination
```

**Redirect limit**:
- Maximum: 10 hops
- Prevents infinite redirect loops

**Redirect error**:
```viro
>> read https://infinite-redirect.example.com
** IO Error: too-many-redirects
Where: (read https://infinite-redirect.example.com)
Metadata: redirect-count=10
```

### HTTP Client Pooling

**Internal optimization**: HTTP clients are pooled based on TLS and timeout settings

**Key**: `(verifyTLS, timeout)`

**Example**:
```viro
; These share the same HTTP client (same TLS/timeout)
port1: open --timeout 5000 https://api.example.com
port2: open --timeout 5000 https://api.example.com

; This uses a different client (different TLS verification)
port3: open --timeout 5000 --insecure https://self-signed.local
```

**Benefits**:
- Connection reuse (faster subsequent requests)
- Reduced memory overhead

---

## TCP Details

### Connection Establishment

```viro
; Connect to TCP server
port: open tcp://localhost:8080

; With timeout
port: open --timeout 3000 tcp://remote-server:9000
```

**Timeout applies to**:
- Connection establishment
- Read operations
- Write operations

### Reading and Writing

**Write text**:
```viro
write port "PING\r\n"
```

**Read response**:
```viro
response: read port
```

**Binary communication**:
```viro
write --binary port binary-data
response: read --binary port
```

### Connection State

```viro
info: query port
print info.state  ; open, closed, eof
```

### Error Handling

**Connection refused**:
```viro
>> port: open tcp://localhost:9999
** IO Error: connection-refused
Where: (open tcp://localhost:9999)
```

**Timeout**:
```viro
>> port: open --timeout 100 tcp://very-slow-server:8080
** IO Error: timeout
Where: (open --timeout 100 tcp://very-slow-server:8080)
```

---

## File I/O Details

### Path Resolution

**Relative paths** (recommended):
```viro
; Sandbox root: /Users/alex/project
read %data/input.txt
; Resolves to: /Users/alex/project/data/input.txt
```

**Subdirectories created automatically**:
```viro
write %reports/2025/q1/summary.txt data
; Creates reports/, reports/2025/, reports/2025/q1/ if they don't exist
```

**File not found**:
```viro
>> read %nonexistent.txt
** IO Error: file-not-found
Where: (read %nonexistent.txt)
```

### Binary vs Text Mode

**Text mode** (default):
- Reads as UTF-8 string
- Line ending normalization (OS-dependent)

**Binary mode**:
- Reads raw bytes
- No encoding conversion
- Preserves exact file contents

```viro
; Text
text: read %data.txt
print type-of text  ; ==> string!

; Binary
data: read --binary %image.png
print type-of data  ; ==> binary!
```

---

## Security Best Practices

### Sandbox Configuration

**Do**:
- Set explicit `--sandbox-root` for production scripts
- Use minimal privilege (smallest necessary sandbox)
- Test sandbox boundaries during development
- Document sandbox requirements in project README

**Don't**:
- Use `/` as sandbox root (grants access to entire filesystem)
- Disable sandbox (not supported; sandbox is always enforced)
- Assume default sandbox location (it's CWD, which varies)

### TLS Verification

**Do**:
- Keep TLS verification enabled for production
- Use `--insecure` only for development/testing
- Add CA certificates to system store for internal CAs
- Monitor logs for `--insecure` usage

**Don't**:
- Use `--insecure` for production HTTPS
- Ignore TLS errors (investigate certificate issues)
- Disable verification without logging/audit trail

### Timeout Configuration

**Do**:
- Set explicit timeouts for network operations
- Use conservative timeouts (prevent hanging)
- Handle timeout errors gracefully
- Log timeout occurrences for monitoring

**Don't**:
- Use infinite timeouts (omitting `--timeout` uses OS defaults, which may be long)
- Set timeouts too short (causes spurious failures)
- Ignore timeout errors

### Error Handling

**Do**:
```viro
; Check for errors before using data
if error? [data: read %config.txt] [
    print "Failed to read config, using defaults"
    data: default-config
]
```

**Don't**:
```viro
; Assume operations always succeed
data: read %config.txt  ; May fail!
print data.field        ; Error if read failed
```

---

## Troubleshooting

### Sandbox Violations

**Symptom**: `Access error: sandbox-violation`

**Causes**:
1. Path traversal outside sandbox (e.g., `../../etc/passwd`)
2. Absolute path used (e.g., `/etc/passwd`)
3. Symlink points outside sandbox

**Solutions**:
- Use relative paths within sandbox
- Check symlink targets
- Verify sandbox root configuration

### TLS Verification Failures

**Symptom**: `Access error: tls-verification-failed`

**Causes**:
1. Self-signed certificate
2. Expired certificate
3. Hostname mismatch
4. Untrusted CA

**Solutions**:
- For development: Use `--insecure` (temporarily)
- For production: Fix certificate issues or add CA to system store
- Verify hostname matches certificate

### Timeout Errors

**Symptom**: `IO error: timeout`

**Causes**:
1. Server too slow
2. Network issues
3. Timeout set too short

**Solutions**:
- Increase timeout (`--timeout 10000` for 10 seconds)
- Check network connectivity
- Verify server is responding

### File Not Found

**Symptom**: `IO error: file-not-found`

**Causes**:
1. File doesn't exist
2. Wrong path (check sandbox root)
3. Typo in filename

**Solutions**:
- Verify file exists: `ls` in sandbox root
- Check path relative to sandbox root
- Verify filename spelling (case-sensitive)

---

## Examples

### Configuration File Loading

```viro
; Load config from sandbox
config: load %config/settings.viro
print config.database-url
print config.api-key
```

### HTTP API Client

```viro
; Fetch data from API
users: read https://api.example.com/users
print length? users

; POST new user
write https://api.example.com/users [
    name: "Alice"
    email: "alice@example.com"
]
```

### Log File Management

```viro
; Append log entries
write --append %logs/activity.log "User logged in"
write --append %logs/activity.log "Document saved"

; Read all logs
logs: read --lines %logs/activity.log
print length? logs
```

### TCP Echo Server Test

```viro
; Connect to echo server
port: open tcp://localhost:7777
write port "Hello, server!\n"
response: read port
print response
close port
```

### Data Import/Export

```viro
; Export data
data: [
    [id: 1 name: "Product A" price: decimal "19.99"]
    [id: 2 name: "Product B" price: decimal "29.99"]
]
save %exports/products.viro data

; Import data
imported: load %exports/products.viro
print first imported  ; First product
```

---

## Performance Considerations

### File I/O

- **Throughput target**: 50 MB/s (integration test SC-012)
- **Buffer size**: Optimize for workload (use `--part` for large files)
- **Caching**: File system cache improves repeated reads

### HTTP I/O

- **Latency target**: 95% of requests <2s on LAN (SC-012)
- **Connection pooling**: Automatic for same TLS/timeout settings
- **Compression**: Not yet supported (future enhancement)

### TCP I/O

- **Latency**: Network-dependent
- **Buffering**: OS-level buffering applies
- **Timeout overhead**: Minimal (<1ms for timeout setup)

---

## Related Documentation

- **REPL Usage**: `docs/repl-usage.md` - Interactive I/O examples
- **Observability**: `docs/observability.md` - Tracing port operations
- **Contracts**: `specs/002-implement-deferred-features/contracts/ports.md` - Technical specifications
- **Quickstart**: `specs/002-implement-deferred-features/quickstart.md` - Port examples

---

## Implementation Status

**Feature 002 - User Story 2**: Complete ✅

**Implemented**:
- ✅ File, TCP, HTTP/HTTPS port drivers
- ✅ Sandbox enforcement with symlink resolution
- ✅ TLS verification (secure by default) with `--insecure` opt-out
- ✅ Timeout controls
- ✅ `open`, `close`, `read`, `write`, `save`, `load`, `query`, `wait`
- ✅ Refinements: `--binary`, `--lines`, `--append`, `--timeout`, `--insecure`
- ✅ Redirect following (HTTP, max 10 hops)
- ✅ Client pooling by TLS/timeout key
- ✅ Port lifecycle trace events

**Tested**:
- ✅ 13/13 contract tests passing
- ✅ SC-012 integration tests (throughput, latency, sandbox enforcement)
- ✅ Backward compatibility verified (84 Feature 001 tests + 13 User Story 1 tests)

See `specs/002-implement-deferred-features/tasks.md` for details.

---

Enjoy safe, powerful I/O in Viro! 🔒
