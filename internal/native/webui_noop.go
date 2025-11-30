//go:build !webui

package native

import (
	"github.com/marcin-radoszewski/viro/internal/core"
	"github.com/marcin-radoszewski/viro/internal/value"
	"github.com/marcin-radoszewski/viro/internal/verror"
)

func WebUINewWindow(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	return value.NewNoneVal(), verror.NewScriptError("webui", [3]string{verror.ErrIDWebUIUnavailable, "WebUI support not compiled in", ""})
}

func WebUINewWindowId(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	return value.NewNoneVal(), verror.NewScriptError("webui", [3]string{verror.ErrIDWebUIUnavailable, "WebUI support not compiled in", ""})
}

func WebUIShow(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	return value.NewNoneVal(), verror.NewScriptError("webui", [3]string{verror.ErrIDWebUIUnavailable, "WebUI support not compiled in", ""})
}

func WebUIShowBrowser(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	return value.NewNoneVal(), verror.NewScriptError("webui", [3]string{verror.ErrIDWebUIUnavailable, "WebUI support not compiled in", ""})
}

func WebUIBind(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	return value.NewNoneVal(), verror.NewScriptError("webui", [3]string{verror.ErrIDWebUIUnavailable, "WebUI support not compiled in", ""})
}

func WebUIRun(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	return value.NewNoneVal(), verror.NewScriptError("webui", [3]string{verror.ErrIDWebUIUnavailable, "WebUI support not compiled in", ""})
}

func WebUICloseDirect(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	return value.NewNoneVal(), verror.NewScriptError("webui", [3]string{verror.ErrIDWebUIUnavailable, "WebUI support not compiled in", ""})
}

func WebUIDestroy(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	return value.NewNoneVal(), verror.NewScriptError("webui", [3]string{verror.ErrIDWebUIUnavailable, "WebUI support not compiled in", ""})
}

func WebUIIsShown(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	return value.NewNoneVal(), verror.NewScriptError("webui", [3]string{verror.ErrIDWebUIUnavailable, "WebUI support not compiled in", ""})
}

func WebUISetTimeout(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	return value.NewNoneVal(), verror.NewScriptError("webui", [3]string{verror.ErrIDWebUIUnavailable, "WebUI support not compiled in", ""})
}

func WebUIWait(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	return value.NewNoneVal(), verror.NewScriptError("webui", [3]string{verror.ErrIDWebUIUnavailable, "WebUI support not compiled in", ""})
}

func WebUIExit(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	return value.NewNoneVal(), verror.NewScriptError("webui", [3]string{verror.ErrIDWebUIUnavailable, "WebUI support not compiled in", ""})
}

func WebUISetConfig(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	return value.NewNoneVal(), verror.NewScriptError("webui", [3]string{verror.ErrIDWebUIUnavailable, "WebUI support not compiled in", ""})
}

func WebUIGetParentProcessId(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	return value.NewNoneVal(), verror.NewScriptError("webui", [3]string{verror.ErrIDWebUIUnavailable, "WebUI support not compiled in", ""})
}

func WebUISetRootFolder(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	return value.NewNoneVal(), verror.NewScriptError("webui", [3]string{verror.ErrIDWebUIUnavailable, "WebUI support not compiled in", ""})
}

func WebUISetDefaultRootFolder(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	return value.NewNoneVal(), verror.NewScriptError("webui", [3]string{verror.ErrIDWebUIUnavailable, "WebUI support not compiled in", ""})
}

func WebUIOpenUrl(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	return value.NewNoneVal(), verror.NewScriptError("webui", [3]string{verror.ErrIDWebUIUnavailable, "WebUI support not compiled in", ""})
}

func WebUISetHide(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	return value.NewNoneVal(), verror.NewScriptError("webui", [3]string{verror.ErrIDWebUIUnavailable, "WebUI support not compiled in", ""})
}

func WebUISetSize(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	return value.NewNoneVal(), verror.NewScriptError("webui", [3]string{verror.ErrIDWebUIUnavailable, "WebUI support not compiled in", ""})
}

func WebUISetPosition(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	return value.NewNoneVal(), verror.NewScriptError("webui", [3]string{verror.ErrIDWebUIUnavailable, "WebUI support not compiled in", ""})
}

func WebUISetProfile(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	return value.NewNoneVal(), verror.NewScriptError("webui", [3]string{verror.ErrIDWebUIUnavailable, "WebUI support not compiled in", ""})
}

func WebUIGetSize(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	return value.NewNoneVal(), verror.NewScriptError("webui", [3]string{verror.ErrIDWebUIUnavailable, "WebUI support not compiled in", ""})
}

func WebUIGetPosition(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	return value.NewNoneVal(), verror.NewScriptError("webui", [3]string{verror.ErrIDWebUIUnavailable, "WebUI support not compiled in", ""})
}

func WebUISetIcon(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	return value.NewNoneVal(), verror.NewScriptError("webui", [3]string{verror.ErrIDWebUIUnavailable, "WebUI support not compiled in", ""})
}

func WebUISendRaw(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	return value.NewNoneVal(), verror.NewScriptError("webui", [3]string{verror.ErrIDWebUIUnavailable, "WebUI support not compiled in", ""})
}
