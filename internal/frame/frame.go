// Package frame implements the variable binding system for Viro.
// Frames map word symbols to values, supporting local-by-default scoping.
// Package frame implements variable binding contexts for the Viro interpreter.
//
// Frames provide lexical scoping by maintaining word-to-value bindings.
// Each function call creates a new frame with its own bindings, and frames
// are linked via parent pointers to support closure semantics.
//
// Frame types:
//   - FrameFunctionArgs: Function parameter bindings
//   - FrameClosure: Closure variable capture
//
// Operations:
//   - Bind: Create new word-to-value binding
//   - Get: Lookup value, traversing parent chain
//   - Set: Update existing binding
//   - HasWord: Check if word is bound
//
// Viro uses local-by-default scoping (differs from REBOL's global-by-default).
package frame

import (
	"github.com/marcin-radoszewski/viro/internal/value"
)

// FrameType distinguishes different kinds of frames.
type FrameType uint8

const (
	FrameFunctionArgs FrameType = iota // Function call frame (arguments + locals)
	FrameClosure                       // Closure captured environment
	// FrameObject, FrameModule deferred to later phases
)

// Frame represents a variable binding context.
// Maps word symbols (strings) to values.
//
// Design per data-model.md §6:
// - Parallel arrays: Words and Values
// - Parent index for lexical scoping (-1 = no parent)
// - Local-by-default: words assigned in frame are local
//
// Constitution Principle IV: Index-based references
// - Parent is an integer index, not a pointer
// - Safe across stack expansion
type Frame struct {
	Type   FrameType     // Frame category
	Words  []string      // Symbol names (parallel to Values)
	Values []value.Value // Bound values (parallel to Words)
	Parent int           // Index of parent frame for closures (-1 if none)
	Name   string        // Optional function or context name for diagnostics
}

// NewFrame creates an empty frame.
func NewFrame(frameType FrameType, parent int) *Frame {
	return &Frame{
		Type:   frameType,
		Words:  []string{},
		Values: []value.Value{},
		Parent: parent,
		Name:   "",
	}
}

// NewFrameWithCapacity creates a frame with pre-allocated capacity.
// Useful for function frames where parameter count is known.
func NewFrameWithCapacity(frameType FrameType, parent int, capacity int) *Frame {
	return &Frame{
		Type:   frameType,
		Words:  make([]string, 0, capacity),
		Values: make([]value.Value, 0, capacity),
		Parent: parent,
		Name:   "",
	}
}

// Bind adds or updates a word binding in this frame.
// Local-by-default: creates new binding if word doesn't exist.
//
// Per data-model.md: This is the core of local-by-default scoping.
// Assignment in a function creates a local variable, NOT a global.
func (f *Frame) Bind(symbol string, val value.Value) {
	// Check if word already exists in this frame
	for i, w := range f.Words {
		if w == symbol {
			// Update existing binding
			f.Values[i] = val
			return
		}
	}

	// Add new binding (local-by-default)
	f.Words = append(f.Words, symbol)
	f.Values = append(f.Values, val)
}

// Get retrieves the value bound to a symbol in this frame.
// Returns (value, true) if found, (NoneVal, false) if not.
//
// LOCAL LOOKUP ONLY - does NOT search parent frame.
// Evaluator is responsible for walking frame chain if needed.
func (f *Frame) Get(symbol string) (value.Value, bool) {
	for i, w := range f.Words {
		if w == symbol {
			return f.Values[i], true
		}
	}
	return value.NoneVal(), false
}

// Set updates an existing binding in this frame.
// Returns true if word was found and updated, false if not found.
// Does NOT create new binding (use Bind for that).
func (f *Frame) Set(symbol string, val value.Value) bool {
	for i, w := range f.Words {
		if w == symbol {
			f.Values[i] = val
			return true
		}
	}
	return false
}

// HasWord checks if a symbol is bound in this frame.
func (f *Frame) HasWord(symbol string) bool {
	for _, w := range f.Words {
		if w == symbol {
			return true
		}
	}
	return false
}

// Count returns the number of bindings in this frame.
func (f *Frame) Count() int {
	return len(f.Words)
}

// GetAll returns all bindings as (symbol, value) pairs.
// Useful for debugging and inspection.
func (f *Frame) GetAll() []Binding {
	bindings := make([]Binding, len(f.Words))
	for i := range f.Words {
		bindings[i] = Binding{
			Symbol: f.Words[i],
			Value:  f.Values[i],
		}
	}
	return bindings
}

// Binding represents a word-to-value binding.
type Binding struct {
	Symbol string
	Value  value.Value
}

// Clone creates a shallow copy of the frame.
// Words and values are copied, but value contents are shared.
// Used for closure capture.
func (f *Frame) Clone() *Frame {
	wordsCopy := make([]string, len(f.Words))
	valuesCopy := make([]value.Value, len(f.Values))
	copy(wordsCopy, f.Words)
	copy(valuesCopy, f.Values)

	return &Frame{
		Type:   f.Type,
		Words:  wordsCopy,
		Values: valuesCopy,
		Parent: f.Parent,
	}
}
