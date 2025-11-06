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
	baseValue
	Frame       core.Frame      // Owned frame for self-contained storage
	ParentProto *ObjectInstance // Parent prototype object (nil = no parent)
}

// NewObject creates an ObjectInstance with owned frame for self-contained storage.
func NewObject(ownedFrame core.Frame) *ObjectInstance {
	return &ObjectInstance{
		Frame:       ownedFrame,
		ParentProto: nil, // No parent by default
	}
}

// String returns a debug representation of the object.
func (o *ObjectInstance) String() string {
	if o == nil {
		return "object[]"
	}

	// Build field representation including inherited fields from prototype chain
	var fields []string
	bindings := o.GetAllFieldsWithProto()
	for _, binding := range bindings {
		fields = append(fields, fmt.Sprintf("%s: %s", binding.Symbol, binding.Value.Form()))
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

	// Build field assignments including inherited fields from prototype chain
	fieldAssignments := []string{}
	bindings := o.GetAllFieldsWithProto()
	for _, binding := range bindings {
		moldedVal := binding.Value.Mold()
		fieldAssignments = append(fieldAssignments, fmt.Sprintf("%s: %s", binding.Symbol, moldedVal))
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

	// Build field display lines including inherited fields from prototype chain
	fieldLines := []string{}
	bindings := o.GetAllFieldsWithProto()
	for _, binding := range bindings {
		displayVal := binding.Value.Form()
		fieldLines = append(fieldLines, fmt.Sprintf("%s: %s", binding.Symbol, displayVal))
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
	obj, ok := v.(*ObjectInstance)
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

// GetAllFieldsWithProto collects all accessible fields (own + inherited).
// Returns bindings in order: parent fields first, then child fields (child overrides parent).
func (obj *ObjectInstance) GetAllFieldsWithProto() []core.Binding {
	seen := make(map[string]bool)
	var result []core.Binding

	// Walk up the prototype chain to collect all objects
	chain := []*ObjectInstance{}
	current := obj
	for current != nil {
		chain = append(chain, current)
		current = current.ParentProto
	}

	// Reverse walk: start from root ancestor to child, so we preserve order
	// Child fields that override parent fields will appear in child's position
	for i := len(chain) - 1; i >= 0; i-- {
		bindings := chain[i].Frame.GetAll()
		for _, binding := range bindings {
			if !seen[binding.Symbol] {
				seen[binding.Symbol] = true
				result = append(result, binding)
			} else {
				// Update the value if field was already seen (child overriding parent)
				for j := range result {
					if result[j].Symbol == binding.Symbol {
						result[j].Value = binding.Value
						break
					}
				}
			}
		}
	}

	return result
}

func (obj *ObjectInstance) GetType() core.ValueType {
	return TypeObject
}

func (obj *ObjectInstance) GetPayload() any {
	return obj
}

func (obj *ObjectInstance) Equals(other core.Value) bool {
	otherObj, ok := other.(*ObjectInstance)
	if !ok {
		return false
	}
	return otherObj == obj
}
