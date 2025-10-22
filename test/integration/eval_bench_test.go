package integration

import (
	"testing"

	"github.com/marcin-radoszewski/viro/internal/core"
	"github.com/marcin-radoszewski/viro/internal/parse"
	"github.com/marcin-radoszewski/viro/internal/value"
)

var (
	simpleEvalResult  core.Value
	complexEvalResult core.Value
)

func BenchmarkEvalSimpleExpression(b *testing.B) {
	source := "1 + 2 * 3 - 4 + 5 / 2"
	values, err := parse.Parse(source)
	if err != nil {
		b.Fatalf("parse failed: %v", err)
	}

	evaluator := NewTestEvaluator()

	warmResult, err := evaluator.DoBlock(values)
	if err != nil {
		b.Fatalf("warm-up evaluation failed: %v", err)
	}
	if got, ok := value.AsInteger(warmResult); !ok || got != 5 {
		b.Fatalf("unexpected warm-up result: %v", warmResult)
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		result, err := evaluator.DoBlock(values)
		if err != nil {
			b.Fatalf("evaluation error: %v", err)
		}
		simpleEvalResult = result
	}

	if got, ok := value.AsInteger(simpleEvalResult); !ok || got != 5 {
		b.Fatalf("unexpected final result: %v", simpleEvalResult)
	}
}

func BenchmarkEvalComplexExpression(b *testing.B) {
	source := `
fib: fn [n] [
	if n <= 1 [
		n
	] [
		fib (n - 1) + fib (n - 2)
	]
]

total: 0
loop 20 [
	total: total + fib 10
]
total
`
	values, err := parse.Parse(source)
	if err != nil {
		b.Fatalf("parse failed: %v", err)
	}

	evaluator := NewTestEvaluator()

	warmResult, err := evaluator.DoBlock(values)
	if err != nil {
		b.Fatalf("warm-up evaluation failed: %v", err)
	}
	if got, ok := value.AsInteger(warmResult); !ok || got != 1100 {
		b.Fatalf("unexpected warm-up result: %v", warmResult)
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		result, err := evaluator.DoBlock(values)
		if err != nil {
			b.Fatalf("evaluation error: %v", err)
		}
		complexEvalResult = result
	}

	if got, ok := value.AsInteger(complexEvalResult); !ok || got != 1100 {
		b.Fatalf("unexpected final result: %v", complexEvalResult)
	}
}
