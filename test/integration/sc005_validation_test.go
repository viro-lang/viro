package integration

import (
	"testing"
	"time"

	"github.com/marcin-radoszewski/viro/internal/parse"
)

// TestSC005_PerformanceBaselines validates success criterion SC-005:
// Evaluation performance supports interactive use: simple expressions (literals, arithmetic)
// complete in under 10 milliseconds, complex expressions (nested function calls) complete
// in under 100 milliseconds on standard hardware
func TestSC005_PerformanceBaselines(t *testing.T) {
	evaluator := NewTestEvaluator()

	// Test 1: Simple expression performance (target: <10ms)
	t.Run("SimpleExpressions", func(t *testing.T) {
		simpleTests := []string{
			"42",
			"3 + 4",
			"10 * 20 - 5",
			"1 + 2 * 3 - 4 + 5 / 2",
		}

		for _, expr := range simpleTests {
			values, err := parse.Parse(expr)
			if err != nil {
				t.Fatalf("parse failed for %q: %v", expr, err)
			}

			start := time.Now()
			_, err = evaluator.DoBlock(values)
			elapsed := time.Since(start)

			if err != nil {
				t.Errorf("evaluation failed for %q: %v", expr, err)
				continue
			}

			if elapsed > 10*time.Millisecond {
				t.Errorf("Simple expression %q took %v, exceeds 10ms baseline", expr, elapsed)
			} else {
				t.Logf("Simple expression %q: %v ✓", expr, elapsed)
			}
		}
	})

	// Test 2: Complex expression performance (target: <100ms)
	t.Run("ComplexExpressions", func(t *testing.T) {
		complexTests := []struct {
			name string
			code string
		}{
			{
				name: "Function definition and call",
				code: "square: fn [n] [(* n n)]  square 25",
			},
			{
				name: "Nested function calls",
				code: `
					add: fn [a b] [(+ a b)]
					mul: fn [a b] [(* a b)]
					result: mul (add 3 4) (add 5 6)
					result
				`,
			},
			{
				name: "Loop with calculations",
				code: `
					total: 0
					loop 10 [total: (+ total 5)]
					total
				`,
			},
			{
				name: "Conditional with function",
				code: `
					max: fn [a b] [if (> a b) [a] [b]]
					max 10 20
				`,
			},
		}

		for _, tt := range complexTests {
			t.Run(tt.name, func(t *testing.T) {
				values, err := parse.Parse(tt.code)
				if err != nil {
					t.Fatalf("parse failed: %v", err)
				}

				start := time.Now()
				_, err = evaluator.DoBlock(values)
				elapsed := time.Since(start)

				if err != nil {
					t.Errorf("evaluation failed: %v", err)
					return
				}

				if elapsed > 100*time.Millisecond {
					t.Errorf("Complex expression took %v, exceeds 100ms baseline", elapsed)
				} else {
					t.Logf("Complex expression: %v ✓", elapsed)
				}
			})
		}
	})

	t.Logf("SC-005 SUCCESS: Performance baselines met for interactive use")
}
