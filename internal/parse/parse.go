package parse

import (
	"github.com/marcin-radoszewski/viro/internal/core"
	"github.com/marcin-radoszewski/viro/internal/tokenize"
	"github.com/marcin-radoszewski/viro/internal/verror"
)

func Parse(input string) ([]core.Value, []core.SourceLocation, error) {
	return ParseWithSource(input, "")
}

func ParseWithSource(input, source string) ([]core.Value, []core.SourceLocation, error) {
	tokenizer := tokenize.NewTokenizer(input)
	tokenizer.SetSource(source)
	tokens, err := tokenizer.Tokenize()
	if err != nil {
		return nil, nil, enrichParseError(err, input, source)
	}

	parser := NewParser(tokens, source)
	values, locations, err := parser.Parse()
	if err != nil {
		return nil, nil, enrichParseError(err, input, source)
	}

	return values, locations, nil
}

func enrichParseError(err error, input, source string) error {
	if vErr, ok := err.(*verror.Error); ok {
		if input != "" {
			vErr.SetNear(input)
		}
		if vErr.File == "" && source != "" {
			line := vErr.Line
			column := vErr.Column
			if line == 0 {
				line = 1
			}
			if column == 0 {
				column = 1
			}
			vErr.SetLocation(source, line, column)
		}
		return vErr
	}

	vErr := verror.NewSyntaxError(verror.ErrIDInvalidSyntax, [3]string{err.Error(), "", ""})
	if input != "" {
		vErr.SetNear(input)
	}
	if source != "" {
		vErr.SetLocation(source, 1, 1)
	}
	return vErr
}
