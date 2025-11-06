package parse

import (
	"strings"

	"github.com/marcin-radoszewski/viro/internal/core"
	"github.com/marcin-radoszewski/viro/internal/tokenize"
	"github.com/marcin-radoszewski/viro/internal/verror"
)

func Parse(input string) ([]core.Value, error) {
	tokenizer := tokenize.NewTokenizer(input)
	tokens, err := tokenizer.Tokenize()
	if err != nil {
		vErr := verror.NewSyntaxError(verror.ErrIDInvalidSyntax, [3]string{err.Error(), "", ""})
		if input != "" {
			vErr.SetNear(input)
		}
		return nil, vErr
	}

	parser := NewParser(tokens)
	values, err := parser.Parse()
	if err != nil {
		errID := verror.ErrIDInvalidSyntax
		errMsg := err.Error()

		if strings.Contains(errMsg, "unclosed block") {
			errID = verror.ErrIDUnclosedBlock
		} else if strings.Contains(errMsg, "unclosed paren") {
			errID = verror.ErrIDUnclosedParen
		}

		vErr := verror.NewSyntaxError(errID, [3]string{errMsg, "", ""})
		if input != "" {
			vErr.SetNear(input)
		}
		return nil, vErr
	}

	return values, nil
}
