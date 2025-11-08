package verror

import (
	"fmt"
	"strings"
)

type Error struct {
	Category ErrorCategory
	Code     int
	ID       string
	Args     [3]string
	Near     string   // String representation of near context (values formatted)
	Where    []string // Call stack (most recent first)
	Message  string   // Formatted message
	File     string
	Line     int
	Column   int
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
	header := fmt.Sprintf("%s error (%d): %s", e.Category, e.Code, e.Message)
	if e.File != "" {
		header = fmt.Sprintf("%s:%d:%d %s", e.File, e.Line, e.Column, header)
	}
	sb.WriteString(header)
	sb.WriteString("\n")

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

func (e *Error) SetLocation(file string, line, column int) *Error {
	e.File = file
	e.Line = line
	e.Column = column
	return e
}

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

func formatMessage(id string, args [3]string) string {
	template, ok := messageTemplates[id]
	if !ok {
		template = "Error: %1 %2 %3" // fallback
	}

	if args[1] == "" && strings.Contains(template, "(%2)") {
		template = strings.ReplaceAll(template, " (%2)", "")
		template = strings.ReplaceAll(template, "(%2)", "")
	}

	msg := template
	msg = strings.ReplaceAll(msg, "%1", args[0])
	msg = strings.ReplaceAll(msg, "%2", args[1])
	msg = strings.ReplaceAll(msg, "%3", args[2])

	return msg
}

var messageTemplates = map[string]string{
	ErrIDUnexpectedEOF:       "Unexpected end of input",
	ErrIDUnclosedBlock:       "Unclosed block '[' - missing ']'",
	ErrIDUnclosedParen:       "Unclosed paren '(' - missing ')'",
	ErrIDInvalidLiteral:      "Invalid literal: %1",
	ErrIDInvalidSyntax:       "Invalid syntax: %1",
	ErrIDUnterminatedString:  "Unterminated string literal",
	ErrIDInvalidEscape:       "Invalid escape sequence: %1",
	ErrIDInvalidCharacter:    "Invalid character: %1",
	ErrIDUnexpectedClosing:   "Unexpected closing delimiter: %1",
	ErrIDInvalidNumberFormat: "Invalid number format: %1",
	ErrIDInvalidBinaryLength: "Binary literal must have even number of hex digits",
	ErrIDInvalidBinaryDigit:  "Invalid hex digit in binary literal: %1",
	ErrIDEmptyPath:           "Empty path",
	ErrIDEmptyPathSegment:    "Empty path segment",
	ErrIDPathLeadingNumber:   "Paths cannot start with numbers",
	ErrIDPathEvalBase:        "Path cannot start with eval segment: %1",

	ErrIDNoValue:          "No value for word: %1",
	ErrIDTypeMismatch:     "Type mismatch for '%1': expected %2, got %3",
	ErrIDInvalidOperation: "Invalid operation: %1",
	ErrIDArgCount:         "Wrong argument count for '%1': expected %2, got %3",
	ErrIDEmptySeries:      "Cannot get %1 of empty series",
	ErrIDOutOfBounds:      "Index %1 out of bounds (length: %2)",
	ErrIDNotImplemented:   "Feature not yet implemented: %1",
	ErrIDActionNoImpl:     "Action not implemented for type: %1",
	ErrIDInvalidToken:     "Invalid token object: %1",

	ErrIDInvalidPath:      "Invalid path (%2): %1",
	ErrIDNonePath:         "Cannot traverse path through none value",
	ErrIDNoSuchField:      "No such field '%1' in object",
	ErrIDPathTypeMismatch: "Type mismatch: path requires object or series type, got %1",
	ErrIDImmutableTarget:  "Cannot assign to immutable target: %1",
	ErrIDObjectFieldDup:   "Duplicate field '%1' in object",
	ErrIDReservedField:    "Field '%1' is reserved in object specifications",

	ErrIDDivByZero: "Division by zero",
	ErrIDOverflow:  "Integer overflow in operation: %1",
	ErrIDUnderflow: "Integer underflow in operation: %1",

	ErrIDSqrtNegative:     "Square root of negative number: %1",
	ErrIDLogDomain:        "Logarithm domain error: %1",
	ErrIDExpOverflow:      "Exponential overflow: %1",
	ErrIDDecimalPrecision: "Decimal precision overflow (max 34 digits): %1",
	ErrIDInvalidDecimal:   "Invalid decimal format: %1",
	ErrIDAsinDomain:       "asin domain error: %1 not in [-1, 1]",
	ErrIDAcosDomain:       "acos domain error: %1 not in [-1, 1]",

	ErrIDPortClosed:            "Port is closed: %1",
	ErrIDTLSVerificationFailed: "TLS certificate verification failed: %1",
	ErrIDSandboxViolation:      "Sandbox violation: path escapes sandbox root: %1",
	ErrIDTimeout:               "I/O timeout: %1",
	ErrIDConnectionRefused:     "Connection refused: %1",
	ErrIDUnknownScheme:         "Unknown port scheme: %1",

	ErrIDSpecUnsupported:   "spec-of: unsupported type %1",
	ErrIDNoBody:            "body-of: %1",
	ErrIDSourceUnsupported: "source: unsupported type %1",

	ErrIDStackOverflow:   "Stack overflow (maximum depth exceeded)",
	ErrIDOutOfMemory:     "Out of memory",
	ErrIDAssertionFailed: "Internal assertion failed: %1",
}

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
