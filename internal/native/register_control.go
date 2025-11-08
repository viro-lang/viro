// Package native provides built-in native functions for the Viro interpreter.
//
// This file contains control flow native function registrations.
package native

import (
	"fmt"

	"github.com/marcin-radoszewski/viro/internal/core"
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
func RegisterControlNatives(rootFrame core.Frame) {
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
		rootFrame.Bind(name, value.NewFuncVal(fn))

		// Mark as registered
		registered[name] = true
	}

	// Group 10: Control flow (4 functions - all need evaluator)
	registerAndBind("when", value.NewNativeFunction(
		"when",
		[]value.ParamSpec{
			value.NewParamSpec("condition", true),
			value.NewParamSpec("body", false),
		},
		When,
		false,
		&NativeDoc{
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
	))

	registerAndBind("if", value.NewNativeFunction(
		"if",
		[]value.ParamSpec{
			value.NewParamSpec("condition", true),
			value.NewParamSpec("true-branch", false),
			value.NewParamSpec("false-branch", false),
		},
		If,
		false,
		&NativeDoc{
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
	))

	registerAndBind("loop", value.NewNativeFunction(
		"loop",
		[]value.ParamSpec{
			value.NewParamSpec("count", true),
			value.NewParamSpec("body", false),
			value.NewRefinementSpec("with-index", true),
		},
		Loop,
		false,
		&NativeDoc{
			Category: "Control",
			Summary:  "Executes a block a specified number of times",
			Description: `Evaluates the body block repeatedly for the specified number of iterations.
The count must be a non-negative integer. Returns the result of the last iteration, or none if count is 0.
Index starts at 0 and increments by 1 per iteration (0-based).

Refinements:
  --with-index 'word: Binds the current iteration index (0, 1, 2, ...) to the specified word.`,
			Parameters: []ParamDoc{
				{Name: "count", Type: "integer!", Description: "The number of times to execute the body (evaluated)", Optional: false},
				{Name: "body", Type: "block!", Description: "The code to execute repeatedly", Optional: false},
			},
			Returns:  "[any-type! none!] The result of the last iteration",
			Examples: []string{"loop 3 [print \"hello\"]  ; prints 'hello' 3 times", "x: 0\nloop 5 [x: x + 1]  ; x becomes 5", "loop 0 [print \"never\"]  ; => none", "loop 3 --with-index 'i [print i]  ; prints: 0 1 2"},
			SeeAlso:  []string{"while", "if", "when"}, Tags: []string{"control", "loop", "iteration", "repeat"},
		},
	))

	registerAndBind("while", value.NewNativeFunction(
		"while",
		[]value.ParamSpec{
			value.NewParamSpec("condition", true),
			value.NewParamSpec("body", false),
		},
		While,
		false,
		&NativeDoc{
			Category: "Control",
			Summary:  "Executes a block repeatedly while a condition is true",
			Description: `Evaluates the condition, and if it's true (non-zero, non-empty), evaluates the body block.
Repeats this process until the condition becomes false. If the condition is a block, it is re-evaluated
before each iteration. If the condition is not a block, it is evaluated once at the start and must remain
constant. Returns the result of the last iteration, or none if the condition is initially false.
Be careful to avoid infinite loops.`,
			Parameters: []ParamDoc{
				{Name: "condition", Type: "any-type!", Description: "The condition to test (blocks are re-evaluated each iteration, other values are constant)", Optional: false},
				{Name: "body", Type: "block!", Description: "The code to execute while condition is true", Optional: false},
			},
			Returns:  "[any-type! none!] The result of the last iteration",
			Examples: []string{"x: 0\nwhile [x < 5] [x: x + 1]  ; x becomes 5", "count: 10\nwhile [count > 0] [print count count: count - 1]", "while false [print \"never\"]  ; => none"},
			SeeAlso:  []string{"loop", "if", "when"}, Tags: []string{"control", "loop", "while", "iteration"},
		},
	))

	registerAndBind("foreach", value.NewNativeFunction(
		"foreach",
		[]value.ParamSpec{
			value.NewParamSpec("series", true),
			value.NewParamSpec("vars", false),
			value.NewParamSpec("body", false),
			value.NewRefinementSpec("with-index", true),
		},
		Foreach,
		false,
		&NativeDoc{
			Category:    "Control",
			Summary:     "Iterates over a series, binding each element to a variable",
			Description: "Iterates over any series type (block!, string!, binary!), binding each element to one or more variables and executing a body block. The loop variable(s) are bound in the current scope (not a new scope), allowing access to outer variables. Returns the result of the last iteration, or none if the series is empty. Supports multiple variables for multi-value assignment. Index represents the iteration number (0-based) regardless of how many elements are consumed per iteration.\n\nRefinements:\n  --with-index 'word: Binds the current iteration index (0, 1, 2, ...) to the specified word.",
			Parameters: []ParamDoc{
				{Name: "series", Type: "block! string! binary!", Description: "The series to iterate over (evaluated)", Optional: false},
				{Name: "vars", Type: "word! block!", Description: "A single word or block of words for the loop variable(s) (quoted)", Optional: false},
				{Name: "body", Type: "block!", Description: "The code to execute for each element", Optional: false},
			},
			Returns: "[any-type! none!] The result of the last iteration, or none if series is empty",
			Examples: []string{
				"foreach [1 2 3] [n] [print n]  ; prints: 1 2 3",
				"foreach [1 2 3] n [print n]  ; single word (quoted)",
				"sum: 0\nforeach [10 20 30] [n] [sum: (+ sum n)]  ; sum becomes 60",
				"foreach \"hello\" [c] [print c]  ; prints each character",
				"foreach [1 2 3 4 5 6] [a b] [print [a b]]  ; multi-value assignment",
				"foreach [a b c] --with-index 'pos [print pos]  ; prints: 0 1 2",
				"foreach [10 20 30] --with-index 'i [v] [print [i v]]  ; prints: [0 10] [1 20] [2 30]",
			},
			SeeAlso: []string{"loop", "while", "map", "filter"},
			Tags:    []string{"control", "iteration", "loop", "foreach"},
		},
	))

	// Group 11: Function creation (1 function - needs evaluator)
	registerAndBind("fn", value.NewNativeFunction(
		"fn",
		[]value.ParamSpec{
			value.NewParamSpec("params", false),
			value.NewParamSpec("body", false),
		},
		Fn,
		false,
		&NativeDoc{
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
	))

	// Group 12: Block manipulation (2 functions - need evaluator)
	registerAndBind("compose", value.NewNativeFunction(
		"compose",
		[]value.ParamSpec{
			value.NewParamSpec("block", false),
		},
		Compose,
		false,
		&NativeDoc{
			Category: "Control",
			Summary:  "Evaluates parenthetical expressions within a block",
			Description: `Takes a block and evaluates any parenthetical expressions (expressions in parentheses)
within it. Other elements remain unevaluated. Returns a new block with the evaluated results.`,
			Parameters: []ParamDoc{
				{Name: "block", Type: "block!", Description: "The block containing expressions to compose", Optional: false},
			},
			Returns:  "[block!] A new block with parenthetical expressions evaluated",
			Examples: []string{"name: \"World\"\ncompose [Hello (name)]  ; => [Hello \"World\"]", "x: 10\ny: 20\ncompose [result: (x + y) is (x * 2)]  ; => [result: 30 is 20]"},
			SeeAlso:  []string{"reduce", "form"}, Tags: []string{"control", "evaluation", "block", "compose"},
		},
	))

	registerAndBind("do", value.NewNativeFunction(
		"do",
		[]value.ParamSpec{
			value.NewParamSpec("value", true),
			{Name: "next", Type: value.TypeNone, Optional: true, Refinement: true, TakesValue: true, Eval: false},
		},
		Do,
		false,
		&NativeDoc{
			Category: "Control",
			Summary:  "Evaluates a value",
			Description: `Evaluates a given value. If the value is a block, evaluates all expressions in the block.
If the value is not a block, evaluates the expression.

Refinements:
  --next word: Evaluate only the next expression in a block and bind the remaining block to the specified word.
               If the value is not a block, the word is not bound.`,
			Parameters: []ParamDoc{
				{Name: "value", Type: "any-type!", Description: "The value to evaluate", Optional: false},
			},
			Returns: "[any-type!] The result of the evaluation",
			Examples: []string{
				"a: [print \"Foo\" 10]\ndo a  ; prints \"Foo\" and returns 10",
				"do a --next b  ; prints \"Foo\" and binds remaining block to b",
				"print head? b  ; prints false",
				"print index? b  ; prints 3",
			},
			SeeAlso: []string{"reduce", "compose", "eval"},
			Tags:    []string{"control", "evaluation", "do"},
		},
	))

	registerAndBind("break", value.NewNativeFunction(
		"break",
		[]value.ParamSpec{
			value.NewRefinementSpec("levels", true),
		},
		Break,
		false,
		&NativeDoc{
			Category:    "Control",
			Summary:     "Exits one or more nested loops immediately",
			Description: "Causes immediate exit from the innermost loop (loop, while, or foreach), or multiple nested loops if --levels is specified. Returns none. Can only be used inside a loop; using break outside a loop causes an error. Break does not cross function boundaries.\n\nRefinements:\n  --levels N: Exit N levels of nested loops (N >= 1, default 1)",
			Parameters:  []ParamDoc{},
			Returns:     "[none!] Always returns none",
			Examples:    []string{"loop 10 [when (= x 5) [break]]  ; exits when x is 5", "foreach [1 2 3 4 5] 'n [when (= n 3) [break]]  ; stops at 3", "loop 3 [loop 3 [when (= x 2) [break --levels 2]]]  ; exits both loops"},
			SeeAlso:     []string{"continue", "loop", "while", "foreach"},
			Tags:        []string{"control", "loop", "break", "multilevel"},
		},
	))

	registerAndBind("continue", value.NewNativeFunction(
		"continue",
		[]value.ParamSpec{
			value.NewRefinementSpec("levels", true),
		},
		Continue,
		false,
		&NativeDoc{
			Category:    "Control",
			Summary:     "Skips to the next iteration of one or more nested loops",
			Description: "Skips the rest of the current iteration and proceeds to the next iteration of the innermost loop (loop, while, or foreach), or multiple nested loops if --levels is specified. Can only be used inside a loop; using continue outside a loop causes an error. Continue does not cross function boundaries.\n\nRefinements:\n  --levels N: Continue N levels of nested loops (N >= 1, default 1)",
			Parameters:  []ParamDoc{},
			Returns:     "[none!] Always returns none",
			Examples:    []string{"loop 5 [when (= (mod i 2) 0) [continue] print i]  ; prints odd numbers", "foreach [1 2 3 4 5] 'n [when (= n 3) [continue] print n]  ; skips 3", "loop 3 [loop 3 [when (= x 3) [continue --levels 2]]]  ; continues outer loop"},
			SeeAlso:     []string{"break", "loop", "while", "foreach"},
			Tags:        []string{"control", "loop", "continue", "multilevel"},
		},
	))

	registerAndBind("return", value.NewNativeFunction(
		"return",
		[]value.ParamSpec{
			{Name: "value", Type: value.TypeNone, Optional: true, Refinement: false, TakesValue: false, Eval: true},
		},
		Return,
		false,
		&NativeDoc{
			Category:    "Control",
			Summary:     "Returns a value from a function",
			Description: "Returns a value from the current function or script, terminating execution and returning control to the caller. If no value is provided, returns none. Can be used inside functions or at the top level of scripts and REPL.",
			Parameters: []ParamDoc{
				{Name: "value", Type: "any-type!", Description: "The value to return (optional, defaults to none)", Optional: true},
			},
			Returns:  "[any-type! none!] The specified value or none",
			Examples: []string{"fn [x] [when x < 0 [return 0] x * 2]  ; returns 0 for negative inputs", "fn [] [return \"hello\"]  ; returns \"hello\"", "fn [] [return]  ; returns none"},
			SeeAlso:  []string{"fn", "break", "continue"},
			Tags:     []string{"control", "function", "return"},
		},
	))
}
