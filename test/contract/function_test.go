package contract

import (
	"fmt"
	"testing"

	"github.com/marcin-radoszewski/viro/internal/core"
	"github.com/marcin-radoszewski/viro/internal/eval"
	"github.com/marcin-radoszewski/viro/internal/parse"
	"github.com/marcin-radoszewski/viro/internal/value"
)

func TestFunction_Definition(t *testing.T) {
	t.Run("captures parameters and body", func(t *testing.T) {
		result, err := Evaluate("fn [name --title [] --verbose] [(print name)]")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		fn, ok := value.AsFunctionValue(result)
		if !ok {
			t.Fatalf("expected function value, got %v", result)
		}

		if len(fn.Params) != 3 {
			t.Fatalf("expected 3 params, got %d", len(fn.Params))
		}

		if fn.Params[0].Name != "name" || fn.Params[0].Refinement {
			t.Fatalf("expected first param to be positional 'name', got %+v", fn.Params[0])
		}
		if fn.Params[1].Name != "title" || !fn.Params[1].Refinement || !fn.Params[1].TakesValue {
			t.Fatalf("expected second param to be value refinement 'title', got %+v", fn.Params[1])
		}
		if fn.Params[2].Name != "verbose" || !fn.Params[2].Refinement || fn.Params[2].TakesValue {
			t.Fatalf("expected third param to be flag refinement 'verbose', got %+v", fn.Params[2])
		}

		if fn.Body == nil || len(fn.Body.Elements) != 1 {
			t.Fatalf("expected body with one element, got %+v", fn.Body)
		}
		if fn.Body.Elements[0].GetType() != value.TypeParen {
			t.Fatalf("expected first body element to be paren, got %v", fn.Body.Elements[0].GetType())
		}
	})

	t.Run("errors for invalid definitions", func(t *testing.T) {
		cases := []string{
			"fn 42 [42]",
			"fn [42] [42]",
			"fn [x x] [x]",
			"fn [x --x []] [x]",
			"fn [] 42",
		}

		for _, src := range cases {
			if _, err := Evaluate(src); err == nil {
				t.Fatalf("expected error for %q", src)
			}
		}
	})
}

func TestFunction_Call(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  core.Value
	}{
		{
			name: "single positional argument",
			input: `square: fn [n] [(* n n)]
    square 5`,
			want: value.NewIntVal(25),
		},
		{
			name: "multiple positional arguments",
			input: `add: fn [a b] [(+ a b)]
    add 3 4`,
			want: value.NewIntVal(7),
		},
		{
			name: "no arguments",
			input: `forty-two: fn [] [42]
    forty-two`,
			want: value.NewIntVal(42),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Evaluate(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !result.Equals(tt.want) {
				t.Fatalf("expected %v, got %v", tt.want, result)
			}
		})
	}
}

func TestFunction_LocalScoping(t *testing.T) {
	script := `counter: 10
increment: fn [] [
    counter: 1
    counter
]
result: increment
counter`

	e, result, err := evalScriptWithEvaluator(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.Equals(value.NewIntVal(10)) {
		t.Fatalf("expected global counter to remain 10, got %v", result)
	}

	local, ok := getGlobal(e, "result")
	if !ok {
		t.Fatalf("expected result binding")
	}
	if !local.Equals(value.NewIntVal(1)) {
		t.Fatalf("expected function return 1, got %v", local)
	}
}

func TestFunction_MutableBlockIsolation(t *testing.T) {
	testCases := []struct {
		name  string
		input string
		check func(e *eval.Evaluator) error
	}{
		{
			name: "flat block isolation",
			input: `create-block: fn [] [
  arr: []
  append arr 1
  arr
]
result1: create-block
result2: create-block
result3: create-block`,
			check: func(e *eval.Evaluator) error {
				expected := value.NewBlockVal([]core.Value{value.NewIntVal(1)})
				for _, name := range []string{"result1", "result2", "result3"} {
					val, ok := getGlobal(e, name)
					if !ok {
						return fmt.Errorf("expected %s binding", name)
					}
					if !val.Equals(expected) {
						return fmt.Errorf("expected %s to be [1], got %v", name, val)
					}
				}
				return nil
			},
		},
		{
			name: "nested block isolation",
			input: `create-nested: fn [] [
  outer: [[1]]
  inner: first outer
  append inner 2
  outer
]
result1: create-nested
inner1: first result1
append inner1 99
result2: create-nested`,
			check: func(e *eval.Evaluator) error {
				result1, ok := getGlobal(e, "result1")
				if !ok {
					return fmt.Errorf("expected result1 binding")
				}
				expected1 := value.NewBlockVal([]core.Value{
					value.NewBlockVal([]core.Value{value.NewIntVal(1), value.NewIntVal(2), value.NewIntVal(99)}),
				})
				if !result1.Equals(expected1) {
					return fmt.Errorf("expected result1 to be [[1 2 99]], got %v", result1)
				}

				result2, ok := getGlobal(e, "result2")
				if !ok {
					return fmt.Errorf("expected result2 binding")
				}
				expected2 := value.NewBlockVal([]core.Value{
					value.NewBlockVal([]core.Value{value.NewIntVal(1), value.NewIntVal(2)}),
				})
				if !result2.Equals(expected2) {
					return fmt.Errorf("expected result2 to be [[1 2]], got %v", result2)
				}
				return nil
			},
		},
		{
			name: "binary isolation",
			input: `create-binary: fn [] [
  bin: #{}
  append bin 1
  bin
]
result1: create-binary
result2: create-binary
result3: create-binary`,
			check: func(e *eval.Evaluator) error {
				expected := value.NewBinaryVal([]byte{1})
				for _, name := range []string{"result1", "result2", "result3"} {
					val, ok := getGlobal(e, name)
					if !ok {
						return fmt.Errorf("expected %s binding", name)
					}
					if !val.Equals(expected) {
						return fmt.Errorf("expected %s to be #{01}, got %v", name, val)
					}
				}
				return nil
			},
		},
		{
			name: "nested binary in block isolation",
			input: `create-mixed: fn [] [
  outer: [#{}]
  inner: first outer
  append inner 42
  outer
]
result1: create-mixed
inner1: first result1
append inner1 99
result2: create-mixed`,
			check: func(e *eval.Evaluator) error {
				result1, ok := getGlobal(e, "result1")
				if !ok {
					return fmt.Errorf("expected result1 binding")
				}
				expected1 := value.NewBlockVal([]core.Value{
					value.NewBinaryVal([]byte{42, 99}),
				})
				if !result1.Equals(expected1) {
					return fmt.Errorf("expected result1 to be [#{2A63}], got %v", result1)
				}

				result2, ok := getGlobal(e, "result2")
				if !ok {
					return fmt.Errorf("expected result2 binding")
				}
				expected2 := value.NewBlockVal([]core.Value{
					value.NewBinaryVal([]byte{42}),
				})
				if !result2.Equals(expected2) {
					return fmt.Errorf("expected result2 to be [#{2A}], got %v", result2)
				}
				return nil
			},
		},
		{
			name: "string isolation",
			input: `create-string: fn [] [
  str: ""
  append str "x"
  str
]
result1: create-string
result2: create-string
result3: create-string`,
			check: func(e *eval.Evaluator) error {
				expected := value.NewStrVal("x")
				for _, name := range []string{"result1", "result2", "result3"} {
					val, ok := getGlobal(e, name)
					if !ok {
						return fmt.Errorf("expected %s binding", name)
					}
					if !val.Equals(expected) {
						return fmt.Errorf("expected %s to be \"x\", got %v", name, val)
					}
				}
				return nil
			},
		},
		{
			name: "nested string in block isolation",
			input: `create-mixed: fn [] [
  outer: [""]
  inner: first outer
  append inner "a"
  outer
]
result1: create-mixed
inner1: first result1
append inner1 "b"
result2: create-mixed`,
			check: func(e *eval.Evaluator) error {
				result1, ok := getGlobal(e, "result1")
				if !ok {
					return fmt.Errorf("expected result1 binding")
				}
				expected1 := value.NewBlockVal([]core.Value{
					value.NewStrVal("ab"),
				})
				if !result1.Equals(expected1) {
					return fmt.Errorf("expected result1 to be [\"ab\"], got %v", result1)
				}

				result2, ok := getGlobal(e, "result2")
				if !ok {
					return fmt.Errorf("expected result2 binding")
				}
				expected2 := value.NewBlockVal([]core.Value{
					value.NewStrVal("a"),
				})
				if !result2.Equals(expected2) {
					return fmt.Errorf("expected result2 to be [\"a\"], got %v", result2)
				}
				return nil
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			e, _, err := evalScriptWithEvaluator(tc.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if err := tc.check(e); err != nil {
				t.Fatalf("check failed: %v", err)
			}
		})
	}
}

func TestFunction_FlagRefinement(t *testing.T) {
	script := `flag-test: fn [msg --verbose] [verbose]
without: flag-test "hello"
with: flag-test "world" --verbose
with`

	e, result, err := evalScriptWithEvaluator(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.Equals(value.NewLogicVal(true)) {
		t.Fatalf("expected final result true, got %v", result)
	}

	without, ok := getGlobal(e, "without")
	if !ok {
		t.Fatalf("expected without binding")
	}
	if !without.Equals(value.NewLogicVal(false)) {
		t.Fatalf("expected flag default false, got %v", without)
	}
}

func TestFunction_ValueRefinement(t *testing.T) {
	script := `greet: fn [name --title []] [title]
no-title: greet "Alice"
with-title: greet "Bob" --title "Dr."
with-title`

	e, result, err := evalScriptWithEvaluator(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.Equals(value.NewStrVal("Dr.")) {
		t.Fatalf("expected final result Dr., got %v", result)
	}

	noTitle, ok := getGlobal(e, "no-title")
	if !ok {
		t.Fatalf("expected no-title binding")
	}
	if noTitle.GetType() != value.TypeNone {
		t.Fatalf("expected value refinement default none, got %v", noTitle)
	}
}

func TestFunction_RefinementOrder(t *testing.T) {
	script := `process: fn [a b --flag --limit []] [limit]
first: process 1 2 --flag --limit 5
second: process 1 2 --limit 7 --flag
third: process 1 2
second`

	e, result, err := evalScriptWithEvaluator(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.Equals(value.NewIntVal(7)) {
		t.Fatalf("expected second call result 7, got %v", result)
	}

	first, ok := getGlobal(e, "first")
	if !ok {
		t.Fatalf("expected first binding")
	}
	if !first.Equals(value.NewIntVal(5)) {
		t.Fatalf("expected first call result 5, got %v", first)
	}

	third, ok := getGlobal(e, "third")
	if !ok {
		t.Fatalf("expected third binding")
	}
	if third.GetType() != value.TypeNone {
		t.Fatalf("expected third call result none when no refinements, got %v", third)
	}
}

func TestFunction_Closure(t *testing.T) {
	result, err := Evaluate(`make-adder: fn [x] [
    fn [y] [
        (+ x y)
    ]
]
add5: make-adder 5
add5 7`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.Equals(value.NewIntVal(12)) {
		t.Fatalf("expected closure to capture x=5, got %v", result)
	}
}

func TestFunction_Recursion(t *testing.T) {
	result, err := Evaluate(`fact: fn [n] [
    if (= n 0) [1] [
        (* n (fact (- n 1)))
    ]
]
fact 5`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.Equals(value.NewIntVal(120)) {
		t.Fatalf("expected factorial 120, got %v", result)
	}
}

func evalScriptWithEvaluator(src string) (*eval.Evaluator, core.Value, error) {
	vals, locations, err := parse.ParseWithSource(src, "(test)")
	if err != nil {
		return nil, value.NewNoneVal(), err
	}

	e := NewTestEvaluator()
	result, evalErr := e.DoBlock(vals, locations)
	return e, result, evalErr
}

func getGlobal(e *eval.Evaluator, name string) (core.Value, bool) {
	if len(e.Frames) == 0 {
		return value.NewNoneVal(), false
	}

	val, ok := e.Frames[0].Get(name)
	return val, ok
}
