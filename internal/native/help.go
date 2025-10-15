package native

import (
	"fmt"
	"strings"

	"github.com/marcin-radoszewski/viro/internal/core"
	"github.com/marcin-radoszewski/viro/internal/value"
)

// buildRegistryFromFrame builds a registry map from all function values in a frame.
// This is used by the help system to access native functions from the root frame.
func buildRegistryFromFrame(f core.Frame) map[string]*value.FunctionValue {
	registry := make(map[string]*value.FunctionValue)
	for _, binding := range f.GetAll() {
		if binding.Value.GetType() == value.TypeFunction {
			if fn, ok := value.AsFunction(binding.Value); ok {
				registry[binding.Symbol] = fn
			}
		}
	}
	return registry
}

// Help displays documentation for a function or category.
// USAGE: ? [word]  (optional argument)
// Returns: none (prints to output)
//
// This function uses FuncEval to support variable arity (0 or 1 arguments).
// With no args: shows category list
// With one arg: shows function help or category listing
//
// NOTE: This function does NOT evaluate its argument, so it accepts word literals.
func Help(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	// Handle 0 or 1 arguments
	if len(args) > 1 {
		return value.NoneVal(), arityError("?", 1, len(args))
	}

	// Build registry from root frame
	rootFrame := eval.GetFrameByIndex(0)
	registry := buildRegistryFromFrame(rootFrame)

	// No arguments - show category list
	if len(args) == 0 {
		fmt.Print(FormatCategoryList(registry))
		return value.NoneVal(), nil
	}

	// One argument - get the word/string directly (not evaluated)
	arg := args[0]

	// Get the word name to look up
	var lookupName string
	if sym, ok := value.AsWord(arg); ok {
		lookupName = sym
	} else if str, ok := value.AsString(arg); ok {
		lookupName = str.String()
	} else {
		return value.NoneVal(), typeError("?", "word or string", arg)
	}

	// Try to find the function in the registry
	if fn, ok := registry[lookupName]; ok {
		// Found a function - show detailed help
		if fn.Doc != nil {
			fmt.Print(FormatHelp(lookupName, fn.Doc))
		} else {
			fmt.Printf("\n%s: Native function (no documentation available)\n\n", lookupName)
		}
		return value.NoneVal(), nil
	}

	// Not a function - maybe it's a category?
	output := FormatFunctionList(lookupName, registry)
	if !strings.Contains(output, "not found") {
		// It's a valid category
		fmt.Print(output)
		return value.NoneVal(), nil
	}

	// Not found - suggest similar functions
	similar := FindSimilar(lookupName, registry, 5)
	if len(similar) > 0 {
		fmt.Printf("\n'%s' not found.\n", lookupName)
		fmt.Printf("Did you mean: %s?\n\n", strings.Join(similar, ", "))
	} else {
		fmt.Printf("\n'%s' not found. Type '?' to see available functions.\n\n", lookupName)
	}

	return value.NoneVal(), nil
}

// Words lists all available function names.
// USAGE: words
// Returns: block of words (function names)
func Words(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 0 {
		return value.NoneVal(), arityError("words", 0, len(args))
	}

	// Build registry from root frame
	rootFrame := eval.GetFrameByIndex(0)
	registry := buildRegistryFromFrame(rootFrame)

	names := make([]core.Value, 0, len(registry))
	for name := range registry {
		names = append(names, value.WordVal(name))
	}

	return value.BlockVal(names), nil
}
