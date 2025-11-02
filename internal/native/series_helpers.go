package native

import (
	"github.com/marcin-radoszewski/viro/internal/core"
	"github.com/marcin-radoszewski/viro/internal/value"
	"github.com/marcin-radoszewski/viro/internal/verror"
)

func seriesBack(series core.Value) (core.Value, error) {
	var currentIndex int

	switch s := series.(type) {
	case *value.BlockValue:
		currentIndex = s.GetIndex()
	case *value.StringValue:
		currentIndex = s.GetIndex()
	case *value.BinaryValue:
		currentIndex = s.GetIndex()
	default:
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"series", value.TypeToString(series.GetType()), ""})
	}

	if currentIndex <= 0 {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDOutOfBounds, [3]string{"series is at head", "", ""})
	}

	var newSeries core.Value
	switch s := series.(type) {
	case *value.BlockValue:
		newSeries = s.Clone()
		newSeries.(*value.BlockValue).SetIndex(currentIndex - 1)
	case *value.StringValue:
		newSeries = s.Clone()
		newSeries.(*value.StringValue).SetIndex(currentIndex - 1)
	case *value.BinaryValue:
		newSeries = s.Clone()
		newSeries.(*value.BinaryValue).SetIndex(currentIndex - 1)
	}

	return newSeries, nil
}

func seriesNext(series core.Value) (core.Value, error) {
	var currentIndex, length int

	switch s := series.(type) {
	case *value.BlockValue:
		currentIndex = s.GetIndex()
		length = len(s.Elements)
	case *value.StringValue:
		currentIndex = s.GetIndex()
		length = len(s.String())
	case *value.BinaryValue:
		currentIndex = s.GetIndex()
		length = len(s.Bytes())
	default:
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"series", value.TypeToString(series.GetType()), ""})
	}

	newIndex := currentIndex + 1
	if newIndex > length {
		newIndex = length
	}

	var newSeries core.Value
	switch s := series.(type) {
	case *value.BlockValue:
		newSeries = s.Clone()
		newSeries.(*value.BlockValue).SetIndex(newIndex)
	case *value.StringValue:
		newSeries = s.Clone()
		newSeries.(*value.StringValue).SetIndex(newIndex)
	case *value.BinaryValue:
		newSeries = s.Clone()
		newSeries.(*value.BinaryValue).SetIndex(newIndex)
	}

	return newSeries, nil
}

func seriesHead(series core.Value) (core.Value, error) {
	switch series.(type) {
	case *value.BlockValue, *value.StringValue, *value.BinaryValue:
		// Valid series types
	default:
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"series", value.TypeToString(series.GetType()), ""})
	}

	var newSeries core.Value
	switch s := series.(type) {
	case *value.BlockValue:
		newSeries = s.Clone()
		newSeries.(*value.BlockValue).SetIndex(0)
	case *value.StringValue:
		newSeries = s.Clone()
		newSeries.(*value.StringValue).SetIndex(0)
	case *value.BinaryValue:
		newSeries = s.Clone()
		newSeries.(*value.BinaryValue).SetIndex(0)
	}

	return newSeries, nil
}

func seriesIndex(series core.Value) (core.Value, error) {
	var index int

	switch s := series.(type) {
	case *value.BlockValue:
		index = s.GetIndex()
	case *value.StringValue:
		index = s.GetIndex()
	case *value.BinaryValue:
		index = s.GetIndex()
	default:
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"series", value.TypeToString(series.GetType()), ""})
	}

	return value.NewIntVal(int64(index + 1)), nil
}

func seriesAt(series core.Value, index int) (core.Value, error) {
	var length int
	var element core.Value

	switch s := series.(type) {
	case *value.BlockValue:
		length = len(s.Elements)
		if index >= 0 && index < length {
			element = s.Elements[index]
		}
	case *value.StringValue:
		str := s.String()
		length = len(str)
		if index >= 0 && index < length {
			element = value.NewStrVal(string(str[index]))
		}
	case *value.BinaryValue:
		data := s.Bytes()
		length = len(data)
		if index >= 0 && index < length {
			element = value.NewIntVal(int64(data[index]))
		}
	default:
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"series", value.TypeToString(series.GetType()), ""})
	}

	if index < 0 || index >= length {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDIndexOutOfRange, [3]string{"at", "series", "index out of range"})
	}

	return element, nil
}

func seriesTail(series core.Value) (core.Value, error) {
	var length int

	switch s := series.(type) {
	case *value.BlockValue:
		length = len(s.Elements)
		if length == 0 {
			return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDOutOfBounds, [3]string{"series is empty", "", ""})
		}
		return value.NewBlockVal(append([]core.Value{}, s.Elements[1:]...)), nil
	case *value.StringValue:
		str := s.String()
		length = len(str)
		if length == 0 {
			return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDOutOfBounds, [3]string{"series is empty", "", ""})
		}
		return value.NewStrVal(str[1:]), nil
	case *value.BinaryValue:
		data := s.Bytes()
		length = len(data)
		if length == 0 {
			return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDOutOfBounds, [3]string{"series is empty", "", ""})
		}
		return value.NewBinaryVal(data[1:]), nil
	default:
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"series", value.TypeToString(series.GetType()), ""})
	}
}
