package native

import (
	"fmt"
	"sort"
	"strings"
)

// FormatHelp formats NativeDoc for display in the REPL.
// Returns a formatted string with sections for usage, parameters, returns, description, examples, etc.
func FormatHelp(funcName string, doc *NativeDoc) string {
	if doc == nil {
		return fmt.Sprintf("%s: No documentation available\n", funcName)
	}

	var b strings.Builder

	// Header
	b.WriteString(fmt.Sprintf("\n%s - %s\n", strings.ToUpper(funcName), doc.Category))
	b.WriteString(strings.Repeat("=", len(funcName)+len(doc.Category)+3))
	b.WriteString("\n\n")

	// Summary
	b.WriteString(doc.Summary)
	b.WriteString("\n\n")

	// Usage section
	b.WriteString("USAGE:\n")
	b.WriteString("    ")
	b.WriteString(funcName)
	for _, param := range doc.Parameters {
		b.WriteString(" ")
		b.WriteString(param.Name)
	}
	b.WriteString("\n\n")

	// Parameters section
	if len(doc.Parameters) > 0 {
		b.WriteString("PARAMETERS:\n")
		for _, param := range doc.Parameters {
			optMarker := ""
			if param.Optional {
				optMarker = " (optional)"
			}
			b.WriteString(fmt.Sprintf("    %-12s [%s]%s\n", param.Name, param.Type, optMarker))
			b.WriteString(fmt.Sprintf("        %s\n", param.Description))
		}
		b.WriteString("\n")
	}

	// Returns section
	b.WriteString("RETURNS:\n")
	b.WriteString(fmt.Sprintf("    %s\n\n", doc.Returns))

	// Description section
	b.WriteString("DESCRIPTION:\n")
	for _, line := range strings.Split(strings.TrimSpace(doc.Description), "\n") {
		b.WriteString("    ")
		b.WriteString(line)
		b.WriteString("\n")
	}
	b.WriteString("\n")

	// Examples section
	if len(doc.Examples) > 0 {
		b.WriteString("EXAMPLES:\n")
		for _, example := range doc.Examples {
			for _, line := range strings.Split(example, "\n") {
				b.WriteString("    ")
				b.WriteString(line)
				b.WriteString("\n")
			}
		}
		b.WriteString("\n")
	}

	// See Also section
	if len(doc.SeeAlso) > 0 {
		b.WriteString("SEE ALSO:\n")
		b.WriteString("    ")
		b.WriteString(strings.Join(doc.SeeAlso, ", "))
		b.WriteString("\n\n")
	}

	return b.String()
}

// FormatCategoryList formats a list of all categories with function counts.
func FormatCategoryList(registry map[string]*NativeInfo) string {
	// Count functions per category
	categoryCount := make(map[string]int)
	for _, info := range registry {
		if info.Doc != nil && info.Doc.Category != "" {
			categoryCount[info.Doc.Category]++
		}
	}

	// Sort categories alphabetically
	categories := make([]string, 0, len(categoryCount))
	for cat := range categoryCount {
		categories = append(categories, cat)
	}
	sort.Strings(categories)

	var b strings.Builder
	b.WriteString("\nAvailable categories:\n")
	for _, cat := range categories {
		count := categoryCount[cat]
		plural := "function"
		if count != 1 {
			plural = "functions"
		}
		b.WriteString(fmt.Sprintf("  - %s (%d %s)\n", cat, count, plural))
	}
	b.WriteString("\nUse '? category' to list functions in a category\n")
	b.WriteString("Use '? function' to see detailed help\n\n")

	return b.String()
}

// FormatFunctionList formats a list of functions in a specific category.
func FormatFunctionList(category string, registry map[string]*NativeInfo) string {
	// Collect functions in this category
	type funcInfo struct {
		name    string
		summary string
	}
	functions := make([]funcInfo, 0)

	for name, info := range registry {
		if info.Doc != nil && strings.EqualFold(info.Doc.Category, category) {
			functions = append(functions, funcInfo{
				name:    name,
				summary: info.Doc.Summary,
			})
		}
	}

	if len(functions) == 0 {
		return fmt.Sprintf("\nCategory '%s' not found or has no functions.\n\n", category)
	}

	// Sort functions alphabetically
	sort.Slice(functions, func(i, j int) bool {
		return functions[i].name < functions[j].name
	})

	var b strings.Builder
	b.WriteString(fmt.Sprintf("\n%s Functions:\n", category))
	b.WriteString(strings.Repeat("-", len(category)+10))
	b.WriteString("\n\n")

	// Find max name length for alignment
	maxLen := 0
	for _, fn := range functions {
		if len(fn.name) > maxLen {
			maxLen = len(fn.name)
		}
	}

	for _, fn := range functions {
		padding := strings.Repeat(" ", maxLen-len(fn.name)+2)
		b.WriteString(fmt.Sprintf("  %s%s%s\n", fn.name, padding, fn.summary))
	}
	b.WriteString("\n")

	return b.String()
}

// FormatWordsList formats a flat list of all function names.
func FormatWordsList(registry map[string]*NativeInfo) string {
	// Collect all function names
	names := make([]string, 0, len(registry))
	for name := range registry {
		names = append(names, name)
	}

	// Sort alphabetically
	sort.Strings(names)

	var b strings.Builder
	b.WriteString("\n")

	// Group by category for better readability
	categories := GetCategories(registry)
	sort.Strings(categories)

	for _, cat := range categories {
		catFuncs := make([]string, 0)
		for _, name := range names {
			if info := registry[name]; info.Doc != nil && info.Doc.Category == cat {
				catFuncs = append(catFuncs, name)
			}
		}
		if len(catFuncs) > 0 {
			b.WriteString(strings.Join(catFuncs, "  "))
			b.WriteString("\n")
		}
	}

	b.WriteString(fmt.Sprintf("\nTotal: %d functions\n", len(names)))
	b.WriteString("Use '? function-name' for detailed help\n\n")

	return b.String()
}

// FindSimilar finds function names similar to the input (for typo suggestions).
// Uses simple string distance heuristics.
func FindSimilar(word string, registry map[string]*NativeInfo, maxResults int) []string {
	type match struct {
		name     string
		distance int
	}

	matches := make([]match, 0)
	word = strings.ToLower(word)

	for name := range registry {
		nameLower := strings.ToLower(name)

		// Check for various similarity criteria
		distance := 0

		// Exact prefix match (very similar)
		if strings.HasPrefix(nameLower, word) {
			distance = len(nameLower) - len(word)
		} else if strings.HasPrefix(word, nameLower) {
			distance = len(word) - len(nameLower)
		} else if strings.Contains(nameLower, word) {
			// Substring match
			distance = len(nameLower)
		} else if strings.Contains(word, nameLower) {
			distance = len(word)
		} else {
			// Levenshtein-like simple distance
			distance = levenshteinDistance(word, nameLower)
		}

		// Only include reasonably similar matches
		if distance <= 3 || (len(word) > 4 && distance <= len(word)/2) {
			matches = append(matches, match{name: name, distance: distance})
		}
	}

	// Sort by distance (most similar first)
	sort.Slice(matches, func(i, j int) bool {
		if matches[i].distance != matches[j].distance {
			return matches[i].distance < matches[j].distance
		}
		return matches[i].name < matches[j].name
	})

	// Return top N results
	result := make([]string, 0, maxResults)
	for i := 0; i < len(matches) && i < maxResults; i++ {
		result = append(result, matches[i].name)
	}

	return result
}

// levenshteinDistance calculates a simple edit distance between two strings.
func levenshteinDistance(s1, s2 string) int {
	if len(s1) == 0 {
		return len(s2)
	}
	if len(s2) == 0 {
		return len(s1)
	}

	// Create matrix
	matrix := make([][]int, len(s1)+1)
	for i := range matrix {
		matrix[i] = make([]int, len(s2)+1)
		matrix[i][0] = i
	}
	for j := range matrix[0] {
		matrix[0][j] = j
	}

	// Fill matrix
	for i := 1; i <= len(s1); i++ {
		for j := 1; j <= len(s2); j++ {
			cost := 0
			if s1[i-1] != s2[j-1] {
				cost = 1
			}

			matrix[i][j] = min3(
				matrix[i-1][j]+1,      // deletion
				matrix[i][j-1]+1,      // insertion
				matrix[i-1][j-1]+cost, // substitution
			)
		}
	}

	return matrix[len(s1)][len(s2)]
}

func min3(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}
