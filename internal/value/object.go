package value

import "fmt"

// ObjectInstance represents an object with frame-based field storage (Feature 002).
//
// Design per data-model.md:
// - FrameIndex: index into frame registry/stack (reuses frame infrastructure)
// - ParentProto: reference to parent prototype object (nil if none) for prototype chain
// - Manifest: published field names and optional type hints
//
// Per FR-009: captures word/value pairs into dedicated frame with nested object support
type ObjectInstance struct {
	FrameIndex  int             // Index into frame storage
	ParentProto *ObjectInstance // Parent prototype object (nil = no parent)
	Manifest    ObjectManifest  // Field metadata

	// Deprecated: Parent field kept for backward compatibility
	Parent int // Parent object frame index (-1 = no parent)
}

// ObjectManifest describes the fields exposed by an object.
type ObjectManifest struct {
	Words []string    // Published field names (case-sensitive)
	Types []ValueType // Optional type hints (TypeNone = any type allowed)
}

// NewObject creates an ObjectInstance with the given frame and field manifest.
func NewObject(frameIndex int, words []string, types []ValueType) *ObjectInstance {
	if types == nil {
		// Default to TypeNone (any type) for all fields
		types = make([]ValueType, len(words))
	}
	return &ObjectInstance{
		FrameIndex:  frameIndex,
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
	return fmt.Sprintf("object[frame:%d fields:%d]", o.FrameIndex, len(o.Manifest.Words))
}

// ObjectVal creates a Value wrapping an ObjectInstance.
func ObjectVal(obj *ObjectInstance) Value {
	return Value{
		Type:    TypeObject,
		Payload: obj,
	}
}

// AsObject extracts the ObjectInstance from a Value, or returns nil if wrong type.
func (v Value) AsObject() (*ObjectInstance, bool) {
	if v.Type != TypeObject {
		return nil, false
	}
	obj, ok := v.Payload.(*ObjectInstance)
	return obj, ok
}
