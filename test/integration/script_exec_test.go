package integration

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestScriptExecution(t *testing.T) {
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
			name:       "successful execution - hello world",
			script:     `print "Hello, World!"`,
			wantExit:   0,
			wantStdout: "Hello, World!",
			wantStderr: "",
		},
		{
			name:       "successful execution - math",
			script:     `print 3 + 4`,
			wantExit:   0,
			wantStdout: "7",
			wantStderr: "",
		},
		{
			name:       "successful execution - multiple statements",
			script:     "print 1\nprint 2\nprint 3",
			wantExit:   0,
			wantStdout: "1\n2\n3",
			wantStderr: "",
		},
		{
			name:       "syntax error - unclosed block",
			script:     `print [unclosed`,
			wantExit:   2,
			wantStdout: "",
			wantStderr: "Syntax error",
		},
		{
			name:       "runtime error - division by zero",
			script:     `1 / 0`,
			wantExit:   1,
			wantStdout: "",
			wantStderr: "Math error",
		},
		{
			name:       "runtime error - undefined variable",
			script:     `print unknown_variable`,
			wantExit:   1,
			wantStdout: "",
			wantStderr: "No value for word",
		},
		{
			name:       "empty script",
			script:     "",
			wantExit:   0,
			wantStdout: "",
			wantStderr: "",
		},
		{
			name:       "script with comments",
			script:     "; This is a comment\nprint 42 ; inline comment",
			wantExit:   0,
			wantStdout: "42",
			wantStderr: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpfile, err := os.CreateTemp("", "test_*.viro")
			if err != nil {
				t.Fatal(err)
			}
			defer os.Remove(tmpfile.Name())

			if _, err := tmpfile.WriteString(tt.script); err != nil {
				t.Fatal(err)
			}
			tmpfile.Close()

			cmd := exec.Command(viroPath, tmpfile.Name())
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

func TestScriptExecutionFromFile(t *testing.T) {
	viroPath := "../../viro"
	if _, err := os.Stat(viroPath); os.IsNotExist(err) {
		t.Skip("viro binary not found, run 'make build' first")
	}

	scriptsDir := "../scripts"
	tests := []struct {
		name       string
		scriptFile string
		wantExit   int
		contains   []string
	}{
		{
			name:       "hello.viro",
			scriptFile: "hello.viro",
			wantExit:   0,
			contains:   []string{"Hello, World!"},
		},
		{
			name:       "math.viro",
			scriptFile: "math.viro",
			wantExit:   0,
			contains:   []string{"7", "50"},
		},
		{
			name:       "error.viro - syntax error",
			scriptFile: "error.viro",
			wantExit:   2,
			contains:   []string{"Syntax error"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scriptPath := filepath.Join(scriptsDir, tt.scriptFile)
			if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
				t.Skipf("test script %s not found", scriptPath)
			}

			cmd := exec.Command(viroPath, scriptPath)
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
			for _, want := range tt.contains {
				if !strings.Contains(outputStr, want) {
					t.Errorf("output missing %q:\nGot: %s", want, outputStr)
				}
			}
		})
	}
}

func TestScriptWithRelativePath(t *testing.T) {
	viroPath := "../../viro"
	if _, err := os.Stat(viroPath); os.IsNotExist(err) {
		t.Skip("viro binary not found, run 'make build' first")
	}

	script := `print "relative path test"`
	tmpfile, err := os.CreateTemp("", "test_*.viro")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.WriteString(script); err != nil {
		t.Fatal(err)
	}
	tmpfile.Close()

	cmd := exec.Command(viroPath, tmpfile.Name())
	output, err := cmd.CombinedOutput()

	if err != nil {
		t.Fatalf("execution failed: %v\nOutput: %s", err, output)
	}

	if !strings.Contains(string(output), "relative path test") {
		t.Errorf("output = %q, want to contain 'relative path test'", output)
	}
}

func TestScriptWithAbsolutePath(t *testing.T) {
	viroPath := "../../viro"
	if _, err := os.Stat(viroPath); os.IsNotExist(err) {
		t.Skip("viro binary not found, run 'make build' first")
	}

	script := `print "absolute path test"`
	tmpfile, err := os.CreateTemp("", "test_*.viro")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.WriteString(script); err != nil {
		t.Fatal(err)
	}
	tmpfile.Close()

	absPath, err := filepath.Abs(tmpfile.Name())
	if err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command(viroPath, absPath)
	output, err := cmd.CombinedOutput()

	if err != nil {
		t.Fatalf("execution failed: %v\nOutput: %s", err, output)
	}

	if !strings.Contains(string(output), "absolute path test") {
		t.Errorf("output = %q, want to contain 'absolute path test'", output)
	}
}

func TestScriptFileNotFound(t *testing.T) {
	viroPath := "../../viro"
	if _, err := os.Stat(viroPath); os.IsNotExist(err) {
		t.Skip("viro binary not found, run 'make build' first")
	}

	cmd := exec.Command(viroPath, "nonexistent_script.viro")
	output, err := cmd.CombinedOutput()

	if err == nil {
		t.Fatal("expected error for nonexistent file, got none")
	}

	exitErr, ok := err.(*exec.ExitError)
	if !ok {
		t.Fatalf("unexpected error type: %v", err)
	}

	if exitErr.ExitCode() != 1 {
		t.Errorf("exit code = %d, want 1 (general error)", exitErr.ExitCode())
	}

	outputStr := string(output)
	if !strings.Contains(strings.ToLower(outputStr), "no such file") &&
		!strings.Contains(strings.ToLower(outputStr), "not found") &&
		!strings.Contains(outputStr, "Error loading script") {
		t.Errorf("error message doesn't indicate file not found: %s", outputStr)
	}
}

func TestScriptWithQuietFlag(t *testing.T) {
	viroPath := "../../viro"
	if _, err := os.Stat(viroPath); os.IsNotExist(err) {
		t.Skip("viro binary not found, run 'make build' first")
	}

	script := `print "this should be suppressed"`
	tmpfile, err := os.CreateTemp("", "test_*.viro")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.WriteString(script); err != nil {
		t.Fatal(err)
	}
	tmpfile.Close()

	cmd := exec.Command(viroPath, "--quiet", tmpfile.Name())
	output, err := cmd.CombinedOutput()

	if err != nil {
		t.Fatalf("execution failed: %v\nOutput: %s", err, output)
	}

	if len(output) > 0 {
		t.Errorf("expected no output with --quiet flag, got: %s", output)
	}
}

func TestScriptWithVerboseFlag(t *testing.T) {
	viroPath := "../../viro"
	if _, err := os.Stat(viroPath); os.IsNotExist(err) {
		t.Skip("viro binary not found, run 'make build' first")
	}

	script := `print "verbose test"`
	tmpfile, err := os.CreateTemp("", "test_*.viro")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.WriteString(script); err != nil {
		t.Fatal(err)
	}
	tmpfile.Close()

	cmd := exec.Command(viroPath, "--verbose", tmpfile.Name())
	output, err := cmd.CombinedOutput()

	if err != nil {
		t.Fatalf("execution failed: %v\nOutput: %s", err, output)
	}

	if !strings.Contains(string(output), "verbose test") {
		t.Errorf("output = %q, want to contain 'verbose test'", output)
	}
}
