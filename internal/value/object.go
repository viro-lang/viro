package value

import (
	"fmt"
	"strings"

	"github.com/marcin-radoszewski/viro/internal/core"
)

// ObjectInstance represents an object with frame-based field storage (Feature 002).
//
// Design per data-model.md:
// - FrameIndex: index into frame registry/stack (reuses frame infrastructure)
// - Frame: owned frame for self-contained field storage (Phase 1 refactor)
// - ParentProto: reference to parent prototype object (nil if none) for prototype chain
// - Manifest: published field names and optional type hints
//
// Per FR-009: captures word/value pairs into dedicated frame with nested object support
type ObjectInstance struct {
	FrameIndex  int             // Index into frame storage (backward compatibility)
	Frame       core.Frame      // Owned frame for self-contained storage
	ParentProto *ObjectInstance // Parent prototype object (nil = no parent)
	Manifest    ObjectManifest  // Field metadata

	// Deprecated: Parent field kept for backward compatibility
	Parent int // Parent object frame index (-1 = no parent)
}

// ObjectManifest describes the fields exposed by an object.
type ObjectManifest struct {
	Words []string         // Published field names (case-sensitive)
	Types []core.ValueType // Optional type hints (TypeNone = any type allowed)
}

// NewObject creates an ObjectInstance with the given frame and field manifest.
func NewObject(frameIndex int, words []string, types []core.ValueType) *ObjectInstance {
	if types == nil {
		// Default to TypeNone (any type) for all fields
		types = make([]core.ValueType, len(words))
	}
	return &ObjectInstance{
		FrameIndex:  frameIndex,
		Frame:       nil, // Owned frame will be set during object creation in native functions
		ParentProto: nil, // No parent by default
		Parent:      -1,  // Deprecated field
		Manifest: ObjectManifest{
			Words: words,
			Types: types,
		},
	}
}

// NewObjectWithFrame creates an ObjectInstance with owned frame for self-contained storage.
func NewObjectWithFrame(frameIndex int, ownedFrame core.Frame, words []string, types []core.ValueType) *ObjectInstance {
	if types == nil {
		// Default to TypeNone (any type) for all fields
		types = make([]core.ValueType, len(words))
	}
	return &ObjectInstance{
		FrameIndex:  frameIndex, // Keep for backward compatibility
		Frame:       ownedFrame,
		ParentProto: nil, // No parent by default
		Parent:      -1,  // Deprecated field
		Manifest: ObjectManifest{
			Words: words,
			Types: types,
		},
	}
}

// String returns a debug representation of the object.
func (o *ObjectInstance) String() string {
	if o == nil {
		return "object[]"
	}

	// Build field representation using owned frame
	var fields []string
	for _, fieldName := range o.Manifest.Words {
		if val, found := o.GetField(fieldName); found {
			fields = append(fields, fmt.Sprintf("%s: %s", fieldName, val.String()))
		} else {
			fields = append(fields, fmt.Sprintf("%s: <missing>", fieldName))
		}
	}

	if len(fields) == 0 {
		return "object[]"
	}

	return fmt.Sprintf("object[%s]", strings.Join(fields, " "))
}

// ObjectVal creates a Value wrapping an ObjectInstance.
func ObjectVal(obj *ObjectInstance) Value {
	return Value{
		Type:    TypeObject,
		Payload: obj,
	}
}

// AsObject extracts the ObjectInstance from a Value, or returns nil if wrong type.
func AsObject(v core.Value) (*ObjectInstance, bool) {
	if v.GetType() != TypeObject {
		return nil, false
	}
	obj, ok := v.GetPayload().(*ObjectInstance)
	return obj, ok
}

// GetField retrieves a field value from the owned frame.
// Returns (value, true) if found, (NoneVal, false) if not.
func (obj *ObjectInstance) GetField(name string) (core.Value, bool) {
	if obj.Frame == nil {
		return NoneVal(), false
	}
	return obj.Frame.Get(name)
}

// SetField sets a field value in the owned frame.
// Creates new binding if field doesn't exist.
// Note: During Phase 2 migration, also updates evaluator frame for compatibility.
func (obj *ObjectInstance) SetField(name string, val core.Value) {
	if obj.Frame == nil {
		return // No-op if no owned frame
	}
	obj.Frame.Bind(name, val)

	// Phase 2 compatibility: also update evaluator frame if it exists
	// This will be removed in Phase 3 when evaluator frames are eliminated
	if obj.FrameIndex >= 0 {
		// We don't have direct access to evaluator here, so this is handled
		// in the native functions that call SetField
	}
}

// GetFieldWithProto retrieves a field value, searching through prototype chain.
// Returns (value, true) if found, (NoneVal, false) if not.
func (obj *ObjectInstance) GetFieldWithProto(name string) (core.Value, bool) {
	// First check owned frame
	if val, found := obj.GetField(name); found {
		return val, true
	}

	// Then check prototype chain
	current := obj.ParentProto
	for current != nil {
		if val, found := current.GetField(name); found {
			return val, true
		}
		current = current.ParentProto
	}

	return NoneVal(), false
}
