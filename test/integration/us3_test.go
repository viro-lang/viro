package integration

import (
	"bytes"
	"strings"
	"testing"

	"github.com/marcin-radoszewski/viro/internal/repl"
)

func TestUS3_SeriesOperations(t *testing.T) {
	evaluator := NewTestEvaluator()
	var out bytes.Buffer
	loop := repl.NewREPLForTest(evaluator, &out)

	// Test series operations
	out.Reset()
	loop.EvalLineForTest("data: [1 2 3]")
	result := strings.TrimSpace(out.String())
	if result != "1 2 3" {
		t.Errorf("expected '1 2 3', got %q", result)
	}

	out.Reset()
	loop.EvalLineForTest("first data")
	result = strings.TrimSpace(out.String())
	if result != "1" {
		t.Errorf("expected '1', got %q", result)
	}

	out.Reset()
	loop.EvalLineForTest("last data")
	result = strings.TrimSpace(out.String())
	if result != "3" {
		t.Errorf("expected '3', got %q", result)
	}

	out.Reset()
	loop.EvalLineForTest("append data 4")
	result = strings.TrimSpace(out.String())
	if result != "1 2 3 4" {
		t.Errorf("expected '1 2 3 4', got %q", result)
	}

	out.Reset()
	loop.EvalLineForTest("data")
	result = strings.TrimSpace(out.String())
	if result != "1 2 3 4" {
		t.Errorf("expected '1 2 3 4', got %q", result)
	}

	out.Reset()
	loop.EvalLineForTest("insert data 0")
	result = strings.TrimSpace(out.String())
	if result != "0 1 2 3 4" {
		t.Errorf("expected '0 1 2 3 4', got %q", result)
	}

	out.Reset()
	loop.EvalLineForTest("data")
	result = strings.TrimSpace(out.String())
	if result != "0 1 2 3 4" {
		t.Errorf("expected '0 1 2 3 4', got %q", result)
	}

	out.Reset()
	loop.EvalLineForTest("text: \"cat\"")
	result = strings.TrimSpace(out.String())
	if result != "cat" {
		t.Errorf("expected 'cat', got %q", result)
	}

	out.Reset()
	loop.EvalLineForTest("first text")
	result = strings.TrimSpace(out.String())
	if result != "c" {
		t.Errorf("expected 'c', got %q", result)
	}

	out.Reset()
	loop.EvalLineForTest("append text \"s\"")
	result = strings.TrimSpace(out.String())
	if result != "cats" {
		t.Errorf("expected 'cats', got %q", result)
	}

	out.Reset()
	loop.EvalLineForTest("insert text \"the \"")
	result = strings.TrimSpace(out.String())
	if result != "the cats" {
		t.Errorf("expected 'the cats', got %q", result)
	}
}
