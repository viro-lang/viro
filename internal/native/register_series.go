package native

import (
	"fmt"

	"github.com/marcin-radoszewski/viro/internal/core"
	"github.com/marcin-radoszewski/viro/internal/value"
)

func registerBlockSeriesActions() {
	RegisterActionImpl(value.TypeBlock, "first", value.NewNativeFunction("first", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, seriesFirst, false, nil))
	RegisterActionImpl(value.TypeBlock, "last", value.NewNativeFunction("last", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, seriesLast, false, nil))
	RegisterActionImpl(value.TypeBlock, "second", value.NewNativeFunction("second", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, seriesSecond, false, nil))
	RegisterActionImpl(value.TypeBlock, "third", value.NewNativeFunction("third", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, seriesThird, false, nil))
	RegisterActionImpl(value.TypeBlock, "fourth", value.NewNativeFunction("fourth", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, seriesFourth, false, nil))
	RegisterActionImpl(value.TypeBlock, "sixth", value.NewNativeFunction("sixth", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, seriesSixth, false, nil))
	RegisterActionImpl(value.TypeBlock, "seventh", value.NewNativeFunction("seventh", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, seriesSeventh, false, nil))
	RegisterActionImpl(value.TypeBlock, "eighth", value.NewNativeFunction("eighth", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, seriesEighth, false, nil))
	RegisterActionImpl(value.TypeBlock, "ninth", value.NewNativeFunction("ninth", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, seriesNinth, false, nil))
	RegisterActionImpl(value.TypeBlock, "tenth", value.NewNativeFunction("tenth", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, seriesTenth, false, nil))
	RegisterActionImpl(value.TypeBlock, "append", value.NewNativeFunction("append", []value.ParamSpec{
		value.NewParamSpec("series", true),
		value.NewParamSpec("value", true),
	}, seriesAppend, false, nil))
	RegisterActionImpl(value.TypeBlock, "insert", value.NewNativeFunction("insert", []value.ParamSpec{
		value.NewParamSpec("series", true),
		value.NewParamSpec("value", true),
	}, seriesInsert, false, nil))
	RegisterActionImpl(value.TypeBlock, "length?", value.NewNativeFunction("length?", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, seriesLength, false, nil))
	RegisterActionImpl(value.TypeBlock, "copy", value.NewNativeFunction("copy", []value.ParamSpec{
		value.NewParamSpec("series", true),
		value.NewRefinementSpec("part", true),
	}, seriesCopy, false, nil))
	RegisterActionImpl(value.TypeBlock, "find", value.NewNativeFunction("find", []value.ParamSpec{
		value.NewParamSpec("series", true),
		value.NewParamSpec("value", true),
		value.NewRefinementSpec("last", false),
	}, BlockFind, false, nil))
	RegisterActionImpl(value.TypeBlock, "remove", value.NewNativeFunction("remove", []value.ParamSpec{
		value.NewParamSpec("series", true),
		value.NewRefinementSpec("part", true),
	}, seriesRemove, false, nil))
	RegisterActionImpl(value.TypeBlock, "skip", value.NewNativeFunction("skip", []value.ParamSpec{
		value.NewParamSpec("series", true),
		value.NewParamSpec("count", true),
	}, seriesSkip, false, nil))
	RegisterActionImpl(value.TypeBlock, "next", value.NewNativeFunction("next", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, seriesNext, false, nil))
	RegisterActionImpl(value.TypeBlock, "back", value.NewNativeFunction("back", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, seriesBack, false, nil))
	RegisterActionImpl(value.TypeBlock, "head", value.NewNativeFunction("head", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, seriesHead, false, nil))
	RegisterActionImpl(value.TypeBlock, "index?", value.NewNativeFunction("index?", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, seriesIndex, false, nil))
	RegisterActionImpl(value.TypeBlock, "take", value.NewNativeFunction("take", []value.ParamSpec{
		value.NewParamSpec("series", true),
		value.NewParamSpec("count", true),
	}, seriesTake, false, nil))
	RegisterActionImpl(value.TypeBlock, "sort", value.NewNativeFunction("sort", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, BlockSort, false, nil))
	RegisterActionImpl(value.TypeBlock, "reverse", value.NewNativeFunction("reverse", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, BlockReverse, false, nil))
	RegisterActionImpl(value.TypeBlock, "at", value.NewNativeFunction("at", []value.ParamSpec{
		value.NewParamSpec("series", true),
		value.NewParamSpec("index", true),
	}, BlockAt, false, nil))
	RegisterActionImpl(value.TypeBlock, "pick", value.NewNativeFunction("pick", []value.ParamSpec{
		value.NewParamSpec("series", true),
		value.NewParamSpec("index", true),
	}, seriesPick, false, nil))
	RegisterActionImpl(value.TypeBlock, "poke", value.NewNativeFunction("poke", []value.ParamSpec{
		value.NewParamSpec("series", true),
		value.NewParamSpec("index", true),
		value.NewParamSpec("value", true),
	}, BlockPoke, false, nil))
	RegisterActionImpl(value.TypeBlock, "select", value.NewNativeFunction("select", []value.ParamSpec{
		value.NewParamSpec("target", true),
		value.NewParamSpec("value", true),
		value.NewRefinementSpec("default", true),
	}, BlockSelect, false, nil))
	RegisterActionImpl(value.TypeBlock, "clear", value.NewNativeFunction("clear", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, seriesClear, false, nil))
	RegisterActionImpl(value.TypeBlock, "change", value.NewNativeFunction("change", []value.ParamSpec{
		value.NewParamSpec("series", true),
		value.NewParamSpec("value", true),
	}, seriesChange, false, nil))
	RegisterActionImpl(value.TypeBlock, "trim", value.NewNativeFunction("trim", []value.ParamSpec{
		value.NewParamSpec("series", true),
		value.NewRefinementSpec("head", false),
		value.NewRefinementSpec("tail", false),
		value.NewRefinementSpec("auto", false),
		value.NewRefinementSpec("lines", false),
		value.NewRefinementSpec("all", false),
		value.NewRefinementSpec("with", true),
	}, BlockTrim, false, nil))
	RegisterActionImpl(value.TypeBlock, "tail", value.NewNativeFunction("tail", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, seriesTail, false, nil))

	RegisterActionImpl(value.TypeBlock, "empty?", value.NewNativeFunction("empty?", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, seriesEmpty, false, nil))
	RegisterActionImpl(value.TypeBlock, "head?", value.NewNativeFunction("head?", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, seriesHeadQ, false, nil))
	RegisterActionImpl(value.TypeBlock, "tail?", value.NewNativeFunction("tail?", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, seriesTailQ, false, nil))
	RegisterActionImpl(value.TypeBlock, "intersect", value.NewNativeFunction("intersect", []value.ParamSpec{
		value.NewParamSpec("s1", true),
		value.NewParamSpec("s2", true),
	}, BlockIntersect, false, nil))
	RegisterActionImpl(value.TypeBlock, "difference", value.NewNativeFunction("difference", []value.ParamSpec{
		value.NewParamSpec("s1", true),
		value.NewParamSpec("s2", true),
	}, BlockDifference, false, nil))
	RegisterActionImpl(value.TypeBlock, "union", value.NewNativeFunction("union", []value.ParamSpec{
		value.NewParamSpec("s1", true),
		value.NewParamSpec("s2", true),
	}, BlockUnion, false, nil))
}

func registerStringSeriesActions() {
	RegisterActionImpl(value.TypeString, "first", value.NewNativeFunction("first", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, seriesFirst, false, nil))
	RegisterActionImpl(value.TypeString, "last", value.NewNativeFunction("last", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, seriesLast, false, nil))
	RegisterActionImpl(value.TypeString, "second", value.NewNativeFunction("second", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, seriesSecond, false, nil))
	RegisterActionImpl(value.TypeString, "third", value.NewNativeFunction("third", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, seriesThird, false, nil))
	RegisterActionImpl(value.TypeString, "fourth", value.NewNativeFunction("fourth", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, seriesFourth, false, nil))
	RegisterActionImpl(value.TypeString, "sixth", value.NewNativeFunction("sixth", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, seriesSixth, false, nil))
	RegisterActionImpl(value.TypeString, "seventh", value.NewNativeFunction("seventh", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, seriesSeventh, false, nil))
	RegisterActionImpl(value.TypeString, "eighth", value.NewNativeFunction("eighth", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, seriesEighth, false, nil))
	RegisterActionImpl(value.TypeString, "ninth", value.NewNativeFunction("ninth", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, seriesNinth, false, nil))
	RegisterActionImpl(value.TypeString, "tenth", value.NewNativeFunction("tenth", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, seriesTenth, false, nil))
	RegisterActionImpl(value.TypeString, "append", value.NewNativeFunction("append", []value.ParamSpec{
		value.NewParamSpec("series", true),
		value.NewParamSpec("value", true),
	}, seriesAppend, false, nil))
	RegisterActionImpl(value.TypeString, "insert", value.NewNativeFunction("insert", []value.ParamSpec{
		value.NewParamSpec("series", true),
		value.NewParamSpec("value", true),
	}, seriesInsert, false, nil))
	RegisterActionImpl(value.TypeString, "length?", value.NewNativeFunction("length?", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, seriesLength, false, nil))
	RegisterActionImpl(value.TypeString, "copy", value.NewNativeFunction("copy", []value.ParamSpec{
		value.NewParamSpec("series", true),
		value.NewRefinementSpec("part", true),
	}, seriesCopy, false, nil))
	RegisterActionImpl(value.TypeString, "find", value.NewNativeFunction("find", []value.ParamSpec{
		value.NewParamSpec("series", true),
		value.NewParamSpec("value", true),
		value.NewRefinementSpec("last", false),
	}, StringFind, false, nil))
	RegisterActionImpl(value.TypeString, "remove", value.NewNativeFunction("remove", []value.ParamSpec{
		value.NewParamSpec("series", true),
		value.NewRefinementSpec("part", true),
	}, seriesRemove, false, nil))
	RegisterActionImpl(value.TypeString, "skip", value.NewNativeFunction("skip", []value.ParamSpec{
		value.NewParamSpec("series", true),
		value.NewParamSpec("count", true),
	}, seriesSkip, false, nil))
	RegisterActionImpl(value.TypeString, "next", value.NewNativeFunction("next", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, seriesNext, false, nil))
	RegisterActionImpl(value.TypeString, "back", value.NewNativeFunction("back", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, seriesBack, false, nil))
	RegisterActionImpl(value.TypeString, "head", value.NewNativeFunction("head", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, seriesHead, false, nil))
	RegisterActionImpl(value.TypeString, "index?", value.NewNativeFunction("index?", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, seriesIndex, false, nil))
	RegisterActionImpl(value.TypeString, "at", value.NewNativeFunction("at", []value.ParamSpec{
		value.NewParamSpec("series", true),
		value.NewParamSpec("index", true),
	}, StringAt, false, nil))
	RegisterActionImpl(value.TypeString, "pick", value.NewNativeFunction("pick", []value.ParamSpec{
		value.NewParamSpec("series", true),
		value.NewParamSpec("index", true),
	}, seriesPick, false, nil))
	RegisterActionImpl(value.TypeString, "poke", value.NewNativeFunction("poke", []value.ParamSpec{
		value.NewParamSpec("series", true),
		value.NewParamSpec("index", true),
		value.NewParamSpec("value", true),
	}, StringPoke, false, nil))
	RegisterActionImpl(value.TypeString, "select", value.NewNativeFunction("select", []value.ParamSpec{
		value.NewParamSpec("target", true),
		value.NewParamSpec("value", true),
		value.NewRefinementSpec("default", true),
	}, StringSelect, false, nil))
	RegisterActionImpl(value.TypeString, "clear", value.NewNativeFunction("clear", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, seriesClear, false, nil))
	RegisterActionImpl(value.TypeString, "change", value.NewNativeFunction("change", []value.ParamSpec{
		value.NewParamSpec("series", true),
		value.NewParamSpec("value", true),
	}, seriesChange, false, nil))
	RegisterActionImpl(value.TypeString, "trim", value.NewNativeFunction("trim", []value.ParamSpec{
		value.NewParamSpec("series", true),
		value.NewRefinementSpec("head", false),
		value.NewRefinementSpec("tail", false),
		value.NewRefinementSpec("auto", false),
		value.NewRefinementSpec("lines", false),
		value.NewRefinementSpec("all", false),
		value.NewRefinementSpec("with", true),
	}, StringTrim, false, nil))
	RegisterActionImpl(value.TypeString, "tail", value.NewNativeFunction("tail", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, seriesTail, false, nil))
	RegisterActionImpl(value.TypeString, "empty?", value.NewNativeFunction("empty?", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, seriesEmpty, false, nil))
	RegisterActionImpl(value.TypeString, "head?", value.NewNativeFunction("head?", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, seriesHeadQ, false, nil))
	RegisterActionImpl(value.TypeString, "tail?", value.NewNativeFunction("tail?", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, seriesTailQ, false, nil))
	RegisterActionImpl(value.TypeString, "sort", value.NewNativeFunction("sort", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, StringSort, false, nil))
	RegisterActionImpl(value.TypeString, "reverse", value.NewNativeFunction("reverse", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, StringReverse, false, nil))
	RegisterActionImpl(value.TypeString, "take", value.NewNativeFunction("take", []value.ParamSpec{
		value.NewParamSpec("series", true),
		value.NewParamSpec("count", true),
	}, seriesTake, false, nil))
	RegisterActionImpl(value.TypeString, "intersect", value.NewNativeFunction("intersect", []value.ParamSpec{
		value.NewParamSpec("s1", true),
		value.NewParamSpec("s2", true),
	}, StringIntersect, false, nil))
	RegisterActionImpl(value.TypeString, "difference", value.NewNativeFunction("difference", []value.ParamSpec{
		value.NewParamSpec("s1", true),
		value.NewParamSpec("s2", true),
	}, StringDifference, false, nil))
	RegisterActionImpl(value.TypeString, "union", value.NewNativeFunction("union", []value.ParamSpec{
		value.NewParamSpec("s1", true),
		value.NewParamSpec("s2", true),
	}, StringUnion, false, nil))
	RegisterActionImpl(value.TypeString, "codepoints-of", value.NewNativeFunction("codepoints-of", []value.ParamSpec{
		value.NewParamSpec("string", true),
	}, CodepointsOf, false, nil))
	RegisterActionImpl(value.TypeString, "codepoint-at", value.NewNativeFunction("codepoint-at", []value.ParamSpec{
		value.NewParamSpec("string", true),
		value.NewParamSpec("index", true),
		value.NewRefinementSpec("default", true),
	}, CodepointAt, false, nil))
	RegisterActionImpl(value.TypeString, "string-from-codepoints", value.NewNativeFunction("string-from-codepoints", []value.ParamSpec{
		value.NewParamSpec("codepoints", true),
	}, StringFromCodepoints, false, nil))
}

func registerBinarySeriesActions() {
	RegisterActionImpl(value.TypeBinary, "first", value.NewNativeFunction("first", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, seriesFirst, false, nil))
	RegisterActionImpl(value.TypeBinary, "last", value.NewNativeFunction("last", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, seriesLast, false, nil))
	RegisterActionImpl(value.TypeBinary, "second", value.NewNativeFunction("second", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, seriesSecond, false, nil))
	RegisterActionImpl(value.TypeBinary, "third", value.NewNativeFunction("third", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, seriesThird, false, nil))
	RegisterActionImpl(value.TypeBinary, "fourth", value.NewNativeFunction("fourth", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, seriesFourth, false, nil))
	RegisterActionImpl(value.TypeBinary, "sixth", value.NewNativeFunction("sixth", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, seriesSixth, false, nil))
	RegisterActionImpl(value.TypeBinary, "seventh", value.NewNativeFunction("seventh", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, seriesSeventh, false, nil))
	RegisterActionImpl(value.TypeBinary, "eighth", value.NewNativeFunction("eighth", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, seriesEighth, false, nil))
	RegisterActionImpl(value.TypeBinary, "ninth", value.NewNativeFunction("ninth", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, seriesNinth, false, nil))
	RegisterActionImpl(value.TypeBinary, "tenth", value.NewNativeFunction("tenth", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, seriesTenth, false, nil))
	RegisterActionImpl(value.TypeBinary, "append", value.NewNativeFunction("append", []value.ParamSpec{
		value.NewParamSpec("series", true),
		value.NewParamSpec("value", true),
	}, seriesAppend, false, nil))
	RegisterActionImpl(value.TypeBinary, "insert", value.NewNativeFunction("insert", []value.ParamSpec{
		value.NewParamSpec("series", true),
		value.NewParamSpec("value", true),
	}, seriesInsert, false, nil))
	RegisterActionImpl(value.TypeBinary, "length?", value.NewNativeFunction("length?", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, seriesLength, false, nil))
	RegisterActionImpl(value.TypeBinary, "copy", value.NewNativeFunction("copy", []value.ParamSpec{
		value.NewParamSpec("series", true),
		value.NewRefinementSpec("part", true),
	}, seriesCopy, false, nil))
	RegisterActionImpl(value.TypeBinary, "find", value.NewNativeFunction("find", []value.ParamSpec{
		value.NewParamSpec("series", true),
		value.NewParamSpec("value", true),
		value.NewRefinementSpec("last", false),
	}, BinaryFind, false, nil))
	RegisterActionImpl(value.TypeBinary, "remove", value.NewNativeFunction("remove", []value.ParamSpec{
		value.NewParamSpec("series", true),
		value.NewRefinementSpec("part", true),
	}, seriesRemove, false, nil))
	RegisterActionImpl(value.TypeBinary, "skip", value.NewNativeFunction("skip", []value.ParamSpec{
		value.NewParamSpec("series", true),
		value.NewParamSpec("count", true),
	}, seriesSkip, false, nil))
	RegisterActionImpl(value.TypeBinary, "next", value.NewNativeFunction("next", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, seriesNext, false, nil))
	RegisterActionImpl(value.TypeBinary, "back", value.NewNativeFunction("back", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, seriesBack, false, nil))
	RegisterActionImpl(value.TypeBinary, "head", value.NewNativeFunction("head", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, seriesHead, false, nil))
	RegisterActionImpl(value.TypeBinary, "index?", value.NewNativeFunction("index?", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, seriesIndex, false, nil))
	RegisterActionImpl(value.TypeBinary, "at", value.NewNativeFunction("at", []value.ParamSpec{
		value.NewParamSpec("series", true),
		value.NewParamSpec("index", true),
	}, BinaryAt, false, nil))
	RegisterActionImpl(value.TypeBinary, "pick", value.NewNativeFunction("pick", []value.ParamSpec{
		value.NewParamSpec("series", true),
		value.NewParamSpec("index", true),
	}, seriesPick, false, nil))
	RegisterActionImpl(value.TypeBinary, "poke", value.NewNativeFunction("poke", []value.ParamSpec{
		value.NewParamSpec("series", true),
		value.NewParamSpec("index", true),
		value.NewParamSpec("value", true),
	}, BinaryPoke, false, nil))
	RegisterActionImpl(value.TypeBinary, "select", value.NewNativeFunction("select", []value.ParamSpec{
		value.NewParamSpec("target", true),
		value.NewParamSpec("value", true),
		value.NewRefinementSpec("default", true),
	}, BinarySelect, false, nil))
	RegisterActionImpl(value.TypeBinary, "clear", value.NewNativeFunction("clear", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, seriesClear, false, nil))
	RegisterActionImpl(value.TypeBinary, "change", value.NewNativeFunction("change", []value.ParamSpec{
		value.NewParamSpec("series", true),
		value.NewParamSpec("value", true),
	}, seriesChange, false, nil))
	RegisterActionImpl(value.TypeBinary, "tail", value.NewNativeFunction("tail", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, seriesTail, false, nil))
	RegisterActionImpl(value.TypeBinary, "empty?", value.NewNativeFunction("empty?", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, seriesEmpty, false, nil))
	RegisterActionImpl(value.TypeBinary, "head?", value.NewNativeFunction("head?", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, seriesHeadQ, false, nil))
	RegisterActionImpl(value.TypeBinary, "tail?", value.NewNativeFunction("tail?", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, seriesTailQ, false, nil))
	RegisterActionImpl(value.TypeBinary, "intersect", value.NewNativeFunction("intersect", []value.ParamSpec{
		value.NewParamSpec("s1", true),
		value.NewParamSpec("s2", true),
	}, BinaryIntersect, false, nil))
	RegisterActionImpl(value.TypeBinary, "difference", value.NewNativeFunction("difference", []value.ParamSpec{
		value.NewParamSpec("s1", true),
		value.NewParamSpec("s2", true),
	}, BinaryDifference, false, nil))
	RegisterActionImpl(value.TypeBinary, "union", value.NewNativeFunction("union", []value.ParamSpec{
		value.NewParamSpec("s1", true),
		value.NewParamSpec("s2", true),
	}, BinaryUnion, false, nil))
	RegisterActionImpl(value.TypeBinary, "sort", value.NewNativeFunction("sort", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, BinarySort, false, nil))
	RegisterActionImpl(value.TypeBinary, "take", value.NewNativeFunction("take", []value.ParamSpec{
		value.NewParamSpec("series", true),
		value.NewParamSpec("count", true),
	}, seriesTake, false, nil))
}

func registerSeriesTypeImpls() {
	registerBlockSeriesActions()
	registerStringSeriesActions()
	registerBinarySeriesActions()
}

func RegisterSeriesNatives(rootFrame core.Frame) {
	registered := make(map[string]bool)

	registerAndBind := func(name string, val core.Value) {
		if val.GetType() == value.TypeNone {
			panic(fmt.Sprintf("RegisterSeriesNatives: attempted to register nil value for '%s'", name))
		}
		if registered[name] {
			panic(fmt.Sprintf("RegisterSeriesNatives: duplicate registration of '%s'", name))
		}

		rootFrame.Bind(name, val)

		registered[name] = true
	}

	registerSeriesTypeImpls()

	registerAndBind("first", CreateAction("first", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, &NativeDoc{
		Category: "Series",
		Summary:  "Returns the first element of a series or none when no element remains (empty series or cursor at tail)",
		Parameters: []ParamDoc{
			{Name: "series", Type: "block! string! binary!", Description: "The series to get first element from"},
		},
		Returns:  "any! The first element of the series or none",
		Examples: []string{"first [1 2 3]  ; => 1", `first "hello"  ; => "h"`, "first #{DEADBEEF}  ; => 222"},
		SeeAlso:  []string{"last", "skip", "take"},
		Tags:     []string{"series"},
	}))

	registerAndBind("last", CreateAction("last", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, &NativeDoc{
		Category: "Series",
		Summary:  "Returns the last element of a series or none if the series is empty",
		Parameters: []ParamDoc{
			{Name: "series", Type: "block! string! binary!", Description: "The series to get last element from"},
		},
		Returns:  "any! The last element of the series or none if empty",
		Examples: []string{"last [1 2 3]  ; => 3", `last "hello"  ; => "o"`, "last #{DEADBEEF}  ; => 239"},
		SeeAlso:  []string{"first", "skip", "take"},
		Tags:     []string{"series"},
	}))

	registerAndBind("second", CreateAction("second", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, &NativeDoc{
		Category: "Series",
		Summary:  "Returns the second element of a series or none when no element remains (empty series or cursor at tail)",
		Parameters: []ParamDoc{
			{Name: "series", Type: "block! string! binary!", Description: "The series to get second element from"},
		},
		Returns:  "any! The second element of the series or none",
		Examples: []string{"second [1 2 3]  ; => 2", `second "hello"  ; => "e"`, "second #{DEADBEEF}  ; => 173"},
		SeeAlso:  []string{"first", "third", "at"},
		Tags:     []string{"series"},
	}))

	registerAndBind("third", CreateAction("third", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, &NativeDoc{
		Category: "Series",
		Summary:  "Returns the third element of a series or none when no element remains (empty series or cursor at tail)",
		Parameters: []ParamDoc{
			{Name: "series", Type: "block! string! binary!", Description: "The series to get third element from"},
		},
		Returns:  "any! The third element of the series or none",
		Examples: []string{"third [1 2 3]  ; => 3", `third "hello"  ; => "l"`, "third #{DEADBEEF}  ; => 190"},
		SeeAlso:  []string{"second", "fourth", "at"},
		Tags:     []string{"series"},
	}))

	registerAndBind("fourth", CreateAction("fourth", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, &NativeDoc{
		Category: "Series",
		Summary:  "Returns the fourth element of a series or none when no element remains (empty series or cursor at tail)",
		Parameters: []ParamDoc{
			{Name: "series", Type: "block! string! binary!", Description: "The series to get fourth element from"},
		},
		Returns:  "any! The fourth element of the series or none",
		Examples: []string{"fourth [1 2 3 4]  ; => 4", `fourth "hello"  ; => "l"`, "fourth #{DEADBEEF}  ; => 239"},
		SeeAlso:  []string{"third", "sixth", "at"},
		Tags:     []string{"series"},
	}))

	registerAndBind("sixth", CreateAction("sixth", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, &NativeDoc{
		Category: "Series",
		Summary:  "Returns the sixth element of a series or none when no element remains (empty series or cursor at tail)",
		Parameters: []ParamDoc{
			{Name: "series", Type: "block! string! binary!", Description: "The series to get sixth element from"},
		},
		Returns:  "any! The sixth element of the series or none",
		Examples: []string{"sixth [1 2 3 4 5 6]  ; => 6", `sixth "hello world"  ; => " "`, "sixth #{DEADBEEF0102}  ; => 2"},
		SeeAlso:  []string{"fourth", "seventh", "at"},
		Tags:     []string{"series"},
	}))

	registerAndBind("seventh", CreateAction("seventh", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, &NativeDoc{
		Category: "Series",
		Summary:  "Returns the seventh element of a series or none when no element remains (empty series or cursor at tail)",
		Parameters: []ParamDoc{
			{Name: "series", Type: "block! string! binary!", Description: "The series to get seventh element from"},
		},
		Returns:  "any! The seventh element of the series or none",
		Examples: []string{"seventh [1 2 3 4 5 6 7]  ; => 7", `seventh "hello world"  ; => "w"`, "seventh #{DEADBEEF010203}  ; => 3"},
		SeeAlso:  []string{"sixth", "eighth", "at"},
		Tags:     []string{"series"},
	}))

	registerAndBind("eighth", CreateAction("eighth", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, &NativeDoc{
		Category: "Series",
		Summary:  "Returns the eighth element of a series or none when no element remains (empty series or cursor at tail)",
		Parameters: []ParamDoc{
			{Name: "series", Type: "block! string! binary!", Description: "The series to get eighth element from"},
		},
		Returns:  "any! The eighth element of the series or none",
		Examples: []string{"eighth [1 2 3 4 5 6 7 8]  ; => 8", `eighth "hello world"  ; => "o"`, "eighth #{DEADBEEF0102030405}  ; => 4"},
		SeeAlso:  []string{"seventh", "ninth", "at"},
		Tags:     []string{"series"},
	}))

	registerAndBind("ninth", CreateAction("ninth", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, &NativeDoc{
		Category: "Series",
		Summary:  "Returns the ninth element of a series or none when no element remains (empty series or cursor at tail)",
		Parameters: []ParamDoc{
			{Name: "series", Type: "block! string! binary!", Description: "The series to get ninth element from"},
		},
		Returns:  "any! The ninth element of the series or none",
		Examples: []string{"ninth [1 2 3 4 5 6 7 8 9]  ; => 9", `ninth "hello world"  ; => "r"`, "ninth #{DEADBEEF010203040506}  ; => 5"},
		SeeAlso:  []string{"eighth", "tenth", "at"},
		Tags:     []string{"series"},
	}))

	registerAndBind("tenth", CreateAction("tenth", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, &NativeDoc{
		Category: "Series",
		Summary:  "Returns the tenth element of a series or none when no element remains (empty series or cursor at tail)",
		Parameters: []ParamDoc{
			{Name: "series", Type: "block! string! binary!", Description: "The series to get tenth element from"},
		},
		Returns:  "any! The tenth element of the series or none",
		Examples: []string{"tenth [1 2 3 4 5 6 7 8 9 10]  ; => 10", `tenth "hello world"  ; => "l"`, "tenth #{DEADBEEF01020304050607}  ; => 6"},
		SeeAlso:  []string{"ninth", "at"},
		Tags:     []string{"series"},
	}))

	registerAndBind("append", CreateAction("append", []value.ParamSpec{
		value.NewParamSpec("series", true),
		value.NewParamSpec("value", true),
	}, &NativeDoc{
		Category: "Series",
		Summary:  "Appends a value to the end of a series",
		Parameters: []ParamDoc{
			{Name: "series", Type: "block! string! binary!", Description: "The series to append to"},
			{Name: "value", Type: "any!", Description: "The value to append"},
		},
		Returns:  "block! string! binary! The modified series",
		Examples: []string{"append [1 2] 3  ; => [1 2 3]", `append "hel" "lo"  ; => "hello"`, "append #{DEAD} 190  ; => #{DEADBE}"},
		SeeAlso:  []string{"insert", "skip", "take"},
		Tags:     []string{"series", "modification"},
	}))

	registerAndBind("at", CreateAction("at", []value.ParamSpec{
		value.NewParamSpec("series", true),
		value.NewParamSpec("index", true),
	}, &NativeDoc{
		Category: "Series",
		Summary:  "Returns the element at the specified 1-based index from a series or none when no element remains (empty series or cursor at tail)",
		Parameters: []ParamDoc{
			{Name: "series", Type: "block! string! binary!", Description: "The series to get element from"},
			{Name: "index", Type: "integer!", Description: "1-based index of the element to return"},
		},
		Returns:  "any! The element at the specified index or none",
		Examples: []string{"at [1 2 3] 2  ; => 2", `at "hello" 1  ; => "h"`, "at #{DEADBEEF} 3  ; => 190"},
		SeeAlso:  []string{"first", "last", "skip", "take"},
		Tags:     []string{"series", "indexing"},
	}))

	registerAndBind("pick", CreateAction("pick", []value.ParamSpec{
		value.NewParamSpec("series", true),
		value.NewParamSpec("index", true),
	}, &NativeDoc{
		Category: "Series",
		Summary:  "Returns the element at the specified 1-based index from a series",
		Parameters: []ParamDoc{
			{Name: "series", Type: "block! string! binary!", Description: "The series to get element from"},
			{Name: "index", Type: "integer!", Description: "1-based index of the element to return"},
		},
		Returns:  "any! The element at the specified index, or none if index is out of bounds",
		Examples: []string{"pick [1 2 3] 2  ; => 2", `pick "hello" 1  ; => "h"`, "pick #{DEADBEEF} 3  ; => 190", "pick [1 2 3] 10  ; => none"},
		SeeAlso:  []string{"at", "first", "last", "poke"},
		Tags:     []string{"series", "indexing"},
	}))

	registerAndBind("poke", CreateAction("poke", []value.ParamSpec{
		value.NewParamSpec("series", true),
		value.NewParamSpec("index", true),
		value.NewParamSpec("value", true),
	}, &NativeDoc{
		Category: "Series",
		Summary:  "Sets the element at the specified 1-based index in a series",
		Parameters: []ParamDoc{
			{Name: "series", Type: "block! string! binary!", Description: "The series to modify"},
			{Name: "index", Type: "integer!", Description: "1-based index of the element to set"},
			{Name: "value", Type: "any!", Description: "The new value to set at the index"},
		},
		Returns:  "any! The value that was set",
		Examples: []string{"poke [1 2 3] 2 99  ; => 99, series becomes [1 99 3]", `poke "hello" 1 "H"  ; => "H", series becomes "Hello"`},
		SeeAlso:  []string{"at", "change", "insert"},
		Tags:     []string{"series", "modification", "indexing"},
	}))

	registerAndBind("select", CreateAction("select", []value.ParamSpec{
		value.NewParamSpec("target", true),
		value.NewParamSpec("value", true),
		value.NewRefinementSpec("default", true),
	}, &NativeDoc{
		Category: "Series",
		Summary:  "Finds a value in a series or field in an object and returns associated value",
		Description: `Polymorphic lookup action that works on series and objects.

For blocks: searches for the value and returns the next element (key-value pairs).
For strings/binary: finds the pattern and returns the remaining portion after it.
For objects: looks up the field name and returns its value (searches prototype chain).

The --default refinement provides a fallback when the value/field is not found.`,
		Parameters: []ParamDoc{
			{Name: "target", Type: "block! string! binary! object!", Description: "The series or object to search"},
			{Name: "value", Type: "any!", Description: "For series: value to find. For objects: field name (word or string)"},
			{Name: "--default", Type: "any!", Description: "Optional fallback value when search/lookup fails"},
		},
		Returns: "any! The found value, or default, or none",
		Examples: []string{
			"select [a 1 b 2] 'b  ; => 2",
			`select "hello world" " "  ; => "world"`,
			"obj: object [x: 10]\nselect obj 'x  ; => 10",
			"select obj 'missing --default 99  ; => 99",
		},
		SeeAlso: []string{"find", "at", "index?", "put", "get"},
		Tags:    []string{"series", "search", "objects", "lookup"},
	}))

	registerAndBind("clear", CreateAction("clear", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, &NativeDoc{
		Category: "Series",
		Summary:  "Removes all elements from a series and resets index to head",
		Parameters: []ParamDoc{
			{Name: "series", Type: "block! string! binary!", Description: "The series to clear"},
		},
		Returns:  "block! string! binary! The cleared series (same reference)",
		Examples: []string{"clear [1 2 3]  ; => [], series becomes empty with index at head", `clear "hello"  ; => "", series becomes empty with index at head`},
		SeeAlso:  []string{"append", "insert", "remove"},
		Tags:     []string{"series", "modification"},
	}))

	registerAndBind("change", CreateAction("change", []value.ParamSpec{
		value.NewParamSpec("series", true),
		value.NewParamSpec("value", true),
	}, &NativeDoc{
		Category: "Series",
		Summary:  "Replaces the element at the current index with a new value",
		Parameters: []ParamDoc{
			{Name: "series", Type: "block! string! binary!", Description: "The series to modify"},
			{Name: "value", Type: "any!", Description: "The new value to set at current index (string: single character, binary: 0-255)"},
		},
		Returns:  "any! The value that was set",
		Examples: []string{"change next [1 2 3] 99  ; => 99, series becomes [1 99 3]", `change next "hello" "H"  ; => "H", series becomes "Hello" (single char only)`},
		SeeAlso:  []string{"poke", "at", "insert"},
		Tags:     []string{"series", "modification"},
	}))

	registerAndBind("trim", CreateAction("trim", []value.ParamSpec{
		value.NewParamSpec("series", true),
		value.NewRefinementSpec("head", false),
		value.NewRefinementSpec("tail", false),
		value.NewRefinementSpec("auto", false),
		value.NewRefinementSpec("lines", false),
		value.NewRefinementSpec("all", false),
		value.NewRefinementSpec("with", true),
	}, &NativeDoc{
		Category: "Series",
		Summary:  "Removes whitespace or none-like values from series",
		Description: `For strings: removes whitespace (default trims from both ends).
For blocks: removes none values (default trims from both ends).

Note: --auto and --lines refinements are only supported for strings.`,
		Parameters: []ParamDoc{
			{Name: "series", Type: "string! block!", Description: "The series to trim"},
			{Name: "--head", Type: "flag", Description: "Trim from head only", Optional: true},
			{Name: "--tail", Type: "flag", Description: "Trim from tail only", Optional: true},
			{Name: "--auto", Type: "flag", Description: "Auto-indent (strings only)", Optional: true},
			{Name: "--lines", Type: "flag", Description: "Remove line breaks (strings only)", Optional: true},
			{Name: "--all", Type: "flag", Description: "Remove all occurrences", Optional: true},
			{Name: "--with", Type: "any!", Description: "Specify what to remove", Optional: true},
		},
		Returns: "string! block! The trimmed series (modified in place)",
		Examples: []string{
			`trim "  hello  "  ; => "hello"`,
			"trim [none 1 none 2 none]  ; => [1 none 2]",
			"trim --all [none 1 none 2 none]  ; => [1 2]",
		},
		SeeAlso: []string{"clear", "change", "remove"},
		Tags:    []string{"series", "modification"},
	}))

	registerAndBind("insert", CreateAction("insert", []value.ParamSpec{
		value.NewParamSpec("series", true),
		value.NewParamSpec("value", true),
	}, &NativeDoc{
		Category: "Series",
		Summary:  "Inserts a value at the beginning of a series",
		Parameters: []ParamDoc{
			{Name: "series", Type: "block! string! binary!", Description: "The series to insert into"},
			{Name: "value", Type: "any!", Description: "The value to insert"},
		},
		Returns:  "block! string! binary! The modified series",
		Examples: []string{"insert [2 3] 1  ; => [1 2 3]", `insert "ello" "h"  ; => "hello"`, "insert #{ADBE} #{DE}  ; => #{DEADBE}"},
		SeeAlso:  []string{"append", "skip", "take"},
		Tags:     []string{"series", "modification"},
	}))

	registerAndBind("length?", CreateAction("length?", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, &NativeDoc{
		Category: "Series",
		Summary:  "Returns the length of a series",
		Parameters: []ParamDoc{
			{Name: "series", Type: "block! string! binary!", Description: "The series to get length of"},
		},
		Returns:  "integer! The number of elements in the series",
		Examples: []string{"length? [1 2 3]  ; => 3", `length? "hello"  ; => 5`, "length? #{DEADBEEF}  ; => 4"},
		SeeAlso:  []string{"first", "last", "skip", "take"},
		Tags:     []string{"series", "query"},
	}))

	registerAndBind("copy", CreateAction("copy", []value.ParamSpec{
		value.NewParamSpec("series", true),
		value.NewRefinementSpec("part", true),
	}, &NativeDoc{
		Category: "Series",
		Summary:  "Copies a series",
		Description: `Creates a copy of the series.

Without --part: copies all remaining elements from the current index position to the end (implicit remainder).
With --part: copies exactly N elements from the current position. Negative counts raise an OutOfBounds error. Zero count returns an empty series. Counts greater than remaining elements raise an OutOfBounds error (no clamping).

Result index of the copied series is always reset to head. To copy the entire series regardless of current position, use: copy head series.

Difference from take: take clamps oversized counts; copy errors instead.

Error example:
    a: next [1 2 3]           ; moves to position 1
    copy --part 5 a           ; ERROR: only 2 elements remaining`,
		Parameters: []ParamDoc{
			{Name: "series", Type: "block! string! binary!", Description: "The series to copy"},
			{Name: "--part", Type: "integer!", Description: "Copy exactly N remaining elements (0 <= N <= remaining)", Optional: true},
		},
		Returns: "block! string! binary! A copy of the series",
		Examples: []string{
			"copy [1 2 3]  ; => [1 2 3]",
			`copy "hello"  ; => "hello"`,
			"copy #{DEADBEEF}  ; => #{DEADBEEF}",
			"copy --part 2 [1 2 3 4]  ; => [1 2]",
			"copy --part 0 [1 2 3]  ; => []",
			"a: next next [1 2 3 4] copy a  ; => [3 4]",
		},
		SeeAlso: []string{"append", "insert", "take"},
		Tags:    []string{"series"},
	}))

	registerAndBind("find", CreateAction("find", []value.ParamSpec{
		value.NewParamSpec("series", true),
		value.NewParamSpec("value", true),
		value.NewRefinementSpec("last", false),
	}, &NativeDoc{
		Category: "Series",
		Summary:  "Finds a value in a series",
		Parameters: []ParamDoc{
			{Name: "series", Type: "block! string! binary!", Description: "The series to search"},
			{Name: "value", Type: "any!", Description: "The value to find"},
			{Name: "--last", Type: "", Description: "Find last occurrence instead of first", Optional: true},
		},
		Returns:  "integer! 1-based index or none",
		Examples: []string{"find [1 2 3] 2  ; => 2", `find "hello" "l"  ; => 3`, "find #{DEADBEEF} 190  ; => 3"},
		SeeAlso:  []string{"first", "last"},
		Tags:     []string{"series", "search"},
	}))

	registerAndBind("remove", CreateAction("remove", []value.ParamSpec{
		value.NewParamSpec("series", true),
		value.NewRefinementSpec("part", true),
	}, &NativeDoc{
		Category: "Series",
		Summary:  "Removes elements from a series",
		Description: `Removes elements from the series starting at the current position.
With --part, removes the specified number of elements. Negative counts raise an OutOfBounds error.
Zero count is a no-op. Oversized counts (where index+count exceeds length) raise an OutOfBounds error.`,
		Parameters: []ParamDoc{
			{Name: "series", Type: "block! string! binary!", Description: "The series to remove from"},
			{Name: "--part", Type: "integer!", Description: "Remove n elements from current position (must be >= 0, index+count <= length)", Optional: true},
		},
		Returns: "block! string! binary! The modified series",
		Examples: []string{
			"remove [1 2 3]  ; => [2 3]",
			"remove --part 2 [1 2 3]  ; => [3]",
			`remove "hello"  ; => "ello"`,
			"remove #{DEADBEEF}  ; => #{ADBE}",
			"remove --part 0 [1 2 3]  ; => [1 2 3] (no-op)",
		},
		SeeAlso: []string{"append", "insert", "clear"},
		Tags:    []string{"series", "modification"},
	}))

	registerAndBind("skip", CreateAction("skip", []value.ParamSpec{
		value.NewParamSpec("series", true),
		value.NewParamSpec("count", true),
	}, &NativeDoc{
		Category: "Series",
		Summary:  "Skips n elements in a series",
		Parameters: []ParamDoc{
			{Name: "series", Type: "block! string! binary!", Description: "The series to skip from"},
			{Name: "count", Type: "integer!", Description: "Number of elements to skip"},
		},
		Returns:  "block! string! binary! Series with index advanced by count",
		Examples: []string{"skip [1 2 3 4] 2  ; => [1 2 3 4] (index at 3)", `skip "hello" 2  ; => "hello" (index at 3)`, "skip #{DEADBEEF} 2  ; => #{DEADBEEF} (index at 3)"},
		SeeAlso:  []string{"take", "first", "last"},
		Tags:     []string{"series"},
	}))

	registerAndBind("next", CreateAction("next", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, &NativeDoc{
		Category: "Series",
		Summary:  "Returns a series reference advanced by one position",
		Parameters: []ParamDoc{
			{Name: "series", Type: "block! string! binary!", Description: "The series to advance"},
		},
		Returns:  "block! string! binary! New series reference at next position",
		Examples: []string{"next [1 2 3]  ; => [1 2 3] (index at 2)", `next "hello"  ; => "hello" (index at 2)`, "next #{DEADBEEF}  ; => #{DEADBEEF} (index at 2)"},
		SeeAlso:  []string{"skip", "back", "head", "tail"},
		Tags:     []string{"series", "navigation"},
	}))

	registerAndBind("back", CreateAction("back", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, &NativeDoc{
		Category: "Series",
		Summary:  "Returns a series reference moved backward by one position",
		Parameters: []ParamDoc{
			{Name: "series", Type: "block! string! binary!", Description: "The series to move backward"},
		},
		Returns:  "block! string! binary! New series reference at previous position",
		Examples: []string{"back next [1 2 3]  ; => [1 2 3] (index at 1)", `back next "hello"  ; => "hello" (index at 1)`, "back next #{DEADBEEF}  ; => #{DEADBEEF} (index at 1)"},
		SeeAlso:  []string{"next", "head", "tail", "skip"},
		Tags:     []string{"series", "navigation"},
	}))

	registerAndBind("head", CreateAction("head", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, &NativeDoc{
		Category: "Series",
		Summary:  "Returns a series reference positioned at the head (position 0)",
		Parameters: []ParamDoc{
			{Name: "series", Type: "block! string! binary!", Description: "The series to position at head"},
		},
		Returns:  "block! string! binary! New series reference at head position",
		Examples: []string{"head [1 2 3]  ; => [1 2 3] (index at 1)", `head "hello"  ; => "hello" (index at 1)`, "head #{DEADBEEF}  ; => #{DEADBEEF} (index at 1)"},
		SeeAlso:  []string{"tail", "next", "back"},
		Tags:     []string{"series", "navigation"},
	}))

	registerAndBind("tail", CreateAction("tail", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, &NativeDoc{
		Category: "Series",
		Summary:  "Returns a series positioned at the tail",
		Parameters: []ParamDoc{
			{Name: "series", Type: "block! string! binary!", Description: "The series to position at tail"},
		},
		Returns:  "block! string! binary! New series reference at tail position",
		Examples: []string{"tail [1 2 3 4]  ; => [1 2 3 4] (index at 4)", `tail "hello"  ; => "hello" (index at 5)`, "tail #{DEADBEEF}  ; => #{DEADBEEF} (index at 4)"},
		SeeAlso:  []string{"head", "next", "back"},
		Tags:     []string{"series", "navigation"},
	}))

	registerAndBind("empty?", CreateAction("empty?", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, &NativeDoc{
		Category: "Series",
		Summary:  "Returns true when the current series index is at or beyond the tail (no remaining elements)",
		Parameters: []ParamDoc{
			{Name: "series", Type: "block! string! binary!", Description: "The series to check"},
		},
		Returns:  "logic! true if no remaining elements from current position, false otherwise",
		Examples: []string{"empty? tail [1 2 3]  ; => true", "empty? back tail [1 2 3]  ; => false", "empty? []  ; => true", "empty? [1 2 3]  ; => false", `empty? ""  ; => true`, `empty? "hello"  ; => false`},
		SeeAlso:  []string{"length?", "head?", "tail?"},
		Tags:     []string{"series", "query"},
	}))

	registerAndBind("head?", CreateAction("head?", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, &NativeDoc{
		Category: "Series",
		Summary:  "Returns true if the series index is at position 0 (head)",
		Parameters: []ParamDoc{
			{Name: "series", Type: "block! string! binary!", Description: "The series to check"},
		},
		Returns:  "logic! true if series is at head position, false otherwise",
		Examples: []string{"head? [1 2 3]  ; => true", "head? next [1 2 3]  ; => false", `head? "hello"  ; => true`, `head? next "hello"  ; => false`},
		SeeAlso:  []string{"tail?", "index?", "head"},
		Tags:     []string{"series", "query"},
	}))

	registerAndBind("tail?", CreateAction("tail?", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, &NativeDoc{
		Category: "Series",
		Summary:  "Returns true if the series index is at the end (index == length)",
		Parameters: []ParamDoc{
			{Name: "series", Type: "block! string! binary!", Description: "The series to check"},
		},
		Returns:  "logic! true if series is at tail position, false otherwise",
		Examples: []string{"tail? [1 2 3]  ; => false", "tail? tail [1 2 3]  ; => true", `tail? "hello"  ; => false`, `tail? tail "hello"  ; => true`},
		SeeAlso:  []string{"head?", "index?", "tail"},
		Tags:     []string{"series", "query"},
	}))

	registerAndBind("index?", CreateAction("index?", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, &NativeDoc{
		Category: "Series",
		Summary:  "Returns the current index position of a series (1-based)",
		Parameters: []ParamDoc{
			{Name: "series", Type: "block! string! binary!", Description: "The series to get index from"},
		},
		Returns:  "integer! The current index position (1-based)",
		Examples: []string{"index? [1 2 3]  ; => 1", `index? next "hello"  ; => 2`, "index? skip #{DEADBEEF} 2  ; => 3"},
		SeeAlso:  []string{"head", "next", "skip"},
		Tags:     []string{"series", "query"},
	}))

	registerAndBind("take", CreateAction("take", []value.ParamSpec{
		value.NewParamSpec("series", true),
		value.NewParamSpec("count", true),
	}, &NativeDoc{
		Category: "Series",
		Summary:  "Takes n elements from a series",
		Description: `Returns a new series containing the first n elements from the current position.
Negative counts raise an OutOfBounds error. Zero count returns an empty series.
Oversized counts are clamped to the remaining elements from current position.`,
		Parameters: []ParamDoc{
			{Name: "series", Type: "block! string! binary!", Description: "The series to take from"},
			{Name: "count", Type: "integer!", Description: "Number of elements to take (must be >= 0)"},
		},
		Returns: "block! string! binary! Series containing first count elements",
		Examples: []string{
			"take [1 2 3 4] 2  ; => [1 2]",
			`take "hello" 2  ; => "he"`,
			"take #{DEADBEEF} 2  ; => #{DEAD}",
			"take [1 2 3] 0  ; => []",
			"take [1 2 3] 10  ; => [1 2 3] (clamped)",
		},
		SeeAlso: []string{"skip", "first", "last", "copy"},
		Tags:    []string{"series"},
	}))

	registerAndBind("sort", CreateAction("sort", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, &NativeDoc{
		Category: "Series",
		Summary:  "Sorts a series in place",
		Parameters: []ParamDoc{
			{Name: "series", Type: "block! string! binary!", Description: "The series to sort"},
		},
		Returns: "block! string! binary! The sorted series",
		Examples: []string{
			"sort [3 1 2]  ; => [1 2 3]",
			`sort "cba"  ; => "abc"`,
			"sort #{030201}  ; => #{010203}",
		},
		SeeAlso: []string{"reverse"},
		Tags:    []string{"series", "sorting"},
	}))

	registerAndBind("reverse", CreateAction("reverse", []value.ParamSpec{
		value.NewParamSpec("series", true),
	}, &NativeDoc{
		Category: "Series",
		Summary:  "Reverses a series in place",
		Parameters: []ParamDoc{
			{Name: "series", Type: "block! string! binary!", Description: "The series to reverse"},
		},
		Returns:  "block! string! binary! The reversed series",
		Examples: []string{"reverse [1 2 3]  ; => [3 2 1]", `reverse "hello"  ; => "olleh"`, "reverse #{DEADBEEF}  ; => #{EFBEADDE}"},
		SeeAlso:  []string{"sort"},
		Tags:     []string{"series"},
	}))

	registerAndBind("split", value.NewNativeFunction(
		"split",
		[]value.ParamSpec{
			value.NewParamSpec("string", true),
			value.NewParamSpec("delimiter", true),
		},
		StringSplit,
		false,
		&NativeDoc{
			Category: "Series",
			Summary:  "Splits a string by delimiter into a block of strings",
			Description: `Splits a string by a delimiter and returns a block containing the resulting substrings.
Empty delimiter is not allowed and will raise an error. Consecutive delimiters create empty strings in the result.`,
			Parameters: []ParamDoc{
				{Name: "string", Type: "string!", Description: "The string to split"},
				{Name: "delimiter", Type: "string!", Description: "The delimiter to split by (cannot be empty)"},
			},
			Returns:  "[block!] Block containing the split string parts",
			Examples: []string{`split "hello world" " "  ; => ["hello" "world"]`, `split "a,b,c" ","  ; => ["a" "b" "c"]`, `split "a,,b" ","  ; => ["a" "" "b"]`},
			SeeAlso:  []string{"join", "form", "mold"},
			Tags:     []string{"series", "string", "split"},
		}))

	registerAndBind("codepoints-of", value.NewNativeFunction(
		"codepoints-of",
		[]value.ParamSpec{
			value.NewParamSpec("string", true),
		},
		CodepointsOf,
		false,
		&NativeDoc{
			Category: "Series",
			Summary:  "Converts a string to a block of Unicode code point integers",
			Description: `Returns a block containing the Unicode code point values (as integers) for each character in the string.
The code points are in the same order as the characters in the string.`,
			Parameters: []ParamDoc{
				{Name: "string", Type: "string!", Description: "The string to convert to code points"},
			},
			Returns:  "[block!] Block of integers representing Unicode code points",
			Examples: []string{`codepoints-of "ABC"  ; => [65 66 67]`, `codepoints-of "🚀"  ; => [128640]`, `codepoints-of ""  ; => []`},
			SeeAlso:  []string{"string-from-codepoints", "codepoint-at"},
			Tags:     []string{"series", "string", "unicode"},
		}))

	registerAndBind("codepoint-at", value.NewNativeFunction(
		"codepoint-at",
		[]value.ParamSpec{
			value.NewParamSpec("string", true),
			value.NewParamSpec("index", true),
			value.NewRefinementSpec("default", true),
		},
		CodepointAt,
		false,
		&NativeDoc{
			Category: "Series",
			Summary:  "Returns the Unicode code point at a specific position in a string",
			Description: `Returns the Unicode code point (as an integer) at the specified 1-based index in the string.
If the index is out of bounds, returns none unless the /default refinement is provided.`,
			Parameters: []ParamDoc{
				{Name: "string", Type: "string!", Description: "The string to get code point from"},
				{Name: "index", Type: "integer!", Description: "1-based index of the character"},
				{Name: "--default", Type: "any!", Description: "Value to return if index is out of bounds", Optional: true},
			},
			Returns:  "integer! The Unicode code point, or none/default if out of bounds",
			Examples: []string{`codepoint-at "ABC" 1  ; => 65`, `codepoint-at "🚀" 1  ; => 128640`, `codepoint-at "ABC" 10  ; => none`, `codepoint-at "ABC" 10 --default 0  ; => 0`},
			SeeAlso:  []string{"codepoints-of", "at", "pick"},
			Tags:     []string{"series", "string", "unicode"},
		}))

	registerAndBind("string-from-codepoints", value.NewNativeFunction(
		"string-from-codepoints",
		[]value.ParamSpec{
			value.NewParamSpec("codepoints", true),
		},
		StringFromCodepoints,
		false,
		&NativeDoc{
			Category: "Series",
			Summary:  "Creates a string from a block of Unicode code point integers",
			Description: `Creates a string from a block of integers representing Unicode code points.
All code points must be valid (0-0x10FFFF) and not in the surrogate range (0xD800-0xDFFF).`,
			Parameters: []ParamDoc{
				{Name: "codepoints", Type: "block!", Description: "Block of integers representing Unicode code points"},
			},
			Returns:  "string! The constructed string",
			Examples: []string{`string-from-codepoints [65 66 67]  ; => "ABC"`, `string-from-codepoints [128640]  ; => "🚀"`, `string-from-codepoints []  ; => ""`},
			SeeAlso:  []string{"codepoints-of", "codepoint-at"},
			Tags:     []string{"series", "string", "unicode"},
		}))

	registerAndBind("intersect", CreateAction("intersect", []value.ParamSpec{
		value.NewParamSpec("s1", true),
		value.NewParamSpec("s2", true),
	}, &NativeDoc{
		Category: "Series",
		Summary:  "Returns the intersection of two series",
		Description: `Returns a new series containing unique elements that exist in both series.
The order of elements in the result follows the order they appear in the first series.
Duplicates are removed from the result.`,
		Parameters: []ParamDoc{
			{Name: "s1", Type: "block! string! binary!", Description: "First series"},
			{Name: "s2", Type: "block! string! binary!", Description: "Second series"},
		},
		Returns: "block! string! binary! Series containing unique common elements",
		Examples: []string{
			"intersect [1 2 3] [2 3 4]  ; => [2 3]",
			`intersect "hello" "world"  ; => "lo"`,
			"intersect #{010203} #{020304}  ; => #{0203}",
		},
		SeeAlso: []string{"union", "difference"},
		Tags:    []string{"series", "set"},
	}))

	registerAndBind("difference", CreateAction("difference", []value.ParamSpec{
		value.NewParamSpec("s1", true),
		value.NewParamSpec("s2", true),
	}, &NativeDoc{
		Category: "Series",
		Summary:  "Returns the difference of two series",
		Description: `Returns a new series containing unique elements from the first series that do not exist in the second series.
The order of elements in the result follows the order they appear in the first series.
Duplicates are removed from the result.`,
		Parameters: []ParamDoc{
			{Name: "s1", Type: "block! string! binary!", Description: "First series"},
			{Name: "s2", Type: "block! string! binary!", Description: "Second series"},
		},
		Returns: "block! string! binary! Series containing unique elements from first series not in second",
		Examples: []string{
			"difference [1 2 3] [2 3 4]  ; => [1]",
			`difference "hello" "world"  ; => "he"`,
			"difference #{010203} #{020304}  ; => #{01}",
		},
		SeeAlso: []string{"union", "intersect"},
		Tags:    []string{"series", "set"},
	}))

	registerAndBind("union", CreateAction("union", []value.ParamSpec{
		value.NewParamSpec("s1", true),
		value.NewParamSpec("s2", true),
	}, &NativeDoc{
		Category: "Series",
		Summary:  "Returns the union of two series",
		Description: `Returns a new series containing all unique elements from both series.
The order of elements follows the order they appear in the first series, then the second series.
Duplicates are removed from the result.`,
		Parameters: []ParamDoc{
			{Name: "s1", Type: "block! string! binary!", Description: "First series"},
			{Name: "s2", Type: "block! string! binary!", Description: "Second series"},
		},
		Returns: "block! string! binary! Series containing all unique elements from both series",
		Examples: []string{
			"union [1 2 3] [2 3 4]  ; => [1 2 3 4]",
			`union "hello" "world"  ; => "helowrd"`,
			"union #{010203} #{020304}  ; => #{01020304}",
		},
		SeeAlso: []string{"intersect", "difference"},
		Tags:    []string{"series", "set"},
	}))
}
