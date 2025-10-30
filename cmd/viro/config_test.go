package main

import (
	"flag"
	"os"
	"testing"
)

func TestNewConfig(t *testing.T) {
	cfg := NewConfig()
	if cfg == nil {
		t.Fatal("NewConfig() returned nil")
	}
	if cfg.SandboxRoot != "" {
		t.Errorf("SandboxRoot = %q, want empty string", cfg.SandboxRoot)
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
			if err := cfg.LoadFromEnv(); err != nil {
				t.Fatalf("LoadFromEnv() error = %v", err)
			}

			if cfg.SandboxRoot != tt.wantRoot {
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
		cfg     *Config
		wantErr bool
	}{
		{
			name:    "empty config valid",
			cfg:     NewConfig(),
			wantErr: false,
		},
		{
			name: "version only",
			cfg: &Config{
				ShowVersion: true,
			},
			wantErr: false,
		},
		{
			name: "help only",
			cfg: &Config{
				ShowHelp: true,
			},
			wantErr: false,
		},
		{
			name: "eval only",
			cfg: &Config{
				EvalExpr: "3 + 4",
			},
			wantErr: false,
		},
		{
			name: "check with script",
			cfg: &Config{
				CheckOnly:  true,
				ScriptFile: "test.viro",
			},
			wantErr: false,
		},
		{
			name: "check without script",
			cfg: &Config{
				CheckOnly: true,
			},
			wantErr: true,
		},
		{
			name: "script only",
			cfg: &Config{
				ScriptFile: "test.viro",
			},
			wantErr: false,
		},
		{
			name: "version and help",
			cfg: &Config{
				ShowVersion: true,
				ShowHelp:    true,
			},
			wantErr: true,
		},
		{
			name: "eval and script",
			cfg: &Config{
				EvalExpr:   "3 + 4",
				ScriptFile: "test.viro",
			},
			wantErr: true,
		},
		{
			name: "stdin without eval",
			cfg: &Config{
				ReadStdin: true,
			},
			wantErr: true,
		},
		{
			name: "stdin with eval",
			cfg: &Config{
				EvalExpr:  "first",
				ReadStdin: true,
			},
			wantErr: false,
		},
		{
			name: "no-print without eval",
			cfg: &Config{
				NoPrint: true,
			},
			wantErr: true,
		},
		{
			name: "no-print with eval",
			cfg: &Config{
				EvalExpr: "3 + 4",
				NoPrint:  true,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
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

	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	os.Args = []string{"cmd", "--sandbox-root=/from/flag"}

	cfg := NewConfig()
	if err := cfg.LoadFromEnv(); err != nil {
		t.Fatalf("LoadFromEnv() error = %v", err)
	}
	if err := cfg.LoadFromFlags(); err != nil {
		t.Fatalf("LoadFromFlags() error = %v", err)
	}

	if cfg.SandboxRoot != "/from/flag" {
		t.Errorf("SandboxRoot = %q, want %q (flags should override env)", cfg.SandboxRoot, "/from/flag")
	}
}
