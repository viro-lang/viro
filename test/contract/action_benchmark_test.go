package contract

import (
	"testing"

	"github.com/marcin-radoszewski/viro/internal/parse"
)

// BenchmarkActionDispatch measures the overhead of action dispatch
// compared to direct native function calls.
// Contract: Phase 6 - T060
func BenchmarkActionDispatch(b *testing.B) {
	e := NewTestEvaluator()

	// Parse once, evaluate many times
	tokens, err := parse.Parse("first [1 2 3 4 5]")
	if err != nil {
		b.Fatalf("Parse error: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := e.Do_Blk(tokens)
		if err != nil {
			b.Fatalf("Eval error: %v", err)
		}
	}
}

// BenchmarkActionDispatchString measures dispatch overhead for string operations.
func BenchmarkActionDispatchString(b *testing.B) {
	e := NewTestEvaluator()

	tokens, err := parse.Parse(`first "hello world"`)
	if err != nil {
		b.Fatalf("Parse error: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := e.Do_Blk(tokens)
		if err != nil {
			b.Fatalf("Eval error: %v", err)
		}
	}
}

// BenchmarkActionAppend measures append action performance.
func BenchmarkActionAppend(b *testing.B) {
	e := NewTestEvaluator()

	// Setup: Create a block variable
	setupTokens, err := parse.Parse("b: [1 2 3]")
	if err != nil {
		b.Fatalf("Parse error: %v", err)
	}
	_, err = e.Do_Blk(setupTokens)
	if err != nil {
		b.Fatalf("Setup error: %v", err)
	}

	// Parse append operation
	tokens, err := parse.Parse("append b 4")
	if err != nil {
		b.Fatalf("Parse error: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := e.Do_Blk(tokens)
		if err != nil {
			b.Fatalf("Eval error: %v", err)
		}
	}
}

// BenchmarkActionLength measures length? action performance.
func BenchmarkActionLength(b *testing.B) {
	e := NewTestEvaluator()

	tokens, err := parse.Parse("length? [1 2 3 4 5 6 7 8 9 10]")
	if err != nil {
		b.Fatalf("Parse error: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := e.Do_Blk(tokens)
		if err != nil {
			b.Fatalf("Eval error: %v", err)
		}
	}
}

// BenchmarkTypeFrameLookup measures the overhead of type frame lookup alone.
func BenchmarkTypeFrameLookup(b *testing.B) {
	e := NewTestEvaluator()

	// Parse multiple action calls to test dispatch overhead
	tokens, err := parse.Parse(`
		first [1 2 3]
		last [1 2 3]
		first "hello"
		last "world"
		length? [1 2 3]
	`)
	if err != nil {
		b.Fatalf("Parse error: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := e.Do_Blk(tokens)
		if err != nil {
			b.Fatalf("Eval error: %v", err)
		}
	}
}
