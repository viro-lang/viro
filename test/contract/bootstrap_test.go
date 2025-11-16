package contract

import (
	"testing"

	"github.com/marcin-radoszewski/viro/internal/core"
	"github.com/marcin-radoszewski/viro/internal/value"
)

func TestBootstrapScriptsLoaded(t *testing.T) {
	evaluator := NewTestEvaluator()

	importFunc, ok := evaluator.GetFrameByIndex(0).Get("import")
	if !ok {
		t.Fatalf("bootstrap script 'import' function not found")
	}

	if importFunc.GetType() != value.TypeFunction {
		t.Errorf("expected 'import' to be a function, got %s", value.TypeToString(importFunc.GetType()))
	}
}

func TestBootstrapScriptsExecuted(t *testing.T) {
	evaluator := NewTestEvaluator()
	importFunc, ok := evaluator.GetFrameByIndex(0).Get("import")
	if !ok {
		t.Fatalf("bootstrap script 'import' function not found")
	}

	if importFunc.GetType() != value.TypeFunction {
		t.Errorf("expected 'import' to be a function, got %s", value.TypeToString(importFunc.GetType()))
	}
}

func TestBootstrapAnyFunction_NoScope(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected core.Value
		wantErr  bool
	}{
		{
			name: "any can see caller parameter",
			code: `
				test: fn [foo] [any [foo]]
				test true
			`,
			expected: value.NewLogicVal(true),
			wantErr:  false,
		},
		{
			name: "any can see caller refinement flag",
			code: `
				test: fn [--flag] [any [flag]]
				test --flag
			`,
			expected: value.NewLogicVal(true),
			wantErr:  false,
		},
		{
			name: "any can see caller refinement value",
			code: `
				test: fn [--refine val] [
					result: none
					any [result: val]
					result
				]
				test --refine 42
			`,
			expected: value.NewIntVal(42),
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Evaluate(tt.code)

			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected evaluation error: %v", err)
			}
			if !result.Equals(tt.expected) {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestBootstrapAllFunction_NoScope(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected core.Value
		wantErr  bool
	}{
		{
			name: "all can see caller parameter",
			code: `
				test: fn [foo] [all [foo foo]]
				test true
			`,
			expected: value.NewLogicVal(true),
			wantErr:  false,
		},
		{
			name: "all can see caller refinement flag",
			code: `
				test: fn [--flag] [all [flag flag]]
				test --flag
			`,
			expected: value.NewLogicVal(true),
			wantErr:  false,
		},
		{
			name: "all can see caller refinement value",
			code: `
				test: fn [--refine val] [
					result: none
					all [result: val]
					result
				]
				test --refine 42
			`,
			expected: value.NewIntVal(42),
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Evaluate(tt.code)

			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected evaluation error: %v", err)
			}
			if !result.Equals(tt.expected) {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}
