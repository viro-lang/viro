package parse

import (
	"strings"

	"github.com/marcin-radoszewski/viro/internal/core"
	"github.com/marcin-radoszewski/viro/internal/parse/peg"
	"github.com/marcin-radoszewski/viro/internal/verror"
)

func Parse(input string) ([]core.Value, error) {
	result, err := peg.ParseReader("", strings.NewReader(input))
	if err != nil {
		errMsg := err.Error()
		errID := verror.ErrIDInvalidSyntax

		if strings.Contains(errMsg, "expected:") && strings.Contains(errMsg, "\"]\"") {
			errID = verror.ErrIDUnclosedBlock
		} else if strings.Contains(errMsg, "expected:") && strings.Contains(errMsg, "\")\"") {
			errID = verror.ErrIDUnclosedParen
		}

		vErr := verror.NewSyntaxError(errID, [3]string{errMsg, "", ""})
		if input != "" {
			vErr.SetNear(input)
		}
		return nil, vErr
	}
	return result.([]core.Value), nil
}
