package integration

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/marcin-radoszewski/viro/internal/api"
)

func TestScriptExecution(t *testing.T) {
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

			var stdout, stderr bytes.Buffer
			ctx := &api.RuntimeContext{
				Args:   []string{tmpfile.Name()},
				Stdin:  nil,
				Stdout: &stdout,
				Stderr: &stderr,
			}
			cfg, err := api.ConfigFromArgs([]string{tmpfile.Name()})
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

func TestScriptExecutionFromFile(t *testing.T) {
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

			var stdout, stderr bytes.Buffer
			ctx := &api.RuntimeContext{
				Args:   []string{scriptPath},
				Stdin:  nil,
				Stdout: &stdout,
				Stderr: &stderr,
			}
			cfg, err := api.ConfigFromArgs([]string{scriptPath})
			if err != nil {
				t.Fatalf("ConfigFromArgs failed: %v", err)
			}

			exitCode := api.Run(ctx, cfg)
			output := stdout.String() + stderr.String()

			if exitCode != tt.wantExit {
				t.Errorf("exit code = %d, want %d\nOutput: %s", exitCode, tt.wantExit, output)
			}

			for _, want := range tt.contains {
				if !strings.Contains(output, want) {
					t.Errorf("output missing %q:\nGot: %s", want, output)
				}
			}
		})
	}
}

func TestScriptWithRelativePath(t *testing.T) {
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

	var stdout, stderr bytes.Buffer
	ctx := &api.RuntimeContext{
		Args:   []string{tmpfile.Name()},
		Stdin:  nil,
		Stdout: &stdout,
		Stderr: &stderr,
	}
	cfg, err := api.ConfigFromArgs([]string{tmpfile.Name()})
	if err != nil {
		t.Fatalf("ConfigFromArgs failed: %v", err)
	}

	exitCode := api.Run(ctx, cfg)
	output := stdout.String() + stderr.String()

	if exitCode != api.ExitSuccess {
		t.Fatalf("execution failed with exit code %d\nOutput: %s", exitCode, output)
	}

	if !strings.Contains(output, "relative path test") {
		t.Errorf("output = %q, want to contain 'relative path test'", output)
	}
}

func TestScriptWithAbsolutePath(t *testing.T) {
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

	var stdout, stderr bytes.Buffer
	ctx := &api.RuntimeContext{
		Args:   []string{absPath},
		Stdin:  nil,
		Stdout: &stdout,
		Stderr: &stderr,
	}
	cfg, err := api.ConfigFromArgs([]string{absPath})
	if err != nil {
		t.Fatalf("ConfigFromArgs failed: %v", err)
	}

	exitCode := api.Run(ctx, cfg)
	output := stdout.String() + stderr.String()

	if exitCode != api.ExitSuccess {
		t.Fatalf("execution failed with exit code %d\nOutput: %s", exitCode, output)
	}

	if !strings.Contains(output, "absolute path test") {
		t.Errorf("output = %q, want to contain 'absolute path test'", output)
	}
}

func TestScriptFileNotFound(t *testing.T) {
	var stdout, stderr bytes.Buffer
	ctx := &api.RuntimeContext{
		Args:   []string{"nonexistent_script.viro"},
		Stdin:  nil,
		Stdout: &stdout,
		Stderr: &stderr,
	}
	cfg, err := api.ConfigFromArgs([]string{"nonexistent_script.viro"})
	if err != nil {
		t.Fatalf("ConfigFromArgs failed: %v", err)
	}

	exitCode := api.Run(ctx, cfg)
	output := stdout.String() + stderr.String()

	if exitCode == api.ExitSuccess {
		t.Fatal("expected error for nonexistent file, got success")
	}

	if exitCode != api.ExitError {
		t.Errorf("exit code = %d, want %d (general error)", exitCode, api.ExitError)
	}

	if !strings.Contains(strings.ToLower(output), "no such file") &&
		!strings.Contains(strings.ToLower(output), "not found") &&
		!strings.Contains(output, "Error loading script") {
		t.Errorf("error message doesn't indicate file not found: %s", output)
	}
}

func TestScriptWithQuietFlag(t *testing.T) {
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

	var stdout, stderr bytes.Buffer
	ctx := &api.RuntimeContext{
		Args:   []string{"--quiet", tmpfile.Name()},
		Stdin:  nil,
		Stdout: &stdout,
		Stderr: &stderr,
	}
	cfg, err := api.ConfigFromArgs([]string{"--quiet", tmpfile.Name()})
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

func TestScriptWithVerboseFlag(t *testing.T) {
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

	var stdout, stderr bytes.Buffer
	ctx := &api.RuntimeContext{
		Args:   []string{"--verbose", tmpfile.Name()},
		Stdin:  nil,
		Stdout: &stdout,
		Stderr: &stderr,
	}
	cfg, err := api.ConfigFromArgs([]string{"--verbose", tmpfile.Name()})
	if err != nil {
		t.Fatalf("ConfigFromArgs failed: %v", err)
	}

	exitCode := api.Run(ctx, cfg)
	output := stdout.String() + stderr.String()

	if exitCode != api.ExitSuccess {
		t.Fatalf("execution failed with exit code %d\nOutput: %s", exitCode, output)
	}

	if !strings.Contains(output, "verbose test") {
		t.Errorf("output = %q, want to contain 'verbose test'", output)
	}
}
