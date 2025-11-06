package parse

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/ericlagergren/decimal"
	"github.com/marcin-radoszewski/viro/internal/core"
	"github.com/marcin-radoszewski/viro/internal/tokenize"
	"github.com/marcin-radoszewski/viro/internal/value"
)

type Parser struct {
	tokens []tokenize.Token
	pos    int
}

func NewParser(tokens []tokenize.Token) *Parser {
	return &Parser{
		tokens: tokens,
		pos:    0,
	}
}

func (p *Parser) Parse() ([]core.Value, error) {
	values := []core.Value{}

	for p.pos < len(p.tokens) {
		token := p.tokens[p.pos]

		if token.Type == tokenize.TokenEOF {
			break
		}

		val, err := p.parseValue()
		if err != nil {
			return nil, err
		}
		values = append(values, val)
	}

	return values, nil
}

func (p *Parser) parseValue() (core.Value, error) {
	if p.pos >= len(p.tokens) {
		return nil, fmt.Errorf("unexpected end of input")
	}

	token := p.tokens[p.pos]
	p.pos++

	switch token.Type {
	case tokenize.TokenLiteral:
		return p.classifyLiteral(token.Value)

	case tokenize.TokenString:
		return value.NewStrVal(token.Value), nil

	case tokenize.TokenBinary:
		return p.parseBinary(token.Value)

	case tokenize.TokenLBracket:
		values, err := p.parseUntil(tokenize.TokenRBracket, "block")
		if err != nil {
			return nil, err
		}
		return value.NewBlockVal(values), nil

	case tokenize.TokenLParen:
		values, err := p.parseUntil(tokenize.TokenRParen, "paren")
		if err != nil {
			return nil, err
		}
		return value.NewParenVal(values), nil

	case tokenize.TokenRBracket, tokenize.TokenRParen:
		return nil, fmt.Errorf("unexpected closing bracket at line %d, column %d", token.Line, token.Column)

	case tokenize.TokenEOF:
		return nil, fmt.Errorf("unexpected EOF")

	default:
		return nil, fmt.Errorf("unknown token type")
	}
}

func (p *Parser) parseUntil(closingType tokenize.TokenType, structName string) ([]core.Value, error) {
	values := []core.Value{}

	for p.pos < len(p.tokens) {
		token := p.tokens[p.pos]

		if token.Type == closingType {
			p.pos++
			return values, nil
		}

		if token.Type == tokenize.TokenEOF {
			return nil, fmt.Errorf("unclosed %s", structName)
		}

		val, err := p.parseValue()
		if err != nil {
			return nil, err
		}
		values = append(values, val)
	}

	return nil, fmt.Errorf("unclosed %s", structName)
}

func (p *Parser) classifyLiteral(text string) (core.Value, error) {
	if strings.HasPrefix(text, "'") {
		return value.NewLitWordVal(text[1:]), nil
	}

	if strings.HasPrefix(text, ":") {
		base := text[1:]
		if len(base) > 0 && base[0] >= '0' && base[0] <= '9' {
			return nil, fmt.Errorf("get-words cannot start with numbers: %s", text)
		}
		if strings.Contains(base, ".") {
			segments, err := p.parsePath(base)
			if err != nil {
				return nil, err
			}
			return value.GetPathVal(value.NewGetPath(segments, value.NewNoneVal())), nil
		}
		return value.NewGetWordVal(base), nil
	}

	if strings.HasSuffix(text, ":") {
		base := text[:len(text)-1]
		if len(base) > 0 && base[0] >= '0' && base[0] <= '9' {
			return nil, fmt.Errorf("set-words cannot start with numbers: %s", text)
		}
		if strings.Contains(base, ".") {
			segments, err := p.parsePath(base)
			if err != nil {
				return nil, err
			}
			return value.SetPathVal(value.NewSetPath(segments, value.NewNoneVal())), nil
		}
		return value.NewSetWordVal(base), nil
	}

	intPattern := regexp.MustCompile(`^-?[0-9]+$`)
	if intPattern.MatchString(text) {
		n, _ := strconv.ParseInt(text, 10, 64)
		return value.NewIntVal(n), nil
	}

	decimalPattern := regexp.MustCompile(`^-?[0-9]+\.[0-9]+([eE][+-]?[0-9]+)?$`)
	if decimalPattern.MatchString(text) {
		d := new(decimal.Big)
		d.SetString(text)
		scale := calculateScale(text)
		return value.DecimalVal(d, scale), nil
	}

	scientificPattern := regexp.MustCompile(`^-?[0-9]+[eE][+-]?[0-9]+$`)
	if scientificPattern.MatchString(text) {
		d := new(decimal.Big)
		d.SetString(text)
		scale := calculateScale(text)
		return value.DecimalVal(d, scale), nil
	}

	if strings.HasSuffix(text, "!") {
		return value.NewDatatypeVal(text), nil
	}

	if strings.Contains(text, ".") {
		firstChar := text[0]
		if firstChar >= '0' && firstChar <= '9' {
			return nil, fmt.Errorf("invalid number format: %s", text)
		}

		segments, err := p.parsePath(text)
		if err != nil {
			return nil, err
		}
		return value.PathVal(value.NewPath(segments, value.NewNoneVal())), nil
	}

	return value.NewWordVal(text), nil
}

func (p *Parser) parsePath(text string) ([]value.PathSegment, error) {
	if text == "" {
		return nil, fmt.Errorf("empty path")
	}

	parts := strings.Split(text, ".")
	segments := make([]value.PathSegment, 0, len(parts))

	for i, part := range parts {
		if part == "" {
			return nil, fmt.Errorf("empty path segment")
		}

		if n, err := strconv.ParseInt(part, 10, 64); err == nil {
			if i == 0 {
				return nil, fmt.Errorf("paths cannot start with numbers")
			}
			segments = append(segments, value.PathSegment{
				Type:  value.PathSegmentIndex,
				Value: n,
			})
		} else {
			segments = append(segments, value.PathSegment{
				Type:  value.PathSegmentWord,
				Value: part,
			})
		}
	}

	return segments, nil
}

func (p *Parser) parseBinary(hexStr string) (core.Value, error) {
	hexStr = strings.ReplaceAll(hexStr, " ", "")
	hexStr = strings.ReplaceAll(hexStr, "\t", "")
	hexStr = strings.ReplaceAll(hexStr, "\n", "")
	hexStr = strings.ReplaceAll(hexStr, "\r", "")

	if len(hexStr)%2 != 0 {
		return nil, fmt.Errorf("binary literal must have even number of hex digits")
	}

	bytes := make([]byte, len(hexStr)/2)
	for i := 0; i < len(hexStr); i += 2 {
		high := hexDigitToInt(hexStr[i])
		low := hexDigitToInt(hexStr[i+1])
		if high == -1 || low == -1 {
			return nil, fmt.Errorf("invalid hex digit in binary literal")
		}
		bytes[i/2] = byte(high<<4 | low)
	}

	return value.NewBinaryVal(bytes), nil
}

func hexDigitToInt(ch byte) int {
	if ch >= '0' && ch <= '9' {
		return int(ch - '0')
	}
	if ch >= 'a' && ch <= 'f' {
		return int(ch - 'a' + 10)
	}
	if ch >= 'A' && ch <= 'F' {
		return int(ch - 'A' + 10)
	}
	return -1
}

func calculateScale(text string) int16 {
	parts := strings.Split(text, ".")
	if len(parts) < 2 {
		return 0
	}

	decimalPart := parts[1]

	eIndex := strings.IndexAny(decimalPart, "eE")
	if eIndex >= 0 {
		decimalPart = decimalPart[:eIndex]
	}

	return int16(len(decimalPart))
}
