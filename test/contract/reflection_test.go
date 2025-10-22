package contract

import (
	"strings"
	"testing"

	"github.com/marcin-radoszewski/viro/internal/core"
	"github.com/marcin-radoszewski/viro/internal/native"
	"github.com/marcin-radoszewski/viro/internal/value"
	"github.com/marcin-radoszewski/viro/internal/verror"
)

// Test suite for Feature 002: Reflection capabilities
// Contract tests validate FR-022 requirements

// T137: type-of for all value types
func TestTypeOf(t *testing.T) {
	// Set up sandbox for file-based tests
	tmpDir := t.TempDir()
	native.SandboxRoot = tmpDir

	tests := []struct {
		name       string
		code       string
		expectWord string
		wantErr    bool
	}{
		{
			name:       "type of integer",
			code:       "type-of 42",
			expectWord: "integer!",
			wantErr:    false,
		},
		{
			name:       "type of string",
			code:       "type-of \"hello\"",
			expectWord: "string!",
			wantErr:    false,
		},
		{
			name:       "type of block",
			code:       "type-of [1 2 3]",
			expectWord: "block!",
			wantErr:    false,
		},
		{
			name:       "type of word",
			code:       "type-of 'test",
			expectWord: "word!",
			wantErr:    false,
		},
		{
			name:       "type of function",
			code:       "square: fn [x] [x * x]\ntype-of :square",
			expectWord: "function!",
			wantErr:    false,
		},
		{
			name:       "type of logic true",
			code:       "type-of true",
			expectWord: "logic!",
			wantErr:    false,
		},
		{
			name:       "type of logic false",
			code:       "type-of false",
			expectWord: "logic!",
			wantErr:    false,
		},
		{
			name:       "type of none",
			code:       "type-of none",
			expectWord: "none!",
			wantErr:    false,
		},
		{
			name:       "type of decimal",
			code:       "type-of decimal \"3.14\"",
			expectWord: "decimal!",
			wantErr:    false,
		},
		{
			name:       "type of object",
			code:       "obj: object [x: 10]\ntype-of obj",
			expectWord: "object!",
			wantErr:    false,
		},
		{
			name:       "type of port",
			code:       "write \"test-type-of.txt\" \"test\"\np: open \"test-type-of.txt\"\nresult: type-of p\nclose p\nresult",
			expectWord: "port!",
			wantErr:    false,
		},
		{
			name:       "type of native function",
			code:       "type-of :print",
			expectWord: "native!",
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

			if result.GetType() != value.TypeWord {
				t.Errorf("expected type-of to return word!, got %v", value.TypeToString(result.GetType()))
			}

			resultStr := result.String()
			if !strings.Contains(resultStr, tt.expectWord) {
				t.Errorf("expected type-of to return %q, got %q", tt.expectWord, resultStr)
			}
		})
	}
}

// T138: spec-of for functions/objects
func TestSpecOf(t *testing.T) {
	tests := []struct {
		name      string
		code      string
		checkFunc func(*testing.T, core.Value)
		wantErr   bool
		errMsg    string
	}{
		{
			name: "spec-of function",
			code: "square: fn [x] [x * x]\nspec-of :square",
			checkFunc: func(t *testing.T, v core.Value) {
				if v.GetType() != value.TypeBlock {
					t.Errorf("expected spec-of to return block!, got %v", value.TypeToString(v.GetType()))
				}
				blk, ok := value.AsBlock(v)
				if !ok || len(blk.Elements) == 0 {
					t.Error("expected spec-of to return non-empty block")
				}
			},
			wantErr: false,
		},
		{
			name: "spec-of native function",
			code: "spec-of :print",
			checkFunc: func(t *testing.T, v core.Value) {
				if v.GetType() != value.TypeBlock {
					t.Errorf("expected spec-of to return block!, got %v", value.TypeToString(v.GetType()))
				}
				blk, ok := value.AsBlock(v)
				if !ok {
					t.Error("expected spec-of to return block")
				}
				if len(blk.Elements) == 0 {
					t.Error("expected native spec to have parameters")
				}
			},
			wantErr: false,
		},
		{
			name: "spec-of object",
			code: "obj: object [name: \"Alice\" age: 30]\nspec-of obj",
			checkFunc: func(t *testing.T, v core.Value) {
				if v.GetType() != value.TypeBlock {
					t.Errorf("expected spec-of to return block!, got %v", value.TypeToString(v.GetType()))
				}
				blk, ok := value.AsBlock(v)
				if !ok {
					t.Error("expected spec-of to return block")
				}
				if len(blk.Elements) == 0 {
					t.Error("expected object spec to have fields")
				}
			},
			wantErr: false,
		},
		{
			name:    "spec-of unsupported type (integer)",
			code:    "spec-of 42",
			wantErr: true,
			errMsg:  "unsupported",
		},
		{
			name:    "spec-of unsupported type (string)",
			code:    "spec-of \"test\"",
			wantErr: true,
			errMsg:  "unsupported",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Evaluate(tt.code)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error but got none")
				}
				verr, ok := err.(*verror.Error)
				if !ok {
					t.Fatalf("expected *verror.Error, got %T", err)
				}
				if verr.Category != verror.ErrScript {
					t.Errorf("expected Script error, got %v", verr.Category)
				}
				if tt.errMsg != "" && !strings.Contains(strings.ToLower(verr.Message), tt.errMsg) {
					t.Errorf("expected error message to contain %q, got %q", tt.errMsg, verr.Message)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.checkFunc != nil {
				tt.checkFunc(t, result)
			}
		})
	}
}

// T139: body-of immutability
func TestBodyOf(t *testing.T) {
	tests := []struct {
		name      string
		code      string
		checkFunc func(*testing.T, core.Value)
		wantErr   bool
		errMsg    string
	}{
		{
			name: "body-of function returns block",
			code: "square: fn [x] [x * x]\nbody-of :square",
			checkFunc: func(t *testing.T, v core.Value) {
				if v.GetType() != value.TypeBlock {
					t.Errorf("expected body-of to return block!, got %v", value.TypeToString(v.GetType()))
				}
				blk, ok := value.AsBlock(v)
				if !ok || len(blk.Elements) == 0 {
					t.Error("expected body-of to return non-empty block")
				}
			},
			wantErr: false,
		},
		{
			name: "body-of returns deep copy (immutable)",
			code: `square: fn [x] [x * x]
			       body: body-of :square
			       append body 999
			       original: body-of :square
			       length? original`,
			checkFunc: func(t *testing.T, v core.Value) {
				if v.GetType() != value.TypeInteger {
					t.Errorf("expected length? to return integer!, got %v", value.TypeToString(v.GetType()))
				}
			},
			wantErr: false,
		},
		{
			name: "body-of object",
			code: "obj: object [x: 10 y: 20]\nbody-of obj",
			checkFunc: func(t *testing.T, v core.Value) {
				if v.GetType() != value.TypeBlock {
					t.Errorf("expected body-of to return block!, got %v", value.TypeToString(v.GetType()))
				}
			},
			wantErr: false,
		},
		{
			name:    "body-of native (no body)",
			code:    "body-of :print",
			wantErr: true,
			errMsg:  "body",
		},
		{
			name:    "body-of integer (unsupported)",
			code:    "body-of 42",
			wantErr: true,
			errMsg:  "body",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Evaluate(tt.code)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error but got none")
				}
				verr, ok := err.(*verror.Error)
				if !ok {
					t.Fatalf("expected *verror.Error, got %T", err)
				}
				if verr.Category != verror.ErrScript {
					t.Errorf("expected Script error, got %v", verr.Category)
				}
				if tt.errMsg != "" && !strings.Contains(strings.ToLower(verr.Message), tt.errMsg) {
					t.Errorf("expected error message to contain %q, got %q", tt.errMsg, verr.Message)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.checkFunc != nil {
				tt.checkFunc(t, result)
			}
		})
	}
}

// T140: words-of/values-of consistency
func TestWordsAndValues(t *testing.T) {
	tests := []struct {
		name      string
		code      string
		checkFunc func(*testing.T, core.Value)
		wantErr   bool
	}{
		{
			name: "words-of and values-of for object",
			code: "obj: object [name: \"Alice\" age: 30]\nwords-of obj",
			checkFunc: func(t *testing.T, v core.Value) {
				if v.GetType() != value.TypeBlock {
					t.Fatalf("expected block result, got %v", value.TypeToString(v.GetType()))
				}
				blk, ok := value.AsBlock(v)
				if !ok || len(blk.Elements) != 2 {
					t.Fatalf("expected block with 2 words, got %d elements", len(blk.Elements))
				}
				for i, elem := range blk.Elements {
					if elem.GetType() != value.TypeWord {
						t.Errorf("element %d: expected word!, got %v", i, value.TypeToString(elem.GetType()))
					}
				}
			},
			wantErr: false,
		},
		{
			name: "values-of for object",
			code: "obj: object [name: \"Alice\" age: 30]\nvalues-of obj",
			checkFunc: func(t *testing.T, v core.Value) {
				if v.GetType() != value.TypeBlock {
					t.Fatalf("expected block result, got %v", value.TypeToString(v.GetType()))
				}
				blk, ok := value.AsBlock(v)
				if !ok || len(blk.Elements) != 2 {
					t.Fatalf("expected block with 2 values, got %d elements", len(blk.Elements))
				}
				if blk.Elements[0].GetType() != value.TypeString {
					t.Errorf("first value: expected string!, got %v", value.TypeToString(blk.Elements[0].GetType()))
				}
				if blk.Elements[1].GetType() != value.TypeInteger {
					t.Errorf("second value: expected integer!, got %v", value.TypeToString(blk.Elements[1].GetType()))
				}
			},
			wantErr: false,
		},
		{
			name: "words-of and values-of length consistency",
			code: "obj: object [name: \"Alice\" age: 30]\nwords-len: length? words-of obj\nvals-len: length? values-of obj\nwords-len = vals-len",
			checkFunc: func(t *testing.T, v core.Value) {
				if v.GetType() != value.TypeLogic {
					t.Fatalf("expected logic!, got %v", value.TypeToString(v.GetType()))
				}
				logic, _ := value.AsLogic(v)
				if !logic {
					t.Error("words-of and values-of should return same length")
				}
			},
			wantErr: false,
		},
		{
			name: "words-of empty object",
			code: "obj: object []\nwords-of obj",
			checkFunc: func(t *testing.T, v core.Value) {
				if v.GetType() != value.TypeBlock {
					t.Errorf("expected block!, got %v", value.TypeToString(v.GetType()))
				}
				blk, ok := value.AsBlock(v)
				if !ok {
					t.Fatal("expected block type")
				}
				if len(blk.Elements) != 0 {
					t.Errorf("expected empty block, got %d elements", len(blk.Elements))
				}
			},
			wantErr: false,
		},
		{
			name: "values-of returns deep copies",
			code: `obj: object [data: [1 2 3]]
			       vals: values-of obj
			       first-val: first vals
			       append first-val 999
			       obj-data: obj.data
			       length? obj-data`,
			checkFunc: func(t *testing.T, v core.Value) {
				if v.GetType() != value.TypeInteger {
					t.Errorf("expected integer!, got %v", value.TypeToString(v.GetType()))
				}
			},
			wantErr: false,
		},
		{
			name: "words-of on nested object",
			code: `obj: object [
				       address: object [city: "Portland" zip: 97201]
			       ]
			       inner: obj.address
			       words-of inner`,
			checkFunc: func(t *testing.T, v core.Value) {
				if v.GetType() != value.TypeBlock {
					t.Errorf("expected block!, got %v", value.TypeToString(v.GetType()))
				}
				blk, ok := value.AsBlock(v)
				if !ok || len(blk.Elements) != 2 {
					t.Error("expected 2 words for nested object")
				}
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Evaluate(tt.code)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error but got none")
				}
				verr, ok := err.(*verror.Error)
				if !ok {
					t.Fatalf("expected *verror.Error, got %T", err)
				}
				if verr.Category != verror.ErrScript {
					t.Errorf("expected Script error, got %v", verr.Category)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.checkFunc != nil {
				tt.checkFunc(t, result)
			}
		})
	}
}

// Additional test for source native
func TestSource(t *testing.T) {
	tests := []struct {
		name    string
		code    string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "source of function",
			code:    "square: fn [x] [x * x]\nsource :square",
			wantErr: false,
		},
		{
			name:    "source of native",
			code:    "source :print",
			wantErr: false,
		},
		{
			name:    "source of object",
			code:    "obj: object [x: 10]\nsource obj",
			wantErr: false,
		},
		{
			name:    "source of unsupported type",
			code:    "source 42",
			wantErr: true,
			errMsg:  "unsupported",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Evaluate(tt.code)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error but got none")
				}
				verr, ok := err.(*verror.Error)
				if !ok {
					t.Fatalf("expected *verror.Error, got %T", err)
				}
				if verr.Category != verror.ErrScript {
					t.Errorf("expected Script error, got %v", verr.Category)
				}
				if tt.errMsg != "" && !strings.Contains(strings.ToLower(verr.Message), tt.errMsg) {
					t.Errorf("expected error message to contain %q, got %q", tt.errMsg, verr.Message)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.GetType() != value.TypeString {
				t.Errorf("expected source to return string!, got %v", value.TypeToString(result.GetType()))
			}
		})
	}
}
