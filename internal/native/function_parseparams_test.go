package native_test

import (
	"testing"

	"github.com/marcin-radoszewski/viro/internal/core"
	"github.com/marcin-radoszewski/viro/internal/native"
	"github.com/marcin-radoszewski/viro/internal/value"
)

func TestParseParamSpecs_EvalFlag(t *testing.T) {
	block := &value.BlockValue{
		Elements: []core.Value{
			value.NewWordVal("a"),
			value.NewLitWordVal("b"),
			value.NewWordVal("c"),
		},
	}
	params, err := native.ParseParamSpecs(block)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(params) != 3 {
		t.Fatalf("expected 3 params, got %d", len(params))
	}
	if !params[0].Eval || params[1].Eval || !params[2].Eval {
		t.Errorf("Eval flags incorrect: got %v", []bool{params[0].Eval, params[1].Eval, params[2].Eval})
	}
}

func TestParseParamSpecs_LitWordRefinementError(t *testing.T) {
	block := &value.BlockValue{
		Elements: []core.Value{
			value.NewLitWordVal("--flag"),
		},
	}
	_, err := native.ParseParamSpecs(block)
	if err == nil {
		t.Error("expected error for lit-word refinement, got nil")
	}
}
