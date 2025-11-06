package tokenize

import (
	"testing"
)

func TestTokenizer_EmptyInput(t *testing.T) {
	tokenizer := NewTokenizer("")
	tokens, err := tokenizer.Tokenize()

	if err != nil {
		t.Errorf("Tokenize() error = %v", err)
		return
	}

	if len(tokens) != 1 {
		t.Errorf("Expected 1 token (EOF), got %d", len(tokens))
		return
	}

	if tokens[0].Type != TokenEOF {
		t.Errorf("Expected TokenEOF, got %v", tokens[0].Type)
	}
}

func TestTokenizer_SingleLiteral(t *testing.T) {
	tokenizer := NewTokenizer("abc")
	tokens, err := tokenizer.Tokenize()

	if err != nil {
		t.Errorf("Tokenize() error = %v", err)
		return
	}

	expected := []Token{
		{Type: TokenLiteral, Value: "abc", Line: 1, Column: 1},
		{Type: TokenEOF, Line: 1, Column: 4},
	}

	if !tokensEqual(tokens, expected) {
		t.Errorf("Expected %v, got %v", expected, tokens)
	}
}

func TestTokenizer_MultipleLiterals(t *testing.T) {
	tokenizer := NewTokenizer("abc def")
	tokens, err := tokenizer.Tokenize()

	if err != nil {
		t.Errorf("Tokenize() error = %v", err)
		return
	}

	expected := []Token{
		{Type: TokenLiteral, Value: "abc", Line: 1, Column: 1},
		{Type: TokenLiteral, Value: "def", Line: 1, Column: 5},
		{Type: TokenEOF, Line: 1, Column: 8},
	}

	if !tokensEqual(tokens, expected) {
		t.Errorf("Expected %v, got %v", expected, tokens)
	}
}

func TestTokenizer_WhitespaceSeparation(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"space", "a b c"},
		{"tab", "a\tb\tc"},
		{"newline", "a\nb\nc"},
		{"mixed", "a \t\n b"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokenizer := NewTokenizer(tt.input)
			tokens, err := tokenizer.Tokenize()

			if err != nil {
				t.Errorf("Tokenize() error = %v", err)
				return
			}

			literalCount := 0
			for _, token := range tokens {
				if token.Type == TokenLiteral {
					literalCount++
				}
			}

			if literalCount != 3 && literalCount != 2 {
				t.Errorf("Expected 2-3 literals, got %d", literalCount)
			}
		})
	}
}

func TestTokenizer_Brackets(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []TokenType
	}{
		{"left bracket", "[", []TokenType{TokenLBracket, TokenEOF}},
		{"right bracket", "]", []TokenType{TokenRBracket, TokenEOF}},
		{"left paren", "(", []TokenType{TokenLParen, TokenEOF}},
		{"right paren", ")", []TokenType{TokenRParen, TokenEOF}},
		{"empty block", "[]", []TokenType{TokenLBracket, TokenRBracket, TokenEOF}},
		{"empty paren", "()", []TokenType{TokenLParen, TokenRParen, TokenEOF}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokenizer := NewTokenizer(tt.input)
			tokens, err := tokenizer.Tokenize()

			if err != nil {
				t.Errorf("Tokenize() error = %v", err)
				return
			}

			if len(tokens) != len(tt.expected) {
				t.Errorf("Expected %d tokens, got %d", len(tt.expected), len(tokens))
				return
			}

			for i, expected := range tt.expected {
				if tokens[i].Type != expected {
					t.Errorf("Token %d: expected %v, got %v", i, expected, tokens[i].Type)
				}
			}
		})
	}
}

func TestTokenizer_Strings(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple string", `"hello"`, "hello"},
		{"empty string", `""`, ""},
		{"string with spaces", `"hello world"`, "hello world"},
		{"string with escaped quotes", `"with \"quotes\""`, `with "quotes"`},
		{"string with newline escape", `"with\nnewline"`, "with\nnewline"},
		{"string with tab escape", `"with\ttab"`, "with\ttab"},
		{"string with backslash escape", `"with\\backslash"`, `with\backslash`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokenizer := NewTokenizer(tt.input)
			tokens, err := tokenizer.Tokenize()

			if err != nil {
				t.Errorf("Tokenize() error = %v", err)
				return
			}

			if len(tokens) < 1 || tokens[0].Type != TokenString {
				t.Errorf("Expected TokenString, got %v", tokens)
				return
			}

			if tokens[0].Value != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, tokens[0].Value)
			}
		})
	}
}

func TestTokenizer_Comments(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []Token
	}{
		{
			"comment only",
			"; this is a comment",
			[]Token{{Type: TokenEOF, Line: 1, Column: 20}},
		},
		{
			"comment before literal",
			"; comment\nabc",
			[]Token{
				{Type: TokenLiteral, Value: "abc", Line: 2, Column: 1},
				{Type: TokenEOF, Line: 2, Column: 4},
			},
		},
		{
			"literal before comment",
			"abc ; comment",
			[]Token{
				{Type: TokenLiteral, Value: "abc", Line: 1, Column: 1},
				{Type: TokenEOF, Line: 1, Column: 14},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokenizer := NewTokenizer(tt.input)
			tokens, err := tokenizer.Tokenize()

			if err != nil {
				t.Errorf("Tokenize() error = %v", err)
				return
			}

			if !tokensEqual(tokens, tt.expected) {
				t.Errorf("Expected %v, got %v", tt.expected, tokens)
			}
		})
	}
}

func TestTokenizer_Mixed(t *testing.T) {
	tokenizer := NewTokenizer(`abc [1 "test"] def`)
	tokens, err := tokenizer.Tokenize()

	if err != nil {
		t.Errorf("Tokenize() error = %v", err)
		return
	}

	expected := []Token{
		{Type: TokenLiteral, Value: "abc", Line: 1, Column: 1},
		{Type: TokenLBracket, Value: "[", Line: 1, Column: 5},
		{Type: TokenLiteral, Value: "1", Line: 1, Column: 6},
		{Type: TokenString, Value: "test", Line: 1, Column: 8},
		{Type: TokenRBracket, Value: "]", Line: 1, Column: 14},
		{Type: TokenLiteral, Value: "def", Line: 1, Column: 16},
		{Type: TokenEOF, Line: 1, Column: 19},
	}

	if !tokensEqual(tokens, expected) {
		t.Errorf("Expected %v, got %v", expected, tokens)
	}
}

func TestTokenizer_UnclosedString(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"unclosed at end", `"hello`},
		{"unclosed with newline", "\"hello\n"},
		{"unclosed with escape", `"hello\"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokenizer := NewTokenizer(tt.input)
			_, err := tokenizer.Tokenize()

			if err == nil {
				t.Errorf("Expected error for unclosed string, got nil")
			}
		})
	}
}

func TestTokenizer_InvalidEscape(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"invalid escape x", `"hello\x"`},
		{"invalid escape z", `"hello\z"`},
		{"invalid escape digit", `"hello\1"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokenizer := NewTokenizer(tt.input)
			_, err := tokenizer.Tokenize()

			if err == nil {
				t.Errorf("Expected error for invalid escape, got nil")
			}
		})
	}
}

func TestTokenizer_PositionTracking(t *testing.T) {
	input := "abc\ndef\n[1]"
	tokenizer := NewTokenizer(input)
	tokens, err := tokenizer.Tokenize()

	if err != nil {
		t.Errorf("Tokenize() error = %v", err)
		return
	}

	expected := []struct {
		typ    TokenType
		value  string
		line   int
		column int
	}{
		{TokenLiteral, "abc", 1, 1},
		{TokenLiteral, "def", 2, 1},
		{TokenLBracket, "[", 3, 1},
		{TokenLiteral, "1", 3, 2},
		{TokenRBracket, "]", 3, 3},
		{TokenEOF, "", 3, 4},
	}

	if len(tokens) != len(expected) {
		t.Errorf("Expected %d tokens, got %d", len(expected), len(tokens))
		return
	}

	for i, exp := range expected {
		token := tokens[i]
		if token.Type != exp.typ {
			t.Errorf("Token %d: expected type %v, got %v", i, exp.typ, token.Type)
		}
		if token.Value != exp.value {
			t.Errorf("Token %d: expected value %q, got %q", i, exp.value, token.Value)
		}
		if token.Line != exp.line {
			t.Errorf("Token %d: expected line %d, got %d", i, exp.line, token.Line)
		}
		if token.Column != exp.column {
			t.Errorf("Token %d: expected column %d, got %d", i, exp.column, token.Column)
		}
	}
}

func TestTokenizer_NextToken(t *testing.T) {
	tokenizer := NewTokenizer("abc def")

	token1, err := tokenizer.NextToken()
	if err != nil {
		t.Errorf("NextToken() error = %v", err)
	}
	if token1.Type != TokenLiteral || token1.Value != "abc" {
		t.Errorf("Expected Literal 'abc', got %v %q", token1.Type, token1.Value)
	}

	token2, err := tokenizer.NextToken()
	if err != nil {
		t.Errorf("NextToken() error = %v", err)
	}
	if token2.Type != TokenLiteral || token2.Value != "def" {
		t.Errorf("Expected Literal 'def', got %v %q", token2.Type, token2.Value)
	}

	token3, err := tokenizer.NextToken()
	if err != nil {
		t.Errorf("NextToken() error = %v", err)
	}
	if token3.Type != TokenEOF {
		t.Errorf("Expected EOF, got %v", token3.Type)
	}
}

func TestTokenizer_ComplexExpression(t *testing.T) {
	input := `abc: [1 2 3]
def: "hello world"
; comment here
result: (abc.1 + 5)`

	tokenizer := NewTokenizer(input)
	tokens, err := tokenizer.Tokenize()

	if err != nil {
		t.Errorf("Tokenize() error = %v", err)
		return
	}

	literalCount := 0
	stringCount := 0
	bracketCount := 0

	for _, token := range tokens {
		switch token.Type {
		case TokenLiteral:
			literalCount++
		case TokenString:
			stringCount++
		case TokenLBracket, TokenRBracket, TokenLParen, TokenRParen:
			bracketCount++
		}
	}

	if literalCount == 0 {
		t.Errorf("Expected some literals, got 0")
	}
	if stringCount != 1 {
		t.Errorf("Expected 1 string, got %d", stringCount)
	}
	if bracketCount != 4 {
		t.Errorf("Expected 4 brackets/parens, got %d", bracketCount)
	}
}

func TestTokenizer_SpecialCharactersInLiterals(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"colon suffix", "abc:", "abc:"},
		{"colon prefix", ":abc", ":abc"},
		{"quote prefix", "'abc", "'abc"},
		{"exclamation suffix", "integer!", "integer!"},
		{"dash in word", "--flag", "--flag"},
		{"dot in path", "obj.field", "obj.field"},
		{"multiple dots", "a.b.c", "a.b.c"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokenizer := NewTokenizer(tt.input)
			tokens, err := tokenizer.Tokenize()

			if err != nil {
				t.Errorf("Tokenize() error = %v", err)
				return
			}

			if len(tokens) < 1 || tokens[0].Type != TokenLiteral {
				t.Errorf("Expected TokenLiteral, got %v", tokens)
				return
			}

			if tokens[0].Value != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, tokens[0].Value)
			}
		})
	}
}

func tokensEqual(a, b []Token) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i].Type != b[i].Type || a[i].Value != b[i].Value ||
			a[i].Line != b[i].Line || a[i].Column != b[i].Column ||
			a[i].Source != b[i].Source {
			return false
		}
	}

	return true
}
