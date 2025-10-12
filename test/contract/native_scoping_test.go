package contract_test

import (
	"testing"

	"github.com/marcin-radoszewski/viro/internal/eval"
	"github.com/marcin-radoszewski/viro/internal/native"
	"github.com/marcin-radoszewski/viro/internal/parse"
	"github.com/marcin-radoszewski/viro/internal/value"
)

// T009: Test that refinements can use native names without errors
// This validates User Story 1: Define Functions with Names Matching Natives
func TestRefinementWithNativeName(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected value.Value
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
			expected: value.IntVal(10),
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
			expected: value.IntVal(5),
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
			expected: value.StrVal("integer"),
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
			expected: value.StrVal("hello"),
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := eval.NewEvaluator()
			tokens, parseErr := parse.Parse(tt.code)
			if parseErr != nil {
				t.Fatalf("parse error: %v", parseErr)
			}

			result, evalErr := e.Do_Blk(tokens)

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

// T009: Test that local variables can use native names without conflicts
func TestLocalVariableWithNativeName(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected value.Value
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
			expected: value.StrVal("custom-type"),
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
			expected: value.IntVal(42),
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
			expected: value.IntVal(100),
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
			expected: value.StrVal("shadowed"),
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
			expected: value.IntVal(6),
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := eval.NewEvaluator()
			tokens, parseErr := parse.Parse(tt.code)
			if parseErr != nil {
				t.Fatalf("parse error: %v", parseErr)
			}

			result, evalErr := e.Do_Blk(tokens)

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

// T010: Test that nested scopes follow lexical scoping rules
func TestNestedScopeShadowing(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected value.Value
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
			expected: value.IntVal(3),
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
			expected: value.IntVal(100),
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
			expected: value.IntVal(200),
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
			expected: value.StrVal("level3"),
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
			expected: value.IntVal(3),
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := eval.NewEvaluator()
			tokens, parseErr := parse.Parse(tt.code)
			if parseErr != nil {
				t.Fatalf("parse error: %v", parseErr)
			}

			result, evalErr := e.Do_Blk(tokens)

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

// Verify that native functions are still accessible via the registry
// This test should pass initially (with registry), then fail when we remove native.Lookup(),
// then pass again after we implement frame-based lookup
func TestNativeFunctionsAccessible(t *testing.T) {
	e := eval.NewEvaluator()

	// Test that we can call native functions directly
	tests := []struct {
		name     string
		code     string
		expected value.Value
	}{
		{"math add", "3 + 4", value.IntVal(7)},
		{"math multiply", "5 * 6", value.IntVal(30)},
		{"print function", `print "test" "ok"`, value.StrVal("ok")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens, parseErr := parse.Parse(tt.code)
			if parseErr != nil {
				t.Fatalf("parse error: %v", parseErr)
			}

			result, evalErr := e.Do_Blk(tokens)
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
		if val.Type != value.TypeFunction {
			t.Errorf("native '%s' should be a function, got type %v", name, val.Type)
		}
	}
}

// Helper function to verify that a native function exists in the registry
// This is used during the transition phase to ensure backward compatibility
func TestNativeRegistryPopulated(t *testing.T) {
	// This test verifies Phase 1 behavior: registry is still populated
	// It will be removed in Phase 3 when registry is deleted

	nativeNames := []string{"+", "-", "*", "/", "print", "type?", "fn", "if", "debug"}
	for _, name := range nativeNames {
		fn, found := native.Lookup(name)
		if !found {
			t.Errorf("native '%s' should be in registry", name)
		}
		if fn == nil {
			t.Errorf("native '%s' should have non-nil function", name)
		}
	}
}
