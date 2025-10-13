// Package value - action value type implementation
package value

import "fmt"

// ActionValue represents a polymorphic function that dispatches to type-specific
// implementations based on the first argument's type.
//
// Actions dispatch at runtime by:
// 1. Evaluating first argument to determine its type
// 2. Looking up type in global TypeRegistry (in frame package)
// 3. Resolving action name in type frame
// 4. Invoking resolved function with arguments
//
// Feature: 004-dynamic-function-invocation
type ActionValue struct {
	Name      string      // Action name (e.g., "first", "append")
	ParamSpec []ParamSpec // Parameter specifications (using same as FunctionValue)
}

// NewAction creates a new action value.
// This is the constructor used during action registration.
func NewAction(name string, params []ParamSpec) *ActionValue {
	return &ActionValue{
		Name:      name,
		ParamSpec: params,
	}
}

// String returns a string representation for debugging.
func (a *ActionValue) String() string {
	return fmt.Sprintf("action[%s]", a.Name)
}

// Equals compares two action values for equality.
// Actions are equal if they have the same name (identity-based).
func (a *ActionValue) Equals(other *ActionValue) bool {
	if a == nil || other == nil {
		return a == other
	}
	return a.Name == other.Name
}

// Arity returns the number of required positional parameters.
// Same logic as FunctionValue.Arity() for consistency.
func (a *ActionValue) Arity() int {
	count := 0
	for _, p := range a.ParamSpec {
		if !p.Refinement && !p.Optional {
			count++
		}
	}
	return count
}
