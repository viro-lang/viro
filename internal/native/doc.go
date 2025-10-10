package native

import "fmt"

// NativeDoc contains documentation metadata for a native function.
// It provides comprehensive information for the help system including
// usage, parameters, examples, and cross-references.
type NativeDoc struct {
	// Category groups related functions (e.g., "Math", "Control", "Series", "Data", "Function", "I/O", "Ports", "Objects")
	Category string

	// Summary is a one-line description of what the function does
	Summary string

	// Description is a detailed multi-line explanation of the function's behavior,
	// edge cases, and usage notes
	Description string

	// Parameters documents each parameter the function accepts
	Parameters []ParamDoc

	// Returns describes what the function returns
	Returns string

	// Examples contains usage examples with expected results
	// Each example should be a valid viro expression or script snippet
	Examples []string

	// SeeAlso lists related function names for cross-referencing
	SeeAlso []string

	// Tags provides searchable keywords for function discovery
	Tags []string
}

// ParamDoc documents a single parameter of a native function.
type ParamDoc struct {
	// Name is the parameter name as used in documentation
	Name string

	// Type describes expected type(s) - e.g., "integer!", "block!", "any-type!"
	// Multiple types can be listed: "integer! decimal!"
	Type string

	// Description explains the parameter's purpose and usage
	Description string

	// Optional indicates whether this parameter can be omitted
	Optional bool
}

// Validate checks if the documentation is complete and well-formed.
// Returns an error message if validation fails, empty string if valid.
func (d *NativeDoc) Validate(funcName string) string {
	if d.Category == "" {
		return funcName + ": missing category"
	}
	if d.Summary == "" {
		return funcName + ": missing summary"
	}
	if d.Description == "" {
		return funcName + ": missing description"
	}
	if d.Returns == "" {
		return funcName + ": missing returns documentation"
	}
	if len(d.Examples) == 0 {
		return funcName + ": missing examples"
	}

	// Validate each parameter
	for i, param := range d.Parameters {
		if param.Name == "" {
			return funcName + ": parameter " + string(rune(i)) + " missing name"
		}
		if param.Type == "" {
			return funcName + ": parameter '" + param.Name + "' missing type"
		}
		if param.Description == "" {
			return funcName + ": parameter '" + param.Name + "' missing description"
		}
	}

	return ""
}

// HasDoc returns true if the documentation is present and non-empty.
func (d *NativeDoc) HasDoc() bool {
	return d != nil && d.Summary != ""
}

// GetCategories returns a list of all unique categories from a registry.
func GetCategories(registry map[string]*NativeInfo) []string {
	categorySet := make(map[string]bool)
	for _, info := range registry {
		if info.Doc != nil && info.Doc.Category != "" {
			categorySet[info.Doc.Category] = true
		}
	}

	categories := make([]string, 0, len(categorySet))
	for cat := range categorySet {
		categories = append(categories, cat)
	}
	return categories
}

// GetFunctionsInCategory returns all function names in a given category.
func GetFunctionsInCategory(registry map[string]*NativeInfo, category string) []string {
	functions := make([]string, 0)
	for name, info := range registry {
		if info.Doc != nil && info.Doc.Category == category {
			functions = append(functions, name)
		}
	}
	return functions
}

// CountDocumented returns the number of documented vs total native functions.
func CountDocumented(registry map[string]*NativeInfo) (documented, total int) {
	total = len(registry)
	for _, info := range registry {
		if info.Doc != nil && info.Doc.HasDoc() {
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
func ValidateRegistry(registry map[string]*NativeInfo) []string {
	errors := make([]string, 0)
	for name, info := range registry {
		if info.Doc != nil {
			if err := info.Doc.Validate(name); err != "" {
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
