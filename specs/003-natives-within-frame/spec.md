# Feature Specification: Natives Within Frame

**Feature Branch**: `003-natives-within-frame`  
**Created**: 2025-10-12  
**Status**: Draft  
**Input**: User description: "natives within frame - Currently the native functions are stored in a special registry and each word lookup first check the native registry and only then checks the frame. It also prevents Viro users from overwriting the natives within frames and reusing the same names which sometimes may lead to confusing errors (`--debug` refinement would collide with native `debug` function within the function scope). Since native functions are represented by `FunctionValue` there is no need for special registry and the functions may be registered in the root frame."

## Clarifications

### Session 2025-10-12

- Q: Root frame mutability strategy - should native functions be immutable, warn on rebinding, or fully mutable? → A: root frame is mutable as everything else, no need to enforce anything
- Q: Native registration timing - when should natives be registered in root frame? → A: During evaluator construction (NewEvaluator)
- Q: Closure capture behavior when native is later shadowed - does closure see original or shadowed value? → A: Closure sees the original native (lexical capture at closure creation time)
- Q: Native registration error handling - what happens if registration fails during construction? → A: Panic/fatal error - evaluator construction fails immediately
- Q: Multi-library shadowing conflicts - what happens when multiple libraries shadow the same native? → A: Standard lexical scoping - innermost binding wins per scope chain

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Define Functions with Names Matching Natives (Priority: P1)

Developers writing Viro code need to define local variables or functions that have the same names as native functions (like `debug`, `trace`, `type`, etc.) without causing naming conflicts or errors. This is essential for natural code expression where common words are used in different contexts.

**Why this priority**: This is the primary motivation for the feature - it directly addresses the naming conflict issue that currently causes confusing errors. Without this, developers must avoid entire categories of common names, which is a significant usability problem.

**Independent Test**: Can be fully tested by writing a function that uses `--debug` refinement and defining a local `debug` variable, then verifying both work correctly without conflicts.

**Acceptance Scenarios**:

1. **Given** a function with a `--debug` refinement parameter, **When** the function is called with that refinement, **Then** the refinement should work correctly without conflicting with the native `debug` function
2. **Given** a local variable named `type` defined in a function, **When** that variable is accessed, **Then** it should return the local value, not invoke the native `type?` function
3. **Given** nested scopes where both define a variable with the same name as a native, **When** the inner scope accesses that name, **Then** it should resolve to the innermost binding following standard lexical scoping rules

---

### User Story 2 - Shadow Native Functions Intentionally (Priority: P2)

Advanced Viro developers need to temporarily override or wrap native functions within specific scopes to customize behavior, add logging, implement proxies, or create domain-specific variants while maintaining access to the original native in outer scopes.

**Why this priority**: This enables advanced patterns like aspect-oriented programming, debugging wrappers, and domain-specific language extensions. While less critical than basic naming freedom, it's a powerful capability for library authors and framework builders.

**Independent Test**: Can be tested by defining a custom `print` function that wraps the native `print`, then verifying both the custom and native versions work in their respective scopes.

**Acceptance Scenarios**:

1. **Given** a user-defined function named `print` in a local scope, **When** `print` is called within that scope, **Then** the user-defined version executes instead of the native
2. **Given** a wrapped native function that calls the original, **When** the wrapper is invoked, **Then** both the wrapper code and the original native execute correctly
3. **Given** a shadowed native in an inner scope and native access needed in that scope, **When** the native is accessed using scope resolution, **Then** the original native remains accessible

---

### User Story 3 - Consistent Word Resolution Order (Priority: P1)

All Viro code should follow a single, predictable word resolution strategy where lookups check the current frame first, then parent frames, without special cases for different word categories. This ensures the language behaves consistently and intuitively.

**Why this priority**: Consistency in language semantics is fundamental for developer understanding and preventing subtle bugs. Without this, the language has two different resolution rules (one for natives, one for user-defined names), which violates the principle of least surprise.

**Independent Test**: Can be tested by creating multiple nested scopes with overlapping names and verifying resolution always follows the lexical scoping chain without special native-first checks.

**Acceptance Scenarios**:

1. **Given** a word that exists both as a native and in a frame, **When** that word is looked up, **Then** the frame value is resolved first (standard lexical scoping)
2. **Given** a word that doesn't exist in any frame, **When** that word is looked up, **Then** an error is raised indicating an unbound word (no fallback to a special registry)
3. **Given** multiple frames in a scope chain, **When** a word is looked up, **Then** resolution proceeds from innermost to outermost frame in standard lexical order

---

### Edge Cases

- Natives are guaranteed to exist in root frame after NewEvaluator returns (registered during construction)
- Users can overwrite or remove natives at the root frame level (standard mutability) - this is intentional flexibility
- Closures use lexical capture semantics - if a closure captures a native and that name is later shadowed in an inner scope, the closure still references the original native from its creation environment
- Native registration failures during NewEvaluator cause immediate panic/fatal error - evaluator construction fails fast ensuring the system never runs in a partially-initialized state
- Multiple libraries can shadow the same native independently - standard lexical scoping applies where innermost binding wins in each scope chain, no conflict detection or special handling needed

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST store all native functions as `FunctionValue` instances in the root frame instead of a separate registry
- **FR-002**: System MUST allow users to define local variables or functions with names matching native functions without errors
- **FR-003**: System MUST follow standard lexical scoping rules for all word lookups, checking the current frame before parent frames
- **FR-004**: System MUST NOT special-case native function lookups - they follow the same resolution path as user-defined functions
- **FR-005**: System MUST initialize the root frame with all native functions during evaluator construction (NewEvaluator), ensuring natives are available before any user code executes
- **FR-006**: System MUST support shadowing native functions in local scopes without affecting the native definitions in outer scopes
- **FR-007**: System MUST maintain backward compatibility - existing code that doesn't shadow natives continues to work identically
- **FR-008**: System MUST preserve native function metadata (documentation, parameter specs, infix flags) when stored in the root frame
- **FR-009**: Root frame bindings (including native functions) are mutable following standard frame semantics - users can rebind or overwrite natives at the root level if needed
- **FR-010**: System MUST eliminate the `native.Registry` global variable and the `native.Lookup()` function once migration is complete

### Key Entities

- **Root Frame**: The outermost lexical scope where all native functions are registered as initial bindings
- **Native Function**: A built-in function implemented in Go, represented as a `FunctionValue` with type `FuncNative`
- **Frame Binding**: A word-to-value mapping within a frame that follows lexical scoping rules
- **Word Lookup**: The process of resolving a word to its value by searching through the frame chain

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can define functions with `--debug` refinement parameters without collision errors, validated by test cases exercising refinement and native function independently
- **SC-002**: All existing Viro code and test suites pass without modification, demonstrating backward compatibility
- **SC-003**: Word lookups follow a single, consistent resolution strategy with no special-case logic for natives, verified by inspecting evaluator code for elimination of `native.Lookup()` calls
- **SC-004**: Native functions remain accessible throughout execution unless explicitly shadowed, confirmed by baseline test suite continuing to pass
- **SC-005**: Lexical scoping behavior is consistent across all word types (natives, user functions, variables), demonstrated by test cases with nested scopes and shadowing
