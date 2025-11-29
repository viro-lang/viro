package native

import (
	"github.com/marcin-radoszewski/viro/internal/core"
	"github.com/marcin-radoszewski/viro/internal/frame"
	"github.com/marcin-radoszewski/viro/internal/value"
	"github.com/marcin-radoszewski/viro/internal/verror"
	"github.com/marcin-radoszewski/viro/internal/webui"
)

func WebUIStart(args []core.Value, refValues map[string]core.Value, eval core.Evaluator, manager webui.Manager) (core.Value, error) {
	if len(args) != 0 {
		return value.NewNoneVal(), arityError("webui.start", 0, len(args))
	}

	return value.NewNoneVal(), nil
}

func WebUIWindow(args []core.Value, refValues map[string]core.Value, eval core.Evaluator, manager webui.Manager) (core.Value, error) {
	if len(args) != 1 {
		return value.NewNoneVal(), arityError("webui.window", 1, len(args))
	}

	_, ok := value.AsBlockValue(args[0])
	if !ok {
		return value.NewNoneVal(), typeError("webui.window", "block", args[0])
	}

	specMap := make(map[string]core.Value)

	id, err := manager.CreateWindow(specMap)
	if err != nil {
		return value.NewNoneVal(), verror.NewScriptError("webui", [3]string{"create-window-failed", err.Error(), ""})
	}

	return value.WebUIWindowVal(value.NewWebUIWindow(id)), nil
}

// WebUIRender renders HTML content to a window
func WebUIRender(args []core.Value, refValues map[string]core.Value, eval core.Evaluator, manager webui.Manager) (core.Value, error) {
	if len(args) != 3 {
		return value.NewNoneVal(), arityError("webui.render", 3, len(args))
	}

	window, ok := value.AsWebUIWindow(args[0])
	if !ok {
		return value.NewNoneVal(), typeError("webui.render", "webui-window", args[0])
	}

	markup := args[1]
	options := make(map[string]core.Value)
	if args[2].GetType() == value.TypeBlock {
		if opts, ok := value.AsBlockValue(args[2]); ok {
			for i := 0; i < len(opts.Elements)-1; i += 2 {
				if word, ok := value.AsWordValue(opts.Elements[i]); ok {
					options[word] = opts.Elements[i+1]
				}
			}
		}
	}

	if contentType, exists := options["content-type"]; exists {
		if ct, ok := value.AsWordValue(contentType); ok {
			validTypes := []string{"text/html", "text/plain", "application/xhtml+xml", "application/json", "text/javascript"}
			valid := false
			for _, vt := range validTypes {
				if ct == vt {
					valid = true
					break
				}
			}
			if !valid {
				return value.NewNoneVal(), verror.NewScriptError("webui", [3]string{"invalid-content-type", ct, ""})
			}
		}
	} else if markup.GetType() == value.TypeBinary {
		return value.NewNoneVal(), verror.NewScriptError("webui", [3]string{"content-type-required", "", ""})
	}

	err := manager.Render(window.ID, markup, options)
	if err != nil {
		return value.NewNoneVal(), verror.NewScriptError("webui", [3]string{"render-failed", err.Error(), ""})
	}

	return value.NewLogicVal(true), nil
}

// WebUISend sends data to JavaScript
func WebUISend(args []core.Value, refValues map[string]core.Value, eval core.Evaluator, manager webui.Manager) (core.Value, error) {
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

	err := manager.Send(window.ID, message.String(), payload)
	if err != nil {
		return value.NewNoneVal(), verror.NewScriptError("webui", [3]string{"send-failed", err.Error(), ""})
	}

	return value.NewLogicVal(true), nil
}

// WebUIOn registers an event handler
func WebUIOn(args []core.Value, refValues map[string]core.Value, eval core.Evaluator, manager webui.Manager) (core.Value, error) {
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

	selector := args[2] // Can be string or block
	if selector.GetType() != value.TypeString && selector.GetType() != value.TypeBlock {
		return value.NewNoneVal(), typeError("webui.on", "string!|block", selector)
	}

	handler, ok := value.AsBlockValue(args[3])
	if !ok {
		return value.NewNoneVal(), typeError("webui.on", "block", args[3])
	}

	frameIndex := eval.CurrentFrameIndex()

	entry := webui.HandlerEntry{
		WindowID:      window.ID,
		Event:         event.String(),
		Selector:      selector.Mold(),
		HandlerBlock:  handler,
		CapturedFrame: frameIndex,
	}

	err := manager.RegisterEvent(entry)
	if err != nil {
		return value.NewNoneVal(), verror.NewScriptError("webui", [3]string{"register-event-failed", err.Error(), ""})
	}

	return value.NewLogicVal(true), nil
}

// WebUIPoll processes pending events
func WebUIPoll(args []core.Value, refValues map[string]core.Value, eval core.Evaluator, manager webui.Manager) (core.Value, error) {
	if len(args) != 1 {
		return value.NewNoneVal(), arityError("webui.poll", 1, len(args))
	}

	var windowID *uint32
	if args[0].GetType() != value.TypeNone {
		window, ok := value.AsWebUIWindow(args[0])
		if !ok {
			return value.NewNoneVal(), typeError("webui.poll", "webui-window", args[0])
		}
		windowID = &window.ID
	}

	events, err := manager.Poll(windowID)
	if err != nil {
		return value.NewNoneVal(), verror.NewScriptError("webui", [3]string{"poll-failed", err.Error(), ""})
	}

	for _, event := range events {
		childFrame := frame.NewFrameWithCapacity(frame.FrameClosure, event.HandlerEntry.CapturedFrame, 4)
		childFrame.Bind("event-name", value.NewStrVal(event.HandlerEntry.Event))
		childFrame.Bind("event-selector", value.NewStrVal(event.HandlerEntry.Selector))
		childFrame.Bind("event-payload", event.Payload)
		childFrame.Bind("event-window", value.WebUIWindowVal(value.NewWebUIWindow(event.HandlerEntry.WindowID)))

		eval.PushFrameContext(childFrame)
		_, err := eval.DoBlock(event.HandlerEntry.HandlerBlock.Elements, event.HandlerEntry.HandlerBlock.Locations())
		eval.PopFrameContext()
		if err != nil {
			return value.NewNoneVal(), err
		}
	}

	return value.NewNoneVal(), nil
}

// WebUIClose closes a window
func WebUIClose(args []core.Value, refValues map[string]core.Value, eval core.Evaluator, manager webui.Manager) (core.Value, error) {
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

	closed := manager.Close(window.ID)
	return value.NewLogicVal(closed), nil
}

// WebUIReady checks if window is ready
func WebUIReady(args []core.Value, refValues map[string]core.Value, eval core.Evaluator, manager webui.Manager) (core.Value, error) {
	if len(args) != 1 {
		return value.NewNoneVal(), arityError("webui.ready?", 1, len(args))
	}

	window, ok := value.AsWebUIWindow(args[0])
	if !ok {
		return value.NewNoneVal(), typeError("webui.ready?", "webui-window", args[0])
	}

	ready := manager.Ready(window.ID)
	return value.NewLogicVal(ready), nil
}

// WebUIInject injects HTML snippet
func WebUIInject(args []core.Value, refValues map[string]core.Value, eval core.Evaluator, manager webui.Manager) (core.Value, error) {
	if len(args) != 2 {
		return value.NewNoneVal(), arityError("webui.inject", 2, len(args))
	}

	window, ok := value.AsWebUIWindow(args[0])
	if !ok {
		return value.NewNoneVal(), typeError("webui.inject", "webui-window", args[0])
	}

	html := args[1]
	if html.GetType() != value.TypeString && html.GetType() != value.TypeBinary {
		return value.NewNoneVal(), typeError("webui.inject", "string!|binary", html)
	}

	err := manager.Inject(window.ID, html)
	if err != nil {
		return value.NewNoneVal(), verror.NewScriptError("webui", [3]string{"inject-failed", err.Error(), ""})
	}

	return value.NewLogicVal(true), nil
}
