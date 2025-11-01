package integration

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
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

	cmd := exec.Command("../../viro", "--profile", scriptPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Profile execution failed: %v\nOutput: %s", err, string(output))
	}

	outputStr := string(output)

	if !strings.Contains(outputStr, "120") {
		t.Errorf("Expected factorial result '120' in output, got:\n%s", outputStr)
	}

	if !strings.Contains(outputStr, "EXECUTION PROFILE") {
		t.Errorf("Expected profile header in output, got:\n%s", outputStr)
	}

	if !strings.Contains(outputStr, "factorial") {
		t.Errorf("Expected 'factorial' function in profile, got:\n%s", outputStr)
	}

	if !strings.Contains(outputStr, "Total Execution Time:") {
		t.Errorf("Expected 'Total Execution Time' in profile, got:\n%s", outputStr)
	}

	if !strings.Contains(outputStr, "Function Statistics") {
		t.Errorf("Expected 'Function Statistics' in profile, got:\n%s", outputStr)
	}

	if strings.Contains(outputStr, `"timestamp"`) || strings.Contains(outputStr, `"word"`) {
		t.Errorf("Profile output should not contain trace JSON, got:\n%s", outputStr)
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

	cmd := exec.Command("../../viro", "--profile", "--quiet", scriptPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Profile execution failed: %v\nOutput: %s", err, string(output))
	}

	outputStr := string(output)

	if strings.Contains(outputStr, "EXECUTION PROFILE") {
		t.Errorf("Profile output should be suppressed with --quiet flag, got:\n%s", outputStr)
	}

	if strings.Contains(outputStr, "30") {
		t.Errorf("Print output should be suppressed with --quiet flag, got:\n%s", outputStr)
	}
}

func TestProfileFlagRequiresScript(t *testing.T) {
	cmd := exec.Command("../../viro", "--profile", "-c", "3 + 4")
	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("Expected error when using --profile with -c, but command succeeded")
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "profile") && !strings.Contains(outputStr, "script") {
		t.Errorf("Expected error message about profile requiring script, got:\n%s", outputStr)
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

	cmd := exec.Command("../../viro", "--profile", scriptPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Profile execution failed: %v\nOutput: %s", err, string(output))
	}

	outputStr := string(output)

	expectedFunctions := []string{"add", "multiply", "power"}
	for _, fn := range expectedFunctions {
		if !strings.Contains(outputStr, fn) {
			t.Errorf("Expected function '%s' in profile output, got:\n%s", fn, outputStr)
		}
	}

	if !strings.Contains(outputStr, "Total Events:") {
		t.Errorf("Expected 'Total Events' in profile, got:\n%s", outputStr)
	}

	if !strings.Contains(outputStr, "Calls") {
		t.Errorf("Expected 'Calls' column header in profile, got:\n%s", outputStr)
	}

	if !strings.Contains(outputStr, "Total Time") {
		t.Errorf("Expected 'Total Time' column header in profile, got:\n%s", outputStr)
	}
}
