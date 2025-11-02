package native

import (
	"strings"

	"github.com/marcin-radoszewski/viro/internal/core"
	"github.com/marcin-radoszewski/viro/internal/value"
	"github.com/marcin-radoszewski/viro/internal/verror"
)

func seriesFirst(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	seriesVal, err := assertSeries(args[0])
	if err != nil {
		return value.NewNoneVal(), err
	}

	val, err := seriesVal.FirstValue()
	if err != nil {
		if strings.Contains(err.Error(), "empty series") {
			return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDEmptySeries, [3]string{"first element", "", ""})
		}
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDOutOfBounds, [3]string{"", "", ""})
	}
	return val, nil
}

func seriesLast(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	seriesVal, err := assertSeries(args[0])
	if err != nil {
		return value.NewNoneVal(), err
	}

	val, err := seriesVal.LastValue()
	if err != nil {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDEmptySeries, [3]string{"last element", "", ""})
	}
	return val, nil
}

func seriesAppend(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	seriesVal, err := assertSeries(args[0])
	if err != nil {
		return value.NewNoneVal(), err
	}

	err = seriesVal.AppendValue(args[1])
	if err != nil {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"compatible value", value.TypeToString(args[1].GetType()), ""})
	}
	return args[0], nil
}

func seriesInsert(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	seriesVal, err := assertSeries(args[0])
	if err != nil {
		return value.NewNoneVal(), err
	}

	err = seriesVal.InsertValue(args[1])
	if err != nil {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"compatible value", value.TypeToString(args[1].GetType()), ""})
	}
	return args[0], nil
}

func seriesLength(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	seriesVal, err := assertSeries(args[0])
	if err != nil {
		return value.NewNoneVal(), err
	}

	return value.NewIntVal(int64(seriesVal.Length())), nil
}

func seriesCopy(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	seriesVal, err := assertSeries(args[0])
	if err != nil {
		return value.NewNoneVal(), err
	}

	count, hasPart, err := readPartCount(refValues)
	if err != nil {
		return value.NewNoneVal(), err
	}

	if !hasPart {
		return seriesVal.Clone().(core.Value), nil
	}

	if err := validatePartCount(seriesVal, count); err != nil {
		return value.NewNoneVal(), err
	}

	copied, err := seriesVal.CopyPart(count)
	if err != nil {
		return value.NewNoneVal(), err
	}
	return copied.(core.Value), nil
}

func seriesRemove(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	seriesVal, err := assertSeries(args[0])
	if err != nil {
		return value.NewNoneVal(), err
	}

	count, hasPart, err := readPartCount(refValues)
	if err != nil {
		return value.NewNoneVal(), err
	}

	if !hasPart {
		count = 1
	}

	if err := validatePartCount(seriesVal, count); err != nil {
		return value.NewNoneVal(), err
	}

	err = seriesVal.RemoveCount(count)
	if err != nil {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDOutOfBounds, [3]string{"", "", ""})
	}
	return args[0], nil
}

func seriesSkip(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	seriesVal, err := assertSeries(args[0])
	if err != nil {
		return value.NewNoneVal(), err
	}

	countVal := args[1]
	if countVal.GetType() != value.TypeInteger {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"integer", value.TypeToString(countVal.GetType()), ""})
	}

	count64, _ := value.AsIntValue(countVal)
	count := int(count64)

	seriesVal.SkipBy(count)
	return args[0], nil
}

func seriesTake(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	seriesVal, err := assertSeries(args[0])
	if err != nil {
		return value.NewNoneVal(), err
	}

	countVal := args[1]
	if countVal.GetType() != value.TypeInteger {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"integer", value.TypeToString(countVal.GetType()), ""})
	}

	count64, _ := value.AsIntValue(countVal)
	count := int(count64)

	taken := seriesVal.TakeCount(count)
	return taken.(core.Value), nil
}

func seriesClear(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	seriesVal, err := assertSeries(args[0])
	if err != nil {
		return value.NewNoneVal(), err
	}

	seriesVal.ClearSeries()
	return args[0], nil
}

func seriesChange(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	seriesVal, err := assertSeries(args[0])
	if err != nil {
		return value.NewNoneVal(), err
	}

	err = seriesVal.ChangeValue(args[1])
	if err != nil {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"compatible value", value.TypeToString(args[1].GetType()), ""})
	}
	return args[1], nil
}
