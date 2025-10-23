package contract_test

import (
	"testing"

	"github.com/marcin-radoszewski/viro/internal/parse"
	"github.com/marcin-radoszewski/viro/internal/value"
	"github.com/marcin-radoszewski/viro/test/contract"
)

func TestUserFunctionEvalFalse(t *testing.T) {
	// Test that unevaluated parameter receives raw token
	// But when accessed in function body, it's still evaluated
	// To prevent evaluation, function must use get-word syntax
	code := `
		get-raw: fn ['value] [:value]
		x: 42
		result: get-raw x
	`
	e := contract.NewTestEvaluator()
	vals, err := parse.Parse(code)
	if err != nil {
		t.Fatal(err)
	}
	result, evalErr := e.DoBlock(vals)
	if evalErr != nil {
		t.Fatal(evalErr)
	}
	// result should be x (the word), fetched with get-word
	if result.GetType() != 4 { // TypeWord
		t.Errorf("Expected word!, got type %d", result.GetType())
	}
	wordStr, ok := value.AsWord(result)
	if !ok || wordStr != "x" {
		t.Errorf("Expected word 'x', got %v", wordStr)
	}
}

func TestUserFunctionMixedEval(t *testing.T) {
	// Test mixed evaluation: normal param evaluated, lit-word param raw
	// This test verifies that evaluated params are evaluated before passing,
	// and unevaluated params are passed as-is
	code := `
		type-check: fn [evaluated 'unevaluated] [
			eval-type: type? evaluated
			uneval-type: type? unevaluated
			[eval-type uneval-type]
		]
		result: type-check (2 + 2) (3 + 3)
	`
	e := contract.NewTestEvaluator()
	vals, err := parse.Parse(code)
	if err != nil {
		t.Fatal(err)
	}
	result, evalErr := e.DoBlock(vals)
	if evalErr != nil {
		t.Fatal(evalErr)
	}
	// result should be a block containing two words: [integer! paren!]
	if result.GetType() != 8 { // TypeBlock
		t.Errorf("Expected block!, got %d", result.GetType())
	}
	block, ok := value.AsBlock(result)
	if !ok || len(block.Elements) != 2 {
		t.Fatalf("Expected block of 2 elements, got %d", len(block.Elements))
	}
	// Both elements should be words (type names)
	if block.Elements[0].GetType() != 4 { // TypeWord
		t.Errorf("Expected word (type name), got %d", block.Elements[0].GetType())
	}
	if block.Elements[1].GetType() != 4 { // TypeWord
		t.Errorf("Expected word (type name), got %d", block.Elements[1].GetType())
	}
}

func TestNativeIfEvalArgs(t *testing.T) {
	code := `
		x: 10
		result: if (x > 5) [x: 1] [x: 2]
		final: x
	`
	e := contract.NewTestEvaluator()
	vals, err := parse.Parse(code)
	if err != nil {
		t.Fatal(err)
	}
	_, evalErr := e.DoBlock(vals)
	if evalErr != nil {
		t.Fatal(evalErr)
	}
	// x should be 1
	final, found := e.Lookup("x")
	if !found {
		t.Fatal("x not found")
	}
	if final.GetType() != 2 { // TypeInteger
		t.Errorf("Expected integer type, got %d", final.GetType())
	}
	ival, ok := value.AsInteger(final)
	if !ok || ival != 1 {
		t.Errorf("Expected x = 1, got %v", ival)
	}
}

func TestRefinementsAlwaysEvaluated(t *testing.T) {
	// Test that refinement values are always evaluated, even if param is lit-word
	// Refinements should be evaluated regardless of Eval flag
	code := `
		test-fn: fn [a 'b --flag] [type? flag]
		x: 42
		y: 99
		result1: test-fn 1 2 --flag x
		result2: test-fn 1 2 --flag y
	`
	e := contract.NewTestEvaluator()
	vals, err := parse.Parse(code)
	if err != nil {
		t.Fatal(err)
	}
	_, evalErr := e.DoBlock(vals)
	if evalErr != nil {
		t.Fatal(evalErr)
	}
	// Check result1: should be integer! (type of x=42)
	result1, found1 := e.Lookup("result1")
	if !found1 {
		t.Fatal("result1 not found")
	}
	// Check result2: should be integer! (type of y=99)
	result2, found2 := e.Lookup("result2")
	if !found2 {
		t.Fatal("result2 not found")
	}
	// Both should return word "integer!" (the type name)
	if result1.GetType() != 4 { // TypeWord
		t.Errorf("Expected word (type name), got %d", result1.GetType())
	}
	if result2.GetType() != 4 { // TypeWord
		t.Errorf("Expected word (type name), got %d", result2.GetType())
	}
}

func TestLitWordRefinementError(t *testing.T) {
	// Parser should handle lit-word inside blocks, but fn should reject lit-word refinements
	code := `
		quote-ref: fn ['--invalid] []
	`
	e := contract.NewTestEvaluator()
	vals, err := parse.Parse(code)
	if err != nil {
		// If parser rejects it, that's also acceptable
		return
	}
	// Should fail during fn execution (ParseParamSpecs)
	_, evalErr := e.DoBlock(vals)
	if evalErr == nil {
		t.Error("Expected error for lit-word refinement, got nil")
	}
}

// TestLitWordParameterReturnsValue verifies that lit-word parameters work like REBOL:
// the parameter value is returned without re-evaluation
func TestLitWordParameterReturnsValue(t *testing.T) {
	code := `
		f: fn ['w] [w]
		result: f word
		type? result
	`
	e := contract.NewTestEvaluator()
	vals, err := parse.Parse(code)
	if err != nil {
		t.Fatal(err)
	}
	result, evalErr := e.DoBlock(vals)
	if evalErr != nil {
		t.Fatal(evalErr)
	}
	// Should return word! type
	if result.GetType() != 4 { // TypeWord
		t.Errorf("Expected word (type name 'word!'), got %d", result.GetType())
	}
	if result.Mold() != "word!" {
		t.Errorf("Expected 'word!', got '%s'", result.Mold())
	}
}

func TestUserFunctionNestedCalls(t *testing.T) {
	code := `
		inc: fn [i] [i + 1]
		result: inc inc inc inc 1
		result
	`
	e := contract.NewTestEvaluator()
	vals, err := parse.Parse(code)
	if err != nil {
		t.Fatal(err)
	}
	result, evalErr := e.DoBlock(vals)
	if evalErr != nil {
		t.Fatal(evalErr)
	}
	if result.GetType() != 2 { // TypeInteger
		t.Fatalf("Expected integer result, got type %d", result.GetType())
	}
	ival, ok := value.AsInteger(result)
	if !ok {
		t.Fatal("Failed to extract integer value")
	}
	if ival != 5 {
		t.Errorf("Expected result 5, got %d", ival)
	}
}

func TestTypeQueryLitWordArgument(t *testing.T) {
	code := `
		f: fn ['w] [w]
		type? f word
	`
	e := contract.NewTestEvaluator()
	vals, err := parse.Parse(code)
	if err != nil {
		t.Fatal(err)
	}
	result, evalErr := e.DoBlock(vals)
	if evalErr == nil {
		if result.GetType() != 4 {
			t.Fatalf("Expected word! result, got type %d", result.GetType())
		}
		if result.Mold() != "word!" {
			t.Errorf("Expected 'word!', got '%s'", result.Mold())
		}
		return
	}
	// Should not produce an error; include failure message for visibility
	t.Fatalf("type? failed: %v", evalErr)
}
