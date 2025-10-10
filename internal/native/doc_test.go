package native

import (
	"strings"
	"testing"

	"github.com/marcin-radoszewski/viro/internal/value"
)

// TestNativeDoc_Validate tests the validation of NativeDoc structs
func TestNativeDoc_Validate(t *testing.T) {
	tests := []struct {
		name     string
		funcName string
		doc      *NativeDoc
		wantErr  string
	}{
		{
			name:     "valid documentation",
			funcName: "testfunc",
			doc: &NativeDoc{
				Category:    "Math",
				Summary:     "Adds two numbers",
				Description: "This function adds two numbers together",
				Parameters: []ParamDoc{
					{Name: "a", Type: "integer!", Description: "First number", Optional: false},
					{Name: "b", Type: "integer!", Description: "Second number", Optional: false},
				},
				Returns:  "[integer!] The sum of a and b",
				Examples: []string{"testfunc 1 2  ; => 3"},
				SeeAlso:  []string{"subtract"},
				Tags:     []string{"arithmetic"},
			},
			wantErr: "",
		},
		{
			name:     "missing category",
			funcName: "testfunc",
			doc: &NativeDoc{
				Category:    "",
				Summary:     "Test summary",
				Description: "Test description",
				Returns:     "something",
				Examples:    []string{"example"},
			},
			wantErr: "testfunc: missing category",
		},
		{
			name:     "missing summary",
			funcName: "testfunc",
			doc: &NativeDoc{
				Category:    "Math",
				Summary:     "",
				Description: "Test description",
				Returns:     "something",
				Examples:    []string{"example"},
			},
			wantErr: "testfunc: missing summary",
		},
		{
			name:     "missing description",
			funcName: "testfunc",
			doc: &NativeDoc{
				Category:    "Math",
				Summary:     "Test summary",
				Description: "",
				Returns:     "something",
				Examples:    []string{"example"},
			},
			wantErr: "testfunc: missing description",
		},
		{
			name:     "missing returns",
			funcName: "testfunc",
			doc: &NativeDoc{
				Category:    "Math",
				Summary:     "Test summary",
				Description: "Test description",
				Returns:     "",
				Examples:    []string{"example"},
			},
			wantErr: "testfunc: missing returns documentation",
		},
		{
			name:     "missing examples",
			funcName: "testfunc",
			doc: &NativeDoc{
				Category:    "Math",
				Summary:     "Test summary",
				Description: "Test description",
				Returns:     "something",
				Examples:    []string{},
			},
			wantErr: "testfunc: missing examples",
		},
		{
			name:     "parameter missing name",
			funcName: "testfunc",
			doc: &NativeDoc{
				Category:    "Math",
				Summary:     "Test summary",
				Description: "Test description",
				Parameters: []ParamDoc{
					{Name: "", Type: "integer!", Description: "Test param"},
				},
				Returns:  "something",
				Examples: []string{"example"},
			},
			wantErr: "missing name",
		},
		{
			name:     "parameter missing type",
			funcName: "testfunc",
			doc: &NativeDoc{
				Category:    "Math",
				Summary:     "Test summary",
				Description: "Test description",
				Parameters: []ParamDoc{
					{Name: "param1", Type: "", Description: "Test param"},
				},
				Returns:  "something",
				Examples: []string{"example"},
			},
			wantErr: "missing type",
		},
		{
			name:     "parameter missing description",
			funcName: "testfunc",
			doc: &NativeDoc{
				Category:    "Math",
				Summary:     "Test summary",
				Description: "Test description",
				Parameters: []ParamDoc{
					{Name: "param1", Type: "integer!", Description: ""},
				},
				Returns:  "something",
				Examples: []string{"example"},
			},
			wantErr: "missing description",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.doc.Validate(tt.funcName)
			if tt.wantErr == "" {
				if got != "" {
					t.Errorf("Validate() = %q, want no error", got)
				}
			} else {
				if !strings.Contains(got, tt.wantErr) {
					t.Errorf("Validate() = %q, want error containing %q", got, tt.wantErr)
				}
			}
		})
	}
}

// TestNativeDoc_HasDoc tests the HasDoc method
func TestNativeDoc_HasDoc(t *testing.T) {
	tests := []struct {
		name string
		doc  *NativeDoc
		want bool
	}{
		{
			name: "nil doc",
			doc:  nil,
			want: false,
		},
		{
			name: "empty doc",
			doc:  &NativeDoc{},
			want: false,
		},
		{
			name: "doc with summary",
			doc: &NativeDoc{
				Summary: "Has documentation",
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.doc.HasDoc(); got != tt.want {
				t.Errorf("HasDoc() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestGetCategories tests the GetCategories function
func TestGetCategories(t *testing.T) {
	registry := map[string]*value.FunctionValue{
		"add": {
			Doc: &NativeDoc{Category: "Math"},
		},
		"subtract": {
			Doc: &NativeDoc{Category: "Math"},
		},
		"if": {
			Doc: &NativeDoc{Category: "Control"},
		},
		"print": {
			Doc: &NativeDoc{Category: "I/O"},
		},
		"undocumented": {
			Doc: nil,
		},
	}

	categories := GetCategories(registry)

	// Should have exactly 3 categories
	if len(categories) != 3 {
		t.Errorf("GetCategories() returned %d categories, want 3", len(categories))
	}

	// Check that all expected categories are present
	categoryMap := make(map[string]bool)
	for _, cat := range categories {
		categoryMap[cat] = true
	}

	expectedCategories := []string{"Math", "Control", "I/O"}
	for _, expected := range expectedCategories {
		if !categoryMap[expected] {
			t.Errorf("GetCategories() missing category %q", expected)
		}
	}
}

// TestGetFunctionsInCategory tests the GetFunctionsInCategory function
func TestGetFunctionsInCategory(t *testing.T) {
	registry := map[string]*value.FunctionValue{
		"add": {
			Doc: &NativeDoc{Category: "Math"},
		},
		"subtract": {
			Doc: &NativeDoc{Category: "Math"},
		},
		"multiply": {
			Doc: &NativeDoc{Category: "Math"},
		},
		"if": {
			Doc: &NativeDoc{Category: "Control"},
		},
		"undocumented": {
			Doc: nil,
		},
	}

	tests := []struct {
		name     string
		category string
		wantLen  int
		wantFns  []string
	}{
		{
			name:     "Math category",
			category: "Math",
			wantLen:  3,
			wantFns:  []string{"add", "subtract", "multiply"},
		},
		{
			name:     "Control category",
			category: "Control",
			wantLen:  1,
			wantFns:  []string{"if"},
		},
		{
			name:     "Non-existent category",
			category: "NonExistent",
			wantLen:  0,
			wantFns:  []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			functions := GetFunctionsInCategory(registry, tt.category)
			if len(functions) != tt.wantLen {
				t.Errorf("GetFunctionsInCategory(%q) returned %d functions, want %d",
					tt.category, len(functions), tt.wantLen)
			}

			// Check that all expected functions are present
			funcMap := make(map[string]bool)
			for _, fn := range functions {
				funcMap[fn] = true
			}

			for _, expected := range tt.wantFns {
				if !funcMap[expected] {
					t.Errorf("GetFunctionsInCategory(%q) missing function %q",
						tt.category, expected)
				}
			}
		})
	}
}

// TestCountDocumented tests the CountDocumented function
func TestCountDocumented(t *testing.T) {
	tests := []struct {
		name           string
		registry       map[string]*value.FunctionValue
		wantDocumented int
		wantTotal      int
	}{
		{
			name: "all documented",
			registry: map[string]*value.FunctionValue{
				"fn1": {Doc: &NativeDoc{Summary: "Doc 1"}},
				"fn2": {Doc: &NativeDoc{Summary: "Doc 2"}},
			},
			wantDocumented: 2,
			wantTotal:      2,
		},
		{
			name: "none documented",
			registry: map[string]*value.FunctionValue{
				"fn1": {Doc: nil},
				"fn2": {Doc: nil},
			},
			wantDocumented: 0,
			wantTotal:      2,
		},
		{
			name: "partially documented",
			registry: map[string]*value.FunctionValue{
				"fn1": {Doc: &NativeDoc{Summary: "Doc 1"}},
				"fn2": {Doc: nil},
				"fn3": {Doc: &NativeDoc{Summary: "Doc 3"}},
				"fn4": {Doc: &NativeDoc{}}, // Empty doc (no summary)
			},
			wantDocumented: 2,
			wantTotal:      4,
		},
		{
			name:           "empty registry",
			registry:       map[string]*value.FunctionValue{},
			wantDocumented: 0,
			wantTotal:      0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			documented, total := CountDocumented(tt.registry)
			if documented != tt.wantDocumented {
				t.Errorf("CountDocumented() documented = %d, want %d",
					documented, tt.wantDocumented)
			}
			if total != tt.wantTotal {
				t.Errorf("CountDocumented() total = %d, want %d",
					total, tt.wantTotal)
			}
		})
	}
}

// TestNewDocTemplate tests the NewDocTemplate function
func TestNewDocTemplate(t *testing.T) {
	tests := []struct {
		name       string
		funcName   string
		category   string
		paramCount int
	}{
		{
			name:       "zero parameters",
			funcName:   "testfunc",
			category:   "Test",
			paramCount: 0,
		},
		{
			name:       "one parameter",
			funcName:   "testfunc",
			category:   "Test",
			paramCount: 1,
		},
		{
			name:       "multiple parameters",
			funcName:   "testfunc",
			category:   "Test",
			paramCount: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := NewDocTemplate(tt.funcName, tt.category, tt.paramCount)

			if doc.Category != tt.category {
				t.Errorf("NewDocTemplate() category = %q, want %q", doc.Category, tt.category)
			}

			if len(doc.Parameters) != tt.paramCount {
				t.Errorf("NewDocTemplate() parameter count = %d, want %d",
					len(doc.Parameters), tt.paramCount)
			}

			// Check that all parameters have default values
			for i, param := range doc.Parameters {
				expectedName := "param" + string(rune('1'+i))
				if param.Name != expectedName {
					t.Errorf("Parameter %d name = %q, want %q", i, param.Name, expectedName)
				}
				if param.Type == "" {
					t.Errorf("Parameter %d missing type", i)
				}
				if param.Description == "" {
					t.Errorf("Parameter %d missing description", i)
				}
			}

			// Check that examples contain function name
			if len(doc.Examples) == 0 {
				t.Error("NewDocTemplate() has no examples")
			} else if !strings.Contains(doc.Examples[0], tt.funcName) {
				t.Errorf("Example doesn't contain function name: %q", doc.Examples[0])
			}
		})
	}
}

// TestValidateRegistry tests the ValidateRegistry function
func TestValidateRegistry(t *testing.T) {
	tests := []struct {
		name         string
		registry     map[string]*value.FunctionValue
		wantErrorCnt int
	}{
		{
			name: "all valid",
			registry: map[string]*value.FunctionValue{
				"fn1": {
					Doc: &NativeDoc{
						Category:    "Test",
						Summary:     "Test function 1",
						Description: "Description 1",
						Returns:     "result",
						Examples:    []string{"example"},
					},
				},
				"fn2": {
					Doc: &NativeDoc{
						Category:    "Test",
						Summary:     "Test function 2",
						Description: "Description 2",
						Returns:     "result",
						Examples:    []string{"example"},
					},
				},
			},
			wantErrorCnt: 0,
		},
		{
			name: "some invalid",
			registry: map[string]*value.FunctionValue{
				"fn1": {
					Doc: &NativeDoc{
						Category:    "Test",
						Summary:     "Test function 1",
						Description: "Description 1",
						Returns:     "result",
						Examples:    []string{"example"},
					},
				},
				"fn2": {
					Doc: &NativeDoc{
						Category:    "",
						Summary:     "Test function 2",
						Description: "Description 2",
						Returns:     "result",
						Examples:    []string{"example"},
					},
				},
				"fn3": {
					Doc: &NativeDoc{
						Category:    "Test",
						Summary:     "",
						Description: "Description 3",
						Returns:     "result",
						Examples:    []string{"example"},
					},
				},
			},
			wantErrorCnt: 2,
		},
		{
			name: "undocumented functions ignored",
			registry: map[string]*value.FunctionValue{
				"fn1": {Doc: nil},
				"fn2": {Doc: nil},
			},
			wantErrorCnt: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := ValidateRegistry(tt.registry)
			if len(errors) != tt.wantErrorCnt {
				t.Errorf("ValidateRegistry() returned %d errors, want %d\nErrors: %v",
					len(errors), tt.wantErrorCnt, errors)
			}
		})
	}
}
