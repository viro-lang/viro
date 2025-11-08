package dialect

import (
	"fmt"

	"github.com/marcin-radoszewski/viro/internal/core"
	"github.com/marcin-radoszewski/viro/internal/value"
	"github.com/marcin-radoszewski/viro/internal/verror"
)

// Engine is the main parse dialect engine.
type Engine struct {
	cursor    SeriesCursor
	state     *ParseState
	evaluator core.Evaluator
}

// NewEngine creates a new parse engine.
func NewEngine(input core.Value, rules []core.Value, options ParseOptions, eval core.Evaluator) (*Engine, error) {
	cursor, ok := NewCursor(input)
	if !ok {
		return nil, verror.NewScriptError("parse-invalid-input", [3]string{"parse", "input must be string! or block!", value.TypeToString(input.GetType())})
	}

	state := NewParseState(input, options)
	state.position = 0

	return &Engine{
		cursor:    cursor,
		state:     state,
		evaluator: eval,
	}, nil
}

// Parse executes the parse operation.
func (e *Engine) Parse(rules []core.Value) (bool, error) {
	// Handle --part option
	limit := e.cursor.Length()
	if e.state.options.Part > 0 && e.state.options.Part < limit {
		limit = e.state.options.Part
	}

	// Try to match rules
	success, err := e.matchRules(rules, 0)
	if err != nil {
		return false, err
	}

	// Check --all option (must consume entire input)
	if e.state.options.MatchAll && success {
		if e.state.position < limit {
			return false, nil
		}
	}

	return success, nil
}

// matchRules attempts to match a sequence of rules.
func (e *Engine) matchRules(rules []core.Value, startPos int) (bool, error) {
	pos := startPos
	ruleIdx := 0

	for ruleIdx < len(rules) {
		rule := rules[ruleIdx]

		// Match the rule
		matched, newPos, err := e.matchRule(rule, pos)
		if err != nil {
			return false, err
		}

		if !matched {
			return false, nil
		}

		pos = newPos
		ruleIdx++
	}

	e.state.position = pos
	return true, nil
}

// matchRule attempts to match a single rule.
func (e *Engine) matchRule(rule core.Value, pos int) (bool, int, error) {
	// Handle different rule types
	switch rule.GetType() {
	case value.TypeString:
		// String literal - match against input
		return e.matchStringLiteral(rule, pos)

	case value.TypeInteger:
		// Integer - match exact count
		return e.matchInteger(rule, pos)

	case value.TypeBlock:
		// Block - alternation or subrule
		return e.matchBlock(rule, pos)

	case value.TypeWord:
		// Word - special commands or datatype matches
		return e.matchWord(rule, pos)

	case value.TypeBitset:
		// Bitset - character class match
		return e.matchBitset(rule, pos)

	default:
		// Unknown rule type
		return false, pos, verror.NewScriptError("parse-invalid-rule", [3]string{"parse", "invalid rule type", value.TypeToString(rule.GetType())})
	}
}

// matchStringLiteral matches a string literal against the input.
func (e *Engine) matchStringLiteral(rule core.Value, pos int) (bool, int, error) {
	strVal, ok := value.AsStringValue(rule)
	if !ok {
		return false, pos, nil
	}

	pattern := strVal.String()
	if e.cursor.IsString() {
		// String input: match character by character
		sc := e.cursor.(*StringCursor)
		for i, r := range []rune(pattern) {
			if pos+i >= sc.Length() {
				return false, pos, nil
			}
			inputRune, _ := sc.RuneAt(pos + i)
			if !e.matchChar(r, inputRune) {
				return false, pos, nil
			}
		}
		return true, pos + len([]rune(pattern)), nil
	} else {
		// Block input: match as a literal value
		elem, ok := e.cursor.At(pos)
		if !ok {
			return false, pos, nil
		}
		if elem.GetType() == value.TypeString {
			if elemStr, ok := value.AsStringValue(elem); ok {
				if MatchString(elemStr.String(), pattern, e.state.options.CaseSensitive) {
					return true, pos + 1, nil
				}
			}
		}
		return false, pos, nil
	}
}

// matchChar matches two characters with case sensitivity option.
func (e *Engine) matchChar(expected, actual rune) bool {
	if e.state.options.CaseSensitive {
		return expected == actual
	}
	// Simple case-insensitive comparison
	if expected >= 'A' && expected <= 'Z' {
		expected = expected - 'A' + 'a'
	}
	if actual >= 'A' && actual <= 'Z' {
		actual = actual - 'A' + 'a'
	}
	return expected == actual
}

// matchInteger matches an exact count of elements.
func (e *Engine) matchInteger(rule core.Value, pos int) (bool, int, error) {
	count, ok := value.AsIntValue(rule)
	if !ok {
		return false, pos, nil
	}

	// Skip exactly count elements
	if pos+int(count) > e.cursor.Length() {
		return false, pos, nil
	}
	return true, pos + int(count), nil
}

// matchBlock matches a block rule (subrules or alternation).
func (e *Engine) matchBlock(rule core.Value, pos int) (bool, int, error) {
	blockVal, ok := value.AsBlockValue(rule)
	if !ok {
		return false, pos, nil
	}

	// Check if it's an alternation (contains | word)
	hasAlternation := false
	for _, elem := range blockVal.Elements {
		if word, ok := value.AsWordValue(elem); ok && word == "|" {
			hasAlternation = true
			break
		}
	}

	if hasAlternation {
		return e.matchAlternation(blockVal.Elements, pos)
	}

	// Otherwise, match as a sequence of subrules
	return e.matchSequence(blockVal.Elements, pos)
}

// matchSequence matches a sequence of rules.
func (e *Engine) matchSequence(rules []core.Value, startPos int) (bool, int, error) {
	pos := startPos
	for _, rule := range rules {
		matched, newPos, err := e.matchRule(rule, pos)
		if err != nil {
			return false, pos, err
		}
		if !matched {
			return false, startPos, nil
		}
		pos = newPos
	}
	return true, pos, nil
}

// matchAlternation matches one of several alternatives separated by |.
func (e *Engine) matchAlternation(rules []core.Value, pos int) (bool, int, error) {
	alternatives := splitAlternatives(rules)

	for _, alt := range alternatives {
		matched, newPos, err := e.matchSequence(alt, pos)
		if err != nil {
			return false, pos, err
		}
		if matched {
			return true, newPos, nil
		}
	}

	return false, pos, nil
}

// splitAlternatives splits rules by | separator.
func splitAlternatives(rules []core.Value) [][]core.Value {
	var result [][]core.Value
	var current []core.Value

	for _, rule := range rules {
		if word, ok := value.AsWordValue(rule); ok && word == "|" {
			if len(current) > 0 {
				result = append(result, current)
				current = []core.Value{}
			}
		} else {
			current = append(current, rule)
		}
	}

	if len(current) > 0 {
		result = append(result, current)
	}

	return result
}

// matchWord matches a word rule (datatype, keyword, etc.).
func (e *Engine) matchWord(rule core.Value, pos int) (bool, int, error) {
	word, ok := value.AsWordValue(rule)
	if !ok {
		return false, pos, nil
	}

	// Handle special keywords
	switch word {
	case "skip":
		// Skip one element
		if pos >= e.cursor.Length() {
			return false, pos, nil
		}
		return true, pos + 1, nil

	case "end":
		// Match end of input
		if pos >= e.cursor.Length() {
			return true, pos, nil
		}
		return false, pos, nil

	default:
		// Try datatype matching
		return e.matchDatatype(word, pos)
	}
}

// matchDatatype matches an element by its datatype.
func (e *Engine) matchDatatype(typename string, pos int) (bool, int, error) {
	elem, ok := e.cursor.At(pos)
	if !ok {
		return false, pos, nil
	}

	// Match datatype name
	actualType := value.TypeToString(elem.GetType())
	if actualType == typename {
		return true, pos + 1, nil
	}

	return false, pos, nil
}

// matchBitset matches a character against a bitset.
func (e *Engine) matchBitset(rule core.Value, pos int) (bool, int, error) {
	bs, ok := value.AsBitsetValue(rule)
	if !ok {
		return false, pos, nil
	}

	if !e.cursor.IsString() {
		return false, pos, verror.NewScriptError("parse-invalid-rule", [3]string{"parse", "bitset can only be used with string input", ""})
	}

	sc := e.cursor.(*StringCursor)
	r, ok := sc.RuneAt(pos)
	if !ok {
		return false, pos, nil
	}

	if bs.Test(r) {
		return true, pos + 1, nil
	}

	return false, pos, nil
}

// GetPosition returns the current parse position.
func (e *Engine) GetPosition() int {
	return e.state.position
}

// GetCaptures returns all captured values.
func (e *Engine) GetCaptures() map[string]core.Value {
	return e.state.captures
}

// Parse is the main entry point for the parse dialect.
func Parse(input core.Value, rules []core.Value, options ParseOptions, eval core.Evaluator) (bool, error) {
	engine, err := NewEngine(input, rules, options, eval)
	if err != nil {
		return false, err
	}

	return engine.Parse(rules)
}

// ParseWithDiagnostics parses and returns diagnostic information.
func ParseWithDiagnostics(input core.Value, rules []core.Value, options ParseOptions, eval core.Evaluator) (bool, int, map[string]core.Value, error) {
	engine, err := NewEngine(input, rules, options, eval)
	if err != nil {
		return false, 0, nil, err
	}

	success, err := engine.Parse(rules)
	return success, engine.GetPosition(), engine.GetCaptures(), err
}

// Helper for error messages with near/where context
func (e *Engine) errorWithContext(id string, message string) error {
	pos := e.state.position
	near := ""
	if pos < e.cursor.Length() {
		near = e.cursor.StringAt(pos)
	}
	return verror.NewScriptError(id, [3]string{"parse", message, fmt.Sprintf("at position %d near %s", pos, near)})
}
