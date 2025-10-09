package native

import (
	"strconv"

	"github.com/marcin-radoszewski/viro/internal/value"
	"github.com/marcin-radoszewski/viro/internal/verror"
)

// arityError returns a script error indicating a wrong number of arguments for a native.
func arityError(name string, expected, actual int) *verror.Error {
	return verror.NewScriptError(
		verror.ErrIDArgCount,
		[3]string{name, strconv.Itoa(expected), strconv.Itoa(actual)},
	)
}

// typeError returns a script error indicating a type mismatch for a native argument.
func typeError(name, expectedType string, actual value.Value) *verror.Error {
	return verror.NewScriptError(
		verror.ErrIDTypeMismatch,
		[3]string{name, expectedType, actual.Type.String()},
	)
}

// mathTypeError is a convenience wrapper for integer expectations in math natives.
func mathTypeError(name string, actual value.Value) *verror.Error {
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
