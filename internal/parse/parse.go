// Package parse implements the Viro parser with traditional operator precedence.
//
// Architecture:
// - Tokenize input into tokens (numbers, words, operators, brackets)
// - Parse tokens into Value structures with precedence-aware parsing
// - Traditional precedence: * and / before + and -, not REBOL's left-to-right
//
// Precedence levels per contracts/math.md:
// 1. not (unary, right associative)
// 2. * / (left associative)
// 3. + - (left associative)
// 4. < > <= >= (left associative)
// 5. = <> (left associative)
// 6. and (left associative)
// 7. or (left associative)
// Package parse converts text input into Value structures for evaluation.
//
// The parser implements a two-stage process:
//   1. Tokenization: Text → Tokens (lexical analysis)
//   2. Parsing: Tokens → Values with operator precedence
//
// Operator precedence (7 levels, highest to lowest):
//   1. Function calls
//   2. Unary negation (-)
//   3. Multiplication and division (*, /)
//   4. Addition and subtraction (+, -)
//   5. Comparisons (<, >, <=, >=)
//   6. Equality (=, <>)
//   7. Logic (and, or)
//
// The parser transforms infix notation (3 + 4) into prefix notation ((+ 3 4))
// for evaluation. Parentheses override precedence by forcing immediate evaluation.
//
// Key difference from REBOL: Viro uses mathematical operator precedence,
// while REBOL uses strict left-to-right evaluation with no precedence.
package parse

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"

	"github.com/ericlagergren/decimal"
	"github.com/marcin-radoszewski/viro/internal/value"
	"github.com/marcin-radoszewski/viro/internal/verror"
)

// Parse parses a string into a slice of Values.
// Returns the parsed values and any syntax error encountered.
func Parse(input string) ([]value.Value, *verror.Error) {
	tokens, err := tokenize(input)
	if err != nil {
		return nil, err
	}

	if len(tokens) == 0 {
		return []value.Value{}, nil
	}

	p := &parser{tokens: tokens, pos: 0, source: input}
	return p.parseSequence()
}

// Token represents a lexical token.
type token struct {
	typ tokenType
	val string
	pos int
}

type tokenType int

const (
	tokNumber tokenType = iota
	tokDecimal
	tokString
	tokWord
	tokOperator
	tokSetWord
	tokGetWord
	tokLitWord
	tokLParen
	tokRParen
	tokLBracket
	tokRBracket
	tokEOF
)

func makeSyntaxError(input string, pos int, id string, args [3]string) *verror.Error {
	err := verror.NewSyntaxError(id, args)
	if input != "" {
		err.SetNear(snippetAround(input, pos))
	}
	return err
}

func snippetAround(input string, pos int) string {
	if input == "" {
		return ""
	}
	runes := []rune(input)
	if len(runes) == 0 {
		return ""
	}
	if pos < 0 {
		pos = 0
	}
	if pos >= len(runes) {
		pos = len(runes) - 1
		if pos < 0 {
			pos = 0
		}
	}
	window := 12
	start := pos - window
	if start < 0 {
		start = 0
	}
	end := pos + window + 1
	if end > len(runes) {
		end = len(runes)
	}
	return string(runes[start:end])
}

// tokenize converts input string into tokens.
func tokenize(input string) ([]token, *verror.Error) {
	var tokens []token
	runes := []rune(input)
	pos := 0

	for pos < len(runes) {
		// Skip whitespace
		if unicode.IsSpace(runes[pos]) {
			pos++
			continue
		}

		// String literals
		if runes[pos] == '"' {
			start := pos
			pos++
			var str strings.Builder
			for pos < len(runes) && runes[pos] != '"' {
				str.WriteRune(runes[pos])
				pos++
			}
			if pos >= len(runes) {
				return nil, makeSyntaxError(input, start, verror.ErrIDInvalidSyntax, [3]string{"unclosed string literal", "", ""})
			}
			pos++ // Skip closing quote
			tokens = append(tokens, token{tokString, str.String(), start})
			continue
		}

		// Numbers (integers and decimals)
		// Format: [-]digits[.digits][e|E[+|-]digits]
		// Examples: 42, -3, 19.99, -3.14, 1.5e10, 2.5E-3
		if unicode.IsDigit(runes[pos]) || (runes[pos] == '-' && pos+1 < len(runes) && unicode.IsDigit(runes[pos+1])) {
			start := pos
			hasDecimal := false
			hasExponent := false
			
			// Optional negative sign
			if runes[pos] == '-' {
				pos++
			}
			
			// Integer part (required)
			for pos < len(runes) && unicode.IsDigit(runes[pos]) {
				pos++
			}
			
			// Decimal point and fractional part (optional)
			if pos < len(runes) && runes[pos] == '.' && pos+1 < len(runes) && unicode.IsDigit(runes[pos+1]) {
				hasDecimal = true
				pos++ // Skip '.'
				for pos < len(runes) && unicode.IsDigit(runes[pos]) {
					pos++
				}
			}
			
			// Exponent part (optional)
			if pos < len(runes) && (runes[pos] == 'e' || runes[pos] == 'E') {
				if pos+1 < len(runes) {
					nextPos := pos + 1
					// Optional sign after 'e'
					if runes[nextPos] == '+' || runes[nextPos] == '-' {
						nextPos++
					}
					// Must have at least one digit after 'e' or 'e+'/'e-'
					if nextPos < len(runes) && unicode.IsDigit(runes[nextPos]) {
						hasExponent = true
						pos = nextPos
						for pos < len(runes) && unicode.IsDigit(runes[pos]) {
							pos++
						}
					}
				}
			}
			
			numStr := string(runes[start:pos])
			
			// Determine token type based on format
			if hasDecimal || hasExponent {
				tokens = append(tokens, token{tokDecimal, numStr, start})
			} else {
				tokens = append(tokens, token{tokNumber, numStr, start})
			}
			continue
		}

		// Brackets and parens
		switch runes[pos] {
		case '[':
			tokens = append(tokens, token{tokLBracket, "[", pos})
			pos++
			continue
		case ']':
			tokens = append(tokens, token{tokRBracket, "]", pos})
			pos++
			continue
		case '(':
			tokens = append(tokens, token{tokLParen, "(", pos})
			pos++
			continue
		case ')':
			tokens = append(tokens, token{tokRParen, ")", pos})
			pos++
			continue
		}

		// Refinement words starting with "--"
		if runes[pos] == '-' && pos+1 < len(runes) && runes[pos+1] == '-' && pos+2 < len(runes) && isWordStart(runes[pos+2]) {
			start := pos
			pos += 2 // Skip leading --
			for pos < len(runes) && isWordChar(runes[pos]) {
				pos++
			}
			word := string(runes[start:pos])
			tokens = append(tokens, token{tokWord, word, start})
			continue
		}

		// Single-character operators and multi-character operator starts
		if runes[pos] == '+' || runes[pos] == '*' || runes[pos] == '/' {
			tokens = append(tokens, token{tokOperator, string(runes[pos]), pos})
			pos++
			continue
		}

		// Handle <, >, =, and their multi-character variants (<=, >=, <>)
		if runes[pos] == '<' {
			start := pos
			pos++
			// Check for <= or <>
			if pos < len(runes) && (runes[pos] == '=' || runes[pos] == '>') {
				pos++
				tokens = append(tokens, token{tokOperator, string(runes[start:pos]), start})
			} else {
				tokens = append(tokens, token{tokOperator, "<", start})
			}
			continue
		}

		if runes[pos] == '>' {
			start := pos
			pos++
			// Check for >=
			if pos < len(runes) && runes[pos] == '=' {
				pos++
				tokens = append(tokens, token{tokOperator, ">=", start})
			} else {
				tokens = append(tokens, token{tokOperator, ">", start})
			}
			continue
		}

		if runes[pos] == '=' {
			tokens = append(tokens, token{tokOperator, "=", pos})
			pos++
			continue
		}

		// Handle minus sign (could be negative number or operator)
		// Negative numbers are already handled above, so here it's an operator
		if runes[pos] == '-' && (pos+1 >= len(runes) || !unicode.IsDigit(runes[pos+1])) {
			tokens = append(tokens, token{tokOperator, "-", pos})
			pos++
			continue
		}

		// Words (including set-word, get-word, lit-word)
		if isWordStart(runes[pos]) {
			start := pos
			pos++
			for pos < len(runes) && isWordChar(runes[pos]) {
				pos++
			}
			word := string(runes[start:pos])

			// Check for set-word (word:)
			if pos < len(runes) && runes[pos] == ':' {
				pos++
				tokens = append(tokens, token{tokSetWord, word, start})
				continue
			}

			// Check if it's an operator
			if isOperator(word) {
				tokens = append(tokens, token{tokOperator, word, start})
			} else {
				tokens = append(tokens, token{tokWord, word, start})
			}
			continue
		}

		// Get-word (:word)
		if runes[pos] == ':' && pos+1 < len(runes) && isWordStart(runes[pos+1]) {
			start := pos
			pos++ // Skip :
			wordStart := pos
			for pos < len(runes) && isWordChar(runes[pos]) {
				pos++
			}
			tokens = append(tokens, token{tokGetWord, string(runes[wordStart:pos]), start})
			continue
		}

		// Lit-word ('word)
		if runes[pos] == '\'' && pos+1 < len(runes) && isWordStart(runes[pos+1]) {
			start := pos
			pos++ // Skip '
			wordStart := pos
			for pos < len(runes) && isWordChar(runes[pos]) {
				pos++
			}
			tokens = append(tokens, token{tokLitWord, string(runes[wordStart:pos]), start})
			continue
		}

		// Unknown character
		return nil, makeSyntaxError(input, pos, verror.ErrIDInvalidSyntax, [3]string{fmt.Sprintf("unexpected character %q", runes[pos]), "", ""})
	}

	return tokens, nil
}

func isWordStart(r rune) bool {
	return unicode.IsLetter(r) || r == '_' || r == '?' || r == '!'
}

func isWordChar(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' || r == '_' || r == '?' || r == '!'
}

func isOperator(word string) bool {
	operators := map[string]bool{
		"+": true, "-": true, "*": true, "/": true,
		"<": true, ">": true, "<=": true, ">=": true,
		"=": true, "<>": true,
		"and": true, "or": true, "not": true,
	}
	return operators[word]
}

// parser holds parsing state.
type parser struct {
	tokens []token
	pos    int
	source string
}

func (p *parser) syntaxError(pos int, id string, args [3]string) *verror.Error {
	return makeSyntaxError(p.source, pos, id, args)
}

// parseSequence parses a sequence of values (top level or within block/paren).
func (p *parser) parseSequence() ([]value.Value, *verror.Error) {
	var values []value.Value

	for !p.isAtEnd() && p.peek().typ != tokRBracket && p.peek().typ != tokRParen {
		val, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		values = append(values, val)
	}

	return values, nil
}

// parseExpression parses an expression with operator precedence.
// For now, we'll use a simple precedence climbing parser.
func (p *parser) parseExpression() (value.Value, *verror.Error) {
	return p.parsePrecedence(7) // Start with lowest precedence (or)
}

// parsePrecedence implements precedence climbing algorithm.
// Higher level number = lower precedence (7 is lowest = or)
func (p *parser) parsePrecedence(level int) (value.Value, *verror.Error) {
	if level == 0 {
		return p.parsePrimary()
	}

	left, err := p.parsePrecedence(level - 1)
	if err != nil {
		return value.NoneVal(), err
	}

	for !p.isAtEnd() && p.isOperatorAtLevel(level) {
		op := p.advance()
		right, err := p.parsePrecedence(level - 1)
		if err != nil {
			return value.NoneVal(), err
		}

		// Build function call: operator word, then arguments
		left = value.ParenVal([]value.Value{
			value.WordVal(op.val),
			left,
			right,
		})
	}

	return left, nil
}

// isOperatorAtLevel checks if current token is an operator at the given precedence level.
func (p *parser) isOperatorAtLevel(level int) bool {
	if p.isAtEnd() || p.peek().typ != tokOperator {
		return false
	}

	op := p.peek().val
	return getOperatorLevel(op) == level
}

// getOperatorLevel returns precedence level for operator (higher = tighter binding).
// Inverted from contracts: level 1 in contract = level 7 here (we parse low to high).
func getOperatorLevel(op string) int {
	switch op {
	case "or":
		return 7 // Lowest precedence
	case "and":
		return 6
	case "=", "<>":
		return 5
	case "<", ">", "<=", ">=":
		return 4
	case "+", "-":
		return 3
	case "*", "/":
		return 2 // Highest precedence (binary)
	case "not":
		return 1 // Unary, handled separately
	default:
		return 0
	}
}

// parsePrimary parses a primary expression (literal, word, block, paren, etc.).
func (p *parser) parsePrimary() (value.Value, *verror.Error) {
	if p.isAtEnd() {
		return value.NoneVal(), p.syntaxError(len([]rune(p.source)), verror.ErrIDUnexpectedEOF, [3]string{"", "", ""})
	}

	tok := p.advance()

	switch tok.typ {
	case tokNumber:
		num, err := strconv.ParseInt(tok.val, 10, 64)
		if err != nil {
			return value.NoneVal(), p.syntaxError(tok.pos, verror.ErrIDInvalidLiteral, [3]string{tok.val, "", ""})
		}
		return value.IntVal(num), nil

	case tokDecimal:
		// Parse decimal literal: creates DecimalValue directly
		// Format: [-]digits[.digits][e|E[+|-]digits]
		d := new(decimal.Big)
		_, ok := d.SetString(tok.val)
		if !ok {
			return value.NoneVal(), p.syntaxError(tok.pos, verror.ErrIDInvalidLiteral, [3]string{tok.val, "", ""})
		}
		
		// Calculate scale from decimal string representation
		scale := int16(0)
		if idx := strings.Index(tok.val, "."); idx >= 0 {
			// Find end of fractional part (before 'e' or 'E' if present)
			endIdx := len(tok.val)
			if eIdx := strings.IndexAny(tok.val, "eE"); eIdx > idx {
				endIdx = eIdx
			}
			scale = int16(endIdx - idx - 1)
		}
		
		return value.DecimalVal(d, scale), nil

	case tokString:
		return value.StrVal(tok.val), nil

	case tokWord:
		// Handle special keyword literals
		switch tok.val {
		case "true":
			return value.LogicVal(true), nil
		case "false":
			return value.LogicVal(false), nil
		case "none":
			return value.NoneVal(), nil
		default:
			return value.WordVal(tok.val), nil
		}

	case tokOperator:
		// When an operator appears in primary position (like in `(+ 3 4)`),
		// treat it as a word value, not as an infix operator
		return value.WordVal(tok.val), nil

	case tokSetWord:
		return value.SetWordVal(tok.val), nil

	case tokGetWord:
		return value.GetWordVal(tok.val), nil

	case tokLitWord:
		return value.LitWordVal(tok.val), nil

	case tokLBracket:
		// Parse block contents
		elements, err := p.parseSequence()
		if err != nil {
			return value.NoneVal(), err
		}
		if p.isAtEnd() || p.peek().typ != tokRBracket {
			return value.NoneVal(), p.syntaxError(tok.pos, verror.ErrIDUnclosedBlock, [3]string{"[", "", ""})
		}
		p.advance() // Consume ]
		return value.BlockVal(elements), nil

	case tokLParen:
		// Parse paren contents
		elements, err := p.parseSequence()
		if err != nil {
			return value.NoneVal(), err
		}
		if p.isAtEnd() || p.peek().typ != tokRParen {
			return value.NoneVal(), p.syntaxError(tok.pos, verror.ErrIDUnclosedParen, [3]string{"(", "", ""})
		}
		p.advance() // Consume )
		return value.ParenVal(elements), nil

	default:
		return value.NoneVal(), p.syntaxError(tok.pos, verror.ErrIDInvalidSyntax, [3]string{tok.val, "", ""})
	}
}

func (p *parser) peek() token {
	if p.isAtEnd() {
		return token{tokEOF, "", len(p.tokens)}
	}
	return p.tokens[p.pos]
}

func (p *parser) advance() token {
	if !p.isAtEnd() {
		tok := p.tokens[p.pos]
		p.pos++
		return tok
	}
	return token{tokEOF, "", len(p.tokens)}
}

func (p *parser) isAtEnd() bool {
	return p.pos >= len(p.tokens)
}

// ParseEval is a convenience function that parses and returns a single expression.
// Used by REPL for single-line evaluation.
func ParseEval(input string) ([]value.Value, *verror.Error) {
	return Parse(input)
}

// Format formats a value back to string (for debugging/display).
func Format(val value.Value) string {
	switch val.Type {
	case value.TypeNone:
		return "none"
	case value.TypeLogic:
		if logic, ok := val.AsLogic(); ok {
			if logic {
				return "true"
			}
			return "false"
		}
		return "logic"
	case value.TypeInteger:
		if num, ok := val.AsInteger(); ok {
			return fmt.Sprintf("%d", num)
		}
		return "integer"
	case value.TypeString:
		if str, ok := val.AsString(); ok {
			return fmt.Sprintf("\"%s\"", str.String())
		}
		return "string"
	case value.TypeWord:
		if word, ok := val.AsWord(); ok {
			return word
		}
		return "word"
	case value.TypeSetWord:
		if word, ok := val.AsWord(); ok {
			return word + ":"
		}
		return "set-word"
	case value.TypeGetWord:
		if word, ok := val.AsWord(); ok {
			return ":" + word
		}
		return "get-word"
	case value.TypeLitWord:
		if word, ok := val.AsWord(); ok {
			return "'" + word
		}
		return "lit-word"
	case value.TypeBlock:
		if block, ok := val.AsBlock(); ok {
			var parts []string
			for _, elem := range block.Elements {
				parts = append(parts, Format(elem))
			}
			return "[" + strings.Join(parts, " ") + "]"
		}
		return "block"
	case value.TypeParen:
		if block, ok := val.AsBlock(); ok {
			var parts []string
			for _, elem := range block.Elements {
				parts = append(parts, Format(elem))
			}
			return "(" + strings.Join(parts, " ") + ")"
		}
		return "paren"
	case value.TypeFunction:
		return "function"
	default:
		return "unknown"
	}
}
