package parse

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/ericlagergren/decimal"
	"github.com/marcin-radoszewski/viro/internal/core"
	"github.com/marcin-radoszewski/viro/internal/tokenize"
	"github.com/marcin-radoszewski/viro/internal/value"
	"github.com/marcin-radoszewski/viro/internal/verror"
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
		return nil, verror.NewSyntaxError(verror.ErrIDUnexpectedEOF, [3]string{"", "", ""})
	}

	token := p.tokens[p.pos]
	p.pos++

	switch token.Type {
	case tokenize.TokenLiteral:
		return p.ClassifyLiteral(token.Value)

	case tokenize.TokenString:
		return value.NewStrVal(token.Value), nil

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
		return nil, verror.NewSyntaxError(verror.ErrIDUnexpectedClosing, [3]string{token.Value, "", ""})

	case tokenize.TokenEOF:
		return nil, verror.NewSyntaxError(verror.ErrIDUnexpectedEOF, [3]string{"", "", ""})

	default:
		return nil, verror.NewSyntaxError(verror.ErrIDInvalidSyntax, [3]string{"unknown token type", "", ""})
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
			errID := verror.ErrIDUnclosedBlock
			if structName == "paren" {
				errID = verror.ErrIDUnclosedParen
			}
			return nil, verror.NewSyntaxError(errID, [3]string{"", "", ""})
		}

		val, err := p.parseValue()
		if err != nil {
			return nil, err
		}
		values = append(values, val)
	}

	errID := verror.ErrIDUnclosedBlock
	if structName == "paren" {
		errID = verror.ErrIDUnclosedParen
	}
	return nil, verror.NewSyntaxError(errID, [3]string{"", "", ""})
}

func (p *Parser) ClassifyLiteral(text string) (core.Value, error) {
	if strings.HasPrefix(text, "#{") && strings.HasSuffix(text, "}") {
		hexStr := text[2 : len(text)-1]
		return p.parseBinary(hexStr)
	}

	if strings.ContainsAny(text, "{}") {
		return nil, verror.NewSyntaxError(verror.ErrIDInvalidLiteral, [3]string{text, "contains braces", ""})
	}

	if strings.HasPrefix(text, "'") {
		return value.NewLitWordVal(text[1:]), nil
	}

	if strings.HasPrefix(text, ":") {
		base := text[1:]
		if len(base) > 0 && base[0] >= '0' && base[0] <= '9' {
			return nil, verror.NewSyntaxError(verror.ErrIDPathLeadingNumber, [3]string{text, "get-word", ""})
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
			return nil, verror.NewSyntaxError(verror.ErrIDPathLeadingNumber, [3]string{text, "set-word", ""})
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
			parts := strings.Split(text, ".")
			if len(parts) == 2 {
				secondPart := parts[1]
				if len(secondPart) > 0 && (secondPart[0] == 'e' || secondPart[0] == 'E') {
					return nil, verror.NewSyntaxError(verror.ErrIDInvalidNumberFormat, [3]string{text, "", ""})
				}
			}
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
	if text == "" || text == "." {
		return nil, verror.NewSyntaxError(verror.ErrIDEmptyPath, [3]string{text, "", ""})
	}

	parts := strings.Split(text, ".")
	segments := make([]value.PathSegment, 0, len(parts))

	for i, part := range parts {
		if part == "" {
			return nil, verror.NewSyntaxError(verror.ErrIDEmptyPathSegment, [3]string{text, "", ""})
		}

		if n, err := strconv.ParseInt(part, 10, 64); err == nil {
			if i == 0 {
				return nil, verror.NewSyntaxError(verror.ErrIDPathLeadingNumber, [3]string{text, "path", ""})
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
	if len(hexStr) == 0 {
		return value.NewBinaryVal([]byte{}), nil
	}

	if len(hexStr)%2 != 0 {
		return nil, verror.NewSyntaxError(verror.ErrIDInvalidBinaryLength, [3]string{hexStr, "", ""})
	}

	bytes := make([]byte, len(hexStr)/2)
	for i := 0; i < len(hexStr); i += 2 {
		high := hexDigitToInt(hexStr[i])
		low := hexDigitToInt(hexStr[i+1])
		if high == -1 || low == -1 {
			return nil, verror.NewSyntaxError(verror.ErrIDInvalidBinaryDigit, [3]string{hexStr, "", ""})
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
