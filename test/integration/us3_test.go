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

	if output := strings.TrimSpace(evalLine(t, loop, &out, "data: [1 2 3]")); output != "[1 2 3]" {
		t.Fatalf("block literal expected [1 2 3], got %q", output)
	}

	if output := strings.TrimSpace(evalLine(t, loop, &out, "first data")); output != "1" {
		t.Fatalf("first data expected 1, got %q", output)
	}

	if output := strings.TrimSpace(evalLine(t, loop, &out, "last data")); output != "3" {
		t.Fatalf("last data expected 3, got %q", output)
	}

	if output := strings.TrimSpace(evalLine(t, loop, &out, "append data 4")); output != "[1 2 3 4]" {
		t.Fatalf("append data expected [1 2 3 4], got %q", output)
	}

	if output := strings.TrimSpace(evalLine(t, loop, &out, "data")); output != "[1 2 3 4]" {
		t.Fatalf("data after append expected [1 2 3 4], got %q", output)
	}

	if output := strings.TrimSpace(evalLine(t, loop, &out, "insert data 0")); output != "[0 1 2 3 4]" {
		t.Fatalf("insert data expected [0 1 2 3 4], got %q", output)
	}

	if output := strings.TrimSpace(evalLine(t, loop, &out, "data")); output != "[0 1 2 3 4]" {
		t.Fatalf("data after insert expected [0 1 2 3 4], got %q", output)
	}

	if output := strings.TrimSpace(evalLine(t, loop, &out, "text: \"cat\"")); output != "\"cat\"" {
		t.Fatalf("string literal expected \"cat\", got %q", output)
	}

	if output := strings.TrimSpace(evalLine(t, loop, &out, "first text")); output != "\"c\"" {
		t.Fatalf("first text expected \"c\", got %q", output)
	}

	if output := strings.TrimSpace(evalLine(t, loop, &out, "append text \"s\"")); output != "\"cats\"" {
		t.Fatalf("append text expected \"cats\", got %q", output)
	}

	if output := strings.TrimSpace(evalLine(t, loop, &out, "insert text \"the \"")); output != "\"the cats\"" {
		t.Fatalf("insert text expected \"the cats\", got %q", output)
	}
}
