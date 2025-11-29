// Package contract validates WebUI natives per contracts/webui.md
package contract

import (
	"testing"
)

// TestWebUI_Window validates webui.window native
func TestWebUI_Window(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string // Molded string for comparison
		wantErr  bool
	}{
		{
			name:     "window with title",
			input:    "webui.window [title: \"Test\"]",
			expected: "webui-window!", // Mock handle
			wantErr:  false,
		},
		{
			name:     "window with dimensions",
			input:    "webui.window [width: 800 height: 600]",
			expected: "webui-window!",
			wantErr:  false,
		},
		{
			name:     "invalid spec type",
			input:    "webui.window \"invalid\"",
			expected: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Evaluate(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("TestWebUI_Window() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && result.Mold() != tt.expected {
				t.Errorf("TestWebUI_Window() = %v, expected %v", result.Mold(), tt.expected)
			}
		})
	}
}

// TestWebUI_Render validates webui.render native
func TestWebUI_Render(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		wantErr  bool
	}{
		{
			name:     "render string",
			input:    "w: webui.window [] webui.render w \"<html></html>\"",
			expected: "true",
			wantErr:  false,
		},
		{
			name:     "render with content-type",
			input:    "w: webui.window [] webui.render w \"data\" [content-type: text/plain]",
			expected: "true",
			wantErr:  false,
		},
		{
			name:     "binary without content-type",
			input:    "w: webui.window [] webui.render w #{00}",
			expected: "",
			wantErr:  true,
		},
		{
			name:     "invalid content-type",
			input:    "w: webui.window [] webui.render w \"html\" [content-type: invalid]",
			expected: "",
			wantErr:  true,
		},
		{
			name:     "binary without content-type",
			input:    "w: webui.window [] webui.render w #{00} []",
			expected: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Evaluate(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("TestWebUI_Render() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && result.Mold() != tt.expected {
				t.Errorf("TestWebUI_Render() = %v, expected %v", result.Mold(), tt.expected)
			}
		})
	}
}

// TestWebUI_Send validates webui.send native
func TestWebUI_Send(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		wantErr  bool
	}{
		{
			name:     "send string",
			input:    "w: webui.window [] webui.send w \"msg\" \"payload\"",
			expected: "true",
			wantErr:  false,
		},
		{
			name:     "send map",
			input:    "w: webui.window [] webui.send w \"msg\" [key: \"value\"]",
			expected: "true",
			wantErr:  false,
		},
		{
			name:     "invalid window",
			input:    "webui.send none \"msg\" \"payload\"",
			expected: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Evaluate(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("TestWebUI_Send() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && result.Mold() != tt.expected {
				t.Errorf("TestWebUI_Send() = %v, expected %v", result.Mold(), tt.expected)
			}
		})
	}
}

// TestWebUI_On validates webui.on native
func TestWebUI_On(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		wantErr  bool
	}{
		{
			name:     "register event handler",
			input:    "w: webui.window [] webui.on w \"click\" \"#btn\" [print event-name]",
			expected: "true",
			wantErr:  false,
		},
		{
			name:     "invalid selector type",
			input:    "w: webui.window [] webui.on w \"click\" 123 [block]",
			expected: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Evaluate(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("TestWebUI_On() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && result.Mold() != tt.expected {
				t.Errorf("TestWebUI_On() = %v, expected %v", result.Mold(), tt.expected)
			}
		})
	}
}

// TestWebUI_Poll validates webui.poll native
func TestWebUI_Poll(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		wantErr  bool
	}{
		{
			name:     "poll window",
			input:    "w: webui.window [] webui.poll w",
			expected: "none",
			wantErr:  false,
		},
		{
			name:     "poll all windows",
			input:    "webui.poll none",
			expected: "none",
			wantErr:  false,
		},
		{
			name:     "poll after close",
			input:    "w: webui.window [] webui.close w webui.poll w",
			expected: "none",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Evaluate(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("TestWebUI_Poll() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && result.Mold() != tt.expected {
				t.Errorf("TestWebUI_Poll() = %v, expected %v", result.Mold(), tt.expected)
			}
		})
	}
}

// TestWebUI_Close validates webui.close native
func TestWebUI_Close(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		wantErr  bool
	}{
		{
			name:     "close window",
			input:    "w: webui.window [] webui.close w",
			expected: "true",
			wantErr:  false,
		},
		{
			name:     "close invalid window",
			input:    "webui.close none",
			expected: "false",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Evaluate(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("TestWebUI_Close() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && result.Mold() != tt.expected {
				t.Errorf("TestWebUI_Close() = %v, expected %v", result.Mold(), tt.expected)
			}
		})
	}
}

// TestWebUI_Ready validates webui.ready? native
func TestWebUI_Ready(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		wantErr  bool
	}{
		{
			name:     "ready check before render",
			input:    "w: webui.window [] webui.ready? w",
			expected: "false",
			wantErr:  false,
		},
		{
			name:     "ready check after render",
			input:    "w: webui.window [] webui.render w \"<html></html>\" webui.ready? w",
			expected: "true",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Evaluate(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("TestWebUI_Ready() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && result.Mold() != tt.expected {
				t.Errorf("TestWebUI_Ready() = %v, expected %v", result.Mold(), tt.expected)
			}
		})
	}
}

// TestWebUI_Inject validates webui.inject native
func TestWebUI_Inject(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		wantErr  bool
	}{
		{
			name:     "inject html",
			input:    "w: webui.window [] webui.inject w \"<div>test</div>\"",
			expected: "true",
			wantErr:  false,
		},
		{
			name:     "inject binary",
			input:    "w: webui.window [] webui.inject w #{3c6469763e}",
			expected: "true",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Evaluate(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("TestWebUI_Inject() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && result.Mold() != tt.expected {
				t.Errorf("TestWebUI_Inject() = %v, expected %v", result.Mold(), tt.expected)
			}
		})
	}
}
