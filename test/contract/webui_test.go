package contract

import (
	"testing"
)

func TestWebUI_Window(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		wantErr  bool
	}{
		{
			name:     "window with title",
			input:    "webui.window [title: \"Test\"]",
			expected: "webui-window!",
			wantErr:  false,
		},
		{
			name:     "window with dimensions",
			input:    "webui.window [width: 800 height: 600]",
			expected: "webui-window!",
			wantErr:  false,
		},
		{
			name:     "window with debug flag",
			input:    "webui.window [debug?: true]",
			expected: "webui-window!",
			wantErr:  false,
		},
		{
			name:     "window with resizable flag",
			input:    "webui.window [resizable?: false]",
			expected: "webui-window!",
			wantErr:  false,
		},
		{
			name:     "window with icon",
			input:    "webui.window [icon: \"icon.png\"]",
			expected: "webui-window!",
			wantErr:  false,
		},
		{
			name:     "window with source",
			input:    "webui.window [source: \"index.html\"]",
			expected: "webui-window!",
			wantErr:  false,
		},
		{
			name:     "window with html",
			input:    "webui.window [html: \"<html></html>\"]",
			expected: "webui-window!",
			wantErr:  false,
		},
		{
			name:     "invalid spec type",
			input:    "webui.window \"invalid\"",
			expected: "",
			wantErr:  true,
		},
		{
			name:     "invalid spec key",
			input:    "webui.window [invalid: \"key\"]",
			expected: "",
			wantErr:  false, // Should ignore unknown keys
		},
		{
			name:     "invalid title type",
			input:    "webui.window [title: 123]",
			expected: "",
			wantErr:  true,
		},
		{
			name:     "odd spec length",
			input:    "webui.window [title: \"Test\" extra]",
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
			name:     "send block",
			input:    "w: webui.window [] webui.send w \"msg\" [1 2 3]",
			expected: "true",
			wantErr:  false,
		},
		{
			name:     "send binary",
			input:    "w: webui.window [] webui.send w \"msg\" #{010203}",
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
			name:     "register handler with selector block",
			input:    "w: webui.window [] webui.on w \"click\" [\"#btn\" \".link\"] [print event-name]",
			expected: "true",
			wantErr:  false,
		},
		{
			name:     "invalid selector type",
			input:    "w: webui.window [] webui.on w \"click\" 123 [block]",
			expected: "",
			wantErr:  true,
		},
		{
			name:     "invalid selector in block",
			input:    "w: webui.window [] webui.on w \"click\" [123] [block]",
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
		{
			name:     "ready reset after consecutive renders",
			input:    "w: webui.window [] webui.render w \"<html></html>\" webui.ready? w webui.render w \"<html>2</html>\" webui.ready? w",
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
