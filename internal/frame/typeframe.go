// Package frame - type frame initialization and management
package frame

import (
	"github.com/marcin-radoszewski/viro/internal/core"
	"github.com/marcin-radoszewski/viro/internal/value"
)

var TypeRegistry map[core.ValueType]core.Frame

func InitTypeFrames() {
	TypeRegistry = make(map[core.ValueType]core.Frame)

	TypeRegistry[value.TypeBlock] = createTypeFrame("block!")
	TypeRegistry[value.TypeString] = createTypeFrame("string!")
	TypeRegistry[value.TypeBinary] = createTypeFrame("binary!")

	TypeRegistry[value.TypeObject] = createTypeFrame("object!")
}

func createTypeFrame(typeName string) core.Frame {
	frame := NewFrame(FrameTypeFrame, 0)
	frame.SetIndex(-1)
	frame.SetName(typeName)
	return frame
}

func GetTypeFrame(typ core.ValueType) (core.Frame, bool) {
	frame, exists := TypeRegistry[typ]
	return frame, exists
}

func RegisterTypeFrame(typ core.ValueType, frame core.Frame) {
	if TypeRegistry == nil {
		TypeRegistry = make(map[core.ValueType]core.Frame)
	}
	TypeRegistry[typ] = frame
}
