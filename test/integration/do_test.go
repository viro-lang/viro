package integration

import (
	"bytes"
	"strings"
	"testing"

	"github.com/marcin-radoszewski/viro/internal/api"
)

func TestDoFunction(t *testing.T) {
	tests := []struct {
		name     string
		expr     string
		want     string
		wantExit int
	}{
		{
			name:     "do evaluates block",
			expr:     "a: [ print \"Foo\" 10 ]\ndo a",
			want:     "Foo",
			wantExit: 0,
		},
		{
			name:     "do with simple value",
			expr:     "do 42",
			want:     "42",
			wantExit: 0,
		},
		{
			name:     "do with expression",
			expr:     "do (3 + 4)",
			want:     "7",
			wantExit: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			cfg, _ := api.ConfigFromArgs([]string{"-c", tt.expr})
			ctx := &api.RuntimeContext{
				Args:   []string{"-c", tt.expr},
				Stdin:  &bytes.Buffer{},
				Stdout: &stdout,
				Stderr: &stderr,
			}

			exitCode := api.Run(ctx, cfg)

			if exitCode != tt.wantExit {
				t.Errorf("exit code = %d, want %d\nStdout: %s\nStderr: %s",
					exitCode, tt.wantExit, stdout.String(), stderr.String())
			}

			output := stdout.String() + stderr.String()
			if !strings.Contains(output, tt.want) {
				t.Errorf("output = %q, want to contain %q", output, tt.want)
			}
		})
	}
}

func TestDoFunctionWithNext(t *testing.T) {
	tests := []struct {
		name     string
		script   string
		want     string
		wantExit int
	}{
		{
			name: "do --next binds remaining block",
			script: `a: [ print "Foo" 10 ]
do a --next 'b
print head? b
print index? b
none`,
			want:     "Foo\nfalse\n3\nnone",
			wantExit: 0,
		},
		{
			name: "do --next with block consuming multiple expressions",
			script: `a: [ 1 + 2 "hello" 99 ]
result: do a --next 'b
print result
print first b
none`,
			want:     "3\nhello\nnone",
			wantExit: 0,
		},
		{
			name: "do --next with single expression block",
			script: `a: [ 42 ]
result: do a --next 'b
print result
print type? b
none`,
			want:     "42\nblock!\nnone",
			wantExit: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			cfg, _ := api.ConfigFromArgs([]string{"-c", tt.script})
			ctx := &api.RuntimeContext{
				Args:   []string{"-c", tt.script},
				Stdin:  &bytes.Buffer{},
				Stdout: &stdout,
				Stderr: &stderr,
			}

			exitCode := api.Run(ctx, cfg)

			if exitCode != tt.wantExit {
				t.Errorf("exit code = %d, want %d\nStdout: %s\nStderr: %s",
					exitCode, tt.wantExit, stdout.String(), stderr.String())
			}

			output := stdout.String()
			outputLines := strings.Split(strings.TrimSpace(output), "\n")
			wantLines := strings.Split(strings.TrimSpace(tt.want), "\n")

			if len(outputLines) != len(wantLines) {
				t.Errorf("output lines = %d, want %d\nOutput: %s\nWant: %s",
					len(outputLines), len(wantLines), output, tt.want)
				return
			}

			for i := range outputLines {
				if !strings.Contains(outputLines[i], wantLines[i]) {
					t.Errorf("line %d: got %q, want to contain %q", i, outputLines[i], wantLines[i])
				}
			}
		})
	}
}
