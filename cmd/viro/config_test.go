package main

import (
	"flag"
	"os"
	"testing"

	"github.com/marcin-radoszewski/viro/internal/api"
)

func setupTestArgs(t *testing.T, args []string) {
	t.Helper()
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	os.Args = args
}

func TestNewConfig(t *testing.T) {
	cfg := NewConfig()
	if cfg == nil {
		t.Fatal("NewConfig() returned nil")
	}
	if cfg.SandboxRoot == "" {
		t.Errorf("SandboxRoot should be set to cwd by default, got empty string")
	}
}

func TestLoadFromEnv(t *testing.T) {
	oldRoot := os.Getenv("VIRO_SANDBOX_ROOT")
	oldTLS := os.Getenv("VIRO_ALLOW_INSECURE_TLS")
	oldHistory := os.Getenv("VIRO_HISTORY_FILE")
	defer func() {
		os.Setenv("VIRO_SANDBOX_ROOT", oldRoot)
		os.Setenv("VIRO_ALLOW_INSECURE_TLS", oldTLS)
		os.Setenv("VIRO_HISTORY_FILE", oldHistory)
	}()

	tests := []struct {
		name        string
		envVars     map[string]string
		wantRoot    string
		wantTLS     bool
		wantHistory string
	}{
		{
			name: "no env vars",
			envVars: map[string]string{
				"VIRO_SANDBOX_ROOT":       "",
				"VIRO_ALLOW_INSECURE_TLS": "",
				"VIRO_HISTORY_FILE":       "",
			},
			wantRoot:    "",
			wantTLS:     false,
			wantHistory: "",
		},
		{
			name: "sandbox root set",
			envVars: map[string]string{
				"VIRO_SANDBOX_ROOT":       "/tmp/test",
				"VIRO_ALLOW_INSECURE_TLS": "",
				"VIRO_HISTORY_FILE":       "",
			},
			wantRoot:    "/tmp/test",
			wantTLS:     false,
			wantHistory: "",
		},
		{
			name: "tls flag true",
			envVars: map[string]string{
				"VIRO_SANDBOX_ROOT":       "",
				"VIRO_ALLOW_INSECURE_TLS": "true",
				"VIRO_HISTORY_FILE":       "",
			},
			wantRoot:    "",
			wantTLS:     true,
			wantHistory: "",
		},
		{
			name: "tls flag 1",
			envVars: map[string]string{
				"VIRO_SANDBOX_ROOT":       "",
				"VIRO_ALLOW_INSECURE_TLS": "1",
				"VIRO_HISTORY_FILE":       "",
			},
			wantRoot:    "",
			wantTLS:     true,
			wantHistory: "",
		},
		{
			name: "history file set",
			envVars: map[string]string{
				"VIRO_SANDBOX_ROOT":       "",
				"VIRO_ALLOW_INSECURE_TLS": "",
				"VIRO_HISTORY_FILE":       "/tmp/.viro_history",
			},
			wantRoot:    "",
			wantTLS:     false,
			wantHistory: "/tmp/.viro_history",
		},
		{
			name: "all env vars set",
			envVars: map[string]string{
				"VIRO_SANDBOX_ROOT":       "/workspace",
				"VIRO_ALLOW_INSECURE_TLS": "true",
				"VIRO_HISTORY_FILE":       "~/.history",
			},
			wantRoot:    "/workspace",
			wantTLS:     true,
			wantHistory: "~/.history",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}

			cfg := NewConfig()
			if err := LoadFromEnv(cfg); err != nil {
				t.Fatalf("LoadFromEnv() error = %v", err)
			}

			if tt.wantRoot != "" && cfg.SandboxRoot != tt.wantRoot {
				t.Errorf("SandboxRoot = %q, want %q", cfg.SandboxRoot, tt.wantRoot)
			}
			if cfg.AllowInsecureTLS != tt.wantTLS {
				t.Errorf("AllowInsecureTLS = %v, want %v", cfg.AllowInsecureTLS, tt.wantTLS)
			}
			if cfg.HistoryFile != tt.wantHistory {
				t.Errorf("HistoryFile = %q, want %q", cfg.HistoryFile, tt.wantHistory)
			}
		})
	}
}

func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *api.Config
		wantErr bool
	}{
		{
			name:    "empty config valid",
			cfg:     NewConfig(),
			wantErr: false,
		},
		{
			name: "version only",
			cfg: &api.Config{
				ShowVersion: true,
			},
			wantErr: false,
		},
		{
			name: "help only",
			cfg: &api.Config{
				ShowHelp: true,
			},
			wantErr: false,
		},
		{
			name: "eval only",
			cfg: &api.Config{
				EvalExpr: "3 + 4",
			},
			wantErr: false,
		},
		{
			name: "check with script",
			cfg: &api.Config{
				CheckOnly:  true,
				ScriptFile: "test.viro",
			},
			wantErr: false,
		},
		{
			name: "check without script",
			cfg: &api.Config{
				CheckOnly: true,
			},
			wantErr: true,
		},
		{
			name: "script only",
			cfg: &api.Config{
				ScriptFile: "test.viro",
			},
			wantErr: false,
		},
		{
			name: "version and help",
			cfg: &api.Config{
				ShowVersion: true,
				ShowHelp:    true,
			},
			wantErr: false,
		},
		{
			name: "eval and script",
			cfg: &api.Config{
				EvalExpr:   "3 + 4",
				ScriptFile: "test.viro",
			},
			wantErr: false,
		},
		{
			name: "stdin without eval",
			cfg: &api.Config{
				ReadStdin: true,
			},
			wantErr: true,
		},
		{
			name: "stdin with eval",
			cfg: &api.Config{
				EvalExpr:  "first",
				ReadStdin: true,
			},
			wantErr: false,
		},
		{
			name: "no-print without eval",
			cfg: &api.Config{
				NoPrint: true,
			},
			wantErr: true,
		},
		{
			name: "no-print with eval",
			cfg: &api.Config{
				EvalExpr: "3 + 4",
				NoPrint:  true,
			},
			wantErr: false,
		},
		{
			name: "profile without script",
			cfg: &api.Config{
				Profile: true,
			},
			wantErr: true,
		},
		{
			name: "profile with script",
			cfg: &api.Config{
				Profile:    true,
				ScriptFile: "test.viro",
			},
			wantErr: false,
		},
		{
			name: "profile with eval",
			cfg: &api.Config{
				Profile:  true,
				EvalExpr: "3 + 4",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(tt.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConfigFlagsPriority(t *testing.T) {
	oldRoot := os.Getenv("VIRO_SANDBOX_ROOT")
	defer os.Setenv("VIRO_SANDBOX_ROOT", oldRoot)

	os.Setenv("VIRO_SANDBOX_ROOT", "/from/env")

	setupTestArgs(t, []string{"cmd", "--sandbox-root=/from/flag"})

	cfg := NewConfig()
	if err := LoadFromEnv(cfg); err != nil {
		t.Fatalf("LoadFromEnv() error = %v", err)
	}
	if err := LoadFromFlags(cfg); err != nil {
		t.Fatalf("LoadFromFlags() error = %v", err)
	}

	if cfg.SandboxRoot != "/from/flag" {
		t.Errorf("SandboxRoot = %q, want %q (flags should override env)", cfg.SandboxRoot, "/from/flag")
	}
}

func TestScriptArgumentParsing(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		wantScriptFile string
		wantArgs       []string
		wantErr        bool
	}{
		{
			name:           "script without args",
			args:           []string{"cmd", "script.viro"},
			wantScriptFile: "script.viro",
			wantArgs:       []string{},
			wantErr:        false,
		},
		{
			name:           "script with args",
			args:           []string{"cmd", "script.viro", "arg1", "arg2"},
			wantScriptFile: "script.viro",
			wantArgs:       []string{"arg1", "arg2"},
			wantErr:        false,
		},
		{
			name:           "script with flag-like args",
			args:           []string{"cmd", "script.viro", "--verbose", "--output", "file.txt"},
			wantScriptFile: "script.viro",
			wantArgs:       []string{"--verbose", "--output", "file.txt"},
			wantErr:        false,
		},
		{
			name:           "viro flags before script",
			args:           []string{"cmd", "--quiet", "--sandbox-root", "/tmp", "script.viro", "arg1"},
			wantScriptFile: "script.viro",
			wantArgs:       []string{"arg1"},
			wantErr:        false,
		},
		{
			name:           "viro flags mixed - flags before script, args after",
			args:           []string{"cmd", "--quiet", "script.viro", "--sandbox-root", "/tmp"},
			wantScriptFile: "script.viro",
			wantArgs:       []string{"--sandbox-root", "/tmp"},
			wantErr:        false,
		},
		{
			name:           "script args with spaces via quoting (simulated)",
			args:           []string{"cmd", "script.viro", "hello world", "arg2"},
			wantScriptFile: "script.viro",
			wantArgs:       []string{"hello world", "arg2"},
			wantErr:        false,
		},
		{
			name:           "empty args (just script)",
			args:           []string{"cmd", "test.viro"},
			wantScriptFile: "test.viro",
			wantArgs:       []string{},
			wantErr:        false,
		},
		{
			name:           "multiple numeric args",
			args:           []string{"cmd", "script.viro", "42", "3.14", "100"},
			wantScriptFile: "script.viro",
			wantArgs:       []string{"42", "3.14", "100"},
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupTestArgs(t, tt.args)

			cfg := NewConfig()
			err := LoadFromFlags(cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadFromFlags() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil {
				return
			}

			if cfg.ScriptFile != tt.wantScriptFile {
				t.Errorf("ScriptFile = %q, want %q", cfg.ScriptFile, tt.wantScriptFile)
			}

			if len(cfg.Args) != len(tt.wantArgs) {
				t.Errorf("Args length = %d, want %d", len(cfg.Args), len(tt.wantArgs))
			}

			for i, arg := range cfg.Args {
				if i >= len(tt.wantArgs) {
					break
				}
				if arg != tt.wantArgs[i] {
					t.Errorf("Args[%d] = %q, want %q", i, arg, tt.wantArgs[i])
				}
			}
		})
	}
}

func TestScriptArgumentsNoScriptMode(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		wantArgs []string
	}{
		{
			name:     "REPL mode - no args",
			args:     []string{"cmd"},
			wantArgs: nil,
		},
		{
			name:     "version mode - no args",
			args:     []string{"cmd", "--version"},
			wantArgs: nil,
		},
		{
			name:     "help mode - no args",
			args:     []string{"cmd", "--help"},
			wantArgs: nil,
		},
		{
			name:     "eval mode - no args",
			args:     []string{"cmd", "-c", "3 + 4"},
			wantArgs: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupTestArgs(t, tt.args)

			cfg := NewConfig()
			if err := LoadFromFlags(cfg); err != nil {
				t.Fatalf("LoadFromFlags() error = %v", err)
			}

			if len(cfg.Args) != 0 {
				t.Errorf("Args = %v, want empty in non-script mode", cfg.Args)
			}
		})
	}
}
