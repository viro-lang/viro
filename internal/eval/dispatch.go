// Package eval - action dispatch logic
package eval

import (
	"github.com/marcin-radoszewski/viro/internal/frame"
	"github.com/marcin-radoszewski/viro/internal/native"
	"github.com/marcin-radoszewski/viro/internal/value"
	"github.com/marcin-radoszewski/viro/internal/verror"
)

// DispatchAction performs type-based dispatch for action values.
//
// Dispatch flow:
// 1. Extract type from first argument
// 2. Look up type in global TypeRegistry
// 3. Resolve action name in type frame
// 4. Invoke resolved function with arguments
//
// Returns the result of the type-specific function or an error.
//
// Feature: 004-dynamic-function-invocation
func (e *Evaluator) DispatchAction(action *value.ActionValue, posArgs []value.Value, refValues map[string]value.Value) (value.Value, *verror.Error) {
	// Validate we have at least one argument (the dispatch argument)
	if len(posArgs) == 0 {
		return value.NoneVal(), verror.NewScriptError(
			verror.ErrIDArgCount,
			[3]string{action.Name, "at least 1", "0"},
		)
	}

	// Get the type of the first argument (dispatch type)
	firstArgType := posArgs[0].Type

	// Look up type frame in global TypeRegistry
	typeFrame, found := frame.GetTypeFrame(firstArgType)
	if !found {
		// Type has no type frame - no implementations registered
		return value.NoneVal(), verror.NewScriptError(
			verror.ErrIDActionNoImpl,
			[3]string{action.Name, firstArgType.String(), ""},
		)
	}

	// Look up action name in type frame
	funcVal, found := typeFrame.Get(action.Name)
	if !found {
		// Type frame exists but doesn't have this action
		return value.NoneVal(), verror.NewScriptError(
			verror.ErrIDActionNoImpl,
			[3]string{action.Name, firstArgType.String(), ""},
		)
	}

	// Extract function from value
	fn, ok := funcVal.AsFunction()
	if !ok {
		// Internal error: type frame contains non-function value
		return value.NoneVal(), verror.NewInternalError(
			"action-frame-corrupt",
			[3]string{action.Name, firstArgType.String(), "type frame binding is not a function"},
		)
	}

	// Invoke the type-specific function with collected arguments
	// The function is always native (we registered it that way)
	if fn.Type != value.FuncNative {
		return value.NoneVal(), verror.NewInternalError(
			"action-frame-corrupt",
			[3]string{action.Name, firstArgType.String(), "type frame contains non-native function"},
		)
	}

	// Call the native function directly with arguments and refinements
	result, callErr := native.Call(fn, posArgs, refValues, e)
	if callErr != nil {
		// Convert error interface back to *verror.Error
		var vErr *verror.Error
		if ve, ok := callErr.(*verror.Error); ok {
			vErr = ve
		} else {
			vErr = verror.NewInternalError(callErr.Error(), [3]string{})
		}
		return value.NoneVal(), vErr
	}

	return result, nil
}
