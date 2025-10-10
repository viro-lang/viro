package native

import (
	"testing"
)

// TestAllNativesDocumented verifies that all native functions have documentation.
func TestAllNativesDocumented(t *testing.T) {
	undocumented := []string{}
	for name, info := range Registry {
		if info.Doc == nil || !info.Doc.HasDoc() {
			undocumented = append(undocumented, name)
		}
	}

	if len(undocumented) > 0 {
		t.Errorf("Found %d undocumented native functions: %v", len(undocumented), undocumented)
	} else {
		t.Logf("All %d native functions are documented ✓", len(Registry))
	}
}

// TestDocumentationValid validates all documentation entries.
func TestDocumentationValid(t *testing.T) {
	errors := ValidateRegistry(Registry)
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
	for name, info := range Registry {
		if info.Doc == nil {
			continue // Skip undocumented (caught by other test)
		}

		doc := info.Doc

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

		// Check parameter count matches arity
		// Arity represents the minimum required parameters.
		// Optional parameters (refinements) are documented but don't count toward arity.
		// Special case: ? function has variable arity (0 or 1) with 1 optional param
		if name == "?" {
			if len(doc.Parameters) != 1 || !doc.Parameters[0].Optional {
				t.Errorf("%s: should have exactly 1 optional parameter", name)
			}
		} else {
			// Count required (non-optional) parameters
			requiredCount := 0
			for _, param := range doc.Parameters {
				if !param.Optional {
					requiredCount++
				}
			}
			if requiredCount != info.Arity {
				t.Errorf("%s: required parameter count (%d) doesn't match arity (%d)",
					name, requiredCount, info.Arity)
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
	for name, info := range Registry {
		if info.Doc != nil && !validCategories[info.Doc.Category] {
			invalidCategories = append(invalidCategories, name+":"+info.Doc.Category)
		}
	}

	if len(invalidCategories) > 0 {
		t.Errorf("Found functions with invalid categories: %v", invalidCategories)
	}
}

// TestDocumentationExamples checks that examples are present and non-empty.
func TestDocumentationExamples(t *testing.T) {
	for name, info := range Registry {
		if info.Doc == nil {
			continue
		}

		if len(info.Doc.Examples) == 0 {
			t.Errorf("%s: no examples provided", name)
			continue
		}

		for i, example := range info.Doc.Examples {
			if example == "" {
				t.Errorf("%s: example %d is empty", name, i)
			}
		}
	}
}

// TestDocumentationStats prints statistics about the documentation.
func TestDocumentationStats(t *testing.T) {
	documented, total := CountDocumented(Registry)
	categories := GetCategories(Registry)

	t.Logf("Documentation Statistics:")
	t.Logf("  Total functions: %d", total)
	t.Logf("  Documented: %d", documented)
	t.Logf("  Undocumented: %d", total-documented)
	t.Logf("  Coverage: %.1f%%", float64(documented)/float64(total)*100)
	t.Logf("  Categories: %d (%v)", len(categories), categories)

	// Count functions per category
	categoryCount := make(map[string]int)
	for _, info := range Registry {
		if info.Doc != nil {
			categoryCount[info.Doc.Category]++
		}
	}

	t.Logf("\nFunctions per category:")
	for _, cat := range categories {
		t.Logf("  %s: %d functions", cat, categoryCount[cat])
	}
}
