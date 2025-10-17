// Package native - action creation utilities
package native

import (
	"github.com/marcin-radoszewski/viro/internal/core"
	"github.com/marcin-radoszewski/viro/internal/frame"
	"github.com/marcin-radoszewski/viro/internal/value"
	"github.com/marcin-radoszewski/viro/internal/verror"
)

// CreateAction creates a new action as a regular function with dispatch logic.
// The action is a native function whose implementation performs type-based dispatch.
//
// Parameters:
//   - name: Action name (e.g., "first", "append")
//   - params: Parameter specifications (same format as FunctionValue)
//   - doc: Documentation for the action (can be nil)
//
// Returns a function value ready to be bound into the root frame.
//
// Feature: 004-dynamic-function-invocation
func CreateAction(name string, params []value.ParamSpec, doc *NativeDoc) value.Value {
	// Create dispatcher closure that captures the action name
	dispatcher := func(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
		// Validate we have at least one argument for dispatch
		if len(args) == 0 {
			return value.NoneVal(), verror.NewScriptError(
				verror.ErrIDArgCount,
				[3]string{name, "at least 1", "0"},
			)
		}

		// Get the type of the first argument for dispatch
		firstArg := args[0]
		firstArgType := firstArg.GetType()

		// Look up type frame in global TypeRegistry
		typeFrame, found := frame.GetTypeFrame(firstArgType)
		if !found {
			// Type has no type frame - no implementations registered
			return value.NoneVal(), verror.NewScriptError(
				verror.ErrIDActionNoImpl,
				[3]string{name, value.TypeToString(firstArgType), ""},
			)
		}

		// Look up action name in type frame
		impl, found := typeFrame.Get(name)
		if !found {
			// Type frame exists but doesn't have this action
			return value.NoneVal(), verror.NewScriptError(
				verror.ErrIDActionNoImpl,
				[3]string{name, value.TypeToString(firstArgType), ""},
			)
		}

		// Extract function from value
		fn, ok := value.AsFunction(impl)
		if !ok {
			// Internal error: type frame contains non-function value
			return value.NoneVal(), verror.NewInternalError(
				"action-frame-corrupt",
				[3]string{name, value.TypeToString(firstArgType), "type frame binding is not a function"},
			)
		}

		// Call the type-specific function directly
		return fn.Native(args, refValues, eval)
	}

	// Create regular native function with dispatcher as implementation
	actionFunc := value.NewNativeFunction(name, params, dispatcher)
	actionFunc.Doc = doc
	return value.FuncVal(actionFunc)
}

// RegisterActionImpl registers a type-specific implementation for an action.
// This binds a function into the type frame for the specified value type.
//
// Parameters:
//   - typ: The value type (e.g., TypeBlock, TypeString)
//   - actionName: The action name (e.g., "first")
//   - fn: The type-specific function implementation
//
// The function is bound into the type frame so it can be resolved during dispatch.
//
// Feature: 004-dynamic-function-invocation
func RegisterActionImpl(typ core.ValueType, actionName string, fn *value.FunctionValue) {
	// Get the type frame for this value type
	typeFrame, found := frame.GetTypeFrame(typ)
	if !found {
		// Internal error: type frame should exist (created by InitTypeFrames)
		panic("RegisterActionImpl: type frame not found for " + value.TypeToString(typ))
	}

	// Bind the function into the type frame
	typeFrame.Bind(actionName, value.FuncVal(fn))
}
