# Contract: Action Dispatch

**Component**: Action Dispatch System
**Version**: 1.0
**Date**: 2025-10-13

## Purpose

Defines the contract for polymorphic action dispatch based on the first argument's type. Actions are invoked with arguments, and the interpreter automatically routes execution to the type-specific implementation.

---

## Function: Action Invocation

### Signature
```viro
<action-name> <first-arg> [<additional-args>...] [<refinements>...]
```

### Description
When an action is invoked, the evaluator:
1. Evaluates the first argument (if marked as evaluated in param spec)
2. Determines the first argument's type
3. Looks up the type in the global TypeRegistry to get the type frame index
4. Resolves the type-specific function from the corresponding type frame using the action's name
5. Invokes the function with all arguments and refinements

### Parameters
| Name | Type | Evaluation | Description |
|------|------|------------|-------------|
| `action-name` | action! | Literal | The action to invoke (e.g., `first`, `append`) |
| `first-arg` | any | Evaluated (typically) | The value whose type determines dispatch |
| `additional-args` | any | Evaluated/Unevaluated per spec | Additional arguments passed to type-specific impl |
| `refinements` | refinement | Literal | Optional refinements (e.g., `/only`) |

### Returns
- **Type**: Depends on type-specific implementation
- **Value**: Result of executing the type-specific function

### Error Conditions

| Error Code | Category | Condition | Example |
|------------|----------|-----------|---------|
| `action-no-impl` | Script (341) | First argument's type not supported by action | `first 42` when first doesn't support integers |
| `wrong-arity` | Script (310) | Incorrect number of arguments for action | `first` with no arguments |
| `action-frame-corrupt` | Internal (941) | Type frame exists but function missing (bug) | Type registered but implementation not in frame |

### Examples

#### Example 1: Block first
```viro
>> first [1 2 3]
== 1
```

**Dispatch Flow**:
1. Action: `first`
2. First arg type: `block!`
3. Type frame lookup: Block frame (index 1)
4. Function lookup: Block-specific `first` in frame 1
5. Invocation: `blockFirst([1 2 3])`
6. Result: `1`

#### Example 2: String first
```viro
>> first "hello"
== "h"
```

**Dispatch Flow**:
1. Action: `first`
2. First arg type: `string!`
3. Type frame lookup: String frame (index 2)
4. Function lookup: String-specific `first` in frame 2
5. Invocation: `stringFirst("hello")`
6. Result: `"h"`

#### Example 3: Unsupported type error
```viro
>> first 42
** Script error: Action 'first' not defined for type integer!
```

**Dispatch Flow**:
1. Action: `first`
2. First arg type: `integer!`
3. Type frame lookup: **Not found in TypeFrames map**
4. Error: `verror.NewScriptError("action-no-impl", ["first", "integer!", ""], ...)`

#### Example 4: Arity error
```viro
>> first
** Script error: Wrong number of arguments for 'first' (expected 1, got 0)
```

**Dispatch Flow**:
1. Action: `first`
2. Argument consumption: **No arguments available**
3. Arity validation: **Fails (expected 1, got 0)**
4. Error: `verror.NewScriptError("wrong-arity", ["first", "1", "0"], ...)`

---

## Invariants

### Pre-Dispatch Invariants
1. Action value must exist in current scope (word lookup succeeds)
2. Global TypeRegistry must be initialized with type-to-frame mappings
3. Action's ParamSpec must define at least one parameter

### Dispatch Invariants
1. If first argument's type is in TypeRegistry, corresponding frame index must be valid
2. Type frame must exist at the specified index
3. Type frame may or may not contain a function binding with the action's name (determines if action is supported for that type)
4. Type-specific function's signature must match action's parameter specification

### Post-Dispatch Invariants
1. If dispatch succeeds, type-specific function is invoked with exact arguments provided
2. If dispatch fails, clear error message indicates action name and unsupported type
3. Stack and frame state remain consistent (no corruption from failed dispatch)

---

## Test Cases

### Test 1: Successful Dispatch - Block
**Input**:
```viro
first [1 2 3]
```
**Expected Output**: `1`
**Validation**: First element of block returned

### Test 2: Successful Dispatch - String
**Input**:
```viro
first "hello"
```
**Expected Output**: `"h"`
**Validation**: First character of string returned

### Test 3: Multiple Arguments - Append
**Input**:
```viro
b: [1 2]
append b 3
b
```
**Expected Output**: `[1 2 3]`
**Validation**: Block modified in-place, element appended

### Test 4: Unsupported Type Error
**Input**:
```viro
first 42
```
**Expected Error**: Script error with code 341, message "Action 'first' not defined for type integer!"
**Validation**: Error clearly identifies action and unsupported type

### Test 5: Arity Error - Zero Arguments
**Input**:
```viro
first
```
**Expected Error**: Script error with code 310, message "Wrong number of arguments for 'first'"
**Validation**: Standard arity validation applies before dispatch

### Test 6: Arity Error - Insufficient Arguments
**Input**:
```viro
append [1 2]
```
**Expected Error**: Script error with code 310, message indicates missing argument
**Validation**: Type-specific function's arity enforced

### Test 7: Refinement Passthrough
**Input**:
```viro
append/only [1 2] [3 4]
```
**Expected Output**: `[1 2 [3 4]]`
**Validation**: Refinement passed to type-specific implementation unchanged

### Test 8: Shadowing Action with Local Binding
**Input**:
```viro
first: fn [x] [x * 2]
first 5
```
**Expected Output**: `10`
**Validation**: Local binding shadows action (lexical scoping respected)

---

## Performance Contract

**Dispatch Overhead**:
- Type lookup in global TypeRegistry: O(1) - direct pointer retrieval
- Function lookup in type frame: O(1) - standard frame.Get()
- Total overhead: Constant time, estimated 2-5x slower than direct function call
- **Improvement**: One fewer indirection compared to index-based type frame storage

**Memory Usage**:
- DispatchContext: Stack-allocated, ~48 bytes (ephemeral, not persisted)
- Action value: ~50 bytes per action (just Name + ParamSpec)
- Type frame: ~100 bytes + (functions × 50 bytes), stored in TypeRegistry
- Global TypeRegistry: ~80-160 bytes (~10 types × 8-16 bytes per pointer)

**Scalability**:
- Adding new type: O(1) - create frame, register in TypeRegistry, add functions
- Adding new action: O(1) - create action, add implementation to relevant type frames
- Runtime dispatch: O(1) - constant time regardless of number of types/actions

---

## Integration Points

### Evaluator Integration
- **Entry Point**: Evaluator's type switch on `TypeAction`
- **Invocation**: When action value encountered with available arguments
- **Error Handling**: Dispatch errors propagate through standard error return path

### Frame System Integration
- **Type Frame Access**: Standard frame lookup via stack index
- **Function Resolution**: Standard word lookup in frame bindings
- **Scoping**: Actions in root frame, subject to shadowing by local bindings

### Native System Integration
- **Registration**: Actions created during native registration phase
- **Migration**: Existing natives converted to actions incrementally
- **Compatibility**: User code sees no behavioral difference

---

## Backward Compatibility

**Guarantees**:
1. All existing series operations continue to work identically
2. Function names remain unchanged (e.g., `first`, `append`)
3. Argument signatures remain unchanged
4. Error messages remain consistent (except for unsupported type case)
5. Refinement behavior unchanged

**Non-Breaking Changes**:
- Internal implementation changes from direct native to action dispatch
- Type frames introduced (transparent to user)
- Error code added for action-no-impl (new error scenario)

**Breaking Changes**:
- None for existing functionality
- New error scenario when calling series operations on unsupported types (but this was always invalid, now explicit error instead of crash)

---

## Future Extensions

**Potential Enhancements** (not in current scope):
1. User-defined types can register into type frames
2. Actions can support custom dispatch strategies (beyond first-argument type)
3. Inline caching for hot dispatch paths
4. Dispatch analytics (track which actions/types most frequently used)

**Extension Points**:
- New type registration: Create type frame, register in global TypeRegistry, add implementations to frame
- Custom dispatch: Extend dispatch logic to support alternative strategies
- Performance: Add caching layer without changing contract
