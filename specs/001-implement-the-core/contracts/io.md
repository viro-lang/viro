# Contract: I/O Natives

**Category**: Input/Output Operations  
**Functions**: `print`, `input`  
**Purpose**: Basic console I/O for REPL interaction

---

## Native: `print`

**Signature**: `print value`

**Parameters**:
- `value`: Any type (evaluated)

**Return**: None

**Behavior**: 
1. Evaluate argument
2. If argument is a block: **reduce** it (evaluate each element), then join results with spaces
3. If argument is any other type: convert value to string representation
4. Output string to standard output (stdout) with newline
5. Return none

**Type Rules**:
- Accepts any value type
- **Special handling for blocks**: reduces (evaluates contents) before printing
- Automatic conversion to string for display

**String Conversion Rules**:
- Integer → decimal representation (e.g., `42` → "42")
- String → content without quotes (e.g., `"hello"` → hello)
- Block → **REDUCE block first**, then join elements with spaces (e.g., `[1 2 3]` → "1 2 3", `["Hello" name]` with name="Alice" → "Hello Alice")
- Word → symbol name (e.g., `'x` → x)
- Logic → `true` or `false`
- None → empty string or "none" (implementation choice)
- Function → `[function]` or function name

**Examples**:
```viro
print 42            → prints "42" (with newline)
print "hello"       → prints "hello"
print [1 2 3]       → prints "1 2 3" (block reduced: each element evaluated and joined)
print ["Hi" "there"]→ prints "Hi there" (strings joined with space)
name: "Alice"
print ["Hello" name]→ prints "Hello Alice" (name evaluated to "Alice", joined)
print true          → prints "true"
print none          → prints "none" or empty line
```

**Test Cases**:
1. `print 42` outputs "42\n" to stdout, returns none
2. `print "hello"` outputs "hello\n", returns none
3. `print [1 2 3]` outputs "1 2 3\n" (reduced block), returns none
4. `print ["Hi" "there"]` outputs "Hi there\n", returns none
5. `name: "Alice"` then `print ["Hello" name]` outputs "Hello Alice\n", returns none
6. `print true` outputs "true\n", returns none
7. `print none` outputs "none\n" or "\n", returns none
8. Capture output: verify exact string with newline
9. Return value: always none

**Block Reduction Behavior**:
- When print receives a block, it evaluates each element in the block
- Results are converted to strings and joined with single spaces
- Example: `print [1 + 1 3 * 4]` → "2 12" (evaluates expressions)
- Example: `x: 5  print ["Value:" x]` → "Value: 5" (evaluates word)

**Side Effects**:
- Writes to stdout (observable in tests via output capture)
- Adds newline after content

**Error Cases**: None (accepts any value, converts to string)

**Usage in Tests**:
```go
func TestNativePrint(t *testing.T) {
    oldStdout := os.Stdout
    r, w, _ := os.Pipe()
    os.Stdout = w
    
    result, err := NativePrint([]Value{IntVal(42)})
    
    w.Close()
    os.Stdout = oldStdout
    
    var buf bytes.Buffer
    io.Copy(&buf, r)
    
    assert.NoError(t, err)
    assert.Equal(t, "42\n", buf.String())
    assert.Equal(t, NoneVal(), result)
}
```

---

## Native: `input`

**Signature**: `input`

**Parameters**: None

**Return**: String (user input line)

**Behavior**: 
1. Read line from standard input (stdin)
2. Remove trailing newline
3. Return line as string value

**Examples**:
```viro
name: input         ; waits for user to type and press Enter
                    ; if user types "Alice", name becomes "Alice"
print name          ; prints "Alice"
```

**Test Cases**:
1. Simulate input "hello\n" → returns `"hello"` (no newline)
2. Simulate input "123\n" → returns `"123"` (string, not integer)
3. Simulate input "\n" (empty line) → returns `""`
4. Return type is always string

**Error Cases**:
- EOF (Ctrl+D) → Return empty string or signal EOF (implementation choice)
- Read error → Access error (500): "Input error"

**Side Effects**:
- Reads from stdin (blocks until input available)
- Consumes input line including newline

**Usage in REPL**:
```viro
print "What is your name?"
name: input
print ["Hello" name]
```

**Usage in Tests**:
```go
func TestNativeInput(t *testing.T) {
    oldStdin := os.Stdin
    r, w, _ := os.Pipe()
    os.Stdin = r
    
    go func() {
        w.Write([]byte("Alice\n"))
        w.Close()
    }()
    
    result, err := NativeInput([]Value{})
    
    os.Stdin = oldStdin
    
    assert.NoError(t, err)
    assert.Equal(t, StrVal("Alice"), result)
}
```

**Implementation Notes**:
- Use `bufio.Scanner` or `bufio.Reader.ReadString('\n')`
- Strip trailing newline character(s)
- Handle both Unix (`\n`) and Windows (`\r\n`) line endings
- Empty input (just Enter) returns empty string

---

## Common Properties

**Standard I/O**:
- `print` writes to stdout
- `input` reads from stdin
- Standard Go I/O streams (`os.Stdout`, `os.Stdin`)

**String Handling**:
- `print` converts any value to string representation
- **`print` reduces blocks**: evaluates each element and joins with spaces
- `input` always returns string (no automatic parsing)
- User must convert input if numeric value needed

**Block Reduction in Print**:
- Critical feature: `print [...]` evaluates block contents before printing
- Enables convenient string interpolation: `print ["Hello" name]`
- Without this, would need separate `reduce` function (not in Phase 1)
- Similar to REBOL's print behavior

**Error Handling**:
- `print` does not error (best effort output)
- `input` can error on read failure (Access error category)

**Return Values**:
- `print` always returns none
- `input` returns string

**Blocking Behavior**:
- `print` is non-blocking (writes immediately)
- `input` blocks until user provides input

**Testing Considerations**:
- Must capture/redirect stdout for print tests
- Must simulate stdin for input tests
- Use Go's `os.Pipe()` for I/O redirection in tests

---

## REPL Integration

**Print in REPL**:
```viro
>> print "hello"
hello              ; output from print
                   ; no => none shown (none results are suppressed in REPL)

>> print [1 2 3]
1 2 3              ; block reduced: elements joined with spaces

>> name: "Alice"
=> "Alice"

>> print ["Hello" name]
Hello Alice        ; name evaluated and interpolated
```

**Input in REPL**:
```viro
>> name: input
[user types: Alice]
=> "Alice"         ; REPL shows return value (non-none)

>> print name
Alice
```

**Interactive Examples**:
```viro
>> loop 3 [print "hi"]
hi
hi
hi

>> print 3 + 4
7

>> x: input  print ["You entered:" x]
[user types: test]
You entered: test    ; block reduced with x evaluated
```

---

## Implementation Checklist

For each native:
- [ ] Function signature matches contract
- [ ] Correct I/O stream (stdout for print, stdin for input)
- [ ] Value to string conversion (print)
- [ ] Newline handling (print adds, input strips)
- [ ] Return correct value (none for print, string for input)
- [ ] All test cases pass
- [ ] I/O redirection in tests works correctly

**Dependencies**:
- Value system (string representation for all types)
- Error system (Access error for input failures)
- Go standard library: `os`, `bufio`, `fmt`

**Testing Strategy**:
- Output capture for print tests (verify exact output)
- Input simulation for input tests (use pipes)
- Round-trip tests (input → print)
- Empty input tests (edge cases)
- Type conversion tests (print with various types)
- **Block reduction tests** (verify print evaluates block contents)
- **Interpolation tests** (print with variables in blocks)

**Performance Considerations**:
- Print operations should be fast (buffered I/O)
- Input blocks (expected behavior, not performance issue)
- No artificial delays or sleeps

**Future Extensions** (out of Phase 1 scope):
- Formatted print (`printf` style)
- File I/O (`read`, `write`)
- Binary I/O
- Network I/O
- Error output stream (stderr)
- Non-blocking input
