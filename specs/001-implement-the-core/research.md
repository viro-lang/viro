# Phase 0: Research & Technical Decisions

**Feature**: Viro Core Language and REPL  
**Date**: 2025-01-07  
**Purpose**: Resolve all NEEDS CLARIFICATION items from Technical Context

## Research Tasks

### 1. REPL Readline Library Selection

**Question**: Which Go readline library should be used for command history and multi-line input?

**Options Evaluated**:
- **chzyer/readline**: Pure Go, cross-platform, actively maintained (2k+ stars)
- **peterh/liner**: Simple, pure Go, good for basic needs (1k+ stars)
- **ergochat/readline**: Fork of chzyer/readline with improvements
- **go-linenoise**: Go bindings to C linenoise (smaller, C dependency)

**Decision**: Use **chzyer/readline** (github.com/chzyer/readline)

**Rationale**:
- Pure Go implementation (no C dependencies, easier cross-platform builds)
- Supports command history persistence to file
- Handles multi-line input with continuation prompts
- Supports Ctrl+C interrupt handling
- Well-documented API with examples
- Active maintenance and wide adoption in Go CLI tools
- Compatible with macOS, Linux, Windows without platform-specific code

**Alternatives Considered**:
- go-linenoise rejected due to C dependency complexity
- peterh/liner considered but chzyer/readline has better multi-line support
- ergochat fork not chosen as base chzyer/readline is stable and sufficient

**Integration Pattern**:
```go
import "github.com/chzyer/readline"

rl, _ := readline.NewEx(&readline.Config{
    Prompt: ">> ",
    HistoryFile: "/tmp/.viro_history",
    InterruptPrompt: "^C",
})
defer rl.Close()

for {
    line, err := rl.Readline()
    // evaluate line
}
```

---

### 2. Table-Driven Testing Patterns for Interpreter

**Question**: How should we structure table-driven tests for native functions and evaluation scenarios?

**Best Practices Researched**:
- Go standard library uses subtests with `t.Run()` for table-driven tests
- Each test case is a struct with inputs and expected outputs
- Subtests enable parallel execution with `t.Parallel()`
- Clear failure messages with descriptive test case names

**Decision**: Use standard Go table-driven test pattern with subtests

**Rationale**:
- Native to Go's testing philosophy, no external dependencies
- Excellent failure reporting (shows which case failed)
- Supports parallel execution for performance
- Easy to add new test cases without duplicating test logic
- Works well with `go test -v` for verbose output

**Pattern for Native Function Contracts**:
```go
func TestNativeAdd(t *testing.T) {
    tests := []struct {
        name     string
        args     []Value
        expected Value
        wantErr  bool
    }{
        {"add integers", []Value{IntVal(3), IntVal(4)}, IntVal(7), false},
        {"type error", []Value{StrVal("x"), IntVal(4)}, nil, true},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := NativeAdd(tt.args)
            if (err != nil) != tt.wantErr {
                t.Errorf("unexpected error status: %v", err)
            }
            if !ValuesEqual(result, tt.expected) {
                t.Errorf("got %v, want %v", result, tt.expected)
            }
        })
    }
}
```

**Pattern for Evaluation Scenarios**:
```go
func TestEvaluator(t *testing.T) {
    tests := []struct {
        name   string
        input  string       // source code
        output string       // expected result
        errCat ErrorCategory // expected error category (0 = no error)
    }{
        {"literal integer", "42", "42", 0},
        {"arithmetic", "3 + 4", "7", 0},
        {"undefined word", "undefined-var", "", ErrScript},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := Eval(tt.input)
            // assert result and error
        })
    }
}
```

**Alternatives Considered**:
- External testing frameworks (testify, ginkgo) rejected per constitution's "minimize external dependencies"
- Benchmark-driven development deferred until correctness is established

---

### 3. Stack Expansion Strategy in Go

**Question**: What is the optimal strategy for automatic stack expansion using Go slices?

**Go Slice Growth Research**:
- Go runtime doubles slice capacity when appending beyond capacity (for slices >1024 elements, growth factor becomes ~1.25x)
- Manual pre-allocation with `make([]T, 0, initialCap)` avoids early reallocations
- Index-based access prevents pointer invalidation (critical per constitution Principle IV)

**Decision**: Use slice with manual capacity management and index-based access

**Rationale**:
- Go slices handle reallocation automatically, but we control initial capacity
- Index-based access (via integers) never invalidates even when slice grows
- Stack frames reference elements by index offset, not pointers
- Performance: pre-allocating reasonable capacity (e.g., 256 slots) avoids most expansions

**Implementation Pattern**:
```go
type Stack struct {
    data []Value
    top  int  // index of next available slot
}

func NewStack(initialCap int) *Stack {
    return &Stack{
        data: make([]Value, 0, initialCap),
        top:  0,
    }
}

func (s *Stack) Push(v Value) {
    if s.top >= len(s.data) {
        s.data = append(s.data, v)
    } else {
        s.data[s.top] = v
    }
    s.top++
}

func (s *Stack) Get(index int) Value {
    return s.data[index]  // index-based, safe across expansions
}

func (s *Stack) Set(index int, v Value) {
    s.data[index] = v
}
```

**Key Safety Properties**:
- Frame pointers are integers (indices), never slice references
- Expansion preserves all existing indices
- Access via `Get(index)` and `Set(index, value)` - no direct slice indexing from outside

**Alternatives Considered**:
- Pre-allocated fixed-size array rejected due to inflexibility and artificial limits
- Linked list of stack chunks rejected due to complexity and cache performance issues

---

### 4. Error Context Capture ("near" and "where")

**Question**: How should we efficiently capture "near" (source expression) and "where" (call stack) context for errors?

**Design Considerations**:
- "Near" context requires tracking current expression during evaluation
- "Where" context requires stack trace of function calls
- Must balance detail vs performance overhead
- Errors should be structured per constitution Principle V

**Decision**: Thread evaluation context through evaluator and maintain call frame chain

**Rationale**:
- Evaluator already maintains index into block being evaluated
- Can extract window of expressions around current index for "near"
- Stack frames link via prior frame pointer, forming natural call chain
- Capture context only when error occurs (zero overhead in success path)

**Near Context Pattern**:
```go
type EvalContext struct {
    source Block      // block being evaluated
    index  int        // current evaluation position
}

func (ctx *EvalContext) NearContext() []Value {
    // Return 3 values before and after current position
    start := max(0, ctx.index-3)
    end := min(len(ctx.source), ctx.index+4)
    return ctx.source[start:end]
}
```

**Where Context Pattern**:
```go
type Frame struct {
    priorFrame int     // index of calling frame (for stack trace)
    function   Value   // function being executed (for call name)
    // ... other fields
}

func (s *Stack) CaptureCallStack() []string {
    var calls []string
    frameIdx := s.currentFrame
    for frameIdx != -1 {
        frame := s.GetFrame(frameIdx)
        calls = append(calls, frame.function.Name())
        frameIdx = frame.priorFrame
    }
    return calls
}
```

**Error Structure**:
```go
type Error struct {
    Category ErrorCategory  // 0-900 range
    Code     int           // specific error code
    ID       string        // error identifier (e.g., "no-value")
    Args     [3]string     // message interpolation arguments
    Near     []Value       // expressions around error location
    Where    []string      // call stack trace
}

func (e Error) Error() string {
    return fmt.Sprintf("%s error (%d): %s\nNear: %v\nWhere: %v",
        e.Category, e.Code, e.Message(), e.Near, e.Where)
}
```

**Alternatives Considered**:
- Copying entire source block rejected as too expensive
- Using runtime.Callers() rejected as it captures Go stack, not Viro call stack
- String-only context rejected as structured values provide better debugging

---

### 5. UTF-8 String Handling for Character Series

**Question**: How should strings be handled to support Viro's series operations while maintaining UTF-8 compatibility?

**Go String Research**:
- Go strings are immutable UTF-8 byte sequences
- Indexing `s[i]` yields bytes, not characters (runes)
- Rune iteration via `for _, r := range s` handles multi-byte characters correctly
- Viro treats strings as character series (not byte series)

**Decision**: Represent strings as `[]rune` internally for series operations

**Rationale**:
- String operations (`first`, `last`, `at`) operate on characters, not bytes
- Converting `string` to `[]rune` gives character-level access
- Mutation operations (append, insert) are easier on `[]rune`
- Only convert to Go `string` for I/O and display
- Performance acceptable for REPL use (optimization possible later if needed)

**Implementation Pattern**:
```go
type StringValue struct {
    runes []rune  // character sequence
}

func (s StringValue) First() rune {
    return s.runes[0]
}

func (s StringValue) Last() rune {
    return s.runes[len(s.runes)-1]
}

func (s StringValue) Append(r rune) {
    s.runes = append(s.runes, r)
}

func (s StringValue) String() string {
    return string(s.runes)  // convert for display
}

// Construction from literal
func ParseString(literal string) StringValue {
    return StringValue{runes: []rune(literal)}
}
```

String operations in Viro:
- `first "hello"` → 'h'
- `last "hello"` → 'o'
- `length? "hello"` → 5
- `append "hello" " world"` → "hello world"

**Edge Cases**:
- Empty string operations check length before access
- Multi-byte Unicode characters handled correctly (emoji, accented chars)
- String comparison via rune slice comparison or convert to string

**Alternatives Considered**:
- Keeping strings as Go `string` type rejected due to awkward character indexing
- Byte-oriented series rejected as incompatible with character semantics
- Normalized Unicode (NFC/NFD) deferred to later phase per spec clarification (case-sensitive, no normalization initially)

---

### 6. Interpreter Benchmarking Strategy

**Question**: How should we establish performance baselines and track performance over time?

**Go Benchmarking Research**:
- Go provides `testing.B` for benchmark tests
- Run with `go test -bench=. -benchmem`
- Reports ns/op (nanoseconds per operation) and allocations
- Can compare benchmarks across commits with benchstat tool

**Decision**: Use Go benchmark tests for each architecture layer with CI tracking

**Rationale**:
- Native to Go tooling, no external dependencies
- Granular benchmarks for each layer (stack ops, value dispatch, native functions, full evaluation)
- Benchmark results tracked in git (markdown file) to detect regressions
- Aligns with success criteria SC-005 (simple <10ms, complex <100ms)

**Benchmark Pattern**:
```go
// Stack operations benchmark
func BenchmarkStackPush(b *testing.B) {
    s := NewStack(256)
    val := IntVal(42)
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        s.Push(val)
    }
}

// Native function benchmark
func BenchmarkNativeAdd(b *testing.B) {
    args := []Value{IntVal(3), IntVal(4)}
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        NativeAdd(args)
    }
}

// Full evaluation benchmark
func BenchmarkEvalSimple(b *testing.B) {
    source := Parse("3 + 4")
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        Eval(source)
    }
}

func BenchmarkEvalComplex(b *testing.B) {
    source := Parse("loop 10 [x: x + 1]")
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        Eval(source)
    }
}
```

**Baseline Targets** (per SC-005):
- Stack push/pop: <100 ns/op
- Type dispatch: <50 ns/op
- Native function call: <1 µs/op
- Simple expression evaluation (literal, arithmetic): <10 ms
- Complex expression (nested function calls): <100 ms

**Performance Tracking**:
- Run benchmarks on every PR
- Store results in `benchmarks/results.md` with commit hash and date
- Compare with previous results to detect >10% regressions
- Optimization deferred per constitution Principle VII (correctness first)

**Alternatives Considered**:
- External profiling tools (pprof) available but not required initially
- Continuous benchmarking service (benchdash) deferred until performance becomes critical
- Manual timing rejected in favor of standardized Go benchmarks

---

## Summary of Resolved Clarifications

| Item | Original Status | Resolution |
|------|----------------|------------|
| Primary Dependencies | NEEDS CLARIFICATION | chzyer/readline for REPL |
| Testing | NEEDS CLARIFICATION | Go table-driven tests with subtests |
| Performance Goals | NEEDS CLARIFICATION | Stack ops <100ns, simple eval <10ms, complex eval <100ms, tracked via Go benchmarks |

All technical unknowns from Technical Context have been researched and resolved. Architecture decisions documented for Phase 1 design artifact generation.

**Next Phase**: Phase 1 - Data Model, Contracts, Quickstart, Agent Context Update
