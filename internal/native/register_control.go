// Package native provides built-in native functions for the Viro interpreter.
//
// This file contains control flow native function registrations.
package native

import (
	"fmt"

	"github.com/marcin-radoszewski/viro/internal/frame"
	"github.com/marcin-radoszewski/viro/internal/value"
)

// RegisterControlNatives registers all control flow native functions in the root frame.
//
// The function panics if:
//   - rootFrame is nil
//   - Any duplicate function name is registered
//   - Any function creation fails
//
// This is intentional fail-fast behavior for critical initialization errors.
func RegisterControlNatives(rootFrame *frame.Frame) {
	// Validation: Track registered names to detect duplicates
	registered := make(map[string]bool)

	// Helper function to register and bind a native function
	registerAndBind := func(name string, fn *value.FunctionValue) {
		if fn == nil {
			panic(fmt.Sprintf("RegisterControlNatives: attempted to register nil function for '%s'", name))
		}
		if registered[name] {
			panic(fmt.Sprintf("RegisterControlNatives: duplicate registration of function '%s'", name))
		}

		// Bind to root frame
		rootFrame.Bind(name, value.FuncVal(fn))

		// Mark as registered
		registered[name] = true
	}

	// Group 10: Control flow (4 functions - all need evaluator)
	fn := value.NewNativeFunction(
		"when",
		[]value.ParamSpec{
			value.NewParamSpec("condition", true), // evaluated
			value.NewParamSpec("body", false),     // NOT evaluated (block)
		},
		func(args []value.Value, refValues map[string]value.Value, eval value.Evaluator) (value.Value, error) {
			reverseAdapter := &nativeEvaluatorAdapter{eval}
			result, err := When(args, refValues, reverseAdapter.unwrap())
			if err == nil {
				return result, nil
			}
			return result, err
		},
	)
	fn.Doc = &NativeDoc{
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
	}
	registerAndBind("when", fn)

	fn = value.NewNativeFunction(
		"if",
		[]value.ParamSpec{
			value.NewParamSpec("condition", true),     // evaluated
			value.NewParamSpec("true-branch", false),  // NOT evaluated (block)
			value.NewParamSpec("false-branch", false), // NOT evaluated (block)
		},
		func(args []value.Value, refValues map[string]value.Value, eval value.Evaluator) (value.Value, error) {
			reverseAdapter := &nativeEvaluatorAdapter{eval}
			result, err := If(args, refValues, reverseAdapter.unwrap())
			if err == nil {
				return result, nil
			}
			return result, err
		},
	)
	fn.Doc = &NativeDoc{
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
	}
	registerAndBind("if", fn)

	fn = value.NewNativeFunction(
		"loop",
		[]value.ParamSpec{
			value.NewParamSpec("count", true), // evaluated
			value.NewParamSpec("body", false), // NOT evaluated (block)
		},
		func(args []value.Value, refValues map[string]value.Value, eval value.Evaluator) (value.Value, error) {
			reverseAdapter := &nativeEvaluatorAdapter{eval}
			result, err := Loop(args, refValues, reverseAdapter.unwrap())
			if err == nil {
				return result, nil
			}
			return result, err
		},
	)
	fn.Doc = &NativeDoc{
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
	}
	registerAndBind("loop", fn)

	fn = value.NewNativeFunction(
		"while",
		[]value.ParamSpec{
			value.NewParamSpec("condition", true), // evaluated
			value.NewParamSpec("body", false),     // NOT evaluated (block)
		},
		func(args []value.Value, refValues map[string]value.Value, eval value.Evaluator) (value.Value, error) {
			reverseAdapter := &nativeEvaluatorAdapter{eval}
			result, err := While(args, refValues, reverseAdapter.unwrap())
			if err == nil {
				return result, nil
			}
			return result, err
		},
	)
	fn.Doc = &NativeDoc{
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
	}
	registerAndBind("while", fn)

	// Group 11: Function creation (1 function - needs evaluator)
	fn = value.NewNativeFunction(
		"fn",
		[]value.ParamSpec{
			value.NewParamSpec("params", false), // NOT evaluated (block)
			value.NewParamSpec("body", false),   // NOT evaluated (block)
		},
		func(args []value.Value, refValues map[string]value.Value, eval value.Evaluator) (value.Value, error) {
			reverseAdapter := &nativeEvaluatorAdapter{eval}
			result, err := Fn(args, refValues, reverseAdapter.unwrap())
			if err == nil {
				return result, nil
			}
			return result, err
		},
	)
	fn.Doc = &NativeDoc{
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
	}
	registerAndBind("fn", fn)
}
