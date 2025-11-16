package contract

import (
	"testing"

	"github.com/marcin-radoszewski/viro/internal/value"
)

func TestBootstrapScriptsLoaded(t *testing.T) {
	evaluator := NewTestEvaluator()

	importFunc, ok := evaluator.GetFrameByIndex(0).Get("import")
	if !ok {
		t.Fatalf("bootstrap script 'import' function not found")
	}

	if importFunc.GetType() != value.TypeFunction {
		t.Errorf("expected 'import' to be a function, got %s", value.TypeToString(importFunc.GetType()))
	}
}

func TestBootstrapScriptsExecuted(t *testing.T) {
	evaluator := NewTestEvaluator()
	importFunc, ok := evaluator.GetFrameByIndex(0).Get("import")
	if !ok {
		t.Fatalf("bootstrap script 'import' function not found")
	}

	if importFunc.GetType() != value.TypeFunction {
		t.Errorf("expected 'import' to be a function, got %s", value.TypeToString(importFunc.GetType()))
	}
}
