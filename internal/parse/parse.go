package parse

import (
	"github.com/marcin-radoszewski/viro/internal/core"
	"github.com/marcin-radoszewski/viro/internal/tokenize"
	"github.com/marcin-radoszewski/viro/internal/verror"
)

func Parse(input string) ([]core.Value, error) {
	tokenizer := tokenize.NewTokenizer(input)
	tokens, err := tokenizer.Tokenize()
	if err != nil {
		if vErr, ok := err.(*verror.Error); ok {
			if input != "" {
				vErr.SetNear(input)
			}
			return nil, vErr
		}
		vErr := verror.NewSyntaxError(verror.ErrIDInvalidSyntax, [3]string{err.Error(), "", ""})
		if input != "" {
			vErr.SetNear(input)
		}
		return nil, vErr
	}

	parser := NewParser(tokens)
	values, err := parser.Parse()
	if err != nil {
		if vErr, ok := err.(*verror.Error); ok {
			if input != "" {
				vErr.SetNear(input)
			}
			return nil, vErr
		}
		vErr := verror.NewSyntaxError(verror.ErrIDInvalidSyntax, [3]string{err.Error(), "", ""})
		if input != "" {
			vErr.SetNear(input)
		}
		return nil, vErr
	}

	return values, nil
}
