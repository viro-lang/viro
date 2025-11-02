package native

import (
	"fmt"

	"github.com/marcin-radoszewski/viro/internal/core"
	"github.com/marcin-radoszewski/viro/internal/value"
	"github.com/marcin-radoszewski/viro/internal/verror"
)

func BlockFirst(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	blk, ok := value.AsBlockValue(args[0])
	if !ok {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"block", value.TypeToString(args[0].GetType()), ""})
	}

	if len(blk.Elements) == 0 {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDEmptySeries, [3]string{"first element", "", ""})
	}

	if blk.GetIndex() >= len(blk.Elements) {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDOutOfBounds, [3]string{fmt.Sprintf("%d", blk.GetIndex()), fmt.Sprintf("%d", len(blk.Elements)), ""})
	}

	return blk.Elements[blk.GetIndex()], nil
}

func BlockLast(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	blk, ok := value.AsBlockValue(args[0])
	if !ok {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"block", value.TypeToString(args[0].GetType()), ""})
	}

	if len(blk.Elements) == 0 {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDEmptySeries, [3]string{"last element", "", ""})
	}

	return blk.Elements[len(blk.Elements)-1], nil
}

func BlockAppend(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	blk, ok := value.AsBlockValue(args[0])
	if !ok {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"block", value.TypeToString(args[0].GetType()), ""})
	}

	blk.Elements = append(blk.Elements, args[1])

	return args[0], nil
}

func BlockInsert(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	blk, ok := value.AsBlockValue(args[0])
	if !ok {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"block", value.TypeToString(args[0].GetType()), ""})
	}

	blk.Elements = append([]core.Value{args[1]}, blk.Elements...)

	return args[0], nil
}

func BlockLength(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	blk, ok := value.AsBlockValue(args[0])
	if !ok {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"block", value.TypeToString(args[0].GetType()), ""})
	}

	return value.NewIntVal(int64(len(blk.Elements))), nil
}

func BlockCopy(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	blk, ok := value.AsBlockValue(args[0])
	if !ok {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"block", value.TypeToString(args[0].GetType()), ""})
	}

	count, hasPart, err := readPartCount(refValues)
	if err != nil {
		return value.NewNoneVal(), err
	}

	if hasPart {
		if err := validatePartCount(blk, count); err != nil {
			return value.NewNoneVal(), err
		}
		elems := make([]core.Value, count)
		copy(elems, blk.Elements[:count])
		return value.NewBlockVal(elems), nil
	}

	return value.NewBlockVal(append([]core.Value{}, blk.Elements...)), nil
}

func BlockFind(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	blk, ok := value.AsBlockValue(args[0])
	if !ok {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"block", value.TypeToString(args[0].GetType()), ""})
	}

	sought := args[1]

	lastVal, hasLast := refValues["last"]
	isLast := hasLast && lastVal.GetType() == value.TypeLogic && lastVal.Equals(value.NewLogicVal(true))

	if isLast {
		for i := len(blk.Elements) - 1; i >= 0; i-- {
			if blk.Elements[i].Equals(sought) {
				return value.NewIntVal(int64(i + 1)), nil
			}
		}
	} else {
		for i, v := range blk.Elements {
			if v.Equals(sought) {
				return value.NewIntVal(int64(i + 1)), nil
			}
		}
	}

	return value.NewNoneVal(), nil
}

func BlockRemove(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	blk, ok := value.AsBlockValue(args[0])
	if !ok {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"block", value.TypeToString(args[0].GetType()), ""})
	}

	count, _, err := readPartCount(refValues)
	if err != nil {
		return value.NewNoneVal(), err
	}

	if err := validatePartCount(blk, count); err != nil {
		return value.NewNoneVal(), err
	}

	blk.SetIndex(0)
	blk.Remove(count)
	return args[0], nil
}

func BlockSkip(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	blk, ok := value.AsBlockValue(args[0])
	if !ok {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"block", value.TypeToString(args[0].GetType()), ""})
	}

	countVal := args[1]
	if countVal.GetType() != value.TypeInteger {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"integer", value.TypeToString(countVal.GetType()), ""})
	}

	count64, _ := value.AsIntValue(countVal)
	count := int(count64)

	newIndex := blk.GetIndex() + count
	if newIndex < 0 || newIndex > len(blk.Elements) {
		newIndex = len(blk.Elements)
	}
	blk.SetIndex(newIndex)

	return args[0], nil
}

func BlockNext(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	return seriesNext(args[0])
}

func BlockBack(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	return seriesBack(args[0])
}

func BlockHead(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	return seriesHead(args[0])
}

func BlockTake(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	blk, ok := value.AsBlockValue(args[0])
	if !ok {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"block", value.TypeToString(args[0].GetType()), ""})
	}

	countVal := args[1]
	if countVal.GetType() != value.TypeInteger {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"integer", value.TypeToString(countVal.GetType()), ""})
	}

	count64, _ := value.AsIntValue(countVal)
	count := int(count64)

	start := blk.GetIndex()
	end := min(start+count, len(blk.Elements))
	newElements := blk.Elements[start:end]
	blk.SetIndex(end)

	return value.NewBlockVal(newElements), nil
}

func BlockSort(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	blk, ok := value.AsBlockValue(args[0])
	if !ok {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"block", value.TypeToString(args[0].GetType()), ""})
	}

	if len(blk.Elements) == 0 {
		return args[0], nil
	}

	firstType := blk.Elements[0].GetType()
	for _, v := range blk.Elements {
		if v.GetType() != firstType || (v.GetType() != value.TypeInteger && v.GetType() != value.TypeString) {
			return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDNotComparable, [3]string{"sort", "mixed types", ""})
		}
	}

	value.SortBlock(blk)
	return args[0], nil
}

func BlockReverse(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	blk, ok := value.AsBlockValue(args[0])
	if !ok {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"block", value.TypeToString(args[0].GetType()), ""})
	}

	for i, j := 0, len(blk.Elements)-1; i < j; i, j = i+1, j-1 {
		blk.Elements[i], blk.Elements[j] = blk.Elements[j], blk.Elements[i]
	}

	return args[0], nil
}

func BlockIndex(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	return seriesIndex(args[0])
}

func BlockAt(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	indexVal := args[1]
	if indexVal.GetType() != value.TypeInteger {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"integer", value.TypeToString(indexVal.GetType()), ""})
	}

	index64, _ := value.AsIntValue(indexVal)
	zeroBasedIndex := int(index64) - 1

	return seriesAt(args[0], zeroBasedIndex)
}

func BlockTail(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	return seriesTail(args[0])
}

func BlockEmpty(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	return seriesEmpty(args[0])
}

func BlockHeadQ(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	return seriesHeadQ(args[0])
}

func BlockTailQ(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	return seriesTailQ(args[0])
}
