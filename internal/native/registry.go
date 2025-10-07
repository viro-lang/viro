// Package native provides a registry of all native functions.
//
// Natives are built-in functions implemented in Go.
// The registry maps word symbols to native implementations.
package native

import (
	"github.com/marcin-radoszewski/viro/internal/value"
	"github.com/marcin-radoszewski/viro/internal/verror"
)

// NativeFunc is the signature for native function implementations.
type NativeFunc func([]value.Value) (value.Value, *verror.Error)

// Registry holds all registered native functions.
var Registry = make(map[string]NativeFunc)

func init() {
	// Register math natives
	Registry["+"] = Add
	Registry["-"] = Subtract
	Registry["*"] = Multiply
	Registry["/"] = Divide
}

// Lookup finds a native function by name.
// Returns the function and true if found, nil and false otherwise.
func Lookup(name string) (NativeFunc, bool) {
	fn, ok := Registry[name]
	return fn, ok
}
