package contract

import (
	"testing"

	"github.com/marcin-radoszewski/viro/internal/core"
	"github.com/marcin-radoszewski/viro/internal/frame"
	"github.com/marcin-radoszewski/viro/internal/value"
)

func TestFrameTypeConstants(t *testing.T) {
	tests := []struct {
		name      string
		frameType core.FrameType
		expected  core.FrameType
	}{
		{"FrameFunctionArgs is 0", frame.FrameFunctionArgs, 0},
		{"FrameClosure is 1", frame.FrameClosure, 1},
		{"FrameObject is 2", frame.FrameObject, 2},
		{"FrameTypeFrame is 3", frame.FrameTypeFrame, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if uint8(tt.frameType) != uint8(tt.expected) {
				t.Errorf("Expected %s to be %d, got %d", tt.name, tt.expected, tt.frameType)
			}
		})
	}
}

func TestFrameTypeConstantsAreDistinct(t *testing.T) {
	types := []core.FrameType{
		frame.FrameFunctionArgs,
		frame.FrameClosure,
		frame.FrameObject,
		frame.FrameTypeFrame,
	}

	seen := make(map[core.FrameType]bool)
	for _, typ := range types {
		if seen[typ] {
			t.Errorf("Duplicate frame type value: %d", typ)
		}
		seen[typ] = true
	}

	if len(seen) != 4 {
		t.Errorf("Expected 4 distinct frame types, got %d", len(seen))
	}
}

func TestTypeFrameHasCorrectType(t *testing.T) {
	frame.InitTypeFrames()

	tests := []struct {
		name      string
		valueType core.ValueType
	}{
		{"block type frame", value.TypeBlock},
		{"string type frame", value.TypeString},
		{"binary type frame", value.TypeBinary},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			typeFrame, exists := frame.GetTypeFrame(tt.valueType)
			if !exists {
				t.Fatalf("Type frame for %s not found", tt.name)
			}

			if typeFrame.GetType() != frame.FrameTypeFrame {
				t.Errorf("Expected type frame to have type FrameTypeFrame, got %d", typeFrame.GetType())
			}

			if typeFrame.GetParent() != 0 {
				t.Errorf("Expected type frame parent to be 0 (root frame), got %d", typeFrame.GetParent())
			}

			if typeFrame.GetIndex() != -1 {
				t.Errorf("Expected type frame index to be -1 (not in frameStore), got %d", typeFrame.GetIndex())
			}
		})
	}
}

func TestRegisterTypeFrameCreatesCorrectType(t *testing.T) {
	customFrame := frame.NewFrame(frame.FrameTypeFrame, 0)
	customFrame.SetIndex(-1)
	customFrame.SetName("custom-type!")

	frame.RegisterTypeFrame(value.TypeInteger, customFrame)

	retrieved, exists := frame.GetTypeFrame(value.TypeInteger)
	if !exists {
		t.Fatal("Custom type frame not found after registration")
	}

	if retrieved.GetType() != frame.FrameTypeFrame {
		t.Errorf("Expected registered type frame to have type FrameTypeFrame, got %d", retrieved.GetType())
	}

	if retrieved.GetName() != "custom-type!" {
		t.Errorf("Expected registered type frame to have name 'custom-type!', got %s", retrieved.GetName())
	}

	if retrieved.GetParent() != 0 {
		t.Errorf("Expected registered type frame parent to be 0, got %d", retrieved.GetParent())
	}

	if retrieved.GetIndex() != -1 {
		t.Errorf("Expected registered type frame index to be -1, got %d", retrieved.GetIndex())
	}
}

func TestFrameTypesAreSemanticallyClear(t *testing.T) {
	functionFrame := frame.NewFrame(frame.FrameFunctionArgs, 0)
	if functionFrame.GetType() == frame.FrameTypeFrame {
		t.Error("Function frame should not have type FrameTypeFrame")
	}

	closureFrame := frame.NewFrame(frame.FrameClosure, -1)
	if closureFrame.GetType() == frame.FrameTypeFrame {
		t.Error("Closure frame should not have type FrameTypeFrame")
	}

	objectFrame := frame.NewObjectFrame(0, []string{"x", "y"}, nil)
	if objectFrame.GetType() == frame.FrameTypeFrame {
		t.Error("Object frame should not have type FrameTypeFrame")
	}

	typeFrame := frame.NewFrame(frame.FrameTypeFrame, 0)
	if typeFrame.GetType() != frame.FrameTypeFrame {
		t.Error("Type frame should have type FrameTypeFrame")
	}
}
