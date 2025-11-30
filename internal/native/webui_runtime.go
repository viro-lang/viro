//go:build webui

package native

import (
	"sync"

	ui "github.com/webui-dev/go-webui/v2"
)

type WebUIWindow interface {
	Show(content string) error
	ShowBrowser(content string, browser ui.Browser) error
}

var webuiInitOnce sync.Once

func ensureWebUIInitialized() {
	webuiInitOnce.Do(func() {
		ui.SetConfig(ui.ShowWaitConnection, false)
	})
}

func ShowWindow(win WebUIWindow, content string) (bool, error) {
	ensureWebUIInitialized()
	err := win.Show(content)
	return err == nil, err
}

func ShowBrowserWindow(win WebUIWindow, content string, browser ui.Browser) (bool, error) {
	ensureWebUIInitialized()
	err := win.ShowBrowser(content, browser)
	return err == nil, err
}
