// Package verror implements structured error handling for the Viro interpreter.
// Errors are categorized with codes, diagnostic context, and human-readable messages.
//
// Design per Constitution Principle V: Structured Errors
// - Category-based error codes (0-900 range)
// - Near context: expressions around error location
// - Where context: call stack trace
// - Message interpolation with arguments
package verror

// ErrorCategory defines error class constants.
// Categories organize errors by severity and origin.
//
// Per contracts/error-handling.md:
// - 0: Loop control (break, continue)
// - 100: Warnings (non-fatal)
// - 200: Syntax errors (parsing)
// - 300: Script errors (runtime)
// - 400: Math errors (arithmetic)
// - 500: Access errors (I/O, security)
// - 900: Internal errors (interpreter bugs, resource exhaustion)
type ErrorCategory uint16

const (
	ErrThrow    ErrorCategory = 0   // Loop control: break outside loop, etc.
	ErrNote     ErrorCategory = 100 // Warnings: non-fatal issues
	ErrSyntax   ErrorCategory = 200 // Syntax: parsing errors, malformed input
	ErrScript   ErrorCategory = 300 // Script: undefined words, type mismatches, invalid operations
	ErrMath     ErrorCategory = 400 // Math: division by zero, overflow, underflow
	ErrAccess   ErrorCategory = 500 // Access: file errors, permissions (future)
	ErrInternal ErrorCategory = 900 // Internal: stack overflow, out-of-memory, interpreter bugs
)

// String returns the category name for display.
func (c ErrorCategory) String() string {
	switch c {
	case ErrThrow:
		return "Throw"
	case ErrNote:
		return "Note"
	case ErrSyntax:
		return "Syntax"
	case ErrScript:
		return "Script"
	case ErrMath:
		return "Math"
	case ErrAccess:
		return "Access"
	case ErrInternal:
		return "Internal"
	default:
		return "Unknown"
	}
}

// Common error IDs (kebab-case identifiers for programmatic handling)
const (
	// Syntax errors (200)
	ErrIDUnexpectedEOF  = "unexpected-eof"
	ErrIDUnclosedBlock  = "unclosed-block"
	ErrIDUnclosedParen  = "unclosed-paren"
	ErrIDInvalidLiteral = "invalid-literal"
	ErrIDInvalidSyntax  = "invalid-syntax"

	// Script errors (300)
	ErrIDNoValue          = "no-value"
	ErrIDTypeMismatch     = "type-mismatch"
	ErrIDInvalidOperation = "invalid-operation"
	ErrIDArgCount         = "arg-count"
	ErrIDEmptySeries      = "empty-series"
	ErrIDOutOfBounds      = "out-of-bounds"
	ErrIDNotImplemented   = "not-implemented" // Feature 002: feature not yet implemented

	// Math errors (400)
	ErrIDDivByZero = "div-zero"
	ErrIDOverflow  = "overflow"
	ErrIDUnderflow = "underflow"

	// Feature 002: Decimal-specific math errors
	ErrIDSqrtNegative     = "sqrt-negative"     // sqrt of negative number
	ErrIDLogDomain        = "log-domain"        // log of zero or negative
	ErrIDExpOverflow      = "exp-overflow"      // exponential overflow
	ErrIDDecimalPrecision = "decimal-precision" // precision overflow (>34 digits)
	ErrIDInvalidDecimal   = "invalid-decimal"   // invalid decimal string format
	ErrIDAsinDomain       = "asin-domain"       // asin outside [-1, 1]
	ErrIDAcosDomain       = "acos-domain"       // acos outside [-1, 1]

	// Access errors (500) - Feature 002: Port I/O
	ErrIDPortClosed            = "port-closed"             // operation on closed port
	ErrIDTLSVerificationFailed = "tls-verification-failed" // HTTPS certificate validation failed
	ErrIDSandboxViolation      = "sandbox-violation"       // file path escapes sandbox root
	ErrIDTimeout               = "timeout"                 // I/O operation timeout
	ErrIDConnectionRefused     = "connection-refused"      // TCP/HTTP connection refused
	ErrIDUnknownScheme         = "unknown-port-scheme"     // unsupported port scheme

	// Internal errors (900)
	ErrIDStackOverflow   = "stack-overflow"
	ErrIDOutOfMemory     = "out-of-memory"
	ErrIDAssertionFailed = "assertion-failed"
)
