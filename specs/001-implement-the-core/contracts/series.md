# Contract: Series Natives

**Category**: Series Operations  
**Functions**: `first`, `last`, `append`, `insert`, `length?`, `next`, `back`, `head`, `tail`, `skip`, `head?`, `tail?`, `index?`  
**Purpose**: Sequential data structure manipulation

---

## Series Protocol

**Series Types** (Phase 1):
- Block: sequence of values
- String: sequence of characters (runes)

**Common Properties**:
- Index-based access (0-based internally, 1-based for REBOL semantics if exposed)
- Mutable (operations modify in-place)
- Bounds-checked operations

---

## Native: `first`

**Signature**: `first series`

**Parameters**:
- `series`: Block or String

**Return**: First element (Value for block, character for string)

**Behavior**: Returns element at position 0

**Type Rules**:
- Argument must be Block or String type (error otherwise)
- Series must be non-empty (error if empty)

**Examples**:
```viro
first [1 2 3]        → 1
first "hello"        → #"h" (character)
first []             → Error: empty series
```

**Error Cases**:
- Argument not series → Script error (300): "First expects series argument"
- Series empty → Script error (300): "Cannot get first of empty series"

**Test Cases**:
1. `first [1 2 3]` returns `1`
2. `first [42]` returns `42` (single element)
3. `first "hello"` returns character `h`
4. `first "x"` returns character `x`
5. `first []` errors (empty block)
6. `first ""` errors (empty string)
7. `first 42` errors (not series)

---

## Native: `last`

**Signature**: `last series`

**Parameters**:
- `series`: Block or String

**Return**: Last element

**Behavior**: Returns element at position (length - 1)

**Examples**:
```viro
last [1 2 3]         → 3
last "hello"         → #"o"
last []              → Error: empty series
```

**Error Cases**: Same as `first`

**Test Cases**:
1. `last [1 2 3]` returns `3`
2. `last [42]` returns `42` (single element)
3. `last "hello"` returns character `o`
4. `last []` errors (empty block)
5. `last ""` errors (empty string)

---

## Native: `append`

**Signature**: `append series value`

**Parameters**:
- `series`: Block or String (modified in-place)
- `value`: Value to add (type must match series)

**Return**: Modified series

**Behavior**: 
- Adds value to end of series
- Modifies series in-place
- Returns the series itself

**Type Rules**:
- First argument must be Block or String
- For Block: value can be any type
- For String: value must be character or string (appends characters)

**Examples**:
```viro
data: [1 2 3]
append data 4        → [1 2 3 4] (data is now [1 2 3 4])

str: "hello"
append str " world"  → "hello world" (str is now "hello world")
```

**Error Cases**:
- First argument not series → Script error: "Append expects series argument"
- String append with non-character/non-string → Script error: "Cannot append non-string to string"

**Test Cases**:
1. `data: [1 2 3]  append data 4` → data becomes `[1 2 3 4]`
2. `append [] 1` → `[1]` (append to empty)
3. `str: "hi"  append str " there"` → str becomes `"hi there"`
4. `append "" "x"` → `"x"` (append to empty string)
5. Block can hold mixed types: `data: [1]  append data "x"` → `[1 "x"]`
6. `append 42 1` errors (not series)

---

## Native: `insert`

**Signature**: `insert series value`

**Parameters**:
- `series`: Block or String (modified in-place)
- `value`: Value to insert (type must match series)

**Return**: Modified series

**Behavior**: 
- Inserts value at beginning (position 0)
- Shifts existing elements right
- Modifies series in-place

**Type Rules**: Same as `append`

**Examples**:
```viro
data: [1 2 3]
insert data 0        → [0 1 2 3] (data is now [0 1 2 3])

str: "world"
insert str "hello "  → "hello world"
```

**Error Cases**: Same as `append`

**Test Cases**:
1. `data: [1 2 3]  insert data 0` → data becomes `[0 1 2 3]`
2. `insert [] 1` → `[1]` (insert to empty)
3. `str: "world"  insert str "hello "` → str becomes `"hello world"`
4. `insert "" "x"` → `"x"`
5. `insert 42 1` errors (not series)

**Note**: Phase 1 inserts at beginning only. Series position-based insert deferred to later phase.

---

## Native: `length?`

**Signature**: `length? series`

**Parameters**:
- `series`: Block or String

**Return**: Integer (number of elements)

**Behavior**: Returns count of elements in series

**Type Rules**:
- Argument must be Block or String type

**Examples**:
```viro
length? [1 2 3]      → 3
length? []           → 0
length? "hello"      → 5
length? ""           → 0
```

**Error Cases**:
- Argument not series → Script error: "Length? expects series argument"

**Test Cases**:
1. `length? [1 2 3]` returns `3`
2. `length? []` returns `0`
3. `length? "hello"` returns `5`
4. `length? ""` returns `0`
5. `length? 42` errors (not series)
6. After append: `data: [1]  append data 2  length? data` → `2`

---

## Native: `next`

**Signature**: `next series`

**Parameters**:
- `series`: Block or String (with position)

**Return**: New series reference at position + 1

**Behavior**: 
- Returns a new series reference advanced one position
- Does not modify the original series
- If already at tail, returns tail (no error)

**Type Rules**:
- Argument must be Block or String type

**Examples**:
```viro
data: [1 2 3]
next data            → series at position 1 (viewing [2 3])
first next data      → 2

str: "hello"
next str             → series at position 1 (viewing "ello")
```

**Error Cases**:
- Argument not series → Script error: "Next expects series argument"

**Test Cases**:
1. `data: [1 2 3]  first next data` returns `2`
2. `str: "hello"  first next str` returns character `e`
3. `data: [1]  next data` → series at tail
4. `next []` → series at tail (empty)
5. `next 42` errors (not series)

---

## Native: `back`

**Signature**: `back series`

**Parameters**:
- `series`: Block or String (with position)

**Return**: New series reference at position - 1

**Behavior**: 
- Returns a new series reference moved back one position
- Does not modify the original series
- If already at head, returns head (no error)

**Type Rules**:
- Argument must be Block or String type

**Examples**:
```viro
data: [1 2 3]
data2: next next data    ; at position 2
back data2               ; at position 1 (viewing [2 3])
first back data2         → 2
```

**Error Cases**:
- Argument not series → Script error: "Back expects series argument"

**Test Cases**:
1. `data: [1 2 3]  data2: next next data  first back data2` returns `2`
2. `data: [1 2 3]  back data` → same as data (already at head)
3. `str: "hello"  str2: next str  first back str2` returns character `h`
4. `back 42` errors (not series)

---

## Native: `head`

**Signature**: `head series`

**Parameters**:
- `series`: Block or String (with position)

**Return**: New series reference at position 0 (head)

**Behavior**: 
- Returns a new series reference at the beginning
- Does not modify the original series
- Always returns head regardless of current position

**Type Rules**:
- Argument must be Block or String type

**Examples**:
```viro
data: [1 2 3]
data2: next next data    ; at position 2
head data2               ; back at position 0 (viewing [1 2 3])
first head data2         → 1
```

**Error Cases**:
- Argument not series → Script error: "Head expects series argument"

**Test Cases**:
1. `data: [1 2 3]  data2: next next data  first head data2` returns `1`
2. `data: [1 2 3]  head data` → same as data (already at head)
3. `str: "hello"  str2: next next str  first head str2` returns character `h`
4. `head 42` errors (not series)

---

## Native: `tail`

**Signature**: `tail series`

**Parameters**:
- `series`: Block or String (with position)

**Return**: New series reference at tail position

**Behavior**: 
- Returns a new series reference at the end (past last element)
- Does not modify the original series
- Tail position is past the last element

**Type Rules**:
- Argument must be Block or String type

**Examples**:
```viro
data: [1 2 3]
tail data                ; at position 3 (past last element)
tail? tail data          → true
```

**Error Cases**:
- Argument not series → Script error: "Tail expects series argument"

**Test Cases**:
1. `data: [1 2 3]  tail? tail data` returns `true`
2. `data: []  tail data` → series at tail (empty series)
3. `str: "hello"  tail? tail str` returns `true`
4. `tail 42` errors (not series)

---

## Native: `skip`

**Signature**: `skip series count`

**Parameters**:
- `series`: Block or String (with position)
- `count`: Integer (positive or negative)

**Return**: New series reference at position + count

**Behavior**: 
- Returns a new series reference advanced by count positions
- Does not modify the original series
- Positive count moves forward, negative moves backward
- Clamps to valid range [0, length] (head to tail)

**Type Rules**:
- First argument must be Block or String
- Second argument must be Integer

**Examples**:
```viro
data: [1 2 3 4 5]
skip data 2              ; at position 2 (viewing [3 4 5])
first skip data 2        → 3

skip data -1             ; moves back (same as back for count -1)
skip data 100            ; clamps to tail
```

**Error Cases**:
- First argument not series → Script error: "Skip expects series argument"
- Second argument not integer → Script error: "Skip expects integer count"

**Test Cases**:
1. `data: [1 2 3 4 5]  first skip data 2` returns `3`
2. `data: [1 2 3]  skip data -1` → same as back data
3. `data: [1 2 3]  skip data 100` → series at tail
4. `data: [1 2 3]  data2: tail data  skip data2 -2` → at position 1
5. `str: "hello"  first skip str 2` returns character `l`
6. `skip 42 1` errors (not series)
7. `skip [1 2 3] "x"` errors (count not integer)

---

## Native: `head?`

**Signature**: `head? series`

**Parameters**:
- `series`: Block or String (with position)

**Return**: Logic (true if at head, false otherwise)

**Behavior**: 
- Returns true if series is at position 0
- Returns false otherwise

**Type Rules**:
- Argument must be Block or String type

**Examples**:
```viro
data: [1 2 3]
head? data               → true
head? next data          → false
head? head next data     → true
```

**Error Cases**:
- Argument not series → Script error: "Head? expects series argument"

**Test Cases**:
1. `data: [1 2 3]  head? data` returns `true`
2. `data: [1 2 3]  head? next data` returns `false`
3. `data: [1 2 3]  head? head next data` returns `true`
4. `str: "hello"  head? str` returns `true`
5. `str: "hello"  head? next str` returns `false`
6. `head? 42` errors (not series)

---

## Native: `tail?`

**Signature**: `tail? series`

**Parameters**:
- `series`: Block or String (with position)

**Return**: Logic (true if at tail, false otherwise)

**Behavior**: 
- Returns true if series is at tail position (past last element)
- Returns false otherwise

**Type Rules**:
- Argument must be Block or String type

**Examples**:
```viro
data: [1 2 3]
tail? data               → false
tail? tail data          → true
tail? next next next data → true
```

**Error Cases**:
- Argument not series → Script error: "Tail? expects series argument"

**Test Cases**:
1. `data: [1 2 3]  tail? data` returns `false`
2. `data: [1 2 3]  tail? tail data` returns `true`
3. `data: [1 2 3]  tail? next next next data` returns `true`
4. `data: []  tail? data` returns `true` (empty series at tail)
5. `str: "hello"  tail? str` returns `false`
6. `str: "hello"  tail? tail str` returns `true`
7. `tail? 42` errors (not series)

---

## Native: `index?`

**Signature**: `index? series`

**Parameters**:
- `series`: Block or String (with position)

**Return**: Integer (1-based position)

**Behavior**: 
- Returns current position in series as 1-based integer
- Head position returns 1
- Tail position returns (length + 1)

**Type Rules**:
- Argument must be Block or String type

**Examples**:
```viro
data: [1 2 3]
index? data              → 1
index? next data         → 2
index? tail data         → 4
```

**Error Cases**:
- Argument not series → Script error: "Index? expects series argument"

**Test Cases**:
1. `data: [1 2 3]  index? data` returns `1`
2. `data: [1 2 3]  index? next data` returns `2`
3. `data: [1 2 3]  index? tail data` returns `4`
4. `data: [1 2 3]  index? next next data` returns `3`
5. `str: "hello"  index? str` returns `1`
6. `str: "hello"  index? next next str` returns `3`
7. `data: []  index? data` returns `1`
8. `index? 42` errors (not series)

---

## Common Properties

**Series Mutability**:
- `append` and `insert` modify series in-place
- Original series reference remains valid
- Return value is the modified series (enables chaining if desired)

**Type Checking**:
- All natives validate series type first
- Clear error messages for type mismatches

**Empty Series Handling**:
- `first` and `last` error on empty series
- `append` and `insert` work on empty series
- `length?` returns 0 for empty series

**String Operations**:
- Strings treated as character (rune) sequences
- String append/insert can take character or string
- Character representation: Go rune type (Unicode code point)

**Block Operations**:
- Blocks can contain any value types (heterogeneous)
- No automatic type coercion

**Error Messages**:
- Include operation name and received types
- Example: "First expects series argument, got integer"

---

## Series Semantics

**REBOL Series Model**:
- Series have current position (internal index)
- Each series reference maintains its own position
- Position ranges from 0 (head) to length (tail)
- Multiple references can point to same underlying data at different positions
- Navigation functions (`next`, `back`, `head`, `tail`, `skip`) return new references
- Position is part of the series reference, not the underlying data

**Position Model**:
```viro
data: [1 2 3]        ; position 0 (head)
data2: next data     ; position 1 (new reference)
first data           → 1 (data still at position 0)
first data2          → 2 (data2 at position 1)
```

**Memory Semantics**:
- Series use Go slices internally (automatic growth)
- Append may reallocate (transparent to user)
- Series values are references (modifications visible to all holders)
- Position is stored separately for each series reference

**Example**:
```viro
a: [1 2 3]
b: a              ; b references same series at same position
append b 4
; Now both a and b see [1 2 3 4]

c: next a         ; c is new reference at position 1
first a           → 1 (a at position 0)
first c           → 2 (c at position 1)
append c 5        ; modifies underlying series
; Now a sees [1 2 3 4 5] and c sees [2 3 4 5] (same data, different positions)
```

**Copy Semantics** (deferred to later phase):
- Phase 1: no `copy` native initially, but series position model is implemented
- Mutations visible through all references to same underlying data
- Each reference maintains its own position
- Deep copy deferred per out-of-scope items

---

## Implementation Checklist

For each native:
- [ ] Function signature matches contract
- [ ] Type validation with clear errors
- [ ] Bounds checking (first/last on non-empty)
- [ ] In-place modification (append/insert)
- [ ] Return correct value
- [ ] All test cases pass
- [ ] Handle empty series correctly

**Dependencies**:
- Value system (Block, String types)
- Type system (series type checking)
- Error system (Script error construction)

**Testing Strategy**:
- Table-driven tests with various series sizes
- Empty series edge cases
- Type mismatch error cases
- Mutation verification (check series contents after operation)
- Reference semantics tests (shared series modifications)

**Performance Considerations**:
- Append to block: amortized O(1) (Go slice growth)
- Insert at beginning: O(n) (shift elements)
- First/last: O(1) access
- Length: O(1) (slice length)
- Navigation (`next`, `back`, `skip`): O(1) (create new reference with different position)
- Position queries (`head?`, `tail?`, `index?`): O(1)

**Future Extensions** (out of Phase 1 scope):
- Series slicing with `copy --part`
- Series search (`find`, `select`)
- Series sorting and transformation
- Advanced insertion at arbitrary positions (`insert at`)
- Series comparison and set operations

````
