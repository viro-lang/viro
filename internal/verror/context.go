package verror

// Context capture functions for error diagnostics.
// These will be implemented once we have the evaluator and value packages integrated.

// CaptureNear creates a string representation of the near context.
// Shows expressions around error location (3 before, current, 3 after).
//
// Implementation deferred until evaluator is available.
// For now, returns empty string.
func CaptureNear(values interface{}, index int) string {
	// TODO: Implement once value package is integrated
	// Should show: value[index-3] value[index-2] value[index-1] >>> value[index] <<< value[index+1] value[index+2] value[index+3]
	return ""
}

// CaptureWhere creates a call stack trace from frame chain.
// Shows function names from most recent to oldest.
//
// Implementation deferred until stack/frame packages are available.
// For now, returns empty slice.
func CaptureWhere(stack interface{}, currentFrame int) []string {
	// TODO: Implement once stack/frame packages are integrated
	// Should walk frame chain via Parent pointers, collecting function names
	return []string{}
}
