package native

import (
	"bufio"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"maps"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/marcin-radoszewski/viro/internal/core"
	"github.com/marcin-radoszewski/viro/internal/trace"
	"github.com/marcin-radoszewski/viro/internal/value"
	"github.com/marcin-radoszewski/viro/internal/verror"
)

// Port driver implementations for Feature 002 (T061-T064)
// Per research.md: pluggable drivers for file, TCP, and HTTP schemes

// SandboxRoot stores the configured sandbox root directory
// Set during initialization from CLI flag --sandbox-root
var SandboxRoot string

// resolveSandboxPath resolves a user path within the sandbox
func resolveSandboxPath(userPath string) (string, error) {
	if SandboxRoot == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("failed to get working directory: %w", err)
		}
		SandboxRoot = cwd
	}

	// Resolve the sandbox root itself through symlinks for proper comparison
	sandboxRootResolved, err := filepath.EvalSymlinks(SandboxRoot)
	if err != nil {
		// If sandbox root doesn't exist, use as-is
		sandboxRootResolved = SandboxRoot
	}

	cleaned := filepath.Clean(userPath)
	var candidate string
	if filepath.IsAbs(cleaned) {
		candidate = cleaned
	} else {
		candidate = filepath.Join(SandboxRoot, cleaned)
	}

	// Evaluate symlinks to detect escape attempts
	resolved, err := filepath.EvalSymlinks(candidate)
	if err != nil {
		// Path doesn't exist yet - check if it's within sandbox
		dir := filepath.Dir(candidate)
		resolvedDir, err := filepath.EvalSymlinks(dir)
		if err != nil {
			// Parent doesn't exist either - verify candidate is within sandbox
			if !strings.HasPrefix(filepath.Clean(candidate)+string(filepath.Separator), filepath.Clean(SandboxRoot)+string(filepath.Separator)) {
				return "", fmt.Errorf("path escapes sandbox: %s", userPath)
			}
			return candidate, nil
		}
		resolved = filepath.Join(resolvedDir, filepath.Base(candidate))
	}

	// Verify resolved path is within sandbox using resolved sandbox root
	// Add separator to prevent false matches like /tmp/sandbox vs /tmp/sandbox-other
	sandboxPrefix := filepath.Clean(sandboxRootResolved) + string(filepath.Separator)
	resolvedClean := filepath.Clean(resolved) + string(filepath.Separator)

	if !strings.HasPrefix(resolvedClean, sandboxPrefix) && resolved != sandboxRootResolved {
		return "", fmt.Errorf("path escapes sandbox: %s resolves to %s", userPath, resolved)
	}

	return resolved, nil
} // fileDriver implements PortDriver for local filesystem operations
type fileDriver struct {
	file *os.File
	path string
}

func (d *fileDriver) Open(ctx context.Context, spec string) error {
	// Resolve path through sandbox (T061)
	resolved, err := resolveSandboxPath(spec)
	if err != nil {
		return fmt.Errorf("sandbox violation: %w", err)
	}
	d.path = resolved

	// Open file for read/write
	f, err := os.OpenFile(resolved, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	d.file = f
	return nil
}

func (d *fileDriver) Read(buf []byte) (int, error) {
	if d.file == nil {
		return 0, fmt.Errorf("file not open")
	}
	return d.file.Read(buf)
}

func (d *fileDriver) Write(buf []byte) (int, error) {
	if d.file == nil {
		return 0, fmt.Errorf("file not open")
	}
	return d.file.Write(buf)
}

func (d *fileDriver) Close() error {
	if d.file == nil {
		return nil // idempotent
	}
	err := d.file.Close()
	d.file = nil
	return err
}

func (d *fileDriver) Query() (map[string]any, error) {
	if d.file == nil {
		return nil, fmt.Errorf("file not open")
	}
	info, err := d.file.Stat()
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"size":     info.Size(),
		"modified": info.ModTime(),
		"mode":     info.Mode().String(),
	}, nil
}

// tcpDriver implements PortDriver for TCP connections
type tcpDriver struct {
	conn    net.Conn
	address string
	timeout *time.Duration
}

func (d *tcpDriver) Open(ctx context.Context, spec string) error {
	// Parse address (format: tcp://host:port)
	address := strings.TrimPrefix(spec, "tcp://")
	d.address = address

	// Create dialer with optional timeout (T062)
	dialer := &net.Dialer{}
	if d.timeout != nil {
		dialer.Timeout = *d.timeout
	}

	// Establish connection
	conn, err := dialer.DialContext(ctx, "tcp", address)
	if err != nil {
		return fmt.Errorf("TCP connection failed: %w", err)
	}
	d.conn = conn
	return nil
}

func (d *tcpDriver) Read(buf []byte) (int, error) {
	if d.conn == nil {
		return 0, fmt.Errorf("connection not open")
	}
	return d.conn.Read(buf)
}

func (d *tcpDriver) Write(buf []byte) (int, error) {
	if d.conn == nil {
		return 0, fmt.Errorf("connection not open")
	}
	return d.conn.Write(buf)
}

func (d *tcpDriver) Close() error {
	if d.conn == nil {
		return nil // idempotent
	}
	err := d.conn.Close()
	d.conn = nil
	return err
}

func (d *tcpDriver) Query() (map[string]any, error) {
	if d.conn == nil {
		return nil, fmt.Errorf("connection not open")
	}
	localAddr := d.conn.LocalAddr().String()
	remoteAddr := d.conn.RemoteAddr().String()
	return map[string]any{
		"local-address":  localAddr,
		"remote-address": remoteAddr,
		"state":          "connected",
	}, nil
}

// httpDriver implements PortDriver for HTTP/HTTPS operations
type httpDriver struct {
	client   *http.Client
	url      string
	response *http.Response
	body     []byte
}

// HTTP client pool for reuse (T063)
var (
	httpClientPool = make(map[clientKey]*http.Client)
	httpClientMu   sync.Mutex
)

type clientKey struct {
	verifyTLS bool
	timeout   time.Duration
}

func getHTTPClient(verifyTLS bool, timeout *time.Duration) *http.Client {
	httpClientMu.Lock()
	defer httpClientMu.Unlock()

	timeoutDur := 30 * time.Second // default
	if timeout != nil {
		timeoutDur = *timeout
	}

	key := clientKey{
		verifyTLS: verifyTLS,
		timeout:   timeoutDur,
	}

	if client, exists := httpClientPool[key]; exists {
		return client
	}

	// Create new client
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: !verifyTLS,
		},
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   timeoutDur,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// Follow max 10 redirects (T064)
			if len(via) >= 10 {
				return fmt.Errorf("stopped after 10 redirects")
			}
			return nil
		},
	}

	httpClientPool[key] = client
	return client
}

func (d *httpDriver) Open(ctx context.Context, spec string) error {
	d.url = spec
	// HTTP driver "open" just stores the URL
	// Actual request happens on Read/Write
	return nil
}

func (d *httpDriver) Read(buf []byte) (int, error) {
	if d.body == nil {
		// Perform GET request
		req, err := http.NewRequest("GET", d.url, nil)
		if err != nil {
			return 0, err
		}

		resp, err := d.client.Do(req)
		if err != nil {
			return 0, err
		}
		defer resp.Body.Close()

		d.response = resp
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return 0, err
		}
		d.body = body
	}

	// Copy body to buffer
	n := copy(buf, d.body)
	d.body = d.body[n:]
	if len(d.body) == 0 {
		return n, io.EOF
	}
	return n, nil
}

func (d *httpDriver) Write(buf []byte) (int, error) {
	// HTTP POST/PUT operation
	req, err := http.NewRequest("POST", d.url, strings.NewReader(string(buf)))
	if err != nil {
		return 0, err
	}

	resp, err := d.client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	d.response = resp
	return len(buf), nil
}

func (d *httpDriver) Close() error {
	// HTTP client is pooled, nothing to close
	d.body = nil
	d.response = nil
	return nil
}

func (d *httpDriver) Query() (map[string]any, error) {
	if d.response == nil {
		return nil, fmt.Errorf("no response available")
	}
	return map[string]any{
		"status":         d.response.StatusCode,
		"content-length": d.response.ContentLength,
		"headers":        d.response.Header,
	}, nil
}

// stdioWriterDriver implements PortDriver for standard output/error streams
type stdioWriterDriver struct {
	writer io.Writer
}

func (d *stdioWriterDriver) Open(ctx context.Context, spec string) error {
	// Always open - stdio streams are always available
	return nil
}

func (d *stdioWriterDriver) Read(buf []byte) (int, error) {
	return 0, fmt.Errorf("stdio writer does not support reading")
}

func (d *stdioWriterDriver) Write(buf []byte) (int, error) {
	return d.writer.Write(buf)
}

func (d *stdioWriterDriver) Close() error {
	// Cannot close stdio streams
	return nil
}

func (d *stdioWriterDriver) Query() (map[string]any, error) {
	return map[string]any{
		"type": "stdio-writer",
	}, nil
}

// stdioReaderDriver implements PortDriver for standard input stream
type stdioReaderDriver struct {
	reader io.Reader
}

func (d *stdioReaderDriver) Open(ctx context.Context, spec string) error {
	// Always open - stdio streams are always available
	return nil
}

func (d *stdioReaderDriver) Read(buf []byte) (int, error) {
	return d.reader.Read(buf)
}

func (d *stdioReaderDriver) Write(buf []byte) (int, error) {
	return 0, fmt.Errorf("stdio reader does not support writing")
}

func (d *stdioReaderDriver) Close() error {
	// Cannot close stdio streams
	return nil
}

func (d *stdioReaderDriver) Query() (map[string]any, error) {
	return map[string]any{
		"type": "stdio-reader",
	}, nil
}

// Port native functions (T065-T072)

// OpenPort implements the `open` native for Feature 002.
// T065: Scheme dispatch and refinement handling
func OpenPort(spec string, opts map[string]core.Value) (core.Value, error) {
	// Parse options
	var timeout *time.Duration
	insecure := false

	if opts != nil {
		if timeoutVal, ok := opts["timeout"]; ok {
			if timeoutVal.GetType() == value.TypeInteger {
				ms, _ := value.AsInteger(timeoutVal)
				dur := time.Duration(ms) * time.Millisecond
				timeout = &dur
			}
		}
		if insecureVal, ok := opts["insecure"]; ok {
			if insecureVal.GetType() == value.TypeLogic {
				insecure, _ = value.AsLogic(insecureVal)
			}
		}
	}

	// Determine scheme
	var driver value.PortDriver
	var scheme string

	if strings.HasPrefix(spec, "http://") || strings.HasPrefix(spec, "https://") {
		scheme = "http"
		if strings.HasPrefix(spec, "https://") {
			scheme = "https"
		}

		// Validate insecure flag usage
		if insecure && scheme == "http" {
			// Allow but ignore
		} else if insecure && scheme == "https" {
			// Emit warning (would go to trace log in full implementation)
			fmt.Fprintf(os.Stderr, "WARNING: TLS verification disabled for %s\n", spec)
		}

		httpDrv := &httpDriver{
			client: getHTTPClient(!insecure, timeout),
			url:    spec,
		}
		driver = httpDrv
	} else if strings.HasPrefix(spec, "tcp://") {
		scheme = "tcp"
		if insecure {
			return value.NoneVal(), fmt.Errorf("--insecure flag not valid for TCP connections")
		}
		tcpDrv := &tcpDriver{timeout: timeout}
		driver = tcpDrv
	} else {
		// File scheme (default)
		scheme = "file"
		if insecure {
			return value.NoneVal(), fmt.Errorf("--insecure flag not valid for file operations")
		}
		driver = &fileDriver{}
	}

	// Create port
	port := value.NewPort(scheme, spec, driver)
	port.Timeout = timeout

	// Open the port
	ctx := context.Background()
	if err := driver.Open(ctx, spec); err != nil {
		trace.TracePortError(scheme, spec, err)
		return value.NoneVal(), err
	}

	port.State = value.PortOpen
	trace.TracePortOpen(scheme, spec)
	return value.PortVal(port), nil
}

// ClosePort implements the `close` native (T066)
func ClosePort(portVal core.Value) error {
	port, ok := value.AsPort(portVal)
	if !ok {
		return fmt.Errorf("expected port value")
	}

	if port.State == value.PortClosed {
		return nil // idempotent
	}

	if err := port.Driver.Close(); err != nil {
		trace.TracePortError(port.Scheme, port.Spec, err)
		return err
	}

	port.State = value.PortClosed
	trace.TracePortClose(port.Scheme, port.Spec)
	return nil
}

// ReadPort implements the `read` native (T067)
func ReadPort(spec string, opts map[string]core.Value) (core.Value, error) {
	// Check if binary mode is requested
	isBinary := false
	if opts != nil {
		if binaryVal, ok := opts["binary"]; ok {
			if binaryVal.GetType() == value.TypeLogic {
				isBinary, _ = value.AsLogic(binaryVal)
			}
		}
	}

	// Open temporary port
	portVal, err := OpenPort(spec, opts)
	if err != nil {
		return value.NoneVal(), err
	}

	port, _ := value.AsPort(portVal)

	// Read all content
	buf := make([]byte, 4096)
	var data []byte
	totalBytes := 0

	for {
		n, err := port.Driver.Read(buf)
		if n > 0 {
			data = append(data, buf[:n]...)
			totalBytes += n
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			trace.TracePortError(port.Scheme, spec, err)
			ClosePort(portVal)
			return value.NoneVal(), err
		}
	}

	trace.TracePortRead(port.Scheme, spec, totalBytes)
	ClosePort(portVal)

	// Return binary or string based on mode
	if isBinary {
		return value.BinaryVal(data), nil
	}
	return value.StrVal(string(data)), nil
}

// WritePort implements the `write` native (T068)
func WritePort(spec string, data core.Value, opts map[string]core.Value) error {
	// Check for append mode
	append := false
	if opts != nil {
		if appendVal, ok := opts["append"]; ok {
			if appendVal.GetType() == value.TypeLogic {
				append, _ = value.AsLogic(appendVal)
			}
		}
	}

	// Get content as bytes (handle both string and binary)
	var contentBytes []byte
	if data.GetType() == value.TypeBinary {
		bin, _ := value.AsBinary(data)
		contentBytes = bin.Bytes()
	} else if data.GetType() == value.TypeString {
		str, _ := value.AsString(data)
		contentBytes = []byte(str.String())
	} else {
		contentBytes = []byte(data.Mold())
	}

	// For file operations with append mode
	if !strings.HasPrefix(spec, "http://") && !strings.HasPrefix(spec, "https://") && !strings.HasPrefix(spec, "tcp://") {
		// File operation
		resolved, err := resolveSandboxPath(spec)
		if err != nil {
			return err
		}

		flags := os.O_WRONLY | os.O_CREATE
		if append {
			flags |= os.O_APPEND
		} else {
			flags |= os.O_TRUNC
		}

		// Ensure directory exists
		dir := filepath.Dir(resolved)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}

		file, err := os.OpenFile(resolved, flags, 0644)
		if err != nil {
			trace.TracePortError("file", spec, err)
			return err
		}
		defer file.Close()

		_, err = file.Write(contentBytes)
		if err != nil {
			trace.TracePortError("file", spec, err)
		} else {
			trace.TracePortWrite("file", spec, len(contentBytes))
		}
		return err
	}

	// Network operation (no append support)
	if append {
		return fmt.Errorf("--append not supported for network operations")
	}

	portVal, err := OpenPort(spec, opts)
	if err != nil {
		trace.TracePortError("network", spec, err)
		return err
	}

	port, _ := value.AsPort(portVal)
	_, err = port.Driver.Write(contentBytes)
	if err != nil {
		trace.TracePortError(port.Scheme, spec, err)
	} else {
		trace.TracePortWrite(port.Scheme, spec, len(contentBytes))
	}
	ClosePort(portVal)
	return err
}

// QueryPort implements the `query` native (T071)
func QueryPort(portVal core.Value) (core.Value, error) {
	port, ok := value.AsPort(portVal)
	if !ok {
		return value.NoneVal(), fmt.Errorf("expected port value")
	}

	if port.State == value.PortClosed {
		return value.NoneVal(), fmt.Errorf("port is closed")
	}

	metadata, err := port.Driver.Query()
	if err != nil {
		return value.NoneVal(), err
	}

	// Convert metadata map to object
	// For now, create a simple object representation
	// Full implementation would use ObjectInstance with proper frame
	obj := value.NewObject(nil, make([]string, 0, len(metadata)), make([]core.ValueType, 0, len(metadata)))

	// Store metadata keys
	for key := range metadata {
		obj.Manifest.Words = append(obj.Manifest.Words, key)
		obj.Manifest.Types = append(obj.Manifest.Types, value.TypeNone)
	}

	return value.ObjectVal(obj), nil
}

// SavePort implements the `save` convenience native (T069)
// Serializes a value using loadable format and writes to file
func SavePort(spec string, val core.Value, opts map[string]core.Value) error {
	// Serialize value to string representation
	serialized := serializeValue(val)

	// Write to file using WritePort
	return WritePort(spec, value.StrVal(serialized), opts)
}

// LoadPort implements the `load` convenience native (T070)
// Reads file and parses into Viro values
func LoadPort(spec string, opts map[string]core.Value) (core.Value, error) {
	// Read file content
	contentVal, err := ReadPort(spec, opts)
	if err != nil {
		return value.NoneVal(), err
	}

	// For now, return the content as-is
	// Full implementation would parse the content into Viro values
	// This requires integration with the parse package
	return contentVal, nil
}

// WaitPort implements the `wait` native (T072)
// Blocks until port is ready or timeout occurs
func WaitPort(portOrBlock core.Value) (core.Value, error) {
	// Handle single port
	if portOrBlock.GetType() == value.TypePort {
		port, ok := value.AsPort(portOrBlock)
		if !ok {
			return value.NoneVal(), fmt.Errorf("expected port value")
		}

		if port.State == value.PortClosed {
			return value.NoneVal(), fmt.Errorf("port is closed")
		}

		// For file ports, they're always ready
		if port.Scheme == "file" {
			return portOrBlock, nil
		}

		// For network ports, check if connection is ready
		// Simple implementation: check if state is open
		if port.State == value.PortOpen {
			return portOrBlock, nil
		}

		return value.NoneVal(), fmt.Errorf("port not ready")
	}

	// Handle block of ports
	if portOrBlock.GetType() == value.TypeBlock {
		blk, ok := value.AsBlock(portOrBlock)
		if !ok {
			return value.NoneVal(), fmt.Errorf("expected block value")
		}

		// Check each port and return first ready one
		for _, elem := range blk.Elements {
			if elem.GetType() == value.TypePort {
				port, ok := value.AsPort(elem)
				if !ok {
					continue
				}

				if port.State == value.PortOpen {
					return elem, nil
				}
			}
		}

		// No ports ready
		return value.NoneVal(), nil
	}

	return value.NoneVal(), fmt.Errorf("expected port or block of ports")
}

// serializeValue converts a value to its loadable string representation
func serializeValue(val core.Value) string {
	switch val.GetType() {
	case value.TypeInteger:
		intVal, _ := value.AsInteger(val)
		return formatInt(intVal)
	case value.TypeDecimal:
		return val.Mold()
	case value.TypeString:
		str, _ := value.AsString(val)
		return fmt.Sprintf(`"%s"`, str.String())
	case value.TypeLogic:
		logicVal, _ := value.AsLogic(val)
		if logicVal {
			return "true"
		}
		return "false"
	case value.TypeNone:
		return "none"
	case value.TypeBlock:
		blk, _ := value.AsBlock(val)
		parts := make([]string, len(blk.Elements))
		for i, elem := range blk.Elements {
			parts[i] = serializeValue(elem)
		}
		return fmt.Sprintf("[%s]", strings.Join(parts, " "))
	default:
		return val.Mold()
	}
}

// Print implements the `print` native.
//
// Contract: print value
// - Accepts any value
// - For blocks: reduce elements (evaluate each) and join with spaces
// - Writes result to stdout followed by newline
// - Returns none
func Print(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 1 {
		return value.NoneVal(), arityError("print", 1, len(args))
	}

	output, err := buildPrintOutput(args[0], eval)
	if err != nil {
		return value.NoneVal(), err
	}

	writer := eval.GetOutputWriter()
	if _, writeErr := fmt.Fprintln(writer, output); writeErr != nil {
		return value.NoneVal(), verror.NewAccessError(
			verror.ErrIDInvalidOperation,
			[3]string{fmt.Sprintf("print output error: %v", writeErr), "", ""},
		)
	}

	return value.NoneVal(), nil
}

// Input implements the `input` native.
//
// Contract: input
// - Reads a line from stdin
// - Returns the line as string value without trailing newline
func Input(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 0 {
		return value.NoneVal(), arityError("input", 0, len(args))
	}

	reader := bufio.NewReader(eval.GetInputReader())
	line, err := reader.ReadString('\n')
	if err != nil {
		if !errors.Is(err, io.EOF) {
			return value.NoneVal(), verror.NewAccessError(
				verror.ErrIDInvalidOperation,
				[3]string{fmt.Sprintf("input read error: %v", err), "", ""},
			)
		}
	}

	line = strings.TrimSuffix(line, "\n")
	line = strings.TrimSuffix(line, "\r")

	return value.StrVal(line), nil
}

func buildPrintOutput(val core.Value, eval core.Evaluator) (string, error) {
	// Use reduce on any value (blocks get reduced, others returned as-is)
	reduced, err := Reduce([]core.Value{val}, nil, eval)
	if err != nil {
		return "", err
	}

	// Use form to format the result
	formed, err := Form([]core.Value{reduced}, nil, eval)
	if err != nil {
		return "", err
	}

	if str, ok := value.AsString(formed); ok {
		return str.String(), nil
	}
	return "", verror.NewInternalError("form did not return string", [3]string{})
}

// Native function wrappers for port operations (T073)
// These adapt the port API to the native function signature

// OpenNative is the native wrapper for open
// Usage: open "file.txt" or open "http://example.com"
// Note: Options/refinements not yet supported in native registry
func OpenNative(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 1 {
		return value.NoneVal(), arityError("open", 1, len(args))
	}

	// Get spec string
	var spec string
	if args[0].GetType() == value.TypeString {
		str, _ := value.AsString(args[0])
		spec = str.String()
	} else {
		spec = args[0].Mold()
	}

	// Call OpenPort with no options (for basic REPL usage)
	result, err := OpenPort(spec, nil)
	if err != nil {
		return value.NoneVal(), verror.NewAccessError(
			verror.ErrIDInvalidOperation,
			[3]string{fmt.Sprintf("open failed: %v", err), spec, ""},
		)
	}

	return result, nil
}

// CloseNative is the native wrapper for close
func CloseNative(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 1 {
		return value.NoneVal(), arityError("close", 1, len(args))
	}

	if args[0].GetType() != value.TypePort {
		return value.NoneVal(), typeError("close", "port!", args[0])
	}

	err := ClosePort(args[0])
	if err != nil {
		return value.NoneVal(), verror.NewAccessError(
			verror.ErrIDPortClosed,
			[3]string{fmt.Sprintf("close failed: %v", err), "", ""},
		)
	}

	return value.NoneVal(), nil
}

// ReadNative is the native wrapper for read
func ReadNative(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 1 {
		return value.NoneVal(), arityError("read", 1, len(args))
	}

	// Get spec string
	var spec string
	if args[0].GetType() == value.TypeString {
		str, _ := value.AsString(args[0])
		spec = str.String()
	} else {
		spec = args[0].Mold()
	}

	// Build options map from refinements
	opts := make(map[string]core.Value)
	if refValues != nil {
		maps.Copy(opts, refValues)
	}

	result, err := ReadPort(spec, opts)
	if err != nil {
		return value.NoneVal(), verror.NewAccessError(
			verror.ErrIDInvalidOperation,
			[3]string{fmt.Sprintf("read failed: %v", err), spec, ""},
		)
	}

	return result, nil
}

// WriteNative is the native wrapper for write
func WriteNative(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 2 {
		return value.NoneVal(), arityError("write", 2, len(args))
	}

	// Get spec string
	var spec string
	if args[0].GetType() == value.TypeString {
		str, _ := value.AsString(args[0])
		spec = str.String()
	} else {
		spec = args[0].Mold()
	}

	err := WritePort(spec, args[1], nil)
	if err != nil {
		return value.NoneVal(), verror.NewAccessError(
			verror.ErrIDInvalidOperation,
			[3]string{fmt.Sprintf("write failed: %v", err), spec, ""},
		)
	}

	return value.NoneVal(), nil
}

// SaveNative is the native wrapper for save
func SaveNative(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 2 {
		return value.NoneVal(), arityError("save", 2, len(args))
	}

	// Get spec string
	var spec string
	if args[0].GetType() == value.TypeString {
		str, _ := value.AsString(args[0])
		spec = str.String()
	} else {
		spec = args[0].Mold()
	}

	err := SavePort(spec, args[1], nil)
	if err != nil {
		return value.NoneVal(), verror.NewAccessError(
			verror.ErrIDInvalidOperation,
			[3]string{fmt.Sprintf("save failed: %v", err), spec, ""},
		)
	}

	return value.NoneVal(), nil
}

// LoadNative is the native wrapper for load
func LoadNative(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 1 {
		return value.NoneVal(), arityError("load", 1, len(args))
	}

	// Get spec string
	var spec string
	if args[0].GetType() == value.TypeString {
		str, _ := value.AsString(args[0])
		spec = str.String()
	} else {
		spec = args[0].Mold()
	}

	result, err := LoadPort(spec, nil)
	if err != nil {
		return value.NoneVal(), verror.NewAccessError(
			verror.ErrIDInvalidOperation,
			[3]string{fmt.Sprintf("load failed: %v", err), spec, ""},
		)
	}

	return result, nil
}

// QueryNative is the native wrapper for query
func QueryNative(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 1 {
		return value.NoneVal(), arityError("query", 1, len(args))
	}

	if args[0].GetType() != value.TypePort {
		return value.NoneVal(), typeError("query", "port!", args[0])
	}

	result, err := QueryPort(args[0])
	if err != nil {
		return value.NoneVal(), verror.NewAccessError(
			verror.ErrIDPortClosed,
			[3]string{fmt.Sprintf("query failed: %v", err), "", ""},
		)
	}

	return result, nil
}

// WaitNative is the native wrapper for wait
func WaitNative(args []core.Value, refValues map[string]core.Value, eval core.Evaluator) (core.Value, error) {
	if len(args) != 1 {
		return value.NoneVal(), arityError("wait", 1, len(args))
	}

	result, err := WaitPort(args[0])
	if err != nil {
		return value.NoneVal(), verror.NewAccessError(
			verror.ErrIDInvalidOperation,
			[3]string{fmt.Sprintf("wait failed: %v", err), "", ""},
		)
	}

	return result, nil
}
