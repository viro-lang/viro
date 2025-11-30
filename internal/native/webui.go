//go:build webui

package native

import (
	"github.com/marcin-radoszewski/viro/internal/core"
	"github.com/marcin-radoszewski/viro/internal/frame"
	"github.com/marcin-radoszewski/viro/internal/value"
	"github.com/marcin-radoszewski/viro/internal/verror"
	ui "github.com/webui-dev/go-webui/v2"
)

func WebUIWindow(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 0 {
		return value.NewNoneVal(), arityError("webui.window", 0, len(args))
	}

	window := ui.NewWindow()
	return value.WebUIWindowVal(value.NewWebUIWindow(window)), nil
}

func WebUIRender(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 2 {
		return value.NewNoneVal(), arityError("webui.render", 2, len(args))
	}

	window, ok := value.AsWebUIWindow(args[0])
	if !ok {
		return value.NewNoneVal(), typeError("webui.render", "webui-window", args[0])
	}

	markup := args[1]
	var content string

	if str, ok := value.AsStringValue(markup); ok {
		content = str.String()
	} else if bin, ok := value.AsBinaryValue(markup); ok {
		content = string(bin.Bytes())
	} else {
		return value.NewNoneVal(), typeError("webui.render", "string!|binary!", markup)
	}

	window.Window.Show(content)
	return value.NewLogicVal(true), nil
}

func WebUISend(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 3 {
		return value.NewNoneVal(), arityError("webui.send", 3, len(args))
	}

	window, ok := value.AsWebUIWindow(args[0])
	if !ok {
		return value.NewNoneVal(), typeError("webui.send", "webui-window", args[0])
	}

	message, ok := value.AsStringValue(args[1])
	if !ok {
		return value.NewNoneVal(), typeError("webui.send", "string", args[1])
	}

	payload := args[2]

	payloadJSON, err := value.ToJSON(payload)
	if err != nil {
		return value.NewNoneVal(), verror.NewScriptError("webui", [3]string{"send-failed", "cannot convert payload to JSON: " + err.Error(), ""})
	}

	window.Window.SendRaw(message.String(), []byte(payloadJSON))
	return value.NewLogicVal(true), nil
}

func WebUIOn(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 4 {
		return value.NewNoneVal(), arityError("webui.on", 4, len(args))
	}

	window, ok := value.AsWebUIWindow(args[0])
	if !ok {
		return value.NewNoneVal(), typeError("webui.on", "webui-window", args[0])
	}

	event, ok := value.AsStringValue(args[1])
	if !ok {
		return value.NewNoneVal(), typeError("webui.on", "string", args[1])
	}

	selector, ok := value.AsStringValue(args[2])
	if !ok {
		return value.NewNoneVal(), typeError("webui.on", "string", args[2])
	}

	handler, ok := value.AsBlockValue(args[3])
	if !ok {
		return value.NewNoneVal(), typeError("webui.on", "block", args[3])
	}

	frameIndex := eval.CurrentFrameIndex()

	window.Window.Bind(selector.String(), func(e ui.Event) any {
		childFrame := frame.NewFrameWithCapacity(frame.FrameClosure, frameIndex, 4)
		childFrame.Bind("event-name", value.NewStrVal(event.String()))
		childFrame.Bind("event-selector", value.NewStrVal(selector.String()))
		payload, _ := ui.GetArg[string](e)
		childFrame.Bind("event-payload", value.NewStrVal(payload))
		childFrame.Bind("event-window", value.WebUIWindowVal(window))

		eval.PushFrameContext(childFrame)
		_, err := eval.DoBlock(handler.Elements, handler.Locations())
		eval.PopFrameContext()

		if err != nil {
			return nil
		}
		return nil
	})

	return value.NewLogicVal(true), nil
}

func WebUIPoll(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 1 {
		return value.NewNoneVal(), arityError("webui.poll", 1, len(args))
	}

	if args[0].GetType() == value.TypeNone {
		ui.Wait()
		return value.NewNoneVal(), nil
	} else {
		// For specific window, no-op as per go-webui semantics
		_, ok := value.AsWebUIWindow(args[0])
		if !ok {
			return value.NewNoneVal(), typeError("webui.poll", "webui-window", args[0])
		}
		return value.NewNoneVal(), nil
	}
}

func WebUIClose(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 1 {
		return value.NewNoneVal(), arityError("webui.close", 1, len(args))
	}

	if args[0].GetType() == value.TypeNone {
		return value.NewLogicVal(false), nil
	}

	window, ok := value.AsWebUIWindow(args[0])
	if !ok {
		return value.NewNoneVal(), typeError("webui.close", "webui-window", args[0])
	}

	window.Window.Close()
	return value.NewLogicVal(true), nil
}

func WebUIReady(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 1 {
		return value.NewNoneVal(), arityError("webui.ready?", 1, len(args))
	}

	window, ok := value.AsWebUIWindow(args[0])
	if !ok {
		return value.NewNoneVal(), typeError("webui.ready?", "webui-window", args[0])
	}

	ready := window.Window.IsShown()
	return value.NewLogicVal(ready), nil
}

func WebUIInject(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 2 {
		return value.NewNoneVal(), arityError("webui.inject", 2, len(args))
	}

	window, ok := value.AsWebUIWindow(args[0])
	if !ok {
		return value.NewNoneVal(), typeError("webui.inject", "webui-window", args[0])
	}

	html := args[1]
	var script string
	if str, ok := value.AsStringValue(html); ok {
		script = str.String()
	} else if bin, ok := value.AsBinaryValue(html); ok {
		script = string(bin.Bytes())
	} else {
		return value.NewNoneVal(), typeError("webui.inject", "string!|binary", html)
	}

	window.Window.Run(script)
	return value.NewLogicVal(true), nil
}
