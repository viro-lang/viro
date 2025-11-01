# Contract: Control Flow Natives

**Category**: Control Flow  
**Functions**: `when`, `if`, `loop`, `while`  
**Purpose**: Conditional execution and iteration primitives

---

## Native: `when`

**Signature**: `when condition [block]`

**Parameters**:
- `condition`: Value (evaluated to logic)
- `block`: Block (executed conditionally)

**Return**: Value from block if condition true, none otherwise

**Behavior**:
1. Evaluate condition expression
2. Convert result to logic (truthy evaluation)
3. If true: evaluate block, return last value
4. If false: return none without evaluating block

**Type Rules**:
- Condition: any value (truthy conversion: none/false → false, all others → true)
- Block: must be Block type (error if not)

**Examples**:
```viro
when true [42]           → 42
when false [42]          → none
when 1 [print "yes"]     → prints "yes", returns none (print returns none)
when none [42]           → none
when x > 10 [
    print "x is large"
]                        → prints if x > 10, returns none

result: when valid? [
    process-data
]                        → result is value or none
```

**Error Cases**:
- Second argument not a block → Script error (300): "Expected block for when body"
- Block evaluation error → propagate error

**Test Cases**:
1. `when true [42]` returns `42`
2. `when false [42]` returns `none`
3. `when 1 [42]` returns `42` (truthy)
4. `when none [42]` returns `none` (falsy)
5. `when true [1 + 1]` evaluates block and returns `2`
6. `when true "string"` errors (not a block)

---

## Native: `if`

**Signature**: `if condition [true-block] [false-block]`

**Parameters**:
- `condition`: Value (evaluated to logic)
- `true-block`: Block (executed if condition true)
- `false-block`: Block (executed if condition false)

**Return**: Value from executed block (last value)

**Behavior**:
1. Evaluate condition expression
2. Convert result to logic (truthy evaluation)
3. If true: evaluate true-block, return last value
4. If false: evaluate false-block, return last value
5. Both blocks required (unlike Viro's if which only has true branch)

**Type Rules**:
- Condition: any value (truthy conversion)
- Both blocks: must be Block type (error if not)

**Examples**:
```viro
if true [1] [2]              → 1
if false [1] [2]             → 2
if 1 < 2 ["less"] ["more"]  → "less"
if none [1] [2]              → 2

result: if x > 10 [
    "large"
] [
    "small"
]

status: if valid? [
    print "Processing..."
    process-data
] [
    print "Error"
    none
]
```

**Error Cases**:
- Second argument not a block → Script error: "Expected block for if true branch"
- Third argument not a block → Script error: "Expected block for if false branch"
- Missing third argument → Script error: "If requires both true and false blocks"
- Block evaluation error → propagate error

**Test Cases**:
1. `if true [1] [2]` returns `1`
2. `if false [1] [2]` returns `2`
3. `if 1 < 2 [10] [20]` returns `10`
4. `if none [10] [20]` returns `20` (none is falsy)
5. `if true [1 + 1] [2 + 2]` returns `2` (evaluates true-block only)
6. `if false [1 + 1] [2 + 2]` returns `4` (evaluates false-block only)
7. `if true 1 [2]` errors (not a block)
8. `if true [1] 2` errors (not a block)
9. `if true [1]` errors (missing false block)

---

## Native: `loop`

**Signature**: `loop count [block]`

**Parameters**:
- `count`: Integer (number of iterations)
- `block`: Block (body to repeat)

**Return**: Value from last iteration (last value of last block evaluation), none if count ≤ 0

**Behavior**:
1. Evaluate count expression
2. Validate count is integer and ≥ 0
3. Execute block count times
4. Return result of last block evaluation

**Type Rules**:
- Count: must be Integer type (error if not)
- Count: must be ≥ 0 (error if negative)
- Block: must be Block type (error if not)

**Examples**:
```viro
loop 3 [print "hi"]    → prints "hi" three times, returns none
loop 0 [print "hi"]    → none (no execution)
loop 5 [42]            → 42 (returns last iteration result)
x: 0  loop 3 [x: x + 1]  → x becomes 3, returns 3
```

**Error Cases**:
- Count not integer → Script error: "Expected integer for loop count"
- Count negative → Script error: "Loop count must be non-negative"
- Block not block → Script error: "Expected block for loop body"
- Block evaluation error → propagate error

**Test Cases**:
1. `loop 3 [42]` returns `42` (last iteration)
2. `loop 0 [42]` returns `none` (no iterations)
3. `loop 1 [42]` returns `42` (single iteration)
4. Counter variable increments: `x: 0  loop 5 [x: x + 1]` results in `x = 5`
5. `loop "3" [42]` errors (not integer)
6. `loop -1 [42]` errors (negative count)
7. `loop 3 42` errors (not block)

---

## Native: `while`

**Signature**: `while [condition] [body]`

**Parameters**:
- `condition`: Block (re-evaluated each iteration for truthiness)
- `body`: Block (executed while condition true)

**Return**: Value from last iteration (last value of last body evaluation), none if never executed

**Behavior**:
1. Evaluate condition block
2. Convert result to logic
3. If true: evaluate body block, go to step 1
4. If false: return result from last body evaluation (or none if never executed)

**Type Rules**:
- Condition: must be Block type (error if not)
- Body: must be Block type (error if not)

**Safety**:
- No automatic timeout (per spec clarification)
- User must interrupt infinite loops via Ctrl+C

**Examples**:
```viro
x: 0  while [x < 3] [x: x + 1]  → x becomes 3, returns 3
while [false] [42]               → none (never executes)
while [true] [42]                → infinite loop (user interrupts)
```

**Error Cases**:
- Condition not block → Script error: "Expected block for while condition"
- Body not block → Script error: "Expected block for while body"
- Condition or body evaluation error → propagate error

**Test Cases**:
1. `x: 0  while [x < 3] [x: x + 1]` results in `x = 3`, returns `3`
2. `while [false] [42]` returns `none` (never executes body)
3. `while [true] [42]` runs indefinitely until interrupted
4. Condition re-evaluated: `x: 0  while [x: x + 1  x < 3] [42]` executes body twice
5. `while true [42]` errors (condition not block)
6. `while [true] 42` errors (body not block)

---

## Common Properties

**Truthy Evaluation** (for all control flow):
- `false` → false
- `none` → false
- All other values → true (including `0`, `""`, `[]`)

**Block Evaluation**:
- Empty block `[]` evaluates to `none`
- Non-empty block returns last evaluated expression

**Error Propagation**:
- Errors in condition or body blocks propagate to caller
- Control flow natives do not catch errors (error handling separate)

**Stack Frames**:
- Each native creates new evaluation context for blocks
- Blocks evaluated with current frame as parent (lexical scoping)

---

## Implementation Checklist

For each native:
- [ ] Function signature matches contract
- [ ] Parameter type validation (return Script error for type mismatch)
- [ ] Truthy conversion implemented correctly
- [ ] Block evaluation uses evaluator (Do_Blk)
- [ ] Return value matches specification
- [ ] All test cases pass
- [ ] Error messages include function name and parameter info

**Specific Requirements**:
- `when`: Single block, returns none if condition false
- `if`: Both blocks required (error if missing), evaluates only one branch
- `loop`: Integer count validation, proper iteration
- `while`: Re-evaluate condition each iteration

**Dependencies**:
- Evaluator (Do_Blk for block evaluation)
- Type system (type checking, truthy conversion)
- Error system (Script error construction)

**Testing Strategy**:
- Table-driven tests with struct { name, args, want, wantErr }
- Each test case from specification becomes table entry
- Parallel execution where safe (no shared state)

**Note**: `either` removed from Viro (compatibility with some languages broken intentionally for clarity)
