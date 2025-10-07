# Feature Specification: Viro Core Language and REPL

**Feature Branch**: `001-implement-the-core`  
**Created**: 2025-01-07  
**Status**: Draft  
**Input**: User description: "Implement the core of Viro language along with REPL"

## Clarifications

### Session 2025-01-07

- Q: How should the interpreter handle word identifier case and Unicode normalization? → A: Case-sensitive (like most languages)
- Q: Should the interpreter enforce time limits on expression evaluation to prevent infinite loops or extremely long-running code? → A: No automatic timeout
- Q: What language should error messages be in, and should the system support multiple languages? → A: English only
- Q: Should the interpreter provide built-in tracing/debugging output capabilities in this phase? → A: No tracing in Phase 1
- Q: What versioning scheme should the interpreter use? → A: Semantic versioning (v1.0.0)

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Evaluate Basic Expressions (Priority: P1)

Users can evaluate simple REBOL-style expressions in the REPL, including literals (numbers, strings), words (variables), and basic arithmetic operations. This forms the foundation of the interpreter and demonstrates the core evaluation engine is working.

**Why this priority**: The core evaluator (`Do_Next`, `Do_Blk`) is the foundation of the entire interpreter. Without expression evaluation, no other features can work. This delivers immediate value by allowing users to perform calculations and see results.

**Independent Test**: Can be fully tested by starting the REPL, entering expressions like `5`, `3 + 4`, `x: 10`, and `x`, and verifying correct output for each expression type.

**Acceptance Scenarios**:

1. **Given** REPL is started, **When** user enters a literal value `42`, **Then** system evaluates and displays `42`
2. **Given** REPL is started, **When** user enters a string `"hello"`, **Then** system evaluates and displays `"hello"`
3. **Given** REPL is started, **When** user enters arithmetic expression `3 + 4`, **Then** system evaluates and displays `7`
4. **Given** REPL is started, **When** user assigns `x: 10`, **Then** system stores the value and displays `10`
5. **Given** variable `x` is bound to `10`, **When** user enters `x`, **Then** system retrieves and displays `10`
6. **Given** user enters an undefined word, **When** evaluation occurs, **Then** system reports "no value" error with word name

---

### User Story 2 - Control Flow with Native Functions (Priority: P2)

Users can use control flow constructs (`if`, `either`, `loop`) to make decisions and repeat operations. This enables basic programming logic and demonstrates the native function dispatch system is working.

**Why this priority**: Control flow is essential for any useful program. It builds on P1's evaluator and adds native function evaluation, which is the next critical architecture layer. Users can now write simple programs, not just evaluate expressions.

**Independent Test**: Can be fully tested by entering control flow expressions like `if true [print "yes"]`, `either 1 < 2 ["less"] ["greater"]`, and `loop 3 [print "hi"]`, verifying each executes correctly.

**Acceptance Scenarios**:

1. **Given** REPL is running, **When** user enters `if true [print "yes"]`, **Then** system prints `yes`
2. **Given** REPL is running, **When** user enters `if false [print "yes"]`, **Then** system prints nothing
3. **Given** REPL is running, **When** user enters `either 1 < 2 ["less"] ["greater"]`, **Then** system evaluates and displays `"less"`
4. **Given** REPL is running, **When** user enters `loop 3 [print "hi"]`, **Then** system prints `hi` three times
5. **Given** control flow evaluates blocks, **When** nested expressions exist, **Then** system evaluates inner blocks recursively

---

### User Story 3 - Series Operations (Priority: P3)

Users can create and manipulate series (blocks, strings) using native functions like `first`, `last`, `append`, `insert`. This demonstrates the value type system is complete enough to handle composite data structures.

**Why this priority**: Series are fundamental REBOL data types. This validates that the type system and native dispatch handle complex types beyond primitives. Users can now work with collections, which is necessary for real programs.

**Independent Test**: Can be fully tested by creating series `data: [1 2 3]`, using operations like `first data`, `append data 4`, `insert data 0`, and verifying results match expected series values.

**Acceptance Scenarios**:

1. **Given** REPL is running, **When** user enters `data: [1 2 3]`, **Then** system creates a block series and binds it to `data`
2. **Given** block `data` contains `[1 2 3]`, **When** user enters `first data`, **Then** system displays `1`
3. **Given** block `data` contains `[1 2 3]`, **When** user enters `last data`, **Then** system displays `3`
4. **Given** block `data` contains `[1 2 3]`, **When** user enters `append data 4`, **Then** system modifies block to `[1 2 3 4]` and displays it
5. **Given** block `data` contains `[1 2 3]`, **When** user enters `insert data 0`, **Then** system modifies block to `[0 1 2 3]` and displays it
6. **Given** string operations needed, **When** user uses series functions on strings, **Then** system handles strings like character series

---

### User Story 4 - Function Definition and Calls (Priority: P4)

Users can define custom functions and call them with arguments. This demonstrates the frame and context system is working, including argument binding and closure support.

**Why this priority**: User-defined functions require the frame system (variable scoping, argument binding) to be operational. This is a major architecture milestone that enables code reuse and abstraction.

**Independent Test**: Can be fully tested by defining `square: func [n] [n * n]`, calling `square 5`, and verifying it returns `25`. Test with multiple arguments and nested calls.

**Acceptance Scenarios**:

1. **Given** REPL is running, **When** user enters `square: func [n] [n * n]`, **Then** system creates function and binds it to `square`
2. **Given** function `square` is defined, **When** user enters `square 5`, **Then** system evaluates function body with `n` bound to `5` and displays `25`
3. **Given** REPL is running, **When** user defines `add: func [a b] [a + b]` and calls `add 3 7`, **Then** system displays `10`
4. **Given** functions can call other functions, **When** user defines nested function calls, **Then** system maintains proper stack frames and returns correct results
5. **Given** function evaluation, **When** error occurs in function body, **Then** system reports error with function context

---

### User Story 5 - Error Handling and Recovery (Priority: P5)

Users encounter structured errors when invalid operations occur (syntax errors, undefined words, type mismatches), and the REPL remains usable after errors. This demonstrates the error system is operational and the REPL has proper error boundaries.

**Why this priority**: Error handling must work correctly from the start (per constitution), but it's naturally tested during previous user stories. This story formalizes error behavior expectations.

**Independent Test**: Can be fully tested by deliberately triggering errors (`undefined-word`, `1 / 0`, `+ "not a number"`), verifying each produces appropriate error messages, and confirming REPL continues accepting input.

**Acceptance Scenarios**:

1. **Given** REPL is running, **When** user enters undefined word `undefined-word`, **Then** system displays Script error (300): "No value for word" with context
2. **Given** REPL is running, **When** user enters `1 / 0`, **Then** system displays Math error (400): "Division by zero"
3. **Given** REPL is running, **When** user enters invalid syntax, **Then** system displays Syntax error (200) with error position
4. **Given** REPL is running, **When** user enters type mismatch like `+ "string" 5`, **Then** system displays Script error with type information
5. **Given** error occurs, **When** error is displayed, **Then** REPL remains operational and accepts next input
6. **Given** error occurs in function, **When** error is reported, **Then** system includes function call stack in error context

---

### User Story 6 - REPL Interactive Features (Priority: P6)

Users can interact with the REPL using standard features: command history (up/down arrows), multi-line input for blocks, and helpful feedback (prompts showing expression continuation). This makes the REPL practical for daily use.

**Why this priority**: These are quality-of-life features that make the REPL usable but aren't strictly necessary for core functionality. They build on the solid foundation of previous stories.

**Independent Test**: Can be fully tested by entering commands, using up arrow to recall, entering incomplete expressions and seeing continuation prompts, and entering multi-line blocks.

**Acceptance Scenarios**:

1. **Given** REPL is running, **When** user presses up arrow, **Then** system displays previous command
2. **Given** user enters incomplete expression `if true [`, **When** user presses enter, **Then** system shows continuation prompt `...` and waits for completion
3. **Given** user enters multi-line block, **When** user completes block with `]`, **Then** system evaluates entire block
4. **Given** REPL is running, **When** user enters `quit` or `exit`, **Then** REPL terminates gracefully
5. **Given** REPL is running, **When** user presses Ctrl+C during evaluation, **Then** system interrupts current evaluation and returns to prompt
6. **Given** REPL starts, **When** initialization complete, **Then** system displays welcome message with version information

---

### Edge Cases

- What happens when user enters deeply nested expressions (e.g., 100 levels of blocks)? System should handle recursion depth gracefully or report stack overflow error.
- What happens when user defines recursive function without base case? System should detect infinite recursion and terminate with error after reasonable depth.
- What happens when block contains mixed types in unexpected positions? System should validate types according to function signatures and report clear type errors.
- What happens when user tries to modify literal values or constants? System should prevent modification or return appropriate error.
- What happens when user enters non-UTF-8 characters? System should handle encoding errors gracefully.
- What happens when evaluation takes too long (potential infinite loop)? User must interrupt via Ctrl+C; no automatic timeout enforced.
- What happens when memory consumption grows large during series operations? System should handle memory allocation failures gracefully.

## Requirements *(mandatory)*

### Functional Requirements

#### Core Evaluation Engine

- **FR-001**: System MUST implement type-based dispatch that classifies values (words, functions, paths, literals) and routes them to appropriate evaluation handlers
- **FR-002**: System MUST support recursive evaluation for nested blocks and parenthesized expressions
- **FR-003**: System MUST evaluate expressions following REBOL semantics: literals evaluate to themselves, words evaluate to bound values, functions execute with arguments
- **FR-004**: System MUST maintain evaluation index through block sequences, advancing after each expression evaluation

#### Type System

- **FR-005**: System MUST support primitive value types: integer, string, word, block, none, logic (true/false)
- **FR-006**: System MUST support function types: native functions (built-in) and user-defined functions
- **FR-007**: System MUST distinguish between word types: word (evaluate), lit-word (quote), get-word (fetch), set-word (assign). Word identifiers are case-sensitive.
- **FR-008**: System MUST implement type checking for function arguments according to type specifications

#### Stack Management

- **FR-009**: System MUST implement unified stack for both data values and function call frames
- **FR-010**: System MUST use index-based (not pointer-based) stack access to prevent invalidation during expansion
- **FR-011**: System MUST automatically expand stack when capacity is reached
- **FR-012**: System MUST maintain stack frame layout: return value slot, prior frame pointer, function metadata, argument values

#### Frame & Context System

- **FR-013**: System MUST implement frames as fundamental unit for variable storage
- **FR-014**: Each frame MUST contain word list (variable names) and value list (corresponding data)
- **FR-015**: System MUST support binding operations: bind word to frame, get variable value, set variable value
- **FR-016**: System MUST implement proper frame types: function argument frames and closure frames (object/module frames deferred to later phases)

#### Native Functions (Minimal Set ~50 functions)

**Control Flow:**
- **FR-017**: System MUST provide `if` function: `if condition [block]` - executes block if condition is true
- **FR-018**: System MUST provide `either` function: `either condition [true-block] [false-block]` - conditional execution
- **FR-019**: System MUST provide `loop` function: `loop count [block]` - repeat block N times
- **FR-020**: System MUST provide `while` function: `while [condition] [block]` - repeat while condition is true

**Data Operations:**
- **FR-021**: System MUST provide `set` function: `set word value` - assigns value to word
- **FR-022**: System MUST provide `get` function: `get word` - retrieves value bound to word
- **FR-023**: System MUST provide `type?` function: `type? value` - returns datatype of value

**I/O Operations:**
- **FR-024**: System MUST provide `print` function: `print value` - outputs value to standard output
- **FR-025**: System MUST provide `input` function: `input` - reads line from standard input

**Math Operations:**
- **FR-026**: System MUST provide arithmetic operators: `+`, `-`, `*`, `/` for integer and decimal arithmetic
- **FR-027**: System MUST provide comparison operators: `<`, `>`, `<=`, `>=`, `=`, `<>` (not equal)
- **FR-028**: System MUST provide logic operators: `and`, `or`, `not`

**Series Operations:**
- **FR-029**: System MUST provide `first` function: returns first element of series
- **FR-030**: System MUST provide `last` function: returns last element of series
- **FR-031**: System MUST provide `append` function: adds element to end of series
- **FR-032**: System MUST provide `insert` function: adds element at current position in series
- **FR-033**: System MUST provide `length?` function: returns number of elements in series

**Function Definition:**
- **FR-034**: System MUST provide `func` function: `func [args] [body]` - creates user-defined function with argument list and body block

#### Error Handling

- **FR-035**: System MUST implement structured errors with category codes: Throw (0), Note (100), Syntax (200), Script (300), Math (400), Access (500), Internal (900)
- **FR-036**: Each error MUST include: error code, category, error ID, up to 3 arguments, near context (expression that caused error), where context (stack location). All error messages are in English.
- **FR-037**: System MUST report undefined word errors when evaluating unbound words
- **FR-038**: System MUST report type errors when function receives wrong argument types
- **FR-039**: System MUST report math errors for invalid operations (division by zero, overflow)
- **FR-040**: System MUST report syntax errors during parsing with position information
- **FR-041**: System MUST maintain interpreter state after errors (REPL continues running)

#### REPL Interface

- **FR-042**: System MUST provide interactive Read-Eval-Print Loop that accepts user input, evaluates expressions, and displays results
- **FR-043**: REPL MUST display prompt indicating ready state (e.g., `>>`)
- **FR-044**: REPL MUST display results of expression evaluation on separate line
- **FR-045**: REPL MUST support multi-line input for incomplete expressions, showing continuation prompt (e.g., `...`)
- **FR-046**: REPL MUST maintain command history accessible via up/down arrow keys
- **FR-047**: REPL MUST support exit commands (`quit`, `exit`) to terminate session
- **FR-048**: REPL MUST support interrupt signal (Ctrl+C) to cancel current evaluation and return to prompt. No automatic timeout is enforced; users rely on manual interruption.
- **FR-049**: REPL MUST display welcome message on startup with interpreter version using semantic versioning format (e.g., "Viro v1.0.0")
- **FR-050**: REPL MUST handle errors gracefully without crashing, displaying error messages and returning to prompt

### Key Entities

- **Value**: Core data representation with type tag and data payload. Contains type discriminator (integer, string, word, block, function, etc.) and associated data. All values are immutable except series.

- **Block**: Ordered sequence of values. Fundamental composite type in REBOL. Supports indexing, iteration, and modification operations. Evaluable as code or used as data.

- **Word**: Symbolic identifier that references a value in a context. Four forms: word (evaluate), lit-word (quote), get-word (fetch), set-word (assign). Bound to frame during evaluation. Case-sensitive: `x` and `X` are distinct identifiers.

- **Function**: Executable entity with argument specification and body block. Native functions are built-in, user functions are created via `func`. Contains formal parameter list, body block, and optional type constraints.

- **Frame**: Variable storage container. Maps word symbols to values. Created for function calls (local variables + arguments), objects (deferred), and modules (deferred). Includes word list and value list.

- **Stack**: Unified storage for values and call frames. Supports push/pop operations, automatic expansion, index-based access. Contains data values between calls and frame structures during calls.

- **Error**: Structured error representation. Contains category code, error ID, up to 3 arguments for message interpolation, near context (source expression), where context (call stack), and formatted message.

- **Series**: Abstract sequence type. Blocks and strings are series. Supports positional operations (first, last, at), modification operations (append, insert, remove), and length queries.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can evaluate at least 20 different expression types correctly (literals, arithmetic, comparisons, control flow, series operations, function calls)

- **SC-002**: REPL session remains stable for continuous operation exceeding 1000 evaluation cycles without memory leaks or crashes

- **SC-003**: Error messages include sufficient context (error category, expression location, relevant values) that users can diagnose and fix issues in under 2 minutes for common errors

- **SC-004**: Users can define and call recursive functions up to depth of 100 without stack overflow

- **SC-005**: Evaluation performance supports interactive use: simple expressions (literals, arithmetic) complete in under 10 milliseconds, complex expressions (nested function calls) complete in under 100 milliseconds on standard hardware

- **SC-006**: Type error detection catches at least 95% of type mismatches before execution (during argument validation phase)

- **SC-007**: Command history supports at least 100 previous commands accessible via arrow keys

- **SC-008**: Multi-line input correctly handles nested blocks up to 10 levels deep without confusion about continuation state

- **SC-009**: Stack expansion occurs transparently without user intervention or noticeable delay (under 1 millisecond for typical expansion)

- **SC-010**: Users can successfully interrupt long-running evaluations via Ctrl+C and return to working REPL prompt in under 500 milliseconds

## Assumptions

1. **Target platform**: Development focuses on macOS initially (per workspace context), but design remains cross-platform compatible
2. **Go version**: Assumes Go 1.21+ available per constitution (for generics and improved error handling)
3. **Unicode support**: UTF-8 encoding assumed for string handling (REBOL standard)
4. **Memory model**: Assumes sufficient memory for typical REPL sessions (defined as supporting programs up to 10,000 values in memory simultaneously)
5. **Performance baseline**: "Interactive" performance defined as response under 100ms for 95th percentile of operations
6. **Arithmetic precision**: Integer arithmetic uses 64-bit signed integers; decimal arithmetic deferred to future phase
7. **Series implementation**: Blocks use slice-based implementation with copy-on-write semantics where practical
8. **Native function count**: Initial implementation targets ~50 natives covering essential operations; full 600+ function library is long-term goal
9. **Parse dialect**: Complex pattern-matching DSL deferred to later phase per constitution principle
10. **Module system**: Object/module frames deferred to later phase; current implementation focuses on function frames only
11. **REPL library**: Standard Go libraries for terminal handling (readline-style libraries acceptable for command history)
12. **Error recovery**: REPL maintains global context across errors; does not reset interpreter state on error (only abandons current evaluation)
13. **Versioning**: Semantic versioning (MAJOR.MINOR.PATCH) used for interpreter releases, consistent with constitution versioning policy

## Out of Scope

The following features are explicitly **not** included in this phase and will be addressed in future phases:

- **Tracing/observability**: Built-in trace flags, evaluation step output, verbose logging (developers rely on external debuggers in Phase 1)
- **Parse dialect**: Pattern-matching DSL for string and block parsing (complex subsystem per design specification)
- **Module system**: Module loading, imports, exports, and module contexts
- **Object system**: Object frames, object creation, and method dispatch
- **File I/O**: Reading from and writing to files (beyond REPL stdio)
- **Network operations**: HTTP, sockets, or other network protocols
- **Advanced math**: Trigonometric functions, logarithms, advanced numerical operations
- **Decimal numbers**: Floating-point arithmetic (integers only in this phase)
- **Date/time**: Date and time datatypes and operations
- **Path expressions**: Path navigation through nested structures (`object/field`)
- **Refinements**: Function refinements for optional parameters (`func/refine`)
- **Advanced series**: Skip, take, copy, find, and other complex series operations beyond basic first/last/append/insert
- **Port system**: Asynchronous I/O and port abstraction
- **Reflection**: Runtime introspection of function signatures, stack inspection
- **Debugging features**: Breakpoints, step execution, variable watches (beyond basic trace output)
- **Optimization**: Performance optimization, bytecode compilation, JIT (correctness first per constitution)
- **Extension system**: C FFI or plugin architecture for extending interpreter

These features may be specified and implemented in subsequent phases after core architecture is validated.
