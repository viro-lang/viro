package native_test

import (
	"testing"

	"github.com/marcin-radoszewski/viro/internal/native"
	"github.com/marcin-radoszewski/viro/internal/value"
)

func TestParseParamSpecs_EvalFlag(t *testing.T) {
	block := &value.BlockValue{
		Elements: []value.Value{
			value.WordVal("a"),
			value.LitWordVal("b"),
			value.WordVal("c"),
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
		Elements: []value.Value{
			value.LitWordVal("--flag"),
		},
	}
	_, err := native.ParseParamSpecs(block)
	if err == nil {
		t.Error("expected error for lit-word refinement, got nil")
	}
}
