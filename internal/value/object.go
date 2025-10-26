package value

import (
	"fmt"
	"strings"

	"github.com/marcin-radoszewski/viro/internal/core"
)

// ObjectInstance represents an object with frame-based field storage (Feature 002).
//
// Design per data-model.md:
// - Frame: owned frame for self-contained field storage
// - ParentProto: reference to parent prototype object (nil if none) for prototype chain
// - Manifest: published field names and optional type hints
//
// Per FR-009: captures word/value pairs into dedicated frame with nested object support
type ObjectInstance struct {
	Frame       core.Frame      // Owned frame for self-contained storage
	ParentProto *ObjectInstance // Parent prototype object (nil = no parent)
	Manifest    ObjectManifest  // Field metadata
}

// ObjectManifest describes the fields exposed by an object.
type ObjectManifest struct {
	Words []string         // Published field names (case-sensitive)
	Types []core.ValueType // Optional type hints (TypeNone = any type allowed)
}

// NewObject creates an ObjectInstance with owned frame for self-contained storage.
func NewObject(ownedFrame core.Frame, words []string, types []core.ValueType) *ObjectInstance {
	if types == nil {
		// Default to TypeNone (any type) for all fields
		types = make([]core.ValueType, len(words))
	}
	return &ObjectInstance{
		Frame:       ownedFrame,
		ParentProto: nil, // No parent by default
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
			fields = append(fields, fmt.Sprintf("%s: %s", fieldName, val.Form()))
		} else {
			fields = append(fields, fmt.Sprintf("%s: <missing>", fieldName))
		}
	}

	if len(fields) == 0 {
		return "object[]"
	}

	return fmt.Sprintf("object[%s]", strings.Join(fields, " "))
}

// Mold returns the mold-formatted object representation (make object! format).
func (o *ObjectInstance) Mold() string {
	if o == nil {
		return "make object! []"
	}

	// Build field assignments using owned frame
	fieldAssignments := []string{}
	for _, fieldName := range o.Manifest.Words {
		if fieldVal, found := o.GetField(fieldName); found {
			// Recursively mold the field value
			moldedVal := fieldVal.Mold() // Use Mold() for proper recursive molding
			fieldAssignments = append(fieldAssignments, fmt.Sprintf("%s: %s", fieldName, moldedVal))
		}
	}

	if len(fieldAssignments) == 0 {
		return "make object! []"
	}

	return fmt.Sprintf("make object! [%s]", strings.Join(fieldAssignments, " "))
}

// Form returns the form-formatted object representation (multi-line field display).
func (o *ObjectInstance) Form() string {
	if o == nil {
		return ""
	}

	// Build field display lines using owned frame
	fieldLines := []string{}
	for _, fieldName := range o.Manifest.Words {
		if fieldVal, found := o.GetField(fieldName); found {
			// Use Form() for human-readable field values
			displayVal := fieldVal.Form()
			fieldLines = append(fieldLines, fmt.Sprintf("%s: %s", fieldName, displayVal))
		}
	}

	if len(fieldLines) == 0 {
		return ""
	}

	return strings.Join(fieldLines, "\n")
}

// ObjectVal creates a Value wrapping an ObjectInstance.
func ObjectVal(obj *ObjectInstance) core.Value {
	return obj
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
// Returns (value, true) if found, (NewNoneVal, false) if not.
func (obj *ObjectInstance) GetField(name string) (core.Value, bool) {
	if obj.Frame == nil {
		return NewNoneVal(), false
	}
	return obj.Frame.Get(name)
}

// SetField sets a field value in the owned frame.
// Creates new binding if field doesn't exist.
func (obj *ObjectInstance) SetField(name string, val core.Value) {
	if obj.Frame == nil {
		return // No-op if no owned frame
	}
	obj.Frame.Bind(name, val)
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

	return NewNoneVal(), false
}

func (obj *ObjectInstance) GetType() core.ValueType {
	return TypeObject
}

func (obj *ObjectInstance) GetPayload() any {
	return obj
}

func (obj *ObjectInstance) Equals(other core.Value) bool {
	if other.GetType() != TypeObject {
		return false
	}
	return other.GetPayload() == obj
}
