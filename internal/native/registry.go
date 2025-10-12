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
// All native functions are unified under the FunctionValue type and stored in the Registry.
package native

import (
	"github.com/marcin-radoszewski/viro/internal/frame"
	"github.com/marcin-radoszewski/viro/internal/value"
	"github.com/marcin-radoszewski/viro/internal/verror"
)

// Evaluator interface for natives that need to evaluate code.
//
// NOTE: This is intentionally a separate interface from value.Evaluator even though
// they have identical method names. The difference is in return types:
//   - native.Evaluator returns *verror.Error (for native function implementations)
//   - value.Evaluator returns error (for FunctionValue.Native field)
//
// This separation exists because of an import cycle constraint:
//   - verror imports value (for error context formatting)
//   - value cannot import verror without creating a cycle
//
// Native function implementations (When, If, Set, Get, etc.) use this interface
// to work directly with *verror.Error for type safety. The registry code provides
// a simple adapter when registering functions to bridge between the two interfaces.
type Evaluator interface {
	Do_Blk(vals []value.Value) (value.Value, *verror.Error)
	Do_Next(val value.Value) (value.Value, *verror.Error)
}

// Registry holds all registered native functions as FunctionValue instances.
// DEPRECATED: This is kept temporarily for backward compatibility with help system.
// Native functions are now stored in the root frame. This will be removed in a future version.
var Registry = make(map[string]*value.FunctionValue)

// Lookup finds a native function by name in the Registry.
// Returns the function value and true if found, nil and false otherwise.
func Lookup(name string) (*value.FunctionValue, bool) {
	fn, ok := Registry[name]
	return fn, ok
}

// Call invokes a native function (FunctionValue) with the given arguments, refinements, and evaluator.
// The evaluator is always passed to the native function (even if the function doesn't use it).
//
// Arguments:
// - posArgs: positional arguments (already evaluated or raw according to ParamSpec.Eval)
// - refValues: refinement values (map of refinement name to value)
// - eval: evaluator for natives that need to evaluate blocks or expressions
func Call(fn *value.FunctionValue, posArgs []value.Value, refValues map[string]value.Value, eval Evaluator) (value.Value, error) {
	if fn.Type != value.FuncNative {
		return value.NoneVal(), verror.NewInternalError(
			"Call() expects native function", [3]string{})
	}

	// Create an adapter to bridge native.Evaluator (returns *verror.Error)
	// to value.Evaluator (returns error)
	adapter := evaluatorAdapter{eval}
	return fn.Native(posArgs, refValues, adapter)
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

// Implement frameManager interface by delegating to the underlying evaluator if it supports these methods
func (a *nativeEvaluatorAdapter) RegisterFrame(f *frame.Frame) int {
	// Try to cast the underlying evaluator to frameManager
	if mgr, ok := a.eval.(interface {
		RegisterFrame(*frame.Frame) int
	}); ok {
		return mgr.RegisterFrame(f)
	}
	// Fallback: return -1 to indicate failure
	return -1
}

func (a *nativeEvaluatorAdapter) GetFrameByIndex(idx int) *frame.Frame {
	// Try to cast the underlying evaluator to frameManager
	if mgr, ok := a.eval.(interface {
		GetFrameByIndex(int) *frame.Frame
	}); ok {
		return mgr.GetFrameByIndex(idx)
	}
	// Fallback: return nil
	return nil
}

func (a *nativeEvaluatorAdapter) MarkFrameCaptured(idx int) {
	// Try to cast the underlying evaluator to frameManager
	if mgr, ok := a.eval.(interface {
		MarkFrameCaptured(int)
	}); ok {
		mgr.MarkFrameCaptured(idx)
	}
}

func (a *nativeEvaluatorAdapter) PushFrameContext(f *frame.Frame) int {
	// Try to cast the underlying evaluator to frameManager
	if mgr, ok := a.eval.(interface {
		PushFrameContext(*frame.Frame) int
	}); ok {
		return mgr.PushFrameContext(f)
	}
	// Fallback: return -1 to indicate failure
	return -1
}

func (a *nativeEvaluatorAdapter) PopFrameContext() {
	// Try to cast the underlying evaluator to frameManager
	if mgr, ok := a.eval.(interface {
		PopFrameContext()
	}); ok {
		mgr.PopFrameContext()
	}
}
