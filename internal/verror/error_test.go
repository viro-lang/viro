package verror

import (
	"strings"
	"testing"
)

func TestErrorStringLocation(t *testing.T) {
	tests := []struct {
		name string
		err  *Error
		want string
	}{
		{
			name: "without location",
			err:  NewSyntaxError(ErrIDInvalidSyntax, [3]string{"token", "", ""}),
			want: "Syntax error (200): Invalid syntax: token",
		},
		{
			name: "with location",
			err:  NewSyntaxError(ErrIDInvalidSyntax, [3]string{"token", "", ""}).SetLocation("script.viro", 12, 5),
			want: "script.viro:12:5 Syntax error (200): Invalid syntax: token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			line := strings.SplitN(got, "\n", 2)[0]

			if line != tt.want {
				t.Fatalf("Error() header mismatch\nwant: %q\ngot: %q", tt.want, line)
			}
		})
	}
}

func TestFormatErrorWithContextLocation(t *testing.T) {
	tests := []struct {
		name string
		err  *Error
		want string
	}{
		{
			name: "without location",
			err:  NewSyntaxError(ErrIDInvalidSyntax, [3]string{"token", "", ""}),
			want: "** Syntax Error (invalid-syntax)",
		},
		{
			name: "with location",
			err:  NewSyntaxError(ErrIDInvalidSyntax, [3]string{"token", "", ""}).SetLocation("script.viro", 12, 5),
			want: "script.viro:12:5 ** Syntax Error (invalid-syntax)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatErrorWithContext(tt.err)
			line := strings.SplitN(got, "\n", 2)[0]

			if line != tt.want {
				t.Fatalf("FormatErrorWithContext() header mismatch\nwant: %q\ngot: %q", tt.want, line)
			}
		})
	}
}

func TestInvalidPathReasonFormatting(t *testing.T) {
	tests := []struct {
		name string
		args [3]string
		want string
	}{
		{
			name: "without reason",
			args: [3]string{"bad", "", ""},
			want: "Invalid path: bad",
		},
		{
			name: "with reason",
			args: [3]string{"bad", "missing", ""},
			want: "Invalid path (missing): bad",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewScriptError(ErrIDInvalidPath, tt.args)
			if err.Message != tt.want {
				t.Fatalf("invalid path message mismatch\nwant: %q\ngot: %q", tt.want, err.Message)
			}
		})
	}
}
