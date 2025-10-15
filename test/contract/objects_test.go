package contract

import (
	"strings"
	"testing"

	"github.com/marcin-radoszewski/viro/internal/core"
	"github.com/marcin-radoszewski/viro/internal/parse"
	"github.com/marcin-radoszewski/viro/internal/value"
	"github.com/marcin-radoszewski/viro/internal/verror"
)

// Test suite for Feature 002: Objects and path expressions
// Contract tests validate FR-009 through FR-011 requirements
// These tests follow TDD: they MUST FAIL initially before implementation

func evalObjectScriptWithEvaluator(src string) (core.Evaluator, core.Value, error) {
	vals, err := parse.Parse(src)
	if err != nil {
		return nil, value.NoneVal(), err
	}

	e := NewTestEvaluator()
	result, evalErr := e.Do_Blk(vals)
	return e, result, evalErr
}

// T080: object construction with field initialization
func TestObjectConstruction(t *testing.T) {
	tests := []struct {
		name        string
		code        string
		expectType  core.ValueType
		checkFields func(*testing.T, core.Value, core.Evaluator)
		wantErr     bool
	}{
		{
			name:       "empty object",
			code:       "obj: object [] obj",
			expectType: value.TypeObject,
			checkFields: func(t *testing.T, v core.Value, e core.Evaluator) {
				obj, ok := value.AsObject(v)
				if !ok {
					t.Fatal("expected object type")
				}
				if len(obj.Manifest.Words) != 0 {
					t.Errorf("expected 0 fields, got %d", len(obj.Manifest.Words))
				}
			},
			wantErr: false,
		},
		{
			name:       "simple fields without values",
			code:       "obj: object [name age] obj",
			expectType: value.TypeObject,
			checkFields: func(t *testing.T, v core.Value, e core.Evaluator) {
				obj, ok := value.AsObject(v)
				if !ok {
					t.Fatal("expected object type")
				}
				if len(obj.Manifest.Words) != 2 {
					t.Errorf("expected 2 fields, got %d", len(obj.Manifest.Words))
				}
				if obj.Manifest.Words[0] != "name" || obj.Manifest.Words[1] != "age" {
					t.Errorf("unexpected field names: %v", obj.Manifest.Words)
				}
			},
			wantErr: false,
		},
		{
			name:       "fields with initialization",
			code:       "obj: object [name: \"Alice\" age: 30] obj",
			expectType: value.TypeObject,
			checkFields: func(t *testing.T, v core.Value, e core.Evaluator) {
				obj, ok := value.AsObject(v)
				if !ok {
					t.Fatal("expected object type")
				}
				if len(obj.Manifest.Words) != 2 {
					t.Errorf("expected 2 fields, got %d", len(obj.Manifest.Words))
				}

				objFrame := e.GetFrameByIndex(obj.FrameIndex)
				if objFrame == nil {
					t.Fatalf("invalid frame index: %d", obj.FrameIndex)
				}

				nameVal, found := objFrame.Get("name")
				if !found {
					t.Error("field 'name' not found in frame")
				}
				if nameVal.GetType() != value.TypeString {
					t.Errorf("expected name to be string, got %v", value.TypeToString(nameVal.GetType()))
				}

				ageVal, found := objFrame.Get("age")
				if !found {
					t.Error("field 'age' not found in frame")
				}
				if ageVal.GetType() != value.TypeInteger {
					t.Errorf("expected age to be integer, got %v", value.TypeToString(ageVal.GetType()))
				}
			},
			wantErr: false,
		},
		{
			name:       "duplicate field names",
			code:       "object [name: \"Alice\" name: \"Bob\"]",
			expectType: value.TypeNone,
			wantErr:    true, // Should raise Script error (object-field-duplicate)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e, result, err := evalObjectScriptWithEvaluator(tt.code)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error but got none")
				}
				if verr, ok := err.(*verror.Error); ok {
					if verr.Category != verror.ErrScript {
						t.Errorf("expected Script error, got %v", verr.Category)
					}
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.GetType() != tt.expectType {
				t.Errorf("expected type %v, got %v", value.TypeToString(tt.expectType), value.TypeToString(result.GetType()))
			}

			if tt.checkFields != nil {
				tt.checkFields(t, result, e)
			}
		})
	}
}

// T081: nested object creation
func TestNestedObjects(t *testing.T) {
	tests := []struct {
		name       string
		code       string
		checkField func(*testing.T, core.Value, core.Evaluator)
		wantErr    bool
	}{
		{
			name: "nested object in field",
			code: `user: object [
				name: "Alice"
				address: object [
					city: "Portland"
					zip: 97201
				]
			]
			user`,
			checkField: func(t *testing.T, v core.Value, e core.Evaluator) {
				obj, ok := value.AsObject(v)
				if !ok {
					t.Fatal("expected object type")
				}

				objFrame := e.GetFrameByIndex(obj.FrameIndex)
				if objFrame == nil {
					t.Fatalf("invalid frame index: %d", obj.FrameIndex)
				}
				addrVal, found := objFrame.Get("address")
				if !found {
					t.Fatal("address field not found")
				}

				if addrVal.GetType() != value.TypeObject {
					t.Errorf("expected address to be object, got %v", value.TypeToString(addrVal.GetType()))
				}

				addrObj, ok := value.AsObject(addrVal)
				if !ok {
					t.Fatal("address is not an object")
				}

				addrFrame := e.GetFrameByIndex(addrObj.FrameIndex)
				if addrFrame == nil {
					t.Fatalf("invalid frame index: %d", addrObj.FrameIndex)
				}
				cityVal, found := addrFrame.Get("city")
				if !found {
					t.Error("city field not found in nested object")
				}
				if cityVal.GetType() != value.TypeString {
					t.Errorf("expected city to be string, got %v", value.TypeToString(cityVal.GetType()))
				}
			},
			wantErr: false,
		},
		{
			name: "three-level nesting",
			code: `org: object [
				dept: object [
					team: object [
						name: "Engineering"
					]
				]
			]
			org`,
			checkField: func(t *testing.T, v core.Value, e core.Evaluator) {
				obj, ok := value.AsObject(v)
				if !ok {
					t.Fatal("expected object type")
				}

				frame1 := e.GetFrameByIndex(obj.FrameIndex)
				if frame1 == nil {
					t.Fatalf("invalid frame index: %d", obj.FrameIndex)
				}
				deptVal, found := frame1.Get("dept")
				if !found || deptVal.GetType() != value.TypeObject {
					t.Fatal("dept not found or not an object")
				}

				deptObj, _ := value.AsObject(deptVal)
				frame2 := e.GetFrameByIndex(deptObj.FrameIndex)
				if frame2 == nil {
					t.Fatalf("invalid frame index: %d", deptObj.FrameIndex)
				}
				teamVal, found := frame2.Get("team")
				if !found || teamVal.GetType() != value.TypeObject {
					t.Fatal("team not found or not an object")
				}
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e, result, err := evalObjectScriptWithEvaluator(tt.code)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.checkField != nil {
				tt.checkField(t, result, e)
			}
		})
	}
}

// T082: path read traversal (object.field.subfield)
func TestPathReadTraversal(t *testing.T) {
	tests := []struct {
		name       string
		code       string
		expectType core.ValueType
		expectStr  string // Expected string representation or partial match
		wantErr    bool
	}{
		{
			name:       "simple field access",
			code:       "obj: object [name: \"Alice\"] obj.name",
			expectType: value.TypeString,
			expectStr:  "Alice",
			wantErr:    false,
		},
		{
			name:       "nested field access",
			code:       "user: object [address: object [city: \"Portland\"]] user.address.city",
			expectType: value.TypeString,
			expectStr:  "Portland",
			wantErr:    false,
		},
		{
			name: "three-level path",
			code: `org: object [
				dept: object [
					team: object [name: "Engineering"]
				]
			]
			org.dept.team.name`,
			expectType: value.TypeString,
			expectStr:  "Engineering",
			wantErr:    false,
		},
		{
			name:    "missing field",
			code:    "obj: object [name: \"Alice\"] obj.age",
			wantErr: true, // Should raise Script error (no-such-field)
		},
		{
			name:    "none-path error",
			code:    "obj: object [data: none] obj.data.field",
			wantErr: true, // Should raise Script error (none-path)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Evaluate(tt.code)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error but got none")
				}
				if verr, ok := err.(*verror.Error); ok {
					if verr.Category != verror.ErrScript {
						t.Errorf("expected Script error, got %v", verr.Category)
					}
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.GetType() != tt.expectType {
				t.Errorf("expected type %v, got %v", value.TypeToString(tt.expectType), value.TypeToString(result.GetType()))
			}

			if tt.expectStr != "" {
				str := result.String()
				if !strings.Contains(str, tt.expectStr) {
					t.Errorf("expected string to contain %q, got %q", tt.expectStr, str)
				}
			}
		})
	}
}

// T083: path write mutation
func TestPathWriteMutation(t *testing.T) {
	tests := []struct {
		name      string
		code      string
		checkFunc func(*testing.T, core.Evaluator)
		wantErr   bool
	}{
		{
			name: "update object field via path",
			code: `obj: object [name: "Alice" age: 30]
			       obj.name: "Bob"
			       obj`,
			checkFunc: func(t *testing.T, e core.Evaluator) {
				rootFrame := e.GetFrameByIndex(0)
				objVal, found := rootFrame.Get("obj")
				if !found {
					t.Fatal("obj not found")
				}
				obj, ok := value.AsObject(objVal)
				if !ok {
					t.Fatal("obj is not an object")
				}

				objFrame := e.GetFrameByIndex(obj.FrameIndex)
				if objFrame == nil {
					t.Fatalf("invalid frame index: %d", obj.FrameIndex)
				}
				nameVal, found := objFrame.Get("name")
				if !found {
					t.Fatal("name field not found")
				}

				str := nameVal.String()
				if !strings.Contains(str, "Bob") {
					t.Errorf("expected name to be Bob, got %s", str)
				}
			},
			wantErr: false,
		},
		{
			name: "update nested field via path",
			code: `user: object [address: object [city: "Portland"]]
			       user.address.city: "Seattle"
			       user`,
			checkFunc: func(t *testing.T, e core.Evaluator) {
				rootFrame := e.GetFrameByIndex(0)
				userVal, found := rootFrame.Get("user")
				if !found {
					t.Fatal("user not found")
				}
				user, ok := value.AsObject(userVal)
				if !ok {
					t.Fatal("user is not an object")
				}

				userFrame := e.GetFrameByIndex(user.FrameIndex)
				if userFrame == nil {
					t.Fatalf("invalid frame index: %d", user.FrameIndex)
				}
				addrVal, found := userFrame.Get("address")
				if !found {
					t.Fatal("address not found")
				}

				addr, ok := value.AsObject(addrVal)
				if !ok {
					t.Fatal("address is not an object")
				}

				addrFrame := e.GetFrameByIndex(addr.FrameIndex)
				if addrFrame == nil {
					t.Fatalf("invalid frame index: %d", addr.FrameIndex)
				}
				cityVal, found := addrFrame.Get("city")
				if !found {
					t.Fatal("city not found")
				}

				str := cityVal.String()
				if !strings.Contains(str, "Seattle") {
					t.Errorf("expected city to be Seattle, got %s", str)
				}
			},
			wantErr: false,
		},
		{
			name:    "assign to immutable literal",
			code:    "42: 100",
			wantErr: true, // Should raise Script error (immutable-target)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e, _, err := evalObjectScriptWithEvaluator(tt.code)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error but got none")
				}
				if verr, ok := err.(interface{ Category() verror.ErrorCategory }); ok {
					if verr.Category() != verror.ErrScript && verr.Category() != verror.ErrSyntax {
						t.Errorf("expected Script or Syntax error, got %v", verr.Category())
					}
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.checkFunc != nil {
				tt.checkFunc(t, e)
			}
		})
	}
}

// T084: path indexing for blocks (block.3)
func TestPathIndexing(t *testing.T) {
	tests := []struct {
		name       string
		code       string
		expectType core.ValueType
		expectStr  string
		wantErr    bool
	}{
		{
			name:       "access block element by index",
			code:       "data: [10 20 30] data.2",
			expectType: value.TypeInteger,
			expectStr:  "20",
			wantErr:    false,
		},
		{
			name:       "access first element",
			code:       "data: [100 200 300] data.1",
			expectType: value.TypeInteger,
			expectStr:  "100",
			wantErr:    false,
		},
		{
			name:    "index out of range",
			code:    "data: [10 20] data.5",
			wantErr: true, // Should raise Script error (index-out-of-range)
		},
		{
			name:    "zero index (1-based indexing)",
			code:    "data: [10 20 30] data.0",
			wantErr: true, // Should raise error (invalid index)
		},
		{
			name:       "nested block access",
			code:       "matrix: [[1 2] [3 4] [5 6]] matrix.2",
			expectType: value.TypeBlock,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Evaluate(tt.code)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.GetType() != tt.expectType {
				t.Errorf("expected type %v, got %v", value.TypeToString(tt.expectType), value.TypeToString(result.GetType()))
			}

			if tt.expectStr != "" {
				str := result.String()
				if !strings.Contains(str, tt.expectStr) {
					t.Errorf("expected string to contain %q, got %q", tt.expectStr, str)
				}
			}
		})
	}
}

// T085: parent prototype lookup
func TestParentPrototype(t *testing.T) {
	tests := []struct {
		name       string
		code       string
		expectType core.ValueType
		expectStr  string
		wantErr    bool
	}{
		{
			name: "inherit field from parent",
			code: `base: make object! [x: 10 y: 20]
			       derived: make base [z: 30]
			       derived.x`,
			expectType: value.TypeInteger,
			expectStr:  "10",
			wantErr:    false,
		},
		{
			name: "override parent field",
			code: `base: make object! [name: "Base"]
			       derived: make base [name: "Derived"]
			       derived.name`,
			expectType: value.TypeString,
			expectStr:  "Derived",
			wantErr:    false,
		},
		{
			name: "multi-level inheritance",
			code: `level1: make object! [a: 1]
			       level2: make level1 [b: 2]
			       level3: make level2 [c: 3]
			       level3.a`,
			expectType: value.TypeInteger,
			expectStr:  "1",
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Evaluate(tt.code)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.GetType() != tt.expectType {
				t.Errorf("expected type %v, got %v", value.TypeToString(tt.expectType), value.TypeToString(result.GetType()))
			}

			if tt.expectStr != "" {
				str := result.String()
				if !strings.Contains(str, tt.expectStr) {
					t.Errorf("expected string to contain %q, got %q", tt.expectStr, str)
				}
			}
		})
	}
}

func TestMakePrototypeErrors(t *testing.T) {
	tests := []struct {
		name    string
		code    string
		errorID string
	}{
		{
			name:    "non-object target",
			code:    "make 10 [x]",
			errorID: verror.ErrIDTypeMismatch,
		},
		{
			name:    "parent field forbidden",
			code:    "make object! [parent: none]",
			errorID: verror.ErrIDReservedField,
		},
		{
			name:    "spec reserved field",
			code:    "make object! [spec: 1]",
			errorID: verror.ErrIDReservedField,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Evaluate(tt.code)
			if err == nil {
				t.Fatalf("expected error but got none")
			}
			if verr, ok := err.(interface{ ID() string }); ok {
				if verr.ID() != tt.errorID {
					t.Fatalf("expected error id %s, got %s", tt.errorID, verr.ID())
				}
			}
		})
	}
}

// T086: path error handling (none-path, index-out-of-range)
func TestPathErrorHandling(t *testing.T) {
	tests := []struct {
		name        string
		code        string
		expectCat   verror.ErrorCategory
		expectToken string // Expected error message token
	}{
		{
			name:        "none-path error",
			code:        "obj: object [data: none] obj.data.field",
			expectCat:   verror.ErrScript,
			expectToken: "none",
		},
		{
			name:        "no-such-field error",
			code:        "obj: object [x: 10] obj.y",
			expectCat:   verror.ErrScript,
			expectToken: "field",
		},
		{
			name:        "index out of range",
			code:        "data: [1 2 3] data.10",
			expectCat:   verror.ErrScript,
			expectToken: "range",
		},
		{
			name:        "path on non-object non-series",
			code:        "x: 42 x.field",
			expectCat:   verror.ErrScript,
			expectToken: "mismatch",
		},
		{
			name:        "immutable target assignment",
			code:        "1.field: 100",
			expectCat:   verror.ErrScript,
			expectToken: "immutable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Evaluate(tt.code)

			if err == nil {
				t.Fatal("expected error but got none")
			}

			if verr, ok := err.(interface {
				Category() verror.ErrorCategory
				Error() string
			}); ok {
				if verr.Category() != tt.expectCat {
					t.Errorf("expected category %v, got %v", tt.expectCat, verr.Category())
				}

				errMsg := verr.Error()
				if !strings.Contains(strings.ToLower(errMsg), strings.ToLower(tt.expectToken)) {
					t.Errorf("expected error message to contain %q, got %q", tt.expectToken, errMsg)
				}
			}
		})
	}
}
