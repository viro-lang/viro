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
	FrameObject                        // Object instance frame (Feature 002)
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
// Feature 002: Objects
// - Manifest: Optional field metadata for object frames (type validation)
type Frame struct {
	Type     FrameType       // Frame category
	Words    []string        // Symbol names (parallel to Values)
	Values   []value.Value   // Bound values (parallel to Words)
	Parent   int             // Index of parent frame for closures (-1 if none) (deprecated: use Index to navigate frameStore)
	Index    int             // Position in evaluator's frameStore (-1 if not yet stored)
	Name     string          // Optional function or context name for diagnostics
	Manifest *ObjectManifest // Optional: field metadata for objects (Feature 002)
}

// ObjectManifest describes the fields and type constraints for an object frame.
// Used for type validation during field assignment (Feature 002, FR-009).
type ObjectManifest struct {
	Words []string          // Published field names (case-sensitive)
	Types []value.ValueType // Optional type hints (TypeNone = any type allowed)
}

// NewFrame creates an empty frame.
func NewFrame(frameType FrameType, parent int) *Frame {
	return &Frame{
		Type:     frameType,
		Words:    []string{},
		Values:   []value.Value{},
		Parent:   parent,
		Index:    -1,
		Name:     "",
		Manifest: nil,
	}
}

// NewFrameWithCapacity creates a frame with pre-allocated capacity.
// Useful for function frames where parameter count is known.
func NewFrameWithCapacity(frameType FrameType, parent int, capacity int) *Frame {
	return &Frame{
		Type:     frameType,
		Words:    make([]string, 0, capacity),
		Values:   make([]value.Value, 0, capacity),
		Parent:   parent,
		Index:    -1,
		Name:     "",
		Manifest: nil,
	}
}

// NewObjectFrame creates an object frame with a manifest for type validation.
// Feature 002: Used by the object native to create typed object instances.
func NewObjectFrame(parent int, words []string, types []value.ValueType) *Frame {
	if len(types) == 0 {
		// Default to TypeNone (any type) for all fields
		types = make([]value.ValueType, len(words))
		for i := range types {
			types[i] = value.TypeNone
		}
	}

	return &Frame{
		Type:   FrameObject,
		Words:  make([]string, 0, len(words)),
		Values: make([]value.Value, 0, len(words)),
		Parent: parent,
		Index:  -1,
		Name:   "",
		Manifest: &ObjectManifest{
			Words: words,
			Types: types,
		},
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

	// Copy manifest if present (Feature 002: objects)
	var manifestCopy *ObjectManifest
	if f.Manifest != nil {
		manifestWordsCopy := make([]string, len(f.Manifest.Words))
		manifestTypesCopy := make([]value.ValueType, len(f.Manifest.Types))
		copy(manifestWordsCopy, f.Manifest.Words)
		copy(manifestTypesCopy, f.Manifest.Types)
		manifestCopy = &ObjectManifest{
			Words: manifestWordsCopy,
			Types: manifestTypesCopy,
		}
	}

	return &Frame{
		Type:     f.Type,
		Words:    wordsCopy,
		Values:   valuesCopy,
		Parent:   f.Parent,
		Index:    -1, // Clone gets a new index when added to frameStore
		Manifest: manifestCopy,
	}
}

// ValidateFieldType checks if a value matches the expected type for a field in an object frame.
// Feature 002: Used during object field assignment to enforce type constraints.
// Returns true if validation passes or if no type constraint exists (TypeNone).
func (f *Frame) ValidateFieldType(symbol string, val value.Value) bool {
	if f.Manifest == nil {
		return true // No manifest = no type validation
	}

	// Find the field index in the manifest
	for i, word := range f.Manifest.Words {
		if word == symbol {
			expectedType := f.Manifest.Types[i]
			if expectedType == value.TypeNone {
				return true // TypeNone allows any type
			}
			return val.Type == expectedType
		}
	}

	return true // Field not in manifest = no type constraint
}

// HasManifestField checks if a field is declared in the object's manifest.
// Feature 002: Used to validate field access in objects.
func (f *Frame) HasManifestField(symbol string) bool {
	if f.Manifest == nil {
		return false
	}

	for _, word := range f.Manifest.Words {
		if word == symbol {
			return true
		}
	}

	return false
}
