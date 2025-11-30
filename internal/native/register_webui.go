package native

import (
	"github.com/marcin-radoszewski/viro/internal/core"
	"github.com/marcin-radoszewski/viro/internal/value"
)

func RegisterWebUINatives(rootFrame core.Frame) {
	rootFrame.Bind("webui-new-window", value.NewFuncVal(value.NewNativeFunction(
		"webui-new-window",
		[]value.ParamSpec{},
		func(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
			return WebUINewWindow(args, refValues, eval)
		},
		false,
		&NativeDoc{
			Category:    "WebUI",
			Summary:     "Create a new WebUI window",
			Description: `Creates a new WebUI window and returns its window ID as an integer.`,
			Parameters:  []ParamDoc{},
			Returns:     "[integer!] Window ID",
			Examples:    []string{"window-id: webui-new-window"},
			Tags:        []string{"webui", "window", "create"},
		},
	)))

	rootFrame.Bind("webui-new-window-id", value.NewFuncVal(value.NewNativeFunction(
		"webui-new-window-id",
		[]value.ParamSpec{},
		func(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
			return WebUINewWindowId(args, refValues, eval)
		},
		false,
		&NativeDoc{
			Category:    "WebUI",
			Summary:     "Get a new WebUI window ID",
			Description: `Returns a new unique window ID that can be used to create a WebUI window.`,
			Parameters:  []ParamDoc{},
			Returns:     "[integer!] New window ID",
			Examples:    []string{"webui-new-window-id"},
			Tags:        []string{"webui", "window", "create"},
		},
	)))

	rootFrame.Bind("webui-show", value.NewFuncVal(value.NewNativeFunction(
		"webui-show",
		[]value.ParamSpec{
			value.NewParamSpec("window-id", true),
			value.NewParamSpec("content", true),
		},
		func(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
			return WebUIShow(args, refValues, eval)
		},
		false,
		&NativeDoc{
			Category:    "WebUI",
			Summary:     "Show window with HTML content",
			Description: `Shows the WebUI window with the specified HTML content.`,
			Parameters: []ParamDoc{
				{Name: "window-id", Type: "integer!", Description: "Window ID", Optional: false},
				{Name: "content", Type: "string!", Description: "HTML content", Optional: false},
			},
			Returns:  "[logic!] Success status",
			Examples: []string{"webui-show window-id \"<h1>Hello</h1>\""},
			Tags:     []string{"webui", "window", "show"},
		},
	)))

	rootFrame.Bind("webui-show-browser", value.NewFuncVal(value.NewNativeFunction(
		"webui-show-browser",
		[]value.ParamSpec{
			value.NewParamSpec("window-id", true),
			value.NewParamSpec("content", true),
			value.NewParamSpec("browser", true),
		},
		func(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
			return WebUIShowBrowser(args, refValues, eval)
		},
		false,
		&NativeDoc{
			Category:    "WebUI",
			Summary:     "Show window in specific browser",
			Description: `Shows the WebUI window in a specific browser with HTML content.`,
			Parameters: []ParamDoc{
				{Name: "window-id", Type: "integer!", Description: "Window ID", Optional: false},
				{Name: "content", Type: "string!", Description: "HTML content", Optional: false},
				{Name: "browser", Type: "integer!", Description: "Browser ID", Optional: false},
			},
			Returns:  "[logic!] Success status",
			Examples: []string{"webui-show-browser window-id \"<h1>Hello</h1>\" 1"},
			Tags:     []string{"webui", "window", "show", "browser"},
		},
	)))

	rootFrame.Bind("webui-bind", value.NewFuncVal(value.NewNativeFunction(
		"webui-bind",
		[]value.ParamSpec{
			value.NewParamSpec("window-id", true),
			value.NewParamSpec("element", true),
			value.NewParamSpec("handler", true),
		},
		func(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
			return WebUIBind(args, refValues, eval)
		},
		false,
		&NativeDoc{
			Category:    "WebUI",
			Summary:     "Bind event handler to element",
			Description: `Binds a Viro block as an event handler for the specified element.`,
			Parameters: []ParamDoc{
				{Name: "window-id", Type: "integer!", Description: "Window ID", Optional: false},
				{Name: "element", Type: "string!", Description: "Element name/ID", Optional: false},
				{Name: "handler", Type: "block!", Description: "Handler block", Optional: false},
			},
			Returns:  "[logic!] Success status",
			Examples: []string{"webui-bind window-id \"myButton\" [print \"clicked\"]"},
			Tags:     []string{"webui", "event", "bind"},
		},
	)))

	rootFrame.Bind("webui-run", value.NewFuncVal(value.NewNativeFunction(
		"webui-run",
		[]value.ParamSpec{
			value.NewParamSpec("window-id", true),
			value.NewParamSpec("script", true),
		},
		func(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
			return WebUIRun(args, refValues, eval)
		},
		false,
		&NativeDoc{
			Category:    "WebUI",
			Summary:     "Execute JavaScript in window",
			Description: `Executes JavaScript code in the specified window.`,
			Parameters: []ParamDoc{
				{Name: "window-id", Type: "integer!", Description: "Window ID", Optional: false},
				{Name: "script", Type: "string!", Description: "JavaScript code", Optional: false},
			},
			Returns:  "[none!] None",
			Examples: []string{"webui-run window-id \"console.log('Hello')\""},
			Tags:     []string{"webui", "script", "javascript"},
		},
	)))

	rootFrame.Bind("webui-close", value.NewFuncVal(value.NewNativeFunction(
		"webui-close",
		[]value.ParamSpec{
			value.NewParamSpec("window-id", true),
		},
		func(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
			return WebUICloseDirect(args, refValues, eval)
		},
		false,
		&NativeDoc{
			Category:    "WebUI",
			Summary:     "Close a window",
			Description: `Closes the specified WebUI window.`,
			Parameters: []ParamDoc{
				{Name: "window-id", Type: "integer!", Description: "Window ID", Optional: false},
			},
			Returns:  "[none!] None",
			Examples: []string{"webui-close window-id"},
			Tags:     []string{"webui", "window", "close"},
		},
	)))

	rootFrame.Bind("webui-destroy", value.NewFuncVal(value.NewNativeFunction(
		"webui-destroy",
		[]value.ParamSpec{
			value.NewParamSpec("window-id", true),
		},
		func(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
			return WebUIDestroy(args, refValues, eval)
		},
		false,
		&NativeDoc{
			Category:    "WebUI",
			Summary:     "Destroy a window",
			Description: `Destroys the specified WebUI window and frees its resources.`,
			Parameters: []ParamDoc{
				{Name: "window-id", Type: "integer!", Description: "Window ID", Optional: false},
			},
			Returns:  "[none!] None",
			Examples: []string{"webui-destroy window-id"},
			Tags:     []string{"webui", "window", "destroy"},
		},
	)))

	rootFrame.Bind("webui-is-shown", value.NewFuncVal(value.NewNativeFunction(
		"webui-is-shown",
		[]value.ParamSpec{
			value.NewParamSpec("window-id", true),
		},
		func(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
			return WebUIIsShown(args, refValues, eval)
		},
		false,
		&NativeDoc{
			Category:    "WebUI",
			Summary:     "Check if window is shown",
			Description: `Returns whether the specified window is currently shown.`,
			Parameters: []ParamDoc{
				{Name: "window-id", Type: "integer!", Description: "Window ID", Optional: false},
			},
			Returns:  "[logic!] Whether window is shown",
			Examples: []string{"shown?: webui-is-shown window-id"},
			Tags:     []string{"webui", "window", "state"},
		},
	)))

	rootFrame.Bind("webui-set-timeout", value.NewFuncVal(value.NewNativeFunction(
		"webui-set-timeout",
		[]value.ParamSpec{
			value.NewParamSpec("timeout", true),
		},
		func(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
			return WebUISetTimeout(args, refValues, eval)
		},
		false,
		&NativeDoc{
			Category:    "WebUI",
			Summary:     "Set global timeout",
			Description: `Sets the global timeout for WebUI operations in seconds.`,
			Parameters: []ParamDoc{
				{Name: "timeout", Type: "integer!", Description: "Timeout in seconds", Optional: false},
			},
			Returns:  "[none!] None",
			Examples: []string{"webui-set-timeout 30"},
			Tags:     []string{"webui", "config", "timeout"},
		},
	)))

	rootFrame.Bind("webui-wait", value.NewFuncVal(value.NewNativeFunction(
		"webui-wait",
		[]value.ParamSpec{},
		func(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
			return WebUIWait(args, refValues, eval)
		},
		false,
		&NativeDoc{
			Category:    "WebUI",
			Summary:     "Wait for all windows to close",
			Description: `Blocks until all WebUI windows are closed.`,
			Parameters:  []ParamDoc{},
			Returns:     "[none!] None",
			Examples:    []string{"webui-wait"},
			Tags:        []string{"webui", "wait", "block"},
		},
	)))

	rootFrame.Bind("webui-exit", value.NewFuncVal(value.NewNativeFunction(
		"webui-exit",
		[]value.ParamSpec{},
		func(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
			return WebUIExit(args, refValues, eval)
		},
		false,
		&NativeDoc{
			Category:    "WebUI",
			Summary:     "Exit WebUI",
			Description: `Exits the WebUI subsystem.`,
			Parameters:  []ParamDoc{},
			Returns:     "[none!] None",
			Examples:    []string{"webui-exit"},
			Tags:        []string{"webui", "exit"},
		},
	)))

	rootFrame.Bind("webui-set-config", value.NewFuncVal(value.NewNativeFunction(
		"webui-set-config",
		[]value.ParamSpec{
			value.NewParamSpec("option", true),
			value.NewParamSpec("value", true),
		},
		func(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
			return WebUISetConfig(args, refValues, eval)
		},
		false,
		&NativeDoc{
			Category:    "WebUI",
			Summary:     "Set WebUI configuration option",
			Description: `Sets a WebUI configuration option.`,
			Parameters: []ParamDoc{
				{Name: "option", Type: "integer!", Description: "Config option ID", Optional: false},
				{Name: "value", Type: "logic!", Description: "Option value", Optional: false},
			},
			Returns:  "[none!] None",
			Examples: []string{"webui-set-config 0 true"},
			Tags:     []string{"webui", "config"},
		},
	)))

	rootFrame.Bind("webui-get-parent-process-id", value.NewFuncVal(value.NewNativeFunction(
		"webui-get-parent-process-id",
		[]value.ParamSpec{
			value.NewParamSpec("window-id", true),
		},
		func(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
			return WebUIGetParentProcessId(args, refValues, eval)
		},
		false,
		&NativeDoc{
			Category:    "WebUI",
			Summary:     "Get parent process ID",
			Description: `Returns the parent process ID for the specified window.`,
			Parameters: []ParamDoc{
				{Name: "window-id", Type: "integer!", Description: "Window ID", Optional: false},
			},
			Returns:  "[integer!] Parent process ID",
			Examples: []string{"pid: webui-get-parent-process-id window-id"},
			Tags:     []string{"webui", "process", "system"},
		},
	)))

	rootFrame.Bind("webui-set-root-folder", value.NewFuncVal(value.NewNativeFunction(
		"webui-set-root-folder",
		[]value.ParamSpec{
			value.NewParamSpec("window-id", true),
			value.NewParamSpec("path", true),
		},
		func(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
			return WebUISetRootFolder(args, refValues, eval)
		},
		false,
		&NativeDoc{
			Category:    "WebUI",
			Summary:     "Set root folder for window",
			Description: `Sets the root folder for serving files in the specified window.`,
			Parameters: []ParamDoc{
				{Name: "window-id", Type: "integer!", Description: "Window ID", Optional: false},
				{Name: "path", Type: "string!", Description: "Root folder path", Optional: false},
			},
			Returns:  "[logic!] Success status",
			Examples: []string{"webui-set-root-folder window-id \"/var/www\""},
			Tags:     []string{"webui", "filesystem", "serve"},
		},
	)))

	rootFrame.Bind("webui-set-default-root-folder", value.NewFuncVal(value.NewNativeFunction(
		"webui-set-default-root-folder",
		[]value.ParamSpec{
			value.NewParamSpec("path", true),
		},
		func(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
			return WebUISetDefaultRootFolder(args, refValues, eval)
		},
		false,
		&NativeDoc{
			Category:    "WebUI",
			Summary:     "Set default root folder",
			Description: `Sets the default root folder for serving files.`,
			Parameters: []ParamDoc{
				{Name: "path", Type: "string!", Description: "Default root folder path", Optional: false},
			},
			Returns:  "[logic!] Success status",
			Examples: []string{"webui-set-default-root-folder \"/var/www\""},
			Tags:     []string{"webui", "filesystem", "serve"},
		},
	)))

	rootFrame.Bind("webui-open-url", value.NewFuncVal(value.NewNativeFunction(
		"webui-open-url",
		[]value.ParamSpec{
			value.NewParamSpec("url", true),
		},
		func(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
			return WebUIOpenUrl(args, refValues, eval)
		},
		false,
		&NativeDoc{
			Category:    "WebUI",
			Summary:     "Open URL in default browser",
			Description: `Opens the specified URL in the default web browser.`,
			Parameters: []ParamDoc{
				{Name: "url", Type: "string!", Description: "URL to open", Optional: false},
			},
			Returns:  "[none!] None",
			Examples: []string{"webui-open-url \"https://webui.me\""},
			Tags:     []string{"webui", "url", "browser"},
		},
	)))

	rootFrame.Bind("webui-set-hide", value.NewFuncVal(value.NewNativeFunction(
		"webui-set-hide",
		[]value.ParamSpec{
			value.NewParamSpec("window-id", true),
			value.NewParamSpec("hidden", true),
		},
		func(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
			return WebUISetHide(args, refValues, eval)
		},
		false,
		&NativeDoc{
			Category:    "WebUI",
			Summary:     "Hide or show window",
			Description: `Hides or shows the specified window.`,
			Parameters: []ParamDoc{
				{Name: "window-id", Type: "integer!", Description: "Window ID", Optional: false},
				{Name: "hidden", Type: "logic!", Description: "Whether to hide window", Optional: false},
			},
			Returns:  "[none!] None",
			Examples: []string{"webui-set-hide window-id true"},
			Tags:     []string{"webui", "window", "visibility"},
		},
	)))

	rootFrame.Bind("webui-set-size", value.NewFuncVal(value.NewNativeFunction(
		"webui-set-size",
		[]value.ParamSpec{
			value.NewParamSpec("window-id", true),
			value.NewParamSpec("width", true),
			value.NewParamSpec("height", true),
		},
		func(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
			return WebUISetSize(args, refValues, eval)
		},
		false,
		&NativeDoc{
			Category:    "WebUI",
			Summary:     "Set window size",
			Description: `Sets the size of the specified window.`,
			Parameters: []ParamDoc{
				{Name: "window-id", Type: "integer!", Description: "Window ID", Optional: false},
				{Name: "width", Type: "integer!", Description: "Window width", Optional: false},
				{Name: "height", Type: "integer!", Description: "Window height", Optional: false},
			},
			Returns:  "[none!] None",
			Examples: []string{"webui-set-size window-id 800 600"},
			Tags:     []string{"webui", "window", "size"},
		},
	)))

	rootFrame.Bind("webui-set-position", value.NewFuncVal(value.NewNativeFunction(
		"webui-set-position",
		[]value.ParamSpec{
			value.NewParamSpec("window-id", true),
			value.NewParamSpec("x", true),
			value.NewParamSpec("y", true),
		},
		func(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
			return WebUISetPosition(args, refValues, eval)
		},
		false,
		&NativeDoc{
			Category:    "WebUI",
			Summary:     "Set window position",
			Description: `Sets the position of the specified window.`,
			Parameters: []ParamDoc{
				{Name: "window-id", Type: "integer!", Description: "Window ID", Optional: false},
				{Name: "x", Type: "integer!", Description: "X coordinate", Optional: false},
				{Name: "y", Type: "integer!", Description: "Y coordinate", Optional: false},
			},
			Returns:  "[none!] None",
			Examples: []string{"webui-set-position window-id 100 100"},
			Tags:     []string{"webui", "window", "position"},
		},
	)))

	rootFrame.Bind("webui-set-profile", value.NewFuncVal(value.NewNativeFunction(
		"webui-set-profile",
		[]value.ParamSpec{
			value.NewParamSpec("window-id", true),
			value.NewParamSpec("name", true),
			value.NewParamSpec("path", true),
		},
		func(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
			return WebUISetProfile(args, refValues, eval)
		},
		false,
		&NativeDoc{
			Category:    "WebUI",
			Summary:     "Set browser profile",
			Description: `Sets the browser profile for the specified window.`,
			Parameters: []ParamDoc{
				{Name: "window-id", Type: "integer!", Description: "Window ID", Optional: false},
				{Name: "name", Type: "string!", Description: "Profile name", Optional: false},
				{Name: "path", Type: "string!", Description: "Profile path", Optional: false},
			},
			Returns:  "[none!] None",
			Examples: []string{"webui-set-profile window-id \"myprofile\" \"/path/to/profile\""},
			Tags:     []string{"webui", "browser", "profile"},
		},
	)))

	rootFrame.Bind("webui-get-size", value.NewFuncVal(value.NewNativeFunction(
		"webui-get-size",
		[]value.ParamSpec{
			value.NewParamSpec("window-id", true),
		},
		func(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
			return WebUIGetSize(args, refValues, eval)
		},
		false,
		&NativeDoc{
			Category:    "WebUI",
			Summary:     "Get window size",
			Description: `Returns the default size of the specified window (800x600).`,
			Parameters: []ParamDoc{
				{Name: "window-id", Type: "integer!", Description: "Window ID", Optional: false},
			},
			Returns:  "[block!] [width height]",
			Examples: []string{"size: webui-get-size window-id"},
			Tags:     []string{"webui", "window", "size"},
		},
	)))

	rootFrame.Bind("webui-get-position", value.NewFuncVal(value.NewNativeFunction(
		"webui-get-position",
		[]value.ParamSpec{
			value.NewParamSpec("window-id", true),
		},
		func(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
			return WebUIGetPosition(args, refValues, eval)
		},
		false,
		&NativeDoc{
			Category:    "WebUI",
			Summary:     "Get window position",
			Description: `Returns the default position of the specified window (100x100).`,
			Parameters: []ParamDoc{
				{Name: "window-id", Type: "integer!", Description: "Window ID", Optional: false},
			},
			Returns:  "[block!] [x y]",
			Examples: []string{"pos: webui-get-position window-id"},
			Tags:     []string{"webui", "window", "position"},
		},
	)))

	rootFrame.Bind("webui-set-icon", value.NewFuncVal(value.NewNativeFunction(
		"webui-set-icon",
		[]value.ParamSpec{
			value.NewParamSpec("window-id", true),
			value.NewParamSpec("icon", true),
			value.NewParamSpec("icon-type", true),
		},
		func(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
			return WebUISetIcon(args, refValues, eval)
		},
		false,
		&NativeDoc{
			Category:    "WebUI",
			Summary:     "Set window icon",
			Description: `Sets the icon for the specified window.`,
			Parameters: []ParamDoc{
				{Name: "window-id", Type: "integer!", Description: "Window ID", Optional: false},
				{Name: "icon", Type: "string!", Description: "Icon data or path", Optional: false},
				{Name: "icon-type", Type: "string!", Description: "Icon type", Optional: false},
			},
			Returns:  "[none!] None",
			Examples: []string{"webui-set-icon window-id \"icon.png\" \"image/png\""},
			Tags:     []string{"webui", "window", "icon"},
		},
	)))

	rootFrame.Bind("webui-send-raw", value.NewFuncVal(value.NewNativeFunction(
		"webui-send-raw",
		[]value.ParamSpec{
			value.NewParamSpec("window-id", true),
			value.NewParamSpec("function", true),
			value.NewParamSpec("raw-data", true),
		},
		func(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
			return WebUISendRaw(args, refValues, eval)
		},
		false,
		&NativeDoc{
			Category:    "WebUI",
			Summary:     "Send raw data to JavaScript",
			Description: `Sends raw binary data to JavaScript in the specified window.`,
			Parameters: []ParamDoc{
				{Name: "window-id", Type: "integer!", Description: "Window ID", Optional: false},
				{Name: "function", Type: "string!", Description: "JavaScript function name", Optional: false},
				{Name: "raw-data", Type: "binary!", Description: "Raw data to send", Optional: false},
			},
			Returns:  "[none!] None",
			Examples: []string{"webui-send-raw window-id \"myFunc\" #{010203}"},
			Tags:     []string{"webui", "send", "raw", "binary"},
		},
	)))
}
