package main

import (
	"errors"
	"testing"

	"github.com/marcin-radoszewski/viro/internal/verror"
)

func TestCategoryToExitCode(t *testing.T) {
	tests := []struct {
		name     string
		category verror.ErrorCategory
		want     int
	}{
		{"syntax error", verror.ErrSyntax, ExitSyntax},
		{"script error", verror.ErrScript, ExitError},
		{"math error", verror.ErrMath, ExitError},
		{"access error", verror.ErrAccess, ExitAccess},
		{"internal error", verror.ErrInternal, ExitInternal},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := categoryToExitCode(tt.category)
			if got != tt.want {
				t.Errorf("categoryToExitCode(%v) = %d, want %d", tt.category, got, tt.want)
			}
		})
	}
}

func TestHandleError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{
			name: "nil error",
			err:  nil,
			want: ExitSuccess,
		},
		{
			name: "syntax error",
			err:  verror.NewSyntaxError(verror.ErrIDInvalidSyntax, [3]string{"test", "", ""}),
			want: ExitSyntax,
		},
		{
			name: "script error",
			err:  verror.NewScriptError(verror.ErrIDNoValue, [3]string{"test", "", ""}),
			want: ExitError,
		},
		{
			name: "math error",
			err:  verror.NewMathError(verror.ErrIDDivByZero, [3]string{"", "", ""}),
			want: ExitError,
		},
		{
			name: "access error",
			err:  verror.NewAccessError(verror.ErrIDSandboxViolation, [3]string{"test", "", ""}),
			want: ExitAccess,
		},
		{
			name: "internal error",
			err:  verror.NewInternalError(verror.ErrIDStackOverflow, [3]string{"", "", ""}),
			want: ExitInternal,
		},
		{
			name: "generic error",
			err:  errors.New("generic error"),
			want: ExitError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := handleError(tt.err)
			if got != tt.want {
				t.Errorf("handleError() = %d, want %d", got, tt.want)
			}
		})
	}
}
