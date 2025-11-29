package value

import (
	"github.com/marcin-radoszewski/viro/internal/core"
)

type WebUIWindow struct {
	ID uint32
}

func NewWebUIWindow(id uint32) *WebUIWindow {
	return &WebUIWindow{ID: id}
}

func (w *WebUIWindow) GetType() core.ValueType {
	return TypeWebUIWindow
}

func (w *WebUIWindow) GetPayload() any {
	return w
}

func (w *WebUIWindow) String() string {
	return w.Mold()
}

func (w *WebUIWindow) Mold() string {
	return "webui-window!"
}

func (w *WebUIWindow) Form() string {
	return w.Mold()
}

func (w *WebUIWindow) Equals(other core.Value) bool {
	if other.GetType() != TypeWebUIWindow {
		return false
	}
	otherWindow, ok := other.GetPayload().(*WebUIWindow)
	if !ok {
		return false
	}
	return w.ID == otherWindow.ID
}

func WebUIWindowVal(window *WebUIWindow) core.Value {
	return window
}

func AsWebUIWindow(v core.Value) (*WebUIWindow, bool) {
	if v.GetType() != TypeWebUIWindow {
		return nil, false
	}
	window, ok := v.GetPayload().(*WebUIWindow)
	return window, ok
}
