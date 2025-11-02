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

func seriesEmpty(series core.Value) (core.Value, error) {
	seriesVal, err := assertSeries(series)
	if err != nil {
		return value.NewNoneVal(), err
	}
	return value.NewLogicVal(seriesVal.Length() == 0), nil
}

func seriesHeadQ(series core.Value) (core.Value, error) {
	seriesVal, err := assertSeries(series)
	if err != nil {
		return value.NewNoneVal(), err
	}
	return value.NewLogicVal(seriesVal.GetIndex() == 0), nil
}

func seriesTailQ(series core.Value) (core.Value, error) {
	seriesVal, err := assertSeries(series)
	if err != nil {
		return value.NewNoneVal(), err
	}
	return value.NewLogicVal(seriesVal.GetIndex() == seriesVal.Length()), nil
}
