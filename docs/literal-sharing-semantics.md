# Literal Sharing Semantics

## Overview

Viro implements an intentional optimization for literal values (blocks, strings, and binaries) where **non-empty literals are shared** between function calls while **empty literals are cloned**. This design prioritizes performance for common use cases while maintaining isolation for mutable empty containers.

## Behavior

### Non-Empty Literals (Shared)

Non-empty block, string, and binary literals in function bodies are returned as-is on each call. This means mutations accumulate across calls:

```viro
shared-block: fn [] [[1 2 3]]
result1: shared-block  ; Returns [1 2 3]
append result1 4       ; Modifies the shared block
result2: shared-block  ; Returns [1 2 3 4] - SAME block!
```

```viro
shared-string: fn [] ["hello"]
result1: shared-string   ; Returns "hello"
append result1 " world"  ; Modifies the shared string
result2: shared-string   ; Returns "hello world" - SAME string!
```

```viro
shared-binary: fn [] [#{010203}]
result1: shared-binary   ; Returns #{010203}
append result1 #{04}     ; Modifies the shared binary
result2: shared-binary   ; Returns #{01020304} - SAME binary!
```

### Empty Literals (Cloned)

Empty literals are cloned on each call to provide isolation:

```viro
isolated-block: fn [] []
result1: isolated-block  ; Returns fresh []
append result1 1         ; Modifies only result1
result2: isolated-block  ; Returns fresh [] - isolated!
```

```viro
isolated-string: fn [] [""]
result1: isolated-string   ; Returns fresh ""
append result1 "x"         ; Modifies only result1
result2: isolated-string   ; Returns fresh "" - isolated!
```

```viro
isolated-binary: fn [] [#{}]
result1: isolated-binary   ; Returns fresh #{}
append result1 #{01}       ; Modifies only result1
result2: isolated-binary   ; Returns fresh #{} - isolated!
```

## Design Rationale

This behavior is **intentional** and optimized for performance:

1. **Non-empty literals** are typically used as templates or constants that shouldn't be modified, so sharing avoids expensive copying of potentially large data structures.

2. **Empty literals** are commonly used as mutable containers, so cloning ensures each call gets a fresh instance.

3. **Performance optimization**: Avoids deep cloning of large literals on every function call while providing expected isolation for empty containers.

## Nested Structure Surprises

Sharing behavior applies to the top-level literal, but nested empty structures within shared non-empty literals are still cloned:

```viro
nested: fn [] [[[]]]      ; Non-empty outer block is shared
result1: nested           ; Returns [[[]]]
inner1: first result1     ; Gets the inner []
append inner1 1           ; Modifies inner1
result2: nested           ; Returns [[[]]] - outer shared, but inner is fresh!
```

The outer `[[[]]]` is shared, but the inner `[]` gets cloned each time the function is called.

## Best Practices

### When You Need Isolation

Use `copy` for non-empty literals that should be isolated:

```viro
isolated-template: fn [] [copy [1 2 3]]
result1: isolated-template  ; Returns [1 2 3]
append result1 4            ; Safe modification
result2: isolated-template  ; Returns [1 2 3] - isolated!
```

### When Sharing Is Desired

Non-empty literals are perfect for shared constants:

```viro
days: fn [] [[mon tue wed thu fri sat sun]]
weekdays: take days 5     ; Modifies shared block
weekends: take days 2     ; Gets remaining from shared block
```

## Implementation

This behavior is implemented in `internal/eval/evaluator.go` lines 477-508, where:

- Empty literals (`Length() == 0`) are deep cloned using `value.DeepCloneValue()`
- Non-empty literals are returned as-is (shared reference)

**WARNING**: Changing this behavior would be a breaking change that could affect performance and existing code expectations. Any modifications must preserve the current semantics.</content>
<parameter name="filePath">docs/literal-sharing-semantics.md