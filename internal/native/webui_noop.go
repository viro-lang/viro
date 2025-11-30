//go:build !webui

package native

import (
	"github.com/marcin-radoszewski/viro/internal/core"
	"github.com/marcin-radoszewski/viro/internal/value"
	"github.com/marcin-radoszewski/viro/internal/verror"
)

func WebUIStart(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	return value.NewNoneVal(), verror.NewScriptError("webui", [3]string{verror.ErrIDWebUIUnavailable, "WebUI support not compiled in", ""})
}

func WebUIWindow(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	return value.NewNoneVal(), verror.NewScriptError("webui", [3]string{verror.ErrIDWebUIUnavailable, "WebUI support not compiled in", ""})
}

func WebUIRender(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	return value.NewNoneVal(), verror.NewScriptError("webui", [3]string{verror.ErrIDWebUIUnavailable, "WebUI support not compiled in", ""})
}

func WebUISend(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	return value.NewNoneVal(), verror.NewScriptError("webui", [3]string{verror.ErrIDWebUIUnavailable, "WebUI support not compiled in", ""})
}

func WebUIOn(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	return value.NewNoneVal(), verror.NewScriptError("webui", [3]string{verror.ErrIDWebUIUnavailable, "WebUI support not compiled in", ""})
}

func WebUIPoll(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	return value.NewNoneVal(), verror.NewScriptError("webui", [3]string{verror.ErrIDWebUIUnavailable, "WebUI support not compiled in", ""})
}

func WebUIClose(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	return value.NewNoneVal(), verror.NewScriptError("webui", [3]string{verror.ErrIDWebUIUnavailable, "WebUI support not compiled in", ""})
}

func WebUIReady(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	return value.NewNoneVal(), verror.NewScriptError("webui", [3]string{verror.ErrIDWebUIUnavailable, "WebUI support not compiled in", ""})
}

func WebUIInject(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	return value.NewNoneVal(), verror.NewScriptError("webui", [3]string{verror.ErrIDWebUIUnavailable, "WebUI support not compiled in", ""})
}
