package native

import (
	"fmt"

	"github.com/marcin-radoszewski/viro/internal/value"
)

// ...istniejący kod...

// Validate checks if the documentation is complete and well-formed.
// Returns an error message if validation fails, empty string if valid.
// ...istniejący kod...

// GetCategories returns a list of all unique categories from a registry.
func GetCategories(registry map[string]*value.FunctionValue) []string {
	categorySet := make(map[string]bool)
	for _, fn := range registry {
		if fn.Doc != nil && fn.Doc.Category != "" {
			categorySet[fn.Doc.Category] = true
		}
	}

	categories := make([]string, 0, len(categorySet))
	for cat := range categorySet {
		categories = append(categories, cat)
	}
	return categories
}

// GetFunctionsInCategory returns all function names in a given category.
func GetFunctionsInCategory(registry map[string]*value.FunctionValue, category string) []string {
	functions := make([]string, 0)
	for name, fn := range registry {
		if fn.Doc != nil && fn.Doc.Category == category {
			functions = append(functions, name)
		}
	}
	return functions
}

// CountDocumented returns the number of documented vs total native functions.
func CountDocumented(registry map[string]*value.FunctionValue) (documented, total int) {
	total = len(registry)
	for _, fn := range registry {
		if fn.Doc != nil && fn.Doc.HasDoc() {
			documented++
		}
	}
	return documented, total
}

// NewDocTemplate creates a documentation template for a new native function.
// This is a helper for developers adding new natives.
func NewDocTemplate(funcName, category string, paramCount int) *NativeDoc {
	params := make([]ParamDoc, paramCount)
	for i := 0; i < paramCount; i++ {
		params[i] = ParamDoc{
			Name:        fmt.Sprintf("param%d", i+1),
			Type:        "any-type!",
			Description: "TODO: describe this parameter",
			Optional:    false,
		}
	}

	return &NativeDoc{
		Category:    category,
		Summary:     "TODO: one-line summary",
		Description: "TODO: detailed description",
		Parameters:  params,
		Returns:     "[any-type!] TODO: describe return value",
		Examples: []string{
			funcName + " example-args  ; => expected-result",
		},
		SeeAlso: []string{},
		Tags:    []string{},
	}
}

// ValidateRegistry checks all documentation in the registry and returns
// a list of validation errors. Returns empty slice if all docs are valid.
func ValidateRegistry(registry map[string]*value.FunctionValue) []string {
	errors := make([]string, 0)
	for name, fn := range registry {
		if fn.Doc != nil {
			if err := fn.Doc.Validate(name); err != "" {
				errors = append(errors, err)
			}
		}
	}
	return errors
}

// DocTemplate provides a string template for developers to copy when documenting natives.
const DocTemplate = `
// Documentation template for native function
Doc: &NativeDoc{
	Category: "Category",  // Math, Control, Series, Data, Function, I/O, Ports, Objects
	Summary: "One-line description of what this function does",
	Description: ` + "`" + `
Detailed explanation of the function including:
- What it does
- When to use it
- Important behavior notes
- Edge cases and limitations
` + "`" + `,
	Parameters: []ParamDoc{
		{
			Name: "param1",
			Type: "type!",  // e.g., "integer!", "block!", "any-type!"
			Description: "Description of the parameter",
			Optional: false,
		},
	},
	Returns: "[return-type!] Description of return value",
	Examples: []string{
		"function-name arg1  ; => result",
		"x: [1 2 3]" + "\n" + "function-name x  ; => modified-result",
	},
	SeeAlso: []string{"related-function-1", "related-function-2"},
	Tags: []string{"tag1", "tag2"},
},
`
