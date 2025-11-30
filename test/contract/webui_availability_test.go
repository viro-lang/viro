package contract

import (
	"testing"

	"github.com/marcin-radoszewski/viro/internal/value"
)

func TestWebUIAvailability(t *testing.T) {
	evaluator := NewTestEvaluator()

	tests := []struct {
		name string
		fn   string
	}{
		{"webui-new-window", "webui-new-window"},
		{"webui-new-window-id", "webui-new-window-id"},
		{"webui-show", "webui-show"},
		{"webui-show-browser", "webui-show-browser"},
		{"webui-bind", "webui-bind"},
		{"webui-run", "webui-run"},
		{"webui-close", "webui-close"},
		{"webui-destroy", "webui-destroy"},
		{"webui-is-shown", "webui-is-shown"},
		{"webui-set-timeout", "webui-set-timeout"},
		{"webui-wait", "webui-wait"},
		{"webui-exit", "webui-exit"},
		{"webui-set-config", "webui-set-config"},
		{"webui-get-parent-process-id", "webui-get-parent-process-id"},
		{"webui-set-root-folder", "webui-set-root-folder"},
		{"webui-set-default-root-folder", "webui-set-default-root-folder"},
		{"webui-open-url", "webui-open-url"},
		{"webui-set-hide", "webui-set-hide"},
		{"webui-set-size", "webui-set-size"},
		{"webui-set-position", "webui-set-position"},
		{"webui-set-profile", "webui-set-profile"},
		{"webui-get-size", "webui-get-size"},
		{"webui-get-position", "webui-get-position"},
		{"webui-set-icon", "webui-set-icon"},
		{"webui-send-raw", "webui-send-raw"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn, ok := evaluator.GetFrameByIndex(0).Get(tt.fn)
			if !ok {
				t.Fatalf("WebUI function '%s' not found in root frame", tt.fn)
			}

			if fn.GetType() != value.TypeFunction {
				t.Errorf("expected '%s' to be a function, got %s", tt.fn, value.TypeToString(fn.GetType()))
			}

			funcVal := fn.(*value.FunctionValue)
			if funcVal.Native == nil {
				t.Errorf("expected '%s' to be a native function", tt.fn)
			}
		})
	}
}
