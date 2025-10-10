package value

import "fmt"

// FunctionType distinguishes native (built-in) from user-defined functions.
type FunctionType uint8

const (
	FuncNative FunctionType = iota // Built-in function (implemented in Go)
	FuncUser                       // User-defined function (via fn native)
)

// ParamSpec defines a function parameter specification.
// Supports positional parameters and refinements (flags and value options).
//
// Design per contracts/function.md:
// - Positional: regular arguments (name, required/optional)
// - Flag refinement: --verbose (boolean, true if present, false otherwise)
// - Value refinement: --title [] (accepts value, none if not provided)
type ParamSpec struct {
	Name       string    // parameter name (without -- prefix for refinements)
	Type       ValueType // expected type (TypeNone = any type accepted)
	Optional   bool      // true if parameter can be omitted
	Refinement bool      // true if this is a refinement (--flag or --option)
	TakesValue bool      // for refinements: true if accepts value, false if boolean flag
	Eval       bool      // NEW: if true, argument is evaluated; if false, passed raw
}

// FunctionValue represents an executable function.
//
// Native functions:
// - Type = FuncNative
// - Native field set to Go function pointer
// - Body is nil
//
// User-defined functions:
// - Type = FuncUser
// - Body field set to block containing function body
// - Native is nil
//
// Design per data-model.md §5:
// - Functions are immutable after creation
// - Local-by-default scoping: all words in body are local by default
// - Closures capture parent frame via Parent field
type FunctionValue struct {
	Type   FunctionType                                        // Native or User
	Name   string                                              // function name (for error messages and debugging)
	Params []ParamSpec                                         // formal parameter specifications
	Body   *BlockValue                                         // function body (nil for natives)
	Native func(args []Value, eval interface{}) (Value, error) // native implementation (nil for user functions)
	Parent int                                                 // parent frame index for closures (-1 if none)
	Infix  bool                                                // true if function can be used as infix operator
}

// NewNativeFunction creates a native (built-in) function.
func NewNativeFunction(name string, params []ParamSpec, impl func([]Value, interface{}) (Value, error)) *FunctionValue {
	return &FunctionValue{
		Type:   FuncNative,
		Name:   name,
		Params: params,
		Body:   nil,
		Native: impl,
		Parent: -1,
	}
}

// NewUserFunction creates a user-defined function.
func NewUserFunction(name string, params []ParamSpec, body *BlockValue, parentFrame int) *FunctionValue {
	return &FunctionValue{
		Type:   FuncUser,
		Name:   name,
		Params: params,
		Body:   body,
		Native: nil,
		Parent: parentFrame,
	}
}

// String returns a string representation for debugging.
func (f *FunctionValue) String() string {
	if f.Type == FuncNative {
		return fmt.Sprintf("native[%s]", f.Name)
	}
	return fmt.Sprintf("function[%s]", f.Name)
}

// Arity returns the number of required positional parameters.
func (f *FunctionValue) Arity() int {
	count := 0
	for _, p := range f.Params {
		if !p.Refinement && !p.Optional {
			count++
		}
	}
	return count
}

// HasRefinement checks if function has a refinement with given name.
func (f *FunctionValue) HasRefinement(name string) bool {
	for _, p := range f.Params {
		if p.Refinement && p.Name == name {
			return true
		}
	}
	return false
}

// GetRefinement returns the ParamSpec for a refinement (nil if not found).
func (f *FunctionValue) GetRefinement(name string) *ParamSpec {
	for i := range f.Params {
		if f.Params[i].Refinement && f.Params[i].Name == name {
			return &f.Params[i]
		}
	}
	return nil
}
