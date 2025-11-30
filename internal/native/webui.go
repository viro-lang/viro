//go:build webui

package native

import (
	"sync"

	"github.com/marcin-radoszewski/viro/internal/core"
	"github.com/marcin-radoszewski/viro/internal/frame"
	"github.com/marcin-radoszewski/viro/internal/value"
	ui "github.com/webui-dev/go-webui/v2"
)

var webuiHandlers = make(map[uint]map[string]core.Value)

var webuiMutex sync.Mutex

func WebUINewWindow(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 0 {
		return value.NewNoneVal(), arityError("webui-new-window", 0, len(args))
	}

	window := ui.NewWindow()
	return value.NewIntVal(int64(window)), nil
}

func WebUINewWindowId(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 0 {
		return value.NewNoneVal(), arityError("webui-new-window-id", 0, len(args))
	}

	result := ui.NewWindowId()
	return value.NewIntVal(int64(result)), nil
}

func WebUIShow(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 2 {
		return value.NewNoneVal(), arityError("webui-show", 2, len(args))
	}

	windowID, ok := value.AsIntValue(args[0])
	if !ok {
		return value.NewNoneVal(), typeError("webui-show", "integer!", args[0])
	}

	content, ok := value.AsStringValue(args[1])
	if !ok {
		return value.NewNoneVal(), typeError("webui-show", "string!", args[1])
	}

	window := ui.Window(uint(windowID))
	err := window.Show(content.String())
	if err != nil {
		return value.NewLogicVal(false), nil
	}
	return value.NewLogicVal(true), nil
}

func WebUIShowBrowser(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 3 {
		return value.NewNoneVal(), arityError("webui-show-browser", 3, len(args))
	}

	windowID, ok := value.AsIntValue(args[0])
	if !ok {
		return value.NewNoneVal(), typeError("webui-show-browser", "integer!", args[0])
	}

	content, ok := value.AsStringValue(args[1])
	if !ok {
		return value.NewNoneVal(), typeError("webui-show-browser", "string!", args[1])
	}

	browserVal, ok := value.AsIntValue(args[2])
	if !ok {
		return value.NewNoneVal(), typeError("webui-show-browser", "integer!", args[2])
	}

	window := ui.Window(uint(windowID))
	err := window.ShowBrowser(content.String(), ui.Browser(uint(browserVal)))
	if err != nil {
		return value.NewLogicVal(false), nil
	}
	return value.NewLogicVal(true), nil
}

func WebUIBind(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 3 {
		return value.NewNoneVal(), arityError("webui-bind", 3, len(args))
	}

	windowID, ok := value.AsIntValue(args[0])
	if !ok {
		return value.NewNoneVal(), typeError("webui-bind", "integer!", args[0])
	}

	element, ok := value.AsStringValue(args[1])
	if !ok {
		return value.NewNoneVal(), typeError("webui-bind", "string!", args[1])
	}

	handler, ok := value.AsBlockValue(args[2])
	if !ok {
		return value.NewNoneVal(), typeError("webui-bind", "block!", args[2])
	}

	wid := uint(uint(windowID))
	if webuiHandlers[wid] == nil {
		webuiHandlers[wid] = make(map[string]core.Value)
	}
	webuiHandlers[wid][element.String()] = handler

	window := ui.Window(uint(windowID))
	frameIndex := eval.CurrentFrameIndex()

	window.Bind(element.String(), func(e ui.Event) any {
		storedHandler, exists := webuiHandlers[wid][element.String()]
		if !exists {
			return nil
		}

		handlerBlock, ok := value.AsBlockValue(storedHandler)
		if !ok {
			return nil
		}

		childFrame := frame.NewFrameWithCapacity(frame.FrameClosure, frameIndex, 4)
		childFrame.Bind("event-element", value.NewStrVal(element.String()))

		if strArg, err := ui.GetArg[string](e); err == nil {
			childFrame.Bind("event-data", value.NewStrVal(strArg))
		} else {
			childFrame.Bind("event-data", value.NewNoneVal())
		}

		childFrame.Bind("event-window-id", value.NewIntVal(int64(wid)))
		childFrame.Bind("event-number", value.NewIntVal(0))

		webuiMutex.Lock()
		eval.PushFrameContext(childFrame)
		_, err := eval.DoBlock(handlerBlock.Elements, handlerBlock.Locations())
		eval.PopFrameContext()
		webuiMutex.Unlock()

		if err != nil {
			return nil
		}
		return nil
	})

	return value.NewLogicVal(true), nil
}

func WebUIRun(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 2 {
		return value.NewNoneVal(), arityError("webui-run", 2, len(args))
	}

	windowID, ok := value.AsIntValue(args[0])
	if !ok {
		return value.NewNoneVal(), typeError("webui-run", "integer!", args[0])
	}

	script, ok := value.AsStringValue(args[1])
	if !ok {
		return value.NewNoneVal(), typeError("webui-run", "string!", args[1])
	}

	window := ui.Window(uint(windowID))
	window.Run(script.String())
	return value.NewNoneVal(), nil
}

func WebUICloseDirect(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 1 {
		return value.NewNoneVal(), arityError("webui-close", 1, len(args))
	}

	windowID, ok := value.AsIntValue(args[0])
	if !ok {
		return value.NewNoneVal(), typeError("webui-close", "integer!", args[0])
	}

	window := ui.Window(uint(windowID))
	window.Close()
	return value.NewNoneVal(), nil
}

func WebUIDestroy(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 1 {
		return value.NewNoneVal(), arityError("webui-destroy", 1, len(args))
	}

	windowID, ok := value.AsIntValue(args[0])
	if !ok {
		return value.NewNoneVal(), typeError("webui-destroy", "integer!", args[0])
	}

	delete(webuiHandlers, uint(uint(windowID)))

	window := ui.Window(uint(windowID))
	window.Destroy()
	return value.NewNoneVal(), nil
}

func WebUIIsShown(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 1 {
		return value.NewNoneVal(), arityError("webui-is-shown", 1, len(args))
	}

	windowID, ok := value.AsIntValue(args[0])
	if !ok {
		return value.NewNoneVal(), typeError("webui-is-shown", "integer!", args[0])
	}

	window := ui.Window(uint(windowID))
	result := window.IsShown()
	return value.NewLogicVal(result), nil
}

func WebUISetTimeout(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 1 {
		return value.NewNoneVal(), arityError("webui-set-timeout", 1, len(args))
	}

	timeout, ok := value.AsIntValue(args[0])
	if !ok {
		return value.NewNoneVal(), typeError("webui-set-timeout", "integer!", args[0])
	}

	ui.SetTimeout(uint(timeout))
	return value.NewNoneVal(), nil
}

func WebUIWait(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 0 {
		return value.NewNoneVal(), arityError("webui-wait", 0, len(args))
	}

	ui.Wait()
	return value.NewNoneVal(), nil
}

func WebUIExit(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 0 {
		return value.NewNoneVal(), arityError("webui-exit", 0, len(args))
	}

	ui.Exit()
	return value.NewNoneVal(), nil
}

func WebUISetConfig(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 2 {
		return value.NewNoneVal(), arityError("webui-set-config", 2, len(args))
	}

	option, ok := value.AsIntValue(args[0])
	if !ok {
		return value.NewNoneVal(), typeError("webui-set-config", "integer!", args[0])
	}

	valueArg, ok := value.AsLogicValue(args[1])
	if !ok {
		return value.NewNoneVal(), typeError("webui-set-config", "logic!", args[1])
	}

	ui.SetConfig(ui.Config(uint(option)), valueArg)
	return value.NewNoneVal(), nil
}

func WebUIGetParentProcessId(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 1 {
		return value.NewNoneVal(), arityError("webui-get-parent-process-id", 1, len(args))
	}

	windowID, ok := value.AsIntValue(args[0])
	if !ok {
		return value.NewNoneVal(), typeError("webui-get-parent-process-id", "integer!", args[0])
	}

	window := ui.Window(uint(windowID))
	result := window.GetParentProcessID()
	return value.NewIntVal(int64(result)), nil
}

func WebUISetRootFolder(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 2 {
		return value.NewNoneVal(), arityError("webui-set-root-folder", 2, len(args))
	}

	windowID, ok := value.AsIntValue(args[0])
	if !ok {
		return value.NewNoneVal(), typeError("webui-set-root-folder", "integer!", args[0])
	}

	path, ok := value.AsStringValue(args[1])
	if !ok {
		return value.NewNoneVal(), typeError("webui-set-root-folder", "string!", args[1])
	}

	window := ui.Window(uint(windowID))
	window.SetRootFolder(path.String())
	return value.NewLogicVal(true), nil
}

func WebUISetDefaultRootFolder(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 1 {
		return value.NewNoneVal(), arityError("webui-set-default-root-folder", 1, len(args))
	}

	path, ok := value.AsStringValue(args[0])
	if !ok {
		return value.NewNoneVal(), typeError("webui-set-default-root-folder", "string!", args[0])
	}

	err := ui.SetDefaultRootFolder(path.String())
	if err != nil {
		return value.NewLogicVal(false), nil
	}
	return value.NewLogicVal(true), nil
}

func WebUIOpenUrl(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 1 {
		return value.NewNoneVal(), arityError("webui-open-url", 1, len(args))
	}

	url, ok := value.AsStringValue(args[0])
	if !ok {
		return value.NewNoneVal(), typeError("webui-open-url", "string!", args[0])
	}

	ui.OpenURL(url.String())
	return value.NewNoneVal(), nil
}

func WebUISetHide(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 2 {
		return value.NewNoneVal(), arityError("webui-set-hide", 2, len(args))
	}

	windowID, ok := value.AsIntValue(args[0])
	if !ok {
		return value.NewNoneVal(), typeError("webui-set-hide", "integer!", args[0])
	}

	hidden, ok := value.AsLogicValue(args[1])
	if !ok {
		return value.NewNoneVal(), typeError("webui-set-hide", "logic!", args[1])
	}

	window := ui.Window(uint(windowID))
	window.SetHide(hidden)
	return value.NewNoneVal(), nil
}

func WebUISetSize(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 3 {
		return value.NewNoneVal(), arityError("webui-set-size", 3, len(args))
	}

	windowID, ok := value.AsIntValue(args[0])
	if !ok {
		return value.NewNoneVal(), typeError("webui-set-size", "integer!", args[0])
	}

	width, ok := value.AsIntValue(args[1])
	if !ok {
		return value.NewNoneVal(), typeError("webui-set-size", "integer!", args[1])
	}

	height, ok := value.AsIntValue(args[2])
	if !ok {
		return value.NewNoneVal(), typeError("webui-set-size", "integer!", args[2])
	}

	window := ui.Window(uint(windowID))
	window.SetSize(uint(width), uint(height))
	return value.NewNoneVal(), nil
}

func WebUISetPosition(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 3 {
		return value.NewNoneVal(), arityError("webui-set-position", 3, len(args))
	}

	windowID, ok := value.AsIntValue(args[0])
	if !ok {
		return value.NewNoneVal(), typeError("webui-set-position", "integer!", args[0])
	}

	x, ok := value.AsIntValue(args[1])
	if !ok {
		return value.NewNoneVal(), typeError("webui-set-position", "integer!", args[1])
	}

	y, ok := value.AsIntValue(args[2])
	if !ok {
		return value.NewNoneVal(), typeError("webui-set-position", "integer!", args[2])
	}

	window := ui.Window(uint(windowID))
	window.SetPosition(uint(x), uint(y))
	return value.NewNoneVal(), nil
}

func WebUISetProfile(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 3 {
		return value.NewNoneVal(), arityError("webui-set-profile", 3, len(args))
	}

	windowID, ok := value.AsIntValue(args[0])
	if !ok {
		return value.NewNoneVal(), typeError("webui-set-profile", "integer!", args[0])
	}

	name, ok := value.AsStringValue(args[1])
	if !ok {
		return value.NewNoneVal(), typeError("webui-set-profile", "string!", args[1])
	}

	path, ok := value.AsStringValue(args[2])
	if !ok {
		return value.NewNoneVal(), typeError("webui-set-profile", "string!", args[2])
	}

	window := ui.Window(uint(windowID))
	window.SetProfile(name.String(), path.String())
	return value.NewNoneVal(), nil
}

func WebUIGetSize(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 1 {
		return value.NewNoneVal(), arityError("webui-get-size", 1, len(args))
	}

	_, ok := value.AsIntValue(args[0])
	if !ok {
		return value.NewNoneVal(), typeError("webui-get-size", "integer!", args[0])
	}

	elements := []core.Value{
		value.NewIntVal(800),
		value.NewIntVal(600),
	}
	return value.NewBlockVal(elements), nil
}

func WebUIGetPosition(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 1 {
		return value.NewNoneVal(), arityError("webui-get-position", 1, len(args))
	}

	_, ok := value.AsIntValue(args[0])
	if !ok {
		return value.NewNoneVal(), typeError("webui-get-position", "integer!", args[0])
	}

	elements := []core.Value{
		value.NewIntVal(100),
		value.NewIntVal(100),
	}
	return value.NewBlockVal(elements), nil
}

func WebUISetIcon(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 3 {
		return value.NewNoneVal(), arityError("webui-set-icon", 3, len(args))
	}

	windowID, ok := value.AsIntValue(args[0])
	if !ok {
		return value.NewNoneVal(), typeError("webui-set-icon", "integer!", args[0])
	}

	icon, ok := value.AsStringValue(args[1])
	if !ok {
		return value.NewNoneVal(), typeError("webui-set-icon", "string!", args[1])
	}

	iconType, ok := value.AsStringValue(args[2])
	if !ok {
		return value.NewNoneVal(), typeError("webui-set-icon", "string!", args[2])
	}

	window := ui.Window(uint(windowID))
	window.SetIcon(icon.String(), iconType.String())
	return value.NewNoneVal(), nil
}

func WebUISendRaw(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 3 {
		return value.NewNoneVal(), arityError("webui-send-raw", 3, len(args))
	}

	windowID, ok := value.AsIntValue(args[0])
	if !ok {
		return value.NewNoneVal(), typeError("webui-send-raw", "integer!", args[0])
	}

	function, ok := value.AsStringValue(args[1])
	if !ok {
		return value.NewNoneVal(), typeError("webui-send-raw", "string!", args[1])
	}

	rawData, ok := value.AsBinaryValue(args[2])
	if !ok {
		return value.NewNoneVal(), typeError("webui-send-raw", "binary!", args[2])
	}

	window := ui.Window(uint(windowID))
	window.SendRaw(function.String(), rawData.Bytes())
	return value.NewNoneVal(), nil
}
