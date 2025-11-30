//go:build webui

package native

import (
	"github.com/marcin-radoszewski/viro/internal/core"
	"github.com/marcin-radoszewski/viro/internal/value"
)

func RegisterWebUINatives(rootFrame core.Frame) {
	webuiObj := value.NewObject(rootFrame)
	webuiObj.Frame.Bind("start", value.NewFuncVal(value.NewNativeFunction(
		"webui.start",
		[]value.ParamSpec{},
		func(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
			return WebUIStart(args, refValues, eval)
		},
		false,
		&NativeDoc{
			Category: "WebUI",
			Summary:  "Initialize the WebUI subsystem",
			Description: `Initializes the WebUI subsystem.
This is called automatically by other webui.* functions.`,
			Returns:  "[none!] None",
			Examples: []string{"webui.start  ; initializes WebUI subsystem"},
			Tags:     []string{"webui", "initialization"},
		},
	)))

	webuiObj.Frame.Bind("window", value.NewFuncVal(value.NewNativeFunction(
		"webui.window",
		[]value.ParamSpec{
			value.NewParamSpec("spec", false),
		},
		func(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
			return WebUIWindow(args, refValues, eval)
		},
		false,
		&NativeDoc{
			Category: "WebUI",
			Summary:  "Create a new WebUI window",
			Description: `Creates a new WebUI window with the specified configuration.
Returns a webui-window! value that can be used with other webui functions.`,
			Parameters: []ParamDoc{
				{Name: "spec", Type: "block!", Description: "Window specification block", Optional: false},
			},
			Returns:  "[webui-window!] Window handle",
			Examples: []string{"webui.window [title: \"My App\" width: 800 height: 600]"},
			Tags:     []string{"webui", "window", "gui"},
		},
	)))

	webuiObj.Frame.Bind("render", value.NewFuncVal(value.NewNativeFunction(
		"webui.render",
		[]value.ParamSpec{
			value.NewParamSpec("window", true),
			value.NewParamSpec("markup", false),
			{Name: "options", Type: value.TypeBlock, Optional: true, Refinement: false, TakesValue: false, Eval: true},
		},
		func(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
			return WebUIRender(args, refValues, eval)
		},
		false,
		&NativeDoc{
			Category: "WebUI",
			Summary:  "Render HTML content in a window",
			Description: `Replaces the current DOM with the provided markup.
Supports string!, binary!, and file! markup with optional content-type validation.`,
			Parameters: []ParamDoc{
				{Name: "window", Type: "webui-window!", Description: "Window to render in", Optional: false},
				{Name: "markup", Type: "string! binary! file!", Description: "HTML content to render", Optional: false},
				{Name: "options", Type: "block!", Description: "Render options", Optional: true},
			},
			Returns:  "[logic!] Success status",
			Examples: []string{"webui.render window \"<h1>Hello</h1>\""},
			Tags:     []string{"webui", "render", "html"},
		},
	)))

	webuiObj.Frame.Bind("inject", value.NewFuncVal(value.NewNativeFunction(
		"webui.inject",
		[]value.ParamSpec{
			value.NewParamSpec("window", true),
			value.NewParamSpec("html", false),
		},
		func(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
			return WebUIInject(args, refValues, eval)
		},
		false,
		&NativeDoc{
			Category:    "WebUI",
			Summary:     "Inject HTML snippet into window",
			Description: `Evaluates HTML snippet in the existing page context without resetting handlers.`,
			Parameters: []ParamDoc{
				{Name: "window", Type: "webui-window!", Description: "Window to inject into", Optional: false},
				{Name: "html", Type: "string! binary!", Description: "HTML snippet", Optional: false},
			},
			Returns:  "[logic!] Success status",
			Examples: []string{"webui.inject window \"<div>New content</div>\""},
			Tags:     []string{"webui", "inject", "html"},
		},
	)))

	webuiObj.Frame.Bind("send", value.NewFuncVal(value.NewNativeFunction(
		"webui.send",
		[]value.ParamSpec{
			value.NewParamSpec("window", true),
			value.NewParamSpec("message", false),
			value.NewParamSpec("payload", false),
		},
		func(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
			return WebUISend(args, refValues, eval)
		},
		false,
		&NativeDoc{
			Category: "WebUI",
			Summary:  "Send data to JavaScript",
			Description: `Pushes data to the JavaScript side of the WebUI window.
Payload is JSON-encoded automatically.`,
			Parameters: []ParamDoc{
				{Name: "window", Type: "webui-window!", Description: "Target window", Optional: false},
				{Name: "message", Type: "string!", Description: "Message channel name", Optional: false},
				{Name: "payload", Type: "any-type!", Description: "Data to send", Optional: false},
			},
			Returns:  "[logic!] Success status",
			Examples: []string{"webui.send window \"update\" [data: \"value\"]"},
			Tags:     []string{"webui", "send", "javascript"},
		},
	)))

	webuiObj.Frame.Bind("on", value.NewFuncVal(value.NewNativeFunction(
		"webui.on",
		[]value.ParamSpec{
			value.NewParamSpec("window", true),
			value.NewParamSpec("event", false),
			value.NewParamSpec("selector", false),
			value.NewParamSpec("handler", false),
		},
		func(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
			return WebUIOn(args, refValues, eval)
		},
		false,
		&NativeDoc{
			Category: "WebUI",
			Summary:  "Register event handler",
			Description: `Registers a callback for browser events.
 Handler block receives event-name, event-selector, event-payload, and event-window bindings.`,
			Parameters: []ParamDoc{
				{Name: "window", Type: "webui-window!", Description: "Window to listen on", Optional: false},
				{Name: "event", Type: "string!", Description: "Event name", Optional: false},
				{Name: "selector", Type: "string! block!", Description: "CSS selector", Optional: false},
				{Name: "handler", Type: "block!", Description: "Handler code", Optional: false},
			},
			Returns:  "[logic!] Success status",
			Examples: []string{"webui.on window \"click\" \"#btn\" [print event-name]"},
			Tags:     []string{"webui", "event", "handler"},
		},
	)))

	webuiObj.Frame.Bind("close", value.NewFuncVal(value.NewNativeFunction(
		"webui.close",
		[]value.ParamSpec{
			value.NewParamSpec("window", true),
		},
		func(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
			return WebUIClose(args, refValues, eval)
		},
		false,
		&NativeDoc{
			Category:    "WebUI",
			Summary:     "Close a window",
			Description: `Closes the window and removes event registrations.`,
			Parameters: []ParamDoc{
				{Name: "window", Type: "webui-window!", Description: "Window to close", Optional: false},
			},
			Returns:  "[logic!] Whether window was closed",
			Examples: []string{"webui.close window"},
			Tags:     []string{"webui", "close", "window"},
		},
	)))

	webuiObj.Frame.Bind("ready?", value.NewFuncVal(value.NewNativeFunction(
		"webui.ready?",
		[]value.ParamSpec{
			value.NewParamSpec("window", true),
		},
		func(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
			return WebUIReady(args, refValues, eval)
		},
		false,
		&NativeDoc{
			Category: "WebUI",
			Summary:  "Register event handler",
			Description: `Registers a callback for browser events.
Handler block receives event-name, event-selector, event-payload, and event-window bindings.`,
			Parameters: []ParamDoc{
				{Name: "window", Type: "webui-window!", Description: "Window to listen on", Optional: false},
				{Name: "event", Type: "string!", Description: "Event name", Optional: false},
				{Name: "selector", Type: "string! block!", Description: "CSS selector", Optional: false},
				{Name: "handler", Type: "block!", Description: "Handler code", Optional: false},
			},
			Returns:  "[logic!] Success status",
			Examples: []string{"webui.on window \"click\" \"#btn\" [print event-name]"},
			Tags:     []string{"webui", "event", "handler"},
		},
	)))

	rootFrame.Bind("webui", webuiObj)
}
