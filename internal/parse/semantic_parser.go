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

var (
	intPattern        = regexp.MustCompile(`^-?[0-9]+$`)
	decimalPattern    = regexp.MustCompile(`^-?[0-9]+\.[0-9]+([eE][+-]?[0-9]+)?$`)
	scientificPattern = regexp.MustCompile(`^-?[0-9]+[eE][+-]?[0-9]+$`)
)

type Parser struct {
	tokens []tokenize.Token
	pos    int
	source string
}

func NewParser(tokens []tokenize.Token, source string) *Parser {
	return &Parser{
		tokens: tokens,
		pos:    0,
		source: source,
	}
}

func (p *Parser) syntaxError(id string, args [3]string, line, column int) *verror.Error {
	return verror.NewSyntaxError(id, args).SetLocation(p.source, line, column)
}

func (p *Parser) eofLocation() (int, int) {
	if len(p.tokens) == 0 {
		return 1, 1
	}
	last := p.tokens[len(p.tokens)-1]
	if last.Line == 0 && last.Column == 0 {
		return 1, 1
	}
	return last.Line, last.Column
}

func (p *Parser) locationFromToken(token tokenize.Token) core.SourceLocation {
	line := token.Line
	if line == 0 {
		line = 1
	}
	column := token.Column
	if column == 0 {
		column = 1
	}
	return core.SourceLocation{
		File:   p.source,
		Line:   line,
		Column: column,
	}
}

func (p *Parser) Parse() ([]core.Value, []core.SourceLocation, error) {
	values := []core.Value{}
	locations := []core.SourceLocation{}

	for p.pos < len(p.tokens) {
		token := p.tokens[p.pos]

		if token.Type == tokenize.TokenEOF {
			break
		}

		val, loc, err := p.parseValue()
		if err != nil {
			return nil, nil, err
		}
		values = append(values, val)
		locations = append(locations, loc)
	}

	return values, locations, nil
}

func (p *Parser) parseValue() (core.Value, core.SourceLocation, error) {
	if p.pos >= len(p.tokens) {
		line, column := p.eofLocation()
		return nil, core.SourceLocation{}, p.syntaxError(verror.ErrIDUnexpectedEOF, [3]string{"", "", ""}, line, column)
	}

	token := p.tokens[p.pos]
	p.pos++

	loc := p.locationFromToken(token)

	switch token.Type {
	case tokenize.TokenLiteral:
		val, err := p.ClassifyLiteral(token)
		return val, loc, err

	case tokenize.TokenString:
		return value.NewStrVal(token.Value), loc, nil

	case tokenize.TokenLBracket:
		values, locations, err := p.parseUntil(tokenize.TokenRBracket, "block", token)
		if err != nil {
			return nil, core.SourceLocation{}, err
		}
		block := value.NewBlockVal(values)
		if blockValue, ok := value.AsBlockValue(block); ok {
			blockValue.SetLocations(locations)
		}
		return block, loc, nil

	case tokenize.TokenLParen:
		values, locations, err := p.parseUntil(tokenize.TokenRParen, "paren", token)
		if err != nil {
			return nil, core.SourceLocation{}, err
		}
		paren := value.NewParenVal(values)
		if blockValue, ok := value.AsBlockValue(paren); ok {
			blockValue.SetLocations(locations)
		}
		return paren, loc, nil

	case tokenize.TokenRBracket, tokenize.TokenRParen:
		return nil, core.SourceLocation{}, p.syntaxError(verror.ErrIDUnexpectedClosing, [3]string{token.Value, "", ""}, token.Line, token.Column)

	case tokenize.TokenEOF:
		return nil, core.SourceLocation{}, p.syntaxError(verror.ErrIDUnexpectedEOF, [3]string{"", "", ""}, token.Line, token.Column)

	default:
		return nil, core.SourceLocation{}, p.syntaxError(verror.ErrIDInvalidSyntax, [3]string{"unknown token type", "", ""}, token.Line, token.Column)
	}
}

func (p *Parser) parseUntil(closingType tokenize.TokenType, structName string, start tokenize.Token) ([]core.Value, []core.SourceLocation, error) {
	values := []core.Value{}
	locations := []core.SourceLocation{}

	for p.pos < len(p.tokens) {
		token := p.tokens[p.pos]

		if token.Type == closingType {
			p.pos++
			return values, locations, nil
		}

		if token.Type == tokenize.TokenEOF {
			errID := verror.ErrIDUnclosedBlock
			if structName == "paren" {
				errID = verror.ErrIDUnclosedParen
			}
			return nil, nil, p.syntaxError(errID, [3]string{"", "", ""}, start.Line, start.Column)
		}

		val, loc, err := p.parseValue()
		if err != nil {
			return nil, nil, err
		}
		values = append(values, val)
		locations = append(locations, loc)
	}

	errID := verror.ErrIDUnclosedBlock
	if structName == "paren" {
		errID = verror.ErrIDUnclosedParen
	}
	return nil, nil, p.syntaxError(errID, [3]string{"", "", ""}, start.Line, start.Column)
}

func (p *Parser) ClassifyLiteral(token tokenize.Token) (core.Value, error) {
	text := token.Value
	if strings.HasPrefix(text, "#{") && strings.HasSuffix(text, "}") {
		hexStr := text[2 : len(text)-1]
		return p.parseBinary(token, hexStr)
	}

	if strings.ContainsAny(text, "{}") {
		return nil, p.syntaxError(verror.ErrIDInvalidLiteral, [3]string{text, "contains braces", ""}, token.Line, token.Column)
	}

	if strings.HasPrefix(text, "'") {
		return value.NewLitWordVal(text[1:]), nil
	}

	if strings.HasPrefix(text, ":") {
		base := text[1:]
		if len(base) > 0 && base[0] >= '0' && base[0] <= '9' {
			return nil, p.syntaxError(verror.ErrIDPathLeadingNumber, [3]string{text, "get-word", ""}, token.Line, token.Column)
		}
		if strings.Contains(base, ".") {
			segments, err := p.parsePath(token, base)
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
			return nil, p.syntaxError(verror.ErrIDPathLeadingNumber, [3]string{text, "set-word", ""}, token.Line, token.Column)
		}
		if strings.Contains(base, ".") {
			segments, err := p.parsePath(token, base)
			if err != nil {
				return nil, err
			}
			return value.SetPathVal(value.NewSetPath(segments, value.NewNoneVal())), nil
		}
		return value.NewSetWordVal(base), nil
	}

	if intPattern.MatchString(text) {
		n, _ := strconv.ParseInt(text, 10, 64)
		return value.NewIntVal(n), nil
	}

	if decimalPattern.MatchString(text) || scientificPattern.MatchString(text) {
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
					return nil, p.syntaxError(verror.ErrIDInvalidNumberFormat, [3]string{text, "", ""}, token.Line, token.Column)
				}
			}
		}

		segments, err := p.parsePath(token, text)
		if err != nil {
			return nil, err
		}
		return value.PathVal(value.NewPath(segments, value.NewNoneVal())), nil
	}

	return value.NewWordVal(text), nil
}

func (p *Parser) parsePath(token tokenize.Token, text string) ([]value.PathSegment, error) {
	if text == "" || text == "." {
		return nil, p.syntaxError(verror.ErrIDEmptyPath, [3]string{text, "", ""}, token.Line, token.Column)
	}

	segments := []value.PathSegment{}
	start := 0
	depth := 0

	for i := 0; i < len(text); i++ {
		ch := text[i]

		if ch == '(' {
			depth++
		} else if ch == ')' {
			if depth > 0 {
				depth--
			}
		}

		if ch == '.' && depth == 0 {
			part := text[start:i]
			if part == "" {
				return nil, p.syntaxError(verror.ErrIDEmptyPathSegment, [3]string{text, "", ""}, token.Line, token.Column)
			}
			seg, err := p.parsePathSegment(token, text, part)
			if err != nil {
				return nil, err
			}
			if len(segments) == 0 && seg.Type == value.PathSegmentIndex {
				return nil, p.syntaxError(verror.ErrIDPathLeadingNumber, [3]string{text, "path", ""}, token.Line, token.Column)
			}
			segments = append(segments, seg)
			start = i + 1
		}
	}

	if depth != 0 {
		return nil, p.syntaxError(verror.ErrIDInvalidPath, [3]string{text, "unbalanced parentheses", ""}, token.Line, token.Column)
	}

	last := text[start:]
	if last == "" {
		return nil, p.syntaxError(verror.ErrIDEmptyPathSegment, [3]string{text, "", ""}, token.Line, token.Column)
	}

	seg, err := p.parsePathSegment(token, text, last)
	if err != nil {
		return nil, err
	}
	if len(segments) == 0 && seg.Type == value.PathSegmentIndex {
		return nil, p.syntaxError(verror.ErrIDPathLeadingNumber, [3]string{text, "path", ""}, token.Line, token.Column)
	}
	segments = append(segments, seg)

	return segments, nil
}

func (p *Parser) parsePathSegment(token tokenize.Token, fullText, part string) (value.PathSegment, error) {
	if len(part) >= 2 && part[0] == '(' && part[len(part)-1] == ')' {
		block, err := p.parseEvalPathSegment(token, part[1:len(part)-1])
		if err != nil {
			return value.PathSegment{}, err
		}
		return value.PathSegment{Type: value.PathSegmentEval, Value: block}, nil
	}

	if n, err := strconv.ParseInt(part, 10, 64); err == nil {
		return value.PathSegment{Type: value.PathSegmentIndex, Value: n}, nil
	}

	return value.PathSegment{Type: value.PathSegmentWord, Value: part}, nil
}

func (p *Parser) parseEvalPathSegment(token tokenize.Token, inner string) (*value.BlockValue, error) {
	tokenizer := tokenize.NewTokenizer(inner)
	tokenizer.SetSource(token.Source)
	tokens, err := tokenizer.Tokenize()
	if err != nil {
		return nil, err
	}

	parser := NewParser(tokens, token.Source)
	values, locations, err := parser.Parse()
	if err != nil {
		return nil, err
	}
	if len(values) == 0 {
		return nil, p.syntaxError(verror.ErrIDEmptyPathSegment, [3]string{token.Value, "eval", ""}, token.Line, token.Column)
	}

	block := value.NewBlockValue(values)
	block.SetLocations(locations)
	return block, nil
}

func (p *Parser) parseBinary(token tokenize.Token, hexStr string) (core.Value, error) {
	if len(hexStr) == 0 {
		return value.NewBinaryVal([]byte{}), nil
	}

	if len(hexStr)%2 != 0 {
		return nil, p.syntaxError(verror.ErrIDInvalidBinaryLength, [3]string{hexStr, "", ""}, token.Line, token.Column)
	}

	bytes := make([]byte, len(hexStr)/2)
	for i := 0; i < len(hexStr); i += 2 {
		high := hexDigitToInt(hexStr[i])
		low := hexDigitToInt(hexStr[i+1])
		if high == -1 || low == -1 {
			return nil, p.syntaxError(verror.ErrIDInvalidBinaryDigit, [3]string{hexStr, "", ""}, token.Line, token.Column)
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
