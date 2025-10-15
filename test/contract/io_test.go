// Package contract validates IO natives per contracts/io.md.
package contract

import (
	"io"
	"os"
	"testing"

	"github.com/marcin-radoszewski/viro/internal/core"
	"github.com/marcin-radoszewski/viro/internal/parse"
	"github.com/marcin-radoszewski/viro/internal/value"
)

// captureStdout redirects stdout during fn execution and returns captured output.
func captureStdout(t *testing.T, fn func() (core.Value, error)) (string, core.Value, error) {
	t.Helper()

	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe failed: %v", err)
	}
	os.Stdout = w

	result, fnErr := fn()

	if err := w.Close(); err != nil {
		t.Fatalf("closing pipe writer failed: %v", err)
	}
	os.Stdout = oldStdout

	output, readErr := io.ReadAll(r)
	if readErr != nil {
		t.Fatalf("reading captured stdout failed: %v", readErr)
	}
	if err := r.Close(); err != nil {
		t.Fatalf("closing pipe reader failed: %v", err)
	}

	return string(output), result, fnErr
}

func TestIO_Print(t *testing.T) {
	tests := []struct {
		name     string
		script   string
		expected string
	}{
		{
			name:     "print integer",
			script:   "print 42",
			expected: "42\n",
		},
		{
			name:     "print string",
			script:   "print \"hello\"",
			expected: "hello\n",
		},
		{
			name:     "print reduced block",
			script:   "print [1 + 1 3 * 4]",
			expected: "2 12\n",
		},
		{
			name:     "print interpolated block",
			script:   "name: \"Alice\"\nprint [\"Hello\" name]",
			expected: "Hello Alice\n",
		},
		{
			name:     "print logic true",
			script:   "print true",
			expected: "true\n",
		},
		{
			name:     "print none",
			script:   "print none",
			expected: "none\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vals, perr := parse.Parse(tt.script)
			if perr != nil {
				t.Fatalf("Parse failed: %v", perr)
			}

			e := NewTestEvaluator()

			captured, result, evalErr := captureStdout(t, func() (core.Value, error) {
				val, derr := e.Do_Blk(vals)
				if derr != nil {
					return value.NoneVal(), derr
				}
				return val, nil
			})

			if evalErr != nil {
				t.Fatalf("Unexpected evaluation error: %v", evalErr)
			}

			if captured != tt.expected {
				t.Fatalf("Unexpected stdout. want %q, got %q", tt.expected, captured)
			}

			if !result.Equals(value.NoneVal()) {
				t.Fatalf("print should return none, got %v", result)
			}
		})
	}
}

func TestIO_Input(t *testing.T) {
	tests := []struct {
		name     string
		provided string
		expected core.Value
	}{
		{
			name:     "simple word",
			provided: "Alice\n",
			expected: value.StrVal("Alice"),
		},
		{
			name:     "numeric input remains string",
			provided: "123\n",
			expected: value.StrVal("123"),
		},
		{
			name:     "empty line",
			provided: "\n",
			expected: value.StrVal(""),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vals, parseErr := parse.Parse("input")
			if parseErr != nil {
				t.Fatalf("Parse failed: %v", parseErr)
			}

			oldStdin := os.Stdin
			r, w, pipeErr := os.Pipe()
			if pipeErr != nil {
				t.Fatalf("os.Pipe failed: %v", pipeErr)
			}
			if _, writeErr := w.Write([]byte(tt.provided)); writeErr != nil {
				t.Fatalf("write to stdin pipe failed: %v", writeErr)
			}
			if err := w.Close(); err != nil {
				t.Fatalf("closing stdin pipe writer failed: %v", err)
			}
			os.Stdin = r

			e := NewTestEvaluator()
			result, evalErr := e.Do_Blk(vals)

			if err := r.Close(); err != nil {
				t.Fatalf("closing stdin pipe reader failed: %v", err)
			}
			os.Stdin = oldStdin

			if evalErr != nil {
				t.Fatalf("Unexpected evaluation error: %v", evalErr)
			}

			if !result.Equals(tt.expected) {
				t.Fatalf("input returned %v, want %v", result, tt.expected)
			}
		})
	}
}
