// Package contract validates IO natives per contracts/io.md.
package contract

import (
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/marcin-radoszewski/viro/internal/core"
	"github.com/marcin-radoszewski/viro/internal/eval"
	"github.com/marcin-radoszewski/viro/internal/parse"
	"github.com/marcin-radoszewski/viro/internal/value"
	"github.com/marcin-radoszewski/viro/internal/verror"
)

// captureOutput configures evaluator output and returns captured output.
func captureOutput(t *testing.T, e *eval.Evaluator, fn func() (core.Value, error)) (string, core.Value, error) {
	t.Helper()

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe failed: %v", err)
	}

	// Configure evaluator to write to our pipe
	oldWriter := e.GetOutputWriter()
	e.SetOutputWriter(w)

	result, fnErr := fn()

	if err := w.Close(); err != nil {
		t.Fatalf("closing pipe writer failed: %v", err)
	}
	// Restore original writer
	e.SetOutputWriter(oldWriter)

	output, readErr := io.ReadAll(r)
	if readErr != nil {
		t.Fatalf("reading captured output failed: %v", readErr)
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
			vals, locations, perr := parse.ParseWithSource(tt.script, "(test)")
			if perr != nil {
				t.Fatalf("Parse failed: %v", perr)
			}

			e := NewTestEvaluator()

			captured, result, evalErr := captureOutput(t, e, func() (core.Value, error) {
				val, derr := e.DoBlock(vals, locations)
				if derr != nil {
					return value.NewNoneVal(), derr
				}
				return val, nil
			})

			if evalErr != nil {
				t.Fatalf("Unexpected evaluation error: %v", evalErr)
			}

			if captured != tt.expected {
				t.Fatalf("Unexpected stdout. want %q, got %q", tt.expected, captured)
			}

			if !result.Equals(value.NewNoneVal()) {
				t.Fatalf("print should return none, got %v", result)
			}
		})
	}
}

func TestIO_Prin(t *testing.T) {
	tests := []struct {
		name     string
		script   string
		expected string
	}{
		{
			name:     "prin integer",
			script:   "prin 42",
			expected: "42",
		},
		{
			name:     "prin string",
			script:   "prin \"hello\"",
			expected: "hello",
		},
		{
			name:     "prin reduced block",
			script:   "prin [1 + 1 3 * 4]",
			expected: "2 12",
		},
		{
			name:     "prin interpolated block",
			script:   "name: \"Alice\"\nprin [\"Hello\" name]",
			expected: "Hello Alice",
		},
		{
			name:     "prin logic true",
			script:   "prin true",
			expected: "true",
		},
		{
			name:     "prin none",
			script:   "prin none",
			expected: "none",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vals, locations, perr := parse.ParseWithSource(tt.script, "(test)")
			if perr != nil {
				t.Fatalf("Parse failed: %v", perr)
			}

			e := NewTestEvaluator()

			captured, result, evalErr := captureOutput(t, e, func() (core.Value, error) {
				val, derr := e.DoBlock(vals, locations)
				if derr != nil {
					return value.NewNoneVal(), derr
				}
				return val, nil
			})

			if evalErr != nil {
				t.Fatalf("Unexpected evaluation error: %v", evalErr)
			}

			if captured != tt.expected {
				t.Fatalf("Unexpected stdout. want %q, got %q", tt.expected, captured)
			}

			if !result.Equals(value.NewNoneVal()) {
				t.Fatalf("prin should return none, got %v", result)
			}
		})
	}
}

func TestIO_PrintPrin_WriteError(t *testing.T) {
	// Create a failing writer
	failingWriter := &failingWriter{}

	// Create evaluator with failing writer
	e := NewTestEvaluator()
	e.SetOutputWriter(failingWriter)

	tests := []struct {
		name   string
		script string
	}{
		{"print write error", "print 42"},
		{"prin write error", "prin 42"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vals, locations, perr := parse.ParseWithSource(tt.script, "(test)")
			if perr != nil {
				t.Fatalf("Parse failed: %v", perr)
			}

			_, evalErr := e.DoBlock(vals, locations)
			if evalErr == nil {
				t.Fatalf("Expected access error due to write failure, got nil")
			}

			// Check that it's an access error
			if vErr, ok := evalErr.(*verror.Error); ok {
				if vErr.Category != verror.ErrAccess {
					t.Fatalf("Expected access error, got %v error: %v", vErr.Category, evalErr)
				}
			} else {
				t.Fatalf("Expected verror.Error, got %T: %v", evalErr, evalErr)
			}
		})
	}
}

// failingWriter always returns an error on Write
type failingWriter struct{}

func (f *failingWriter) Write(p []byte) (n int, err error) {
	return 0, fmt.Errorf("simulated write error")
}

// flushTrackingWriter records writes and flush calls
type flushTrackingWriter struct {
	writes  []string
	flushes int
}

func (f *flushTrackingWriter) Write(p []byte) (n int, err error) {
	f.writes = append(f.writes, string(p))
	return len(p), nil
}

func (f *flushTrackingWriter) Flush() error {
	f.flushes++
	return nil
}

func TestIO_PrinFlush(t *testing.T) {
	// Test that prin triggers flush while print does not
	trackingWriter := &flushTrackingWriter{}

	e := NewTestEvaluator()
	e.SetOutputWriter(trackingWriter)

	tests := []struct {
		name           string
		script         string
		expectFlush    bool
		expectedOutput string
	}{
		{
			name:           "prin triggers flush",
			script:         "prin 42",
			expectFlush:    true,
			expectedOutput: "42",
		},
		{
			name:           "print does not trigger flush",
			script:         "print 42",
			expectFlush:    false,
			expectedOutput: "42\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset tracking writer state
			trackingWriter.writes = nil
			trackingWriter.flushes = 0

			vals, locations, perr := parse.ParseWithSource(tt.script, "(test)")
			if perr != nil {
				t.Fatalf("Parse failed: %v", perr)
			}

			_, evalErr := e.DoBlock(vals, locations)
			if evalErr != nil {
				t.Fatalf("Unexpected evaluation error: %v", evalErr)
			}

			// Check output
			if len(trackingWriter.writes) != 1 || trackingWriter.writes[0] != tt.expectedOutput {
				t.Fatalf("Unexpected output. want %q, got %q", tt.expectedOutput, trackingWriter.writes)
			}

			// Check flush behavior
			if tt.expectFlush && trackingWriter.flushes != 1 {
				t.Fatalf("Expected flush to be called once, got %d calls", trackingWriter.flushes)
			}
			if !tt.expectFlush && trackingWriter.flushes != 0 {
				t.Fatalf("Expected flush to not be called, got %d calls", trackingWriter.flushes)
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
			expected: value.NewStrVal("Alice"),
		},
		{
			name:     "numeric input remains string",
			provided: "123\n",
			expected: value.NewStrVal("123"),
		},
		{
			name:     "empty line",
			provided: "\n",
			expected: value.NewStrVal(""),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vals, locations, parseErr := parse.ParseWithSource("input", "(test)")
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
			result, evalErr := e.DoBlock(vals, locations)

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
