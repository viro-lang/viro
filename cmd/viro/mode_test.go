package main

import "testing"

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
		cfg     *Config
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
			cfg: &Config{
				ShowVersion: true,
			},
			want: ModeVersion,
		},
		{
			name: "help flag",
			cfg: &Config{
				ShowHelp: true,
			},
			want: ModeHelp,
		},
		{
			name: "eval flag",
			cfg: &Config{
				EvalExpr: "3 + 4",
			},
			want: ModeEval,
		},
		{
			name: "check with script",
			cfg: &Config{
				CheckOnly:  true,
				ScriptFile: "test.viro",
			},
			want: ModeCheck,
		},
		{
			name: "check without script",
			cfg: &Config{
				CheckOnly: true,
			},
			want:    ModeCheck,
			wantErr: false,
		},
		{
			name: "script file",
			cfg: &Config{
				ScriptFile: "test.viro",
			},
			want: ModeScript,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := detectMode(tt.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("detectMode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("detectMode() = %v, want %v", got, tt.want)
			}
		})
	}
}
