package integration

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/marcin-radoszewski/viro/internal/api"
)

func TestProfileFlag(t *testing.T) {
	tmpDir := t.TempDir()
	scriptPath := filepath.Join(tmpDir, "test.viro")

	script := `
factorial: fn [n] [
    if (= n 0) [
        1
    ] [
        * n (factorial (- n 1))
    ]
]

result: factorial 5
print result
`

	if err := os.WriteFile(scriptPath, []byte(script), 0644); err != nil {
		t.Fatalf("Failed to create test script: %v", err)
	}

	var stdout, stderr bytes.Buffer
	ctx := &api.RuntimeContext{
		Args:   []string{"--profile", scriptPath},
		Stdin:  nil,
		Stdout: &stdout,
		Stderr: &stderr,
	}
	cfg, err := api.ConfigFromArgs([]string{"--profile", scriptPath})
	if err != nil {
		t.Fatalf("ConfigFromArgs failed: %v", err)
	}

	exitCode := api.Run(ctx, cfg)
	output := stdout.String() + stderr.String()

	if exitCode != api.ExitSuccess {
		t.Fatalf("Profile execution failed with exit code %d\nOutput: %s", exitCode, output)
	}

	if !strings.Contains(output, "120") {
		t.Errorf("Expected factorial result '120' in output, got:\n%s", output)
	}

	if !strings.Contains(output, "EXECUTION PROFILE") {
		t.Errorf("Expected profile header in output, got:\n%s", output)
	}

	if !strings.Contains(output, "factorial") {
		t.Errorf("Expected 'factorial' function in profile, got:\n%s", output)
	}

	if !strings.Contains(output, "Total Execution Time:") {
		t.Errorf("Expected 'Total Execution Time' in profile, got:\n%s", output)
	}

	if !strings.Contains(output, "Function Statistics") {
		t.Errorf("Expected 'Function Statistics' in profile, got:\n%s", output)
	}

	if strings.Contains(output, `"timestamp"`) || strings.Contains(output, `"word"`) {
		t.Errorf("Profile output should not contain trace JSON, got:\n%s", output)
	}
}

func TestProfileWithQuietFlag(t *testing.T) {
	tmpDir := t.TempDir()
	scriptPath := filepath.Join(tmpDir, "test.viro")

	script := `
sum: fn [a b] [+ a b]
result: sum 10 20
print result
`

	if err := os.WriteFile(scriptPath, []byte(script), 0644); err != nil {
		t.Fatalf("Failed to create test script: %v", err)
	}

	var stdout, stderr bytes.Buffer
	ctx := &api.RuntimeContext{
		Args:   []string{"--profile", "--quiet", scriptPath},
		Stdin:  nil,
		Stdout: &stdout,
		Stderr: &stderr,
	}
	cfg, err := api.ConfigFromArgs([]string{"--profile", "--quiet", scriptPath})
	if err != nil {
		t.Fatalf("ConfigFromArgs failed: %v", err)
	}

	exitCode := api.Run(ctx, cfg)
	output := stdout.String() + stderr.String()

	if exitCode != api.ExitSuccess {
		t.Fatalf("Profile execution failed with exit code %d\nOutput: %s", exitCode, output)
	}

	if strings.Contains(output, "EXECUTION PROFILE") {
		t.Errorf("Profile output should be suppressed with --quiet flag, got:\n%s", output)
	}

	if strings.Contains(output, "30") {
		t.Errorf("Print output should be suppressed with --quiet flag, got:\n%s", output)
	}
}

func TestProfileFlagRequiresScript(t *testing.T) {
	var stdout, stderr bytes.Buffer
	ctx := &api.RuntimeContext{
		Args:   []string{"--profile", "-c", "3 + 4"},
		Stdin:  nil,
		Stdout: &stdout,
		Stderr: &stderr,
	}
	cfg, err := api.ConfigFromArgs([]string{"--profile", "-c", "3 + 4"})
	if err != nil {
		t.Fatalf("ConfigFromArgs failed: %v", err)
	}

	exitCode := api.Run(ctx, cfg)
	output := stdout.String() + stderr.String()

	if exitCode == api.ExitSuccess {
		t.Fatalf("Expected error when using --profile with -c, but command succeeded")
	}

	if !strings.Contains(output, "profile") && !strings.Contains(output, "script") {
		t.Errorf("Expected error message about profile requiring script, got:\n%s", output)
	}
}

func TestProfileWithComplexScript(t *testing.T) {
	tmpDir := t.TempDir()
	scriptPath := filepath.Join(tmpDir, "complex.viro")

	script := `
add: fn [a b] [+ a b]
multiply: fn [a b] [* a b]
power: fn [base exp] [
    if (= exp 0) [
        1
    ] [
        multiply base (power base (- exp 1))
    ]
]

result1: add 5 10
result2: multiply 3 4
result3: power 2 5

print result1
print result2
print result3
`

	if err := os.WriteFile(scriptPath, []byte(script), 0644); err != nil {
		t.Fatalf("Failed to create test script: %v", err)
	}

	var stdout, stderr bytes.Buffer
	ctx := &api.RuntimeContext{
		Args:   []string{"--profile", scriptPath},
		Stdin:  nil,
		Stdout: &stdout,
		Stderr: &stderr,
	}
	cfg, err := api.ConfigFromArgs([]string{"--profile", scriptPath})
	if err != nil {
		t.Fatalf("ConfigFromArgs failed: %v", err)
	}

	exitCode := api.Run(ctx, cfg)
	output := stdout.String() + stderr.String()

	if exitCode != api.ExitSuccess {
		t.Fatalf("Profile execution failed with exit code %d\nOutput: %s", exitCode, output)
	}

	expectedFunctions := []string{"add", "multiply", "power"}
	for _, fn := range expectedFunctions {
		if !strings.Contains(output, fn) {
			t.Errorf("Expected function '%s' in profile output, got:\n%s", fn, output)
		}
	}

	if !strings.Contains(output, "Total Events:") {
		t.Errorf("Expected 'Total Events' in profile, got:\n%s", output)
	}

	if !strings.Contains(output, "Calls") {
		t.Errorf("Expected 'Calls' column header in profile, got:\n%s", output)
	}

	if !strings.Contains(output, "Total Time") {
		t.Errorf("Expected 'Total Time' column header in profile, got:\n%s", output)
	}
}
