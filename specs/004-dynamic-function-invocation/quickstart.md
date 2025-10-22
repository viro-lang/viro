# Quickstart: Dynamic Function Invocation (Actions)

**Feature**: 004-dynamic-function-invocation
**Audience**: Viro language users
**Date**: 2025-10-13

## What Are Actions?

Actions are polymorphic functions in Viro that automatically dispatch to the correct implementation based on the type of their first argument. This allows you to use the same function name (like `first`, `append`, or `length?`) with different data types (blocks, strings, etc.) without needing to know which specific implementation to call.

**Example**:
```viro
>> first [1 2 3]        ; Works with blocks
== 1

>> first "hello"        ; Works with strings
== "h"

>> length? [1 2 3]      ; Works with blocks
== 3

>> length? "hello"      ; Works with strings
== 5
```

Behind the scenes, Viro automatically routes `first [1 2 3]` to the block-specific implementation of `first`, and `first "hello"` to the string-specific implementation. You don't need to do anything special—it just works!

---

## Available Actions

The following series operations are implemented as actions:

### `first` - Get First Element
```viro
>> first [1 2 3]
== 1

>> first "hello"
== "h"
```

**Supported types**: block!, string!

---

### `last` - Get Last Element
```viro
>> last [1 2 3]
== 3

>> last "hello"
== "o"
```

**Supported types**: block!, string!

---

### `append` - Add to End
```viro
>> b: [1 2]
>> append b 3
== [1 2 3]

>> s: "hel"
>> append s "lo"
== "hello"
```

**Supported types**: block!, string!
**Note**: Modifies the series in-place and returns it.

---

### `insert` - Add to Beginning
```viro
>> b: [2 3]
>> insert b 1
== [1 2 3]

>> s: "orld"
>> insert s "W"
== "World"
```

**Supported types**: block!, string!
**Note**: Modifies the series in-place and returns it.

---

### `length?` - Get Series Length
```viro
>> length? [1 2 3]
== 3

>> length? "hello"
== 5

>> length? []
== 0
```

**Supported types**: block!, string!

---

## How Dispatch Works

When you call an action, Viro:

1. **Evaluates the first argument** to determine its type
2. **Looks up the type-specific implementation** for that action
3. **Invokes the correct function** with your arguments

**Example flow** for `first [1 2 3]`:
```
1. User calls: first [1 2 3]
2. Evaluator sees action 'first'
3. Evaluates first arg: [1 2 3] → type is block!
4. Looks up block! in first's type map → finds block-specific first
5. Calls block-first with [1 2 3]
6. Returns: 1
```

---

## Error Handling

### Unsupported Type Error

If you call an action on a type that doesn't support it, you'll get a clear error:

```viro
>> first 42
** Script error: Action 'first' not defined for type integer!
```

This tells you:
- Which action you tried to call (`first`)
- Which type doesn't support it (`integer!`)

**Fix**: Use the action only with supported types (check the "Supported types" for each action).

---

### Empty Series Error

Some actions (like `first` and `last`) require non-empty series:

```viro
>> first []
** Script error: Series is empty

>> last ""
** Script error: Series is empty
```

**Fix**: Check if series is empty before calling, or handle the error.

---

### Type Mismatch Error

String operations only accept string values:

```viro
>> append "hello" 42
** Script error: Type mismatch (expected string, got integer)
```

**Fix**: Ensure you're appending/inserting values of the correct type.

---

## Common Patterns

### Pattern 1: Chaining Series Operations

```viro
>> b: [1 2 3]
>> append (append b 4) 5
== [1 2 3 4 5]
```

Since `append` returns the modified series, you can chain operations.

---

### Pattern 2: Polymorphic Functions

Write functions that work with any series type:

```viro
>> get-middle: fn [series] [
     len: length? series
     if len < 2 [return none]
     first (skip series (len / 2))
   ]

>> get-middle [1 2 3 4 5]
== 3

>> get-middle "hello"
== "l"
```

Your function doesn't need to know whether it's working with a block or string—the actions handle the dispatch automatically!

---

### Pattern 3: Type-Specific Behavior

Different types may have different behavior for the same action:

```viro
>> append [1 2] [3 4]
== [1 2 [3 4]]          ; Appends block as nested element

>> append "hel" "lo"
== "hello"              ; Concatenates strings
```

Understanding type-specific semantics helps you write correct code.

---

## Scoping and Shadowing

Actions follow Viro's local-by-default scoping rules. You can shadow an action with a local binding:

```viro
>> first [1 2 3]
== 1

>> first: fn [x] [x * 2]    ; Shadow 'first' with local function

>> first 5
== 10                        ; Uses local binding, not action

>> do [first [1 2 3]]        ; New scope, action binding visible
== 1
```

**Key point**: Local bindings take precedence over actions in the root frame.

---

## Performance Considerations

Actions have a small dispatch overhead compared to direct function calls (typically 2-5x slower). For most use cases, this overhead is negligible. If you're in a performance-critical loop, consider:

1. **Measure first**: Profile your code to confirm the action dispatch is the bottleneck
2. **Optimize if needed**: Use type-specific implementations directly (advanced usage)

For typical Viro programs, action dispatch is fast enough and provides significant ergonomic benefits.

---

## FAQ

### Q: What's the difference between an action and a function?

**A**: Actions are a special kind of function that dispatches based on the first argument's type. Regular functions have a single implementation. Actions have multiple implementations (one per type) and automatically choose the right one.

### Q: Can I create my own actions?

**A**: Not in the current implementation. Actions are currently limited to built-in series operations. Future versions may support user-defined actions.

### Q: Do actions work with user-defined types?

**A**: Not yet. The action system is designed to support user-defined types in the future, but user-defined types are not yet implemented in Viro.

### Q: Why do I get "action-no-impl" errors?

**A**: You're calling an action on a type that doesn't support it. For example, `first 42` fails because `first` only works with blocks and strings, not integers. Check the "Supported types" list for each action.

### Q: Can actions dispatch on multiple argument types?

**A**: No, actions only dispatch on the **first argument's type**. Subsequent arguments are validated by the type-specific implementation, but they don't affect which implementation is chosen.

### Q: What happens if I pass the wrong type for a subsequent argument?

**A**: The type-specific implementation will generate an error. For example, `append "hello" 42` fails with a type-mismatch error because the string implementation of `append` expects a string value.

---

## Next Steps

- **Read the contracts**: See `contracts/series-actions.md` for detailed specifications
- **Explore the code**: Type-specific implementations are in `internal/native/series_*.go`
- **Run the tests**: Contract tests in `test/contract/*_test.go` show comprehensive usage examples
- **Experiment in REPL**: Try different combinations to understand dispatch behavior

---

## Summary

Actions provide **polymorphic dispatch** in Viro, allowing you to:
- Use the same function name with different types
- Write generic code that works with any series type
- Get clear errors when types aren't supported

**Key takeaways**:
1. Actions dispatch based on first argument type
2. Supported types: block!, string! (for series actions)
3. Errors are clear and actionable
4. Follows local-by-default scoping (can be shadowed)
5. Small performance overhead, negligible for most use cases

Enjoy writing polymorphic Viro code! 🚀
