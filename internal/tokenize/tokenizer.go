package tokenize

import (
	"fmt"
	"strings"
)

type TokenType int

const (
	TokenLiteral TokenType = iota
	TokenString
	TokenBinary
	TokenLParen
	TokenRParen
	TokenLBracket
	TokenRBracket
	TokenEOF
)

type Token struct {
	Type   TokenType
	Value  string
	Line   int
	Column int
}

type Tokenizer struct {
	input  string
	pos    int
	line   int
	column int
}

func NewTokenizer(input string) *Tokenizer {
	return &Tokenizer{
		input:  input,
		pos:    0,
		line:   1,
		column: 1,
	}
}

func (t *Tokenizer) NextToken() (Token, error) {
	t.skipWhitespaceAndComments()

	if t.pos >= len(t.input) {
		return Token{Type: TokenEOF, Line: t.line, Column: t.column}, nil
	}

	ch := t.input[t.pos]
	tokenLine := t.line
	tokenColumn := t.column

	if ch == '@' || ch == '`' || ch == '~' {
		return Token{}, fmt.Errorf("invalid character '%c' at line %d, column %d", ch, tokenLine, tokenColumn)
	}

	if ch == '{' || ch == '}' {
		if ch == '{' && t.pos > 0 && t.input[t.pos-1] != '#' {
			return Token{}, fmt.Errorf("invalid character '%c' at line %d, column %d", ch, tokenLine, tokenColumn)
		}
		if ch == '}' {
			nextPos := t.pos + 1
			if nextPos < len(t.input) && t.input[nextPos] != ' ' && t.input[nextPos] != '\t' &&
			   t.input[nextPos] != '\n' && t.input[nextPos] != '\r' &&
			   t.input[nextPos] != '[' && t.input[nextPos] != ']' &&
			   t.input[nextPos] != '(' && t.input[nextPos] != ')' &&
			   t.input[nextPos] != ';' {
				return Token{}, fmt.Errorf("invalid character '%c' at line %d, column %d", ch, tokenLine, tokenColumn)
			}
		}
	}

	switch ch {
	case '[':
		t.advance()
		return Token{Type: TokenLBracket, Value: "[", Line: tokenLine, Column: tokenColumn}, nil
	case ']':
		t.advance()
		return Token{Type: TokenRBracket, Value: "]", Line: tokenLine, Column: tokenColumn}, nil
	case '(':
		t.advance()
		return Token{Type: TokenLParen, Value: "(", Line: tokenLine, Column: tokenColumn}, nil
	case ')':
		t.advance()
		return Token{Type: TokenRParen, Value: ")", Line: tokenLine, Column: tokenColumn}, nil
	case '"':
		str, err := t.readString()
		if err != nil {
			return Token{}, err
		}
		return Token{Type: TokenString, Value: str, Line: tokenLine, Column: tokenColumn}, nil
	case '#':
		if t.pos+1 < len(t.input) && t.input[t.pos+1] == '{' {
			bin, err := t.readBinary()
			if err != nil {
				return Token{}, err
			}
			return Token{Type: TokenBinary, Value: bin, Line: tokenLine, Column: tokenColumn}, nil
		}
		literal := t.readLiteral()
		return Token{Type: TokenLiteral, Value: literal, Line: tokenLine, Column: tokenColumn}, nil
	default:
		literal := t.readLiteral()
		return Token{Type: TokenLiteral, Value: literal, Line: tokenLine, Column: tokenColumn}, nil
	}
}

func (t *Tokenizer) Tokenize() ([]Token, error) {
	tokens := []Token{}

	for {
		token, err := t.NextToken()
		if err != nil {
			return nil, err
		}

		tokens = append(tokens, token)

		if token.Type == TokenEOF {
			break
		}
	}

	return tokens, nil
}

func (t *Tokenizer) skipWhitespaceAndComments() {
	for t.pos < len(t.input) {
		ch := t.input[t.pos]

		if ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r' {
			t.advance()
			continue
		}

		if ch == ';' {
			t.skipComment()
			continue
		}

		break
	}
}

func (t *Tokenizer) skipComment() {
	for t.pos < len(t.input) && t.input[t.pos] != '\n' {
		t.pos++
		t.column++
	}
	if t.pos < len(t.input) {
		t.advance()
	}
}

func (t *Tokenizer) readString() (string, error) {
	startLine := t.line
	startColumn := t.column

	t.advance()

	var result strings.Builder

	for t.pos < len(t.input) {
		ch := t.input[t.pos]

		if ch == '"' {
			t.advance()
			return result.String(), nil
		}

		if ch == '\\' {
			t.pos++
			if t.pos >= len(t.input) {
				return "", fmt.Errorf("unclosed string at line %d, column %d", startLine, startColumn)
			}

			escapedChar := t.input[t.pos]
			switch escapedChar {
			case 'n':
				result.WriteByte('\n')
			case 't':
				result.WriteByte('\t')
			case 'r':
				result.WriteByte('\r')
			case '\\':
				result.WriteByte('\\')
			case '"':
				result.WriteByte('"')
			default:
				return "", fmt.Errorf("invalid escape sequence '\\%c' at line %d, column %d", escapedChar, t.line, t.column)
			}
			t.advance()
			continue
		}

		result.WriteByte(ch)
		t.advance()
	}

	return "", fmt.Errorf("unclosed string at line %d, column %d", startLine, startColumn)
}

func (t *Tokenizer) readBinary() (string, error) {
	startLine := t.line
	startColumn := t.column

	t.advance()
	t.advance()

	var result strings.Builder

	for t.pos < len(t.input) {
		ch := t.input[t.pos]

		if ch == '}' {
			t.advance()
			return result.String(), nil
		}

		if ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r' {
			t.advance()
			continue
		}

		if (ch >= '0' && ch <= '9') || (ch >= 'a' && ch <= 'f') || (ch >= 'A' && ch <= 'F') {
			result.WriteByte(ch)
			t.advance()
			continue
		}

		return "", fmt.Errorf("invalid character in binary literal '%c' at line %d, column %d", ch, t.line, t.column)
	}

	return "", fmt.Errorf("unclosed binary literal at line %d, column %d", startLine, startColumn)
}

func (t *Tokenizer) readLiteral() string {
	start := t.pos

	for t.pos < len(t.input) {
		ch := t.input[t.pos]

		if ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r' ||
			ch == '[' || ch == ']' || ch == '(' || ch == ')' || ch == ';' ||
			ch == '{' || ch == '}' {
			break
		}

		if ch == '@' || ch == '`' || ch == '~' {
			break
		}

		if (ch == 'e' || ch == 'E') && t.pos > start {
			if t.pos+1 >= len(t.input) || (t.input[t.pos+1] != '+' && t.input[t.pos+1] != '-' && (t.input[t.pos+1] < '0' || t.input[t.pos+1] > '9')) {
				literal := t.input[start:t.pos]
				if strings.Contains(literal, ".") && strings.IndexAny(literal, "0123456789") == 0 {
					break
				}
			}
		}

		t.pos++
		t.column++
	}

	return t.input[start:t.pos]
}

func (t *Tokenizer) peek() byte {
	if t.pos >= len(t.input) {
		return 0
	}
	return t.input[t.pos]
}

func (t *Tokenizer) advance() {
	if t.pos < len(t.input) {
		if t.input[t.pos] == '\n' {
			t.line++
			t.column = 1
		} else {
			t.column++
		}
		t.pos++
	}
}
