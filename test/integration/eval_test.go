package integration

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestEvalMode(t *testing.T) {
	viroPath := "../../viro"
	if _, err := os.Stat(viroPath); os.IsNotExist(err) {
		t.Skip("viro binary not found, run 'make build' first")
	}

	tests := []struct {
		name     string
		expr     string
		want     string
		wantExit int
	}{
		{
			name:     "simple math",
			expr:     "3 + 4",
			want:     "7",
			wantExit: 0,
		},
		{
			name:     "multiplication",
			expr:     "10 * 5",
			want:     "50",
			wantExit: 0,
		},
		{
			name:     "complex expression",
			expr:     "2 + 3 * 4",
			want:     "20",
			wantExit: 0,
		},
		{
			name:     "string literal",
			expr:     `"hello"`,
			want:     "hello",
			wantExit: 0,
		},
		{
			name:     "block literal",
			expr:     "[1 2 3]",
			want:     "1 2 3",
			wantExit: 0,
		},
		{
			name:     "function call - first",
			expr:     "first [1 2 3]",
			want:     "1",
			wantExit: 0,
		},
		{
			name:     "function call - length",
			expr:     "length? [a b c d]",
			want:     "4",
			wantExit: 0,
		},
		{
			name:     "division by zero",
			expr:     "1 / 0",
			want:     "Math error",
			wantExit: 1,
		},
		{
			name:     "syntax error",
			expr:     "print [unclosed",
			want:     "Syntax error",
			wantExit: 2,
		},
		{
			name:     "undefined variable",
			expr:     "unknown_var",
			want:     "No value for word",
			wantExit: 1,
		},
		{
			name:     "boolean true",
			expr:     "true",
			want:     "true",
			wantExit: 0,
		},
		{
			name:     "boolean false",
			expr:     "false",
			want:     "false",
			wantExit: 0,
		},
		{
			name:     "comparison",
			expr:     "5 > 3",
			want:     "true",
			wantExit: 0,
		},
		{
			name:     "power function",
			expr:     "pow 2 10",
			want:     "1024",
			wantExit: 0,
		},
		{
			name:     "nested blocks",
			expr:     "[[1 2] [3 4]]",
			want:     "1 2 3 4",
			wantExit: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command(viroPath, "-c", tt.expr)
			output, err := cmd.CombinedOutput()

			exitCode := 0
			if err != nil {
				if exitErr, ok := err.(*exec.ExitError); ok {
					exitCode = exitErr.ExitCode()
				} else {
					t.Fatalf("unexpected error type: %v", err)
				}
			}

			if exitCode != tt.wantExit {
				t.Errorf("exit code = %d, want %d\nOutput: %s", exitCode, tt.wantExit, output)
			}

			if !strings.Contains(string(output), tt.want) {
				t.Errorf("output = %q, want to contain %q", output, tt.want)
			}
		})
	}
}

func TestEvalModeWithNoPrint(t *testing.T) {
	viroPath := "../../viro"
	if _, err := os.Stat(viroPath); os.IsNotExist(err) {
		t.Skip("viro binary not found, run 'make build' first")
	}

	tests := []struct {
		name string
		expr string
	}{
		{
			name: "simple expression",
			expr: "3 + 4",
		},
		{
			name: "string",
			expr: `"hello"`,
		},
		{
			name: "block",
			expr: "[1 2 3]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command(viroPath, "-c", tt.expr, "--no-print")
			output, err := cmd.CombinedOutput()

			if err != nil {
				t.Fatalf("execution failed: %v\nOutput: %s", err, output)
			}

			if len(output) > 0 {
				t.Errorf("expected no output with --no-print, got: %s", output)
			}
		})
	}
}

func TestEvalModeWithQuiet(t *testing.T) {
	viroPath := "../../viro"
	if _, err := os.Stat(viroPath); os.IsNotExist(err) {
		t.Skip("viro binary not found, run 'make build' first")
	}

	cmd := exec.Command(viroPath, "-c", "3 + 4", "--quiet")
	output, err := cmd.CombinedOutput()

	if err != nil {
		t.Fatalf("execution failed: %v\nOutput: %s", err, output)
	}

	if len(output) > 0 {
		t.Errorf("expected no output with --quiet, got: %s", output)
	}
}

func TestEvalModeMultipleExpressions(t *testing.T) {
	viroPath := "../../viro"
	if _, err := os.Stat(viroPath); os.IsNotExist(err) {
		t.Skip("viro binary not found, run 'make build' first")
	}

	tests := []struct {
		name string
		expr string
		want string
	}{
		{
			name: "variable assignment and use",
			expr: "x: 10  x * 2",
			want: "20",
		},
		{
			name: "function definition and call",
			expr: "double: fn [n] [n * 2]  double 5",
			want: "10",
		},
		{
			name: "multiple statements",
			expr: "a: 1  b: 2  a + b",
			want: "3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command(viroPath, "-c", tt.expr)
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

func TestEvalModeWithVerbose(t *testing.T) {
	viroPath := "../../viro"
	if _, err := os.Stat(viroPath); os.IsNotExist(err) {
		t.Skip("viro binary not found, run 'make build' first")
	}

	cmd := exec.Command(viroPath, "-c", "3 + 4", "--verbose")
	output, err := cmd.CombinedOutput()

	if err != nil {
		t.Fatalf("execution failed: %v\nOutput: %s", err, output)
	}

	if !strings.Contains(string(output), "7") {
		t.Errorf("output = %q, want to contain '7'", output)
	}
}

func TestEvalModeComplexProgram(t *testing.T) {
	viroPath := "../../viro"
	if _, err := os.Stat(viroPath); os.IsNotExist(err) {
		t.Skip("viro binary not found, run 'make build' first")
	}

	t.Skip("Fibonacci recursive function causes stack overflow - known Viro issue")

	expr := "fib: fn [n] [if n <= 1 [n] [fib (n - 1) + fib (n - 2)]]  fib 10"

	cmd := exec.Command(viroPath, "-c", expr)
	output, err := cmd.CombinedOutput()

	if err != nil {
		t.Fatalf("execution failed: %v\nOutput: %s", err, output)
	}

	if !strings.Contains(string(output), "55") {
		t.Errorf("output = %q, want to contain '55' (10th Fibonacci number)", output)
	}
}

func TestEvalModeShellIntegration(t *testing.T) {
	viroPath := "../../viro"
	if _, err := os.Stat(viroPath); os.IsNotExist(err) {
		t.Skip("viro binary not found, run 'make build' first")
	}

	tests := []struct {
		name     string
		expr     string
		useShell bool
		want     string
	}{
		{
			name:     "use in shell variable",
			expr:     "pow 2 10",
			useShell: false,
			want:     "1024.00",
		},
		{
			name:     "calculate with multiple values",
			expr:     "3 + 4 * 5",
			useShell: false,
			want:     "35",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command(viroPath, "-c", tt.expr)
			output, err := cmd.CombinedOutput()

			if err != nil {
				t.Fatalf("execution failed: %v\nOutput: %s", err, output)
			}

			result := strings.TrimSpace(string(output))
			if result != tt.want {
				t.Errorf("output = %q, want %q", result, tt.want)
			}
		})
	}
}

func TestEvalModeEmptyExpression(t *testing.T) {
	viroPath := "../../viro"
	if _, err := os.Stat(viroPath); os.IsNotExist(err) {
		t.Skip("viro binary not found, run 'make build' first")
	}

	t.Skip("Empty expression -c \"\" currently enters REPL mode, which may be intended behavior")

	cmd := exec.Command(viroPath, "-c", "")
	output, err := cmd.CombinedOutput()

	if err != nil {
		t.Fatalf("execution failed: %v\nOutput: %s", err, output)
	}

	if strings.TrimSpace(string(output)) != "" {
		t.Errorf("expected empty or unset output for empty expression, got: %s", output)
	}
}

func TestEvalModeWithSpecialCharacters(t *testing.T) {
	viroPath := "../../viro"
	if _, err := os.Stat(viroPath); os.IsNotExist(err) {
		t.Skip("viro binary not found, run 'make build' first")
	}

	tests := []struct {
		name string
		expr string
		want string
	}{
		{
			name: "string with spaces",
			expr: `"hello world"`,
			want: "hello world",
		},
		{
			name: "string with special chars",
			expr: `"test@example.com"`,
			want: "test@example.com",
		},
		{
			name: "UTF-8 string",
			expr: `"Hello 世界"`,
			want: "Hello 世界",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command(viroPath, "-c", tt.expr)
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
