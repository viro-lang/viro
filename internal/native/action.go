// Package native - action creation utilities
package native

import (
	"github.com/marcin-radoszewski/viro/internal/core"
	"github.com/marcin-radoszewski/viro/internal/frame"
	"github.com/marcin-radoszewski/viro/internal/value"
)

// CreateAction creates a new action value with the given name and parameter specifications.
// This is the primary constructor used when registering action natives.
//
// Parameters:
//   - name: Action name (e.g., "first", "append")
//   - params: Parameter specifications (same format as FunctionValue)
//
// Returns an action value ready to be bound into the root frame.
//
// Feature: 004-dynamic-function-invocation
func CreateAction(name string, params []value.ParamSpec) value.Value {
	action := value.NewAction(name, params)
	return value.ActionVal(action)
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
