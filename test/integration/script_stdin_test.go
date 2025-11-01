package integration

import (
	"bytes"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestScriptFromStdin(t *testing.T) {
	viroPath := "../../viro"
	if _, err := os.Stat(viroPath); os.IsNotExist(err) {
		t.Skip("viro binary not found, run 'make build' first")
	}

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
			wantStderr: "Syntax error",
		},
		{
			name:       "runtime error from stdin",
			script:     `1 / 0`,
			wantExit:   1,
			wantStdout: "",
			wantStderr: "Math error",
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
			cmd := exec.Command(viroPath, "-")
			cmd.Stdin = strings.NewReader(tt.script)

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

			outputStr := string(output)
			if tt.wantStdout != "" && !strings.Contains(outputStr, tt.wantStdout) {
				t.Errorf("stdout missing expected content:\nWant: %q\nGot: %q", tt.wantStdout, outputStr)
			}

			if tt.wantStderr != "" && !strings.Contains(outputStr, tt.wantStderr) {
				t.Errorf("stderr missing expected content:\nWant: %q\nGot: %q", tt.wantStderr, outputStr)
			}
		})
	}
}

func TestScriptFromStdinWithPipe(t *testing.T) {
	viroPath := "../../viro"
	if _, err := os.Stat(viroPath); os.IsNotExist(err) {
		t.Skip("viro binary not found, run 'make build' first")
	}

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
			echoCmd := exec.Command("echo", tt.echoInput)
			viroCmd := exec.Command(viroPath, "-")

			var buf bytes.Buffer
			echoCmd.Stdout = &buf
			viroCmd.Stdin = &buf

			if err := echoCmd.Run(); err != nil {
				t.Fatalf("echo failed: %v", err)
			}

			output, err := viroCmd.CombinedOutput()
			if err != nil {
				t.Fatalf("viro failed: %v\nOutput: %s", err, output)
			}

			if !strings.Contains(string(output), tt.wantOutput) {
				t.Errorf("output = %q, want to contain %q", output, tt.wantOutput)
			}
		})
	}
}

func TestScriptFromStdinWithQuietFlag(t *testing.T) {
	viroPath := "../../viro"
	if _, err := os.Stat(viroPath); os.IsNotExist(err) {
		t.Skip("viro binary not found, run 'make build' first")
	}

	script := `print "should be suppressed"`
	cmd := exec.Command(viroPath, "--quiet", "-")
	cmd.Stdin = strings.NewReader(script)

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("execution failed: %v\nOutput: %s", err, output)
	}

	if len(output) > 0 {
		t.Errorf("expected no output with --quiet flag, got: %s", output)
	}
}

func TestScriptFromStdinComplexProgram(t *testing.T) {
	viroPath := "../../viro"
	if _, err := os.Stat(viroPath); os.IsNotExist(err) {
		t.Skip("viro binary not found, run 'make build' first")
	}

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

	cmd := exec.Command(viroPath, "-")
	cmd.Stdin = strings.NewReader(script)

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("execution failed: %v\nOutput: %s", err, output)
	}

	if !strings.Contains(string(output), "120") {
		t.Errorf("output = %q, want to contain '120'", output)
	}
}

func TestScriptFromStdinLargeInput(t *testing.T) {
	viroPath := "../../viro"
	if _, err := os.Stat(viroPath); os.IsNotExist(err) {
		t.Skip("viro binary not found, run 'make build' first")
	}

	var scriptBuilder strings.Builder
	scriptBuilder.WriteString("total: 0\n")
	for i := 1; i <= 100; i++ {
		scriptBuilder.WriteString("total: total + 1\n")
	}
	scriptBuilder.WriteString("print total\n")

	script := scriptBuilder.String()

	cmd := exec.Command(viroPath, "-")
	cmd.Stdin = strings.NewReader(script)

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("execution failed: %v\nOutput: %s", err, output)
	}

	if !strings.Contains(string(output), "100") {
		t.Errorf("output = %q, want to contain '100'", output)
	}
}

func TestScriptFromStdinBinaryData(t *testing.T) {
	viroPath := "../../viro"
	if _, err := os.Stat(viroPath); os.IsNotExist(err) {
		t.Skip("viro binary not found, run 'make build' first")
	}

	binaryData := []byte{0xFF, 0xFE, 0xFD}

	cmd := exec.Command(viroPath, "-")
	cmd.Stdin = bytes.NewReader(binaryData)

	output, err := cmd.CombinedOutput()

	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		}
	}

	if exitCode != 2 && exitCode != 0 {
		t.Logf("Note: Binary data resulted in exit code %d (output: %s)", exitCode, output)
	}
}

func TestScriptFromStdinUTF8(t *testing.T) {
	viroPath := "../../viro"
	if _, err := os.Stat(viroPath); os.IsNotExist(err) {
		t.Skip("viro binary not found, run 'make build' first")
	}

	script := `print "Hello 世界 🌍"`

	cmd := exec.Command(viroPath, "-")
	cmd.Stdin = strings.NewReader(script)

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
