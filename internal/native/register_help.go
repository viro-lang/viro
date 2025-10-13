package native

import (
	"github.com/marcin-radoszewski/viro/internal/frame"
	"github.com/marcin-radoszewski/viro/internal/value"
	"github.com/marcin-radoszewski/viro/internal/verror"
)

// RegisterHelpNatives registers all help, debug, and reflection native functions into the root frame.
//
// Panic behavior:
// - Panics if rootFrame is nil (fail-fast on invalid state)
// - Panics if duplicate binding detected (programming error)
// - Panics if function value creation fails (critical error)
//
// Note: This function also populates the legacy Registry map for backward compatibility during Phase 1 migration.
func RegisterHelpNatives(rootFrame *frame.Frame) {
	if rootFrame == nil {
		panic("RegisterHelpNatives: rootFrame is nil")
	}

	var fn *value.FunctionValue

	// Helper function to wrap simple math/reflection functions
	registerSimpleMathFunc := func(name string, impl func([]value.Value) (value.Value, *verror.Error), arity int, doc *NativeDoc) {
		// Extract parameter names from existing documentation
		params := make([]value.ParamSpec, arity)

		if doc != nil && len(doc.Parameters) == arity {
			// Use parameter names from documentation
			for i := 0; i < arity; i++ {
				params[i] = value.NewParamSpec(doc.Parameters[i].Name, true)
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

		fn := value.NewNativeFunction(
			name,
			params,
			func(args []value.Value, refValues map[string]value.Value, eval value.Evaluator) (value.Value, error) {
				result, err := impl(args)
				if err == nil {
					return result, nil
				}
				return result, err
			},
		)
		fn.Doc = doc

		// Bind to root frame
		rootFrame.Bind(name, value.FuncVal(fn))
	}

	// Group 12: Help system (2 functions)
	fn = value.NewNativeFunction(
		"?",
		[]value.ParamSpec{
			value.NewParamSpec("topic", false), // NOT evaluated (word/string)
		},
		func(args []value.Value, refValues map[string]value.Value, eval value.Evaluator) (value.Value, error) {
			reverseAdapter := &nativeEvaluatorAdapter{eval}
			result, err := Help(args, reverseAdapter.unwrap())
			if err == nil {
				return result, nil
			}
			return result, err
		},
	)
	fn.Doc = &NativeDoc{
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
	}
	// Bind to root frame
	rootFrame.Bind("?", value.FuncVal(fn))

	fn = value.NewNativeFunction(
		"words",
		[]value.ParamSpec{},
		func(args []value.Value, refValues map[string]value.Value, eval value.Evaluator) (value.Value, error) {
			reverseAdapter := &nativeEvaluatorAdapter{eval}
			result, err := Words(args, reverseAdapter.unwrap())
			if err == nil {
				return result, nil
			}
			return result, err
		},
	)
	fn.Doc = &NativeDoc{
		Category: "Help",
		Summary:  "Lists all available native function names",
		Description: `Returns a block containing all native function names as words.
Does not print by default - use 'print words' to display the list.
Useful for programmatic access to available functionality.`,
		Parameters: []ParamDoc{},
		Returns:    "[block!] A block containing all function names as words",
		Examples:   []string{"words  ; return all function names", "fns: words\nlength? fns  ; count available functions", "print words  ; display function names"},
		SeeAlso:    []string{"?", "type?"}, Tags: []string{"help", "documentation", "discovery", "list"},
	}
	// Bind to root frame
	rootFrame.Bind("words", value.FuncVal(fn))

	// Group 13: Trace/Debug/Reflection (9 functions - Feature 002, FR-020 to FR-022)
	fn = value.NewNativeFunction(
		"trace",
		[]value.ParamSpec{
			value.NewRefinementSpec("on", false),
			value.NewRefinementSpec("off", false),
			value.NewRefinementSpec("only", true),
			value.NewRefinementSpec("exclude", true),
			value.NewRefinementSpec("file", true),
			value.NewRefinementSpec("append", false),
		},
		func(args []value.Value, refValues map[string]value.Value, eval value.Evaluator) (value.Value, error) {
			reverseAdapter := &nativeEvaluatorAdapter{eval}
			result, err := Trace(args, refValues, reverseAdapter.unwrap())
			if err == nil {
				return result, nil
			}
			return result, err
		},
	)
	fn.Doc = &NativeDoc{
		Category: "Debug",
		Summary:  "Controls execution tracing",
		Description: `Enables or disables tracing of code execution. When enabled, traces function calls
and other execution events to a log file. Supports filtering and custom output destinations.`,
		Parameters: []ParamDoc{
			{Name: "--on", Type: "logic!", Description: "Enable tracing", Optional: true},
			{Name: "--off", Type: "logic!", Description: "Disable tracing", Optional: true},
			{Name: "--only", Type: "block!", Description: "Block of words to include in trace (whitelist)", Optional: true},
			{Name: "--exclude", Type: "block!", Description: "Block of words to exclude from trace (blacklist)", Optional: true},
			{Name: "--file", Type: "string!", Description: "Custom file path for trace output", Optional: true},
			{Name: "--append", Type: "logic!", Description: "Append to trace file instead of overwriting", Optional: true},
		},
		Returns:  "[none!] Always returns none",
		Examples: []string{"trace --on  ; enable tracing", "trace --on --only [calculate-interest]  ; trace specific function", "trace --off  ; disable tracing"},
		SeeAlso:  []string{"trace?", "debug"}, Tags: []string{"debug", "trace", "observability"},
	}
	// Bind to root frame
	rootFrame.Bind("trace", value.FuncVal(fn))

	fn = value.NewNativeFunction(
		"trace?",
		[]value.ParamSpec{},
		func(args []value.Value, refValues map[string]value.Value, eval value.Evaluator) (value.Value, error) {
			reverseAdapter := &nativeEvaluatorAdapter{eval}
			result, err := TraceQuery(args, refValues, reverseAdapter.unwrap())
			if err == nil {
				return result, nil
			}
			return result, err
		},
	)
	fn.Doc = &NativeDoc{
		Category:    "Debug",
		Summary:     "Queries trace status",
		Description: `Returns a boolean indicating whether tracing is currently enabled.`,
		Parameters:  []ParamDoc{},
		Returns:     "[logic!] true if tracing is enabled, false otherwise",
		Examples:    []string{"trace?  ; => false", "trace --on\ntrace?  ; => true"},
		SeeAlso:     []string{"trace", "debug"}, Tags: []string{"debug", "trace", "query"},
	}
	// Bind to root frame
	rootFrame.Bind("trace?", value.FuncVal(fn))

	fn = value.NewNativeFunction(
		"debug",
		[]value.ParamSpec{
			value.NewRefinementSpec("on", false),
			value.NewRefinementSpec("off", false),
			{Name: "breakpoint", Type: value.TypeNone, Optional: true, Refinement: true, TakesValue: true, Eval: false}, // Don't evaluate - we want lit-word
			{Name: "remove", Type: value.TypeNone, Optional: true, Refinement: true, TakesValue: true, Eval: true},      // Evaluate - we want integer
			value.NewRefinementSpec("step", false),
			value.NewRefinementSpec("next", false),
			value.NewRefinementSpec("finish", false),
			value.NewRefinementSpec("continue", false),
			value.NewRefinementSpec("locals", false),
			value.NewRefinementSpec("stack", false),
		},
		func(args []value.Value, refValues map[string]value.Value, eval value.Evaluator) (value.Value, error) {
			reverseAdapter := &nativeEvaluatorAdapter{eval}
			result, err := Debug(args, refValues, reverseAdapter.unwrap())
			if err == nil {
				return result, nil
			}
			return result, err
		},
	)
	fn.Doc = &NativeDoc{
		Category: "Debug",
		Summary:  "Controls the interactive debugger",
		Description: `Provides debugging capabilities including breakpoints, stepping, and inspection.
Use --on to enable the debugger, set breakpoints with --breakpoint, and control execution flow
with stepping commands. Inspect state with --locals and --stack.`,
		Parameters: []ParamDoc{
			{Name: "--on", Type: "logic!", Description: "Enable debugger", Optional: true},
			{Name: "--off", Type: "logic!", Description: "Disable debugger", Optional: true},
			{Name: "--breakpoint", Type: "word!", Description: "Set breakpoint on word (returns breakpoint ID)", Optional: true},
			{Name: "--remove", Type: "integer!", Description: "Remove breakpoint by ID", Optional: true},
			{Name: "--step", Type: "logic!", Description: "Step into next expression", Optional: true},
			{Name: "--next", Type: "logic!", Description: "Step over next expression", Optional: true},
			{Name: "--finish", Type: "logic!", Description: "Continue until function returns", Optional: true},
			{Name: "--continue", Type: "logic!", Description: "Resume normal execution", Optional: true},
			{Name: "--locals", Type: "logic!", Description: "Show local variables", Optional: true},
			{Name: "--stack", Type: "logic!", Description: "Show call stack", Optional: true},
		},
		Returns: "[integer! block! none!] Breakpoint ID, inspection data, or none",
		Examples: []string{
			"debug --on  ; enable debugger",
			"debug --breakpoint 'square  ; set breakpoint",
			"debug --step  ; step into",
			"debug --locals  ; show locals",
		},
		SeeAlso: []string{"trace", "trace?"}, Tags: []string{"debug", "breakpoint", "stepping"},
	}
	// Bind to root frame
	rootFrame.Bind("debug", value.FuncVal(fn))

	registerSimpleMathFunc("type-of", TypeOf, 1, &NativeDoc{
		Category: "Reflection",
		Summary:  "Returns the type name of a value",
		Description: `Returns a word representing the canonical type name of any value.
Type names follow the pattern 'type!' (e.g., integer!, string!, block!).`,
		Parameters: []ParamDoc{
			{Name: "value", Type: "any-type!", Description: "The value to get the type of", Optional: false},
		},
		Returns:  "[word!] The type name as a word",
		Examples: []string{"type-of 42  ; => integer!", `type-of "hello"  ; => string!`, "type-of [1 2 3]  ; => block!"},
		SeeAlso:  []string{"type?", "spec-of", "body-of"}, Tags: []string{"reflection", "type", "introspection"},
	})

	registerSimpleMathFunc("spec-of", SpecOf, 1, &NativeDoc{
		Category: "Reflection",
		Summary:  "Returns the specification of a function or object",
		Description: `Extracts the specification (parameter list or field definitions) from a function or object.
Returns an immutable copy as a block. For functions, returns the parameter spec. For objects,
returns the field names and type hints.`,
		Parameters: []ParamDoc{
			{Name: "value", Type: "function! object!", Description: "The function or object to inspect", Optional: false},
		},
		Returns:  "[block!] The specification as a block",
		Examples: []string{"square: fn [x] [x * x]\nspec-of :square  ; => [x]", "obj: object [name: \"Alice\"]\nspec-of obj  ; => [name]"},
		SeeAlso:  []string{"body-of", "type-of", "words-of"}, Tags: []string{"reflection", "spec", "introspection"},
	})

	registerSimpleMathFunc("body-of", BodyOf, 1, &NativeDoc{
		Category: "Reflection",
		Summary:  "Returns the body of a function or object",
		Description: `Extracts the body block from a function or object. Returns an immutable deep copy
to prevent mutation of the original. For functions, returns the function body. For objects,
returns a block of set-word/value pairs.`,
		Parameters: []ParamDoc{
			{Name: "value", Type: "function! object!", Description: "The function or object to inspect", Optional: false},
		},
		Returns:  "[block!] The body as a block",
		Examples: []string{"square: fn [x] [x * x]\nbody-of :square  ; => [x * x]", "obj: object [x: 10]\nbody-of obj  ; => [x: 10]"},
		SeeAlso:  []string{"spec-of", "type-of", "source"}, Tags: []string{"reflection", "body", "introspection"},
	})

	fn = value.NewNativeFunction(
		"words-of",
		[]value.ParamSpec{
			value.NewParamSpec("object", true),
		},
		func(args []value.Value, refValues map[string]value.Value, eval value.Evaluator) (value.Value, error) {
			reverseAdapter := &nativeEvaluatorAdapter{eval}
			result, err := WordsOf(args, refValues, reverseAdapter.unwrap())
			if err == nil {
				return result, nil
			}
			return result, err
		},
	)
	fn.Doc = &NativeDoc{
		Category: "Reflection",
		Summary:  "Returns the field names of an object",
		Description: `Extracts all field names (words) from an object as a block. The order matches
the object's manifest. Returns an immutable block of words.`,
		Parameters: []ParamDoc{
			{Name: "object", Type: "object!", Description: "The object to inspect", Optional: false},
		},
		Returns:  "[block!] Block of field names as words",
		Examples: []string{"obj: object [name: \"Alice\" age: 30]\nwords-of obj  ; => [name age]", "words-of object []  ; => []"},
		SeeAlso:  []string{"values-of", "spec-of"}, Tags: []string{"reflection", "object", "fields"},
	}
	// Bind to root frame
	rootFrame.Bind("words-of", value.FuncVal(fn))

	fn = value.NewNativeFunction(
		"values-of",
		[]value.ParamSpec{
			value.NewParamSpec("object", true),
		},
		func(args []value.Value, refValues map[string]value.Value, eval value.Evaluator) (value.Value, error) {
			reverseAdapter := &nativeEvaluatorAdapter{eval}
			result, err := ValuesOf(args, refValues, reverseAdapter.unwrap())
			if err == nil {
				return result, nil
			}
			return result, err
		},
	)
	fn.Doc = &NativeDoc{
		Category: "Reflection",
		Summary:  "Returns the field values of an object",
		Description: `Extracts all field values from an object as a block. The order matches the object's
manifest and corresponds to words-of. Returns deep copies to prevent mutation.`,
		Parameters: []ParamDoc{
			{Name: "object", Type: "object!", Description: "The object to inspect", Optional: false},
		},
		Returns:  "[block!] Block of field values",
		Examples: []string{"obj: object [name: \"Alice\" age: 30]\nvalues-of obj  ; => [\"Alice\" 30]", "values-of object []  ; => []"},
		SeeAlso:  []string{"words-of", "spec-of"}, Tags: []string{"reflection", "object", "values"},
	}
	// Bind to root frame
	rootFrame.Bind("values-of", value.FuncVal(fn))

	registerSimpleMathFunc("source", Source, 1, &NativeDoc{
		Category: "Reflection",
		Summary:  "Returns formatted source code for a function or object",
		Description: `Reconstructs a readable source code representation of a function or object.
For functions, returns the complete definition including spec and body. For objects,
returns the object definition with field names.`,
		Parameters: []ParamDoc{
			{Name: "value", Type: "function! object!", Description: "The function or object to format", Optional: false},
		},
		Returns:  "[string!] The formatted source code",
		Examples: []string{"square: fn [x] [x * x]\nsource :square  ; => \"fn [x] [x * x]\"", "obj: object [x: 10]\nsource obj  ; => \"object [x]\""},
		SeeAlso:  []string{"spec-of", "body-of", "type-of"}, Tags: []string{"reflection", "source", "format"},
	})
}
