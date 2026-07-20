package native

import (
	"github.com/marcin-radoszewski/viro/internal/core"
	"github.com/marcin-radoszewski/viro/internal/parse/dialect"
	"github.com/marcin-radoszewski/viro/internal/value"
	"github.com/marcin-radoszewski/viro/internal/verror"
)

// NativeParseDialect implements the new parse dialect for string and block pattern matching.
// This is the new implementation that will replace the old NativeParse (now parse-values).
func NativeParseDialect(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	// parse input rules [--case] [--all] [--part N]
	if len(args) != 2 {
		return value.NewNoneVal(), arityError("parse", 2, len(args))
	}

	input := args[0]
	rulesVal := args[1]

	// Validate input type (must be string or block)
	inputType := input.GetType()
	if inputType != value.TypeString && inputType != value.TypeBlock && inputType != value.TypeParen {
		return value.NewNoneVal(), typeError("parse", "string! block!", input)
	}

	// Extract rules block
	var rules []core.Value
	if blockVal, ok := value.AsBlockValue(rulesVal); ok {
		rules = blockVal.Elements
	} else {
		return value.NewNoneVal(), typeError("parse", "block!", rulesVal)
	}

	// Build options from refinements
	opts := dialect.DefaultOptions()

	// Check for --case refinement (case-sensitive matching)
	if caseVal, hasCaseRef := refValues["case"]; hasCaseRef && caseVal.GetType() != value.TypeNone {
		opts.CaseSensitive = true
	}

	// Check for --all refinement (match entire input)
	if allVal, hasAllRef := refValues["all"]; hasAllRef && allVal.GetType() != value.TypeNone {
		opts.MatchAll = true
	}

	// Check for --part refinement (parse only N elements)
	if partVal, hasPartRef := refValues["part"]; hasPartRef && partVal.GetType() != value.TypeNone {
		if intVal, ok := value.AsIntValue(partVal); ok {
			opts.Part = int(intVal)
		} else {
			return value.NewNoneVal(), verror.NewScriptError("type-mismatch", [3]string{"parse", "--part requires integer!", value.TypeToString(partVal.GetType())})
		}
	}

	// Check for --any refinement (allow partial matches)
	if anyVal, hasAnyRef := refValues["any"]; hasAnyRef && anyVal.GetType() != value.TypeNone {
		opts.Any = true
		opts.MatchAll = false // --any implies not matching all
	}

	// Execute parse
	success, err := dialect.Parse(input, rules, opts, eval)
	if err != nil {
		return value.NewNoneVal(), err
	}

	return value.NewLogicVal(success), nil
}
