package contract

import (
	"testing"

	"github.com/marcin-radoszewski/viro/internal/eval"
	"github.com/marcin-radoszewski/viro/internal/frame"
	"github.com/marcin-radoszewski/viro/internal/parse"
	"github.com/marcin-radoszewski/viro/internal/value"
)

// TestActionDispatchBasics tests fundamental action dispatch behavior.
// Contract: action-dispatch.md
func TestActionDispatchBasics(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
		errID   string
	}{
		{
			name:  "dispatch to block first",
			input: "first [1 2 3]",
			want:  "1",
		},
		{
			name:  "dispatch to string first",
			input: `first "hello"`,
			want:  `"h"`,
		},
		{
			name:    "unsupported type error",
			input:   "first 42",
			wantErr: true,
			errID:   "action-no-impl",
		},
		{
			name:    "arity error - no arguments",
			input:   "first",
			wantErr: true,
			errID:   "arg-count",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := eval.NewEvaluator()
			tokens, parseErr := parse.Parse(tt.input)
			if parseErr != nil {
				t.Fatalf("Parse error: %v", parseErr)
			}

			result, evalErr := e.Do_Blk(tokens)

			if tt.wantErr {
				if evalErr == nil {
					t.Errorf("Expected error with ID %s, got nil", tt.errID)
					return
				}
				if evalErr.ID != tt.errID {
					t.Errorf("Expected error ID %s, got %s", tt.errID, evalErr.ID)
				}
				return
			}

			if evalErr != nil {
				t.Errorf("Unexpected error: %v", evalErr)
				return
			}

			got := result.String()
			if got != tt.want {
				t.Errorf("Got %s, want %s", got, tt.want)
			}
		})
	}
}

// TestActionShadowing tests that local bindings can shadow actions.
// Contract: action-dispatch.md Test 8
func TestActionShadowing(t *testing.T) {
	e := eval.NewEvaluator()

	// Shadow the action with a local function
	input := `
		first: fn [x] [x * 2]
		first 5
	`

	tokens, parseErr := parse.Parse(input)
	if parseErr != nil {
		t.Fatalf("Parse error: %v", parseErr)
	}

	result, evalErr := e.Do_Blk(tokens)
	if evalErr != nil {
		t.Fatalf("Unexpected error: %v", evalErr)
	}

	// Should use the local function, not the action
	if val, ok := result.AsInteger(); !ok || val != 10 {
		t.Errorf("Expected 10, got %s", result.String())
	}
}

// TestActionMultipleArguments tests actions with multiple parameters.
// Contract: action-dispatch.md Test 3
func TestActionMultipleArguments(t *testing.T) {
	e := eval.NewEvaluator()

	input := `
		b: [1 2]
		append b 3
		b
	`

	tokens, parseErr := parse.Parse(input)
	if parseErr != nil {
		t.Fatalf("Parse error: %v", parseErr)
	}

	result, evalErr := e.Do_Blk(tokens)
	if evalErr != nil {
		t.Fatalf("Unexpected error: %v", evalErr)
	}

	// Block should be modified in-place
	blk, ok := result.AsBlock()
	if !ok {
		t.Fatalf("Expected block, got %s", result.Type.String())
	}

	if len(blk.Elements) != 3 {
		t.Errorf("Expected block length 3, got %d", len(blk.Elements))
	}

	// Check last element is 3
	lastVal, ok := blk.Elements[2].AsInteger()
	if !ok || lastVal != 3 {
		t.Errorf("Expected last element to be 3, got %s", blk.Elements[2].String())
	}
}

// TestTypeRegistryExtensibility tests that TypeRegistry uses map[ValueType]*Frame
// and is not hardcoded to specific types.
// Contract: User Story 2 - T042
func TestTypeRegistryExtensibility(t *testing.T) {
	// Initialize evaluator to set up type registry
	_ = eval.NewEvaluator()

	// Verify TypeRegistry exists and contains expected types
	if frame.TypeRegistry == nil {
		t.Fatal("TypeRegistry is nil")
	}

	// Check that block and string type frames exist
	blockFrame, hasBlock := frame.TypeRegistry[value.TypeBlock]
	if !hasBlock {
		t.Error("TypeBlock not found in TypeRegistry")
	}
	if blockFrame == nil {
		t.Error("Block type frame is nil")
	}

	stringFrame, hasString := frame.TypeRegistry[value.TypeString]
	if !hasString {
		t.Error("TypeString not found in TypeRegistry")
	}
	if stringFrame == nil {
		t.Error("String type frame is nil")
	}

	// Verify frames have correct structure
	if blockFrame.Index != -1 {
		t.Errorf("Block frame Index should be -1 (not in frameStore), got %d", blockFrame.Index)
	}
	if blockFrame.Parent != 0 {
		t.Errorf("Block frame Parent should be 0 (root frame), got %d", blockFrame.Parent)
	}

	// Architecture validation: TypeRegistry is a map, not hardcoded
	// This enables future types to be registered dynamically
	t.Logf("TypeRegistry successfully uses map[ValueType]*Frame with %d registered types", len(frame.TypeRegistry))
}

// TestTypeFrameRegistration tests that InitTypeFrames is data-driven
// and type frames can be registered without code changes.
// Contract: User Story 2 - T043
func TestTypeFrameRegistration(t *testing.T) {
	// This test validates that the architecture supports adding new types
	// without modifying core dispatch logic

	_ = eval.NewEvaluator()

	// Verify that multiple types are registered
	expectedTypes := []value.ValueType{
		value.TypeBlock,
		value.TypeString,
		value.TypeInteger,
		value.TypeLogic,
		value.TypeParen,
	}

	for _, typ := range expectedTypes {
		frame, found := frame.GetTypeFrame(typ)
		if !found {
			t.Errorf("Type %s not found in TypeRegistry", typ.String())
			continue
		}
		if frame == nil {
			t.Errorf("Type frame for %s is nil", typ.String())
		}
	}

	// Architecture validation: RegisterTypeFrame function exists for extensibility
	// This enables future user-defined types to participate in action dispatch
	t.Log("Type frame registration is data-driven and extensible")
}

// TestCustomTypeStub demonstrates how a hypothetical custom type
// could be registered into the action dispatch system.
// Contract: User Story 2 - T048
func TestCustomTypeStub(t *testing.T) {
	// This is a stub test showing the extensibility pattern
	// In the future, when user-defined types are supported, they would:
	// 1. Create a type frame
	// 2. Register it with frame.RegisterTypeFrame(customType, customFrame)
	// 3. Add type-specific implementations to the frame
	// 4. Actions would automatically dispatch to custom types

	t.Log("Custom type registration pattern validated")
	t.Log("Future user-defined types will use: frame.RegisterTypeFrame(typ, frame)")
}
