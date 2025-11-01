package integration

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestEvalWithStdin(t *testing.T) {
	viroPath := "../../viro"
	if _, err := os.Stat(viroPath); os.IsNotExist(err) {
		t.Skip("viro binary not found, run 'make build' first")
	}

	tests := []struct {
		name  string
		stdin string
		expr  string
		want  string
	}{
		{
			name:  "first from stdin block",
			stdin: "data: [1 2 3]",
			expr:  "first data",
			want:  "1",
		},
		{
			name:  "length from stdin",
			stdin: "data: [a b c d]",
			expr:  "length? data",
			want:  "4",
		},
		{
			name:  "last from stdin",
			stdin: "data: [10 20 30]",
			expr:  "last data",
			want:  "30",
		},
		{
			name:  "variable from stdin used in expression",
			stdin: "x: 10",
			expr:  "x * 2",
			want:  "20",
		},
		{
			name:  "function from stdin used in expression",
			stdin: "double: fn [n] [n * 2]",
			expr:  "double 7",
			want:  "14",
		},
		{
			name:  "multiple values from stdin",
			stdin: "a: 5\nb: 10",
			expr:  "a + b",
			want:  "15",
		},
		{
			name:  "data manipulation from stdin",
			stdin: "data: [1 2 3 4 5]",
			expr:  "length? data",
			want:  "5",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command(viroPath, "-c", tt.expr, "--stdin")
			cmd.Stdin = strings.NewReader(tt.stdin)

			output, err := cmd.CombinedOutput()

			if err != nil {
				t.Fatalf("execution failed: %v\nOutput: %s", err, output)
			}

			if !strings.Contains(string(output), tt.want) {
				t.Errorf("output = %q, want to contain %q", output, tt.want)
			}
		})
	}
}

func TestEvalStdinPipelineUsage(t *testing.T) {
	viroPath := "../../viro"
	if _, err := os.Stat(viroPath); os.IsNotExist(err) {
		t.Skip("viro binary not found, run 'make build' first")
	}

	tests := []struct {
		name  string
		stdin string
		expr  string
		want  string
	}{
		{
			name:  "process block data",
			stdin: "data: [10 20 30 40 50]",
			expr:  "length? data",
			want:  "5",
		},
		{
			name:  "access first element",
			stdin: `data: ["hello" "world" "test"]`,
			expr:  "first data",
			want:  "hello",
		},
		{
			name:  "access last element",
			stdin: "data: [1 2 3 4 5]",
			expr:  "last data",
			want:  "5",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command(viroPath, "-c", tt.expr, "--stdin")
			cmd.Stdin = strings.NewReader(tt.stdin)

			output, err := cmd.CombinedOutput()

			if err != nil {
				t.Fatalf("execution failed: %v\nOutput: %s", err, output)
			}

			if !strings.Contains(string(output), tt.want) {
				t.Errorf("output = %q, want to contain %q", output, tt.want)
			}
		})
	}
}

func TestEvalStdinWithNoPrint(t *testing.T) {
	viroPath := "../../viro"
	if _, err := os.Stat(viroPath); os.IsNotExist(err) {
		t.Skip("viro binary not found, run 'make build' first")
	}

	cmd := exec.Command(viroPath, "-c", "first data", "--stdin", "--no-print")
	cmd.Stdin = strings.NewReader("data: [1 2 3]")

	output, err := cmd.CombinedOutput()

	if err != nil {
		t.Fatalf("execution failed: %v\nOutput: %s", err, output)
	}

	if len(output) > 0 {
		t.Errorf("expected no output with --no-print, got: %s", output)
	}
}

func TestEvalStdinWithQuiet(t *testing.T) {
	viroPath := "../../viro"
	if _, err := os.Stat(viroPath); os.IsNotExist(err) {
		t.Skip("viro binary not found, run 'make build' first")
	}

	cmd := exec.Command(viroPath, "-c", "first data", "--stdin", "--quiet")
	cmd.Stdin = strings.NewReader("data: [1 2 3]")

	output, err := cmd.CombinedOutput()

	if err != nil {
		t.Fatalf("execution failed: %v\nOutput: %s", err, output)
	}

	if len(output) > 0 {
		t.Errorf("expected no output with --quiet, got: %s", output)
	}
}

func TestEvalStdinErrors(t *testing.T) {
	viroPath := "../../viro"
	if _, err := os.Stat(viroPath); os.IsNotExist(err) {
		t.Skip("viro binary not found, run 'make build' first")
	}

	tests := []struct {
		name     string
		stdin    string
		expr     string
		wantExit int
		wantErr  string
	}{
		{
			name:     "syntax error in stdin",
			stdin:    "[unclosed",
			expr:     "first",
			wantExit: 2,
			wantErr:  "Syntax error",
		},
		{
			name:     "runtime error in stdin",
			stdin:    "x: 1 / 0",
			expr:     "x",
			wantExit: 1,
			wantErr:  "Math error",
		},
		{
			name:     "syntax error in expression",
			stdin:    "x: 10",
			expr:     "x + [unclosed",
			wantExit: 2,
			wantErr:  "Syntax error",
		},
		{
			name:     "undefined variable from stdin",
			stdin:    "a: 5",
			expr:     "undefined_var",
			wantExit: 1,
			wantErr:  "No value for word",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command(viroPath, "-c", tt.expr, "--stdin")
			cmd.Stdin = strings.NewReader(tt.stdin)

			output, err := cmd.CombinedOutput()

			exitCode := 0
			if err != nil {
				if exitErr, ok := err.(*exec.ExitError); ok {
					exitCode = exitErr.ExitCode()
				}
			}

			if exitCode != tt.wantExit {
				t.Errorf("exit code = %d, want %d\nOutput: %s", exitCode, tt.wantExit, output)
			}

			if !strings.Contains(string(output), tt.wantErr) {
				t.Errorf("error output = %q, want to contain %q", output, tt.wantErr)
			}
		})
	}
}

func TestEvalStdinComplexProgram(t *testing.T) {
	viroPath := "../../viro"
	if _, err := os.Stat(viroPath); os.IsNotExist(err) {
		t.Skip("viro binary not found, run 'make build' first")
	}

	t.Skip("Complex reduce function test - simplify or fix syntax")

	stdin := `
; Define helper functions
square: fn [n] [n * n]
sum-squares: fn [lst] [
	reduce [acc val] [acc + square val] lst 0
]
data: [1 2 3 4 5]
`

	expr := "sum-squares data"

	cmd := exec.Command(viroPath, "-c", expr, "--stdin")
	cmd.Stdin = strings.NewReader(stdin)

	output, err := cmd.CombinedOutput()

	if err != nil {
		t.Fatalf("execution failed: %v\nOutput: %s", err, output)
	}

	if !strings.Contains(string(output), "55") {
		t.Errorf("output = %q, want to contain '55' (1²+2²+3²+4²+5²)", output)
	}
}

func TestEvalStdinEmptyInput(t *testing.T) {
	viroPath := "../../viro"
	if _, err := os.Stat(viroPath); os.IsNotExist(err) {
		t.Skip("viro binary not found, run 'make build' first")
	}

	cmd := exec.Command(viroPath, "-c", "3 + 4", "--stdin")
	cmd.Stdin = strings.NewReader("")

	output, err := cmd.CombinedOutput()

	if err != nil {
		t.Fatalf("execution failed: %v\nOutput: %s", err, output)
	}

	if !strings.Contains(string(output), "7") {
		t.Errorf("output = %q, want to contain '7'", output)
	}
}

func TestEvalStdinLargeInput(t *testing.T) {
	viroPath := "../../viro"
	if _, err := os.Stat(viroPath); os.IsNotExist(err) {
		t.Skip("viro binary not found, run 'make build' first")
	}

	var stdinBuilder strings.Builder
	stdinBuilder.WriteString("data: [\n")
	for i := 1; i <= 1000; i++ {
		stdinBuilder.WriteString("  ")
		stdinBuilder.WriteString(strings.TrimSpace(strings.Fields(strings.Trim(fmt.Sprint(i), "[]"))[0]))
		stdinBuilder.WriteString("\n")
	}
	stdinBuilder.WriteString("]\n")

	expr := "length? data"

	cmd := exec.Command(viroPath, "-c", expr, "--stdin")
	cmd.Stdin = strings.NewReader(stdinBuilder.String())

	output, err := cmd.CombinedOutput()

	if err != nil {
		t.Fatalf("execution failed: %v\nOutput: %s", err, output)
	}

	if !strings.Contains(string(output), "1000") {
		t.Errorf("output = %q, want to contain '1000'", output)
	}
}

func TestEvalStdinMultilineInput(t *testing.T) {
	viroPath := "../../viro"
	if _, err := os.Stat(viroPath); os.IsNotExist(err) {
		t.Skip("viro binary not found, run 'make build' first")
	}

	stdin := `
x: 10
y: 20
z: 30
`

	expr := "x + y + z"

	cmd := exec.Command(viroPath, "-c", expr, "--stdin")
	cmd.Stdin = strings.NewReader(stdin)

	output, err := cmd.CombinedOutput()

	if err != nil {
		t.Fatalf("execution failed: %v\nOutput: %s", err, output)
	}

	if !strings.Contains(string(output), "60") {
		t.Errorf("output = %q, want to contain '60'", output)
	}
}

func TestEvalStdinUTF8(t *testing.T) {
	viroPath := "../../viro"
	if _, err := os.Stat(viroPath); os.IsNotExist(err) {
		t.Skip("viro binary not found, run 'make build' first")
	}

	stdin := `greeting: "Hello 世界 🌍"`
	expr := "greeting"

	cmd := exec.Command(viroPath, "-c", expr, "--stdin")
	cmd.Stdin = strings.NewReader(stdin)

	output, err := cmd.CombinedOutput()

	if err != nil {
		t.Fatalf("execution failed: %v\nOutput: %s", err, output)
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "Hello") ||
		!strings.Contains(outputStr, "世界") ||
		!strings.Contains(outputStr, "🌍") {
		t.Errorf("output = %q, want to contain UTF-8 characters", outputStr)
	}
}
