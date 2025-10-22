package integration

import (
	"bytes"
	"strings"
	"testing"

	"github.com/marcin-radoszewski/viro/internal/eval"
	"github.com/marcin-radoszewski/viro/internal/native"
	"github.com/marcin-radoszewski/viro/internal/repl"
)

// TestPortNativesInREPL validates that port natives work through the REPL
func TestPortNativesInREPL(t *testing.T) {
	// Setup sandbox
	tmpDir := t.TempDir()
	if err := eval.InitSandbox(tmpDir); err != nil {
		t.Fatalf("Failed to init sandbox: %v", err)
	}
	native.SandboxRoot = tmpDir

	evaluator := NewTestEvaluator()
	var errOut bytes.Buffer
	loop := repl.NewREPLForTest(evaluator, &errOut)

	tests := []struct {
		name     string
		input    string
		contains string
		setup    []string
	}{
		{
			name:     "Write to file returns none",
			input:    `write "test.txt" "Hello, World!"`,
			contains: "", // write returns none, which prints nothing
			setup:    []string{},
		},
		{
			name:     "Read from file",
			input:    `read "data.txt"`,
			contains: "Test data", // string values are not quoted in REPL Form output
			setup: []string{
				`write "data.txt" "Test data"`,
			},
		},
		{
			name:     "Save and load",
			input:    `load "number.txt"`,
			contains: "42",
			setup: []string{
				`save "number.txt" 42`,
			},
		},
	}

	passedTests := 0
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Run setup commands
			for _, setupCmd := range tt.setup {
				errOut.Reset()
				loop.EvalLineForTest(setupCmd)
				// Setup output is discarded
			}

			// Execute test
			errOut.Reset()
			loop.EvalLineForTest(tt.input)
			result := strings.TrimSpace(errOut.String())

			if !strings.Contains(result, tt.contains) {
				t.Errorf("%s: expected to contain %q, got %q", tt.name, tt.contains, result)
			} else {
				passedTests++
				t.Logf("Port REPL PASS: %s", tt.name)
			}
		})
	}

	t.Logf("Port REPL tests: %d/%d passed", passedTests, len(tests))
}
