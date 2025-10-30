# Viro Language Examples

This directory contains examples demonstrating Viro's features and capabilities.

## Running Examples

```bash
./viro examples/01_basics.viro
```

Or use the REPL to run individual snippets:

```bash
./viro
>> do %examples/01_basics.viro
```

## Example Files

### 01_basics.viro
Introduction to Viro fundamentals:
- Literals (integers, strings, logic, none)
- Variables and assignment
- Arithmetic operations
- Left-to-right evaluation vs parentheses
- Blocks
- Comparison operators

### 02_control_flow.viro
Control flow constructs:
- `when` - conditional execution
- `if` - if/else branching
- `loop` - fixed iteration
- `while` - conditional iteration

### 03_functions.viro
Function definition and usage:
- Simple functions
- Functions with return values
- Multiple parameters
- Recursive functions (factorial, fibonacci)
- Functions with refinements (flags)

### 04_series.viro
Series (blocks and strings) operations:
- Creating blocks
- Accessing elements: `first`, `last`, `at`
- Series length: `length?`
- Modifying: `append`, `insert`
- Navigation: `head`, `tail`, `next`, `back`
- Working with strings

### 05_data_manipulation.viro
Data manipulation functions:
- `set` / `get` - explicit variable binding
- `type?` - type introspection
- `form` - human-readable formatting
- `mold` - REBOL-readable formatting
- `reduce` - evaluate block elements
- `compose` - selective evaluation
- `join` - string concatenation

### 06_objects.viro
Object-oriented features:
- Creating objects with `make object!`
- Object fields and methods
- Path notation for field access (`object/field`)
- Nested objects
- Object initialization

### 07_advanced.viro
Advanced programming patterns:
- Higher-order functions (map, filter)
- Collection operations (sum, max)
- Algorithms (bubble sort)
- Functional programming techniques

### 08_practical.viro
Practical utility functions:
- Temperature conversion
- Prime number checking
- Greatest Common Divisor (GCD)
- String manipulation
- Word counting

## Key Language Features

### Left-to-Right Evaluation

Viro evaluates expressions left-to-right, unlike most languages with operator precedence:

```viro
3 + 4 * 2    ; Result: 14 (not 11!)
3 + (4 * 2)  ; Result: 11 (parentheses force evaluation)
```

### Blocks vs Parentheses

- **Blocks `[]`**: Deferred evaluation (quoted code)
- **Parentheses `()`**: Immediate evaluation

```viro
[1 + 2]      ; Block containing: 1, +, 2
(1 + 2)      ; Evaluates to: 3
```

### Local-by-Default Scoping

Variables are local to their function by default (safer than REBOL's global-by-default):

```viro
x: 10          ; Global x
func [] [
    x: 20      ; Local x (doesn't affect global)
]
```

### Refinements

Bash-style flags for function options:

```viro
func [a b --add --multiply] [
    if :add [+ a b] [
        if :multiply [* a b] [a]
    ]
]

calculate 5 3 --add       ; 8
calculate 5 3 --multiply  ; 15
```

### Type-Based Dispatch

Different word types have different evaluation semantics:

- `word` - Retrieve value
- `'word` - Literal word (lit-word)
- `:word` - Get value (get-word)
- `word:` - Set value (set-word)

## Learning Path

1. Start with **01_basics.viro** - Learn fundamental syntax
2. Progress to **02_control_flow.viro** - Understand flow control
3. Master **03_functions.viro** - Function definition and recursion
4. Explore **04_series.viro** - Data structure manipulation
5. Study **05_data_manipulation.viro** - Type system and conversion
6. Learn **06_objects.viro** - Object-oriented programming
7. Challenge yourself with **07_advanced.viro** - Advanced patterns
8. Apply knowledge with **08_practical.viro** - Real-world utilities

## Additional Resources

- [Main README](../README.md) - Project overview
- [Specification](../specs/001-implement-the-core/spec.md) - Language specification
- [Quickstart Guide](../specs/001-implement-the-core/quickstart.md) - Build and test instructions
- [Native Function Contracts](../specs/001-implement-the-core/contracts/) - Detailed function documentation

## Contributing Examples

When adding new examples:

1. Follow TDD methodology (write tests first)
2. Keep examples focused and well-commented
3. Use descriptive print statements for output clarity
4. Demonstrate one concept at a time
5. Build from simple to complex
