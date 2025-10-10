// Package native provides built-in native functions for the Viro interpreter.
//
// Native functions are implemented in Go and registered in the global Registry.
// They are invoked by the evaluator when a function value with native type
// is called.
//
// Categories:
//   - Math: +, -, *, /, <, >, <=, >=, =, <>, and, or, not
//   - Control: when, if, loop, while
//   - Series: first, last, append, insert, length?
//   - Data: set, get, type?
//   - Function: fn (function definition)
//   - I/O: print, input
//
// Native types:
//   - Simple natives (NativeFunc): Don't need evaluator access
//   - Eval natives (NativeFuncWithEval): Need evaluator for code evaluation
//
// All natives are registered in the Registry map with metadata (arity, eval requirement).
package native

import (
	"fmt"

	"github.com/marcin-radoszewski/viro/internal/value"
	"github.com/marcin-radoszewski/viro/internal/verror"
)

// Evaluator interface for natives that need to evaluate code.
type Evaluator interface {
	Do_Blk(vals []value.Value) (value.Value, *verror.Error)
	Do_Next(val value.Value) (value.Value, *verror.Error)
}

// NativeFunc is the signature for simple native functions (like math operations).
type NativeFunc func([]value.Value) (value.Value, *verror.Error)

// NativeFuncWithEval is the signature for natives that need evaluator access (like control flow).
type NativeFuncWithEval func([]value.Value, Evaluator) (value.Value, *verror.Error)

// NativeInfo wraps a native function with metadata.
type NativeInfo struct {
	Func      NativeFunc         // Simple native (if NeedsEval is false)
	FuncEval  NativeFuncWithEval // Native needing evaluator (if NeedsEval is true)
	NeedsEval bool               // True if this native needs evaluator access
	Arity     int                // Number of arguments expected
	Infix     bool               // True if this function uses infix notation (consumes lastResult as first arg)
	EvalArgs  []bool             // NEW: per-arg evaluation control (nil = all eval)
	Doc       *NativeDoc         // Documentation metadata (nil for undocumented functions)
}

// Registry holds all registered native functions.
var Registry = make(map[string]*NativeInfo)

// FunctionRegistry holds native functions as FunctionValue instances (new unified representation).
// This is the new registry that will eventually replace Registry after migration is complete.
var FunctionRegistry = make(map[string]*value.FunctionValue)

func init() {
	// Register math natives (simple - don't need evaluator)
	Registry["+"] = &NativeInfo{
		Func:      Add,
		NeedsEval: false,
		Arity:     2,
		Infix:     true,
		Doc: &NativeDoc{
			Category: "Math",
			Summary:  "Adds two numbers together",
			Description: `Performs addition on two numeric values (integers or decimals).
Returns an integer if both operands are integers, otherwise returns a decimal.
Supports infix notation for natural mathematical expressions.`,
			Parameters: []ParamDoc{
				{Name: "left", Type: "integer! decimal!", Description: "The first number to add", Optional: false},
				{Name: "right", Type: "integer! decimal!", Description: "The second number to add", Optional: false},
			},
			Returns:  "[integer! decimal!] The sum of the two numbers",
			Examples: []string{"3 + 4  ; => 7", "2.5 + 1.5  ; => 4.0", "10 + -5  ; => 5"},
			SeeAlso:  []string{"-", "*", "/"}, Tags: []string{"arithmetic", "math", "addition"},
		},
	}
	Registry["-"] = &NativeInfo{
		Func:      Subtract,
		NeedsEval: false,
		Arity:     2,
		Infix:     true,
		Doc: &NativeDoc{
			Category: "Math",
			Summary:  "Subtracts the second number from the first",
			Description: `Performs subtraction on two numeric values (integers or decimals).
Returns an integer if both operands are integers, otherwise returns a decimal.
Supports infix notation for natural mathematical expressions.`,
			Parameters: []ParamDoc{
				{Name: "left", Type: "integer! decimal!", Description: "The number to subtract from", Optional: false},
				{Name: "right", Type: "integer! decimal!", Description: "The number to subtract", Optional: false},
			},
			Returns:  "[integer! decimal!] The difference between the two numbers",
			Examples: []string{"10 - 3  ; => 7", "5.5 - 2.0  ; => 3.5", "0 - 5  ; => -5"},
			SeeAlso:  []string{"+", "*", "/"}, Tags: []string{"arithmetic", "math", "subtraction"},
		},
	}
	Registry["*"] = &NativeInfo{
		Func:      Multiply,
		NeedsEval: false,
		Arity:     2,
		Infix:     true,
		Doc: &NativeDoc{
			Category: "Math",
			Summary:  "Multiplies two numbers together",
			Description: `Performs multiplication on two numeric values (integers or decimals).
Returns an integer if both operands are integers, otherwise returns a decimal.
Supports infix notation for natural mathematical expressions.`,
			Parameters: []ParamDoc{
				{Name: "left", Type: "integer! decimal!", Description: "The first number to multiply", Optional: false},
				{Name: "right", Type: "integer! decimal!", Description: "The second number to multiply", Optional: false},
			},
			Returns:  "[integer! decimal!] The product of the two numbers",
			Examples: []string{"3 * 4  ; => 12", "2.5 * 2.0  ; => 5.0", "7 * -2  ; => -14"},
			SeeAlso:  []string{"+", "-", "/", "pow"}, Tags: []string{"arithmetic", "math", "multiplication"},
		},
	}
	Registry["/"] = &NativeInfo{
		Func:      Divide,
		NeedsEval: false,
		Arity:     2,
		Infix:     true,
		Doc: &NativeDoc{
			Category: "Math",
			Summary:  "Divides the first number by the second",
			Description: `Performs division on two numeric values (integers or decimals).
Always returns a decimal for precision, even when dividing integers.
Raises an error if dividing by zero.`,
			Parameters: []ParamDoc{
				{Name: "left", Type: "integer! decimal!", Description: "The dividend (number to be divided)", Optional: false},
				{Name: "right", Type: "integer! decimal!", Description: "The divisor (number to divide by)", Optional: false},
			},
			Returns:  "[decimal!] The quotient of the division",
			Examples: []string{"10 / 2  ; => 5.0", "7 / 2  ; => 3.5", "1.0 / 4.0  ; => 0.25"},
			SeeAlso:  []string{"+", "-", "*", "pow"}, Tags: []string{"arithmetic", "math", "division"},
		},
	}

	// Register comparison operators
	Registry["<"] = &NativeInfo{
		Func:      LessThan,
		NeedsEval: false,
		Arity:     2,
		Infix:     true,
		Doc: &NativeDoc{
			Category: "Math",
			Summary:  "Tests if the first number is less than the second",
			Description: `Compares two numeric values and returns true if the first is less than the second.
Works with both integers and decimals. Uses lexicographic ordering for strings.`,
			Parameters: []ParamDoc{
				{Name: "left", Type: "integer! decimal! string!", Description: "The first value to compare", Optional: false},
				{Name: "right", Type: "integer! decimal! string!", Description: "The second value to compare", Optional: false},
			},
			Returns:  "[logic!] True if left < right, false otherwise",
			Examples: []string{"3 < 5  ; => true", "10 < 10  ; => false", "5 < 3  ; => false"},
			SeeAlso:  []string{">", "<=", ">=", "=", "<>"}, Tags: []string{"comparison", "math", "logic"},
		},
	}
	Registry[">"] = &NativeInfo{
		Func:      GreaterThan,
		NeedsEval: false,
		Arity:     2,
		Infix:     true,
		Doc: &NativeDoc{
			Category: "Math",
			Summary:  "Tests if the first number is greater than the second",
			Description: `Compares two numeric values and returns true if the first is greater than the second.
Works with both integers and decimals. Uses lexicographic ordering for strings.`,
			Parameters: []ParamDoc{
				{Name: "left", Type: "integer! decimal! string!", Description: "The first value to compare", Optional: false},
				{Name: "right", Type: "integer! decimal! string!", Description: "The second value to compare", Optional: false},
			},
			Returns:  "[logic!] True if left > right, false otherwise",
			Examples: []string{"10 > 5  ; => true", "10 > 10  ; => false", "3 > 5  ; => false"},
			SeeAlso:  []string{"<", "<=", ">=", "=", "<>"}, Tags: []string{"comparison", "math", "logic"},
		},
	}
	Registry["<="] = &NativeInfo{
		Func:      LessOrEqual,
		NeedsEval: false,
		Arity:     2,
		Infix:     true,
		Doc: &NativeDoc{
			Category: "Math",
			Summary:  "Tests if the first number is less than or equal to the second",
			Description: `Compares two numeric values and returns true if the first is less than or equal to the second.
Works with both integers and decimals. Uses lexicographic ordering for strings.`,
			Parameters: []ParamDoc{
				{Name: "left", Type: "integer! decimal! string!", Description: "The first value to compare", Optional: false},
				{Name: "right", Type: "integer! decimal! string!", Description: "The second value to compare", Optional: false},
			},
			Returns:  "[logic!] True if left <= right, false otherwise",
			Examples: []string{"3 <= 5  ; => true", "10 <= 10  ; => true", "15 <= 10  ; => false"},
			SeeAlso:  []string{"<", ">", ">=", "=", "<>"}, Tags: []string{"comparison", "math", "logic"},
		},
	}
	Registry[">="] = &NativeInfo{
		Func:      GreaterOrEqual,
		NeedsEval: false,
		Arity:     2,
		Infix:     true,
		Doc: &NativeDoc{
			Category: "Math",
			Summary:  "Tests if the first number is greater than or equal to the second",
			Description: `Compares two numeric values and returns true if the first is greater than or equal to the second.
Works with both integers and decimals. Uses lexicographic ordering for strings.`,
			Parameters: []ParamDoc{
				{Name: "left", Type: "integer! decimal! string!", Description: "The first value to compare", Optional: false},
				{Name: "right", Type: "integer! decimal! string!", Description: "The second value to compare", Optional: false},
			},
			Returns:  "[logic!] True if left >= right, false otherwise",
			Examples: []string{"10 >= 5  ; => true", "10 >= 10  ; => true", "3 >= 5  ; => false"},
			SeeAlso:  []string{"<", ">", "<=", "=", "<>"}, Tags: []string{"comparison", "math", "logic"},
		},
	}
	Registry["="] = &NativeInfo{
		Func:      Equal,
		NeedsEval: false,
		Arity:     2,
		Infix:     true,
		Doc: &NativeDoc{
			Category: "Math",
			Summary:  "Tests if two values are equal",
			Description: `Compares two values for equality. Works with all value types including
integers, decimals, strings, blocks, and objects. Returns true if values are equivalent.`,
			Parameters: []ParamDoc{
				{Name: "left", Type: "any-type!", Description: "The first value to compare", Optional: false},
				{Name: "right", Type: "any-type!", Description: "The second value to compare", Optional: false},
			},
			Returns:  "[logic!] True if values are equal, false otherwise",
			Examples: []string{"5 = 5  ; => true", "3 = 4  ; => false", `"hello" = "hello"  ; => true`},
			SeeAlso:  []string{"<>", "<", ">", "<=", ">="}, Tags: []string{"comparison", "equality", "logic"},
		},
	}
	Registry["<>"] = &NativeInfo{
		Func:      NotEqual,
		NeedsEval: false,
		Arity:     2,
		Infix:     true,
		Doc: &NativeDoc{
			Category: "Math",
			Summary:  "Tests if two values are not equal",
			Description: `Compares two values for inequality. Works with all value types including
integers, decimals, strings, blocks, and objects. Returns true if values differ.`,
			Parameters: []ParamDoc{
				{Name: "left", Type: "any-type!", Description: "The first value to compare", Optional: false},
				{Name: "right", Type: "any-type!", Description: "The second value to compare", Optional: false},
			},
			Returns:  "[logic!] True if values are not equal, false otherwise",
			Examples: []string{"3 <> 4  ; => true", "5 <> 5  ; => false", `"hello" <> "world"  ; => true`},
			SeeAlso:  []string{"=", "<", ">", "<=", ">="}, Tags: []string{"comparison", "inequality", "logic"},
		},
	}

	// Register logic operators
	Registry["and"] = &NativeInfo{
		Func:      And,
		NeedsEval: false,
		Arity:     2,
		Infix:     true,
		Doc: &NativeDoc{
			Category: "Math",
			Summary:  "Performs logical AND on two boolean values",
			Description: `Returns true only if both operands are true. In viro, any non-zero integer
and non-empty string is considered true; zero, empty strings, and false are considered false.`,
			Parameters: []ParamDoc{
				{Name: "left", Type: "logic! integer!", Description: "The first boolean value", Optional: false},
				{Name: "right", Type: "logic! integer!", Description: "The second boolean value", Optional: false},
			},
			Returns:  "[logic!] True if both values are true, false otherwise",
			Examples: []string{"true and true  ; => true", "true and false  ; => false", "1 and 1  ; => true"},
			SeeAlso:  []string{"or", "not"}, Tags: []string{"logic", "boolean", "and"},
		},
	}
	Registry["or"] = &NativeInfo{
		Func:      Or,
		NeedsEval: false,
		Arity:     2,
		Infix:     true,
		Doc: &NativeDoc{
			Category: "Math",
			Summary:  "Performs logical OR on two boolean values",
			Description: `Returns true if at least one operand is true. In viro, any non-zero integer
and non-empty string is considered true; zero, empty strings, and false are considered false.`,
			Parameters: []ParamDoc{
				{Name: "left", Type: "logic! integer!", Description: "The first boolean value", Optional: false},
				{Name: "right", Type: "logic! integer!", Description: "The second boolean value", Optional: false},
			},
			Returns:  "[logic!] True if either value is true, false if both are false",
			Examples: []string{"true or false  ; => true", "false or false  ; => false", "0 or 1  ; => true"},
			SeeAlso:  []string{"and", "not"}, Tags: []string{"logic", "boolean", "or"},
		},
	}
	Registry["not"] = &NativeInfo{
		Func:      Not,
		NeedsEval: false,
		Arity:     1,
		Doc: &NativeDoc{
			Category: "Math",
			Summary:  "Negates a boolean value",
			Description: `Returns the logical negation of a value. True becomes false, false becomes true.
In viro, any non-zero integer and non-empty string is considered true; zero, empty strings, and false are considered false.`,
			Parameters: []ParamDoc{
				{Name: "value", Type: "logic! integer!", Description: "The boolean value to negate", Optional: false},
			},
			Returns:  "[logic!] The negated boolean value",
			Examples: []string{"not true  ; => false", "not false  ; => true", "not 0  ; => true", "not 1  ; => false"},
			SeeAlso:  []string{"and", "or"}, Tags: []string{"logic", "boolean", "not", "negation"},
		},
	}

	// Register decimal and advanced math natives (Feature 002)
	Registry["decimal"] = &NativeInfo{
		Func:      DecimalConstructor,
		NeedsEval: false,
		Arity:     1,
		Doc: &NativeDoc{
			Category: "Math",
			Summary:  "Converts a value to high-precision decimal",
			Description: `Converts an integer or string to a decimal value with arbitrary precision.
Useful for financial calculations and when standard floating-point arithmetic is insufficient.`,
			Parameters: []ParamDoc{
				{Name: "value", Type: "integer! decimal! string!", Description: "The value to convert to decimal", Optional: false},
			},
			Returns:  "[decimal!] The decimal representation of the input value",
			Examples: []string{`decimal 42  ; => 42.0`, `decimal "3.14159265358979323846"  ; => 3.14159265358979323846`},
			SeeAlso:  []string{"pow", "sqrt"}, Tags: []string{"conversion", "decimal", "precision"},
		},
	}
	Registry["pow"] = &NativeInfo{
		Func:      Pow,
		NeedsEval: false,
		Arity:     2,
		Doc: &NativeDoc{
			Category: "Math",
			Summary:  "Raises a number to a power",
			Description: `Calculates base raised to the exponent power (base^exponent).
Always returns a decimal for precision. Supports negative exponents for fractional results.`,
			Parameters: []ParamDoc{
				{Name: "base", Type: "integer! decimal!", Description: "The base number", Optional: false},
				{Name: "exponent", Type: "integer! decimal!", Description: "The power to raise the base to", Optional: false},
			},
			Returns:  "[decimal!] The result of base^exponent",
			Examples: []string{"pow 2 8  ; => 256.0", "pow 10 -2  ; => 0.01", "pow 2.5 2  ; => 6.25"},
			SeeAlso:  []string{"sqrt", "exp", "*"}, Tags: []string{"math", "power", "exponent"},
		},
	}
	Registry["sqrt"] = &NativeInfo{
		Func:      Sqrt,
		NeedsEval: false,
		Arity:     1,
		Doc: &NativeDoc{
			Category: "Math",
			Summary:  "Calculates the square root of a number",
			Description: `Returns the square root of a number. Always returns a decimal for precision.
Raises an error if the input is negative.`,
			Parameters: []ParamDoc{
				{Name: "value", Type: "integer! decimal!", Description: "The number to take the square root of (must be non-negative)", Optional: false},
			},
			Returns:  "[decimal!] The square root of the input",
			Examples: []string{"sqrt 16  ; => 4.0", "sqrt 2  ; => 1.414213562373095", "sqrt 0  ; => 0.0"},
			SeeAlso:  []string{"pow", "exp"}, Tags: []string{"math", "root", "square"},
		},
	}
	Registry["exp"] = &NativeInfo{
		Func:      Exp,
		NeedsEval: false,
		Arity:     1,
		Doc: &NativeDoc{
			Category: "Math",
			Summary:  "Calculates e raised to a power",
			Description: `Returns e (Euler's number, approximately 2.71828) raised to the given power.
Useful for exponential growth calculations and mathematical analysis.`,
			Parameters: []ParamDoc{
				{Name: "exponent", Type: "integer! decimal!", Description: "The power to raise e to", Optional: false},
			},
			Returns:  "[decimal!] The value of e^exponent",
			Examples: []string{"exp 0  ; => 1.0", "exp 1  ; => 2.718281828459045", "exp 2  ; => 7.38905609893065"},
			SeeAlso:  []string{"log", "pow"}, Tags: []string{"math", "exponential", "euler"},
		},
	}
	Registry["log"] = &NativeInfo{
		Func:      Log,
		NeedsEval: false,
		Arity:     1,
		Doc: &NativeDoc{
			Category: "Math",
			Summary:  "Calculates the natural logarithm (base e)",
			Description: `Returns the natural logarithm (ln) of a number, which is the logarithm to base e.
The input must be positive. This is the inverse of the exp function.`,
			Parameters: []ParamDoc{
				{Name: "value", Type: "integer! decimal!", Description: "The number to take the logarithm of (must be positive)", Optional: false},
			},
			Returns:  "[decimal!] The natural logarithm of the input",
			Examples: []string{"log 1  ; => 0.0", "log 2.718281828459045  ; => 1.0", "log 10  ; => 2.302585092994046"},
			SeeAlso:  []string{"exp", "log-10", "pow"}, Tags: []string{"math", "logarithm", "natural"},
		},
	}
	Registry["log-10"] = &NativeInfo{
		Func:      Log10,
		NeedsEval: false,
		Arity:     1,
		Doc: &NativeDoc{
			Category: "Math",
			Summary:  "Calculates the base-10 logarithm",
			Description: `Returns the logarithm to base 10 of a number. The input must be positive.
Useful for scientific calculations and order-of-magnitude estimations.`,
			Parameters: []ParamDoc{
				{Name: "value", Type: "integer! decimal!", Description: "The number to take the logarithm of (must be positive)", Optional: false},
			},
			Returns:  "[decimal!] The base-10 logarithm of the input",
			Examples: []string{"log-10 1  ; => 0.0", "log-10 10  ; => 1.0", "log-10 100  ; => 2.0", "log-10 1000  ; => 3.0"},
			SeeAlso:  []string{"log", "exp", "pow"}, Tags: []string{"math", "logarithm", "base10"},
		},
	}
	Registry["sin"] = &NativeInfo{
		Func:      Sin,
		NeedsEval: false,
		Arity:     1,
		Doc: &NativeDoc{
			Category: "Math",
			Summary:  "Calculates the sine of an angle",
			Description: `Returns the sine of an angle given in radians.
Use multiplication by pi/180 to convert from degrees to radians.`,
			Parameters: []ParamDoc{
				{Name: "angle", Type: "integer! decimal!", Description: "The angle in radians", Optional: false},
			},
			Returns:  "[decimal!] The sine of the angle",
			Examples: []string{"sin 0  ; => 0.0", "sin 1.5707963267948966  ; => 1.0 (pi/2)", "sin 3.141592653589793  ; => 0.0 (pi)"},
			SeeAlso:  []string{"cos", "tan", "asin"}, Tags: []string{"math", "trigonometry", "sine"},
		},
	}
	Registry["cos"] = &NativeInfo{
		Func:      Cos,
		NeedsEval: false,
		Arity:     1,
		Doc: &NativeDoc{
			Category: "Math",
			Summary:  "Calculates the cosine of an angle",
			Description: `Returns the cosine of an angle given in radians.
Use multiplication by pi/180 to convert from degrees to radians.`,
			Parameters: []ParamDoc{
				{Name: "angle", Type: "integer! decimal!", Description: "The angle in radians", Optional: false},
			},
			Returns:  "[decimal!] The cosine of the angle",
			Examples: []string{"cos 0  ; => 1.0", "cos 1.5707963267948966  ; => 0.0 (pi/2)", "cos 3.141592653589793  ; => -1.0 (pi)"},
			SeeAlso:  []string{"sin", "tan", "acos"}, Tags: []string{"math", "trigonometry", "cosine"},
		},
	}
	Registry["tan"] = &NativeInfo{
		Func:      Tan,
		NeedsEval: false,
		Arity:     1,
		Doc: &NativeDoc{
			Category: "Math",
			Summary:  "Calculates the tangent of an angle",
			Description: `Returns the tangent of an angle given in radians.
Use multiplication by pi/180 to convert from degrees to radians.`,
			Parameters: []ParamDoc{
				{Name: "angle", Type: "integer! decimal!", Description: "The angle in radians", Optional: false},
			},
			Returns:  "[decimal!] The tangent of the angle",
			Examples: []string{"tan 0  ; => 0.0", "tan 0.7853981633974483  ; => 1.0 (pi/4)"},
			SeeAlso:  []string{"sin", "cos", "atan"}, Tags: []string{"math", "trigonometry", "tangent"},
		},
	}
	Registry["asin"] = &NativeInfo{
		Func:      Asin,
		NeedsEval: false,
		Arity:     1,
		Doc: &NativeDoc{
			Category: "Math",
			Summary:  "Calculates the arcsine (inverse sine) of a value",
			Description: `Returns the angle in radians whose sine is the given value.
The input must be between -1 and 1 (inclusive). Result is in range [-pi/2, pi/2].`,
			Parameters: []ParamDoc{
				{Name: "value", Type: "integer! decimal!", Description: "The sine value (must be between -1 and 1)", Optional: false},
			},
			Returns:  "[decimal!] The angle in radians",
			Examples: []string{"asin 0  ; => 0.0", "asin 1  ; => 1.5707963267948966 (pi/2)", "asin -1  ; => -1.5707963267948966 (-pi/2)"},
			SeeAlso:  []string{"sin", "acos", "atan"}, Tags: []string{"math", "trigonometry", "arcsine", "inverse"},
		},
	}
	Registry["acos"] = &NativeInfo{
		Func:      Acos,
		NeedsEval: false,
		Arity:     1,
		Doc: &NativeDoc{
			Category: "Math",
			Summary:  "Calculates the arccosine (inverse cosine) of a value",
			Description: `Returns the angle in radians whose cosine is the given value.
The input must be between -1 and 1 (inclusive). Result is in range [0, pi].`,
			Parameters: []ParamDoc{
				{Name: "value", Type: "integer! decimal!", Description: "The cosine value (must be between -1 and 1)", Optional: false},
			},
			Returns:  "[decimal!] The angle in radians",
			Examples: []string{"acos 1  ; => 0.0", "acos 0  ; => 1.5707963267948966 (pi/2)", "acos -1  ; => 3.141592653589793 (pi)"},
			SeeAlso:  []string{"cos", "asin", "atan"}, Tags: []string{"math", "trigonometry", "arccosine", "inverse"},
		},
	}
	Registry["atan"] = &NativeInfo{
		Func:      Atan,
		NeedsEval: false,
		Arity:     1,
		Doc: &NativeDoc{
			Category: "Math",
			Summary:  "Calculates the arctangent (inverse tangent) of a value",
			Description: `Returns the angle in radians whose tangent is the given value.
Accepts any real number as input. Result is in range (-pi/2, pi/2).`,
			Parameters: []ParamDoc{
				{Name: "value", Type: "integer! decimal!", Description: "The tangent value", Optional: false},
			},
			Returns:  "[decimal!] The angle in radians",
			Examples: []string{"atan 0  ; => 0.0", "atan 1  ; => 0.7853981633974483 (pi/4)", "atan -1  ; => -0.7853981633974483 (-pi/4)"},
			SeeAlso:  []string{"tan", "asin", "acos"}, Tags: []string{"math", "trigonometry", "arctangent", "inverse"},
		},
	}
	Registry["round"] = &NativeInfo{
		Func:      Round,
		NeedsEval: false,
		Arity:     1,
		Doc: &NativeDoc{
			Category: "Math",
			Summary:  "Rounds a number to the nearest integer",
			Description: `Rounds a decimal number to the nearest integer using standard rounding rules
(0.5 rounds up). Returns an integer value.`,
			Parameters: []ParamDoc{
				{Name: "value", Type: "integer! decimal!", Description: "The number to round", Optional: false},
			},
			Returns:  "[integer!] The rounded integer value",
			Examples: []string{"round 3.2  ; => 3", "round 3.7  ; => 4", "round 3.5  ; => 4", "round -2.5  ; => -2"},
			SeeAlso:  []string{"ceil", "floor", "truncate"}, Tags: []string{"math", "rounding"},
		},
	}
	Registry["ceil"] = &NativeInfo{
		Func:      Ceil,
		NeedsEval: false,
		Arity:     1,
		Doc: &NativeDoc{
			Category: "Math",
			Summary:  "Rounds a number up to the nearest integer",
			Description: `Returns the smallest integer greater than or equal to the input (ceiling function).
Always rounds upward, even for negative numbers.`,
			Parameters: []ParamDoc{
				{Name: "value", Type: "integer! decimal!", Description: "The number to round up", Optional: false},
			},
			Returns:  "[integer!] The ceiling value",
			Examples: []string{"ceil 3.1  ; => 4", "ceil 3.9  ; => 4", "ceil -2.1  ; => -2", "ceil 5  ; => 5"},
			SeeAlso:  []string{"floor", "round", "truncate"}, Tags: []string{"math", "rounding", "ceiling"},
		},
	}
	Registry["floor"] = &NativeInfo{
		Func:      Floor,
		NeedsEval: false,
		Arity:     1,
		Doc: &NativeDoc{
			Category: "Math",
			Summary:  "Rounds a number down to the nearest integer",
			Description: `Returns the largest integer less than or equal to the input (floor function).
Always rounds downward, even for negative numbers.`,
			Parameters: []ParamDoc{
				{Name: "value", Type: "integer! decimal!", Description: "The number to round down", Optional: false},
			},
			Returns:  "[integer!] The floor value",
			Examples: []string{"floor 3.1  ; => 3", "floor 3.9  ; => 3", "floor -2.1  ; => -3", "floor 5  ; => 5"},
			SeeAlso:  []string{"ceil", "round", "truncate"}, Tags: []string{"math", "rounding", "floor"},
		},
	}
	Registry["truncate"] = &NativeInfo{
		Func:      Truncate,
		NeedsEval: false,
		Arity:     1,
		Doc: &NativeDoc{
			Category: "Math",
			Summary:  "Truncates a number toward zero",
			Description: `Removes the fractional part of a number, rounding toward zero.
For positive numbers, behaves like floor; for negative numbers, behaves like ceil.`,
			Parameters: []ParamDoc{
				{Name: "value", Type: "integer! decimal!", Description: "The number to truncate", Optional: false},
			},
			Returns:  "[integer!] The truncated integer value",
			Examples: []string{"truncate 3.7  ; => 3", "truncate -3.7  ; => -3", "truncate 5  ; => 5"},
			SeeAlso:  []string{"floor", "ceil", "round"}, Tags: []string{"math", "rounding", "truncate"},
		},
	}

	// Register series natives
	Registry["first"] = &NativeInfo{
		Func:      First,
		NeedsEval: false,
		Arity:     1,
		Doc: &NativeDoc{
			Category: "Series",
			Summary:  "Returns the first element of a series",
			Description: `Gets the first element of a block or string. Raises an error if the series is empty.
For strings, returns the first character as a string.`,
			Parameters: []ParamDoc{
				{Name: "series", Type: "block! string!", Description: "The series to get the first element from", Optional: false},
			},
			Returns:  "[any-type!] The first element of the series",
			Examples: []string{"first [1 2 3]  ; => 1", `first "hello"  ; => "h"`, "first [[a b] c]  ; => [a b]"},
			SeeAlso:  []string{"last", "length?", "append", "insert"}, Tags: []string{"series", "access", "first"},
		},
	}
	Registry["last"] = &NativeInfo{
		Func:      Last,
		NeedsEval: false,
		Arity:     1,
		Doc: &NativeDoc{
			Category: "Series",
			Summary:  "Returns the last element of a series",
			Description: `Gets the last element of a block or string. Raises an error if the series is empty.
For strings, returns the last character as a string.`,
			Parameters: []ParamDoc{
				{Name: "series", Type: "block! string!", Description: "The series to get the last element from", Optional: false},
			},
			Returns:  "[any-type!] The last element of the series",
			Examples: []string{"last [1 2 3]  ; => 3", `last "hello"  ; => "o"`, "last [[a b] c]  ; => c"},
			SeeAlso:  []string{"first", "length?", "append", "insert"}, Tags: []string{"series", "access", "last"},
		},
	}
	Registry["append"] = &NativeInfo{
		Func:      Append,
		NeedsEval: false,
		Arity:     2,
		Doc: &NativeDoc{
			Category: "Series",
			Summary:  "Appends a value to the end of a series",
			Description: `Adds a value to the end of a block or string, modifying the series in place.
Returns the modified series. For strings, the value is converted to a string before appending.`,
			Parameters: []ParamDoc{
				{Name: "series", Type: "block! string!", Description: "The series to append to (modified in place)", Optional: false},
				{Name: "value", Type: "any-type!", Description: "The value to append", Optional: false},
			},
			Returns:  "[block! string!] The modified series",
			Examples: []string{"data: [1 2 3]\nappend data 4  ; => [1 2 3 4]", `text: "hello"\nappend text " world"  ; => "hello world"`},
			SeeAlso:  []string{"insert", "first", "last", "length?"}, Tags: []string{"series", "modification", "append"},
		},
	}
	Registry["insert"] = &NativeInfo{
		Func:      Insert,
		NeedsEval: false,
		Arity:     2,
		Doc: &NativeDoc{
			Category: "Series",
			Summary:  "Inserts a value at the beginning of a series",
			Description: `Adds a value to the start of a block or string, modifying the series in place.
Returns the modified series. For strings, the value is converted to a string before inserting.`,
			Parameters: []ParamDoc{
				{Name: "series", Type: "block! string!", Description: "The series to insert into (modified in place)", Optional: false},
				{Name: "value", Type: "any-type!", Description: "The value to insert", Optional: false},
			},
			Returns:  "[block! string!] The modified series",
			Examples: []string{"data: [2 3 4]\ninsert data 1  ; => [1 2 3 4]", `text: "world"\ninsert text "hello "  ; => "hello world"`},
			SeeAlso:  []string{"append", "first", "last", "length?"}, Tags: []string{"series", "modification", "insert"},
		},
	}
	Registry["length?"] = &NativeInfo{
		Func:      LengthQ,
		NeedsEval: false,
		Arity:     1,
		Doc: &NativeDoc{
			Category: "Series",
			Summary:  "Returns the number of elements in a series",
			Description: `Counts the elements in a block or characters in a string.
Returns an integer representing the length of the series.`,
			Parameters: []ParamDoc{
				{Name: "series", Type: "block! string!", Description: "The series to measure", Optional: false},
			},
			Returns:  "[integer!] The number of elements in the series",
			Examples: []string{"length? [1 2 3 4]  ; => 4", `length? "hello"  ; => 5`, "length? []  ; => 0"},
			SeeAlso:  []string{"first", "last", "append", "insert"}, Tags: []string{"series", "query", "length", "count"},
		},
	}

	// Register data natives
	Registry["set"] = &NativeInfo{
		FuncEval:  Set,
		NeedsEval: true,
		Arity:     2,
		Doc: &NativeDoc{
			Category: "Data",
			Summary:  "Sets a word to a value in the current context",
			Description: `Assigns a value to a word (variable) in the current frame.
The word is not evaluated; the value is evaluated before assignment. Returns the assigned value.`,
			Parameters: []ParamDoc{
				{Name: "word", Type: "word!", Description: "The word to set (not evaluated)", Optional: false},
				{Name: "value", Type: "any-type!", Description: "The value to assign (evaluated)", Optional: false},
			},
			Returns:  "[any-type!] The value that was assigned",
			Examples: []string{"set 'x 42  ; => 42 (x is now 42)", "set 'name \"Alice\"  ; => \"Alice\"", "set 'data [1 2 3]  ; => [1 2 3]"},
			SeeAlso:  []string{"get", ":", "type?"}, Tags: []string{"data", "assignment", "variable"},
		},
	}
	Registry["get"] = &NativeInfo{
		FuncEval:  Get,
		NeedsEval: true,
		Arity:     1,
		Doc: &NativeDoc{
			Category: "Data",
			Summary:  "Gets the value of a word from the current context",
			Description: `Retrieves the value associated with a word (variable) in the current frame.
The word is not evaluated. Raises an error if the word is not bound to a value.`,
			Parameters: []ParamDoc{
				{Name: "word", Type: "word!", Description: "The word to look up (not evaluated)", Optional: false},
			},
			Returns:  "[any-type!] The value bound to the word",
			Examples: []string{"x: 42\nget 'x  ; => 42", "name: \"Bob\"\nget 'name  ; => \"Bob\""},
			SeeAlso:  []string{"set", ":", "type?"}, Tags: []string{"data", "access", "variable"},
		},
	}
	Registry["type?"] = &NativeInfo{
		Func:      TypeQ,
		NeedsEval: false,
		Arity:     1,
		Doc: &NativeDoc{
			Category: "Data",
			Summary:  "Returns the type of a value",
			Description: `Returns a word representing the type of the given value.
Possible types include: integer!, decimal!, string!, block!, word!, function!, object!, port!, logic!, none!`,
			Parameters: []ParamDoc{
				{Name: "value", Type: "any-type!", Description: "The value to check the type of", Optional: false},
			},
			Returns:  "[word!] A word representing the value's type",
			Examples: []string{"type? 42  ; => integer!", `type? "hello"  ; => string!`, "type? [1 2 3]  ; => block!", "type? :print  ; => function!"},
			SeeAlso:  []string{"set", "get"}, Tags: []string{"data", "type", "introspection", "reflection"},
		},
	}

	// Register object natives (Feature 002 - User Story 3)
	Registry["object"] = &NativeInfo{
		FuncEval:  Object,
		NeedsEval: true,
		Arity:     1,
		Doc: &NativeDoc{
			Category: "Objects",
			Summary:  "Creates a new object from a block of definitions",
			Description: `Creates a new object (context) by evaluating a block of word-value pairs.
The block is evaluated in a new frame, and all word definitions become fields of the object.
Returns the newly created object.`,
			Parameters: []ParamDoc{
				{Name: "spec", Type: "block!", Description: "A block containing word definitions to become object fields", Optional: false},
			},
			Returns:  "[object!] The newly created object",
			Examples: []string{"obj: object [x: 10 y: 20]  ; => object with fields x and y", "person: object [name: \"Alice\" age: 30]"},
			SeeAlso:  []string{"context", "make"}, Tags: []string{"objects", "context", "creation"},
		},
	}
	Registry["context"] = &NativeInfo{
		FuncEval:  Context,
		NeedsEval: true,
		Arity:     1,
		Doc: &NativeDoc{
			Category: "Objects",
			Summary:  "Creates a new context (alias for object)",
			Description: `Creates a new object (context) by evaluating a block of word-value pairs.
This is an alias for the 'object' function. The block is evaluated in a new frame,
and all word definitions become fields of the context.`,
			Parameters: []ParamDoc{
				{Name: "spec", Type: "block!", Description: "A block containing word definitions to become context fields", Optional: false},
			},
			Returns:  "[object!] The newly created context",
			Examples: []string{"ctx: context [counter: 0 increment: fn [] [counter: counter + 1]]", "config: context [debug: true port: 8080]"},
			SeeAlso:  []string{"object", "make"}, Tags: []string{"objects", "context", "creation"},
		},
	}
	Registry["make"] = &NativeInfo{
		FuncEval:  Make,
		NeedsEval: true,
		Arity:     2,
		Doc: &NativeDoc{
			Category: "Objects",
			Summary:  "Creates a derived object from a parent object",
			Description: `Creates a new object that inherits from a parent object and adds or overrides fields.
The first argument is the parent object (prototype), and the second is a block of
new or overriding field definitions. The new object shares the parent's fields but can shadow them.`,
			Parameters: []ParamDoc{
				{Name: "parent", Type: "object!", Description: "The parent object to derive from", Optional: false},
				{Name: "spec", Type: "block!", Description: "A block of field definitions to add or override", Optional: false},
			},
			Returns:  "[object!] The newly created derived object",
			Examples: []string{"base: object [x: 1 y: 2]\nderived: make base [z: 3]  ; => object with x, y, z", "point: object [x: 0 y: 0]\npoint3d: make point [z: 0]"},
			SeeAlso:  []string{"object", "context"}, Tags: []string{"objects", "inheritance", "derivation"},
		},
	}
	Registry["select"] = &NativeInfo{
		FuncEval:  Select,
		NeedsEval: true,
		Arity:     2,
		Doc: &NativeDoc{
			Category: "Objects",
			Summary:  "Retrieves a field value from an object or block",
			Description: `Looks up a field in an object or searches for a key in a block.
For objects: returns the field value, checking parent prototypes if needed.
For blocks: searches for key-value pairs (alternating pattern) and returns the value.
Use --default refinement to provide a fallback when field/key is not found.`,
			Parameters: []ParamDoc{
				{Name: "target", Type: "object! or block!", Description: "The object or block to search", Optional: false},
				{Name: "field", Type: "word! or string!", Description: "The field name or key to look up", Optional: false},
				{Name: "--default", Type: "any-type!", Description: "Optional default value when field not found", Optional: true},
			},
			Returns: "[any-type!] The field/key value, or default, or none",
			Examples: []string{
				"obj: object [x: 10 y: 20]\nselect obj 'x  ; => 10",
				"select obj 'z --default 99  ; => 99 (field not found)",
				"data: ['name \"Alice\" 'age 30]\nselect data 'age  ; => 30",
			},
			SeeAlso: []string{"put", "get", "object"},
			Tags:    []string{"objects", "lookup", "field-access"},
		},
	}
	Registry["put"] = &NativeInfo{
		FuncEval:  Put,
		NeedsEval: true,
		Arity:     3,
		Doc: &NativeDoc{
			Category: "Objects",
			Summary:  "Sets a field value in an object",
			Description: `Updates an existing field in an object with a new value.
The field must already exist in the object's manifest - dynamic field addition is not allowed.
If the field has a type hint, the new value must match that type.
Returns the assigned value.`,
			Parameters: []ParamDoc{
				{Name: "object", Type: "object!", Description: "The object to modify", Optional: false},
				{Name: "field", Type: "word! or string!", Description: "The field name to update", Optional: false},
				{Name: "value", Type: "any-type!", Description: "The new value to assign", Optional: false},
			},
			Returns: "[any-type!] The assigned value",
			Examples: []string{
				"obj: object [x: 10 y: 20]\nput obj 'x 42  ; => 42, obj.x is now 42",
				"person: object [name: \"Alice\" age: 30]\nput person 'age 31",
			},
			SeeAlso: []string{"select", "set", "object"},
			Tags:    []string{"objects", "mutation", "field-update"},
		},
	}

	// Register IO natives
	Registry["print"] = &NativeInfo{
		FuncEval:  Print,
		NeedsEval: true,
		Arity:     1,
		Doc: &NativeDoc{
			Category: "I/O",
			Summary:  "Prints a value to standard output",
			Description: `Evaluates and prints a value to standard output, followed by a newline.
Blocks are formatted with spaces between elements. Returns none.`,
			Parameters: []ParamDoc{
				{Name: "value", Type: "any-type!", Description: "The value to print (will be evaluated)", Optional: false},
			},
			Returns:  "[none!] Always returns none",
			Examples: []string{`print "Hello, world!"  ; prints: Hello, world!`, "print 42  ; prints: 42", "print [1 2 3]  ; prints: 1 2 3"},
			SeeAlso:  []string{"input"}, Tags: []string{"io", "output", "print", "display"},
		},
	}
	Registry["input"] = &NativeInfo{
		Func:      Input,
		NeedsEval: false,
		Arity:     0,
		Doc: &NativeDoc{
			Category: "I/O",
			Summary:  "Reads a line of text from standard input",
			Description: `Reads a line of text from standard input (stdin) and returns it as a string.
The trailing newline is removed. Blocks until input is received.`,
			Parameters: []ParamDoc{},
			Returns:    "[string!] The line of text read from standard input",
			Examples:   []string{`name: input  ; waits for user input`, `print "Enter your name:"\nname: input\nprint ["Hello" name]`},
			SeeAlso:    []string{"print", "read"}, Tags: []string{"io", "input", "stdin", "read"},
		},
	}

	// Register port natives (Feature 002 - User Story 2)
	Registry["open"] = &NativeInfo{
		Func:      OpenNative,
		NeedsEval: false,
		Arity:     1,
		Doc: &NativeDoc{
			Category: "Ports",
			Summary:  "Opens a port for file or network I/O",
			Description: `Opens a port specified by a URL or file path string. Supports file:// URLs and
potentially other schemes. Returns a port value that can be used with read, write, close, etc.
File operations are subject to sandbox restrictions if configured.`,
			Parameters: []ParamDoc{
				{Name: "spec", Type: "string!", Description: "A URL or file path (e.g., \"file://data.txt\")", Optional: false},
			},
			Returns:  "[port!] An open port ready for I/O operations",
			Examples: []string{`p: open "file://data.txt"  ; => port`, `p: open "file:///tmp/output.log"`},
			SeeAlso:  []string{"close", "read", "write", "save", "load"}, Tags: []string{"ports", "io", "file", "open"},
		},
	}
	Registry["close"] = &NativeInfo{
		Func:      CloseNative,
		NeedsEval: false,
		Arity:     1,
		Doc: &NativeDoc{
			Category: "Ports",
			Summary:  "Closes an open port",
			Description: `Closes a previously opened port, releasing any associated resources.
After closing, the port should not be used for further I/O operations. Returns none.`,
			Parameters: []ParamDoc{
				{Name: "port", Type: "port!", Description: "The port to close", Optional: false},
			},
			Returns:  "[none!] Always returns none",
			Examples: []string{`p: open "file://data.txt"\nclose p  ; closes the port`},
			SeeAlso:  []string{"open", "read", "write"}, Tags: []string{"ports", "io", "close", "cleanup"},
		},
	}
	Registry["read"] = &NativeInfo{
		Func:      ReadNative,
		NeedsEval: false,
		Arity:     1,
		Doc: &NativeDoc{
			Category: "Ports",
			Summary:  "Reads data from a port or file",
			Description: `Reads all data from a port or directly from a file path.
If given a port, reads from that open port. If given a string (file path),
opens the file, reads its contents, and closes it automatically. Returns the data as a string.`,
			Parameters: []ParamDoc{
				{Name: "source", Type: "port! string!", Description: "A port or file path to read from", Optional: false},
			},
			Returns:  "[string!] The data read from the source",
			Examples: []string{`content: read "file://data.txt"  ; read entire file`, `p: open "file://data.txt"\ndata: read p\nclose p`},
			SeeAlso:  []string{"write", "load", "open", "close"}, Tags: []string{"ports", "io", "read", "file"},
		},
	}
	Registry["write"] = &NativeInfo{
		Func:      WriteNative,
		NeedsEval: false,
		Arity:     2,
		Doc: &NativeDoc{
			Category: "Ports",
			Summary:  "Writes data to a port or file",
			Description: `Writes data to a port or directly to a file path.
If the target is a port, writes to that open port. If given a string (file path),
opens the file, writes the data, and closes it automatically. Overwrites existing content.`,
			Parameters: []ParamDoc{
				{Name: "target", Type: "port! string!", Description: "A port or file path to write to", Optional: false},
				{Name: "data", Type: "string!", Description: "The data to write", Optional: false},
			},
			Returns:  "[none!] Always returns none",
			Examples: []string{`write "file://output.txt" "Hello, world!"  ; write to file`, `p: open "file://output.txt"\nwrite p "data"\nclose p`},
			SeeAlso:  []string{"read", "save", "open", "close"}, Tags: []string{"ports", "io", "write", "file"},
		},
	}
	Registry["save"] = &NativeInfo{
		Func:      SaveNative,
		NeedsEval: false,
		Arity:     2,
		Doc: &NativeDoc{
			Category: "Ports",
			Summary:  "Saves a value to a file in viro format",
			Description: `Serializes a viro value (block, object, etc.) and writes it to a file.
The value is converted to viro source code format that can be loaded back with 'load'.
This is the recommended way to persist viro data structures.`,
			Parameters: []ParamDoc{
				{Name: "file", Type: "string!", Description: "The file path to save to", Optional: false},
				{Name: "value", Type: "any-type!", Description: "The value to save", Optional: false},
			},
			Returns:  "[none!] Always returns none",
			Examples: []string{`save "file://config.viro" [debug: true port: 8080]`, `data: object [x: 1 y: 2]\nsave "file://data.viro" data`},
			SeeAlso:  []string{"load", "write", "read"}, Tags: []string{"ports", "io", "save", "serialize", "persist"},
		},
	}
	Registry["load"] = &NativeInfo{
		Func:      LoadNative,
		NeedsEval: false,
		Arity:     1,
		Doc: &NativeDoc{
			Category: "Ports",
			Summary:  "Loads and parses a viro source file",
			Description: `Reads a file containing viro source code, parses it, and returns the parsed value.
This is the recommended way to load data structures saved with 'save'.
Returns the parsed viro value (block, object, etc.).`,
			Parameters: []ParamDoc{
				{Name: "file", Type: "string!", Description: "The file path to load from", Optional: false},
			},
			Returns:  "[any-type!] The parsed viro value from the file",
			Examples: []string{`config: load "file://config.viro"  ; load and parse`, `data: load "file://data.viro"`},
			SeeAlso:  []string{"save", "read"}, Tags: []string{"ports", "io", "load", "parse", "deserialize"},
		},
	}
	Registry["query"] = &NativeInfo{
		Func:      QueryNative,
		NeedsEval: false,
		Arity:     1,
		Doc: &NativeDoc{
			Category: "Ports",
			Summary:  "Queries metadata about a port or file",
			Description: `Returns metadata about a port or file, such as size, modification time, or status.
The exact information returned depends on the port type. Returns an object with metadata fields.`,
			Parameters: []ParamDoc{
				{Name: "target", Type: "port! string!", Description: "A port or file path to query", Optional: false},
			},
			Returns:  "[object!] An object containing metadata about the target",
			Examples: []string{`info: query "file://data.txt"  ; get file info`, `p: open "file://data.txt"\ninfo: query p\nclose p`},
			SeeAlso:  []string{"open", "read"}, Tags: []string{"ports", "io", "metadata", "query", "info"},
		},
	}
	Registry["wait"] = &NativeInfo{
		Func:      WaitNative,
		NeedsEval: false,
		Arity:     1,
		Doc: &NativeDoc{
			Category: "Ports",
			Summary:  "Waits for a port to be ready or for a timeout",
			Description: `Waits for a port to become ready for I/O operations, or for a specified duration.
If given a number, waits for that many seconds. If given a port, waits until the port is ready.
Returns the port that became ready, or none if a timeout occurred.`,
			Parameters: []ParamDoc{
				{Name: "target", Type: "port! integer! decimal!", Description: "A port to wait on or a duration in seconds", Optional: false},
			},
			Returns:  "[port! none!] The ready port or none on timeout",
			Examples: []string{"wait 2  ; wait for 2 seconds", "wait 0.5  ; wait for half a second", `p: open "file://data.txt"\nwait p  ; wait until port is ready`},
			SeeAlso:  []string{"open", "read", "write"}, Tags: []string{"ports", "io", "wait", "delay", "timeout"},
		},
	}

	// Register function native
	Registry["fn"] = &NativeInfo{
		FuncEval:  Fn,
		NeedsEval: true,
		Arity:     2,
		EvalArgs:  []bool{false, false},
		Doc: &NativeDoc{
			Category: "Function",
			Summary:  "Creates a new function",
			Description: `Defines a new function with parameters and a body. The first argument is a block
containing parameter names, and the second is a block containing the function body code.
Returns a function value that can be called. Functions capture their defining context (closure).`,
			Parameters: []ParamDoc{
				{Name: "params", Type: "block!", Description: "A block of parameter names (words)", Optional: false},
				{Name: "body", Type: "block!", Description: "A block of code to execute when the function is called", Optional: false},
			},
			Returns:  "[function!] The newly created function",
			Examples: []string{"square: fn [n] [n * n]  ; => function", "add: fn [a b] [a + b]\nadd 3 4  ; => 7", "greet: fn [name] [print [\"Hello\" name]]\ngreet \"Alice\"  ; prints: Hello Alice"},
			SeeAlso:  []string{"set", "get"}, Tags: []string{"function", "definition", "lambda", "closure"},
		},
	}

	// Register control flow natives (need evaluator)
	Registry["when"] = &NativeInfo{
		FuncEval:  When,
		NeedsEval: true,
		Arity:     2,
		EvalArgs:  []bool{true, false},
		Doc: &NativeDoc{
			Category: "Control",
			Summary:  "Executes a block of code if a condition is true",
			Description: `Evaluates the condition, and if it's true (non-zero, non-empty), evaluates and returns
the result of the body block. If the condition is false, returns none. This is a one-branch conditional.`,
			Parameters: []ParamDoc{
				{Name: "condition", Type: "logic! integer!", Description: "The condition to test (evaluated)", Optional: false},
				{Name: "body", Type: "block!", Description: "The code to execute if condition is true (not evaluated unless condition is true)", Optional: false},
			},
			Returns:  "[any-type! none!] The result of the body if condition is true, otherwise none",
			Examples: []string{"x: 10\nwhen x > 5 [print \"x is large\"]  ; prints: x is large", "when false [print \"not printed\"]  ; => none"},
			SeeAlso:  []string{"if", "loop", "while"}, Tags: []string{"control", "conditional", "when"},
		},
	}
	Registry["if"] = &NativeInfo{
		FuncEval:  If,
		NeedsEval: true,
		Arity:     3,
		EvalArgs:  []bool{true, false, false},
		Doc: &NativeDoc{
			Category: "Control",
			Summary:  "Executes one of two blocks based on a condition",
			Description: `Evaluates the condition, and if it's true (non-zero, non-empty), evaluates and returns
the result of the true-branch. If the condition is false, evaluates and returns the result of the false-branch.
This is a two-branch conditional (if-then-else).`,
			Parameters: []ParamDoc{
				{Name: "condition", Type: "logic! integer!", Description: "The condition to test (evaluated)", Optional: false},
				{Name: "true-branch", Type: "block!", Description: "The code to execute if condition is true", Optional: false},
				{Name: "false-branch", Type: "block!", Description: "The code to execute if condition is false", Optional: false},
			},
			Returns:  "[any-type!] The result of whichever branch was executed",
			Examples: []string{"x: 10\nif x > 5 [\"large\"] [\"small\"]  ; => \"large\"", "if false [1] [2]  ; => 2", "result: if 3 = 3 [print \"equal\"] [print \"not equal\"]"},
			SeeAlso:  []string{"when", "loop", "while"}, Tags: []string{"control", "conditional", "if", "else"},
		},
	}
	Registry["loop"] = &NativeInfo{
		FuncEval:  Loop,
		NeedsEval: true,
		Arity:     2,
		EvalArgs:  []bool{true, false},
		Doc: &NativeDoc{
			Category: "Control",
			Summary:  "Executes a block a specified number of times",
			Description: `Evaluates the body block repeatedly for the specified number of iterations.
The count must be a non-negative integer. Returns the result of the last iteration, or none if count is 0.`,
			Parameters: []ParamDoc{
				{Name: "count", Type: "integer!", Description: "The number of times to execute the body (evaluated)", Optional: false},
				{Name: "body", Type: "block!", Description: "The code to execute repeatedly", Optional: false},
			},
			Returns:  "[any-type! none!] The result of the last iteration",
			Examples: []string{"loop 3 [print \"hello\"]  ; prints 'hello' 3 times", "x: 0\nloop 5 [x: x + 1]  ; x becomes 5", "loop 0 [print \"never\"]  ; => none"},
			SeeAlso:  []string{"while", "if", "when"}, Tags: []string{"control", "loop", "iteration", "repeat"},
		},
	}
	Registry["while"] = &NativeInfo{
		FuncEval:  While,
		NeedsEval: true,
		Arity:     2,
		EvalArgs:  []bool{true, false},
		Doc: &NativeDoc{
			Category: "Control",
			Summary:  "Executes a block repeatedly while a condition is true",
			Description: `Evaluates the condition, and if it's true (non-zero, non-empty), evaluates the body block.
Repeats this process until the condition becomes false. Returns the result of the last iteration,
or none if the condition is initially false. Be careful to avoid infinite loops.`,
			Parameters: []ParamDoc{
				{Name: "condition", Type: "block! logic! integer!", Description: "The condition to test (evaluated before each iteration)", Optional: false},
				{Name: "body", Type: "block!", Description: "The code to execute while condition is true", Optional: false},
			},
			Returns:  "[any-type! none!] The result of the last iteration",
			Examples: []string{"x: 0\nwhile [x < 5] [x: x + 1]  ; x becomes 5", "count: 10\nwhile [count > 0] [print count count: count - 1]", "while [false] [print \"never\"]  ; => none"},
			SeeAlso:  []string{"loop", "if", "when"}, Tags: []string{"control", "loop", "while", "iteration"},
		},
	}

	// Register help system natives
	Registry["?"] = &NativeInfo{
		Func:      Help,
		NeedsEval: false,         // Do not evaluate args - we want word literals
		Arity:     1,             // Requires one argument in scripts
		EvalArgs:  []bool{false}, // Do not evaluate the first arg
		Doc: &NativeDoc{
			Category: "Help",
			Summary:  "Displays help for functions or lists functions in a category",
			Description: `Interactive help system for discovering and learning about viro functions.
Provide a word argument to show detailed documentation for that function or list functions in that category.
Provides usage examples, parameter descriptions, and cross-references.

Note: In the REPL, typing just '?' (without arguments) is a special shortcut that shows all categories.
In scripts, you must provide an argument: '? math' or '? append'.`,
			Parameters: []ParamDoc{
				{Name: "topic", Type: "word! string!", Description: "Function name or category to get help for", Optional: true},
			},
			Returns:  "[none!] Always returns none (displays help to stdout)",
			Examples: []string{"? math  ; list functions in Math category", "? append  ; show detailed help for append", "? \"sqrt\"  ; help using string"},
			SeeAlso:  []string{"words", "type?"},
			Tags:     []string{"help", "documentation", "discovery", "introspection"},
		},
	}

	Registry["words"] = &NativeInfo{
		Func:      Words,
		NeedsEval: false,
		Arity:     0,
		Doc: &NativeDoc{
			Category: "Help",
			Summary:  "Lists all available native function names",
			Description: `Returns a block containing all native function names as words.
Does not print by default - use 'print words' to display the list.
Useful for programmatic access to available functionality.`,
			Parameters: []ParamDoc{},
			Returns:    "[block!] A block containing all function names as words",
			Examples:   []string{"words  ; return all function names", "fns: words\nlength? fns  ; count available functions", "print words  ; display function names"},
			SeeAlso:    []string{"?", "type?"}, Tags: []string{"help", "documentation", "discovery", "list"},
		},
	}

	// ===== NEW: Populate FunctionRegistry =====
	// Group 1: Simple math operations (4 functions)
	FunctionRegistry["+"] = value.NewNativeFunction(
		"+",
		[]value.ParamSpec{
			value.NewParamSpec("left", true),
			value.NewParamSpec("right", true),
		},
		func(args []value.Value, eval value.Evaluator) (value.Value, error) {
			result, err := Add(args)
			if err == nil {
				return result, nil
			}
			return result, err
		},
	)
	fn := FunctionRegistry["+"]
	fn.Infix = true
	fn.Doc = Registry["+"].Doc

	FunctionRegistry["-"] = value.NewNativeFunction(
		"-",
		[]value.ParamSpec{
			value.NewParamSpec("left", true),
			value.NewParamSpec("right", true),
		},
		func(args []value.Value, eval value.Evaluator) (value.Value, error) {
			result, err := Subtract(args)
			if err == nil {
				return result, nil
			}
			return result, err
		},
	)
	fn = FunctionRegistry["-"]
	fn.Infix = true
	fn.Doc = Registry["-"].Doc

	FunctionRegistry["*"] = value.NewNativeFunction(
		"*",
		[]value.ParamSpec{
			value.NewParamSpec("left", true),
			value.NewParamSpec("right", true),
		},
		func(args []value.Value, eval value.Evaluator) (value.Value, error) {
			result, err := Multiply(args)
			if err == nil {
				return result, nil
			}
			return result, err
		},
	)
	fn = FunctionRegistry["*"]
	fn.Infix = true
	fn.Doc = Registry["*"].Doc

	FunctionRegistry["/"] = value.NewNativeFunction(
		"/",
		[]value.ParamSpec{
			value.NewParamSpec("left", true),
			value.NewParamSpec("right", true),
		},
		func(args []value.Value, eval value.Evaluator) (value.Value, error) {
			result, err := Divide(args)
			if err == nil {
				return result, nil
			}
			return result, err
		},
	)
	fn = FunctionRegistry["/"]
	fn.Infix = true
	fn.Doc = Registry["/"].Doc

	// Group 2: Comparison operators (6 functions)
	FunctionRegistry["<"] = value.NewNativeFunction(
		"<",
		[]value.ParamSpec{
			value.NewParamSpec("left", true),
			value.NewParamSpec("right", true),
		},
		func(args []value.Value, eval value.Evaluator) (value.Value, error) {
			result, err := LessThan(args)
			if err == nil {
				return result, nil
			}
			return result, err
		},
	)
	fn = FunctionRegistry["<"]
	fn.Infix = true
	fn.Doc = Registry["<"].Doc

	FunctionRegistry[">"] = value.NewNativeFunction(
		">",
		[]value.ParamSpec{
			value.NewParamSpec("left", true),
			value.NewParamSpec("right", true),
		},
		func(args []value.Value, eval value.Evaluator) (value.Value, error) {
			result, err := GreaterThan(args)
			if err == nil {
				return result, nil
			}
			return result, err
		},
	)
	fn = FunctionRegistry[">"]
	fn.Infix = true
	fn.Doc = Registry[">"].Doc

	FunctionRegistry["<="] = value.NewNativeFunction(
		"<=",
		[]value.ParamSpec{
			value.NewParamSpec("left", true),
			value.NewParamSpec("right", true),
		},
		func(args []value.Value, eval value.Evaluator) (value.Value, error) {
			result, err := LessOrEqual(args)
			if err == nil {
				return result, nil
			}
			return result, err
		},
	)
	fn = FunctionRegistry["<="]
	fn.Infix = true
	fn.Doc = Registry["<="].Doc

	FunctionRegistry[">="] = value.NewNativeFunction(
		">=",
		[]value.ParamSpec{
			value.NewParamSpec("left", true),
			value.NewParamSpec("right", true),
		},
		func(args []value.Value, eval value.Evaluator) (value.Value, error) {
			result, err := GreaterOrEqual(args)
			if err == nil {
				return result, nil
			}
			return result, err
		},
	)
	fn = FunctionRegistry[">="]
	fn.Infix = true
	fn.Doc = Registry[">="].Doc

	FunctionRegistry["="] = value.NewNativeFunction(
		"=",
		[]value.ParamSpec{
			value.NewParamSpec("left", true),
			value.NewParamSpec("right", true),
		},
		func(args []value.Value, eval value.Evaluator) (value.Value, error) {
			result, err := Equal(args)
			if err == nil {
				return result, nil
			}
			return result, err
		},
	)
	fn = FunctionRegistry["="]
	fn.Infix = true
	fn.Doc = Registry["="].Doc

	FunctionRegistry["<>"] = value.NewNativeFunction(
		"<>",
		[]value.ParamSpec{
			value.NewParamSpec("left", true),
			value.NewParamSpec("right", true),
		},
		func(args []value.Value, eval value.Evaluator) (value.Value, error) {
			result, err := NotEqual(args)
			if err == nil {
				return result, nil
			}
			return result, err
		},
	)
	fn = FunctionRegistry["<>"]
	fn.Infix = true
	fn.Doc = Registry["<>"].Doc

	// Group 3: Logic operators (3 functions)
	FunctionRegistry["and"] = value.NewNativeFunction(
		"and",
		[]value.ParamSpec{
			value.NewParamSpec("left", true),
			value.NewParamSpec("right", true),
		},
		func(args []value.Value, eval value.Evaluator) (value.Value, error) {
			result, err := And(args)
			if err == nil {
				return result, nil
			}
			return result, err
		},
	)
	fn = FunctionRegistry["and"]
	fn.Infix = true
	fn.Doc = Registry["and"].Doc

	FunctionRegistry["or"] = value.NewNativeFunction(
		"or",
		[]value.ParamSpec{
			value.NewParamSpec("left", true),
			value.NewParamSpec("right", true),
		},
		func(args []value.Value, eval value.Evaluator) (value.Value, error) {
			result, err := Or(args)
			if err == nil {
				return result, nil
			}
			return result, err
		},
	)
	fn = FunctionRegistry["or"]
	fn.Infix = true
	fn.Doc = Registry["or"].Doc

	FunctionRegistry["not"] = value.NewNativeFunction(
		"not",
		[]value.ParamSpec{
			value.NewParamSpec("value", true),
		},
		func(args []value.Value, eval value.Evaluator) (value.Value, error) {
			result, err := Not(args)
			if err == nil {
				return result, nil
			}
			return result, err
		},
	)
	fn = FunctionRegistry["not"]
	fn.Infix = false
	fn.Doc = Registry["not"].Doc

	// Group 4: Advanced math functions (16 functions)
	// Helper function to wrap simple math functions
	registerSimpleMathFunc := func(name string, impl func([]value.Value) (value.Value, *verror.Error), arity int) {
		// Extract parameter names from existing documentation
		params := make([]value.ParamSpec, arity)
		oldInfo := Registry[name]

		if oldInfo.Doc != nil && len(oldInfo.Doc.Parameters) == arity {
			// Use parameter names from documentation
			for i := 0; i < arity; i++ {
				params[i] = value.NewParamSpec(oldInfo.Doc.Parameters[i].Name, true)
			}
		} else {
			// Fallback to generic names if documentation is missing or mismatched
			paramNames := []string{"value", "left", "right", "base", "exponent"}
			for i := 0; i < arity; i++ {
				if i < len(paramNames) {
					params[i] = value.NewParamSpec(paramNames[i], true)
				} else {
					params[i] = value.NewParamSpec("arg", true)
				}
			}
		}

		FunctionRegistry[name] = value.NewNativeFunction(
			name,
			params,
			func(args []value.Value, eval value.Evaluator) (value.Value, error) {
				result, err := impl(args)
				if err == nil {
					return result, nil
				}
				return result, err
			},
		)
		fn := FunctionRegistry[name]
		fn.Doc = Registry[name].Doc
	}

	registerSimpleMathFunc("decimal", DecimalConstructor, 1)
	registerSimpleMathFunc("pow", Pow, 2)
	registerSimpleMathFunc("sqrt", Sqrt, 1)
	registerSimpleMathFunc("exp", Exp, 1)
	registerSimpleMathFunc("log", Log, 1)
	registerSimpleMathFunc("log-10", Log10, 1)
	registerSimpleMathFunc("sin", Sin, 1)
	registerSimpleMathFunc("cos", Cos, 1)
	registerSimpleMathFunc("tan", Tan, 1)
	registerSimpleMathFunc("asin", Asin, 1)
	registerSimpleMathFunc("acos", Acos, 1)
	registerSimpleMathFunc("atan", Atan, 1)
	registerSimpleMathFunc("round", Round, 1)
	registerSimpleMathFunc("ceil", Ceil, 1)
	registerSimpleMathFunc("floor", Floor, 1)
	registerSimpleMathFunc("truncate", Truncate, 1)

	// Group 5: Series operations (5 functions)
	registerSimpleMathFunc("first", First, 1)
	registerSimpleMathFunc("last", Last, 1)
	registerSimpleMathFunc("append", Append, 2)
	registerSimpleMathFunc("insert", Insert, 2)
	registerSimpleMathFunc("length?", LengthQ, 1)

	// Group 6: Data operations (3 functions)
	// set and get need evaluator, type? doesn't
	FunctionRegistry["set"] = value.NewNativeFunction(
		"set",
		[]value.ParamSpec{
			value.NewParamSpec("word", false), // NOT evaluated (lit-word)
			value.NewParamSpec("value", true), // evaluated
		},
		func(args []value.Value, eval value.Evaluator) (value.Value, error) {
			// We need to pass a native.Evaluator to Set, but we have value.Evaluator
			// Create a reverse adapter that converts value.Evaluator back to native.Evaluator
			reverseAdapter := &nativeEvaluatorAdapter{eval}
			result, err := Set(args, reverseAdapter.unwrap())
			if err == nil {
				return result, nil
			}
			return result, err
		},
	)
	fn = FunctionRegistry["set"]
	fn.Doc = Registry["set"].Doc

	FunctionRegistry["get"] = value.NewNativeFunction(
		"get",
		[]value.ParamSpec{
			value.NewParamSpec("word", false), // NOT evaluated (lit-word)
		},
		func(args []value.Value, eval value.Evaluator) (value.Value, error) {
			reverseAdapter := &nativeEvaluatorAdapter{eval}
			result, err := Get(args, reverseAdapter.unwrap())
			if err == nil {
				return result, nil
			}
			return result, err
		},
	)
	fn = FunctionRegistry["get"]
	fn.Doc = Registry["get"].Doc

	registerSimpleMathFunc("type?", TypeQ, 1)

	// Group 7: Object operations (5 functions - all need evaluator)
	FunctionRegistry["object"] = value.NewNativeFunction(
		"object",
		[]value.ParamSpec{
			value.NewParamSpec("spec", false), // NOT evaluated (block)
		},
		func(args []value.Value, eval value.Evaluator) (value.Value, error) {
			reverseAdapter := &nativeEvaluatorAdapter{eval}
			result, err := Object(args, reverseAdapter.unwrap())
			if err == nil {
				return result, nil
			}
			return result, err
		},
	)
	fn = FunctionRegistry["object"]
	fn.Doc = Registry["object"].Doc

	FunctionRegistry["context"] = value.NewNativeFunction(
		"context",
		[]value.ParamSpec{
			value.NewParamSpec("spec", false), // NOT evaluated (block)
		},
		func(args []value.Value, eval value.Evaluator) (value.Value, error) {
			reverseAdapter := &nativeEvaluatorAdapter{eval}
			result, err := Context(args, reverseAdapter.unwrap())
			if err == nil {
				return result, nil
			}
			return result, err
		},
	)
	fn = FunctionRegistry["context"]
	fn.Doc = Registry["context"].Doc

	FunctionRegistry["make"] = value.NewNativeFunction(
		"make",
		[]value.ParamSpec{
			value.NewParamSpec("parent", true), // evaluated
			value.NewParamSpec("spec", false),  // NOT evaluated (block)
		},
		func(args []value.Value, eval value.Evaluator) (value.Value, error) {
			reverseAdapter := &nativeEvaluatorAdapter{eval}
			result, err := Make(args, reverseAdapter.unwrap())
			if err == nil {
				return result, nil
			}
			return result, err
		},
	)
	fn = FunctionRegistry["make"]
	fn.Doc = Registry["make"].Doc

	FunctionRegistry["select"] = value.NewNativeFunction(
		"select",
		[]value.ParamSpec{
			value.NewParamSpec("target", true), // evaluated
			value.NewParamSpec("field", false), // NOT evaluated (word/string)
		},
		func(args []value.Value, eval value.Evaluator) (value.Value, error) {
			reverseAdapter := &nativeEvaluatorAdapter{eval}
			result, err := Select(args, reverseAdapter.unwrap())
			if err == nil {
				return result, nil
			}
			return result, err
		},
	)
	fn = FunctionRegistry["select"]
	fn.Doc = Registry["select"].Doc

	FunctionRegistry["put"] = value.NewNativeFunction(
		"put",
		[]value.ParamSpec{
			value.NewParamSpec("object", true), // evaluated
			value.NewParamSpec("field", false), // NOT evaluated (word/string)
			value.NewParamSpec("value", true),  // evaluated
		},
		func(args []value.Value, eval value.Evaluator) (value.Value, error) {
			reverseAdapter := &nativeEvaluatorAdapter{eval}
			result, err := Put(args, reverseAdapter.unwrap())
			if err == nil {
				return result, nil
			}
			return result, err
		},
	)
	fn = FunctionRegistry["put"]
	fn.Doc = Registry["put"].Doc

	// Group 8: I/O operations (2 functions - print needs evaluator)
	FunctionRegistry["print"] = value.NewNativeFunction(
		"print",
		[]value.ParamSpec{
			value.NewParamSpec("value", true), // evaluated
		},
		func(args []value.Value, eval value.Evaluator) (value.Value, error) {
			reverseAdapter := &nativeEvaluatorAdapter{eval}
			result, err := Print(args, reverseAdapter.unwrap())
			if err == nil {
				return result, nil
			}
			return result, err
		},
	)
	fn = FunctionRegistry["print"]
	fn.Doc = Registry["print"].Doc

	registerSimpleMathFunc("input", Input, 0)

	// Group 9: Port operations (8 functions)
	registerSimpleMathFunc("open", OpenNative, 1)
	registerSimpleMathFunc("close", CloseNative, 1)
	registerSimpleMathFunc("read", ReadNative, 1)
	registerSimpleMathFunc("write", WriteNative, 2)
	registerSimpleMathFunc("save", SaveNative, 2)
	registerSimpleMathFunc("load", LoadNative, 1)
	registerSimpleMathFunc("query", QueryNative, 1)
	registerSimpleMathFunc("wait", WaitNative, 1)

	// Group 10: Control flow (4 functions - all need evaluator)
	FunctionRegistry["when"] = value.NewNativeFunction(
		"when",
		[]value.ParamSpec{
			value.NewParamSpec("condition", true), // evaluated
			value.NewParamSpec("body", false),     // NOT evaluated (block)
		},
		func(args []value.Value, eval value.Evaluator) (value.Value, error) {
			reverseAdapter := &nativeEvaluatorAdapter{eval}
			result, err := When(args, reverseAdapter.unwrap())
			if err == nil {
				return result, nil
			}
			return result, err
		},
	)
	fn = FunctionRegistry["when"]
	fn.Doc = Registry["when"].Doc

	FunctionRegistry["if"] = value.NewNativeFunction(
		"if",
		[]value.ParamSpec{
			value.NewParamSpec("condition", true),     // evaluated
			value.NewParamSpec("true-branch", false),  // NOT evaluated (block)
			value.NewParamSpec("false-branch", false), // NOT evaluated (block)
		},
		func(args []value.Value, eval value.Evaluator) (value.Value, error) {
			reverseAdapter := &nativeEvaluatorAdapter{eval}
			result, err := If(args, reverseAdapter.unwrap())
			if err == nil {
				return result, nil
			}
			return result, err
		},
	)
	fn = FunctionRegistry["if"]
	fn.Doc = Registry["if"].Doc

	FunctionRegistry["loop"] = value.NewNativeFunction(
		"loop",
		[]value.ParamSpec{
			value.NewParamSpec("count", true), // evaluated
			value.NewParamSpec("body", false), // NOT evaluated (block)
		},
		func(args []value.Value, eval value.Evaluator) (value.Value, error) {
			reverseAdapter := &nativeEvaluatorAdapter{eval}
			result, err := Loop(args, reverseAdapter.unwrap())
			if err == nil {
				return result, nil
			}
			return result, err
		},
	)
	fn = FunctionRegistry["loop"]
	fn.Doc = Registry["loop"].Doc

	FunctionRegistry["while"] = value.NewNativeFunction(
		"while",
		[]value.ParamSpec{
			value.NewParamSpec("condition", true), // evaluated
			value.NewParamSpec("body", false),     // NOT evaluated (block)
		},
		func(args []value.Value, eval value.Evaluator) (value.Value, error) {
			reverseAdapter := &nativeEvaluatorAdapter{eval}
			result, err := While(args, reverseAdapter.unwrap())
			if err == nil {
				return result, nil
			}
			return result, err
		},
	)
	fn = FunctionRegistry["while"]
	fn.Doc = Registry["while"].Doc

	// Group 11: Function creation (1 function - needs evaluator)
	FunctionRegistry["fn"] = value.NewNativeFunction(
		"fn",
		[]value.ParamSpec{
			value.NewParamSpec("params", false), // NOT evaluated (block)
			value.NewParamSpec("body", false),   // NOT evaluated (block)
		},
		func(args []value.Value, eval value.Evaluator) (value.Value, error) {
			reverseAdapter := &nativeEvaluatorAdapter{eval}
			result, err := Fn(args, reverseAdapter.unwrap())
			if err == nil {
				return result, nil
			}
			return result, err
		},
	)
	fn = FunctionRegistry["fn"]
	fn.Doc = Registry["fn"].Doc

	// Group 12: Help system (2 functions)
	FunctionRegistry["?"] = value.NewNativeFunction(
		"?",
		[]value.ParamSpec{
			value.NewParamSpec("topic", false), // NOT evaluated (word/string)
		},
		func(args []value.Value, eval value.Evaluator) (value.Value, error) {
			result, err := Help(args)
			if err == nil {
				return result, nil
			}
			return result, err
		},
	)
	fn = FunctionRegistry["?"]
	fn.Doc = Registry["?"].Doc

	registerSimpleMathFunc("words", Words, 0)

	// Validate EvalArgs length matches Arity
	for name, info := range Registry {
		if info.EvalArgs != nil && len(info.EvalArgs) != info.Arity {
			panic(fmt.Sprintf("Native %s: EvalArgs length (%d) != Arity (%d)",
				name, len(info.EvalArgs), info.Arity))
		}
	}
}

// Lookup finds a native function by name.
// Returns the function info and true if found, nil and false otherwise.
// DEPRECATED: Use LookupFunction for new code. This will be removed after migration.
func Lookup(name string) (*NativeInfo, bool) {
	info, ok := Registry[name]
	return info, ok
}

// Call invokes a native function with the given arguments and evaluator.
// Handles both simple natives and natives that need evaluator access.
// DEPRECATED: Use CallFunction for new code. This will be removed after migration.
func Call(info *NativeInfo, args []value.Value, eval Evaluator) (value.Value, *verror.Error) {
	if info.NeedsEval {
		return info.FuncEval(args, eval)
	}
	return info.Func(args)
}

// LookupFunction finds a native function by name in the new FunctionRegistry.
// Returns the function value and true if found, nil and false otherwise.
func LookupFunction(name string) (*value.FunctionValue, bool) {
	fn, ok := FunctionRegistry[name]
	return fn, ok
}

// CallFunction invokes a native function (FunctionValue) with the given arguments and evaluator.
// The evaluator is always passed to the native function (even if the function doesn't use it).
func CallFunction(fn *value.FunctionValue, args []value.Value, eval Evaluator) (value.Value, error) {
	if fn.Type != value.FuncNative {
		return value.NoneVal(), verror.NewInternalError(
			"CallFunction() expects native function", [3]string{})
	}

	// Create an adapter to bridge native.Evaluator (returns *verror.Error)
	// to value.Evaluator (returns error)
	adapter := evaluatorAdapter{eval}
	return fn.Native(args, adapter)
}

// evaluatorAdapter wraps native.Evaluator to implement value.Evaluator.
// This bridges the difference in return types (*verror.Error vs error).
type evaluatorAdapter struct {
	eval Evaluator
}

func (a evaluatorAdapter) Do_Blk(vals []value.Value) (value.Value, error) {
	result, err := a.eval.Do_Blk(vals)
	return result, err // *verror.Error implements error interface
}

func (a evaluatorAdapter) Do_Next(val value.Value) (value.Value, error) {
	result, err := a.eval.Do_Next(val)
	return result, err // *verror.Error implements error interface
}

// nativeEvaluatorAdapter wraps value.Evaluator to implement native.Evaluator.
// This is the reverse of evaluatorAdapter - converts value.Evaluator (error) back to native.Evaluator (*verror.Error).
// Special case: if the value.Evaluator is actually an evaluatorAdapter, unwrap it to get the original native.Evaluator.
type nativeEvaluatorAdapter struct {
	eval value.Evaluator
}

func (a *nativeEvaluatorAdapter) unwrap() Evaluator {
	// If the eval is an evaluatorAdapter, unwrap it to get the original
	if adapter, ok := a.eval.(evaluatorAdapter); ok {
		return adapter.eval
	}
	// Otherwise, this adapter is the best we can do
	return a
}

func (a *nativeEvaluatorAdapter) Do_Blk(vals []value.Value) (value.Value, *verror.Error) {
	result, err := a.eval.Do_Blk(vals)
	if err == nil {
		return result, nil
	}
	// Convert error to *verror.Error
	if verr, ok := err.(*verror.Error); ok {
		return result, verr
	}
	// If it's not a *verror.Error, wrap it
	return value.NoneVal(), verror.NewInternalError(err.Error(), [3]string{})
}

func (a *nativeEvaluatorAdapter) Do_Next(val value.Value) (value.Value, *verror.Error) {
	result, err := a.eval.Do_Next(val)
	if err == nil {
		return result, nil
	}
	// Convert error to *verror.Error
	if verr, ok := err.(*verror.Error); ok {
		return result, verr
	}
	// If it's not a *verror.Error, wrap it
	return value.NoneVal(), verror.NewInternalError(err.Error(), [3]string{})
}
