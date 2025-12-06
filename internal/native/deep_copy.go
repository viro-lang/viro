package native

import (
	"github.com/marcin-radoszewski/viro/internal/core"
	"github.com/marcin-radoszewski/viro/internal/value"
	"github.com/marcin-radoszewski/viro/internal/verror"
)

const maxDepth = 1000

func isImmutableType(t core.ValueType) bool {
	switch t {
	case value.TypeInteger, value.TypeString, value.TypeBinary, value.TypeLogic, value.TypeNone, value.TypeFunction:
		return true
	default:
		return false
	}
}

func deepCopySeries(series value.Series, visited map[core.Value]core.Value, depth int) (value.Series, error) {
	if depth > maxDepth {
		return nil, verror.NewScriptError("invalid-operation", [3]string{"deep copy recursion limit exceeded", "", ""})
	}

	if existing, ok := visited[series.(core.Value)]; ok {
		return existing.(value.Series), nil
	}

	var result value.Series
	switch series.GetType() {
	case value.TypeBlock, value.TypeParen:
		block := series.(*value.BlockValue)
		newElements := make([]core.Value, block.Length())
		newBlock := value.NewBlockValue(newElements)
		visited[series.(core.Value)] = newBlock

		for i := 0; i < block.Length(); i++ {
			elem := block.ElementAt(i)
			copiedElem, err := deepCopyValue(elem, visited, depth+1)
			if err != nil {
				return nil, err
			}
			newElements[i] = copiedElem
		}
		result = newBlock
	case value.TypeString:
		str := series.(*value.StringValue)
		newStr := value.NewStringValue(str.String())
		visited[series.(core.Value)] = newStr
		result = newStr
	case value.TypeBinary:
		bin := series.(*value.BinaryValue)
		newData := make([]byte, bin.Length())
		copy(newData, bin.Bytes())
		newBin := value.NewBinaryValue(newData)
		visited[series.(core.Value)] = newBin
		result = newBin
	default:
		return nil, verror.NewScriptError("type-mismatch", [3]string{"series type for deep copy", value.TypeToString(series.GetType()), ""})
	}

	return result, nil
}

func deepCopyValue(val core.Value, visited map[core.Value]core.Value, depth int) (core.Value, error) {
	if depth > maxDepth {
		return value.NewNoneVal(), verror.NewScriptError("invalid-operation", [3]string{"deep copy recursion limit exceeded", "", ""})
	}

	if existing, ok := visited[val]; ok {
		return existing, nil
	}

	if isImmutableType(val.GetType()) {
		return val, nil
	}

	switch val.GetType() {
	case value.TypeBlock, value.TypeParen:
		block := val.(*value.BlockValue)
		newElements := make([]core.Value, block.Length())
		newBlock := value.NewBlockValue(newElements)
		visited[val] = newBlock

		for i := 0; i < block.Length(); i++ {
			elem := block.ElementAt(i)
			copiedElem, err := deepCopyValue(elem, visited, depth+1)
			if err != nil {
				return value.NewNoneVal(), err
			}
			newElements[i] = copiedElem
		}
		return newBlock, nil
	case value.TypeObject:
		obj, ok := value.AsObject(val)
		if !ok {
			return value.NewNoneVal(), verror.NewScriptError("type-mismatch", [3]string{"object", value.TypeToString(val.GetType()), ""})
		}
		newFrame := obj.Frame.Clone()
		newObj := value.ObjectVal(value.NewObject(newFrame))
		visited[val] = newObj

		for _, binding := range newFrame.GetAll() {
			copiedVal, err := deepCopyValue(binding.Value, visited, depth+1)
			if err != nil {
				return value.NewNoneVal(), err
			}
			newFrame.Bind(binding.Symbol, copiedVal)
		}
		return newObj, nil
	default:
		return val, nil
	}
}
