//go:build !webui

package native

import (
	"github.com/marcin-radoszewski/viro/internal/verror"
)

func ShowWindow(win interface{}, content string) (bool, error) {
	return false, verror.NewScriptError("webui", [3]string{verror.ErrIDWebUIUnavailable, "WebUI support not compiled in", ""})
}

func ShowBrowserWindow(win interface{}, content string, browser interface{}) (bool, error) {
	return false, verror.NewScriptError("webui", [3]string{verror.ErrIDWebUIUnavailable, "WebUI support not compiled in", ""})
}
