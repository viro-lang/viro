package verror

import (
	"fmt"
	"strings"

	"github.com/marcin-radoszewski/viro/internal/core"
	"github.com/marcin-radoszewski/viro/internal/value"
)

// Context capture functions for error diagnostics.
// Provides Near (expression window) and Where (call stack) context for errors.

// CaptureNear creates a string representation of the near context.
// Shows expressions around error location (3 before, current, 3 after).
//
// Per data-model.md §8 and error-handling.md:
// - Window size: 3 expressions before, current (marked), 3 after
// - Boundary handling: clamps to available elements
// - Format: "value1 value2 value3 >>> ERROR_HERE <<< value4 value5 value6"
// - For decimal values, includes magnitude and scale metadata
func CaptureNear(values []core.Value, index int) string {
	if len(values) == 0 || index < 0 || index >= len(values) {
		return ""
	}

	// Calculate window bounds (3 before, 3 after)
	start := index - 3
	if start < 0 {
		start = 0
	}
	end := index + 4 // +4 because end is exclusive: [start, end)
	if end > len(values) {
		end = len(values)
	}

	var parts []string

	// Values before error
	for i := start; i < index; i++ {
		parts = append(parts, formatValueWithMetadata(values[i]))
	}

	// Error location marker
	parts = append(parts, ">>>")
	parts = append(parts, formatValueWithMetadata(values[index]))
	parts = append(parts, "<<<")

	// Values after error
	for i := index + 1; i < end; i++ {
		parts = append(parts, formatValueWithMetadata(values[i]))
	}

	return strings.Join(parts, " ")
}

// formatValueWithMetadata formats a value with type-specific metadata.
// For decimal values, includes scale information for better diagnostics.
func formatValueWithMetadata(v core.Value) string {
	if v.GetType() == value.TypeDecimal {
		if dec, ok := value.AsDecimal(v); ok && dec != nil {
			// Include scale metadata for decimal diagnostics
			return fmt.Sprintf("%s[scale:%d]", dec.Mold(), dec.Scale)
		}
	}
	return v.Mold()
}

// CaptureWhere creates a call stack trace from frame chain.
// Shows function names from most recent to oldest.
//
// Per data-model.md §6 and error-handling.md:
// - Walks frame chain via Parent indices
// - Collects function names from function metadata
// - Returns slice with most recent call first
// - Empty slice if no frames (top-level evaluation)
//
// Note: This function takes a frame retrieval function to avoid
// circular dependencies between packages.
type FrameInfo struct {
	FunctionName string
	Parent       int
}

// CaptureWhere builds call stack from frame chain.
// frameGetter should return (FrameInfo, true) for valid frame index, (empty, false) otherwise.
func CaptureWhere(currentFrameIdx int, frameGetter func(int) (FrameInfo, bool)) []string {
	if currentFrameIdx == -1 {
		return []string{} // top-level, no call stack
	}

	var callStack []string
	frameIdx := currentFrameIdx

	// Walk frame chain via Parent pointers
	// Limit to reasonable depth (100) to prevent infinite loops
	maxDepth := 100
	for i := 0; i < maxDepth && frameIdx != -1; i++ {
		frameInfo, ok := frameGetter(frameIdx)
		if !ok {
			break // invalid frame index
		}

		// Add function name to call stack
		if frameInfo.FunctionName != "" {
			callStack = append(callStack, frameInfo.FunctionName)
		}

		// Move to parent frame
		frameIdx = frameInfo.Parent
	}

	return callStack
}

// FormatErrorWithContext formats an error message with Near and Where context.
// Used by REPL to display errors with diagnostic information.
//
// Format:
//
//	** Category Error (ID)
//	Message with interpolated arguments
//	Near: expression window showing error location
//	Where: call stack from most recent to oldest
func FormatErrorWithContext(err *Error) string {
	var parts []string

	// Error header: ** Category Error (ID)
	category := "Error"
	switch err.Category {
	case 0:
		category = "Throw"
	case 100:
		category = "Note"
	case 200:
		category = "Syntax"
	case 300:
		category = "Script"
	case 400:
		category = "Math"
	case 500:
		category = "Access"
	case 900:
		category = "Internal"
	}
	parts = append(parts, fmt.Sprintf("** %s Error (%s)", category, err.ID))

	// Error message
	parts = append(parts, err.Message)

	// Near context (if available)
	if err.Near != "" {
		parts = append(parts, fmt.Sprintf("Near: %s", err.Near))
	}

	// Where context (if available)
	if len(err.Where) > 0 {
		whereStr := strings.Join(err.Where, " > ")
		parts = append(parts, fmt.Sprintf("Where: %s", whereStr))
	}

	return strings.Join(parts, "\n")
}
