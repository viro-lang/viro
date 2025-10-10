package native

import (
	"testing"
)

// TestAllNativesDocumented verifies that all native functions have documentation.
func TestAllNativesDocumented(t *testing.T) {
	undocumented := []string{}
	for name, fn := range FunctionRegistry {
		if fn.Doc == nil || !fn.Doc.HasDoc() {
			undocumented = append(undocumented, name)
		}
	}

	if len(undocumented) > 0 {
		t.Errorf("Found %d undocumented native functions: %v", len(undocumented), undocumented)
	} else {
		t.Logf("All %d native functions are documented ✓", len(FunctionRegistry))
	}
}

// TestDocumentationValid validates all documentation entries.
func TestDocumentationValid(t *testing.T) {
	errors := ValidateRegistry(FunctionRegistry)
	if len(errors) > 0 {
		t.Errorf("Found %d documentation validation errors:", len(errors))
		for _, err := range errors {
			t.Errorf("  - %s", err)
		}
	} else {
		t.Logf("All documentation entries are valid ✓")
	}
}

// TestDocumentationCompleteness checks that each documented function has all required fields.
func TestDocumentationCompleteness(t *testing.T) {
	for name, fn := range FunctionRegistry {
		if fn.Doc == nil {
			continue // Skip undocumented (caught by other test)
		}

		doc := fn.Doc

		// Check required fields
		if doc.Category == "" {
			t.Errorf("%s: missing category", name)
		}
		if doc.Summary == "" {
			t.Errorf("%s: missing summary", name)
		}
		if doc.Description == "" {
			t.Errorf("%s: missing description", name)
		}
		if doc.Returns == "" {
			t.Errorf("%s: missing returns documentation", name)
		}
		if len(doc.Examples) == 0 {
			t.Errorf("%s: missing examples", name)
		}

		// Check parameter count matches function's parameter spec
		// Special case: ? function has variable arity (0 or 1) with 1 optional param
		if name == "?" {
			if len(doc.Parameters) != 1 || !doc.Parameters[0].Optional {
				t.Errorf("%s: should have exactly 1 optional parameter", name)
			}
		} else {
			// Count required (non-optional) parameters in documentation
			requiredCount := 0
			for _, param := range doc.Parameters {
				if !param.Optional {
					requiredCount++
				}
			}
			// Count required (non-optional, non-refinement) parameters from function spec
			funcRequiredCount := 0
			for _, param := range fn.Params {
				if !param.Optional && !param.Refinement {
					funcRequiredCount++
				}
			}
			if requiredCount != funcRequiredCount {
				t.Errorf("%s: required parameter count in doc (%d) doesn't match function spec (%d)",
					name, requiredCount, funcRequiredCount)
			}
		}

		// Check each parameter is complete
		for i, param := range doc.Parameters {
			if param.Name == "" {
				t.Errorf("%s: parameter %d missing name", name, i)
			}
			if param.Type == "" {
				t.Errorf("%s: parameter '%s' missing type", name, param.Name)
			}
			if param.Description == "" {
				t.Errorf("%s: parameter '%s' missing description", name, param.Name)
			}
		}
	}
} // TestDocumentationCategories verifies all functions are in valid categories.
func TestDocumentationCategories(t *testing.T) {
	validCategories := map[string]bool{
		"Math":     true,
		"Control":  true,
		"Series":   true,
		"Data":     true,
		"Function": true,
		"I/O":      true,
		"Ports":    true,
		"Objects":  true,
		"Help":     true,
	}

	invalidCategories := []string{}
	for name, fn := range FunctionRegistry {
		if fn.Doc != nil && !validCategories[fn.Doc.Category] {
			invalidCategories = append(invalidCategories, name+":"+fn.Doc.Category)
		}
	}

	if len(invalidCategories) > 0 {
		t.Errorf("Found functions with invalid categories: %v", invalidCategories)
	}
}

// TestDocumentationExamples checks that examples are present and non-empty.
func TestDocumentationExamples(t *testing.T) {
	for name, fn := range FunctionRegistry {
		if fn.Doc == nil {
			continue
		}

		if len(fn.Doc.Examples) == 0 {
			t.Errorf("%s: no examples provided", name)
			continue
		}

		for i, example := range fn.Doc.Examples {
			if example == "" {
				t.Errorf("%s: example %d is empty", name, i)
			}
		}
	}
}

// TestDocumentationStats prints statistics about the documentation.
func TestDocumentationStats(t *testing.T) {
	documented, total := CountDocumented(FunctionRegistry)
	categories := GetCategories(FunctionRegistry)

	t.Logf("Documentation Statistics:")
	t.Logf("  Total functions: %d", total)
	t.Logf("  Documented: %d", documented)
	t.Logf("  Undocumented: %d", total-documented)
	t.Logf("  Coverage: %.1f%%", float64(documented)/float64(total)*100)
	t.Logf("  Categories: %d (%v)", len(categories), categories)

	// Count functions per category
	categoryCount := make(map[string]int)
	for _, fn := range FunctionRegistry {
		if fn.Doc != nil {
			categoryCount[fn.Doc.Category]++
		}
	}

	t.Logf("\nFunctions per category:")
	for _, cat := range categories {
		t.Logf("  %s: %d functions", cat, categoryCount[cat])
	}
}
