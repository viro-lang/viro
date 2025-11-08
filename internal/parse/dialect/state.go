// Package dialect implements the Viro parse dialect - a pattern matching DSL
// for strings and blocks, compatible with Rebol's parse semantics.
package dialect

import (
	"github.com/marcin-radoszewski/viro/internal/core"
)

// MatchResult represents the outcome of a parse operation.
type MatchResult int

const (
	MatchSuccess MatchResult = iota // Parse succeeded
	MatchFailure                    // Parse failed
	MatchError                      // Parse error (invalid rule, etc.)
)

// ParseOptions configures parse behavior via refinements.
type ParseOptions struct {
	CaseSensitive bool // --case: case-sensitive string matching
	MatchAll      bool // --all: match entire input (no remainder)
	Part          int  // --part N: only parse first N elements
	Any           bool // --any: allow partial matches
}

// DefaultOptions returns default parse options (case-insensitive, match-all).
func DefaultOptions() ParseOptions {
	return ParseOptions{
		CaseSensitive: false,
		MatchAll:      true,
		Part:          -1, // -1 means no limit
		Any:           false,
	}
}

// ParseState tracks the current state during parsing.
type ParseState struct {
	input    core.Value    // The input series (string or block)
	position int           // Current position in input
	options  ParseOptions  // Parse options
	captures map[string]core.Value // Captured values (word -> value)
	marks    map[string]int         // Named positions (word: -> position)
}

// NewParseState creates a new parse state for the given input.
func NewParseState(input core.Value, options ParseOptions) *ParseState {
	return &ParseState{
		input:    input,
		position: 0,
		options:  options,
		captures: make(map[string]core.Value),
		marks:    make(map[string]int),
	}
}

// GetPosition returns the current position.
func (ps *ParseState) Position() int {
	return ps.position
}

// SetPosition sets the current position.
func (ps *ParseState) SetPosition(pos int) {
	ps.position = pos
}

// AtEnd returns true if at end of input.
func (ps *ParseState) AtEnd() bool {
	// This will be checked via cursor in Engine
	return false
}

// Remaining returns the number of elements remaining.
func (ps *ParseState) Remaining() int {
	// This will be checked via cursor in Engine
	return 0
}

// GetCapture retrieves a captured value by name.
func (ps *ParseState) GetCapture(name string) (core.Value, bool) {
	val, ok := ps.captures[name]
	return val, ok
}

// SetCapture stores a captured value.
func (ps *ParseState) SetCapture(name string, value core.Value) {
	ps.captures[name] = value
}

// GetMark retrieves a marked position by name.
func (ps *ParseState) GetMark(name string) (int, bool) {
	pos, ok := ps.marks[name]
	return pos, ok
}

// SetMark stores a marked position.
func (ps *ParseState) SetMark(name string, pos int) {
	ps.marks[name] = pos
}
