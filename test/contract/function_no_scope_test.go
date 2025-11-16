package contract

import (
	"testing"

	"github.com/marcin-radoszewski/viro/internal/core"
	"github.com/marcin-radoszewski/viro/internal/parse"
	"github.com/marcin-radoszewski/viro/internal/value"
)

func TestFunction_NoScope(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected core.Value
		wantErr  bool
	}{
		{
			name: "mutation visibility - function modifies caller variable",
			code: `
				x: 1
				bump: fn --no-scope [] [x: x + 1]
				bump
				x
			`,
			expected: value.NewIntVal(2),
			wantErr:  false,
		},
		{
			name: "caller variable access - function can read caller locals",
			code: `
				y: 5
				add-to-y: fn --no-scope [z] [y + z]
				add-to-y 3
			`,
			expected: value.NewIntVal(8),
			wantErr:  false,
		},
		{
			name: "parameter restoration - parameter names don't remain bound",
			code: `
				a: 100
				temp: fn --no-scope [a] [a: 50]
				temp 200
				a
			`,
			expected: value.NewIntVal(100),
			wantErr:  false,
		},
		{
			name: "nested no-scope functions - dynamic scope demonstration",
			code: `
				outer-var: 10
				inner: fn --no-scope [] [outer-var + 1]
				outer: fn [] [
					outer-var: 20
					inner
				]
				outer
			`,
			expected: value.NewIntVal(21),
			wantErr:  false,
		},
		{
			name: "default behavior regression - normal fn still creates local scope",
			code: `
				x: 10
				normal: fn [] [x: 20]
				normal
				x
			`,
			expected: value.NewIntVal(10),
			wantErr:  false,
		},
		{
			name: "closure preservation - lexical parent still accessible",
			code: `
				make-adder: fn [base] [
					fn --no-scope [x] [base + x]
				]
				add5: make-adder 5
				add5 3
			`,
			expected: value.NewIntVal(8),
			wantErr:  false,
		},
		{
			name: "multiple assignments persist",
			code: `
				a: 1
				b: 2
				modify-both: fn --no-scope [] [
					a: 10
					b: 20
					c: 30
				]
				modify-both
				a + b + c
			`,
			expected: value.NewIntVal(60),
			wantErr:  false,
		},
		{
			name: "refinement parameters work with no-scope",
			code: `
				test-val: 0
				process: fn --no-scope [val --flag] [
					if flag [test-val: 1] [test-val: 0]
				]
				process 5 --flag
				test-val
			`,
			expected: value.NewIntVal(1),
			wantErr:  false,
		},
		{
			name: "new parameter names remain undefined after call",
			code: `
				existing: 100
				test: fn --no-scope [new-param] [
					new-param + 10
				]
				result: test 5
				existing
			`,
			expected: value.NewIntVal(100),
			wantErr:  false,
		},
		{
			name: "accessing new parameter after call fails",
			code: `
				test: fn --no-scope [new-param] [
					new-param + 10
				]
				result: test 5
				new-param
			`,
			expected: value.NewNoneVal(),
			wantErr:  true,
		},
		{
			name: "new refinement words remain undefined after call",
			code: `
				existing: 200
				test: fn --no-scope [val --new-flag] [
					if new-flag [val + 100] [val + 10]
				]
				result: test 5 --new-flag
				existing
			`,
			expected: value.NewIntVal(200),
			wantErr:  false,
		},
		{
			name: "accessing new refinement after call fails",
			code: `
				test: fn --no-scope [val --new-flag] [
					if new-flag [val + 100] [val + 10]
				]
				result: test 5 --new-flag
				new-flag
			`,
			expected: value.NewNoneVal(),
			wantErr:  true,
		},
		{
			name: "recursion works with parameter restoration",
			code: `
				counter: 0
				recurse: fn --no-scope [depth] [
					counter: counter + 1
					if depth > 0 [
						recurse depth - 1
					] [
						counter
					]
				]
				final: recurse 3
				depth
			`,
			expected: value.NewNoneVal(),
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := NewTestEvaluator()
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
