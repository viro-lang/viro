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
//   - FrameObject: Object instance frame
//   - FrameTypeFrame: Type frame for action dispatch
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
	"github.com/marcin-radoszewski/viro/internal/core"
	"github.com/marcin-radoszewski/viro/internal/value"
	"slices"
)

const (
	FrameFunctionArgs core.FrameType = iota // Function call frame (arguments + locals)
	FrameClosure                            // Closure captured environment
	FrameObject                             // Object instance frame
	FrameTypeFrame                          // Type frame for action dispatch
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
//
// Frame Chain Navigation:
// - Parent field is essential for lexical scoping and frame chain traversal
// - Used by evaluator's Lookup method to walk parent chain
// - Parent=0 for root frame and type frames (links to root)
// - Parent=-1 for frames with no parent (root frame itself)
type Frame struct {
	Type   core.FrameType // Frame category
	Words  []string       // Symbol names (parallel to Values)
	Values []core.Value   // Bound values (parallel to Words)
	Parent int            // Parent frame index for lexical scoping (-1 if none). Essential for frame chain traversal.
	Index  int            // Position in evaluator's frameStore (-1 if not yet stored)
	Name   string         // Optional function or context name for diagnostics
}

// NewFrame creates an empty frame.
func NewFrame(frameType core.FrameType, parent int) core.Frame {
	return &Frame{
		Type:   frameType,
		Words:  []string{},
		Values: []core.Value{},
		Parent: parent,
		Index:  -1,
		Name:   "",
	}
}

// NewFrameWithCapacity creates a frame with pre-allocated capacity.
// Useful for function frames where parameter count is known.
func NewFrameWithCapacity(frameType core.FrameType, parent int, capacity int) *Frame {
	return &Frame{
		Type:   frameType,
		Words:  make([]string, 0, capacity),
		Values: make([]core.Value, 0, capacity),
		Parent: parent,
		Index:  -1,
		Name:   "",
	}
}

// NewObjectFrame creates an object frame.
// Feature 002: Used by the object native to create object instances.
func NewObjectFrame(parent int, words []string, types []core.ValueType) *Frame {
	return &Frame{
		Type:   FrameObject,
		Words:  make([]string, 0, len(words)),
		Values: make([]core.Value, 0, len(words)),
		Parent: parent,
		Index:  -1,
		Name:   "",
	}
}

func (f *Frame) GetType() core.FrameType {
	return f.Type
}

func (f *Frame) ChangeType(newType core.FrameType) {
	f.Type = newType
}

func (f *Frame) GetParent() int {
	return f.Parent
}

func (f *Frame) GetIndex() int {
	return f.Index
}

func (f *Frame) SetIndex(idx int) {
	f.Index = idx
}

func (f *Frame) GetName() string {
	return f.Name
}

func (f *Frame) SetName(name string) {
	f.Name = name
}

// Bind adds or updates a word binding in this frame.
// Local-by-default: creates new binding if word doesn't exist.
//
// Per data-model.md: This is the core of local-by-default scoping.
// Assignment in a function creates a local variable, NOT a global.
func (f *Frame) Bind(symbol string, val core.Value) {
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
func (f *Frame) Get(symbol string) (core.Value, bool) {
	for i, w := range f.Words {
		if w == symbol {
			return f.Values[i], true
		}
	}
	return value.NewNoneVal(), false
}

// Set updates an existing binding in this frame.
// Returns true if word was found and updated, false if not found.
// Does NOT create new binding (use Bind for that).
func (f *Frame) Set(symbol string, val core.Value) bool {
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
	return slices.Contains(f.Words, symbol)
}

// Count returns the number of bindings in this frame.
func (f *Frame) Count() int {
	return len(f.Words)
}

// GetAll returns all bindings as (symbol, value) pairs.
// Useful for debugging and inspection.
func (f *Frame) GetAll() []core.Binding {
	bindings := make([]core.Binding, len(f.Words))
	for i := range f.Words {
		bindings[i] = core.Binding{
			Symbol: f.Words[i],
			Value:  f.Values[i],
		}
	}
	return bindings
}

// Clone creates a shallow copy of the frame.
// Words and values are copied, but value contents are shared.
// Used for closure capture.
func (f *Frame) Clone() core.Frame {
	wordsCopy := make([]string, len(f.Words))
	valuesCopy := make([]core.Value, len(f.Values))
	copy(wordsCopy, f.Words)
	copy(valuesCopy, f.Values)

	return &Frame{
		Type:   f.Type,
		Words:  wordsCopy,
		Values: valuesCopy,
		Parent: f.Parent,
		Index:  -1, // Clone gets a new index when added to frameStore
	}
}
