package value

import (
	"fmt"

	"github.com/marcin-radoszewski/viro/internal/docmodel"
)

// Evaluator is a minimal interface for evaluating Viro code.
// Used by native functions that need to evaluate arguments or blocks.
// The actual implementation is in the eval package to avoid circular dependencies.
//
// NOTE: This interface returns the standard error interface (not *verror.Error)
// to avoid an import cycle (verror imports value for context formatting).
// Native functions use native.Evaluator which returns *verror.Error directly.
// The adapter pattern in registry.go bridges between these two interfaces.
type Evaluator interface {
	Do_Blk(vals []Value) (Value, error)
	Do_Next(val Value) (Value, error)
}

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

// NewParamSpec creates a ParamSpec for a positional parameter.
// This is a convenience function for native function registration.
//
// Parameters:
//   - name: parameter name
//   - eval: if true, argument is evaluated before passing to function;
//     if false, argument is passed raw (unevaluated)
//
// The parameter is non-optional, accepts any type, and is not a refinement.
func NewParamSpec(name string, eval bool) ParamSpec {
	return ParamSpec{
		Name:       name,
		Type:       TypeNone, // any type by default
		Optional:   false,
		Refinement: false,
		TakesValue: false,
		Eval:       eval,
	}
}

// NewRefinementSpec creates a ParamSpec for a refinement (flag or option).
// This is a convenience function for native function registration.
//
// Parameters:
//   - name: refinement name (without -- prefix)
//   - takesValue: if true, refinement accepts a value (e.g., --title "text");
//     if false, refinement is a boolean flag (e.g., --verbose)
//
// Refinements are always optional and their arguments are always evaluated.
func NewRefinementSpec(name string, takesValue bool) ParamSpec {
	return ParamSpec{
		Name:       name,
		Type:       TypeNone, // any type by default
		Optional:   true,     // refinements are always optional
		Refinement: true,
		TakesValue: takesValue,
		Eval:       true, // refinement values always evaluated
	}
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
	Type   FunctionType                                                                  // Native or User
	Name   string                                                                        // function name (for error messages and debugging)
	Params []ParamSpec                                                                   // formal parameter specifications
	Body   *BlockValue                                                                   // function body (nil for natives)
	Native func(args []Value, refValues map[string]Value, eval Evaluator) (Value, error) // native implementation (nil for user functions)
	Parent int                                                                           // parent frame index for closures (-1 if none)
	Infix  bool                                                                          // true if function can be used as infix operator
	Doc    *docmodel.FuncDoc                                                             // dokumentacja funkcji użytkownika (nil jeśli brak)
}

// NewNativeFunction creates a native (built-in) function.
func NewNativeFunction(name string, params []ParamSpec, impl func([]Value, map[string]Value, Evaluator) (Value, error)) *FunctionValue {
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
// Dodano argument doc typu *docmodel.FuncDoc (może być nil)
func NewUserFunction(name string, params []ParamSpec, body *BlockValue, parentFrame int, doc *docmodel.FuncDoc) *FunctionValue {
	return &FunctionValue{
		Type:   FuncUser,
		Name:   name,
		Params: params,
		Body:   body,
		Native: nil,
		Parent: parentFrame,
		Doc:    doc,
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
