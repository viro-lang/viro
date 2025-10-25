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
		vErr := verror.NewSyntaxError(verror.ErrIDInvalidSyntax, [3]string{err.Error(), "", ""})
		if input != "" {
			vErr.SetNear(input)
		}
		return nil, vErr
	}
	return result.([]core.Value), nil
}
