package native

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/marcin-radoszewski/viro/internal/core"
	"github.com/marcin-radoszewski/viro/internal/frame"
	"github.com/marcin-radoszewski/viro/internal/value"
	"github.com/marcin-radoszewski/viro/internal/verror"
	ui "github.com/webui-dev/go-webui/v2"
)

var (
	runtime     *webuiRuntime
	runtimeOnce sync.Once
)

type webuiRuntime struct {
	mu      sync.RWMutex
	windows map[uint32]*webuiWindow
	nextID  uint32
}

type webuiWindow struct {
	id       uint32
	window   ui.Window
	handlers map[string][]handlerEntry
}

type handlerEntry struct {
	event         string
	selector      string
	handlerBlock  *value.BlockValue
	capturedFrame int
}

func getWebUIRuntime() *webuiRuntime {
	runtimeOnce.Do(func() {
		runtime = &webuiRuntime{
			windows: make(map[uint32]*webuiWindow),
		}
	})
	return runtime
}

func WebUIStart(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 0 {
		return value.NewNoneVal(), arityError("webui.start", 0, len(args))
	}

	return value.NewNoneVal(), nil
}

func WebUIWindow(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 1 {
		return value.NewNoneVal(), arityError("webui.window", 1, len(args))
	}

	block, ok := value.AsBlockValue(args[0])
	if !ok {
		return value.NewNoneVal(), typeError("webui.window", "block", args[0])
	}

	rt := getWebUIRuntime()
	rt.mu.Lock()
	defer rt.mu.Unlock()

	rt.nextID++
	id := rt.nextID

	window := ui.NewWindow()

	spec := make(map[string]interface{})

	for i := 0; i < len(block.Elements)-1; i += 2 {
		key := block.Elements[i]
		val := block.Elements[i+1]

		evaluatedVal, err := eval.DoBlock([]core.Value{val}, nil)
		if err != nil {
			return value.NewNoneVal(), err
		}

		word, ok := value.AsWordValue(key)
		if !ok {
			return value.NewNoneVal(), verror.NewScriptError("webui", [3]string{"invalid-spec-key", "keys must be words", ""})
		}

		switch word {
		case "title":
			if str, ok := value.AsStringValue(evaluatedVal); ok {
				spec["title"] = str.String()
			} else {
				return value.NewNoneVal(), verror.NewScriptError("webui", [3]string{"invalid-spec-value", "title must be string!", ""})
			}
		case "width":
			if intVal, ok := value.AsIntValue(evaluatedVal); ok {
				spec["width"] = int(intVal)
			} else {
				return value.NewNoneVal(), verror.NewScriptError("webui", [3]string{"invalid-spec-value", "width must be integer!", ""})
			}
		case "height":
			if intVal, ok := value.AsIntValue(evaluatedVal); ok {
				spec["height"] = int(intVal)
			} else {
				return value.NewNoneVal(), verror.NewScriptError("webui", [3]string{"invalid-spec-value", "height must be integer!", ""})
			}
		case "resizable?":
			if evaluatedVal.GetType() == value.TypeLogic {
				if logic, ok := value.AsLogicValue(evaluatedVal); ok {
					spec["resizable"] = bool(logic)
				}
			} else {
				return value.NewNoneVal(), verror.NewScriptError("webui", [3]string{"invalid-spec-value", "resizable? must be logic!", ""})
			}
		case "icon":
			if str, ok := value.AsStringValue(evaluatedVal); ok {
				spec["icon"] = str.String()
			} else {
				return value.NewNoneVal(), verror.NewScriptError("webui", [3]string{"invalid-spec-value", "icon must be string!", ""})
			}
		case "source":
			if str, ok := value.AsStringValue(evaluatedVal); ok {
				spec["source"] = str.String()
			} else {
				return value.NewNoneVal(), verror.NewScriptError("webui", [3]string{"invalid-spec-value", "source must be string!", ""})
			}
		case "html":
			if str, ok := value.AsStringValue(evaluatedVal); ok {
				spec["html"] = str.String()
			} else if bin, ok := value.AsBinaryValue(evaluatedVal); ok {
				spec["html"] = string(bin.Bytes())
			} else {
				return value.NewNoneVal(), verror.NewScriptError("webui", [3]string{"invalid-spec-value", "html must be string! or binary!", ""})
			}
		default:
			return value.NewNoneVal(), verror.NewScriptError("webui", [3]string{"invalid-spec-key", word + " is not a valid window spec key", ""})
		}
	}

	if len(block.Elements)%2 != 0 {
		return value.NewNoneVal(), verror.NewScriptError("webui", [3]string{"invalid-spec", "spec block must have even number of elements", ""})
	}

	webuiWin := &webuiWindow{
		id:       id,
		window:   window,
		handlers: make(map[string][]handlerEntry),
	}

	rt.windows[id] = webuiWin

	// Note: go-webui doesn't support title, size, resizable, or icon settings directly
	// These would need to be set via JavaScript after showing the window

	if html, ok := spec["html"].(string); ok {
		window.Show(html)
	} else if source, ok := spec["source"].(string); ok {
		window.Show(source)
	} else {
		window.Show("")
	}

	return value.WebUIWindowVal(value.NewWebUIWindow(id)), nil
}

func WebUIRender(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 3 {
		return value.NewNoneVal(), arityError("webui.render", 3, len(args))
	}

	window, ok := value.AsWebUIWindow(args[0])
	if !ok {
		return value.NewNoneVal(), typeError("webui.render", "webui-window", args[0])
	}

	rt := getWebUIRuntime()
	rt.mu.RLock()
	webuiWin, exists := rt.windows[window.ID]
	rt.mu.RUnlock()

	if !exists {
		return value.NewNoneVal(), verror.NewScriptError("webui", [3]string{verror.ErrIDWindowNotFound, "", ""})
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
		var ct string
		if wordVal, ok := value.AsWordValue(contentType); ok {
			ct = wordVal
		} else if strVal, ok := value.AsStringValue(contentType); ok {
			ct = strVal.String()
		} else {
			return value.NewNoneVal(), verror.NewScriptError("webui", [3]string{"invalid-content-type", "must be word! or string!", ""})
		}

		validTypes := []string{"text/html", "text/plain", "application/xhtml+xml", "application/json", "text/javascript"}
		valid := false
		for _, vt := range validTypes {
			if strings.EqualFold(ct, vt) {
				valid = true
				break
			}
		}
		if !valid {
			return value.NewNoneVal(), verror.NewScriptError("webui", [3]string{"invalid-content-type", ct, ""})
		}
	} else if markup.GetType() == value.TypeBinary {
		return value.NewNoneVal(), verror.NewScriptError("webui", [3]string{"content-type-required", "", ""})
	}

	err := webuiWin.window.Show(content)
	if err != nil {
		return value.NewNoneVal(), verror.NewScriptError("webui", [3]string{"render-failed", err.Error(), ""})
	}

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

	rt := getWebUIRuntime()
	rt.mu.RLock()
	webuiWin, exists := rt.windows[window.ID]
	rt.mu.RUnlock()

	if !exists {
		return value.NewNoneVal(), verror.NewScriptError("webui", [3]string{verror.ErrIDWindowNotFound, "", ""})
	}

	message, ok := value.AsStringValue(args[1])
	if !ok {
		return value.NewNoneVal(), typeError("webui.send", "string", args[1])
	}

	payload := args[2]

	// Convert payload to JSON
	payloadJSON, err := value.ToJSON(payload)
	if err != nil {
		return value.NewNoneVal(), verror.NewScriptError("webui", [3]string{"send-failed", "cannot convert payload to JSON: " + err.Error(), ""})
	}

	// Send via JavaScript
	js := fmt.Sprintf("if (window.webui && window.webui['%s']) { window.webui['%s'](%s); }", message.String(), message.String(), payloadJSON)
	webuiWin.window.Run(js)

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

	rt := getWebUIRuntime()
	rt.mu.Lock()
	webuiWin, exists := rt.windows[window.ID]
	if !exists {
		rt.mu.Unlock()
		return value.NewNoneVal(), verror.NewScriptError("webui", [3]string{verror.ErrIDWindowNotFound, "", ""})
	}

	event, ok := value.AsStringValue(args[1])
	if !ok {
		rt.mu.Unlock()
		return value.NewNoneVal(), typeError("webui.on", "string", args[1])
	}

	selector := args[2] // Can be string or block
	if selector.GetType() != value.TypeString && selector.GetType() != value.TypeBlock {
		rt.mu.Unlock()
		return value.NewNoneVal(), typeError("webui.on", "string!|block", selector)
	}

	handler, ok := value.AsBlockValue(args[3])
	if !ok {
		rt.mu.Unlock()
		return value.NewNoneVal(), typeError("webui.on", "block", args[3])
	}

	frameIndex := eval.CurrentFrameIndex()

	selectors := []string{}
	if selector.GetType() == value.TypeString {
		if str, ok := value.AsStringValue(selector); ok {
			selectors = []string{str.String()}
		}
	} else if selector.GetType() == value.TypeBlock {
		if block, ok := value.AsBlockValue(selector); ok {
			for _, elem := range block.Elements {
				if str, ok := value.AsStringValue(elem); ok {
					selectors = append(selectors, str.String())
				} else {
					rt.mu.Unlock()
					return value.NewNoneVal(), verror.NewScriptError("webui", [3]string{"invalid-selector", "selector block elements must be strings", ""})
				}
			}
		}
	}

	for _, sel := range selectors {
		entry := handlerEntry{
			event:         event.String(),
			selector:      sel,
			handlerBlock:  handler,
			capturedFrame: frameIndex,
		}

		eventKey := event.String() + ":" + sel
		if webuiWin.handlers[eventKey] == nil {
			webuiWin.handlers[eventKey] = []handlerEntry{}
		}
		webuiWin.handlers[eventKey] = append(webuiWin.handlers[eventKey], entry)

		// Bind the event to go-webui
		webuiWin.window.Bind(sel, func(e ui.Event) any {
			rt.mu.RLock()
			entries, exists := webuiWin.handlers[eventKey]
			rt.mu.RUnlock()

			if !exists {
				return nil
			}

			for _, entry := range entries {
				childFrame := frame.NewFrameWithCapacity(frame.FrameClosure, entry.capturedFrame, 4)
				childFrame.Bind("event-name", value.NewStrVal(entry.event))
				childFrame.Bind("event-selector", value.NewStrVal(entry.selector))
				payload, _ := ui.GetArg[string](e)
				childFrame.Bind("event-payload", value.NewStrVal(payload))
				childFrame.Bind("event-payload", value.NewStrVal(payload))
				childFrame.Bind("event-window", value.WebUIWindowVal(value.NewWebUIWindow(window.ID)))

				rt.mu.Unlock() // Release lock before executing handler
				eval.PushFrameContext(childFrame)
				_, err := eval.DoBlock(entry.handlerBlock.Elements, entry.handlerBlock.Locations())
				eval.PopFrameContext()
				rt.mu.Lock() // Re-acquire lock

				if err != nil {
					// Log error but continue
					continue
				}
			}
			return nil
		})
	}

	rt.mu.Unlock()
	return value.NewLogicVal(true), nil
}

func WebUIPoll(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 1 {
		return value.NewNoneVal(), arityError("webui.poll", 1, len(args))
	}

	if args[0].GetType() == value.TypeNone {
		// When called with none, wait for all windows to close
		ui.Wait()
		return value.NewNoneVal(), nil
	} else {
		// When called with a specific window, no-op (per-window wait not available in go-webui)
		window, ok := value.AsWebUIWindow(args[0])
		if !ok {
			return value.NewNoneVal(), typeError("webui.poll", "webui-window", args[0])
		}
		// Check if window exists
		rt := getWebUIRuntime()
		rt.mu.RLock()
		_, exists := rt.windows[window.ID]
		rt.mu.RUnlock()
		if !exists {
			return value.NewNoneVal(), verror.NewScriptError("webui", [3]string{verror.ErrIDWindowNotFound, "", ""})
		}
		// No-op for specific window
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

	rt := getWebUIRuntime()
	rt.mu.Lock()
	webuiWin, exists := rt.windows[window.ID]
	if exists {
		webuiWin.window.Close()
		delete(rt.windows, window.ID)
	}
	rt.mu.Unlock()

	return value.NewLogicVal(exists), nil
}

func WebUIReady(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 1 {
		return value.NewNoneVal(), arityError("webui.ready?", 1, len(args))
	}

	window, ok := value.AsWebUIWindow(args[0])
	if !ok {
		return value.NewNoneVal(), typeError("webui.ready?", "webui-window", args[0])
	}

	rt := getWebUIRuntime()
	rt.mu.RLock()
	webuiWin, exists := rt.windows[window.ID]
	rt.mu.RUnlock()

	if !exists {
		return value.NewNoneVal(), verror.NewScriptError("webui", [3]string{verror.ErrIDWindowNotFound, "", ""})
	}

	ready := webuiWin.window.IsShown()
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

	rt := getWebUIRuntime()
	rt.mu.RLock()
	webuiWin, exists := rt.windows[window.ID]
	rt.mu.RUnlock()

	if !exists {
		return value.NewNoneVal(), verror.NewScriptError("webui", [3]string{verror.ErrIDWindowNotFound, "", ""})
	}

	html := args[1]
	var htmlStr string
	if str, ok := value.AsStringValue(html); ok {
		htmlStr = str.String()
	} else if bin, ok := value.AsBinaryValue(html); ok {
		htmlStr = string(bin.Bytes())
	} else {
		return value.NewNoneVal(), typeError("webui.inject", "string!|binary", html)
	}

	// Inject HTML by running JavaScript
	htmlJSON, _ := json.Marshal(htmlStr)
	js := fmt.Sprintf("document.body.insertAdjacentHTML('beforeend', %s);", string(htmlJSON))
	webuiWin.window.Run(js)

	return value.NewLogicVal(true), nil
}
