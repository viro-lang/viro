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
	vals, locations, err := parse.ParseWithSource(src, "(test)")
	if err != nil {
		return nil, value.NewNoneVal(), err
	}

	e := NewTestEvaluator()
	result, evalErr := e.DoBlock(vals, locations)
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
				bindings := obj.Frame.GetAll()
				if len(bindings) != 0 {
					t.Errorf("expected 0 fields, got %d", len(bindings))
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
				bindings := obj.Frame.GetAll()
				if len(bindings) != 2 {
					t.Errorf("expected 2 fields, got %d", len(bindings))
				}
				if bindings[0].Symbol != "name" || bindings[1].Symbol != "age" {
					t.Errorf("unexpected field names: %v %v", bindings[0].Symbol, bindings[1].Symbol)
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
				bindings := obj.Frame.GetAll()
				if len(bindings) != 2 {
					t.Errorf("expected 2 fields, got %d", len(bindings))
				}

				nameVal, found := obj.GetField("name")
				if !found {
					t.Error("field 'name' not found in frame")
				}
				if nameVal.GetType() != value.TypeString {
					t.Errorf("expected name to be string, got %v", value.TypeToString(nameVal.GetType()))
				}

				ageVal, found := obj.GetField("age")
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

				addrVal, found := obj.GetField("address")
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

				// Check city field in nested object
				cityVal, found := addrObj.GetField("city")
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

				deptVal, found := obj.GetField("dept")
				if !found || deptVal.GetType() != value.TypeObject {
					t.Fatal("dept not found or not an object")
				}

				deptObj, _ := value.AsObject(deptVal)
				teamVal, found := deptObj.GetField("team")
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
			name:       "path invokes no-arg function",
			code:       "obj: object [get-name: fn [] [\"Alice\"]] obj.get-name",
			expectType: value.TypeString,
			expectStr:  "Alice",
			wantErr:    false,
		},
		{
			name:       "get-path returns function value",
			code:       "obj: object [get-name: fn [] [\"Alice\"]] :obj.get-name",
			expectType: value.TypeFunction,
			wantErr:    false,
		},
		{
			name:       "nested path invokes no-arg function",
			code:       "user: object [profile: object [get-age: fn [] [30]]] user.profile.get-age",
			expectType: value.TypeInteger,
			expectStr:  "30",
			wantErr:    false,
		},
		{
			name:       "nested get-path returns function",
			code:       "user: object [profile: object [get-age: fn [] [30]]] :user.profile.get-age",
			expectType: value.TypeFunction,
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
				str := result.Form()
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
			code: `obj: object []
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

				nameVal, found := obj.GetField("name")
				if !found {
					t.Fatal("name field not found")
				}

				str := nameVal.Form()
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

				addrVal, found := user.GetField("address")
				if !found {
					t.Fatal("address not found")
				}

				addr, ok := value.AsObject(addrVal)
				if !ok {
					t.Fatal("address is not an object")
				}

				// Check city field in nested object
				cityVal, found := addr.GetField("city")
				if !found {
					t.Fatal("city not found")
				}

				str := cityVal.Form()
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
		{
			name:       "access string element by index",
			code:       "str: \"hello\" str.2",
			expectType: value.TypeString,
			expectStr:  "e",
			wantErr:    false,
		},
		{
			name:       "access first string element",
			code:       "str: \"world\" str.1",
			expectType: value.TypeString,
			expectStr:  "w",
			wantErr:    false,
		},
		{
			name:    "string index out of range",
			code:    "str: \"hi\" str.10",
			wantErr: true, // Should raise Script error (index-out-of-range)
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
				str := result.Form()
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
				str := result.Form()
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
			name:        "string index out of range",
			code:        "str: \"hi\" str.10",
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

// T087: path function invocation
func TestPathFunctionInvocation(t *testing.T) {
	tests := []struct {
		name       string
		code       string
		expectType core.ValueType
		expectStr  string
		wantErr    bool
	}{
		{
			name:       "path invokes no-arg function",
			code:       "obj: object [get-name: fn [] [\"Alice\"]] obj.get-name",
			expectType: value.TypeString,
			expectStr:  "Alice",
			wantErr:    false,
		},
		{
			name:       "get-path returns function value",
			code:       "obj: object [get-name: fn [] [\"Alice\"]] :obj.get-name",
			expectType: value.TypeFunction,
			wantErr:    false,
		},
		{
			name:       "nested path invokes no-arg function",
			code:       "user: object [profile: object [get-age: fn [] [30]]] user.profile.get-age",
			expectType: value.TypeInteger,
			expectStr:  "30",
			wantErr:    false,
		},
		{
			name:       "nested get-path returns function",
			code:       "user: object [profile: object [get-age: fn [] [30]]] :user.profile.get-age",
			expectType: value.TypeFunction,
			wantErr:    false,
		},
		{
			name:       "path with regular field access",
			code:       "obj: object [name: \"Bob\"] obj.name",
			expectType: value.TypeString,
			expectStr:  "Bob",
			wantErr:    false,
		},
		{
			name:       "path with function that has args (fails with no args)",
			code:       "obj: object [add: fn [a b] [a + b]] obj.add",
			expectType: value.TypeNone,
			wantErr:    true, // Should fail with arity error
		},
		{
			name:       "call function accessed by path with arguments",
			code:       "obj: object [add: fn [a b] [a + b]] obj.add 1 2",
			expectType: value.TypeInteger,
			expectStr:  "3",
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
				str := result.Form()
				if !strings.Contains(str, tt.expectStr) {
					t.Errorf("expected string to contain %q, got %q", tt.expectStr, str)
				}
			}
		})
	}
}

// T088: get-path evaluation
func TestGetPathEvaluation(t *testing.T) {
	tests := []struct {
		name       string
		code       string
		expectType core.ValueType
		expectStr  string
		wantErr    bool
	}{
		{
			name:       "get-path returns function without invocation",
			code:       "obj: object [get-name: fn [] [\"Alice\"]] :obj.get-name",
			expectType: value.TypeFunction,
			wantErr:    false,
		},
		{
			name:       "get-path on regular field",
			code:       "obj: object [name: \"Bob\"] :obj.name",
			expectType: value.TypeString,
			expectStr:  "Bob",
			wantErr:    false,
		},
		{
			name:       "get-path on nested function",
			code:       "user: object [profile: object [get-age: fn [] [30]]] :user.profile.get-age",
			expectType: value.TypeFunction,
			wantErr:    false,
		},
		{
			name:       "get-path on nested field",
			code:       "user: object [profile: object [name: \"Alice\"]] :user.profile.name",
			expectType: value.TypeString,
			expectStr:  "Alice",
			wantErr:    false,
		},
		{
			name:    "get-path on missing field",
			code:    "obj: object [x: 10] :obj.missing",
			wantErr: true, // Should raise Script error (no-such-field)
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
				str := result.Form()
				if !strings.Contains(str, tt.expectStr) {
					t.Errorf("expected string to contain %q, got %q", tt.expectStr, str)
				}
			}
		})
	}
}
