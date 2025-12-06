package native

import (
	"github.com/marcin-radoszewski/viro/internal/core"
	"github.com/marcin-radoszewski/viro/internal/value"
	"github.com/marcin-radoszewski/viro/internal/verror"
)

const maxDepth = 1000

func deepCopyBlock(block *value.BlockValue, visited map[core.Value]core.Value, depth int) (core.Value, error) {
	newElements := make([]core.Value, block.Length())
	newBlock := value.NewBlockVal(newElements).(value.Series)
	visited[block] = newBlock

	for i := 0; i < block.Length(); i++ {
		elem := block.ElementAt(i)
		copiedElem, err := deepCopyValue(elem, visited, depth+1)
		if err != nil {
			return nil, err
		}
		newElements[i] = copiedElem
	}
	return newBlock, nil
}

func isImmutableType(t core.ValueType) bool {
	switch t {
	case value.TypeInteger, value.TypeLogic, value.TypeNone, value.TypeFunction:
		return true
	default:
		return false
	}
}

func deepCopySeries(series value.Series, visited map[core.Value]core.Value, depth int) (value.Series, error) {
	if depth > maxDepth {
		return nil, verror.NewInternalError(verror.ErrIDStackOverflow, [3]string{"deep copy", "", ""})
	}

	if existing, ok := visited[series.(core.Value)]; ok {
		return existing.(value.Series), nil
	}

	var result value.Series
	switch series.GetType() {
	case value.TypeBlock, value.TypeParen:
		block := series.(*value.BlockValue)
		newBlock, err := deepCopyBlock(block, visited, depth)
		if err != nil {
			return nil, err
		}
		result = newBlock.(value.Series)
	case value.TypeString:
		str := series.(*value.StringValue)
		newStr := value.NewStrVal(str.String()).(value.Series)
		visited[series.(core.Value)] = newStr
		result = newStr
	case value.TypeBinary:
		bin := series.(*value.BinaryValue)
		newData := make([]byte, bin.Length())
		copy(newData, bin.Bytes())
		newBin := value.NewBinaryVal(newData).(value.Series)
		visited[series.(core.Value)] = newBin
		result = newBin
	default:
		return nil, verror.NewScriptError("type-mismatch", [3]string{"series type for deep copy", value.TypeToString(series.GetType()), ""})
	}

	return result, nil
}

func deepCopyValue(val core.Value, visited map[core.Value]core.Value, depth int) (core.Value, error) {
	if depth > maxDepth {
		return nil, verror.NewInternalError(verror.ErrIDStackOverflow, [3]string{"deep copy", "", ""})
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
		return deepCopyBlock(block, visited, depth)
	case value.TypeString:
		str := val.(*value.StringValue)
		newStr := value.NewStrVal(str.String()).(value.Series)
		visited[val] = newStr
		return newStr, nil
	case value.TypeBinary:
		bin := val.(*value.BinaryValue)
		newData := make([]byte, bin.Length())
		copy(newData, bin.Bytes())
		newBin := value.NewBinaryVal(newData).(value.Series)
		visited[val] = newBin
		return newBin, nil
	case value.TypeObject:
		obj, ok := value.AsObject(val)
		if !ok {
			return nil, verror.NewScriptError("type-mismatch", [3]string{"object", value.TypeToString(val.GetType()), ""})
		}
		newFrame := obj.Frame.Clone()
		newObjInstance := value.NewObject(newFrame)
		newObj := value.ObjectVal(newObjInstance)
		visited[val] = newObj

		// Deep copy the prototype chain
		if obj.ParentProto != nil {
			protoVal := value.ObjectVal(obj.ParentProto)
			copiedProto, err := deepCopyValue(protoVal, visited, depth+1)
			if err != nil {
				return nil, err
			}
			if copiedProtoObj, ok := value.AsObject(copiedProto); ok {
				newObjInstance.ParentProto = copiedProtoObj
			}
		}

		for _, binding := range newFrame.GetAll() {
			copiedVal, err := deepCopyValue(binding.Value, visited, depth+1)
			if err != nil {
				return nil, err
			}
			newFrame.Bind(binding.Symbol, copiedVal)
		}
		return newObj, nil
	default:
		return val, nil
	}
}
