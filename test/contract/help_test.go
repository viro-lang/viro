package contract

import (
	"strings"
	"testing"

	"github.com/marcin-radoszewski/viro/internal/core"
	"github.com/marcin-radoszewski/viro/internal/native"
	"github.com/marcin-radoszewski/viro/internal/value"
)

func TestHelpDirectCall(t *testing.T) {
	e := NewTestEvaluator()

	val, err := native.Help([]core.Value{}, e)
	if err != nil {
		t.Fatalf("Help with no args failed: %v", err)
	}
	if val.GetType() != value.TypeNone {
		t.Errorf("Expected none!, got %v", val.GetType())
	}

	val, err = native.Help([]core.Value{value.WordVal("append")}, e)
	if err != nil {
		t.Fatalf("Help with 'append' failed: %v", err)
	}
	if val.GetType() != value.TypeNone {
		t.Errorf("Expected none!, got %v", val.GetType())
	}

	val, err = native.Help([]core.Value{value.WordVal("math")}, e)
	if err != nil {
		t.Fatalf("Help with 'math' failed: %v", err)
	}
	if val.GetType() != value.TypeNone {
		t.Errorf("Expected none!, got %v", val.GetType())
	}

	val, err = native.Help([]core.Value{value.WordVal("+")}, e)
	if err != nil {
		t.Fatalf("Help with '+' failed: %v", err)
	}
	if val.GetType() != value.TypeNone {
		t.Errorf("Expected none!, got %v", val.GetType())
	}
}

func TestWords(t *testing.T) {
	val, err := Evaluate("words")
	if err != nil {
		t.Fatalf("Eval error: %v", err)
	}

	if val.GetType() != value.TypeBlock {
		t.Fatalf("Expected block!, got %v", val.GetType())
	}

	block, ok := value.AsBlock(val)
	if !ok {
		t.Fatal("Failed to convert to BlockValue")
	}

	for i, elem := range block.Elements {
		if elem.GetType() != value.TypeWord {
			t.Errorf("Element %d: expected word!, got %v", i, elem.GetType())
		}
	}

	names := make(map[string]bool)
	for _, elem := range block.Elements {
		word, _ := value.AsWord(elem)
		names[word] = true
	}

	expectedFunctions := []string{"+", "?", "words", "print", "if"}
	for _, fn := range expectedFunctions {
		if !names[fn] {
			t.Errorf("Expected function '%s' not found in words output", fn)
		}
	}
}

func TestWordsDirectCall(t *testing.T) {
	e := NewTestEvaluator()
	val, err := native.Words([]core.Value{}, e)
	if err != nil {
		t.Fatalf("Words failed: %v", err)
	}

	if val.GetType() != value.TypeBlock {
		t.Fatalf("Expected block!, got %v", val.GetType())
	}

	block, ok := value.AsBlock(val)
	if !ok {
		t.Fatal("Failed to convert to BlockValue")
	}

	if len(block.Elements) < 57 {
		t.Errorf("Expected at least 57 functions, got %d", len(block.Elements))
	}
}

func TestHelpFunctionExists(t *testing.T) {
	e := NewTestEvaluator()
	rootFrame := e.GetFrameByIndex(0)

	fnValue, found := rootFrame.Get("?")
	if !found {
		t.Fatal("? function not found in root frame")
	}

	fn, ok := value.AsFunction(fnValue)
	if !ok {
		t.Fatal("? value is not a function")
	}

	if fn.Native == nil {
		t.Error("? function has nil Native")
	}

	if fn.Doc == nil {
		t.Error("? function has no documentation")
	}

	if fn.Doc.Category != "Help" {
		t.Errorf("Expected category 'Help', got '%s'", fn.Doc.Category)
	}
}

func TestHelpFormatterOutput(t *testing.T) {
	e := NewTestEvaluator()
	rootFrame := e.GetFrameByIndex(0)

	registry := make(map[string]*value.FunctionValue)
	for _, binding := range rootFrame.GetAll() {
		if binding.Value.GetType() == value.TypeFunction {
			if fn, ok := value.AsFunction(binding.Value); ok {
				registry[binding.Symbol] = fn
			}
		}
	}

	output := native.FormatCategoryList(registry)

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
	e := NewTestEvaluator()
	rootFrame := e.GetFrameByIndex(0)

	fnValue, found := rootFrame.Get("+")
	if !found {
		t.Fatal("+ function not found in root frame")
	}

	fn, ok := value.AsFunction(fnValue)
	if !ok {
		t.Fatal("+ value is not a function")
	}

	output := native.FormatHelp("+", fn.Doc)

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
