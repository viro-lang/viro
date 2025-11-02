package native

import (
	"fmt"

	"github.com/marcin-radoszewski/viro/internal/core"
	"github.com/marcin-radoszewski/viro/internal/value"
	"github.com/marcin-radoszewski/viro/internal/verror"
)

func seriesBack(series core.Value) (core.Value, error) {
	seriesVal, err := assertSeries(series)
	if err != nil {
		return value.NewNoneVal(), err
	}

	currentIndex := seriesVal.GetIndex()
	if currentIndex <= 0 {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDOutOfBounds, [3]string{"-1", fmt.Sprintf("%d", seriesVal.Length()), ""})
	}

	newSeries := seriesVal.Clone()
	newSeries.SetIndex(currentIndex - 1)

	return newSeries.(core.Value), nil
}

func seriesNext(series core.Value) (core.Value, error) {
	seriesVal, err := assertSeries(series)
	if err != nil {
		return value.NewNoneVal(), err
	}

	currentIndex := seriesVal.GetIndex()
	length := seriesVal.Length()

	newIndex := min(currentIndex+1, length)

	newSeries := seriesVal.Clone()
	newSeries.SetIndex(newIndex)

	return newSeries.(core.Value), nil
}

func seriesHead(series core.Value) (core.Value, error) {
	seriesVal, err := assertSeries(series)
	if err != nil {
		return value.NewNoneVal(), err
	}

	newSeries := seriesVal.Clone()
	newSeries.SetIndex(0)

	return newSeries.(core.Value), nil
}

func seriesIndex(series core.Value) (core.Value, error) {
	seriesVal, err := assertSeries(series)
	if err != nil {
		return value.NewNoneVal(), err
	}

	index := seriesVal.GetIndex()
	return value.NewIntVal(int64(index + 1)), nil
}

func seriesAt(series core.Value, index int) (core.Value, error) {
	seriesVal, err := assertSeries(series)
	if err != nil {
		return value.NewNoneVal(), err
	}

	length := seriesVal.Length()
	if index < 0 || index >= length {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDOutOfBounds, [3]string{fmt.Sprintf("%d", index), fmt.Sprintf("%d", length), ""})
	}

	return seriesVal.ElementAt(index), nil
}

func seriesTail(series core.Value) (core.Value, error) {
	seriesVal, err := assertSeries(series)
	if err != nil {
		return value.NewNoneVal(), err
	}

	tailSeries := seriesVal.Clone()
	tailSeries.SetIndex(seriesVal.Length())
	return tailSeries, nil
}

func assertSeries(series core.Value) (value.Series, error) {
	seriesVal, ok := series.(value.Series)
	if !ok {
		return nil, verror.NewScriptError(verror.ErrIDActionNoImpl, [3]string{value.TypeToString(series.GetType()), "", ""})
	}
	return seriesVal, nil
}

func seriesEmpty(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	seriesVal, err := assertSeries(args[0])
	if err != nil {
		return value.NewNoneVal(), err
	}
	return value.NewLogicVal(seriesVal.Length() == 0), nil
}

func seriesHeadQ(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	seriesVal, err := assertSeries(args[0])
	if err != nil {
		return value.NewNoneVal(), err
	}
	return value.NewLogicVal(seriesVal.GetIndex() == 0), nil
}

func seriesTailQ(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	seriesVal, err := assertSeries(args[0])
	if err != nil {
		return value.NewNoneVal(), err
	}
	return value.NewLogicVal(seriesVal.GetIndex() == seriesVal.Length()), nil
}

func readPartCount(refValues map[string]core.Value) (int, bool, error) {
	partVal, hasPart := refValues["part"]
	hasPart = hasPart && partVal.GetType() != value.TypeNone

	if !hasPart {
		return 1, false, nil // default count is 1 when no --part refinement
	}

	if partVal.GetType() != value.TypeInteger {
		return 0, false, verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"integer", value.TypeToString(partVal.GetType()), ""})
	}

	count64, ok := value.AsIntValue(partVal)
	if !ok {
		return 0, false, verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"integer", value.TypeToString(partVal.GetType()), ""})
	}

	return int(count64), true, nil
}

func validatePartCount(series value.Series, count int) error {
	if count < 0 {
		return verror.NewScriptError(verror.ErrIDOutOfBounds, [3]string{fmt.Sprintf("%d", count), fmt.Sprintf("%d", series.Length()), ""})
	}
	if count > series.Length() {
		return verror.NewScriptError(verror.ErrIDOutOfBounds, [3]string{fmt.Sprintf("%d", count), fmt.Sprintf("%d", series.Length()), ""})
	}
	return nil
}
