package integration

import (
	"bytes"
	"strings"
	"testing"

	"github.com/marcin-radoszewski/viro/internal/api"
)

func TestScriptFromStdin(t *testing.T) {
	tests := []struct {
		name       string
		script     string
		wantExit   int
		wantStdout string
		wantStderr string
	}{
		{
			name:       "simple print from stdin",
			script:     `print "from stdin"`,
			wantExit:   0,
			wantStdout: "from stdin",
			wantStderr: "",
		},
		{
			name:       "math expression from stdin",
			script:     `print 10 + 20`,
			wantExit:   0,
			wantStdout: "30",
			wantStderr: "",
		},
		{
			name:       "multiple statements from stdin",
			script:     "print 1\nprint 2\nprint 3",
			wantExit:   0,
			wantStdout: "1\n2\n3",
			wantStderr: "",
		},
		{
			name:       "syntax error from stdin",
			script:     `print [unclosed`,
			wantExit:   2,
			wantStdout: "",
			wantStderr: "** Syntax Error",
		},
		{
			name:       "runtime error from stdin",
			script:     `1 / 0`,
			wantExit:   1,
			wantStdout: "",
			wantStderr: "** Math Error",
		},
		{
			name:       "empty input from stdin",
			script:     "",
			wantExit:   0,
			wantStdout: "",
			wantStderr: "",
		},
		{
			name:       "function definition and call from stdin",
			script:     "double: fn [x] [x * 2]\nprint double 5",
			wantExit:   0,
			wantStdout: "10",
			wantStderr: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			ctx := &api.RuntimeContext{
				Args:   []string{"-"},
				Stdin:  strings.NewReader(tt.script),
				Stdout: &stdout,
				Stderr: &stderr,
			}
			cfg, err := api.ConfigFromArgs([]string{"-"})
			if err != nil {
				t.Fatalf("ConfigFromArgs failed: %v", err)
			}

			exitCode := api.Run(ctx, cfg)
			output := stdout.String() + stderr.String()

			if exitCode != tt.wantExit {
				t.Errorf("exit code = %d, want %d\nOutput: %s", exitCode, tt.wantExit, output)
			}

			if tt.wantStdout != "" && !strings.Contains(output, tt.wantStdout) {
				t.Errorf("stdout missing expected content:\nWant: %q\nGot: %q", tt.wantStdout, output)
			}

			if tt.wantStderr != "" && !strings.Contains(output, tt.wantStderr) {
				t.Errorf("stderr missing expected content:\nWant: %q\nGot: %q", tt.wantStderr, output)
			}
		})
	}
}

func TestScriptFromStdinWithPipe(t *testing.T) {
	tests := []struct {
		name       string
		echoInput  string
		wantOutput string
	}{
		{
			name:       "echo piped to viro",
			echoInput:  `print "piped input"`,
			wantOutput: "piped input",
		},
		{
			name:       "multiline piped input",
			echoInput:  "x: 5\nprint x * 2",
			wantOutput: "10",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			ctx := &api.RuntimeContext{
				Args:   []string{"-"},
				Stdin:  strings.NewReader(tt.echoInput),
				Stdout: &stdout,
				Stderr: &stderr,
			}
			cfg, err := api.ConfigFromArgs([]string{"-"})
			if err != nil {
				t.Fatalf("ConfigFromArgs failed: %v", err)
			}

			exitCode := api.Run(ctx, cfg)
			output := stdout.String() + stderr.String()

			if exitCode != api.ExitSuccess {
				t.Fatalf("viro failed with exit code %d\nOutput: %s", exitCode, output)
			}

			if !strings.Contains(output, tt.wantOutput) {
				t.Errorf("output = %q, want to contain %q", output, tt.wantOutput)
			}
		})
	}
}

func TestScriptFromStdinWithQuietFlag(t *testing.T) {
	script := `print "should be suppressed"`
	var stdout, stderr bytes.Buffer
	ctx := &api.RuntimeContext{
		Args:   []string{"--quiet", "-"},
		Stdin:  strings.NewReader(script),
		Stdout: &stdout,
		Stderr: &stderr,
	}
	cfg, err := api.ConfigFromArgs([]string{"--quiet", "-"})
	if err != nil {
		t.Fatalf("ConfigFromArgs failed: %v", err)
	}

	exitCode := api.Run(ctx, cfg)
	output := stdout.String() + stderr.String()

	if exitCode != api.ExitSuccess {
		t.Fatalf("execution failed with exit code %d\nOutput: %s", exitCode, output)
	}

	if len(output) > 0 {
		t.Errorf("expected no output with --quiet flag, got: %s", output)
	}
}

func TestScriptFromStdinComplexProgram(t *testing.T) {
	script := `
; Factorial function
fac: fn [n] [
	if n <= 1 [
		1
	] [
		n * fac (n - 1)
	]
]

print fac 5
`

	var stdout, stderr bytes.Buffer
	ctx := &api.RuntimeContext{
		Args:   []string{"-"},
		Stdin:  strings.NewReader(script),
		Stdout: &stdout,
		Stderr: &stderr,
	}
	cfg, err := api.ConfigFromArgs([]string{"-"})
	if err != nil {
		t.Fatalf("ConfigFromArgs failed: %v", err)
	}

	exitCode := api.Run(ctx, cfg)
	output := stdout.String() + stderr.String()

	if exitCode != api.ExitSuccess {
		t.Fatalf("execution failed with exit code %d\nOutput: %s", exitCode, output)
	}

	if !strings.Contains(output, "120") {
		t.Errorf("output = %q, want to contain '120'", output)
	}
}

func TestScriptFromStdinLargeInput(t *testing.T) {
	var scriptBuilder strings.Builder
	scriptBuilder.WriteString("total: 0\n")
	for i := 1; i <= 100; i++ {
		scriptBuilder.WriteString("total: total + 1\n")
	}
	scriptBuilder.WriteString("print total\n")

	script := scriptBuilder.String()

	var stdout, stderr bytes.Buffer
	ctx := &api.RuntimeContext{
		Args:   []string{"-"},
		Stdin:  strings.NewReader(script),
		Stdout: &stdout,
		Stderr: &stderr,
	}
	cfg, err := api.ConfigFromArgs([]string{"-"})
	if err != nil {
		t.Fatalf("ConfigFromArgs failed: %v", err)
	}

	exitCode := api.Run(ctx, cfg)
	output := stdout.String() + stderr.String()

	if exitCode != api.ExitSuccess {
		t.Fatalf("execution failed with exit code %d\nOutput: %s", exitCode, output)
	}

	if !strings.Contains(output, "100") {
		t.Errorf("output = %q, want to contain '100'", output)
	}
}

func TestScriptFromStdinBinaryData(t *testing.T) {
	binaryData := []byte{0xFF, 0xFE, 0xFD}

	var stdout, stderr bytes.Buffer
	ctx := &api.RuntimeContext{
		Args:   []string{"-"},
		Stdin:  bytes.NewReader(binaryData),
		Stdout: &stdout,
		Stderr: &stderr,
	}
	cfg, err := api.ConfigFromArgs([]string{"-"})
	if err != nil {
		t.Fatalf("ConfigFromArgs failed: %v", err)
	}

	exitCode := api.Run(ctx, cfg)
	output := stdout.String() + stderr.String()

	if exitCode != api.ExitSyntax && exitCode != api.ExitSuccess {
		t.Logf("Note: Binary data resulted in exit code %d (output: %s)", exitCode, output)
	}
}

func TestScriptFromStdinUTF8(t *testing.T) {
	script := `print "Hello 世界 🌍"`

	var stdout, stderr bytes.Buffer
	ctx := &api.RuntimeContext{
		Args:   []string{"-"},
		Stdin:  strings.NewReader(script),
		Stdout: &stdout,
		Stderr: &stderr,
	}
	cfg, err := api.ConfigFromArgs([]string{"-"})
	if err != nil {
		t.Fatalf("ConfigFromArgs failed: %v", err)
	}

	exitCode := api.Run(ctx, cfg)
	output := stdout.String() + stderr.String()

	if exitCode != api.ExitSuccess {
		t.Fatalf("execution failed with exit code %d\nOutput: %s", exitCode, output)
	}

	if !strings.Contains(output, "Hello") ||
		!strings.Contains(output, "世界") ||
		!strings.Contains(output, "🌍") {
		t.Errorf("output = %q, want to contain UTF-8 characters", output)
	}
}
