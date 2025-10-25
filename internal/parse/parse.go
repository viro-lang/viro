package parse

import (
	"fmt"
	"strings"

	"github.com/marcin-radoszewski/viro/internal/core"
	"github.com/marcin-radoszewski/viro/internal/parse/peg"
	"github.com/marcin-radoszewski/viro/internal/value"
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

func ParseEval(input string) ([]core.Value, error) {
	return Parse(input)
}

func Format(val core.Value) string {
	switch val.GetType() {
	case value.TypeNone:
		return "none"
	case value.TypeLogic:
		if logic, ok := value.AsLogic(val); ok {
			if logic {
				return "true"
			}
			return "false"
		}
		return "logic"
	case value.TypeInteger:
		if num, ok := value.AsInteger(val); ok {
			return fmt.Sprintf("%d", num)
		}
		return "integer"
	case value.TypeString:
		if str, ok := value.AsString(val); ok {
			return fmt.Sprintf("\"%s\"", str.String())
		}
		return "string"
	case value.TypeWord:
		if word, ok := value.AsWord(val); ok {
			return word
		}
		return "word"
	case value.TypeSetWord:
		if word, ok := value.AsWord(val); ok {
			return word + ":"
		}
		return "set-word"
	case value.TypeGetWord:
		if word, ok := value.AsWord(val); ok {
			return ":" + word
		}
		return "get-word"
	case value.TypeLitWord:
		if word, ok := value.AsWord(val); ok {
			return "'" + word
		}
		return "lit-word"
	case value.TypeBlock:
		if block, ok := value.AsBlock(val); ok {
			var parts []string
			for _, elem := range block.Elements {
				parts = append(parts, Format(elem))
			}
			return "[" + strings.Join(parts, " ") + "]"
		}
		return "block"
	case value.TypeParen:
		if block, ok := value.AsBlock(val); ok {
			var parts []string
			for _, elem := range block.Elements {
				parts = append(parts, Format(elem))
			}
			return "(" + strings.Join(parts, " ") + ")"
		}
		return "paren"
	case value.TypeFunction:
		return "function"
	case value.TypePath:
		if path, ok := value.AsPath(val); ok {
			return path.String()
		}
		return "path"
	default:
		return "unknown"
	}
}
