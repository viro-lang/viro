package native

import (
	"github.com/marcin-radoszewski/viro/internal/core"
	"github.com/marcin-radoszewski/viro/internal/value"
	"github.com/marcin-radoszewski/viro/internal/verror"
)

func seriesBack(series core.Value) (core.Value, error) {
	seriesVal, ok := series.(value.Series)
	if !ok {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"series", value.TypeToString(series.GetType()), ""})
	}

	currentIndex := seriesVal.GetIndex()
	if currentIndex <= 0 {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDOutOfBounds, [3]string{"series is at head", "", ""})
	}

	newSeries := seriesVal.Clone()
	newSeries.SetIndex(currentIndex - 1)

	return newSeries.(core.Value), nil
}

func seriesNext(series core.Value) (core.Value, error) {
	seriesVal, ok := series.(value.Series)
	if !ok {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"series", value.TypeToString(series.GetType()), ""})
	}

	currentIndex := seriesVal.GetIndex()
	length := seriesVal.Length()

	newIndex := currentIndex + 1
	if newIndex > length {
		newIndex = length
	}

	newSeries := seriesVal.Clone()
	newSeries.SetIndex(newIndex)

	return newSeries.(core.Value), nil
}

func seriesHead(series core.Value) (core.Value, error) {
	seriesVal, ok := series.(value.Series)
	if !ok {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"series", value.TypeToString(series.GetType()), ""})
	}

	newSeries := seriesVal.Clone()
	newSeries.SetIndex(0)

	return newSeries.(core.Value), nil
}

func seriesIndex(series core.Value) (core.Value, error) {
	seriesVal, ok := series.(value.Series)
	if !ok {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"series", value.TypeToString(series.GetType()), ""})
	}

	index := seriesVal.GetIndex()
	return value.NewIntVal(int64(index + 1)), nil
}

func seriesAt(series core.Value, index int) (core.Value, error) {
	seriesVal, ok := series.(value.Series)
	if !ok {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"series", value.TypeToString(series.GetType()), ""})
	}

	length := seriesVal.Length()
	if index < 0 || index >= length {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDIndexOutOfRange, [3]string{"at", "series", "index out of range"})
	}

	return seriesVal.ElementAt(index), nil
}

func seriesTail(series core.Value) (core.Value, error) {
	seriesVal, ok := series.(value.Series)
	if !ok {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"series", value.TypeToString(series.GetType()), ""})
	}

	tailSeries := seriesVal.Clone()
	tailSeries.SetIndex(seriesVal.Length())
	return tailSeries, nil
}
