// Package native provides built-in native functions for the Viro interpreter.
//
// Native functions are implemented in Go and registered in the root frame (frame index 0).
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
// All native functions are unified under the FunctionValue type and stored in the root frame.
package native

import (
	"github.com/marcin-radoszewski/viro/internal/core"
	"github.com/marcin-radoszewski/viro/internal/value"
	"github.com/marcin-radoszewski/viro/internal/verror"
)

// Call invokes a native function (FunctionValue) with the given arguments, refinements, and evaluator.
// The evaluator is always passed to the native function (even if the function doesn't use it).
//
// Arguments:
// - posArgs: positional arguments (already evaluated or raw according to ParamSpec.Eval)
// - refValues: refinement values (map of refinement name to value)
// - eval: evaluator for natives that need to evaluate blocks or expressions
func Call(fn *value.FunctionValue, posArgs []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if fn.Type != value.FuncNative {
		return value.NoneVal(), verror.NewInternalError(
			"Call() expects native function", [3]string{})
	}

	return fn.Native(posArgs, refValues, eval)
}
