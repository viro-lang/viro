# Viro Execution Model

This document describes how Viro code is parsed and executed. Understanding this model is **critical** for implementing language features correctly.

**IMPORTANT:** This document describes the **correct** execution model based on the reference implementation. The current Go implementation may deviate from this model and should be refactored to match.

## Overview

Viro uses a **homoiconic** design where code is data. There is NO separate Abstract Syntax Tree (AST). The parser directly converts source text into Viro values, and the evaluator operates on these values.

## Two-Phase Execution

### Phase 1: Parsing (String → Values)

The parser converts source text directly into a `BlockVal` containing Viro values.

**Example:**
```viro
1 hello x: 10
```

**Parses to:**
```
BlockVal containing:
  - IntegerVal(1)
  - WordVal("hello")
  - SetWordVal("x")
  - IntegerVal(10)
```

**Key Point:** Parsing does NOT interpret meaning or create special structures. It only tokenizes the input into typed values.

### Phase 2: Evaluation (Values → Results)

The evaluator processes the `BlockVal` by reading values sequentially from start to finish, evaluating one expression at a time.

## Core Evaluation Primitive: Evaluate Expression

The fundamental evaluation operation is **evaluate-expression**:

**Input:**
- `context` - Current execution context (scope)
- `block` - The block of values being evaluated
- `position` - Current position in the block
- `last_result` - Result of the previous expression (for infix)

**Output:**
- `new_position` - Position after consuming this expression
- `result_value` - The evaluated result

**Critical Insight:** Evaluate-expression processes **ONE** expression starting at the current position and returns the **new position** after consuming all values that belong to that expression.

### Expression Evaluation Dispatch

```
function evaluate_expression(context, block, position, last_result):
    element = block[position]
    
    switch element.type:
        case LitWord:
            // 'word → return WordVal("word")
            return (position + 1, WordVal(element.name))
        
        case GetWord:
            // :word → lookup and return value (no invocation)
            value = context.lookup(element.name)
            return (position + 1, value)
        
        case SetWord:
            // word: value → evaluate next expr, bind, return
            (new_pos, value) = evaluate_expression(context, block, position + 1, last_result)
            context.bind(element.name, value)
            return (new_pos, value)
        
        case Word:
            value = context.lookup(element.name)
            if value == undefined:
                error("word has no value")
            
            // Is it a function?
            if value.is_function():
                // YES: handle function call (consumes arguments)
                return invoke_function(context, block, position, element.name, value, last_result)
            else:
                // NO: just return the value
                return (position + 1, value)
        
        default:
            // Literals (integers, strings, blocks, etc.) evaluate to themselves
            return (position + 1, element)
```

**Key Insights:**

1. **Position tracking**: Each case returns the new position after consuming values
2. **Recursive calls**: SetWord and function calls recursively evaluate sub-expressions
3. **last_result**: Passed through to support infix operations
4. **No pre-dispatch**: Values are evaluated on-demand during sequential traversal

## Block Evaluation

```
function evaluate_block(context, block):
    position = 0
    last_result = none
    
    while position < block.length:
        (position, last_result) = evaluate_expression(context, block, position, last_result)
    
    return last_result
```

**Key Insights:**

1. **Sequential execution**: Loop from start to end
2. **Position advancement**: Evaluate-expression returns new position (may skip multiple values)
3. **Result threading**: Result of previous expression is passed to next (enables infix)
4. **Return last result**: Block returns the result of the final expression

## Value Type Evaluation Rules

### Literal Values
**Types:** Integer, String, Decimal, Logic, None, Block, Function

**Behavior:** Return themselves unchanged, advance position by 1.

**Example:**
- `42` → returns `IntegerVal(42)`, position + 1
- `"hello"` → returns `StringVal("hello")`, position + 1
- `[1 2 3]` → returns `BlockVal(...)`, position + 1 (NOT evaluated)

### LitWord
**Syntax:** `'word`

**Behavior:** Returns a `WordVal` containing the same string, advance position by 1.

**Example:**
```viro
'hello  ; → WordVal("hello"), position + 1
```

### GetWord
**Syntax:** `:word`

**Behavior:** 
1. Look up the word in context
2. Return the value WITHOUT checking if it's a function
3. Advance position by 1

**Use Case:** Access function values without invoking them.

**Example:**
```viro
add: func [a b] [a + b]
add 5 3         ; Invokes function → 8
:add            ; GetWord → returns the function value itself
```

### SetWord
**Syntax:** `word:`

**Behavior:**
1. Recursively evaluate the NEXT expression
2. Bind the result to the word in current context
3. Return the result AND the new position

**Critical:** SetWord **consumes** the next expression. The position advances past both the set-word and its value.

**Example:**
```viro
x: 10
; position=0: SetWordVal("x")
; → evaluate_expression(context, block, position=1, last_result) 
; → returns (position=2, IntegerVal(10))
; → binds x to 10
; → returns (position=2, IntegerVal(10))

y: add 5 3
; position=0: SetWordVal("y")
; → evaluate_expression(context, block, position=1, last_result)
; → WordVal("add") is a function
; → invoke_function consumes positions 1, 2, 3
; → returns (position=4, result)
; → binds y to result
; → returns (position=4, result)
```

### Word
**Syntax:** `word`

**Behavior:**
1. Look up the word in context
2. If not found → error
3. **Check if value is a function**
   - **If YES**: Invoke the function (collecting arguments)
   - **If NO**: Return the value, advance position by 1

**This is CRITICAL:** Words **automatically invoke** functions when looked up.

**Example:**
```viro
x: 10
x           ; WordVal("x") → lookup → IntegerVal(10) (not a function) → return 10

add: func [a b] [a + b]
add 5 3     ; WordVal("add") → lookup → FunctionVal (IS a function) → invoke with args
```

## Function Invocation

When a `Word` resolves to a function, the invocation process begins:

```
function invoke_function(context, block, position, name, function, last_result):
    args = []
    refinements = {}
    position = position + 1  // Move past the function word
    
    arg_index = 0
    
    // Handle infix
    if function.is_infix:
        if last_result == none:
            error("Missing left argument for infix function")
        args.append(last_result)
        arg_index = 1
    
    // Collect remaining arguments
    while arg_index < function.arity:
        arg_spec = function.parameters[arg_index]
        arg_index = arg_index + 1
        
        // Read any refinements before this arg
        position = read_refinements(context, block, position, function, refinements)
        
        // Evaluate or use literal based on parameter spec
        if arg_spec.should_evaluate:
            (position, arg_value) = evaluate_expression(context, block, position, none)
        else:
            arg_value = block[position]
            position = position + 1
        
        args.append(arg_value)
    
    // Read trailing refinements
    position = read_refinements(context, block, position, function, refinements)
    
    // Invoke function
    result = function.execute(context, args, refinements)
    return (position, result)
```

**Key Insights:**

1. **Position advancement**: Starts at function word, advances past each argument
2. **Argument evaluation**: Each arg may recursively evaluate expressions, advancing position
3. **Infix handling**: If function is infix, uses `last_result` as first argument
4. **Refinements**: Can appear before, between, or after positional arguments
5. **Returns new position**: After consuming all arguments

### Argument Evaluation Control

Function parameters specify whether to evaluate the argument or use it literally:

```
// Evaluated argument (default)
parameter { name: "value", evaluate: true }   // Calls evaluate_expression()

// Literal argument (for control flow)
parameter { name: "block", evaluate: false }  // Uses raw value from block
```

**Example:**
```viro
when: func [condition [evaluate: true] block [evaluate: false]] [...]

when x > 5 [print "big"]
; condition: evaluates "x > 5" → LogicVal(true)
; block: literal BlockVal (NOT evaluated yet - when decides if/when to evaluate it)
```

## Infix Functions

Infix functions **require**:
1. Exactly 2 arguments
2. First argument MUST be evaluated
3. Function has infix property set to true

**Behavior:**
1. Use `last_result` (result of previous expression) as first argument
2. Evaluate next expression for second argument
3. Invoke function

**Example:**
```viro
+: func [a b] [...]
; Mark + as infix

5 + 3
; position=0: evaluate_expression → IntegerVal(5), last_result = 5
; position=1: evaluate_expression(last_result=5)
;   → WordVal("+") → lookup → FunctionVal
;   → function.is_infix == true
;   → args[0] = last_result (5)
;   → evaluate_expression at position 2 → IntegerVal(3)
;   → args[1] = 3
;   → invoke: +(5, 3) → 8
; → returns (position=3, 8)

1 + 2 + 3
; position=0: evaluate → 1, last=1
; position=1: evaluate(last=1) → +(1,2) → 3, last=3
; position=2: evaluate(last=3) → +(3,3) → 6
```

**Critical:** Infix is left-associative due to `last_result` threading through the evaluation loop.

## Refinements

Refinements are function modifiers starting with `--`:

**Two types:**
1. **Flag refinements**: Boolean flags (`--flag`)
2. **Value refinements**: Take a value (`--option value`)

**Parsing:**
```
function read_refinements(context, block, position, function, refinements):
    element = block[position]
    
    while element != undefined and element.is_refinement():
        name = element.name.remove_prefix("--")
        
        if not function.has_refinement(name):
            error("Unknown refinement: " + name)
        
        if function.refinement_takes_value(name):
            // Value refinement: evaluate next expression
            (position, value) = evaluate_expression(context, block, position + 1, none)
            refinements[name] = value
        else:
            // Flag refinement
            position = position + 1
            refinements[name] = true
        
        element = block[position]
    
    return position
```

**Example:**
```viro
foo: func [x] [refinements: [--bar --baz <value>]] [...]

foo --bar 10 --baz 20
; position=0: WordVal("foo") → function
; position=1: read_refinements → "--bar" (flag), position=2
; position=2: evaluate_expression → 10, bind to x, position=3
; position=3: read_refinements → "--baz" (value), evaluate → 20, position=5
```

## Path Evaluation (Get Path)

**Syntax:** `word.field` or `block.1`

**Behavior:**
1. Evaluate first segment (word lookup)
2. For each subsequent segment:
   - **Word segment**: Object field access
   - **Index segment**: Collection indexing (1-based)
3. Return value at final location
4. Advance position by 1 (path is a single token)

**Example:**
```viro
obj: make object! [x: 10]
obj.x           ; PathVal → looks up "obj" → accesses field "x" → returns 10

data: [1 2 3]
data.2          ; PathVal → looks up "data" → indexes at 2 → returns 2
```

## Path Assignment (Set Path)

**Syntax:** `word.field:` or `block.1:`

**Behavior:**
1. Evaluate the NEXT expression (recursive evaluate-expression)
2. Traverse path to container
3. Assign result to final location
4. Return result and new position

**Example:**
```viro
obj: make object! [x: 10]
obj.x: 20       ; SetPath → evaluates 20 → assigns to obj.x

data: [1 2 3]
data.2: 99      ; SetPath → evaluates 99 → assigns to data[2]
```

**Note:** Set-path consumes the next expression, just like set-word.

## Context and Scoping

Contexts form a hierarchy for lexical scoping:

```
Current Context
    ↓ (if not found, check parent)
Parent Context
    ↓
Global Context
```

**Lookup process:**
1. Check current context
2. If not found and context has parent, check parent
3. Continue until found or reach global
4. Error if not found

**Function execution creates new context:**
1. Create new context with parent = function's definition context
2. Bind parameters to argument values
3. Evaluate function body in this context
4. Discard context on return

## Native Function Implementation

Native functions receive:
- Execution context
- Positional arguments (already evaluated or literal per parameter spec)
- Refinements map

**Example - Simple Native:**
```
function native_add(context, args, refinements):
    a = args[0]  // Already evaluated
    b = args[1]  // Already evaluated
    return IntegerVal(a.value + b.value)
```

**Example - Control Flow Native:**
```
function native_when(context, args, refinements):
    condition = args[0]  // Already evaluated (evaluate: true)
    block = args[1]      // Literal block (evaluate: false)
    
    if to_truthy(condition):
        // NOW evaluate the block
        return evaluate_block(context, block)
    
    return NoneVal
```

**Key Point:** Natives control when/if to evaluate literal arguments (like blocks for control flow).

## Key Implementation Rules

### 1. evaluate_expression Returns Position
**Every** evaluation must return `(new_position, value)`. This is how the evaluator knows how many values were consumed.

```
// Literal value - consumes 1 position
case Integer:
    return (position + 1, element)

// SetWord - consumes set-word + expression
case SetWord:
    (new_pos, value) = evaluate_expression(context, block, position + 1, last_result)
    context.bind(element.name, value)
    return (new_pos, value)
```

### 2. Function Calls Consume Arguments
When a word resolves to a function, the invocation **consumes** all argument positions:

```
function invoke_function(...):
    position = position + 1  // Move past function word
    
    for each parameter:
        (position, arg) = evaluate_expression(context, block, position, none)
        args.append(arg)
    
    result = function.execute(args)
    return (position, result)  // Return position AFTER all args
```

### 3. Infix Uses last_result
Infix functions **require** the `last_result` parameter:

```
if function.is_infix and last_result != none:
    args[0] = last_result  // Previous expression result
    // Collect remaining args from current position
```

### 4. NO Pre-Scanning or AST Building
Do NOT scan ahead or build intermediate structures. Evaluate on-demand:

```
// WRONG - Don't build AST
function parse_expressions(block):
    expressions = []
    for value in block:
        expr = identify_expression_type(value)
        expressions.append(expr)
    return expressions

// CORRECT - Evaluate on-demand
function evaluate_expression(context, block, position, last_result):
    // Evaluate current position, return new position
    element = block[position]
    // ... handle based on element type
```

### 5. Literal Arguments Don't Get Evaluated
Control flow functions need unevaluated blocks:

```
// Function definition
when: func [
    condition [evaluate: true]    // Will be evaluated before passed
    block [evaluate: false]        // Will be passed as-is
] [...]

// Native implementation
function native_when(context, args, refinements):
    condition = args[0]  // Already evaluated
    block = args[1]      // Literal block - NOT evaluated yet
    
    if to_truthy(condition):
        // Native controls when to evaluate
        return evaluate_block(context, block.elements)
    
    return none
```

### 6. Value Constructors Always
Always use constructor functions to create values:

```
// CORRECT
IntegerVal(42)
StringVal("hello")
BlockVal(elements)

// WRONG - Never do direct construction
new IntegerValue(42)
IntegerValue{value: 42}
```

## Common Pitfalls

### ❌ Wrong: Evaluate all arguments upfront
```
// Don't do this - you don't know how many args the function takes
args = []
for i from position+1 to block.length:
    value = evaluate_expression(context, block, i, none)
    args.append(value)
```

### ✓ Correct: Evaluate arguments sequentially, tracking position
```
args = []
current_position = position + 1
for each parameter in function.parameters:
    (current_position, arg) = evaluate_expression(context, block, current_position, none)
    args.append(arg)
return (current_position, function.invoke(args))
```

### ❌ Wrong: Separate "parse" and "execute" phases
```
// Don't build an AST!
class FunctionCall:
    name: string
    arguments: list[Expression]
```

### ✓ Correct: On-demand evaluation
```
// Words that resolve to functions invoke immediately during evaluation
case Word:
    value = context.lookup(element.name)
    if value.is_function():
        return invoke_function(context, block, position, value, last_result)
    return (position + 1, value)
```

### ❌ Wrong: Pre-evaluate blocks passed to natives
```
function native_when(context, args, refinements):
    condition = args[0]
    // WRONG - block is already literal, don't evaluate here automatically
    block_result = evaluate_block(context, args[1])
```

### ✓ Correct: Let native control evaluation
```
function native_when(context, args, refinements):
    condition = args[0]
    block = args[1]  // Literal block
    
    // Native decides IF and WHEN to evaluate
    if to_truthy(condition):
        return evaluate_block(context, block.elements)
    
    return none
```

## Testing Strategy

When implementing evaluator features:

1. **Test position tracking**: Verify evaluate-expression returns correct position
2. **Test argument consumption**: Functions should consume exact number of values
3. **Test infix chaining**: `1 + 2 + 3` should work left-to-right
4. **Test refinements**: Flags and value refinements between arguments
5. **Test literal args**: `when false [print "no"]` should NOT print
6. **Test nested expressions**: `x: add 5 multiply 2 3` should work correctly

**Example Test Cases:**
```
Input: "x: 10"
Expected: x bound to 10, returns 10

Input: "add 5 3"
Expected: returns 8 (assuming add function exists)

Input: "5 + 3"
Expected: returns 8 (infix)

Input: "when false [print "no"]"
Expected: returns none, does NOT print

Input: "1 + 2 + 3"
Expected: returns 6 (left-associative)
```

## Summary Checklist

When implementing the evaluator:

- [ ] `evaluate_expression()` returns `(position, value)`
- [ ] `evaluate_block()` loops until position exceeds block length
- [ ] Words automatically invoke functions
- [ ] Get-words return values without invocation
- [ ] Set-words evaluate next expression and bind
- [ ] Function calls consume argument positions
- [ ] Infix functions use `last_result` as first argument
- [ ] Refinements can appear anywhere in argument list
- [ ] Native functions control evaluation of literal arguments
- [ ] No AST or pre-scanning - pure sequential evaluation
- [ ] Position tracking is consistent across all evaluation paths

## Series Position and Copy Behavior

Series (blocks, strings, binary) maintain an internal index position that affects operations:

**Position-Aware Operations**:
- `copy` - Copies from current index to end, result always at head
- `first` - Returns element at current index
- `next` - Advances index by 1
- `tail` - Sets index to end of series

**Example:**
```viro
a: [1 2 3 4]
a: next next a      ; index now at position 2 (element 3)
b: copy a           ; copies [3 4], result at head
first a             ; returns 3 (element at current position)
```

**Important**: The `copy` function only copies remaining elements from the current index position forward. To copy an entire series regardless of position, use `copy head series`.

### Series Operations Comparison

Different series operations handle bounds checking and position awareness differently:

| Operation | Behavior with --part | Validation | Position Awareness |
|-----------|---------------------|------------|-------------------|
| `copy` | Copies N elements from current position | Strict: 0 ≤ N ≤ remaining, error if exceeded | Yes - copies from index onward |
| `take` | Takes N elements from current position | Strict for negative: N < 0 errors<br/>Clamped for overflow: N > remaining returns remaining | Yes - takes from index onward |
| `remove` | Removes N elements from current position | Strict: 0 ≤ N ≤ remaining, error if exceeded | Yes - removes from index onward |
| `first` | Returns element at current position | Error if at tail (no elements remaining) | Yes - operates at current index |
| `length?` | Returns total series length | N/A | No - always returns full length |

**Key Differences**:

1. **copy vs take**: Both work from current position, but `copy` errors on overflow while `take` clamps
2. **Validation timing**: `copy` and `remove` validate at native layer, `take` validates negative only
3. **Result positioning**: `copy` always returns result at head (index 0), `take` also returns at head

**Practical Examples**:
```viro
a: next next [1 2 3 4 5]    ; index at position 2 (element 3)

copy a                       ; returns [3 4 5]
copy --part 2 a             ; returns [3 4] - strict count
copy --part 10 a            ; ERROR: out of bounds (asked for 10, only 3 remaining)

take 2 a                    ; returns [3 4]  
take 10 a                   ; returns [3 4 5] - clamped to remaining

remove --part 2 a           ; removes [3 4], a now [1 2 5] at index 2
remove --part 10 a          ; ERROR: out of bounds

first a                     ; returns 3 (element at current index)
length? a                   ; returns 5 (total length, ignores position)
```

**Design Rationale**:

- **copy strict**: Ensures you get exactly what you asked for or an error (no silent truncation)
- **take clamped**: Convenient for "give me up to N" scenarios where partial results are acceptable  
- **remove strict**: Safety-critical - removing wrong count could corrupt data structures

## Core Principle

**The execution model is fundamentally sequential value consumption with position tracking.**

The evaluator reads values from left to right, one at a time. Each evaluation:
1. Examines the value at the current position
2. Performs type-specific evaluation logic
3. Returns the new position and resulting value

Function calls extend this by consuming additional positions for arguments. Infix functions use the previous result as their first argument. That's it - there are no special cases beyond this simple model.

Master this concept and the rest follows naturally.
