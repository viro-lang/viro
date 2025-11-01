package integration

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestScriptArgumentsIntegration(t *testing.T) {
	viroPath := "../../viro"
	if _, err := os.Stat(viroPath); os.IsNotExist(err) {
		t.Skip("viro binary not found, run 'make build' first")
	}

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
			cmd := exec.Command(viroPath, tt.args...)
			output, err := cmd.CombinedOutput()

			if tt.wantStderr != "" {
				if err == nil {
					t.Errorf("Expected error, got none")
				}
				if !strings.Contains(string(output), tt.wantStderr) {
					t.Errorf("Output = %q, want to contain %q", string(output), tt.wantStderr)
				}
				return
			}

			if err != nil {
				t.Fatalf("Command failed: %v\nOutput: %s", err, output)
			}

			for _, want := range tt.wantStdout {
				if !strings.Contains(string(output), want) {
					t.Errorf("Output missing expected line:\nWant: %q\nGot: %s", want, output)
				}
			}
		})
	}
}

func TestScriptArgumentsWithViroFlags(t *testing.T) {
	viroPath := "../../viro"
	if _, err := os.Stat(viroPath); os.IsNotExist(err) {
		t.Skip("viro binary not found, run 'make build' first")
	}

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
			cmd := exec.Command(viroPath, tt.args...)
			output, err := cmd.CombinedOutput()

			if err != nil {
				t.Fatalf("Command failed: %v\nOutput: %s", err, output)
			}

			if len(tt.wantStdout) == 0 {
				if len(output) > 0 {
					t.Errorf("Expected no output with --quiet, got: %s", output)
				}
				return
			}

			for _, want := range tt.wantStdout {
				if !strings.Contains(string(output), want) {
					t.Errorf("Output missing expected line:\nWant: %q\nGot: %s", want, output)
				}
			}
		})
	}
}

func TestScriptArgumentsEmptyList(t *testing.T) {
	viroPath := "../../viro"
	if _, err := os.Stat(viroPath); os.IsNotExist(err) {
		t.Skip("viro binary not found, run 'make build' first")
	}

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

	cmd := exec.Command(viroPath, tmpfile.Name())
	output, err := cmd.CombinedOutput()

	if err != nil {
		t.Fatalf("Command failed: %v\nOutput: %s", err, output)
	}

	if !strings.Contains(string(output), "Args type: block!") {
		t.Errorf("Expected system.args to be block!, got: %s", output)
	}

	if !strings.Contains(string(output), "Args length: 0") {
		t.Errorf("Expected system.args length to be 0, got: %s", output)
	}
}
