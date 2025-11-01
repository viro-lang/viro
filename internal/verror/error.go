// Package verror provides structured error handling for the Viro interpreter.
//
// Errors are categorized into ranges (0, 100, 200, 300, 400, 500, 900) to
// distinguish between syntax errors, script errors, math errors, access errors,
// user errors, and internal errors.
//
// Each error includes:
//   - Category: Error type (syntax, script, math, access, user, internal)
//   - Code: Specific error code within category
//   - ID: String identifier for the error
//   - Args: Up to 3 arguments for message interpolation
//   - Near: Expression context showing where the error occurred
//   - Where: Call stack location
//   - Message: Formatted error message
//
// Error factory functions (NewSyntaxError, NewScriptError, etc.) provide
// convenient error creation with proper categorization.
package verror

import (
	"fmt"
	"strings"
)

// Error represents a structured interpreter error with diagnostic context.
//
// Design per contracts/error-handling.md:
// - Category: Error class (0-900)
// - Code: Specific error within category
// - ID: Symbolic identifier for programmatic handling
// - Args: Up to 3 arguments for message interpolation (%1, %2, %3)
// - Near: Expression window showing error location (3 before, 3 after)
// - Where: Call stack trace (function names)
// - Message: Formatted human-readable error message
type Error struct {
	Category ErrorCategory
	Code     int
	ID       string
	Args     [3]string
	Near     string   // String representation of near context (values formatted)
	Where    []string // Call stack (most recent first)
	Message  string   // Formatted message
}

// NewError creates an error with given category, ID, and arguments.
// Message is generated automatically from ID and Args.
func NewError(category ErrorCategory, id string, args [3]string) *Error {
	return &Error{
		Category: category,
		Code:     int(category), // Base code from category
		ID:       id,
		Args:     args,
		Near:     "",
		Where:    []string{},
		Message:  formatMessage(id, args),
	}
}

// Error implements the Go error interface.
func (e *Error) Error() string {
	var sb strings.Builder

	// Main error line
	sb.WriteString(fmt.Sprintf("%s error (%d): %s\n", e.Category, e.Code, e.Message))

	// Near context (if available)
	if e.Near != "" {
		sb.WriteString(fmt.Sprintf("Near: %s\n", e.Near))
	}

	// Where context (if available)
	if len(e.Where) > 0 {
		sb.WriteString(fmt.Sprintf("Where: %s\n", strings.Join(e.Where, " <- ")))
	}

	return sb.String()
}

// SetNear adds near context (expression window around error).
func (e *Error) SetNear(near string) *Error {
	e.Near = near
	return e
}

// SetWhere adds call stack context.
func (e *Error) SetWhere(where []string) *Error {
	e.Where = where
	return e
}

// Factory functions for each error category

// NewSyntaxError creates a syntax error (parsing).
func NewSyntaxError(id string, args [3]string) *Error {
	return NewError(ErrSyntax, id, args)
}

// NewScriptError creates a script error (runtime).
func NewScriptError(id string, args [3]string) *Error {
	return NewError(ErrScript, id, args)
}

// NewMathError creates a math error (arithmetic).
func NewMathError(id string, args [3]string) *Error {
	return NewError(ErrMath, id, args)
}

// NewAccessError creates an access error (I/O).
func NewAccessError(id string, args [3]string) *Error {
	return NewError(ErrAccess, id, args)
}

// NewInternalError creates an internal error (interpreter bug).
func NewInternalError(id string, args [3]string) *Error {
	return NewError(ErrInternal, id, args)
}

// formatMessage generates human-readable error message from ID and args.
// Uses simple template substitution: %1, %2, %3.
func formatMessage(id string, args [3]string) string {
	template, ok := messageTemplates[id]
	if !ok {
		template = "Error: %1 %2 %3" // fallback
	}

	msg := template
	msg = strings.ReplaceAll(msg, "%1", args[0])
	msg = strings.ReplaceAll(msg, "%2", args[1])
	msg = strings.ReplaceAll(msg, "%3", args[2])

	return msg
}

// messageTemplates maps error IDs to message templates.
// Templates use %1, %2, %3 for argument interpolation.
var messageTemplates = map[string]string{
	// Syntax errors
	ErrIDUnexpectedEOF:  "Unexpected end of input",
	ErrIDUnclosedBlock:  "Unclosed block '[' - missing ']'",
	ErrIDUnclosedParen:  "Unclosed paren '(' - missing ')'",
	ErrIDInvalidLiteral: "Invalid literal: %1",
	ErrIDInvalidSyntax:  "Invalid syntax: %1",

	// Script errors
	ErrIDNoValue:          "No value for word: %1",
	ErrIDTypeMismatch:     "Type mismatch for '%1': expected %2, got %3",
	ErrIDInvalidOperation: "Invalid operation: %1",
	ErrIDArgCount:         "Wrong argument count for '%1': expected %2, got %3",
	ErrIDEmptySeries:      "Cannot get %1 of empty series",
	ErrIDOutOfBounds:      "Index %1 out of bounds (length: %2)",
	ErrIDNotImplemented:   "Feature not yet implemented: %1",

	// Feature 002: Path evaluation errors (T091)
	ErrIDInvalidPath:      "Invalid path: %1",
	ErrIDNonePath:         "Cannot traverse path through none value",
	ErrIDNoSuchField:      "No such field '%1' in object",
	ErrIDIndexOutOfRange:  "%1", // full message in arg
	ErrIDPathTypeMismatch: "Type mismatch: path requires object or series type, got %1",
	ErrIDImmutableTarget:  "Cannot assign to immutable target: %1",
	ErrIDObjectFieldDup:   "Duplicate field '%1' in object",
	ErrIDReservedField:    "Field '%1' is reserved in object specifications",

	// Math errors
	ErrIDDivByZero: "Division by zero",
	ErrIDOverflow:  "Integer overflow in operation: %1",
	ErrIDUnderflow: "Integer underflow in operation: %1",

	// Feature 002: Decimal-specific math errors
	ErrIDSqrtNegative:     "Square root of negative number: %1",
	ErrIDLogDomain:        "Logarithm domain error: %1",
	ErrIDExpOverflow:      "Exponential overflow: %1",
	ErrIDDecimalPrecision: "Decimal precision overflow (max 34 digits): %1",
	ErrIDInvalidDecimal:   "Invalid decimal format: %1",
	ErrIDAsinDomain:       "asin domain error: %1 not in [-1, 1]",
	ErrIDAcosDomain:       "acos domain error: %1 not in [-1, 1]",

	// Feature 002: Access errors (Port I/O)
	ErrIDPortClosed:            "Port is closed: %1",
	ErrIDTLSVerificationFailed: "TLS certificate verification failed: %1",
	ErrIDSandboxViolation:      "Sandbox violation: path escapes sandbox root: %1",
	ErrIDTimeout:               "I/O timeout: %1",
	ErrIDConnectionRefused:     "Connection refused: %1",
	ErrIDUnknownScheme:         "Unknown port scheme: %1",

	// Feature 002: Reflection errors (T162)
	ErrIDSpecUnsupported:   "spec-of: unsupported type %1",
	ErrIDNoBody:            "body-of: %1",
	ErrIDSourceUnsupported: "source: unsupported type %1",

	// Internal errors
	ErrIDStackOverflow:   "Stack overflow (maximum depth exceeded)",
	ErrIDOutOfMemory:     "Out of memory",
	ErrIDAssertionFailed: "Internal assertion failed: %1",
}

// ToExitCode converts an error category to an appropriate exit code.
// This centralizes exit code mapping logic across the application.
func ToExitCode(category ErrorCategory) int {
	switch category {
	case ErrSyntax:
		return 2 // ExitSyntax
	case ErrAccess:
		return 3 // ExitAccess
	case ErrInternal:
		return 70 // ExitInternal
	default:
		return 1 // ExitError
	}
}
