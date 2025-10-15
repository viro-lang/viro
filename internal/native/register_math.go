// Package native provides built-in native functions for the Viro interpreter.
package native

import (
	"fmt"

	"github.com/marcin-radoszewski/viro/internal/core"
	"github.com/marcin-radoszewski/viro/internal/value"
)

// RegisterMathNatives registers all math-related native functions to the root frame.
//
// Panics if any function is nil or if a duplicate name is detected during registration.
func RegisterMathNatives(rootFrame core.Frame) {
	// Validation: Track registered names to detect duplicates
	registered := make(map[string]bool)

	// Helper function to register and bind a native function
	registerAndBind := func(name string, fn *value.FunctionValue) {
		if fn == nil {
			panic(fmt.Sprintf("RegisterMathNatives: attempted to register nil function for '%s'", name))
		}
		if registered[name] {
			panic(fmt.Sprintf("RegisterMathNatives: duplicate registration of function '%s'", name))
		}

		// Bind to root frame
		rootFrame.Bind(name, value.FuncVal(fn))

		// Mark as registered
		registered[name] = true
	}

	// ===== Group 1: Simple math operations (4 functions) =====
	fn := value.NewNativeFunction(
		"+",
		[]value.ParamSpec{
			value.NewParamSpec("left", true),
			value.NewParamSpec("right", true),
		},
		func(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
			result, err := Add(args)
			if err == nil {
				return result, nil
			}
			return result, err
		},
	)
	fn.Infix = true
	fn.Doc = &NativeDoc{
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
	}
	registerAndBind("+", fn)

	fn = value.NewNativeFunction(
		"-",
		[]value.ParamSpec{
			value.NewParamSpec("left", true),
			value.NewParamSpec("right", true),
		},
		func(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
			result, err := Subtract(args)
			if err == nil {
				return result, nil
			}
			return result, err
		},
	)
	fn.Infix = true
	fn.Doc = &NativeDoc{
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
	}
	registerAndBind("-", fn)

	fn = value.NewNativeFunction(
		"*",
		[]value.ParamSpec{
			value.NewParamSpec("left", true),
			value.NewParamSpec("right", true),
		},
		func(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
			result, err := Multiply(args)
			if err == nil {
				return result, nil
			}
			return result, err
		},
	)
	fn.Infix = true
	fn.Doc = &NativeDoc{
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
	}
	registerAndBind("*", fn)

	fn = value.NewNativeFunction(
		"/",
		[]value.ParamSpec{
			value.NewParamSpec("left", true),
			value.NewParamSpec("right", true),
		},
		func(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
			result, err := Divide(args)
			if err == nil {
				return result, nil
			}
			return result, err
		},
	)
	fn.Infix = true
	fn.Doc = &NativeDoc{
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
	}
	registerAndBind("/", fn)

	// ===== Group 2: Comparison operators (6 functions) =====
	fn = value.NewNativeFunction(
		"<",
		[]value.ParamSpec{
			value.NewParamSpec("left", true),
			value.NewParamSpec("right", true),
		},
		func(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
			result, err := LessThan(args)
			if err == nil {
				return result, nil
			}
			return result, err
		},
	)
	fn.Infix = true
	fn.Doc = &NativeDoc{
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
	}
	registerAndBind("<", fn)

	fn = value.NewNativeFunction(
		">",
		[]value.ParamSpec{
			value.NewParamSpec("left", true),
			value.NewParamSpec("right", true),
		},
		func(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
			result, err := GreaterThan(args)
			if err == nil {
				return result, nil
			}
			return result, err
		},
	)
	fn.Infix = true
	fn.Doc = &NativeDoc{
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
	}
	registerAndBind(">", fn)

	fn = value.NewNativeFunction(
		"<=",
		[]value.ParamSpec{
			value.NewParamSpec("left", true),
			value.NewParamSpec("right", true),
		},
		func(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
			result, err := LessOrEqual(args)
			if err == nil {
				return result, nil
			}
			return result, err
		},
	)
	fn.Infix = true
	fn.Doc = &NativeDoc{
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
	}
	registerAndBind("<=", fn)

	fn = value.NewNativeFunction(
		">=",
		[]value.ParamSpec{
			value.NewParamSpec("left", true),
			value.NewParamSpec("right", true),
		},
		func(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
			result, err := GreaterOrEqual(args)
			if err == nil {
				return result, nil
			}
			return result, err
		},
	)
	fn.Infix = true
	fn.Doc = &NativeDoc{
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
	}
	registerAndBind(">=", fn)

	fn = value.NewNativeFunction(
		"=",
		[]value.ParamSpec{
			value.NewParamSpec("left", true),
			value.NewParamSpec("right", true),
		},
		func(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
			result, err := Equal(args)
			if err == nil {
				return result, nil
			}
			return result, err
		},
	)
	fn.Infix = true
	fn.Doc = &NativeDoc{
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
	}
	registerAndBind("=", fn)

	fn = value.NewNativeFunction(
		"<>",
		[]value.ParamSpec{
			value.NewParamSpec("left", true),
			value.NewParamSpec("right", true),
		},
		func(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
			result, err := NotEqual(args)
			if err == nil {
				return result, nil
			}
			return result, err
		},
	)
	fn.Infix = true
	fn.Doc = &NativeDoc{
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
	}
	registerAndBind("<>", fn)

	// ===== Group 3: Logic operators (3 functions) =====
	fn = value.NewNativeFunction(
		"and",
		[]value.ParamSpec{
			value.NewParamSpec("left", true),
			value.NewParamSpec("right", true),
		},
		func(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
			result, err := And(args)
			if err == nil {
				return result, nil
			}
			return result, err
		},
	)
	fn.Infix = true
	fn.Doc = &NativeDoc{
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
	}
	registerAndBind("and", fn)

	fn = value.NewNativeFunction(
		"or",
		[]value.ParamSpec{
			value.NewParamSpec("left", true),
			value.NewParamSpec("right", true),
		},
		func(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
			result, err := Or(args)
			if err == nil {
				return result, nil
			}
			return result, err
		},
	)
	fn.Infix = true
	fn.Doc = &NativeDoc{
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
	}
	registerAndBind("or", fn)

	fn = value.NewNativeFunction(
		"not",
		[]value.ParamSpec{
			value.NewParamSpec("value", true),
		},
		func(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
			result, err := Not(args)
			if err == nil {
				return result, nil
			}
			return result, err
		},
	)
	fn.Infix = false
	fn.Doc = &NativeDoc{
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
	}
	registerAndBind("not", fn)

	// ===== Group 4: Advanced math functions (16 functions) =====
	// Helper function to wrap simple math functions
	registerSimpleMathFunc := func(name string, impl func([]core.Value) (core.Value, error), arity int, doc *NativeDoc) {
		// Extract parameter names from existing documentation
		params := make([]value.ParamSpec, arity)

		if doc != nil && len(doc.Parameters) == arity {
			// Use parameter names from documentation
			for i := range arity {
				params[i] = value.NewParamSpec(doc.Parameters[i].Name, true)
			}
		} else {
			// Fallback to generic names if documentation is missing or mismatched
			paramNames := []string{"value", "left", "right", "base", "exponent"}
			for i := range arity {
				if i < len(paramNames) {
					params[i] = value.NewParamSpec(paramNames[i], true)
				} else {
					params[i] = value.NewParamSpec("arg", true)
				}
			}
		}

		fn := value.NewNativeFunction(
			name,
			params,
			func(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
				result, err := impl(args)
				if err == nil {
					return result, nil
				}
				return result, err
			},
		)
		fn.Doc = doc
		registerAndBind(name, fn)
	}

	registerSimpleMathFunc("decimal", DecimalConstructor, 1, &NativeDoc{
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
	})
	registerSimpleMathFunc("pow", Pow, 2, &NativeDoc{
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
	})
	registerSimpleMathFunc("sqrt", Sqrt, 1, &NativeDoc{
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
	})
	registerSimpleMathFunc("exp", Exp, 1, &NativeDoc{
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
	})
	registerSimpleMathFunc("log", Log, 1, &NativeDoc{
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
	})
	registerSimpleMathFunc("log-10", Log10, 1, &NativeDoc{
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
	})
	registerSimpleMathFunc("sin", Sin, 1, &NativeDoc{
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
	})
	registerSimpleMathFunc("cos", Cos, 1, &NativeDoc{
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
	})
	registerSimpleMathFunc("tan", Tan, 1, &NativeDoc{
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
	})
	registerSimpleMathFunc("asin", Asin, 1, &NativeDoc{
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
	})
	registerSimpleMathFunc("acos", Acos, 1, &NativeDoc{
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
	})
	registerSimpleMathFunc("atan", Atan, 1, &NativeDoc{
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
	})
	registerSimpleMathFunc("round", Round, 1, &NativeDoc{
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
	})
	registerSimpleMathFunc("ceil", Ceil, 1, &NativeDoc{
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
	})
	registerSimpleMathFunc("floor", Floor, 1, &NativeDoc{
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
	})
	registerSimpleMathFunc("truncate", Truncate, 1, &NativeDoc{
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
	})
}
