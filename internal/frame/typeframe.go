// Package frame - type frame initialization and management
package frame

import "github.com/marcin-radoszewski/viro/internal/value"

// TypeRegistry maps each value type to its corresponding type frame.
// Type frames contain type-specific function implementations for actions.
//
// Type frames are stored directly in this registry (not on stack).
// They use Parent=0 (index to root frame) and Index=-1 (not in frameStore).
//
// Feature: 004-dynamic-function-invocation
var TypeRegistry map[value.ValueType]*Frame

// InitTypeFrames creates type frames for all native value types.
// Called during interpreter startup after root frame creation.
//
// Each type frame:
// - Parent = 0 (index to root frame on stack)
// - Index = -1 (not in frameStore)
// - Type = FrameFunctionArgs (reusing existing type for consistency)
//
// Type-specific implementations will be registered into these frames
// by the native package during its initialization.
func InitTypeFrames() {
	TypeRegistry = make(map[value.ValueType]*Frame)

	// Create type frames for series types (primary use case for actions)
	TypeRegistry[value.TypeBlock] = createTypeFrame("block!")
	TypeRegistry[value.TypeString] = createTypeFrame("string!")

	// Create type frames for other native types (future extensibility)
	TypeRegistry[value.TypeInteger] = createTypeFrame("integer!")
	TypeRegistry[value.TypeLogic] = createTypeFrame("logic!")
	TypeRegistry[value.TypeParen] = createTypeFrame("paren!")

	// Note: Additional types can be registered later via RegisterTypeFrame
}

// createTypeFrame creates a type frame with standard configuration.
func createTypeFrame(typeName string) *Frame {
	frame := NewFrame(FrameFunctionArgs, 0) // Parent = 0 (root frame)
	frame.Index = -1                        // Not in frameStore
	frame.Name = typeName                   // For diagnostics
	return frame
}

// GetTypeFrame retrieves the type frame for a given value type.
// Returns (frame, true) if type has a frame, (nil, false) otherwise.
func GetTypeFrame(typ value.ValueType) (*Frame, bool) {
	frame, exists := TypeRegistry[typ]
	return frame, exists
}

// RegisterTypeFrame registers a type frame for a custom value type.
// This enables user-defined types to participate in action dispatch.
//
// The frame must have:
// - Parent = 0 (root frame)
// - Index = -1 (not in frameStore)
//
// Example usage for a hypothetical custom type:
//
//	// 1. Create a type frame
//	customFrame := frame.NewFrame(frame.FrameFunctionArgs, 0)
//	customFrame.Index = -1
//	customFrame.Name = "custom-type!"
//
//	// 2. Add type-specific implementations
//	customFrame.Bind("first", value.FuncVal(customFirstImpl))
//	customFrame.Bind("last", value.FuncVal(customLastImpl))
//
//	// 3. Register the type frame
//	frame.RegisterTypeFrame(value.TypeCustom, customFrame)
//
//	// 4. Actions will now dispatch to custom types automatically
//	// first custom-value  ; → calls customFirstImpl
//
// Feature: 004-dynamic-function-invocation (extensibility)
func RegisterTypeFrame(typ value.ValueType, frame *Frame) {
	if TypeRegistry == nil {
		TypeRegistry = make(map[value.ValueType]*Frame)
	}
	TypeRegistry[typ] = frame
}
