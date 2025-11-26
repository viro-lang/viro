package native

import (
	"fmt"

	"github.com/marcin-radoszewski/viro/internal/core"
	"github.com/marcin-radoszewski/viro/internal/value"
	"github.com/marcin-radoszewski/viro/internal/verror"
)

func seriesFirst(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	seriesVal, err := assertSeries(args[0])
	if err != nil {
		return value.NewNoneVal(), err
	}

	if seriesVal.Length() == 0 {
		return value.NewNoneVal(), nil
	}

	val, err := seriesVal.FirstValue()
	if err != nil {
		return value.NewNoneVal(), nil
	}
	return val, nil
}

func seriesLast(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	seriesVal, err := assertSeries(args[0])
	if err != nil {
		return value.NewNoneVal(), err
	}

	if seriesVal.Length() == 0 {
		return value.NewNoneVal(), nil
	}

	val, err := seriesVal.LastValue()
	if err != nil {
		return value.NewNoneVal(), nil
	}
	return val, nil
}

func seriesOrdinalAccess(series core.Value, ordinal int) (core.Value, error) {
	s, err := assertSeries(series)
	if err != nil {
		return value.NewNoneVal(), err
	}

	current := s.GetIndex()
	length := s.Length()

	if current == 0 {
		if ordinal < 0 || length <= ordinal {
			return value.NewNoneVal(), nil
		}
		return s.ElementAt(ordinal), nil
	}

	target := current + ordinal
	if ordinal < 0 || target >= length {
		return value.NewNoneVal(), nil
	}
	return s.ElementAt(target), nil
}

func seriesSecond(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	return seriesOrdinalAccess(args[0], 1)
}

func seriesThird(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	return seriesOrdinalAccess(args[0], 2)
}

func seriesFourth(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	return seriesOrdinalAccess(args[0], 3)
}

func seriesSixth(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	return seriesOrdinalAccess(args[0], 5)
}

func seriesSeventh(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	return seriesOrdinalAccess(args[0], 6)
}

func seriesEighth(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	return seriesOrdinalAccess(args[0], 7)
}

func seriesNinth(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	return seriesOrdinalAccess(args[0], 8)
}

func seriesTenth(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	return seriesOrdinalAccess(args[0], 9)
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
		count = seriesVal.Length() - seriesVal.GetIndex()
	} else {
		if err := validatePartCount(seriesVal, count); err != nil {
			return value.NewNoneVal(), err
		}
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
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDOutOfBounds, [3]string{fmt.Sprintf("%d", count), fmt.Sprintf("%d", seriesVal.Length()), fmt.Sprintf("%d", seriesVal.GetIndex())})
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

	currentIndex := seriesVal.GetIndex()
	length := seriesVal.Length()

	newIndex := currentIndex + count
	if newIndex < 0 {
		newIndex = 0
	}
	if newIndex > length {
		newIndex = length
	}

	newSeries := seriesVal.Clone()
	newSeries.SetIndex(newIndex)

	return newSeries.(core.Value), nil
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

	if count < 0 {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDOutOfBounds, [3]string{fmt.Sprintf("%d", count), fmt.Sprintf("%d", seriesVal.Length()), fmt.Sprintf("%d", seriesVal.GetIndex())})
	}

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
