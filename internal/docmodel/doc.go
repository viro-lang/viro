package docmodel

// ParamDoc describes a parameter in function documentation.
type ParamDoc struct {
	Name        string
	Type        string
	Description string
	Optional    bool
}

// FuncDoc is neutralna reprezentacja dokumentacji funkcji.
type FuncDoc struct {
	Category    string
	Summary     string
	Description string
	Parameters  []ParamDoc
	Returns     string
	Examples    []string
	SeeAlso     []string
	Tags        []string
}

// NewFuncDoc convenience helper
func NewFuncDoc(category, summary, description, returns string, params []ParamDoc, examples, seeAlso, tags []string) *FuncDoc {
	return &FuncDoc{Category: category, Summary: summary, Description: description, Parameters: params, Returns: returns, Examples: examples, SeeAlso: seeAlso, Tags: tags}
}

// Validate checks if the documentation is complete and well-formed.
// Returns an error message if validation fails, empty string if valid.
func (d *FuncDoc) Validate(funcName string) string {
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
func (d *FuncDoc) HasDoc() bool {
	return d != nil && d.Summary != ""
}
