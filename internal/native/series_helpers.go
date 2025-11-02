package native

import (
	"bytes"
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
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDOutOfBounds, [3]string{"-1", fmt.Sprintf("%d", seriesVal.Length()), ""})
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

	length := seriesVal.Length()
	if index < 0 || index >= length {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDOutOfBounds, [3]string{fmt.Sprintf("%d", index), fmt.Sprintf("%d", length), ""})
	}

	return seriesVal.ElementAt(index), nil
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
	if count < 0 {
		return verror.NewScriptError(verror.ErrIDOutOfBounds, [3]string{fmt.Sprintf("%d", count), fmt.Sprintf("%d", series.Length()), ""})
	}
	if count > series.Length() {
		return verror.NewScriptError(verror.ErrIDOutOfBounds, [3]string{fmt.Sprintf("%d", count), fmt.Sprintf("%d", series.Length()), ""})
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

func seriesPoke(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	indexVal := args[1]
	if indexVal.GetType() != value.TypeInteger {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"integer", value.TypeToString(indexVal.GetType()), ""})
	}

	index64, _ := value.AsIntValue(indexVal)
	zeroBasedIndex := int(index64) - 1

	if blk, ok := value.AsBlockValue(args[0]); ok {
		if err := validateIndex(zeroBasedIndex, len(blk.Elements)); err != nil {
			return value.NewNoneVal(), err
		}
		blk.Elements[zeroBasedIndex] = args[2]
		return args[2], nil
	}

	if str, ok := value.AsStringValue(args[0]); ok {
		if err := validateIndex(zeroBasedIndex, len(str.Runes())); err != nil {
			return value.NewNoneVal(), err
		}
		r, err := validateStringValue(args[2])
		if err != nil {
			return value.NewNoneVal(), err
		}
		str.SetRunes(append(str.Runes()[:zeroBasedIndex], append([]rune{r}, str.Runes()[zeroBasedIndex+1:]...)...))
		return args[2], nil
	}

	if bin, ok := value.AsBinaryValue(args[0]); ok {
		if err := validateIndex(zeroBasedIndex, len(bin.Bytes())); err != nil {
			return value.NewNoneVal(), err
		}
		b, err := validateByteValue(args[2])
		if err != nil {
			return value.NewNoneVal(), err
		}
		bin.Bytes()[zeroBasedIndex] = b
		return args[2], nil
	}

	return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDActionNoImpl, [3]string{value.TypeToString(args[0].GetType()), "", ""})
}

func seriesSelect(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	hasDefault := false
	defaultVal, ok := refValues["default"]
	if ok && defaultVal.GetType() != value.TypeNone {
		hasDefault = true
	}

	if blk, ok := value.AsBlockValue(args[0]); ok {
		for i := 0; i < len(blk.Elements); i++ {
			matches := false
			elem := blk.Elements[i]
			searchVal := args[1]

			if isWordLike(elem.GetType()) && isWordLike(searchVal.GetType()) {
				elemSymbol, _ := value.AsWordValue(elem)
				searchSymbol, _ := value.AsWordValue(searchVal)
				matches = elemSymbol == searchSymbol
			} else {
				matches = elem.Equals(searchVal)
			}

			if matches {
				if i+1 < len(blk.Elements) {
					return blk.Elements[i+1], nil
				}
				return value.NewNoneVal(), nil
			}
		}
		if hasDefault {
			return defaultVal, nil
		}
		return value.NewNoneVal(), nil
	}

	if str, ok := value.AsStringValue(args[0]); ok {
		if targetStr, ok2 := value.AsStringValue(args[1]); ok2 {
			haystack := string(str.Runes())
			needle := string(targetStr.Runes())
			pos := strings.Index(haystack, needle)
			if pos == -1 {
				if hasDefault {
					return defaultVal, nil
				}
				return value.NewNoneVal(), nil
			}
			remainder := haystack[pos+len(needle):]
			return value.NewStringValue(remainder), nil
		}
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"string", value.TypeToString(args[1].GetType()), ""})
	}

	if bin, ok := value.AsBinaryValue(args[0]); ok {
		if targetBytes, ok2 := value.AsBinaryValue(args[1]); ok2 {
			haystack := bin.Bytes()
			needle := targetBytes.Bytes()
			pos := bytes.Index(haystack, needle)
			if pos == -1 {
				if hasDefault {
					return defaultVal, nil
				}
				return value.NewNoneVal(), nil
			}
			remainder := haystack[pos+len(needle):]
			return value.NewBinaryValue(remainder), nil
		}
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"binary", value.TypeToString(args[1].GetType()), ""})
	}

	return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDActionNoImpl, [3]string{value.TypeToString(args[0].GetType()), "", ""})
}

func seriesClear(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if blk, ok := value.AsBlockValue(args[0]); ok {
		blk.Elements = blk.Elements[:0]
		blk.SetIndex(0)
		return args[0], nil
	}

	if str, ok := value.AsStringValue(args[0]); ok {
		str.SetRunes(str.Runes()[:0])
		str.SetIndex(0)
		return args[0], nil
	}

	if bin, ok := value.AsBinaryValue(args[0]); ok {
		bin.SetIndex(0)
		bin.Remove(bin.Length())
		return args[0], nil
	}

	return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDActionNoImpl, [3]string{value.TypeToString(args[0].GetType()), "", ""})
}

func seriesChange(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if blk, ok := value.AsBlockValue(args[0]); ok {
		currentIndex := blk.GetIndex()
		if err := validateIndex(currentIndex, len(blk.Elements)); err != nil {
			return value.NewNoneVal(), err
		}
		blk.Elements[currentIndex] = args[1]
		newIndex := currentIndex + 1
		if newIndex > len(blk.Elements) {
			newIndex = len(blk.Elements)
		}
		blk.SetIndex(newIndex)
		return args[1], nil
	}

	if str, ok := value.AsStringValue(args[0]); ok {
		currentIndex := str.GetIndex()
		runes := str.Runes()
		if err := validateIndex(currentIndex, len(runes)); err != nil {
			return value.NewNoneVal(), err
		}
		r, err := validateStringValue(args[1])
		if err != nil {
			return value.NewNoneVal(), err
		}
		newRunes := append(runes[:currentIndex], append([]rune{r}, runes[currentIndex+1:]...)...)
		str.SetRunes(newRunes)
		newIndex := currentIndex + 1
		if newIndex > len(newRunes) {
			newIndex = len(newRunes)
		}
		str.SetIndex(newIndex)
		return args[1], nil
	}

	if bin, ok := value.AsBinaryValue(args[0]); ok {
		currentIndex := bin.GetIndex()
		bytes := bin.Bytes()
		if err := validateIndex(currentIndex, len(bytes)); err != nil {
			return value.NewNoneVal(), err
		}
		b, err := validateByteValue(args[1])
		if err != nil {
			return value.NewNoneVal(), err
		}
		bytes[currentIndex] = b
		newIndex := currentIndex + 1
		if newIndex > len(bytes) {
			newIndex = len(bytes)
		}
		bin.SetIndex(newIndex)
		return args[1], nil
	}

	return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDActionNoImpl, [3]string{value.TypeToString(args[0].GetType()), "", ""})
}

func seriesPick(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	indexVal := args[1]
	if indexVal.GetType() != value.TypeInteger {
		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDTypeMismatch, [3]string{"integer", value.TypeToString(indexVal.GetType()), ""})
	}

	index64, _ := value.AsIntValue(indexVal)
	zeroBasedIndex := int(index64) - 1

	if blk, ok := value.AsBlockValue(args[0]); ok {
		if zeroBasedIndex < 0 || zeroBasedIndex >= len(blk.Elements) {
			return value.NewNoneVal(), nil
		}
		return blk.Elements[zeroBasedIndex], nil
	}

	if str, ok := value.AsStringValue(args[0]); ok {
		runes := str.Runes()
		if zeroBasedIndex < 0 || zeroBasedIndex >= len(runes) {
			return value.NewNoneVal(), nil
		}
		return value.NewStringValue(string(runes[zeroBasedIndex])), nil
	}

	if bin, ok := value.AsBinaryValue(args[0]); ok {
		bytes := bin.Bytes()
		if zeroBasedIndex < 0 || zeroBasedIndex >= len(bytes) {
			return value.NewNoneVal(), nil
		}
		return value.NewIntVal(int64(bytes[zeroBasedIndex])), nil
	}

	return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDActionNoImpl, [3]string{value.TypeToString(args[0].GetType()), "", ""})
}

func seriesTrim(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if str, ok := value.AsStringValue(args[0]); ok {
		input := string(str.Runes())

		hasHead := hasRefinement(refValues, "head")
		hasTail := hasRefinement(refValues, "tail")
		hasAuto := hasRefinement(refValues, "auto")
		hasLines := hasRefinement(refValues, "lines")
		hasAll := hasRefinement(refValues, "all")
		hasWith, withVal := getRefinementValue(refValues, "with")

		flags := []bool{hasHead, hasTail, hasAuto, hasLines, hasAll, hasWith}
		flagCount := 0
		for _, flag := range flags {
			if flag {
				flagCount++
			}
		}

		if flagCount > 1 {
			return value.NewNoneVal(), verror.NewScriptError(
				verror.ErrIDInvalidOperation,
				[3]string{"trim refinements are mutually exclusive", "", ""},
			)
		}

		if hasWith {
			withStr, ok := value.AsStringValue(withVal)
			if !ok {
				return value.NewNoneVal(), verror.NewScriptError(
					verror.ErrIDTypeMismatch,
					[3]string{"string", value.TypeToString(withVal.GetType()), "--with"},
				)
			}
			charsToRemove := string(withStr.Runes())
			return value.NewStringValue(trimWith(input, charsToRemove)), nil
		}

		if flagCount == 0 {
			return value.NewStringValue(trimDefault(input)), nil
		}

		if hasHead {
			return value.NewStringValue(trimHead(input)), nil
		}
		if hasTail {
			return value.NewStringValue(trimTail(input)), nil
		}
		if hasAuto {
			return value.NewStringValue(trimAuto(input)), nil
		}
		if hasLines {
			return value.NewStringValue(trimLines(input)), nil
		}
		if hasAll {
			return value.NewStringValue(trimAll(input)), nil
		}

		return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDAssertionFailed, [3]string{"unexpected trim refinement state", "", ""})
	}

	return value.NewNoneVal(), verror.NewScriptError(verror.ErrIDActionNoImpl, [3]string{value.TypeToString(args[0].GetType()), "", ""})
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

func trimWith(input string, chars string) string {
	result := input
	for _, char := range chars {
		result = strings.ReplaceAll(result, string(char), "")
	}
	return result
}
