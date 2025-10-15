package native

import (
	"github.com/marcin-radoszewski/viro/internal/core"
	"github.com/marcin-radoszewski/viro/internal/value"
	"github.com/marcin-radoszewski/viro/internal/verror"
)

// arityError returns a script error indicating a wrong number of arguments for a native.
func arityError(name string, expected, actual int) *verror.Error {
	return verror.NewScriptError(
		verror.ErrIDArgCount,
		[3]string{name, formatInt(int64(expected)), formatInt(int64(actual))},
	)
}

// typeError returns a script error indicating a type mismatch for a native argument.
func typeError(name, expectedType string, actual core.Value) error {
	return verror.NewScriptError(
		verror.ErrIDTypeMismatch,
		[3]string{name, expectedType, value.TypeToString(actual.GetType())},
	)
}

// mathTypeError is a convenience wrapper for integer expectations in math natives.
func mathTypeError(name string, actual core.Value) error {
	return typeError(name, "integer", actual)
}

// overflowError returns a math error for overflow scenarios.
func overflowError(op string) *verror.Error {
	return verror.NewMathError(verror.ErrIDOverflow, [3]string{op, "", ""})
}

// underflowError returns a math error for underflow scenarios.
func underflowError(op string) *verror.Error {
	return verror.NewMathError(verror.ErrIDUnderflow, [3]string{op, "", ""})
}

// invalidParamSpecError returns an error for invalid function parameter specifications.
func invalidParamSpecError(spec string) *verror.Error {
	return verror.NewScriptError(
		verror.ErrIDInvalidOperation,
		[3]string{"Invalid parameter specification: " + spec, "", ""},
	)
}

// duplicateParamError returns an error for duplicate parameter names in function definitions.
func duplicateParamError(name string) *verror.Error {
	return verror.NewScriptError(
		verror.ErrIDInvalidOperation,
		[3]string{"Duplicate parameter name: " + name, "", ""},
	)
}
