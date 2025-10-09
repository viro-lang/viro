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
}

// Registry holds all registered native functions.
var Registry = make(map[string]*NativeInfo)

func init() {
	// Register math natives (simple - don't need evaluator)
	Registry["+"] = &NativeInfo{Func: Add, NeedsEval: false, Arity: 2}
	Registry["-"] = &NativeInfo{Func: Subtract, NeedsEval: false, Arity: 2}
	Registry["*"] = &NativeInfo{Func: Multiply, NeedsEval: false, Arity: 2}
	Registry["/"] = &NativeInfo{Func: Divide, NeedsEval: false, Arity: 2}

	// Register comparison operators
	Registry["<"] = &NativeInfo{Func: LessThan, NeedsEval: false, Arity: 2}
	Registry[">"] = &NativeInfo{Func: GreaterThan, NeedsEval: false, Arity: 2}
	Registry["<="] = &NativeInfo{Func: LessOrEqual, NeedsEval: false, Arity: 2}
	Registry[">="] = &NativeInfo{Func: GreaterOrEqual, NeedsEval: false, Arity: 2}
	Registry["="] = &NativeInfo{Func: Equal, NeedsEval: false, Arity: 2}
	Registry["<>"] = &NativeInfo{Func: NotEqual, NeedsEval: false, Arity: 2}

	// Register logic operators
	Registry["and"] = &NativeInfo{Func: And, NeedsEval: false, Arity: 2}
	Registry["or"] = &NativeInfo{Func: Or, NeedsEval: false, Arity: 2}
	Registry["not"] = &NativeInfo{Func: Not, NeedsEval: false, Arity: 1}

	// Register decimal and advanced math natives (Feature 002)
	Registry["decimal"] = &NativeInfo{Func: DecimalConstructor, NeedsEval: false, Arity: 1}
	Registry["pow"] = &NativeInfo{Func: Pow, NeedsEval: false, Arity: 2}
	Registry["sqrt"] = &NativeInfo{Func: Sqrt, NeedsEval: false, Arity: 1}
	Registry["exp"] = &NativeInfo{Func: Exp, NeedsEval: false, Arity: 1}
	Registry["log"] = &NativeInfo{Func: Log, NeedsEval: false, Arity: 1}
	Registry["log-10"] = &NativeInfo{Func: Log10, NeedsEval: false, Arity: 1}
	Registry["sin"] = &NativeInfo{Func: Sin, NeedsEval: false, Arity: 1}
	Registry["cos"] = &NativeInfo{Func: Cos, NeedsEval: false, Arity: 1}
	Registry["tan"] = &NativeInfo{Func: Tan, NeedsEval: false, Arity: 1}
	Registry["asin"] = &NativeInfo{Func: Asin, NeedsEval: false, Arity: 1}
	Registry["acos"] = &NativeInfo{Func: Acos, NeedsEval: false, Arity: 1}
	Registry["atan"] = &NativeInfo{Func: Atan, NeedsEval: false, Arity: 1}
	Registry["round"] = &NativeInfo{Func: Round, NeedsEval: false, Arity: 1} // TODO: support --places refinement
	Registry["ceil"] = &NativeInfo{Func: Ceil, NeedsEval: false, Arity: 1}
	Registry["floor"] = &NativeInfo{Func: Floor, NeedsEval: false, Arity: 1}
	Registry["truncate"] = &NativeInfo{Func: Truncate, NeedsEval: false, Arity: 1}

	// Register series natives
	Registry["first"] = &NativeInfo{Func: First, NeedsEval: false, Arity: 1}
	Registry["last"] = &NativeInfo{Func: Last, NeedsEval: false, Arity: 1}
	Registry["append"] = &NativeInfo{Func: Append, NeedsEval: false, Arity: 2}
	Registry["insert"] = &NativeInfo{Func: Insert, NeedsEval: false, Arity: 2}
	Registry["length?"] = &NativeInfo{Func: LengthQ, NeedsEval: false, Arity: 1}

	// Register data natives
	Registry["set"] = &NativeInfo{FuncEval: Set, NeedsEval: true, Arity: 2}
	Registry["get"] = &NativeInfo{FuncEval: Get, NeedsEval: true, Arity: 1}
	Registry["type?"] = &NativeInfo{Func: TypeQ, NeedsEval: false, Arity: 1}

	// Register IO natives
	Registry["print"] = &NativeInfo{FuncEval: Print, NeedsEval: true, Arity: 1}
	Registry["input"] = &NativeInfo{Func: Input, NeedsEval: false, Arity: 0}

	// Register function native
	Registry["fn"] = &NativeInfo{FuncEval: Fn, NeedsEval: true, Arity: 2}

	// Register control flow natives (need evaluator)
	Registry["when"] = &NativeInfo{FuncEval: When, NeedsEval: true, Arity: 2}
	Registry["if"] = &NativeInfo{FuncEval: If, NeedsEval: true, Arity: 3}
	Registry["loop"] = &NativeInfo{FuncEval: Loop, NeedsEval: true, Arity: 2}
	Registry["while"] = &NativeInfo{FuncEval: While, NeedsEval: true, Arity: 2}
}

// Lookup finds a native function by name.
// Returns the function info and true if found, nil and false otherwise.
func Lookup(name string) (*NativeInfo, bool) {
	info, ok := Registry[name]
	return info, ok
}

// Call invokes a native function with the given arguments and evaluator.
// Handles both simple natives and natives that need evaluator access.
func Call(info *NativeInfo, args []value.Value, eval Evaluator) (value.Value, *verror.Error) {
	if info.NeedsEval {
		return info.FuncEval(args, eval)
	}
	return info.Func(args)
}
