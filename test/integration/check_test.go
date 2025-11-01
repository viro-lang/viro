package integration

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/marcin-radoszewski/viro/internal/api"
)

func TestCheckMode(t *testing.T) {
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

			var stdout, stderr bytes.Buffer
			ctx := &api.RuntimeContext{
				Args:   []string{"--check", tmpfile.Name()},
				Stdin:  &bytes.Buffer{},
				Stdout: &stdout,
				Stderr: &stderr,
			}

			cfg := mustConfigFromArgs(t, []string{"--check", tmpfile.Name()})
			exitCode := api.Run(ctx, cfg)

			output := stdout.String() + stderr.String()

			if exitCode != tt.wantExit {
				t.Errorf("exit code = %d, want %d\nOutput: %s", exitCode, tt.wantExit, output)
			}

			if tt.wantOutput != "" {
				if !strings.Contains(output, tt.wantOutput) {
					t.Errorf("output = %q, want to contain %q", output, tt.wantOutput)
				}
			}
		})
	}
}

func TestCheckModeDoesNotExecute(t *testing.T) {
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

	var stdout, stderr bytes.Buffer
	ctx := &api.RuntimeContext{
		Args:   []string{"--check", tmpfile.Name()},
		Stdin:  &bytes.Buffer{},
		Stdout: &stdout,
		Stderr: &stderr,
	}

	cfg := mustConfigFromArgs(t, []string{"--check", tmpfile.Name()})
	exitCode := api.Run(ctx, cfg)

	if exitCode != 0 {
		t.Fatalf("check failed on valid syntax: exit code %d\nOutput: %s", exitCode, stdout.String()+stderr.String())
	}

	outputStr := stdout.String() + stderr.String()
	if strings.Contains(outputStr, "This should not be printed") {
		t.Error("--check mode executed the script (found print output)")
	}

	if strings.Contains(outputStr, "Math error") {
		t.Error("--check mode executed the script (found runtime error)")
	}
}

func TestCheckModeWithVerbose(t *testing.T) {
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

	var stdout, stderr bytes.Buffer
	ctx := &api.RuntimeContext{
		Args:   []string{"--check", "--verbose", tmpfile.Name()},
		Stdin:  &bytes.Buffer{},
		Stdout: &stdout,
		Stderr: &stderr,
	}

	cfg := mustConfigFromArgs(t, []string{"--check", "--verbose", tmpfile.Name()})
	exitCode := api.Run(ctx, cfg)

	if exitCode != 0 {
		t.Fatalf("check failed: exit code %d\nOutput: %s", exitCode, stdout.String()+stderr.String())
	}

	outputStr := stdout.String() + stderr.String()
	if !strings.Contains(outputStr, "Syntax valid") &&
		!strings.Contains(outputStr, "Parsed") {
		t.Logf("Note: --verbose output for --check mode: %s", outputStr)
	}
}

func TestCheckModeFromExistingScripts(t *testing.T) {
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

			var stdout, stderr bytes.Buffer
			ctx := &api.RuntimeContext{
				Args:   []string{"--check", tt.scriptPath},
				Stdin:  &bytes.Buffer{},
				Stdout: &stdout,
				Stderr: &stderr,
			}

			cfg := mustConfigFromArgs(t, []string{"--check", tt.scriptPath})
			exitCode := api.Run(ctx, cfg)

			if exitCode != tt.wantExit {
				t.Errorf("exit code = %d, want %d\nOutput: %s", exitCode, tt.wantExit, stdout.String()+stderr.String())
			}
		})
	}
}

func TestCheckModeMultipleErrors(t *testing.T) {
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

	var stdout, stderr bytes.Buffer
	ctx := &api.RuntimeContext{
		Args:   []string{"--check", tmpfile.Name()},
		Stdin:  &bytes.Buffer{},
		Stdout: &stdout,
		Stderr: &stderr,
	}

	cfg := mustConfigFromArgs(t, []string{"--check", tmpfile.Name()})
	exitCode := api.Run(ctx, cfg)

	if exitCode == 0 {
		t.Fatal("expected syntax error, got exit code 0")
	}

	if exitCode != 2 {
		t.Errorf("exit code = %d, want 2 (syntax error)", exitCode)
	}

	output := stdout.String() + stderr.String()
	if !strings.Contains(output, "Syntax error") {
		t.Errorf("output = %q, want to contain 'Syntax error'", output)
	}
}

func TestCheckModeFileNotFound(t *testing.T) {
	var stdout, stderr bytes.Buffer
	ctx := &api.RuntimeContext{
		Args:   []string{"--check", "nonexistent_file.viro"},
		Stdin:  &bytes.Buffer{},
		Stdout: &stdout,
		Stderr: &stderr,
	}

	cfg := mustConfigFromArgs(t, []string{"--check", "nonexistent_file.viro"})
	exitCode := api.Run(ctx, cfg)

	if exitCode == 0 {
		t.Fatal("expected error for nonexistent file, got exit code 0")
	}

	if exitCode != 1 {
		t.Errorf("exit code = %d, want 1 (file not found error)", exitCode)
	}

	outputStr := stdout.String() + stderr.String()
	if !strings.Contains(strings.ToLower(outputStr), "no such file") &&
		!strings.Contains(strings.ToLower(outputStr), "not found") &&
		!strings.Contains(outputStr, "Error loading script") {
		t.Errorf("error message doesn't indicate file not found: %s", outputStr)
	}
}

func TestCheckModeWithQuiet(t *testing.T) {
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

	var stdout, stderr bytes.Buffer
	ctx := &api.RuntimeContext{
		Args:   []string{"--check", "--quiet", tmpfile.Name()},
		Stdin:  &bytes.Buffer{},
		Stdout: &stdout,
		Stderr: &stderr,
	}

	cfg := mustConfigFromArgs(t, []string{"--check", "--quiet", tmpfile.Name()})
	exitCode := api.Run(ctx, cfg)

	if exitCode != 0 {
		t.Fatalf("check failed: exit code %d\nOutput: %s", exitCode, stdout.String()+stderr.String())
	}

	output := stdout.String() + stderr.String()
	if len(output) > 0 {
		t.Logf("Note: --quiet with --check produced output: %s", output)
	}
}
