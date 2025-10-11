package contract

import (
	"github.com/marcin-radoszewski/viro/internal/eval"
	"github.com/marcin-radoszewski/viro/internal/parse"
	"github.com/marcin-radoszewski/viro/internal/value"
	"github.com/marcin-radoszewski/viro/internal/verror"
)

// Evaluate is a helper function to evaluate Viro code in tests.
func Evaluate(src string) (value.Value, *verror.Error) {
	vals, err := parse.Parse(src)
	if err != nil {
		return value.NoneVal(), err
	}

	e := eval.NewEvaluator()
	return e.Do_Blk(vals)
}
