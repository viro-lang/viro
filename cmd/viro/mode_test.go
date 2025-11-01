package main

import (
	"testing"

	"github.com/marcin-radoszewski/viro/internal/api"
)

func TestModeString(t *testing.T) {
	tests := []struct {
		mode Mode
		want string
	}{
		{ModeREPL, "REPL"},
		{ModeScript, "Script"},
		{ModeEval, "Eval"},
		{ModeCheck, "Check"},
		{ModeVersion, "Version"},
		{ModeHelp, "Help"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.mode.String()
			if got != tt.want {
				t.Errorf("Mode.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDetectMode(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *api.Config
		want    Mode
		wantErr bool
	}{
		{
			name: "default REPL",
			cfg:  NewConfig(),
			want: ModeREPL,
		},
		{
			name: "version flag",
			cfg: &api.Config{
				ShowVersion: true,
			},
			want: ModeVersion,
		},
		{
			name: "help flag",
			cfg: &api.Config{
				ShowHelp: true,
			},
			want: ModeHelp,
		},
		{
			name: "eval flag",
			cfg: &api.Config{
				EvalExpr: "3 + 4",
			},
			want: ModeEval,
		},
		{
			name: "check with script",
			cfg: &api.Config{
				CheckOnly:  true,
				ScriptFile: "test.viro",
			},
			want: ModeCheck,
		},
		{
			name: "check without script",
			cfg: &api.Config{
				CheckOnly: true,
			},
			want:    ModeCheck,
			wantErr: false,
		},
		{
			name: "script file",
			cfg: &api.Config{
				ScriptFile: "test.viro",
			},
			want: ModeScript,
		},
		{
			name: "multiple modes - version and help",
			cfg: &api.Config{
				ShowVersion: true,
				ShowHelp:    true,
			},
			want:    ModeREPL,
			wantErr: true,
		},
		{
			name: "multiple modes - eval and script",
			cfg: &api.Config{
				EvalExpr:   "3 + 4",
				ScriptFile: "test.viro",
			},
			want:    ModeREPL,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DetectMode(tt.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("DetectMode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("DetectMode() = %v, want %v", got, tt.want)
			}
		})
	}
}
