package integration

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestCheckMode(t *testing.T) {
	viroPath := "../../viro"
	if _, err := os.Stat(viroPath); os.IsNotExist(err) {
		t.Skip("viro binary not found, run 'make build' first")
	}

	tests := []struct {
		name       string
		script     string
		wantExit   int
		wantOutput string
	}{
		{
			name:       "valid syntax - simple expression",
			script:     `print "hello"`,
			wantExit:   0,
			wantOutput: "",
		},
		{
			name:       "valid syntax - math",
			script:     `3 + 4 * 5`,
			wantExit:   0,
			wantOutput: "",
		},
		{
			name:       "valid syntax - function definition",
			script:     `double: fn [x] [x * 2]`,
			wantExit:   0,
			wantOutput: "",
		},
		{
			name:       "valid syntax - block",
			script:     `[1 2 3]`,
			wantExit:   0,
			wantOutput: "",
		},
		{
			name:       "invalid syntax - unclosed block",
			script:     `[1 2 3`,
			wantExit:   2,
			wantOutput: "Syntax error",
		},
		{
			name:       "invalid syntax - unclosed string",
			script:     `"unclosed string`,
			wantExit:   2,
			wantOutput: "Syntax error",
		},
		{
			name:       "invalid syntax - malformed expression",
			script:     `print [unclosed`,
			wantExit:   2,
			wantOutput: "Syntax error",
		},
		{
			name:       "empty script",
			script:     "",
			wantExit:   0,
			wantOutput: "",
		},
		{
			name:       "comments only",
			script:     "; This is a comment\n; Another comment",
			wantExit:   0,
			wantOutput: "",
		},
		{
			name:       "valid complex script",
			script:     "fib: fn [n] [if n <= 1 [n] [fib (n - 1) + fib (n - 2)]]\nfib 10",
			wantExit:   0,
			wantOutput: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpfile, err := os.CreateTemp("", "test_check_*.viro")
			if err != nil {
				t.Fatal(err)
			}
			defer os.Remove(tmpfile.Name())

			if _, err := tmpfile.WriteString(tt.script); err != nil {
				t.Fatal(err)
			}
			tmpfile.Close()

			cmd := exec.Command(viroPath, "--check", tmpfile.Name())
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

			if tt.wantOutput != "" {
				if !strings.Contains(string(output), tt.wantOutput) {
					t.Errorf("output = %q, want to contain %q", output, tt.wantOutput)
				}
			}
		})
	}
}

func TestCheckModeDoesNotExecute(t *testing.T) {
	viroPath := "../../viro"
	if _, err := os.Stat(viroPath); os.IsNotExist(err) {
		t.Skip("viro binary not found, run 'make build' first")
	}

	script := `
print "This should not be printed"
1 / 0
`

	tmpfile, err := os.CreateTemp("", "test_check_no_exec_*.viro")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.WriteString(script); err != nil {
		t.Fatal(err)
	}
	tmpfile.Close()

	cmd := exec.Command(viroPath, "--check", tmpfile.Name())
	output, err := cmd.CombinedOutput()

	if err != nil {
		t.Fatalf("check failed on valid syntax: %v\nOutput: %s", err, output)
	}

	outputStr := string(output)
	if strings.Contains(outputStr, "This should not be printed") {
		t.Error("--check mode executed the script (found print output)")
	}

	if strings.Contains(outputStr, "Math error") {
		t.Error("--check mode executed the script (found runtime error)")
	}
}

func TestCheckModeWithVerbose(t *testing.T) {
	viroPath := "../../viro"
	if _, err := os.Stat(viroPath); os.IsNotExist(err) {
		t.Skip("viro binary not found, run 'make build' first")
	}

	script := "print \"hello\"\nprint \"world\""

	tmpfile, err := os.CreateTemp("", "test_check_verbose_*.viro")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.WriteString(script); err != nil {
		t.Fatal(err)
	}
	tmpfile.Close()

	cmd := exec.Command(viroPath, "--check", "--verbose", tmpfile.Name())
	output, err := cmd.CombinedOutput()

	if err != nil {
		t.Fatalf("check failed: %v\nOutput: %s", err, output)
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "Syntax valid") &&
		!strings.Contains(outputStr, "Parsed") {
		t.Logf("Note: --verbose output for --check mode: %s", outputStr)
	}
}

func TestCheckModeFromExistingScripts(t *testing.T) {
	viroPath := "../../viro"
	if _, err := os.Stat(viroPath); os.IsNotExist(err) {
		t.Skip("viro binary not found, run 'make build' first")
	}

	tests := []struct {
		name       string
		scriptPath string
		wantExit   int
	}{
		{
			name:       "hello.viro is valid",
			scriptPath: "../scripts/hello.viro",
			wantExit:   0,
		},
		{
			name:       "math.viro is valid",
			scriptPath: "../scripts/math.viro",
			wantExit:   0,
		},
		{
			name:       "error.viro is invalid",
			scriptPath: "../scripts/error.viro",
			wantExit:   2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := os.Stat(tt.scriptPath); os.IsNotExist(err) {
				t.Skipf("test script %s not found", tt.scriptPath)
			}

			cmd := exec.Command(viroPath, "--check", tt.scriptPath)
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
		})
	}
}

func TestCheckModeMultipleErrors(t *testing.T) {
	viroPath := "../../viro"
	if _, err := os.Stat(viroPath); os.IsNotExist(err) {
		t.Skip("viro binary not found, run 'make build' first")
	}

	script := `
[unclosed block
print "test"
[another unclosed
`

	tmpfile, err := os.CreateTemp("", "test_check_multi_err_*.viro")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.WriteString(script); err != nil {
		t.Fatal(err)
	}
	tmpfile.Close()

	cmd := exec.Command(viroPath, "--check", tmpfile.Name())
	output, err := cmd.CombinedOutput()

	if err == nil {
		t.Fatal("expected syntax error, got none")
	}

	exitErr, ok := err.(*exec.ExitError)
	if !ok {
		t.Fatalf("unexpected error type: %v", err)
	}

	if exitErr.ExitCode() != 2 {
		t.Errorf("exit code = %d, want 2 (syntax error)", exitErr.ExitCode())
	}

	if !strings.Contains(string(output), "Syntax error") {
		t.Errorf("output = %q, want to contain 'Syntax error'", output)
	}
}

func TestCheckModeFileNotFound(t *testing.T) {
	viroPath := "../../viro"
	if _, err := os.Stat(viroPath); os.IsNotExist(err) {
		t.Skip("viro binary not found, run 'make build' first")
	}

	cmd := exec.Command(viroPath, "--check", "nonexistent_file.viro")
	output, err := cmd.CombinedOutput()

	if err == nil {
		t.Fatal("expected error for nonexistent file, got none")
	}

	exitErr, ok := err.(*exec.ExitError)
	if !ok {
		t.Fatalf("unexpected error type: %v", err)
	}

	if exitErr.ExitCode() != 1 {
		t.Errorf("exit code = %d, want 1 (file not found error)", exitErr.ExitCode())
	}

	outputStr := string(output)
	if !strings.Contains(strings.ToLower(outputStr), "no such file") &&
		!strings.Contains(strings.ToLower(outputStr), "not found") &&
		!strings.Contains(outputStr, "Error loading script") {
		t.Errorf("error message doesn't indicate file not found: %s", outputStr)
	}
}

func TestCheckModeWithQuiet(t *testing.T) {
	viroPath := "../../viro"
	if _, err := os.Stat(viroPath); os.IsNotExist(err) {
		t.Skip("viro binary not found, run 'make build' first")
	}

	script := `print "test"`

	tmpfile, err := os.CreateTemp("", "test_check_quiet_*.viro")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.WriteString(script); err != nil {
		t.Fatal(err)
	}
	tmpfile.Close()

	cmd := exec.Command(viroPath, "--check", "--quiet", tmpfile.Name())
	output, err := cmd.CombinedOutput()

	if err != nil {
		t.Fatalf("check failed: %v\nOutput: %s", err, output)
	}

	if len(output) > 0 {
		t.Logf("Note: --quiet with --check produced output: %s", output)
	}
}
