package contract

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/marcin-radoszewski/viro/internal/core"
	"github.com/marcin-radoszewski/viro/internal/eval"
	"github.com/marcin-radoszewski/viro/internal/native"
	"github.com/marcin-radoszewski/viro/internal/value"
)

// Test suite for Feature 002: Port abstraction (file, TCP, HTTP)
// Contract tests validate FR-006 through FR-008 and FR-020 requirements

// T052: open file port with sandbox resolution
func TestFilePortSandbox(t *testing.T) {
	// Setup sandbox root
	tmpDir := t.TempDir()
	if err := eval.InitSandbox(tmpDir); err != nil {
		t.Fatalf("Failed to init sandbox: %v", err)
	}
	// Also set native.SandboxRoot for port operations
	native.SandboxRoot = tmpDir

	// Create test file within sandbox
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("hello"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test 1: Open file within sandbox should succeed
	t.Run("OpenWithinSandbox", func(t *testing.T) {
		port, err := native.OpenPort("test.txt", nil)
		if err != nil {
			t.Errorf("Expected success opening file within sandbox, got error: %v", err)
		}
		if port.GetType() != value.TypePort {
			t.Errorf("Expected TypePort, got %v", port.GetType())
		}
		p, ok := value.AsPort(port)
		if !ok {
			t.Fatal("Failed to extract Port from value")
		}
		if p.State != value.PortOpen {
			t.Errorf("Expected PortOpen state, got %v", p.State)
		}
		// Close the port to clean up
		defer native.ClosePort(port)
	})

	// Test 2: Open file outside sandbox should fail
	t.Run("OpenOutsideSandbox", func(t *testing.T) {
		_, err := native.OpenPort("/etc/passwd", nil)
		if err == nil {
			t.Error("Expected error when opening file outside sandbox")
		}
	})

	// Test 3: Attempt to escape using ../
	t.Run("EscapeAttempt", func(t *testing.T) {
		_, err := native.OpenPort("../../etc/passwd", nil)
		if err == nil {
			t.Error("Expected error when attempting to escape sandbox with ../")
		}
	})
}

// T053: open HTTP port with TLS verification
func TestHTTPPortTLS(t *testing.T) {
	t.Run("HTTPSWithInsecureFlag", func(t *testing.T) {
		// Create test HTTPS server with self-signed cert
		server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
		}))
		defer server.Close()

		// Test with --insecure flag (should succeed)
		opts := map[string]core.Value{
			"insecure": value.LogicVal(true),
		}
		port, err := native.OpenPort(server.URL, opts)
		if err != nil {
			t.Errorf("Expected success with --insecure flag, got: %v", err)
		}
		if port.GetType() == value.TypePort {
			p, _ := value.AsPort(port)
			if p.State != value.PortOpen {
				t.Errorf("Expected PortOpen, got %v", p.State)
			}
			defer native.ClosePort(port)
		}
	})

	t.Run("HTTPSRequestCompletes", func(t *testing.T) {
		// Verify that HTTPS requests actually work end-to-end
		testData := "test response data"
		server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(testData))
		}))
		defer server.Close()

		// Read from HTTPS port
		opts := map[string]core.Value{
			"insecure": value.LogicVal(true),
		}
		content, err := native.ReadPort(server.URL, opts)
		if err != nil {
			t.Errorf("Expected successful HTTPS read, got: %v", err)
		}

		str, ok := value.AsString(content)
		if !ok {
			t.Fatal("Expected string response")
		}
		if !strings.Contains(str.String(), testData) {
			t.Errorf("Expected response to contain '%s', got: %s", testData, str.String())
		}
	})
}

// T054: open TCP port with timeout
func TestTCPPortTimeout(t *testing.T) {
	t.Run("TCPWithTimeout", func(t *testing.T) {
		// Test TCP connection with custom timeout
		opts := map[string]core.Value{
			"timeout": value.IntVal(100), // 100ms timeout
		}
		_, err := native.OpenPort("tcp://localhost:9999", opts)
		// Connection should fail or timeout, but should not panic
		if err == nil {
			t.Log("TCP connection succeeded (unexpected but not an error)")
		}
	})

	t.Run("TCPWithoutTimeout", func(t *testing.T) {
		// Test TCP connection with OS default timeout
		_, err := native.OpenPort("tcp://localhost:9999", nil)
		// Should use OS default timeout
		if err == nil {
			t.Log("TCP connection succeeded (unexpected but not an error)")
		}
	})
}

// T055: read/write file operations
func TestFilePortOperations(t *testing.T) {
	tmpDir := t.TempDir()
	if err := eval.InitSandbox(tmpDir); err != nil {
		t.Fatalf("Failed to init sandbox: %v", err)
	}
	// Also set native.SandboxRoot for port operations
	native.SandboxRoot = tmpDir

	t.Run("WriteAndReadFile", func(t *testing.T) {
		testFile := "test-write.txt"
		// Write data to file
		data := value.StrVal("Hello, Viro!")
		err := native.WritePort(testFile, data, nil)
		if err != nil {
			t.Fatalf("Failed to write file: %v", err)
		}
		defer os.Remove(filepath.Join(tmpDir, testFile))

		// Read data back
		content, err := native.ReadPort(testFile, nil)
		if err != nil {
			t.Fatalf("Failed to read file: %v", err)
		}

		str, ok := value.AsString(content)
		if !ok {
			t.Fatal("Expected string result from read")
		}
		if str.String() != "Hello, Viro!" {
			t.Errorf("Expected 'Hello, Viro!', got '%s'", str.String())
		}
	})

	t.Run("AppendToFile", func(t *testing.T) {
		testFile := "test-append.txt"
		// Write initial data
		data1 := value.StrVal("Line 1\n")
		if err := native.WritePort(testFile, data1, nil); err != nil {
			t.Fatalf("Failed to write initial data: %v", err)
		}
		defer os.Remove(filepath.Join(tmpDir, testFile))

		// Append more data
		data2 := value.StrVal("Line 2\n")
		opts := map[string]core.Value{
			"append": value.LogicVal(true),
		}
		if err := native.WritePort(testFile, data2, opts); err != nil {
			t.Fatalf("Failed to append data: %v", err)
		}

		// Read and verify
		content, err := native.ReadPort(testFile, nil)
		if err != nil {
			t.Fatalf("Failed to read file: %v", err)
		}

		str, _ := value.AsString(content)
		expected := "Line 1\nLine 2\n"
		if str.String() != expected {
			t.Errorf("Expected '%s', got '%s'", expected, str.String())
		}
	})
}

// T056: HTTP GET/POST/HEAD with redirects
func TestHTTPMethods(t *testing.T) {
	t.Run("HTTPGet", func(t *testing.T) {
		// Create test HTTP server
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "GET" {
				t.Errorf("Expected GET method, got %s", r.Method)
			}
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("GET response"))
		}))
		defer server.Close()

		// Test HTTP GET request
		content, err := native.ReadPort(server.URL, nil)
		if err != nil {
			t.Errorf("HTTP GET failed: %v", err)
		}
		if content.GetType() != value.TypeString {
			t.Errorf("Expected string response, got %v", content.GetType())
		}
		str, _ := value.AsString(content)
		if !strings.Contains(str.String(), "GET response") {
			t.Errorf("Expected 'GET response' in content, got: %s", str.String())
		}
	})

	t.Run("HTTPRedirect", func(t *testing.T) {
		// Create test server that handles redirects
		redirectCount := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if redirectCount < 3 {
				redirectCount++
				http.Redirect(w, r, "/redirect", http.StatusFound)
			} else {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("Final destination"))
			}
		}))
		defer server.Close()

		// Test that redirects are followed automatically
		content, err := native.ReadPort(server.URL, nil)
		if err != nil {
			t.Errorf("HTTP redirect failed: %v", err)
		}
		if content.GetType() != value.TypeString {
			t.Errorf("Expected string response after redirects, got %v", content.GetType())
		}
		if redirectCount != 3 {
			t.Errorf("Expected 3 redirects, got %d", redirectCount)
		}
	})

	t.Run("HTTPPost", func(t *testing.T) {
		// Create test server that accepts POST
		var receivedBody string
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "POST" {
				t.Errorf("Expected POST method, got %s", r.Method)
			}
			body, _ := io.ReadAll(r.Body)
			receivedBody = string(body)
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("POST received"))
		}))
		defer server.Close()

		// Test HTTP POST request
		data := value.StrVal("test data")
		opts := map[string]core.Value{
			"method": value.WordVal("POST"),
		}
		err := native.WritePort(server.URL, data, opts)
		if err != nil {
			t.Errorf("HTTP POST failed: %v", err)
		}
		if receivedBody != "test data" {
			t.Errorf("Expected 'test data' in POST body, got: %s", receivedBody)
		}
	})
}

// T057: port query metadata
func TestPortQuery(t *testing.T) {
	tmpDir := t.TempDir()
	if err := eval.InitSandbox(tmpDir); err != nil {
		t.Fatalf("Failed to init sandbox: %v", err)
	}
	// Also set native.SandboxRoot for port operations
	native.SandboxRoot = tmpDir

	// Create test file
	testFile := "query-test.txt"
	testContent := "test content"
	if err := native.WritePort(testFile, value.StrVal(testContent), nil); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	defer os.Remove(filepath.Join(tmpDir, testFile))

	t.Run("QueryFilePort", func(t *testing.T) {
		port, err := native.OpenPort(testFile, nil)
		if err != nil {
			t.Fatalf("Failed to open port: %v", err)
		}
		defer native.ClosePort(port)

		metadata, err := native.QueryPort(port)
		if err != nil {
			t.Errorf("Failed to query port: %v", err)
		}

		if metadata.GetType() != value.TypeObject {
			t.Errorf("Expected object metadata, got %v", metadata.GetType())
		}

		// Metadata should contain: size, modified time, etc.
		obj, ok := value.AsObject(metadata)
		if !ok {
			t.Fatal("Failed to extract object from metadata")
		}
		if obj == nil {
			t.Error("Expected non-nil object metadata")
		}
	})

	t.Run("QueryClosedPort", func(t *testing.T) {
		port, _ := native.OpenPort(testFile, nil)
		native.ClosePort(port)

		_, err := native.QueryPort(port)
		if err == nil {
			t.Error("Expected error when querying closed port")
		}
	})
}

// T058: sandbox escape prevention
func TestSandboxEscapePrevention(t *testing.T) {
	tmpDir := t.TempDir()
	if err := eval.InitSandbox(tmpDir); err != nil {
		t.Fatalf("Failed to init sandbox: %v", err)
	}
	// Also set native.SandboxRoot since port operations use it
	native.SandboxRoot = tmpDir

	escapeAttempts := []string{
		"../../../etc/passwd",
		"/../etc/passwd",
		"/etc/passwd",
		"../outside.txt",
		"subdir/../../outside.txt",
	}

	for _, attempt := range escapeAttempts {
		t.Run(attempt, func(t *testing.T) {
			_, err := native.OpenPort(attempt, nil)
			if err == nil {
				t.Errorf("Expected error for escape attempt: %s", attempt)
			}
		})
	}

	// Test symlink escape prevention
	t.Run("SymlinkEscape", func(t *testing.T) {
		// Create symlink pointing outside sandbox
		outsideFile := filepath.Join(os.TempDir(), "outside.txt")
		os.WriteFile(outsideFile, []byte("outside"), 0644)
		defer os.Remove(outsideFile)

		symlinkPath := filepath.Join(tmpDir, "escape-link")
		if err := os.Symlink(outsideFile, symlinkPath); err != nil {
			t.Skip("Cannot create symlink (may require privileges)")
		}
		defer os.Remove(symlinkPath)

		_, err := native.OpenPort("escape-link", nil)
		if err == nil {
			t.Error("Expected error when following symlink outside sandbox")
		}
	})
}

// T059: TLS --insecure flag behavior
func TestTLSInsecureFlag(t *testing.T) {
	t.Run("InsecureFlagOnHTTPS", func(t *testing.T) {
		// Create test HTTPS server with self-signed cert
		server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		// Test with --insecure flag (should succeed and emit warning)
		opts := map[string]core.Value{
			"insecure": value.LogicVal(true),
		}
		port, err := native.OpenPort(server.URL, opts)
		if err != nil {
			t.Errorf("Expected --insecure to allow self-signed cert, got: %v", err)
		}
		if port.GetType() == value.TypePort {
			native.ClosePort(port)
		}
	})

	t.Run("InsecureFlagOnHTTP", func(t *testing.T) {
		// Create test HTTP server
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		// --insecure should be allowed but has no effect on HTTP
		opts := map[string]core.Value{
			"insecure": value.LogicVal(true),
		}
		port, err := native.OpenPort(server.URL, opts)
		if err != nil {
			t.Errorf("--insecure flag should be allowed on HTTP: %v", err)
		}
		if port.GetType() == value.TypePort {
			native.ClosePort(port)
		}
	})

	t.Run("InsecureFlagOnFile", func(t *testing.T) {
		tmpDir := t.TempDir()
		if err := eval.InitSandbox(tmpDir); err != nil {
			t.Fatalf("Failed to init sandbox: %v", err)
		}
		// Also set native.SandboxRoot for port operations
		native.SandboxRoot = tmpDir

		// --insecure on file:// should raise error
		opts := map[string]core.Value{
			"insecure": value.LogicVal(true),
		}
		_, err := native.OpenPort("test.txt", opts)
		if err == nil {
			t.Error("Expected error when using --insecure with file://")
		}
	})
}
