<!--
Sync Impact Report:
Version Change: Template → 1.0.0
Constitution Type: Initial ratification for Viro REBOL-inspired interpreter project

Added Sections:
- All core principles (7 principles defined)
- Implementation Architecture section
- Quality & Reliability section
- Governance section

Modified Principles: N/A (initial creation)

Templates Status:
✅ plan-template.md - Reviewed, TDD principles align
✅ spec-template.md - Reviewed, user story approach compatible
✅ tasks-template.md - Reviewed, test-first task ordering compatible
✅ All command prompts in .github/prompts/ - Generic guidance maintained

Follow-up TODOs: None - all placeholders resolved
-->

# Viro Interpreter Constitution

## Core Principles

### I. Test-Driven Development (NON-NEGOTIABLE)

**Tests MUST be written before implementation.** Every feature follows the strict Red-Green-Refactor cycle:
1. Write failing tests that define the expected behavior
2. Obtain user approval on test scenarios
3. Verify tests fail for the right reasons
4. Implement minimal code to make tests pass
5. Refactor with confidence knowing tests protect correctness

**Rationale**: The REBOL interpreter is a complex system with type dispatch, stack management, and frame contexts. TDD ensures each component works correctly in isolation and integration before building the next layer. This is critical for an interpreter where subtle bugs can cascade.

### II. Incremental Implementation by Architecture Layer

Implementation MUST follow the layered architecture recommended in the design specification:
1. **Core Evaluator** (`Do_Next`, `Do_Blk`) - Foundation first
2. **Type System** - Value types and evaluation strategies
3. **Minimal Native Set** (~50 essential natives: control flow, basic I/O, data operations)
4. **Frame & Context System** - Variable binding and function calls
5. **Error Handling** - Structured errors from the start
6. **Extended Natives** - Expand function library incrementally
7. **Parse Dialect** - Complex subsystem added last

Each layer MUST be fully tested and operational before proceeding to the next.

**Rationale**: The design specification explicitly warns that the interpreter is "highly tuned code that should only be modified by experts." Building incrementally with tests at each layer prevents cascading failures and maintains system integrity.

### III. Type-Based Dispatch Fidelity

The interpreter MUST implement faithful type-based dispatch mirroring REBOL R3 behavior:
- Each value type (words, functions, paths, literals) has specific evaluation rules
- Evaluation type maps MUST route to appropriate handlers
- Stack-based execution MUST be preserved for data and control flow
- Frame-based variable binding MUST be maintained (not global namespace)

**Rationale**: REBOL's power comes from homoiconicity and consistent type semantics. Deviation from type-based dispatch breaks the fundamental execution model.

### IV. Stack and Frame Safety

Stack and frame management MUST prioritize safety:
- Use index-based access (not pointer-based) to prevent invalidation on expansion
- Implement automatic stack expansion with proper bounds checking
- Maintain stack frame layout: return value slot, prior frame info, function metadata, arguments
- Frame types (objects, modules, function arguments, closures) MUST be distinct and properly isolated

**Rationale**: The design specification emphasizes safety through index-based access. Interpreter crashes due to stack corruption are catastrophic and difficult to debug.

### V. Error Categories and Structured Errors

Error handling MUST follow the REBOL error category system (0-900 range):
- **Throw (0)**: Loop control errors
- **Note (100)**: Warnings
- **Syntax (200)**: Parsing errors
- **Script (300)**: Runtime errors
- **Math (400)**: Arithmetic errors
- **Access (500)**: I/O/security errors
- **Internal (900)**: System errors

Each error MUST include: code, type, id, arg1-3, near, where context.

**Rationale**: Structured errors enable precise error handling and debugging. The numeric categorization allows programmatic error filtering and recovery strategies.

### VI. Observable Behavior Through Text I/O

Where practical, interpreter operations SHOULD expose text-based I/O:
- Trace flags for evaluation steps (when enabled)
- Human-readable error messages with context
- Debug output for frame dumps and value inspection
- Optional verbose logging for stack operations

**Rationale**: Text I/O ensures debuggability. Interpreters are black boxes; observable execution traces are essential for diagnosing evaluation issues.

### VII. Simplicity and YAGNI (You Aren't Gonna Need It)

Start with minimal necessary features and expand only when needed:
- Implement ~50 core natives before considering the full 600+
- Skip complex features (Parse dialect, module system) until core interpreter is solid
- Avoid premature optimization - correctness first, performance later
- Each native function MUST have a clear, justified use case

**Rationale**: The design specification recommends starting small and expanding incrementally. Over-engineering early creates maintenance burden and delays core functionality.

## Implementation Architecture

### Language and Tooling

- **Language**: Go (as specified)
- **Testing Framework**: Go standard library `testing` package
- **Minimum Go Version**: 1.21 or later (for generics and improved error handling)
- **Code Organization**: Layered modules matching architecture phases
- **Build Tool**: Standard `go build` / `go test`

### Source Structure

```
viro/
├── docs/                    # Design specifications
│   └── interpreter.md
├── internal/
│   ├── value/              # Value types and type system
│   ├── eval/               # Core evaluator (Do_Next, Do_Blk)
│   ├── stack/              # Stack management
│   ├── frame/              # Frame & context system
│   ├── native/             # Native function implementations
│   ├── error/              # Structured error system
│   └── parse/              # Parse dialect (later phase)
├── pkg/
│   └── viro/               # Public API
├── cmd/
│   └── viro/               # CLI entry point
└── test/
    ├── contract/           # Contract tests for native functions
    ├── integration/        # End-to-end interpreter tests
    └── fixtures/           # Test REBOL scripts
```

### Dependency Constraints

- **Minimize external dependencies**: Prefer Go standard library
- **No reflection for core evaluation**: Type dispatch MUST use explicit type switching for performance
- **Allowed dependencies**: Testing utilities, benchmarking tools, optional debugging libraries
- **Forbidden dependencies**: Interpreter frameworks that impose their own evaluation model

## Quality & Reliability

### Test Coverage Requirements

- **Minimum 80% code coverage** for core evaluator, stack, and frame systems
- **100% coverage** for error handling paths (all error categories must be tested)
- **Contract tests** for every native function specifying input/output behavior
- **Integration tests** covering multi-step evaluation scenarios

### Performance Baselines

Performance MUST NOT degrade without explicit justification:
- **Stack operations**: O(1) access, amortized O(1) expansion
- **Frame lookup**: O(1) for direct bindings, O(depth) for nested contexts
- **Value dispatch**: O(1) type classification and routing
- **Memory**: Reasonable memory usage for typical REBOL programs (define "reasonable" during Phase 1)

Benchmarks MUST be established during Phase 1 (Core Evaluator) and maintained thereafter.

### Code Review Requirements

All code changes MUST:
1. Include tests (written first, failing initially)
2. Pass existing test suite
3. Include architectural justification for new abstractions
4. Document deviations from REBOL R3 behavior (if any)
5. Update relevant design documentation

## Governance

### Constitution Authority

This constitution supersedes all other development practices and guidelines. When conflicts arise between this constitution and external guidance, the constitution takes precedence.

### Amendment Process

1. **Proposal**: Document proposed change with rationale and impact analysis
2. **Review**: Evaluate against project goals and architectural principles
3. **Approval**: Requires explicit approval (define approval authority during project initialization)
4. **Migration**: Update affected code, tests, and documentation
5. **Version Bump**: Increment version according to semantic versioning rules below

### Versioning Policy

Constitution version follows semantic versioning:
- **MAJOR**: Backward incompatible changes (e.g., removing a core principle, changing TDD requirement)
- **MINOR**: New principles or sections added (e.g., adding security requirements)
- **PATCH**: Clarifications, wording improvements, non-semantic refinements

### Compliance Review

All pull requests MUST verify compliance with:
- Test-first development (failing tests before implementation)
- Layered architecture adherence (no skipping to advanced features)
- Type system and stack safety requirements
- Error handling coverage
- Code review checklist completion

**Complexity Justification**: Any violation of core principles (e.g., implementing features out of order, skipping tests) MUST be explicitly justified with:
- Why the violation is necessary
- What simpler alternative was considered and rejected
- Mitigation plan to restore compliance

### Development Guidance

Runtime development guidance for AI agents and developers should reference `docs/interpreter.md` for architectural details. This constitution provides non-negotiable principles; the design specification provides implementation patterns.

**Version**: 1.0.0 | **Ratified**: 2025-01-07 | **Last Amended**: 2025-01-07