package native

import "testing"

func TestFormatInt(t *testing.T) {
	tests := []struct {
		name string
		in   int
		out  string
	}{
		{"zero", 0, "0"},
		{"positive", 42, "42"},
		{"negative", -17, "-17"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatInt(tt.in); got != tt.out {
				t.Fatalf("formatInt(%d) = %s, want %s", tt.in, got, tt.out)
			}
		})
	}
}

func TestFormatUint(t *testing.T) {
	tests := []struct {
		name string
		in   uint64
		out  string
	}{
		{"zero", 0, "0"},
		{"small", 7, "7"},
		{"large", 123456789, "123456789"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatUint(tt.in); got != tt.out {
				t.Fatalf("formatUint(%d) = %s, want %s", tt.in, got, tt.out)
			}
		})
	}
}
