# Data Model: Dynamic Function Invocation (Action Types)

**Feature**: 004-dynamic-function-invocation
**Date**: 2025-10-13

## Overview

This document defines the core entities and their relationships for the action dispatch system in Viro. The system enables type-based polymorphic dispatch while maintaining Viro's index-based architecture and local-by-default scoping.

---

## Entity Definitions

### 1. Action (Value Type)

**Description**: A polymorphic function that dispatches to type-specific implementations based on the first argument's type.

**Fields**:
| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `Name` | `string` | Action name (e.g., "first", "append") | Non-empty, valid word characters |
| `ParamSpec` | `[]Parameter` | Parameter specifications (arity, eval flags, refinements) | Must have at least 1 parameter (the dispatch argument) |

**Relationships**:
- References: `TypeRegistry` (for runtime type frame lookup during dispatch)
- Stored in: Root frame (like other native functions)
- Used by: Evaluator during function invocation

**State Transitions**: None (immutable after creation)

**Invariants**:
- ParamSpec must include at least the dispatch parameter (first argument)
- Name must be a valid Viro word (alphanumeric, hyphens, underscores, question marks)
- Actions conceptually support ALL types; implementation availability determined at runtime

**Example**:
```go
ActionValue{
    Name: "first",
    ParamSpec: []Parameter{
        {Name: "series", Eval: true},  // First arg, evaluated, used for dispatch
    },
}

// Runtime dispatch flow:
// 1. Evaluate first argument → block! value
// 2. Look up TypeBlock in TypeRegistry → frame index 1
// 3. Look up "first" in frame 1 → block-specific first implementation
// 4. If found: invoke it; if not found: error "action not defined for type"
```

---

### 2. TypeFrame (Regular Frame)

**Description**: A type frame is a **regular Frame** containing type-specific function implementations. It uses the standard frame mechanism (Words/Values arrays) with no special fields. The only thing that makes it a "type frame" is that it's stored in the TypeRegistry instead of the frameStore.

**Structure**: Uses the existing `Frame` struct from `internal/frame`:
```go
// Existing Frame structure (no modifications needed)
type Frame struct {
    Type     FrameType       // Frame category (e.g., FrameGlobal for type frames)
    Words    []string        // Function names: ["first", "last", "append", ...]
    Values   []value.Value   // FunctionValue instances (parallel to Words)
    Parent   int             // Index of parent frame (always 0 for type frames)
    Index    int             // Position in frameStore (-1 for type frames, not in store)
    Name     string          // Optional name for diagnostics (e.g., "block!")
}
```

**Storage Location**:
- Type frames are stored directly in TypeRegistry as `*Frame` pointers
- They are **NOT** stored in the frameStore (stack)
- Their Index field is set to -1 to indicate they're not in frameStore
- Root frame (index 0) remains on stack for normal execution frame access

**Relationships**:
- Parent: Root frame (index 0 on stack) - accessed via index-based lookup
- Contains: Type-specific function implementations bound using standard frame mechanism
- Referenced by: TypeRegistry (maps ValueType → *Frame)

**State Transitions**:
```
[Created at startup] → [Populated with functions via Set()] → [Stored in TypeRegistry] → [Immutable]
```

**Invariants**:
- All type frames have Parent = 0 (root frame) for standard word resolution
- All type frames have Index = -1 (not in frameStore)
- Type frames use FrameGlobal type (same as root frame)
- Functions bound using existing `frame.Set(word, funcValue)` method
- Each ValueType has at most one type frame
- Type frames are never pushed/popped from stack (permanent, not execution context)

**Example**:
```go
// Block type frame (created using existing Frame, stored in TypeRegistry)
// No special fields needed - just a regular frame!

// During initialization:
blockFrame := frame.New(frame.FrameGlobal, 0, "block!")
blockFrame.Index = -1  // Not in frameStore
blockFrame.Set("first", value.FuncVal(blockFirstImpl))
blockFrame.Set("last", value.FuncVal(blockLastImpl))
blockFrame.Set("append", value.FuncVal(blockAppendImpl))

// Stored in TypeRegistry (not frameStore):
TypeRegistry[TypeBlock] = blockFrame

// Result is a standard Frame:
// Words:  ["first", "last", "append"]
// Values: [FuncVal(...), FuncVal(...), FuncVal(...)]
// Parent: 0 (points to root frame on stack)
// Index:  -1 (not in frameStore)
```

---

### 3. TypeRegistry (Lookup Table)

**Description**: A global mapping from value types to their corresponding type frames, storing direct pointers to Frame structs. Type frames are stored here instead of in the frameStore (stack).

**Fields**:
| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `TypeToFrame` | `map[ValueType]*Frame` | Maps each type to its frame pointer | All types must have at most one frame |

**Relationships**:
- Stores: All `TypeFrame` entities (owns them)
- Used by: Evaluator during action dispatch

**State Transitions**:
```
[Empty at startup] → [Populated during type frame init] → [Immutable]
```

**Invariants**:
- Each ValueType appears at most once as a key
- All frame pointers reference valid type frames
- Registry is fully populated before any actions are created
- Never modified after initialization completes
- Type frames stored here have Index = -1 (not in frameStore)

**Example**:
```go
var TypeRegistry = map[ValueType]*Frame{
    TypeBlock:   &blockFrame,   // Direct pointer to block type frame
    TypeString:  &stringFrame,  // Direct pointer to string type frame
    TypeInteger: &intFrame,     // Direct pointer to integer type frame (if it has type-specific ops)
    // ... etc.
}

// Each frame has:
// - Parent = 0 (root frame on stack)
// - Index = -1 (not in frameStore)
```

---

### 4. DispatchContext (Runtime Data)

**Description**: Ephemeral context data used during action dispatch. Not persisted; created on-demand during evaluation.

**Fields**:
| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `ActionName` | `string` | Name of action being dispatched | Non-empty |
| `FirstArgType` | `ValueType` | Type of first argument | Valid ValueType |
| `TypeFrame` | `*Frame` | Resolved type frame pointer | Non-nil if type found |
| `ResolvedFunc` | `*FunctionValue` | Type-specific function to invoke | Non-nil |

**Relationships**:
- Created by: Evaluator during action invocation
- References: One `Action`, one `TypeFrame`, one `FunctionValue`
- Lifetime: Single evaluation step (not stored)

**State Transitions**:
```
[Created] → [Type Resolved] → [Function Resolved] → [Function Invoked] → [Destroyed]
```

**Invariants**:
- FirstArgType may or may not exist in global TypeRegistry (determines success or error)
- If TypeRegistry has FirstArgType, ResolvedFunc may or may not exist in TypeFrame (determines success or error)
- Context exists only during evaluation (not stored in stack/frames)

**Example**:
```go
// During evaluation of: first [1 2 3]
DispatchContext{
    ActionName:      "first",
    FirstArgType:    TypeBlock,
    TypeFrame:       TypeRegistry[TypeBlock],  // Direct pointer to block frame
    ResolvedFunc:    &blockFirstFunc,           // Block's first implementation
}
```

---

## Relationships Diagram

```
┌─────────────┐
│  Root Frame │ (index 0, on stack)
│  (Globals)  │
└──────┬──────┘
       │
       │ parent (index-based, all type frames point to index 0)
       │
       ├──────────────┬──────────────┬──────────────┐
       │              │              │              │
       ▼              ▼              ▼              ▼
┌────────────┐ ┌────────────┐ ┌────────────┐ ┌────────────┐
│Block Frame │ │String Frame│ │Integer Fr..│ │  ... etc   │
│ (Index=-1) │ │ (Index=-1) │ │ (Index=-1) │ │ (Index=-1) │
│ Parent=0   │ │ Parent=0   │ │ Parent=0   │ │ Parent=0   │
└─────┬──────┘ └─────┬──────┘ └─────┬──────┘ └─────┬──────┘
      │              │              │              │
      │ contains     │ contains     │ contains     │ contains
      │              │              │              │
      ▼              ▼              ▼              ▼
┌──────────┐   ┌──────────┐   ┌──────────┐   ┌──────────┐
│  first   │   │  first   │   │  (none)  │   │  ...     │
│  last    │   │  last    │   │          │   │          │
│  append  │   │  append  │   │          │   │          │
│  ...     │   │  ...     │   │          │   │          │
└──────────┘   └──────────┘   └──────────┘   └──────────┘
   (Block        (String        (Integer       (Type-
   impls)         impls)         impls)        specific
                                               impls)

                ┌─────────────────┐
                │  Action: first  │
                │                 │
                │  Name: "first"  │
                │  ParamSpec: [..]│
                └────────┬────────┘
                         │
                         │ runtime lookup via
                         │ first arg type
                         │
                         ▼
                ┌─────────────────┐
                │  TypeRegistry   │
                │  (global map)   │
                │                 │
                │  Block → *Frame │───┐ (direct pointers)
                │  String → *Frame│───┼─┐
                │  Integer→ *Frame│   │ │
                └─────────────────┘   │ │
                          │           │ │
                          │stores     │ │
                          │owns       │ │
                          ▼           ▼ ▼
                   (Block Frame)  (String Frame)
                   Has "first"?   Has "first"?
                   → Yes, invoke  → Yes, invoke
```

**Key Relationships**:
1. **Action → TypeRegistry**: Actions query TypeRegistry to resolve type frames
2. **TypeRegistry → TypeFrame**: Registry stores direct pointers to type frames (not indices)
3. **TypeFrame → Root Frame**: All type frames have Parent=0 (index-based reference to root on stack)
4. **TypeFrame → Functions**: Type frames contain type-specific function implementations
5. **Evaluator → Action**: Evaluator uses action during dispatch
6. **DispatchContext → All**: Temporary context connects action, type frame, and function
7. **Type Frames NOT on Stack**: Type frames stored in TypeRegistry, not in frameStore (Index=-1)

---

## Validation Rules

### Action Creation
```go
func ValidateAction(a *ActionValue) error {
    if a.Name == "" {
        return errors.New("action name cannot be empty")
    }
    if len(a.ParamSpec) == 0 {
        return errors.New("action must have at least one parameter")
    }
    return nil
}
```

### Type Frame Initialization
```go
func ValidateTypeFrame(f *Frame) error {
    if f.Parent != 0 {
        return errors.New("type frames must have root as parent")
    }
    // Validate all bindings are functions using standard frame lookup
    for i, word := range f.Words {
        val := f.Values[i]
        if val.Type != TypeFunction {
            return fmt.Errorf("type frame binding %s must be function, got %s", word, val.Type)
        }
    }
    return nil
}
```

### Dispatch Resolution
```go
func DispatchAction(action *ActionValue, argType ValueType, registry map[ValueType]*Frame) (*FunctionValue, error) {
    // Lookup type frame from global registry (direct pointer)
    frame, exists := registry[argType]
    if !exists {
        // Type has no type frame - no implementations registered for this type
        return nil, fmt.Errorf("action %s not defined for type %s", action.Name, argType)
    }

    // Look up action name using standard frame.Get() method
    funcVal, found := frame.Get(action.Name)
    if !found {
        // Type frame exists but doesn't have this action
        return nil, fmt.Errorf("action %s not defined for type %s", action.Name, argType)
    }

    return funcVal.AsFunction(), nil
}
```

---

## Usage Examples

### Example 1: Calling `first` on a block

```viro
first [1 2 3]
```

**Entity Interactions**:
1. Evaluator encounters action value `first` in root frame
2. Evaluator reads next argument: `[1 2 3]` (evaluates to block value)
3. Extracts type from first argument → TypeBlock
4. Looks up TypeBlock in global TypeRegistry → gets direct pointer to block frame
5. Creates DispatchContext:
   - ActionName: "first"
   - FirstArgType: TypeBlock
   - TypeFrame: &blockFrame (direct pointer)
6. Looks up "first" in block frame using frame.Get("first") → gets block-specific first function
7. Invokes function with `[1 2 3]` as argument
8. Returns `1`

### Example 2: Calling `append` on a string

```viro
append "hel" "lo"
```

**Entity Interactions**:
1. Evaluator encounters action value `append` in root frame
2. Evaluator reads first argument: `"hel"` (evaluates to string value)
3. Extracts type from first argument → TypeString
4. Looks up TypeString in global TypeRegistry → gets direct pointer to string frame
5. Creates DispatchContext:
   - ActionName: "append"
   - FirstArgType: TypeString
   - TypeFrame: &stringFrame (direct pointer)
6. Looks up "append" in string frame using frame.Get("append") → gets string-specific append function
7. Invokes function with `"hel"` and `"lo"` as arguments
8. Returns `"hello"`

### Example 3: Error - Unsupported Type

```viro
first 42
```

**Entity Interactions**:
1. Evaluator encounters action value `first` in root frame
2. Evaluator reads first argument: `42` (evaluates to integer value)
3. Extracts type from first argument → TypeInteger
4. Looks up TypeInteger in global TypeRegistry → may or may not find frame pointer
5. If frame found: looks up "first" in integer frame → **not found**
6. Generates error: `verror.NewScriptError("action-no-impl", ["first", "integer!", ""], ...)`
7. Returns error to user: "Action 'first' not defined for type integer!"

---

## Migration Path

### Current State (Before Actions)
```
Root Frame (index 0, on stack):
├── first:  FunctionValue{impl: native.First, spec: [series]}
├── last:   FunctionValue{impl: native.Last, spec: [series]}
├── append: FunctionValue{impl: native.Append, spec: [series value]}
└── ...

(No type frames, no TypeRegistry)
```

### Target State (After Actions)
```
Root Frame (index 0, on stack):
├── first:  ActionValue{Name: "first", ParamSpec: [...]}
├── last:   ActionValue{Name: "last", ParamSpec: [...]}
├── append: ActionValue{Name: "append", ParamSpec: [...]}
└── ...

Block Frame (Index=-1, in TypeRegistry):
├── first:  FunctionValue{impl: native.BlockFirst, ...}
├── last:   FunctionValue{impl: native.BlockLast, ...}
└── append: FunctionValue{impl: native.BlockAppend, ...}

String Frame (Index=-1, in TypeRegistry):
├── first:  FunctionValue{impl: native.StringFirst, ...}
├── last:   FunctionValue{impl: native.StringLast, ...}
└── append: FunctionValue{impl: native.StringAppend, ...}

Global TypeRegistry:
├── TypeBlock → &blockFrame (direct pointer)
├── TypeString → &stringFrame (direct pointer)
└── ...
```

**Migration Steps**:
1. Create type frames during startup (before registering natives)
2. Implement type-specific functions in separate files
3. Register functions into type frames
4. Create action values with TypeFrames maps
5. Bind actions to root frame (replacing old natives)
6. Verify existing tests pass
7. Add new tests for type dispatch and errors

---

## Performance Considerations

**Lookup Complexity**:
- Action dispatch: O(1) TypeRegistry lookup for type → frame pointer
- Function lookup: O(1) map lookup in frame bindings
- Total dispatch overhead: O(1) with ~2 hash lookups (one less than stack-based approach)

**Memory Overhead**:
- One type frame per native type: ~10 frames × ~100 bytes = ~1KB
- One action per series operation: ~10 actions × ~50 bytes = ~500 bytes (minimal - just name + param spec)
- One global TypeRegistry: ~10 types × 8-16 bytes per pointer = ~80-160 bytes
- Total overhead: ~1.6KB (negligible for interpreter)

**Benefits of Direct Pointer Storage**:
- Faster dispatch: One fewer indirection (no frameStore.Get() call)
- Type frames don't pollute stack traces (Index=-1, not in frameStore)
- Clearer separation: execution frames on stack, type metadata in TypeRegistry

**Optimization Opportunities** (if needed):
- Inline caching: Store last (action, type) → function resolution
- Devirtualization: Generate specialized code paths for common types
- Method table: Pre-compute action → type → function mappings at startup

---

## Summary

The action dispatch system introduces three primary entities:
1. **Action**: Polymorphic function (just Name + ParamSpec)
2. **TypeFrame**: Regular Frame containing type-specific implementations (stored in TypeRegistry)
3. **TypeRegistry**: Global map storing direct pointers to type frames

These entities maintain Viro's architectural principles:
- **Index-based parent references**: Type frames use Parent=0 (index) to reference root frame on stack
- **Separation of concerns**: Execution frames on stack, type metadata in TypeRegistry
- **Immutable**: Type frames and actions created once at startup
- **Consistent**: Reuses existing Frame structure (Words/Values arrays)
- **Extensible**: New types can register type frames without core changes
- **Clean stack traces**: Type frames have Index=-1 (not in frameStore), don't pollute execution traces

The system enables polymorphic dispatch while preserving backward compatibility and maintaining predictable performance characteristics. Direct pointer storage provides faster dispatch (one fewer indirection) compared to index-based type frame storage.
