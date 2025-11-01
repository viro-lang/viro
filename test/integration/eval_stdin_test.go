package integration

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/marcin-radoszewski/viro/internal/api"
)

func TestEvalWithStdin(t *testing.T) {
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
			var stdout, stderr bytes.Buffer
			ctx := &api.RuntimeContext{
				Args:   []string{"-c", tt.expr, "--stdin"},
				Stdin:  strings.NewReader(tt.stdin),
				Stdout: &stdout,
				Stderr: &stderr,
			}

			cfg, _ := api.ConfigFromArgs([]string{"-c", tt.expr, "--stdin"})
			exitCode := api.Run(ctx, cfg)

			if exitCode != 0 {
				t.Fatalf("execution failed with exit code %d\nOutput: %s", exitCode, stdout.String()+stderr.String())
			}

			output := stdout.String() + stderr.String()
			if !strings.Contains(output, tt.want) {
				t.Errorf("output = %q, want to contain %q", output, tt.want)
			}
		})
	}
}

func TestEvalStdinPipelineUsage(t *testing.T) {
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
			var stdout, stderr bytes.Buffer
			ctx := &api.RuntimeContext{
				Args:   []string{"-c", tt.expr, "--stdin"},
				Stdin:  strings.NewReader(tt.stdin),
				Stdout: &stdout,
				Stderr: &stderr,
			}

			cfg, _ := api.ConfigFromArgs([]string{"-c", tt.expr, "--stdin"})
			exitCode := api.Run(ctx, cfg)

			if exitCode != 0 {
				t.Fatalf("execution failed with exit code %d\nOutput: %s", exitCode, stdout.String()+stderr.String())
			}

			output := stdout.String() + stderr.String()
			if !strings.Contains(output, tt.want) {
				t.Errorf("output = %q, want to contain %q", output, tt.want)
			}
		})
	}
}

func TestEvalStdinWithNoPrint(t *testing.T) {
	var stdout, stderr bytes.Buffer
	ctx := &api.RuntimeContext{
		Args:   []string{"-c", "first data", "--stdin", "--no-print"},
		Stdin:  strings.NewReader("data: [1 2 3]"),
		Stdout: &stdout,
		Stderr: &stderr,
	}

	cfg, _ := api.ConfigFromArgs([]string{"-c", "first data", "--stdin", "--no-print"})
	exitCode := api.Run(ctx, cfg)

	if exitCode != 0 {
		t.Fatalf("execution failed with exit code %d\nOutput: %s", exitCode, stdout.String()+stderr.String())
	}

	output := stdout.String() + stderr.String()
	if len(output) > 0 {
		t.Errorf("expected no output with --no-print, got: %s", output)
	}
}

func TestEvalStdinWithQuiet(t *testing.T) {
	var stdout, stderr bytes.Buffer
	ctx := &api.RuntimeContext{
		Args:   []string{"-c", "first data", "--stdin", "--quiet"},
		Stdin:  strings.NewReader("data: [1 2 3]"),
		Stdout: &stdout,
		Stderr: &stderr,
	}

	cfg, _ := api.ConfigFromArgs([]string{"-c", "first data", "--stdin", "--quiet"})
	exitCode := api.Run(ctx, cfg)

	if exitCode != 0 {
		t.Fatalf("execution failed with exit code %d\nOutput: %s", exitCode, stdout.String()+stderr.String())
	}

	output := stdout.String() + stderr.String()
	if len(output) > 0 {
		t.Errorf("expected no output with --quiet, got: %s", output)
	}
}

func TestEvalStdinErrors(t *testing.T) {
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
			var stdout, stderr bytes.Buffer
			ctx := &api.RuntimeContext{
				Args:   []string{"-c", tt.expr, "--stdin"},
				Stdin:  strings.NewReader(tt.stdin),
				Stdout: &stdout,
				Stderr: &stderr,
			}

			cfg, _ := api.ConfigFromArgs([]string{"-c", tt.expr, "--stdin"})
			exitCode := api.Run(ctx, cfg)

			if exitCode != tt.wantExit {
				t.Errorf("exit code = %d, want %d\nOutput: %s", exitCode, tt.wantExit, stdout.String()+stderr.String())
			}

			output := stdout.String() + stderr.String()
			if !strings.Contains(output, tt.wantErr) {
				t.Errorf("error output = %q, want to contain %q", output, tt.wantErr)
			}
		})
	}
}

func TestEvalStdinComplexProgram(t *testing.T) {
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

	var stdout, stderr bytes.Buffer
	ctx := &api.RuntimeContext{
		Args:   []string{"-c", expr, "--stdin"},
		Stdin:  strings.NewReader(stdin),
		Stdout: &stdout,
		Stderr: &stderr,
	}

	cfg, _ := api.ConfigFromArgs([]string{"-c", expr, "--stdin"})
	exitCode := api.Run(ctx, cfg)

	if exitCode != 0 {
		t.Fatalf("execution failed with exit code %d\nOutput: %s", exitCode, stdout.String()+stderr.String())
	}

	output := stdout.String() + stderr.String()
	if !strings.Contains(output, "55") {
		t.Errorf("output = %q, want to contain '55' (1²+2²+3²+4²+5²)", output)
	}
}

func TestEvalStdinEmptyInput(t *testing.T) {
	var stdout, stderr bytes.Buffer
	ctx := &api.RuntimeContext{
		Args:   []string{"-c", "3 + 4", "--stdin"},
		Stdin:  strings.NewReader(""),
		Stdout: &stdout,
		Stderr: &stderr,
	}

	cfg, _ := api.ConfigFromArgs([]string{"-c", "3 + 4", "--stdin"})
	exitCode := api.Run(ctx, cfg)

	if exitCode != 0 {
		t.Fatalf("execution failed with exit code %d\nOutput: %s", exitCode, stdout.String()+stderr.String())
	}

	output := stdout.String() + stderr.String()
	if !strings.Contains(output, "7") {
		t.Errorf("output = %q, want to contain '7'", output)
	}
}

func TestEvalStdinLargeInput(t *testing.T) {
	var stdinBuilder strings.Builder
	stdinBuilder.WriteString("data: [\n")
	for i := 1; i <= 1000; i++ {
		stdinBuilder.WriteString("  ")
		stdinBuilder.WriteString(strings.TrimSpace(strings.Fields(strings.Trim(fmt.Sprint(i), "[]"))[0]))
		stdinBuilder.WriteString("\n")
	}
	stdinBuilder.WriteString("]\n")

	expr := "length? data"

	var stdout, stderr bytes.Buffer
	ctx := &api.RuntimeContext{
		Args:   []string{"-c", expr, "--stdin"},
		Stdin:  strings.NewReader(stdinBuilder.String()),
		Stdout: &stdout,
		Stderr: &stderr,
	}

	cfg, _ := api.ConfigFromArgs([]string{"-c", expr, "--stdin"})
	exitCode := api.Run(ctx, cfg)

	if exitCode != 0 {
		t.Fatalf("execution failed with exit code %d\nOutput: %s", exitCode, stdout.String()+stderr.String())
	}

	output := stdout.String() + stderr.String()
	if !strings.Contains(output, "1000") {
		t.Errorf("output = %q, want to contain '1000'", output)
	}
}

func TestEvalStdinMultilineInput(t *testing.T) {
	stdin := `
x: 10
y: 20
z: 30
`

	expr := "x + y + z"

	var stdout, stderr bytes.Buffer
	ctx := &api.RuntimeContext{
		Args:   []string{"-c", expr, "--stdin"},
		Stdin:  strings.NewReader(stdin),
		Stdout: &stdout,
		Stderr: &stderr,
	}

	cfg, _ := api.ConfigFromArgs([]string{"-c", expr, "--stdin"})
	exitCode := api.Run(ctx, cfg)

	if exitCode != 0 {
		t.Fatalf("execution failed with exit code %d\nOutput: %s", exitCode, stdout.String()+stderr.String())
	}

	output := stdout.String() + stderr.String()
	if !strings.Contains(output, "60") {
		t.Errorf("output = %q, want to contain '60'", output)
	}
}

func TestEvalStdinUTF8(t *testing.T) {
	stdin := `greeting: "Hello 世界 🌍"`
	expr := "greeting"

	var stdout, stderr bytes.Buffer
	ctx := &api.RuntimeContext{
		Args:   []string{"-c", expr, "--stdin"},
		Stdin:  strings.NewReader(stdin),
		Stdout: &stdout,
		Stderr: &stderr,
	}

	cfg, _ := api.ConfigFromArgs([]string{"-c", expr, "--stdin"})
	exitCode := api.Run(ctx, cfg)

	if exitCode != 0 {
		t.Fatalf("execution failed with exit code %d\nOutput: %s", exitCode, stdout.String()+stderr.String())
	}

	outputStr := stdout.String() + stderr.String()
	if !strings.Contains(outputStr, "Hello") ||
		!strings.Contains(outputStr, "世界") ||
		!strings.Contains(outputStr, "🌍") {
		t.Errorf("output = %q, want to contain UTF-8 characters", outputStr)
	}
}
