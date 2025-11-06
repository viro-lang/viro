package integration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/marcin-radoszewski/viro/internal/eval"
	"github.com/marcin-radoszewski/viro/internal/native"
	"github.com/marcin-radoszewski/viro/internal/parse"
)

func TestReadRefinements(t *testing.T) {
	tmpDir := t.TempDir()
	if err := eval.InitSandbox(tmpDir); err != nil {
		t.Fatalf("Failed to init sandbox: %v", err)
	}
	native.SandboxRoot = tmpDir

	testFile := filepath.Join(tmpDir, "test.txt")
	testContent := "line 1\nline 2\nline 3\nline 4\nline 5\n"
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	testBinaryFile := filepath.Join(tmpDir, "binary.dat")
	binaryData := []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09}
	if err := os.WriteFile(testBinaryFile, binaryData, 0644); err != nil {
		t.Fatalf("Failed to write binary test file: %v", err)
	}

	tests := []struct {
		name     string
		script   string
		expected string
	}{
		{
			name:     "Read as lines",
			script:   `read --lines "test.txt"`,
			expected: `["line 1" "line 2" "line 3" "line 4" "line 5"]`,
		},
		{
			name:     "Read partial bytes",
			script:   `read --part 6 "test.txt"`,
			expected: "line 1",
		},
		{
			name:     "Read partial lines",
			script:   `read --lines --part 3 "test.txt"`,
			expected: `["line 1" "line 2" "line 3"]`,
		},
		{
			name:     "Read with seek",
			script:   `read --seek 7 "test.txt"`,
			expected: "line 2",
		},
		{
			name:     "Read with seek and part",
			script:   `read --seek 7 --part 6 "test.txt"`,
			expected: "line 2",
		},
		{
			name:     "Read binary partial",
			script:   `form read --binary --part 5 "binary.dat"`,
			expected: "#{0001020304}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evaluator := NewTestEvaluator()
			vals, parseErr := parse.ParseWithSource(tt.script, "(test)")
			if parseErr != nil {
				t.Fatalf("Parse failed for %q: %v", tt.script, parseErr)
			}
			result, err := evaluator.DoBlock(vals)
			if err != nil {
				t.Fatalf("Evaluation error: %v", err)
			}

			resultStr := strings.TrimSpace(result.Mold())
			if !strings.Contains(resultStr, tt.expected) && resultStr != tt.expected {
				t.Errorf("Expected output to contain %q, got %q", tt.expected, resultStr)
			}
		})
	}
}

func TestReadRefinementsErrors(t *testing.T) {
	tmpDir := t.TempDir()
	if err := eval.InitSandbox(tmpDir); err != nil {
		t.Fatalf("Failed to init sandbox: %v", err)
	}
	native.SandboxRoot = tmpDir

	testFile := filepath.Join(tmpDir, "test.txt")
	testContent := "test content"
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	tests := []struct {
		name        string
		script      string
		shouldError bool
	}{
		{
			name:        "Binary and lines conflict",
			script:      `read --binary --lines "test.txt"`,
			shouldError: true,
		},
		{
			name:        "Negative seek position",
			script:      `read --seek -1 "test.txt"`,
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evaluator := NewTestEvaluator()
			vals, parseErr := parse.ParseWithSource(tt.script, "(test)")
			if parseErr != nil {
				t.Fatalf("Parse failed for %q: %v", tt.script, parseErr)
			}
			_, err := evaluator.DoBlock(vals)

			if tt.shouldError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.shouldError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestReadRefinementsEdgeCases(t *testing.T) {
	tmpDir := t.TempDir()
	if err := eval.InitSandbox(tmpDir); err != nil {
		t.Fatalf("Failed to init sandbox: %v", err)
	}
	native.SandboxRoot = tmpDir

	testFile := filepath.Join(tmpDir, "test.txt")
	testContent := "line1\n\n"
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	testFileNoNewline := filepath.Join(tmpDir, "test_no_newline.txt")
	testContentNoNewline := "line1"
	if err := os.WriteFile(testFileNoNewline, []byte(testContentNoNewline), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	tests := []struct {
		name     string
		script   string
		expected string
	}{
		{
			name:     "Lines with empty line preserved",
			script:   `read --lines "test.txt"`,
			expected: `["line1" ""]`,
		},
		{
			name:     "Lines without trailing newline",
			script:   `read --lines "test_no_newline.txt"`,
			expected: `["line1"]`,
		},
		{
			name:     "Seek to position 0",
			script:   `read --seek 0 "test_no_newline.txt"`,
			expected: `"line1"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evaluator := NewTestEvaluator()
			vals, parseErr := parse.ParseWithSource(tt.script, "(test)")
			if parseErr != nil {
				t.Fatalf("Parse failed for %q: %v", tt.script, parseErr)
			}
			result, err := evaluator.DoBlock(vals)
			if err != nil {
				t.Fatalf("Evaluation error: %v", err)
			}

			resultStr := strings.TrimSpace(result.Mold())
			if resultStr != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, resultStr)
			}
		})
	}
}
