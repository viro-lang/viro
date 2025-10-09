# Quickstart: Deferred Language Capabilities

**Feature**: Deferred Language Capabilities (002)  
**Audience**: Engineers implementing or validating Phase 002 features  
**Prerequisites**: Core interpreter (Feature 001) operational

---

## 1. High-Precision Decimals

### Creating Decimals
```viro
price: decimal "19.99"
qty: 3

print price * qty         ; ==> 59.97 (decimal, scale preserved)
print round --places price 2 ; ==> 19.99
```

### Advanced Math
```viro
interest-rate: decimal "0.035"
years: 5

future: principal * exp interest-rate * years
print future

print sin decimal "1.570796326794897" ; approx π/2 → 1.0
```

### Rounding Modes
```viro
round --mode decimal "2.5" 'half-up      ; ==> 3
round --mode decimal "2.5" 'half-down    ; ==> 2
round --places decimal "3.14159" 3      ; ==> 3.142
```

---

## 2. Sandbox Ports

### Reading & Writing Files
```viro
; sandbox root supplied via CLI flag: --sandbox-root %~/viro-sandbox
save %reports/q3.txt [revenue 125000 profit 37800]
data: load %reports/q3.txt
print data.profit
```

### HTTP Requests with TLS Controls
```viro
resp: read https://status.viro.dev
print length? resp

secure: open https://api.viro.dev
insecure: open --insecure https://self-signed.example ; allowed but logged
close secure
close insecure
```

### TCP with Timeouts
```viro
port: open --timeout 500 tcp://localhost:9000
write port "PING\n"
reply: read --part port 4
close port
```

---

## 3. Objects and Paths

### Objects & Path Mutation
```viro
invoice: object [
    id: 42
    customer: "Acme"
    items: [
        [description: "Widget" price: decimal "19.99" qty: 3]
    ]
]

; Update quantity via path set
invoice/items/1/qty: 4
print invoice/items/1/qty
```

### Path Access Patterns
```viro
user: object [
    name: "Ana"
    address: object [
        city: "Porto"
        zip: 4000
    ]
]

; Read nested path
print user/address/city  ; ==> "Porto"

; Modify nested value
user/address/city: "Lisboa"
print user/address/city  ; ==> "Lisboa"
```

---

## 4. Parse Dialect Essentials

```viro
rule: [copy name some letter space copy amount some digit]
text: "Invoice ABC123 199"

if parse text rule [
    print ["Name:" name "Amount:" amount]
]
```

### Case-Sensitive Parsing
```viro
parse --case "ABC" ["ABC"] ; true
parse --case "abc" ["ABC"] ; false
```

### Nested Structures
```viro
parse data [
    some [into [word set value integer!]]
]
```

---

## 5. Observability: Trace & Debug

### Trace Workflow
```viro
trace --on --only ['calculate-interest]
result: calculate-interest 100 decimal "0.05" 3
trace --off
```
- Trace events saved to `viro-trace.log` (rotating file).
- Use `trace --on --file` to select custom path within sandbox.

### Debug Session
```viro
debug --on
debug --breakpoint 'calculate-interest
result: calculate-interest 100 decimal "0.05" 3
; REPL drops into (debug) prompt when breakpoint hits
; Use: debug --locals / debug --stack / debug --step / debug --continue
debug --off
```

---

## 6. Reflection Helpers

```viro
print type-of invoice                     ; object!
print words-of invoice                    ; [id customer items]
print values-of invoice                   ; deep copy of fields
print spec-of :my-function                ; function spec block
print body-of :my-function                ; function body
```

---

## Validation Checklist

- [ ] Decimals create, round, and participate in advanced math operations.
- [ ] Ports respect sandbox, TLS verification, and timeout rules.
- [ ] Objects support nested structures with path-based access and mutation.
- [ ] Objects support nested path access and mutation.
- [ ] Parse dialect handles literals, quantifiers, captures, and nested rules.
- [ ] Trace logs created and debugger can set breakpoints/inspect state.
- [ ] Reflection functions return immutable snapshots.

Use this quickstart as a smoke test guide once implementation is complete.
