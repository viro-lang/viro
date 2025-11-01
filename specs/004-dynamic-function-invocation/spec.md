# Feature Specification: Dynamic Function Invocation (Action Types)

**Feature Branch**: `004-dynamic-function-invocation`
**Created**: 2025-10-13
**Status**: Draft
**Input**: User description: "dynamic function invocation - research a way to implement type-based dynamic dispatch for native and user-defined types"

## Clarifications

### Session 2025-10-13

- Q: When an action is invoked with zero arguments (no first argument to determine type for dispatch), what should the interpreter do? → A: Actions follow normal function call semantics for parameter validation (including zero/insufficient arguments)
- Q: When an action dispatches based on the first argument's type (e.g., `append [1 2] 3`), how should the system validate or constrain subsequent arguments? → A: Dispatch occurs only on first argument type; subsequent argument validation is delegated to the type-specific implementation
- Q: When are type frames (e.g., block! frame, string! frame) created and initialized with their type-specific function implementations? → A: All type frames are created and populated at interpreter startup (eager initialization)

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Type-Safe Series Operations (Priority: P1)

A language user wants to apply the same operation name to different data types and have the interpreter automatically dispatch to the correct type-specific implementation. For example, calling `first` on a block, string, or future user-defined series types should work seamlessly without the user needing to know which specific function variant to call.

**Why this priority**: This is the foundational behavior that enables polymorphism in Viro. Without it, users would need separate function names for each type (e.g., `first-block`, `first-string`), which breaks the homoiconic design philosophy and makes the language less intuitive.

**Independent Test**: Can be fully tested by defining a single action (e.g., `first`) with type-specific implementations for blocks and strings, then verifying that `first [1 2 3]` and `first "hello"` both execute correctly and deliver the appropriate first element.

**Acceptance Scenarios**:

1. **Given** an action `first` is defined with implementations for block and string types, **When** user executes `first [1 2 3]`, **Then** the interpreter dispatches to the block-specific `first` implementation and returns `1`
2. **Given** an action `first` is defined with implementations for block and string types, **When** user executes `first "hello"`, **Then** the interpreter dispatches to the string-specific `first` implementation and returns `"h"`
3. **Given** an action `append` is defined for multiple types, **When** user executes `append [1 2] 3`, **Then** the block is modified in-place and returns `[1 2 3]`
4. **Given** an action `length?` is defined for series types, **When** user executes `length? "test"`, **Then** the string-specific implementation returns `4`

---

### User Story 2 - Extensibility for Future User-Defined Types (Priority: P2)

A future language extension will allow users to define custom data types with their own behavior. The action dispatch system must be designed so that when user-defined types are eventually implemented, they can participate in the same polymorphic dispatch without requiring changes to the core dispatch mechanism.

**Why this priority**: While user-defined types are not implemented yet, the dispatch architecture must support them. This ensures that when Phase 3 adds objects or custom types, the action system doesn't require a complete redesign.

**Independent Test**: Can be verified by examining the design and ensuring that type-specific function frames are not hardcoded to native types only. A test could stub a hypothetical user-defined type and verify the dispatch logic would handle it identically to native types.

**Acceptance Scenarios**:

1. **Given** the action dispatch system is designed with type frames, **When** evaluating the architecture, **Then** the system must not hardcode native types and must allow registration of new type frames
2. **Given** a hypothetical user-defined type `custom!`, **When** evaluating dispatch logic, **Then** the same lookup mechanism used for native types would work for custom types
3. **Given** documentation of the action system, **When** reviewed by developers, **Then** it clearly explains how future user-defined types can register their own function frames

---

### User Story 3 - Error Handling for Missing Type Implementations (Priority: P3)

When a user calls an action on a value whose type doesn't have an implementation for that action, the interpreter must provide a clear, helpful error message indicating which action was called, on what type, and that no implementation exists.

**Why this priority**: Good error messages are essential for developer experience, but the core dispatch functionality is more critical to establish first. This can be implemented after the basic dispatch works.

**Independent Test**: Can be tested by defining an action with implementations for only some types (e.g., `first` for blocks but not integers), then calling it on an unsupported type and verifying the error message is clear and actionable.

**Acceptance Scenarios**:

1. **Given** an action `first` has no implementation for integer type, **When** user executes `first 42`, **Then** the interpreter returns an error like "Action 'first' not defined for type integer!"
2. **Given** an action `append` has implementations for blocks and strings but not integers, **When** user executes `append 100 5`, **Then** the error message indicates that `append` cannot operate on integers
3. **Given** a user calls a non-existent action, **When** evaluating `nonexistent [1 2]`, **Then** the error indicates the word is undefined (not a type mismatch)

---

### Edge Cases

- When an action is called with zero or insufficient arguments, the existing function parameter validation mechanism applies (same arity error handling as regular functions)
- Actions dispatch only on the first argument's type; subsequent argument validation is handled by the type-specific implementation (e.g., block-append validates that the second argument is appropriate for appending to a block)
- If a user shadows an action name with a local variable in their scope, the local binding takes precedence (FR-008 lexical scoping)
- Action refinements are passed through to the type-specific implementation unchanged (per Assumptions)
- Type frames are initialized at interpreter startup and are not modified at runtime (eager initialization, static structure)
- The dispatch system uses the existing frame chain for lexical scoping (actions live in root frame, subject to shadowing)

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST introduce a new value type `action!` that represents a polymorphic function capable of dispatching based on the type of its first argument
- **FR-002**: System MUST maintain separate function frames for each data type (block!, string!, integer!, etc.) containing type-specific implementations, created and populated at interpreter startup
- **FR-003**: When an action is invoked, the evaluator MUST look up the first argument's type, find the corresponding type frame, and execute the type-specific function (validation of subsequent arguments is delegated to the type-specific implementation)
- **FR-004**: Actions MUST support the same parameter specifications as regular functions (evaluated args, unevaluated args, refinements) and use the same parameter validation mechanism (arity checking occurs before dispatch)
- **FR-005**: The dispatch mechanism MUST NOT hardcode native types - it must use a registry or mapping that can be extended
- **FR-006**: When an action is called on a type with no implementation, the system MUST generate a clear script error indicating the action name and unsupported type
- **FR-007**: Type-specific function frames MUST be accessible in a uniform way that does not distinguish between native types and future user-defined types
- **FR-008**: Actions MUST respect lexical scoping - if a user shadows an action name with a local binding, the local binding takes precedence
- **FR-009**: All existing series operations (first, last, append, insert, length?, etc.) MUST be converted from direct native functions to actions with type-specific implementations
- **FR-010**: The action dispatch system MUST integrate with the existing evaluator without requiring changes to the core evaluation loop (type-based dispatch)
- **FR-011**: Type frames MUST support multiple functions per type (e.g., block! frame has first, last, append, insert, etc.)
- **FR-012**: The system MUST document the action creation and registration process clearly for future extensibility

### Key Entities

- **Action**: A polymorphic function value that dispatches to type-specific implementations based on its first argument's type. Contains only a name and parameter specification; dispatch uses the global TypeRegistry to locate type frames.
- **Type Frame**: A regular frame containing functions that operate on a specific value type (e.g., a "block! frame" with first, last, append functions). Uses standard frame mechanism (Words/Values arrays) with no special fields. Stored in TypeRegistry (not on stack), with Index=-1 and Parent=0 (points to root frame on stack).
- **Type Registry**: A global mapping from value types (TypeBlock, TypeString, etc.) to type frame pointers. Stores type frames directly (not indices). Initialized at startup and used by all actions to locate the correct type frame during dispatch.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: All existing series operations in the standard library can be invoked on appropriate types and dispatch correctly (100% of current series functions work via action dispatch)
- **SC-002**: Language users can call the same operation name on different types without type-specific function name prefixes (e.g., `first [1 2]` and `first "ab"` both work)
- **SC-003**: Test coverage demonstrates that adding a new type-specific implementation requires only registering the function in the type frame, with no changes to the dispatch logic
- **SC-004**: Error messages for unsupported type operations clearly identify the action name and unsupported type in 100% of test cases
- **SC-005**: Performance overhead of action dispatch compared to direct function calls is measurable and acceptable (to be defined during planning)

## Assumptions

- The action dispatch system will follow Viro's existing local-by-default scoping rules (actions in root frame can be shadowed by local bindings)
- Actions always dispatch based on the first argument's type (not multiple dispatch based on all arguments)
- Type frames are implemented as regular frames in the stack, maintaining the existing index-based architecture
- Refinements on actions are passed through to the type-specific implementation unchanged
- The action type will be implemented similarly to function!, with evaluation behavior defined by the evaluator
- Multi-method dispatch (dispatching on multiple argument types) is out of scope for this feature
- Actions do not modify the current type-based dispatch architecture, they layer on top of it
- Migration of existing native functions to actions can be done incrementally without breaking existing code

## Dependencies

- Requires existing frame system (internal/frame/) to store type-specific function frames
- Requires existing value type system (internal/value/) to add the new action! type
- Requires existing evaluator (internal/eval/) to handle action invocation
- Requires existing native function registration system to be refactored to register into type frames instead of root frame
- No external dependencies

## Out of Scope

- Implementation of user-defined types (this feature only prepares the architecture for future user-defined types)
- Multiple dispatch (dispatching based on types of multiple arguments, not just the first)
- Dynamic modification of type frames at runtime (type frame structure is eagerly initialized at interpreter startup and remains static)
- Performance optimization beyond basic dispatch (profiling and optimization can be addressed later)
- Conversion of non-series native functions to actions (only series operations initially)
- Type inference or automatic type promotion
- Macro or compile-time dispatch optimization

## Open Questions

None at this time. The feature description provides sufficient detail about the suggested implementation approach (type frames + action! type), and reasonable defaults can be applied for unspecified details.
