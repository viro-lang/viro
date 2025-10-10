package contract

import (
	"strings"
	"testing"

	"github.com/marcin-radoszewski/viro/internal/eval"
	"github.com/marcin-radoszewski/viro/internal/native"
	"github.com/marcin-radoszewski/viro/internal/parse"
	"github.com/marcin-radoszewski/viro/internal/value"
)

func TestHelpDirectCall(t *testing.T) {
	// Test calling Help function directly (bypasses parser arity checking)

	// Test with no arguments - should show category list
	val, err := native.Help([]value.Value{})
	if err != nil {
		t.Fatalf("Help with no args failed: %v", err)
	}
	if val.Type != value.TypeNone {
		t.Errorf("Expected none!, got %v", val.Type)
	}

	// Test with function name
	val, err = native.Help([]value.Value{value.WordVal("append")})
	if err != nil {
		t.Fatalf("Help with 'append' failed: %v", err)
	}
	if val.Type != value.TypeNone {
		t.Errorf("Expected none!, got %v", val.Type)
	}

	// Test with category name
	val, err = native.Help([]value.Value{value.WordVal("math")})
	if err != nil {
		t.Fatalf("Help with 'math' failed: %v", err)
	}
	if val.Type != value.TypeNone {
		t.Errorf("Expected none!, got %v", val.Type)
	}

	// Test with operator
	val, err = native.Help([]value.Value{value.WordVal("+")})
	if err != nil {
		t.Fatalf("Help with '+' failed: %v", err)
	}
	if val.Type != value.TypeNone {
		t.Errorf("Expected none!, got %v", val.Type)
	}
}

func TestWords(t *testing.T) {
	evaluator := eval.NewEvaluator()
	result, parseErr := parse.Parse("words")
	if parseErr != nil {
		t.Fatalf("Parse error: %v", parseErr)
	}

	val, evalErr := evaluator.Do_Blk(result)
	if evalErr != nil {
		t.Fatalf("Eval error: %v", evalErr)
	}

	if val.Type != value.TypeBlock {
		t.Fatalf("Expected block!, got %v", val.Type)
	}

	block, ok := val.AsBlock()
	if !ok {
		t.Fatal("Failed to convert to BlockValue")
	}
	if len(block.Elements) < 57 {
		t.Errorf("Expected at least 57 functions, got %d", len(block.Elements))
	}

	for i, elem := range block.Elements {
		if elem.Type != value.TypeWord {
			t.Errorf("Element %d: expected word!, got %v", i, elem.Type)
		}
	}

	names := make(map[string]bool)
	for _, elem := range block.Elements {
		word, _ := elem.AsWord()
		names[word] = true
	}

	expectedFunctions := []string{"+", "append", "?", "words", "print", "if"}
	for _, fn := range expectedFunctions {
		if !names[fn] {
			t.Errorf("Expected function '%s' not found in words output", fn)
		}
	}
}

func TestHelpFunctionExists(t *testing.T) {
	info, found := native.Lookup("?")
	if !found {
		t.Fatal("? function not found in native.Registry")
	}

	if info.Func == nil {
		t.Error("? function has nil Func")
	}

	if info.NeedsEval {
		t.Error("? function should have NeedsEval=false (to receive unevaluated words)")
	}

	if info.Doc == nil {
		t.Error("? function has no documentation")
	}

	if info.Doc.Category != "Help" {
		t.Errorf("Expected category 'Help', got '%s'", info.Doc.Category)
	}
}

func TestHelpFormatterOutput(t *testing.T) {
	output := native.FormatCategoryList(native.Registry)

	if !strings.Contains(output, "Available categories") {
		t.Error("FormatCategoryList missing 'Available categories' header")
	}

	if !strings.Contains(output, "Math") {
		t.Error("FormatCategoryList missing Math category")
	}

	if !strings.Contains(output, "Help") {
		t.Error("FormatCategoryList missing Help category")
	}
}

func TestHelpFunctionDetail(t *testing.T) {
	info, found := native.Lookup("+")
	if !found {
		t.Fatal("+ function not found")
	}

	output := native.FormatHelp("+", info.Doc)

	if !strings.Contains(output, "+") {
		t.Error("FormatHelp missing function name")
	}

	if !strings.Contains(output, "Math") {
		t.Error("FormatHelp missing category")
	}

	if !strings.Contains(output, "Adds two numbers") {
		t.Error("FormatHelp missing summary")
	}

	if !strings.Contains(output, "PARAMETERS:") && !strings.Contains(output, "left") {
		t.Error("FormatHelp missing parameters")
	}

	if !strings.Contains(output, "EXAMPLES:") && !strings.Contains(output, "3 + 4") {
		t.Error("FormatHelp missing examples")
	}
}
