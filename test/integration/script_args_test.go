package integration

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/marcin-radoszewski/viro/internal/api"
)

func TestScriptArgumentsIntegration(t *testing.T) {
	testScript := "../scripts/test_args.viro"

	tests := []struct {
		name       string
		args       []string
		wantStdout []string
		wantStderr string
	}{
		{
			name: "no arguments",
			args: []string{testScript},
			wantStdout: []string{
				"Arguments received:",
				"Number of arguments: 0",
			},
			wantStderr: "",
		},
		{
			name: "single argument",
			args: []string{testScript, "hello"},
			wantStdout: []string{
				"Arguments received: hello",
				"Number of arguments: 1",
				"First argument: hello",
			},
			wantStderr: "",
		},
		{
			name: "multiple arguments",
			args: []string{testScript, "hello", "world"},
			wantStdout: []string{
				"Arguments received: hello world",
				"Number of arguments: 2",
				"First argument: hello",
				"Last argument: world",
			},
			wantStderr: "",
		},
		{
			name: "flag-like arguments",
			args: []string{testScript, "--verbose", "--output", "file.txt"},
			wantStdout: []string{
				"Arguments received: --verbose --output file.txt",
				"Number of arguments: 3",
				"First argument: --verbose",
				"Last argument: file.txt",
			},
			wantStderr: "",
		},
		{
			name: "numeric arguments",
			args: []string{testScript, "42", "3.14", "100"},
			wantStdout: []string{
				"Arguments received: 42 3.14 100",
				"Number of arguments: 3",
				"First argument: 42",
				"Last argument: 100",
			},
			wantStderr: "",
		},
		{
			name: "mixed arguments",
			args: []string{testScript, "hello", "42", "--flag"},
			wantStdout: []string{
				"Arguments received: hello 42 --flag",
				"Number of arguments: 3",
				"First argument: hello",
				"Last argument: --flag",
			},
			wantStderr: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			ctx := &api.RuntimeContext{
				Args:   tt.args,
				Stdin:  nil,
				Stdout: &stdout,
				Stderr: &stderr,
			}
			cfg, err := api.ConfigFromArgs(tt.args)
			if err != nil {
				t.Fatalf("ConfigFromArgs failed: %v", err)
			}

			exitCode := api.Run(ctx, cfg)
			output := stdout.String() + stderr.String()

			if tt.wantStderr != "" {
				if exitCode == api.ExitSuccess {
					t.Errorf("Expected error, got success")
				}
				if !strings.Contains(output, tt.wantStderr) {
					t.Errorf("Output = %q, want to contain %q", output, tt.wantStderr)
				}
				return
			}

			if exitCode != api.ExitSuccess {
				t.Fatalf("Command failed with exit code %d\nOutput: %s", exitCode, output)
			}

			for _, want := range tt.wantStdout {
				if !strings.Contains(output, want) {
					t.Errorf("Output missing expected line:\nWant: %q\nGot: %s", want, output)
				}
			}
		})
	}
}

func TestScriptArgumentsWithViroFlags(t *testing.T) {
	testScript := "../scripts/test_args.viro"

	tests := []struct {
		name       string
		args       []string
		wantStdout []string
	}{
		{
			name: "viro flags before script",
			args: []string{"--verbose", testScript, "arg1", "arg2"},
			wantStdout: []string{
				"Arguments received: arg1 arg2",
				"Number of arguments: 2",
				"First argument: arg1",
				"Last argument: arg2",
			},
		},
		{
			name:       "quiet flag suppresses output",
			args:       []string{"--quiet", testScript, "arg1"},
			wantStdout: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			ctx := &api.RuntimeContext{
				Args:   tt.args,
				Stdin:  nil,
				Stdout: &stdout,
				Stderr: &stderr,
			}
			cfg, err := api.ConfigFromArgs(tt.args)
			if err != nil {
				t.Fatalf("ConfigFromArgs failed: %v", err)
			}

			exitCode := api.Run(ctx, cfg)
			output := stdout.String() + stderr.String()

			if exitCode != api.ExitSuccess {
				t.Fatalf("Command failed with exit code %d\nOutput: %s", exitCode, output)
			}

			if len(tt.wantStdout) == 0 {
				if len(output) > 0 {
					t.Errorf("Expected no output with --quiet, got: %s", output)
				}
				return
			}

			for _, want := range tt.wantStdout {
				if !strings.Contains(output, want) {
					t.Errorf("Output missing expected line:\nWant: %q\nGot: %s", want, output)
				}
			}
		})
	}
}

func TestScriptArgumentsEmptyList(t *testing.T) {
	scriptContent := `print ["Args type:" type-of system.args]
print ["Args length:" length? system.args]`

	tmpfile, err := os.CreateTemp("", "test_empty_args_*.viro")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.WriteString(scriptContent); err != nil {
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
		t.Fatalf("Command failed with exit code %d\nOutput: %s", exitCode, output)
	}

	if !strings.Contains(output, "Args type: block!") {
		t.Errorf("Expected system.args to be block!, got: %s", output)
	}

	if !strings.Contains(output, "Args length: 0") {
		t.Errorf("Expected system.args length to be 0, got: %s", output)
	}
}
