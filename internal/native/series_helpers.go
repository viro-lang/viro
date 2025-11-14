package native

import (
	"fmt"
	"strings"

	"github.com/marcin-radoszewski/viro/internal/core"
	"github.com/marcin-radoszewski/viro/internal/value"
	"github.com/marcin-radoszewski/viro/internal/verror"
)

func seriesBack(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	seriesVal, err := assertSeries(args[0])
	if err != nil {
		return value.NewNoneVal(), err
	}

	currentIndex := seriesVal.GetIndex()
	if currentIndex <= 0 {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDOutOfBounds, [3]string{fmt.Sprintf("%d", currentIndex-1), fmt.Sprintf("%d", seriesVal.Length()), fmt.Sprintf("%d", currentIndex)})
	}

	newSeries := seriesVal.Clone()
	newSeries.SetIndex(currentIndex - 1)

	return newSeries.(core.Value), nil
}

func seriesNext(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	seriesVal, err := assertSeries(args[0])
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

func seriesHead(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	seriesVal, err := assertSeries(args[0])
	if err != nil {
		return value.NewNoneVal(), err
	}

	newSeries := seriesVal.Clone()
	newSeries.SetIndex(0)

	return newSeries.(core.Value), nil
}

func seriesIndex(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	seriesVal, err := assertSeries(args[0])
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

	currentIndex := seriesVal.GetIndex()
	length := seriesVal.Length()

	absoluteIndex := currentIndex + index

	if index < 0 || length == 0 || absoluteIndex >= length {
		return value.NewNoneVal(), nil
	}

	return seriesVal.ElementAt(absoluteIndex), nil
}

func seriesTail(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	seriesVal, err := assertSeries(args[0])
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

func isWordLike(t core.ValueType) bool {
	return t == value.TypeWord || t == value.TypeGetWord || t == value.TypeLitWord || t == value.TypeSetWord
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
		return 1, false, nil
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
	remaining := series.Length() - series.GetIndex()
	if count < 0 || count > remaining {
		return verror.NewScriptError(verror.ErrIDOutOfBounds, [3]string{fmt.Sprintf("%d", count), fmt.Sprintf("%d", remaining), fmt.Sprintf("%d", series.GetIndex())})
	}
	return nil
}

func validateIndex(index int, length int) error {
	if index < 0 || index >= length {
		return verror.NewScriptError(verror.ErrIDOutOfBounds, [3]string{fmt.Sprintf("%d", index), fmt.Sprintf("%d", length), ""})
	}
	return nil
}

func validateStringValue(val core.Value) (rune, error) {
	if strVal, ok := value.AsStringValue(val); ok && len(strVal.Runes()) == 1 {
		return strVal.Runes()[0], nil
	}
	return 0, verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"single character string", value.TypeToString(val.GetType()), ""})
}

func validateByteValue(val core.Value) (byte, error) {
	if byteVal, ok := value.AsIntValue(val); ok {
		if byteVal >= 0 && byteVal <= 255 {
			return byte(byteVal), nil
		}
	}
	return 0, verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"byte value (0-255)", value.TypeToString(val.GetType()), ""})
}

func seriesPick(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	seriesVal, err := assertSeries(args[0])
	if err != nil {
		return value.NewNoneVal(), err
	}

	indexVal := args[1]
	if indexVal.GetType() != value.TypeInteger {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"integer", value.TypeToString(indexVal.GetType()), ""})
	}

	index64, _ := value.AsIntValue(indexVal)
	zeroBasedIndex := int(index64) - 1

	if zeroBasedIndex < 0 || zeroBasedIndex >= seriesVal.Length() {
		return value.NewNoneVal(), nil
	}

	return seriesVal.ElementAt(zeroBasedIndex), nil
}

func hasRefinement(refValues map[string]core.Value, name string) bool {
	val, ok := refValues[name]
	return ok && val.GetType() == value.TypeLogic && val.Equals(value.NewLogicVal(true))
}

func getRefinementValue(refValues map[string]core.Value, name string) (bool, core.Value) {
	val, ok := refValues[name]
	return ok && val.GetType() != value.TypeNone, val
}

func trimDefault(input string) string {
	lines := strings.Split(input, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimSpace(line)
	}
	return strings.Join(lines, "\n")
}

func trimHead(input string) string {
	return strings.TrimLeft(input, " \t")
}

func trimTail(input string) string {
	return strings.TrimRight(input, " \t")
}

func trimAuto(input string) string {
	lines := strings.Split(input, "\n")
	if len(lines) == 0 {
		return input
	}

	if len(lines) == 1 {
		return strings.TrimSpace(input)
	}

	var baseIndent string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			for i, char := range line {
				if char != ' ' && char != '\t' {
					baseIndent = line[:i]
					break
				}
			}
			break
		}
	}

	if baseIndent == "" {
		return trimDefault(input)
	}

	result := make([]string, len(lines))
	for i, line := range lines {
		if strings.HasPrefix(line, baseIndent) {
			result[i] = line[len(baseIndent):]
		} else {
			result[i] = line
		}
	}

	return strings.Join(result, "\n")
}

func trimLines(input string) string {
	result := strings.ReplaceAll(input, "\n", " ")
	result = strings.ReplaceAll(result, "\r", " ")

	result = strings.Join(strings.Fields(result), " ")

	return strings.TrimSpace(result)
}

func trimAll(input string) string {
	result := strings.ReplaceAll(input, " ", "")
	result = strings.ReplaceAll(result, "\t", "")
	result = strings.ReplaceAll(result, "\n", "")
	result = strings.ReplaceAll(result, "\r", "")
	return result
}

func countTrue(flags ...bool) int {
	count := 0
	for _, flag := range flags {
		if flag {
			count++
		}
	}
	return count
}

func trimWith(input string, chars string) string {
	result := input
	for _, char := range chars {
		result = strings.ReplaceAll(result, string(char), "")
	}
	return result
}
