package contract_test

import (
	"testing"

	"github.com/marcin-radoszewski/viro/internal/core"
	"github.com/marcin-radoszewski/viro/internal/parse"
	"github.com/marcin-radoszewski/viro/internal/value"
	"github.com/marcin-radoszewski/viro/test/contract"
)

func TestRefinementWithNativeName(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected core.Value
		wantErr  bool
	}{
		{
			name: "refinement parameter named --debug",
			code: `
				test-fn: fn [x --debug] [
					if debug [x * 2] [x]
				]
				test-fn 5 --debug
			`,
			expected: value.NewIntVal(10),
			wantErr:  false,
		},
		{
			name: "refinement parameter named --debug without flag",
			code: `
				test-fn: fn [x --debug] [
					if debug [x * 2] [x]
				]
				test-fn 5
			`,
			expected: value.NewIntVal(5),
			wantErr:  false,
		},
		{
			name: "refinement parameter named --type",
			code: `
				test-fn: fn [val --type] [
					if type [type] [none]
				]
				test-fn 42 --type "integer"
			`,
			expected: value.NewStrVal("integer"),
			wantErr:  false,
		},
		{
			name: "refinement parameter named --print",
			code: `
				test-fn: fn [msg --print] [
					if print [msg] ["disabled"]
				]
				test-fn "hello" --print
			`,
			expected: value.NewStrVal("hello"),
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := contract.NewTestEvaluator()
			tokens, locations, parseErr := parse.ParseWithSource(tt.code, "(test)")
			if parseErr != nil {
				t.Fatalf("parse error: %v", parseErr)
			}

			result, evalErr := e.DoBlock(tokens, locations)

			if tt.wantErr {
				if evalErr == nil {
					t.Fatalf("expected error but got none")
				}
				return
			}

			if evalErr != nil {
				t.Fatalf("unexpected evaluation error: %v", evalErr)
			}
			if !result.Equals(tt.expected) {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestLocalVariableWithNativeName(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected core.Value
		wantErr  bool
	}{
		{
			name: "local variable named 'type' shadows native type?",
			code: `
				test-fn: fn [] [
					type: "custom-type"
					type
				]
				test-fn
			`,
			expected: value.NewStrVal("custom-type"),
			wantErr:  false,
		},
		{
			name: "local variable named 'print' shadows native print",
			code: `
				test-fn: fn [] [
					print: 42
					print
				]
				test-fn
			`,
			expected: value.NewIntVal(42),
			wantErr:  false,
		},
		{
			name: "local variable named 'if' shadows native if",
			code: `
				test-fn: fn [] [
					if: 100
					if
				]
				test-fn
			`,
			expected: value.NewIntVal(100),
			wantErr:  false,
		},
		{
			name: "local variable with native name, native still accessible in outer scope",
			code: `
				outer: fn [] [
					inner: fn [] [
						print: "shadowed"
						print
					]
					print "after inner"
					inner
				]
				outer
			`,
			expected: value.NewStrVal("shadowed"),
			wantErr:  false,
		},
		{
			name: "multiple local variables with native names",
			code: `
				test-fn: fn [] [
					print: 1
					type: 2
					debug: 3
					print + type + debug
				]
				test-fn
			`,
			expected: value.NewIntVal(6),
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := contract.NewTestEvaluator()
			tokens, locations, parseErr := parse.ParseWithSource(tt.code, "(test)")
			if parseErr != nil {
				t.Fatalf("parse error: %v", parseErr)
			}

			result, evalErr := e.DoBlock(tokens, locations)

			if tt.wantErr {
				if evalErr == nil {
					t.Fatalf("expected error but got none")
				}
				return
			}

			if evalErr != nil {
				t.Fatalf("unexpected evaluation error: %v", evalErr)
			}
			if !result.Equals(tt.expected) {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestNestedScopeShadowing(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected core.Value
		wantErr  bool
	}{
		{
			name: "3-level nested scopes with native name",
			code: `
				level1: fn [] [
					print: 1
					level2: fn [] [
						print: 2
						level3: fn [] [
							print: 3
							print
						]
						level3
					]
					level2
				]
				level1
			`,
			expected: value.NewIntVal(3),
			wantErr:  false,
		},
		{
			name: "nested scope sees parent binding when not shadowed",
			code: `
				outer: fn [] [
					x: 100
					inner: fn [] [
						x
					]
					inner
				]
				outer
			`,
			expected: value.NewIntVal(100),
			wantErr:  false,
		},
		{
			name: "nested scope shadows parent binding",
			code: `
				outer: fn [] [
					x: 100
					inner: fn [] [
						x: 200
						x
					]
					inner
				]
				outer
			`,
			expected: value.NewIntVal(200),
			wantErr:  false,
		},
		{
			name: "multiple levels of shadowing same native",
			code: `
				level1: fn [] [
					type: "level1"
					level2: fn [] [
						type: "level2"
						level3: fn [] [
							type: "level3"
							type
						]
						result: level3
						result
					]
					level2
				]
				level1
			`,
			expected: value.NewStrVal("level3"),
			wantErr:  false,
		},
		{
			name: "innermost binding wins at each level",
			code: `
				outer: fn [] [
					val: 1
					middle: fn [] [
						val: 2
						inner: fn [] [
							val: 3
							val
						]
						inner
					]
					middle
				]
				outer
			`,
			expected: value.NewIntVal(3),
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := contract.NewTestEvaluator()
			tokens, locations, parseErr := parse.ParseWithSource(tt.code, "(test)")
			if parseErr != nil {
				t.Fatalf("parse error: %v", parseErr)
			}

			result, evalErr := e.DoBlock(tokens, locations)

			if tt.wantErr {
				if evalErr == nil {
					t.Fatalf("expected error but got none")
				}
				return
			}

			if evalErr != nil {
				t.Fatalf("unexpected evaluation error: %v", evalErr)
			}
			if !result.Equals(tt.expected) {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestNativeFunctionsAccessible(t *testing.T) {
	e := contract.NewTestEvaluator()

	// Test that we can call native functions directly
	tests := []struct {
		name     string
		code     string
		expected core.Value
	}{
		{"math add", "3 + 4", value.NewIntVal(7)},
		{"math multiply", "5 * 6", value.NewIntVal(30)},
		{"print function", `print "test" "ok"`, value.NewStrVal("ok")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens, locations, parseErr := parse.ParseWithSource(tt.code, "(test)")
			if parseErr != nil {
				t.Fatalf("parse error: %v", parseErr)
			}

			result, evalErr := e.DoBlock(tokens, locations)
			if evalErr != nil {
				t.Fatalf("unexpected evaluation error: %v", evalErr)
			}
			if !result.Equals(tt.expected) {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}

	// Verify root frame contains natives
	rootFrame := e.GetFrameByIndex(0)
	if rootFrame == nil {
		t.Fatalf("root frame should exist")
	}

	// Check a few key natives are in root frame
	nativeNames := []string{"+", "-", "*", "/", "print", "type?", "fn", "if", "debug"}
	for _, name := range nativeNames {
		val, found := rootFrame.Get(name)
		if !found {
			t.Errorf("native '%s' should be in root frame", name)
		}
		if val.GetType() != value.TypeFunction {
			t.Errorf("native '%s' should be a function, got type %v", name, value.TypeToString(val.GetType()))
		}
	}
}
