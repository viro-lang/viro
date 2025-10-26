package contract

import (
	"testing"

	"github.com/ericlagergren/decimal"
	"github.com/marcin-radoszewski/viro/internal/core"
	"github.com/marcin-radoszewski/viro/internal/frame"
	"github.com/marcin-radoszewski/viro/internal/value"
)

// T025.1: Foundation TDD checkpoint - Contract tests for new value types
// These tests MUST FAIL before implementation to verify TDD compliance per Constitution Principle I

// TestDecimalValueConstruction validates DecimalValue creation and basic properties
func TestDecimalValueConstruction(t *testing.T) {
	// Test decimal construction from Big number
	mag := decimal.New(1999, 2) // 19.99
	dec := value.NewDecimal(mag, 2)

	if dec == nil {
		t.Fatal("NewDecimal returned nil")
	}

	if dec.Scale != 2 {
		t.Errorf("expected scale 2, got %d", dec.Scale)
	}

	if dec.Context.Precision != 34 {
		t.Errorf("expected precision 34, got %d", dec.Context.Precision)
	}

	// Test String() output
	str := dec.String()
	if str == "" {
		t.Error("DecimalValue.String() returned empty string")
	}
}

// TestDecimalValueWrapping validates Value wrapping for decimals
func TestDecimalValueWrapping(t *testing.T) {
	mag := decimal.New(42, 0)
	val := value.DecimalVal(mag, 0)

	if val.GetType() != value.TypeDecimal {
		t.Errorf("expected TypeDecimal, got %v", val.GetType())
	}

	// Test AsDecimal extraction
	dec, ok := value.AsDecimal(val)
	if !ok {
		t.Error("AsDecimal returned false for decimal value")
	}
	if dec == nil {
		t.Error("AsDecimal returned nil")
	}

	// Test wrong type
	intVal := value.NewIntVal(42)
	_, ok = value.AsDecimal(intVal)
	if ok {
		t.Error("AsDecimal returned true for integer value")
	}
}

// TestObjectInstanceConstruction validates ObjectInstance creation
func TestObjectInstanceConstruction(t *testing.T) {
	words := []string{"name", "age"}
	types := []core.ValueType{value.TypeString, value.TypeInteger}

	obj := value.NewObject(nil, words, types)

	if obj == nil {
		t.Fatal("NewObject returned nil")
	}

	if len(obj.Manifest.Words) != 2 {
		t.Errorf("expected 2 words, got %d", len(obj.Manifest.Words))
	}

	// Test String() output
	str := obj.String()
	if str == "" {
		t.Error("ObjectInstance.String() returned empty string")
	}
}

// TestObjectValueWrapping validates Value wrapping for objects
func TestObjectValueWrapping(t *testing.T) {
	obj := value.NewObject(nil, []string{"x"}, nil)
	val := value.ObjectVal(obj)

	if val.GetType() != value.TypeObject {
		t.Errorf("expected TypeObject, got %v", val.GetType())
	}

	// Test AsObject extraction
	extracted, ok := value.AsObject(val)
	if !ok {
		t.Error("AsObject returned false for object value")
	}
	if extracted == nil {
		t.Error("AsObject returned nil")
	}

	// Test wrong type
	intVal := value.NewIntVal(42)
	_, ok = value.AsObject(intVal)
	if ok {
		t.Error("AsObject returned true for integer value")
	}
}

// TestPortConstruction validates Port creation
func TestPortConstruction(t *testing.T) {
	// Port requires a driver, use nil for now (will implement drivers later)
	port := value.NewPort("file", "/tmp/test.txt", nil)

	if port == nil {
		t.Fatal("NewPort returned nil")
	}

	if port.Scheme != "file" {
		t.Errorf("expected scheme 'file', got %s", port.Scheme)
	}

	if port.State != value.PortClosed {
		t.Errorf("expected PortClosed, got %v", port.State)
	}

	// Test String() output
	str := port.String()
	if str == "" {
		t.Error("Port.String() returned empty string")
	}
}

// TestPortValueWrapping validates Value wrapping for ports
func TestPortValueWrapping(t *testing.T) {
	port := value.NewPort("tcp", "localhost:8080", nil)
	val := value.PortVal(port)

	if val.GetType() != value.TypePort {
		t.Errorf("expected TypePort, got %v", val.GetType())
	}

	// Test AsPort extraction
	extracted, ok := value.AsPort(val)
	if !ok {
		t.Error("AsPort returned false for port value")
	}
	if extracted == nil {
		t.Error("AsPort returned nil")
	}

	// Test wrong type
	intVal := value.NewIntVal(42)
	_, ok = value.AsPort(intVal)
	if ok {
		t.Error("AsPort returned true for integer value")
	}
}

// TestPathExpressionConstruction validates PathExpression creation
func TestPathExpressionConstruction(t *testing.T) {
	segments := []value.PathSegment{
		{Type: value.PathSegmentWord, Value: "user"},
		{Type: value.PathSegmentWord, Value: "address"},
		{Type: value.PathSegmentWord, Value: "city"},
	}
	base := value.NewWordVal("user")

	path := value.NewPath(segments, base)

	if path == nil {
		t.Fatal("NewPath returned nil")
	}

	if len(path.Segments) != 3 {
		t.Errorf("expected 3 segments, got %d", len(path.Segments))
	}

	// Test String() output
	str := path.String()
	if str == "" {
		t.Error("PathExpression.String() returned empty string")
	}
}

// TestPathValueWrapping validates Value wrapping for paths
func TestPathValueWrapping(t *testing.T) {
	segments := []value.PathSegment{
		{Type: value.PathSegmentWord, Value: "test"},
	}
	path := value.NewPath(segments, value.NewNoneVal())
	val := value.PathVal(path)

	if val.GetType() != value.TypePath {
		t.Errorf("expected TypePath, got %v", val.GetType())
	}

	// Test AsPath extraction
	extracted, ok := value.AsPath(val)
	if !ok {
		t.Error("AsPath returned false for path value")
	}
	if extracted == nil {
		t.Error("AsPath returned nil")
	}

	// Test wrong type
	intVal := value.NewIntVal(42)
	_, ok = value.AsPath(intVal)
	if ok {
		t.Error("AsPath returned true for integer value")
	}
}

// TestValueTypeDispatch validates type dispatch for new types
func TestValueTypeDispatch(t *testing.T) {
	// Create object with a field
	objFrame := frame.NewObjectFrame(0, []string{"name"}, []core.ValueType{value.TypeString})
	obj := value.NewObject(objFrame, []string{"name"}, []core.ValueType{value.TypeString})
	obj.SetField("name", value.NewStrVal("Alice"))

	tests := []struct {
		name     string
		val      core.Value
		expected core.ValueType
	}{
		{"decimal", value.DecimalVal(decimal.New(42, 0), 0), value.TypeDecimal},
		{"object", value.ObjectVal(obj), value.TypeObject},
		{"port", value.PortVal(value.NewPort("file", "test", nil)), value.TypePort},
		{"path", value.PathVal(value.NewPath([]value.PathSegment{{Type: value.PathSegmentWord, Value: "test"}}, value.NewNoneVal())), value.TypePath},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.val.GetType() != tt.expected {
				t.Errorf("expected type %v, got %v", tt.expected, tt.val.GetType())
			}

			// Verify Form() method works
			str := tt.val.Form()
			if str == "" {
				t.Errorf("Value.Form() returned empty for type %v", tt.expected)
			}
		})
	}
}

// TestInvalidConversions validates error cases for type assertions
func TestInvalidConversions(t *testing.T) {
	intVal := value.NewIntVal(42)

	// Test all AsXXX methods return false for wrong type
	if _, ok := value.AsDecimal(intVal); ok {
		t.Error("AsDecimal succeeded on integer")
	}
	if _, ok := value.AsObject(intVal); ok {
		t.Error("AsObject succeeded on integer")
	}
	if _, ok := value.AsPort(intVal); ok {
		t.Error("AsPort succeeded on integer")
	}
	if _, ok := value.AsPath(intVal); ok {
		t.Error("AsPath succeeded on integer")
	}
}
