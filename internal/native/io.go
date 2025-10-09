package native

import (
	"bufio"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

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

func (d *fileDriver) Query() (map[string]interface{}, error) {
	if d.file == nil {
		return nil, fmt.Errorf("file not open")
	}
	info, err := d.file.Stat()
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
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

func (d *tcpDriver) Query() (map[string]interface{}, error) {
	if d.conn == nil {
		return nil, fmt.Errorf("connection not open")
	}
	localAddr := d.conn.LocalAddr().String()
	remoteAddr := d.conn.RemoteAddr().String()
	return map[string]interface{}{
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

func (d *httpDriver) Query() (map[string]interface{}, error) {
	if d.response == nil {
		return nil, fmt.Errorf("no response available")
	}
	return map[string]interface{}{
		"status":         d.response.StatusCode,
		"content-length": d.response.ContentLength,
		"headers":        d.response.Header,
	}, nil
}

// Port native functions (T065-T072)

// OpenPort implements the `open` native for Feature 002.
// T065: Scheme dispatch and refinement handling
func OpenPort(spec string, opts map[string]value.Value) (value.Value, error) {
	// Parse options
	var timeout *time.Duration
	insecure := false

	if opts != nil {
		if timeoutVal, ok := opts["timeout"]; ok {
			if timeoutVal.Type == value.TypeInteger {
				ms := timeoutVal.Payload.(int64)
				dur := time.Duration(ms) * time.Millisecond
				timeout = &dur
			}
		}
		if insecureVal, ok := opts["insecure"]; ok {
			if insecureVal.Type == value.TypeLogic {
				insecure = insecureVal.Payload.(bool)
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
		return value.NoneVal(), err
	}

	port.State = value.PortOpen
	return value.PortVal(port), nil
}

// ClosePort implements the `close` native (T066)
func ClosePort(portVal value.Value) error {
	port, ok := portVal.AsPort()
	if !ok {
		return fmt.Errorf("expected port value")
	}

	if port.State == value.PortClosed {
		return nil // idempotent
	}

	if err := port.Driver.Close(); err != nil {
		return err
	}

	port.State = value.PortClosed
	return nil
}

// ReadPort implements the `read` native (T067)
func ReadPort(spec string, opts map[string]value.Value) (value.Value, error) {
	// Open temporary port
	portVal, err := OpenPort(spec, opts)
	if err != nil {
		return value.NoneVal(), err
	}

	port, _ := portVal.AsPort()

	// Read all content
	buf := make([]byte, 4096)
	var result strings.Builder

	for {
		n, err := port.Driver.Read(buf)
		if n > 0 {
			result.Write(buf[:n])
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			ClosePort(portVal)
			return value.NoneVal(), err
		}
	}

	ClosePort(portVal)
	return value.StrVal(result.String()), nil
}

// WritePort implements the `write` native (T068)
func WritePort(spec string, data value.Value, opts map[string]value.Value) error {
	// Check for append mode
	append := false
	if opts != nil {
		if appendVal, ok := opts["append"]; ok {
			if appendVal.Type == value.TypeLogic {
				append = appendVal.Payload.(bool)
			}
		}
	}

	// Get string content
	var content string
	if data.Type == value.TypeString {
		str, _ := data.AsString()
		content = str.String()
	} else {
		content = data.String()
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
			return err
		}
		defer file.Close()

		_, err = file.WriteString(content)
		return err
	}

	// Network operation (no append support)
	if append {
		return fmt.Errorf("--append not supported for network operations")
	}

	portVal, err := OpenPort(spec, opts)
	if err != nil {
		return err
	}

	port, _ := portVal.AsPort()
	_, err = port.Driver.Write([]byte(content))
	ClosePort(portVal)
	return err
}

// QueryPort implements the `query` native (T071)
func QueryPort(portVal value.Value) (value.Value, error) {
	port, ok := portVal.AsPort()
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
	obj := &value.ObjectInstance{
		FrameIndex: -1, // Temporary object without frame
		Parent:     -1,
		Manifest: value.ObjectManifest{
			Words: make([]string, 0, len(metadata)),
			Types: make([]value.ValueType, 0, len(metadata)),
		},
	}

	// Store metadata keys
	for key := range metadata {
		obj.Manifest.Words = append(obj.Manifest.Words, key)
		obj.Manifest.Types = append(obj.Manifest.Types, value.TypeNone)
	}

	return value.ObjectVal(obj), nil
}

// Print implements the `print` native.
//
// Contract: print value
// - Accepts any value
// - For blocks: reduce elements (evaluate each) and join with spaces
// - Writes result to stdout followed by newline
// - Returns none
func Print(args []value.Value, eval Evaluator) (value.Value, *verror.Error) {
	if len(args) != 1 {
		return value.NoneVal(), verror.NewScriptError(
			verror.ErrIDArgCount,
			[3]string{"print", "1", formatInt(len(args))},
		)
	}

	output, err := buildPrintOutput(args[0], eval)
	if err != nil {
		return value.NoneVal(), err
	}

	if _, writeErr := fmt.Fprintln(os.Stdout, output); writeErr != nil {
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
func Input(args []value.Value) (value.Value, *verror.Error) {
	if len(args) != 0 {
		return value.NoneVal(), verror.NewScriptError(
			verror.ErrIDArgCount,
			[3]string{"input", "0", formatInt(len(args))},
		)
	}

	reader := bufio.NewReader(os.Stdin)
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

func buildPrintOutput(val value.Value, eval Evaluator) (string, *verror.Error) {
	if val.Type == value.TypeBlock {
		blk, ok := val.AsBlock()
		if !ok {
			return "", verror.NewInternalError("block value missing payload in print", [3]string{})
		}

		if len(blk.Elements) == 0 {
			return "", nil
		}

		parts := make([]string, 0, len(blk.Elements))
		for idx, elem := range blk.Elements {
			evaluated, err := eval.Do_Next(elem)
			if err != nil {
				if err.Near == "" {
					err.SetNear(verror.CaptureNear(blk.Elements, idx))
				}
				return "", err
			}
			parts = append(parts, valueToPrintString(evaluated))
		}
		return strings.Join(parts, " "), nil
	}

	return valueToPrintString(val), nil
}

func valueToPrintString(val value.Value) string {
	if val.Type == value.TypeString {
		if str, ok := val.AsString(); ok {
			return str.String()
		}
	}
	return val.String()
}
