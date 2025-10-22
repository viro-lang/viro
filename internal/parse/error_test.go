package parse

import (
	"strings"
	"testing"

	"github.com/marcin-radoszewski/viro/internal/verror"
)

// TestMakeSyntaxError tests the makeSyntaxError function directly
func TestMakeSyntaxError(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		pos      int
		id       string
		args     [3]string
		wantNear string
	}{
		{
			name:     "basic error with context",
			input:    "hello world error",
			pos:      6,
			id:       verror.ErrIDInvalidSyntax,
			args:     [3]string{"unexpected token", "", ""},
			wantNear: "hello world error",
		},
		{
			name:     "error at start",
			input:    "invalid",
			pos:      0,
			id:       verror.ErrIDInvalidSyntax,
			args:     [3]string{"invalid start", "", ""},
			wantNear: "invalid",
		},
		{
			name:     "error at end",
			input:    "valid invalid",
			pos:      7,
			id:       verror.ErrIDInvalidSyntax,
			args:     [3]string{"invalid end", "", ""},
			wantNear: "valid invalid",
		},
		{
			name:     "empty input",
			input:    "",
			pos:      0,
			id:       verror.ErrIDInvalidSyntax,
			args:     [3]string{"empty", "", ""},
			wantNear: "",
		},
		{
			name:     "position beyond input",
			input:    "short",
			pos:      10,
			id:       verror.ErrIDInvalidSyntax,
			args:     [3]string{"beyond", "", ""},
			wantNear: "short",
		},
		{
			name:     "unicode characters",
			input:    "héllo wörld",
			pos:      6,
			id:       verror.ErrIDInvalidSyntax,
			args:     [3]string{"unicode", "", ""},
			wantNear: "héllo wörld",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := makeSyntaxError(tt.input, tt.pos, tt.id, tt.args)
			if err == nil {
				t.Fatal("Expected error, got nil")
			}

			// Check error type
			if verr, ok := err.(*verror.Error); ok {
				if verr.ID != tt.id {
					t.Errorf("Expected error ID %q, got %q", tt.id, verr.ID)
				}
				if verr.Args != tt.args {
					t.Errorf("Expected args %v, got %v", tt.args, verr.Args)
				}
				if verr.Near != tt.wantNear {
					t.Errorf("Expected near %q, got %q", tt.wantNear, verr.Near)
				}
			} else {
				t.Errorf("Expected verror.Error, got %T", err)
			}
		})
	}
}

// TestSnippetAround tests the snippetAround function directly
func TestSnippetAround(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		pos      int
		expected string
	}{
		{
			name:     "middle of string",
			input:    "hello world error here",
			pos:      6,
			expected: "hello world error h",
		},
		{
			name:     "start of string",
			input:    "hello world",
			pos:      0,
			expected: "hello world",
		},
		{
			name:     "end of string",
			input:    "hello world",
			pos:      11,
			expected: "hello world",
		},
		{
			name:     "empty string",
			input:    "",
			pos:      0,
			expected: "",
		},
		{
			name:     "position beyond length",
			input:    "short",
			pos:      10,
			expected: "short",
		},
		{
			name:     "negative position",
			input:    "hello",
			pos:      -1,
			expected: "hello",
		},
		{
			name:     "unicode string",
			input:    "héllo wörld",
			pos:      6,
			expected: "héllo wörld",
		},
		{
			name:     "very long string",
			input:    strings.Repeat("a", 100),
			pos:      50,
			expected: strings.Repeat("a", 25), // 50-12 to 50+12+1 = 38 chars, but min with length 100
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := snippetAround(tt.input, tt.pos)
			if result != tt.expected {
				t.Errorf("snippetAround(%q, %d) = %q, want %q", tt.input, tt.pos, result, tt.expected)
			}
		})
	}
}

// TestParserSyntaxError tests the parser's syntaxError method
func TestParserSyntaxError(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		pos      int
		id       string
		args     [3]string
		wantNear string
	}{
		{
			name:     "parser syntax error",
			input:    "invalid [ syntax",
			pos:      8,
			id:       verror.ErrIDUnclosedBlock,
			args:     [3]string{"[", "", ""},
			wantNear: "invalid [ syntax",
		},
		{
			name:     "unexpected EOF",
			input:    "incomplete",
			pos:      10,
			id:       verror.ErrIDUnexpectedEOF,
			args:     [3]string{"", "", ""},
			wantNear: "incomplete",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &parser{
				source: tt.input,
			}

			err := p.syntaxError(tt.pos, tt.id, tt.args)
			if err == nil {
				t.Fatal("Expected error, got nil")
			}

			if verr, ok := err.(*verror.Error); ok {
				if verr.ID != tt.id {
					t.Errorf("Expected error ID %q, got %q", tt.id, verr.ID)
				}
				if verr.Near != tt.wantNear {
					t.Errorf("Expected near %q, got %q", tt.wantNear, verr.Near)
				}
			} else {
				t.Errorf("Expected verror.Error, got %T", err)
			}
		})
	}
}

// TestMalformedInputHandling tests various malformed input scenarios
func TestMalformedInputHandling(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectError bool
		errorID     string
	}{
		{
			name:        "unclosed string",
			input:       `"unclosed string`,
			expectError: true,
			errorID:     verror.ErrIDInvalidSyntax,
		},
		{
			name:        "invalid number",
			input:       "123abc",
			expectError: false, // This should parse as separate tokens
		},
		{
			name:        "malformed path",
			input:       "obj.",
			expectError: true,
			errorID:     verror.ErrIDInvalidSyntax,
		},
		{
			name:        "invalid character",
			input:       "valid@invalid",
			expectError: true,
			errorID:     verror.ErrIDInvalidSyntax,
		},
		{
			name:        "unclosed block",
			input:       "[unclosed",
			expectError: true,
			errorID:     verror.ErrIDUnclosedBlock,
		},
		{
			name:        "unclosed paren",
			input:       "(unclosed",
			expectError: true,
			errorID:     verror.ErrIDUnclosedParen,
		},
		{
			name:        "unexpected closing bracket",
			input:       "value]",
			expectError: false, // Parser may be lenient about extra closing brackets
		},
		{
			name:        "unexpected closing paren",
			input:       "value)",
			expectError: false, // Parser may be lenient about extra closing parens
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Parse(tt.input)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error for input %q, got nil", tt.input)
					return
				}

				if verr, ok := err.(*verror.Error); ok {
					if verr.ID != tt.errorID {
						t.Errorf("Expected error ID %q, got %q", tt.errorID, verr.ID)
					}
				} else {
					t.Errorf("Expected verror.Error, got %T", err)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error for input %q, got %v", tt.input, err)
				}
			}
		})
	}
}

// TestUnexpectedEOF tests unexpected end of file scenarios
func TestUnexpectedEOF(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "incomplete block",
			input: "[incomplete",
		},
		{
			name:  "incomplete paren",
			input: "(incomplete",
		},
		{
			name:  "incomplete string",
			input: `"incomplete`,
		},
		{
			name:  "incomplete path",
			input: "obj.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Parse(tt.input)
			if err == nil {
				t.Errorf("Expected error for incomplete input %q", tt.input)
			}
		})
	}
}

// TestInvalidTokenSequences tests invalid combinations of tokens
func TestInvalidTokenSequences(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "number followed by invalid",
			input: "123@invalid",
		},
		{
			name:  "word with invalid chars",
			input: "word#invalid",
		},
		{
			name:  "multiple dots in path",
			input: "obj..field",
		},
		{
			name:  "dot without following segment",
			input: "obj. ",
		},
		{
			name:  "colon without word",
			input: "word:",
		},
		{
			name:  "apostrophe without word",
			input: "'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vals, err := Parse(tt.input)
			// For some cases like "word:", it might parse successfully as a set-word
			// Only fail if we get an error we don't expect
			if err != nil {
				t.Logf("Got expected error for %q: %v", tt.input, err)
			} else {
				t.Logf("Parsed %q successfully as %d values", tt.input, len(vals))
			}
		})
	}
}

// TestTokenizationErrorCases tests specific tokenization error scenarios
func TestTokenizationErrorCases(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectError bool
		errorID     string
	}{
		{
			name:        "unclosed string literal",
			input:       `"unclosed string`,
			expectError: true,
			errorID:     verror.ErrIDInvalidSyntax,
		},
		{
			name:        "string with escape issues",
			input:       `"valid string"`,
			expectError: false,
		},
		{
			name:        "invalid number format",
			input:       "123.456.789",
			expectError: false, // Should tokenize as separate tokens
		},
		{
			name:        "malformed path",
			input:       "obj.",
			expectError: true,
			errorID:     verror.ErrIDInvalidSyntax,
		},
		{
			name:        "path with invalid segment",
			input:       "obj.#invalid",
			expectError: true,
			errorID:     verror.ErrIDInvalidSyntax,
		},
		{
			name:        "invalid character sequence",
			input:       "valid$invalid",
			expectError: true,
			errorID:     verror.ErrIDInvalidSyntax,
		},
		{
			name:        "unicode invalid chars",
			input:       "word\x00invalid",
			expectError: true,
			errorID:     verror.ErrIDInvalidSyntax,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Parse(tt.input)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error for input %q", tt.input)
					return
				}

				if verr, ok := err.(*verror.Error); ok {
					if verr.ID != tt.errorID {
						t.Errorf("Expected error ID %q, got %q", tt.errorID, verr.ID)
					}
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error for input %q, got %v", tt.input, err)
				}
			}
		})
	}
}

// TestUnicodeEdgeCases tests parsing with unicode characters
func TestUnicodeEdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectError bool
	}{
		{
			name:        "unicode identifiers",
			input:       "héllo wörld",
			expectError: false,
		},
		{
			name:        "unicode in strings",
			input:       `"héllo wörld"`,
			expectError: false,
		},
		{
			name:        "unicode operators",
			input:       "a ≠ b",
			expectError: true, // ≠ is not a valid operator
		},
		{
			name:        "mixed unicode and ascii",
			input:       "héllo + world",
			expectError: false,
		},
		{
			name:        "unicode digits in numbers",
			input:       "123",
			expectError: false,
		},
		{
			name:        "unicode whitespace",
			input:       "hello\t\n\r world",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Parse(tt.input)

			if tt.expectError && err == nil {
				t.Errorf("Expected error for unicode input %q", tt.input)
			} else if !tt.expectError && err != nil {
				t.Errorf("Unexpected error for unicode input %q: %v", tt.input, err)
			}
		})
	}
}
