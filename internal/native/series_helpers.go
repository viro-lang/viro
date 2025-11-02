package native

import (
	"fmt"

	"github.com/marcin-radoszewski/viro/internal/core"
	"github.com/marcin-radoszewski/viro/internal/value"
	"github.com/marcin-radoszewski/viro/internal/verror"
)

// ensureArgCount validates that the correct number of arguments is provided for a native function.
func ensureArgCount(args []core.Value, expected int, funcName string) error {
	if len(args) != expected {
		return verror.NewScriptError(verror.ErrIDArgCount, [3]string{funcName, fmt.Sprintf("%d", expected), fmt.Sprintf("%d", len(args))})
	}
	return nil
}

// seriesBack implements the shared back logic for all series types.
// It creates a new series reference with index moved backward by 1.
// Returns an error if already at head position.
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
