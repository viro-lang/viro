package native

import (
	"os"
	"testing"

	"github.com/marcin-radoszewski/viro/internal/frame"
	"github.com/marcin-radoszewski/viro/internal/value"
	"github.com/marcin-radoszewski/viro/internal/verror"
)

// TestMain sets up the native Registry before running tests
func TestMain(m *testing.M) {
	// Create a temporary frame to register natives
	tempFrame := frame.NewFrameWithCapacity(frame.FrameClosure, -1, 80)

	// Register all natives to populate the Registry
	RegisterMathNatives(tempFrame)
	RegisterSeriesNatives(tempFrame)
	RegisterDataNatives(tempFrame)
	RegisterIONatives(tempFrame)
	RegisterControlNatives(tempFrame)
	RegisterHelpNatives(tempFrame)

	// Run tests
	os.Exit(m.Run())
}

// mockEvaluator is a simple mock implementation of the Evaluator interface for testing.
type mockEvaluator struct{}

func (m mockEvaluator) Do_Blk(vals []value.Value) (value.Value, *verror.Error) {
	// Simple mock - just return none
	return value.NoneVal(), nil
}

func (m mockEvaluator) Do_Next(val value.Value) (value.Value, *verror.Error) {
	// Simple mock - just return the value
	return val, nil
}

func TestRegistrySimpleMath(t *testing.T) {
	// Test addition (+)
	t.Run("Add", func(t *testing.T) {
		fn, ok := Lookup("+")
		if !ok {
			t.Fatal("Function '+' not found in Registry")
		}
		if fn.Name != "+" {
			t.Errorf("Expected name '+', got %q", fn.Name)
		}
		if !fn.Infix {
			t.Error("Expected '+' to be infix")
		}
		if fn.Type != value.FuncNative {
			t.Error("Expected function type to be FuncNative")
		}
		if len(fn.Params) != 2 {
			t.Errorf("Expected 2 params, got %d", len(fn.Params))
		}

		// Test invocation
		result, err := Call(fn, []value.Value{value.IntVal(3), value.IntVal(4)}, map[string]value.Value{}, mockEvaluator{})
		if err != nil {
			t.Fatalf("Call failed: %v", err)
		}
		if result.Type != value.TypeInteger {
			t.Errorf("Expected integer result, got %v", result.Type)
		}
		intVal, ok := result.AsInteger()
		if !ok || intVal != 7 {
			t.Errorf("Expected 7, got %v", intVal)
		}
	})

	// Test subtraction (-)
	t.Run("Subtract", func(t *testing.T) {
		fn, ok := Lookup("-")
		if !ok {
			t.Fatal("Function '-' not found in Registry")
		}
		if !fn.Infix {
			t.Error("Expected '-' to be infix")
		}

		result, err := Call(fn, []value.Value{value.IntVal(10), value.IntVal(3)}, map[string]value.Value{}, mockEvaluator{})
		if err != nil {
			t.Fatalf("Call failed: %v", err)
		}
		intVal, ok := result.AsInteger()
		if !ok || intVal != 7 {
			t.Errorf("Expected 7, got %v", intVal)
		}
	})

	// Test multiplication (*)
	t.Run("Multiply", func(t *testing.T) {
		fn, ok := Lookup("*")
		if !ok {
			t.Fatal("Function '*' not found in Registry")
		}
		if !fn.Infix {
			t.Error("Expected '*' to be infix")
		}

		result, err := Call(fn, []value.Value{value.IntVal(3), value.IntVal(4)}, map[string]value.Value{}, mockEvaluator{})
		if err != nil {
			t.Fatalf("Call failed: %v", err)
		}
		intVal, ok := result.AsInteger()
		if !ok || intVal != 12 {
			t.Errorf("Expected 12, got %v", intVal)
		}
	})

	// Test division (/)
	t.Run("Divide", func(t *testing.T) {
		fn, ok := Lookup("/")
		if !ok {
			t.Fatal("Function '/' not found in Registry")
		}
		if !fn.Infix {
			t.Error("Expected '/' to be infix")
		}

		// Test integer division (10 / 2 = 5 as integer)
		result, err := Call(fn, []value.Value{value.IntVal(10), value.IntVal(2)}, map[string]value.Value{}, mockEvaluator{})
		if err != nil {
			t.Fatalf("Call failed: %v", err)
		}
		// When both inputs are integers, mathOp returns integer result
		if result.Type != value.TypeInteger {
			t.Errorf("Expected integer result for int/int, got %v", result.Type)
		}
		intVal, ok := result.AsInteger()
		if !ok || intVal != 5 {
			t.Errorf("Expected 5, got %v", intVal)
		}
	})
}

func TestRegistryComparison(t *testing.T) {
	// Test less than (<)
	t.Run("LessThan", func(t *testing.T) {
		fn, ok := Lookup("<")
		if !ok {
			t.Fatal("Function '<' not found in Registry")
		}
		if !fn.Infix {
			t.Error("Expected '<' to be infix")
		}

		result, err := Call(fn, []value.Value{value.IntVal(3), value.IntVal(5)}, map[string]value.Value{}, mockEvaluator{})
		if err != nil {
			t.Fatalf("Call failed: %v", err)
		}
		if result.Type != value.TypeLogic {
			t.Errorf("Expected logic result, got %v", result.Type)
		}
		logicVal, ok := result.AsLogic()
		if !ok || !logicVal {
			t.Errorf("Expected true, got %v", logicVal)
		}
	})

	// Test greater than (>)
	t.Run("GreaterThan", func(t *testing.T) {
		fn, ok := Lookup(">")
		if !ok {
			t.Fatal("Function '>' not found in Registry")
		}

		result, err := Call(fn, []value.Value{value.IntVal(10), value.IntVal(5)}, map[string]value.Value{}, mockEvaluator{})
		if err != nil {
			t.Fatalf("Call failed: %v", err)
		}
		logicVal, ok := result.AsLogic()
		if !ok || !logicVal {
			t.Errorf("Expected true, got %v", logicVal)
		}
	})

	// Test equal (=)
	t.Run("Equal", func(t *testing.T) {
		fn, ok := Lookup("=")
		if !ok {
			t.Fatal("Function '=' not found in Registry")
		}

		result, err := Call(fn, []value.Value{value.IntVal(5), value.IntVal(5)}, map[string]value.Value{}, mockEvaluator{})
		if err != nil {
			t.Fatalf("Call failed: %v", err)
		}
		logicVal, ok := result.AsLogic()
		if !ok || !logicVal {
			t.Errorf("Expected true, got %v", logicVal)
		}
	})

	// Test not equal (<>)
	t.Run("NotEqual", func(t *testing.T) {
		fn, ok := Lookup("<>")
		if !ok {
			t.Fatal("Function '<>' not found in Registry")
		}

		result, err := Call(fn, []value.Value{value.IntVal(3), value.IntVal(4)}, map[string]value.Value{}, mockEvaluator{})
		if err != nil {
			t.Fatalf("Call failed: %v", err)
		}
		logicVal, ok := result.AsLogic()
		if !ok || !logicVal {
			t.Errorf("Expected true, got %v", logicVal)
		}
	})
}

func TestRegistryMetadata(t *testing.T) {
	t.Run("Documentation", func(t *testing.T) {
		fn, ok := Lookup("+")
		if !ok {
			t.Fatal("Function '+' not found")
		}
		if fn.Doc == nil {
			t.Error("Expected documentation for '+'")
		} else {
			if fn.Doc.Category != "Math" {
				t.Errorf("Expected category 'Math', got %q", fn.Doc.Category)
			}
			if fn.Doc.Summary == "" {
				t.Error("Expected non-empty summary")
			}
		}
	})

	t.Run("ParamSpecs", func(t *testing.T) {
		fn, ok := Lookup("+")
		if !ok {
			t.Fatal("Function '+' not found")
		}
		if len(fn.Params) != 2 {
			t.Fatalf("Expected 2 params, got %d", len(fn.Params))
		}
		// Both params should be evaluated
		if !fn.Params[0].Eval {
			t.Error("Expected first param to be evaluated")
		}
		if !fn.Params[1].Eval {
			t.Error("Expected second param to be evaluated")
		}
		// Check param names
		if fn.Params[0].Name != "left" {
			t.Errorf("Expected first param name 'left', got %q", fn.Params[0].Name)
		}
		if fn.Params[1].Name != "right" {
			t.Errorf("Expected second param name 'right', got %q", fn.Params[1].Name)
		}
	})

	t.Run("ParamNamesMatchDocumentation", func(t *testing.T) {
		// Verify that parameter names in ParamSpec match the documentation
		for name, fn := range Registry {
			if fn.Doc == nil || len(fn.Doc.Parameters) == 0 {
				continue // Skip functions without documentation
			}

			// Count non-refinement params in ParamSpec
			positionalParams := []value.ParamSpec{}
			for _, param := range fn.Params {
				if !param.Refinement {
					positionalParams = append(positionalParams, param)
				}
			}

			// Count non-refinement params in Documentation (parameters that don't start with --)
			positionalDocParams := []string{}
			for _, docParam := range fn.Doc.Parameters {
				if len(docParam.Name) == 0 || docParam.Name[0] != '-' {
					positionalDocParams = append(positionalDocParams, docParam.Name)
				}
			}

			if len(positionalParams) != len(positionalDocParams) {
				t.Errorf("Function %q: param count mismatch - ParamSpec has %d positional, Doc has %d positional",
					name, len(positionalParams), len(positionalDocParams))
				continue
			}

			// Check that names match
			for i, param := range positionalParams {
				expectedName := positionalDocParams[i]
				if param.Name != expectedName {
					t.Errorf("Function %q param %d: expected name %q, got %q",
						name, i, expectedName, param.Name)
				}
			}
		}
	})
}
