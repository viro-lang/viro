# Data Model: Deferred Language Capabilities

**Feature**: Deferred Language Capabilities (002)  
**Date**: 2025-10-08  
**Source**: Consolidated from spec.md (Key Entities, Functional Requirements, Clarifications)

## Overview

Phase 002 extends the Viro interpreter data model established in Feature 001 with richer numeric types, I/O abstractions, advanced parsing, and observability. Enhancements focus on:

1. **Value System Extensions**: `decimal!`, `object!`, `port!`, and enriched path semantics.
2. **I/O & Integration**: Unified port drivers for filesystem, TCP, and HTTP with sandbox enforcement.
3. **Program Structure**: Objects and path expressions enabling encapsulation and data organization.
4. **Pattern Matching**: Parse dialect representation supporting composable rules and captures.
5. **Operational Insight**: Trace sessions, debugger state, and reflection metadata for insight into runtime behavior.

All new entities respect constitutional principles—particularly index-based stack/frame safety, structured errors, and incremental layering. Existing entities from Feature 001 remain unchanged unless explicitly noted.

---

## Entity Definitions

### 1. DecimalValue (`decimal!`)

**Purpose**: High-precision decimal floating point values with IEEE 754 decimal128 semantics, including rounding context metadata.

**Structure**:
```go
type DecimalValue struct {
    Magnitude *decimal.Big // from github.com/ericlagergren/decimal
    Context   decimal.Context // precision, rounding mode, traps
    Scale     int16           // digits to the right of decimal point for round-trip formatting
}
```

**Fields**:
- `Magnitude`: Normalized decimal number supporting ±10^−6143…±10^6144.
- `Context`: Shared context controlling precision (default 34 digits) and rounding (half-even by default).
- `Scale`: Explicit scale metadata used to preserve formatting when re-serialising values (e.g., differentiates `1.20` vs `1.2`).

**Operations**:
- Arithmetic: `Add`, `Sub`, `Mul`, `Div`, `Pow` (delegates to decimal library with context).
- Transcendentals: `Exp`, `Log`, `Log10`, `Sin`, `Cos`, `Tan`, `Asin`, `Acos`, `Atan`.
- Rounding helpers: `Round`, `Ceil`, `Floor`, `Truncate` (accept precision refinements and rounding mode overrides).
- Conversion: `FromString`, `FromInt64`, `ToStringWithScale`.

**Validation Rules**:
- Context precision ≥ 1 and ≤ 34 (decimal128 target). Higher precision rejected by phase scope.
- Scale between −6143 and +6144 (library bounds). Values outside raise Math error (400).
- All operations trap on NaN or Infinity; interpreter converts to Math errors per FR-004/FR-005.

**Relationships**:
- Wrapped by `Value{Type: TypeDecimal, Payload: *DecimalValue}`.
- Interacts with integer values via automatic promotion (integers converted to decimal with scale 0).

---

### 2. Value Type Extensions

**Purpose**: Extend global value enumeration to accommodate new first-class types.

**Additions**:
```go
const (
    TypeDecimal ValueType = iota + 11 // continue after existing types
    TypeObject                        // object! instance
    TypePort                          // open port handle
    TypePath                          // evaluated path result placeholder
)
```

**Notes**:
- `TypePath` acts as transient container when evaluating path segments; ultimately resolves to another Value type and should not persist outside evaluation.
- Constructors and type assertions must be added for each new type to preserve type-based dispatch fidelity (Principle III).

---

### 3. ObjectInstance (`object!`)

**Purpose**: Encapsulate word/value bindings in a dedicated frame with optional parent prototype, enabling hierarchical data structures.

**Structure**:
```go
type ObjectInstance struct {
    Frame       core.Frame      // owned frame for self-contained field storage
    ParentProto *ObjectInstance // parent prototype object (nil if none)
    Manifest    ObjectManifest  // field metadata
}
```

**Key Changes (Phase 3 Refactor)**:
- **Owned Frames**: Objects now own their frames directly instead of referencing evaluator frameStore
- **Self-Contained**: Objects are portable and serializable (no evaluator coupling)
- **Prototype Chain**: ParentProto enables inheritance without frame index dependencies

**Operations**:
- Construction via native `object [spec]` – creates owned Frame, binds words, evaluates initializers
- Field access: `GetField(name)` searches owned frame; `GetFieldWithProto(name)` traverses prototype chain
- Path traversal: `object.field` uses owned frame lookup with prototype inheritance
- Mutation: `object.field: value` updates owned frame with validation

**Validation Rules**:
- Field names unique within object; duplicates raise Script error (300)
- Type hints (when provided) enforced on assignment
- Parent chain must be acyclic to prevent infinite lookup loops

**Relationships**:
- **Self-Contained**: No evaluator dependency - objects can be created, passed, and serialized independently
- **Portable**: Objects work across evaluator instances and can be saved/loaded
- **Memory Managed**: Object frames managed by object lifecycle, not evaluator frameStore
- Path expressions use ObjectInstance.GetFieldWithProto() for resolution

---

### 4. PathExpression

**Purpose**: Evaluate nested access/mutation across objects, blocks, series, and future map types.

**Structure**:
```go
type PathSegment struct {
    Kind   SegmentKind
    Word   string
    Index  int
    Refinement string
    Evaluated Value // cached result for mutation stage
}

type PathExpression struct {
    Segments []PathSegment
}
```

**Segment Kinds**:
- `WordSegment`: standard `object.field` (dot-separated path).
- `IndexSegment`: numeric access `block.2` (dot-separated for series).
- `RefinementSegment`: function refinements `func --only` (Phase 3 usage).
- `EvalSegment`: paren expression `array.(index)` (evaluated at runtime).

**Parser Disambiguation**: The tokenizer distinguishes decimal literals, refinements, and paths by examining the first character(s). Numbers start with a digit or `-` followed immediately by a digit (e.g., `19.99`, `-3.14`), refinements start with `--` followed by a letter (e.g., `--option`, `--places`), and words/paths start with a letter (e.g., `config.timeout`, `items.5`). This allows unambiguous parsing of decimal fractions, refinements, and path expressions using the dot separator.

**Eval Segment Restrictions**:
- Leading eval segments (e.g., `.(expr).field`) are syntactically invalid and rejected by the parser.
- Eval segments are only permitted after a valid base (word or index segment).
- Valid result types from eval segment expressions: `word!`, `string!`, `integer!`.
- Other types (block, object, decimal, etc.) raise Script error (300) with clear diagnostic message.
- Eval segments are evaluated once per path traversal and results cached to prevent re-evaluation.

**Evaluation Flow**:
1. Resolve first segment (word) via current frame.
2. Iteratively apply segments: for objects use Frame lookup; for blocks/strings adjust index; for functions apply refinement metadata.
3. Eval segments are materialized on-demand during traversal, and results are cached.
4. For assignment, keep track of penultimate target and final segment info to update underlying storage.
5. Set-path operations never re-evaluate eval segments; cached materialized values are reused.

**Validation Rules**:
- All intermediate values must support requested segment type; mismatches raise Script error (300).
- Assignment disallowed on immutable targets (e.g., decimal literal) raising Script error.
- Eval segments in assignment paths must resolve to valid field/index identifiers.
- Attempting to assign through eval segments that yield unsupported types produces clear error context.

**Relationships**:
- Uses ObjectInstance, BlockValue, StringValue to traverse structures.
- Works closely with evaluator's dispatch to differentiate read vs write paths.

---

### 5. Port

**Purpose**: Abstract I/O channels (files, TCP sockets, HTTP streams) with optional timeouts and unified API.

**Structure**:
```go
type Port struct {
    Scheme   string       // file, tcp, http
    Spec     PortSpec     // parsed target details
    Driver   PortDriver   // scheme-specific implementation
    Timeout  *time.Duration // optional user-provided timeout (nil => OS default)
    State    PortState
}

type PortSpec struct {
    Path   string        // file path or URL/host information
    Mode   PortMode      // read/write/append
    Options map[string]Value // protocol-specific options (headers, method, etc.)
}

type PortState uint8
const (
    PortClosed PortState = iota
    PortOpen
    PortEOF
)
```

**Driver Interface** (see Research Task 4): ensures index-based references by storing driver state inside Port struct, not external global.

**Validation Rules**:
- File operations must resolve path through sandbox root helper before driver invocation (Access error on escape attempts).
- HTTP operations enforce TLS verification unless `--insecure` refinement present.
- TCP/HTTP operations rely on OS timeouts when `Timeout == nil` (clarification #4). When provided, driver uses `context.WithTimeout`.

**Relationships**:
- Stored as `Value{Type: TypePort, Payload: *Port}`; when REPL closes port or GC runs, driver Close invoked.
- Trace events log port lifecycle transitions (open, read, write, close).

---

### 7. ParsePattern

**Purpose**: Represent declarative parse rules enabling recursive descent evaluation with captures and control combinators.

**Structure**:
```go
type ParseRuleKind uint8
const (
    RuleLiteral ParseRuleKind = iota
    RuleWord
    RuleSet
    RuleCopy
    RuleSome
    RuleAny
    RuleOpt
    RuleNot
    RuleInto
    RuleAhead
    RuleBlock
)

type ParsePattern struct {
    Rules []ParseRule
}

type ParseRule struct {
    Kind       ParseRuleKind
    Literal    Value        // for literal match
    Word       string       // symbol reference
    Children   []ParseRule  // nested rules (e.g., into, block)
    TargetWord string       // for set/copy operations
}
```

**Evaluation State**:
```go
type ParseState struct {
    Input   []Value   // block or rune sequence for strings
    Index   int
    Stack   []parseFrame // recursion stack for into/blocks
    Captures map[string]Value // words set/copied during parse
}
```

**Validation Rules**:
- Detect infinite loops by tracking rule/input pairs visited; fail with Syntax error (200) if repeated without progress (edge case requirement).
- Non-boolean return (should be success flag) converted to boolean per Viro semantics.

**Relationships**:
- Integrates with evaluator to supply context for `parse` native (FR-015/FR-016).
- Captures assign words in caller frame consistent with existing frame safety rules.

---

### 8. TraceSession & TraceEvent

**Purpose**: Provide structured observability for evaluation steps, emissions directed to rotating log file.

**Structure**:
```go
type TraceSession struct {
    Enabled    bool
    Sink       TraceSink // interface for writing events (file, stdout, custom)
    Filters    TraceFilters
    SequenceID uint64
}

type TraceFilters struct {
    IncludeWords []string
    ExcludeWords []string
    MinDuration  time.Duration
}

type TraceEvent struct {
    ID        uint64
    Timestamp time.Time
    Word      string
    Value     Value
    Duration  time.Duration
    Category  string // eval, native, error, trace
    Metadata  map[string]any
}
```

**Trace Sink**: Default sink uses `lumberjack.Logger` writing JSON lines. Additional sinks can wrap TraceSink interface for custom outputs.

**Validation Rules**:
- When session disabled, Sink writes no events and SequenceID not incremented.
- Enabling/disabling trace toggles should flush outstanding events and record control event.

**Relationships**:
- Evaluator instruments entry/exit of `Do_Next`, `Do_Blk`, native invocations.
- Debug commands interact with TraceSession to step or inspect events.

---

### 9. Debugger State & Breakpoints

**Purpose**: Manage breakpoints, stepping, and inspection for the `debug` native.

**Structure**:
```go
type Breakpoint struct {
    ID        int
    Target    BreakpointTarget
    Condition func(*Evaluator) bool // optional conditional break
}

type BreakpointTarget struct {
    WordName string // function name or word symbol
    Location LocationSpec // optional block index/offset
}

type Debugger struct {
    Breakpoints map[int]Breakpoint
    NextID      int
    Mode        DebugMode // continue, step-into, step-over
}
```

**Relationships**:
- Debugger owned by REPL session; interacts with TraceSession to emit events.
- Commands `debug --breakpoint`, `debug --remove`, `debug --step`, `debug --locals`, etc., act on Debugger.

---

### 10. Reflection Metadata

**Purpose**: Serve reflection natives `type-of`, `spec-of`, `body-of`, `words-of`, `values-of` with immutable snapshots.

**Structure**:
```go
type ReflectionSnapshot struct {
    TargetType ValueType
    TypeName   string
    Spec       []Value   // parameter block copy for functions
    Body       *BlockValue // deep copy for safety
    Words      []string
    Values     []Value
}
```

**Notes**:
- Snapshots generated on demand; expensive copies flagged so REPL warns when large structures encountered (trace event `reflection-large`).
- Maintains local-by-default semantics by copying values rather than exposing live frames.

---

## Updated Interactions

### Value Promotion & Arithmetic Flow

1. Mixed expression `19.99 * 3`:
   - Integer promoted to DecimalValue with scale 0.
   - Multiplication executes in Decimal context, resulting scale derived from operands (scale 2).
   - Result stored as DecimalValue (payload decimal.Big) and displayed respecting scale.

2. Rounding `round --places total 2`:
   - Refinement translates into temporary Context override (precision = scale + requested places).
   - Result DecimalValue returns with updated scale 2.

### File Sandbox Resolution

1. REPL receives `save %reports/summary.txt data`.
2. Sandbox helper resolves CLI-provided root (e.g., `/Users/alex/viro-sandbox`).
3. Target path cleaned: `filepath.Join(root, "reports/summary.txt")`.
4. `EvalSymlinks` ensures final path remains under root; Access error if not.
5. Port driver writes file, updates PortState.

### Trace & Debug Interaction

1. User executes `trace --on`.
2. TraceSession `Enabled=true`, sink uses lumberjack-backed logger writing JSON lines.
3. Each evaluation step emits event with SequenceID++.
4. User triggers `debug --breakpoint 'square`.
5. Breakpoint stored, and when function invoked evaluator checks breakpoints before executing body, handing control to interactive prompt.

---

## Compliance with Constitution

- **Type Dispatch (Principle III)**: New types registered with evaluator dispatch tables before native resolution.
- **Stack/Frame Safety (Principle IV)**: Objects reuse existing frame infrastructure with index references; no pointer shortcuts introduced.
- **Structured Errors (Principle V)**: All new failure modes (TLS override, sandbox violation, invalid parse rule, whitelist miss) mapped to appropriate categories (Math, Access, Syntax, Script).
- **Observability (Principle VI)**: TraceSession and Debugger produce structured, user-readable output.
- **YAGNI (Principle VII)**: Scope limited to spec-mandated features; advanced protocols (FTP, WebSocket) deferred.

---

## Future Considerations (Post-Phase 002)

- **Decimal Performance Optimisations**: Investigate pooled contexts or hardware acceleration if benchmarks reveal hotspots.
- **Parse Dialect Extensions**: Add `thru`, `reject`, and custom productions once core combinators stable.
- **Port Drivers**: Add UDP, WebSocket, and TLS socket support in future phases.
- **Trace Streaming**: Optional integration with external telemetry sinks (HTTP POST) when security review complete.

This data model supplements Feature 001 documentation and should be read in conjunction with the original core entities to understand the full interpreter architecture.
