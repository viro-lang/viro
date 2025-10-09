package contract

import (
	"os"
	"path/filepath"
	"testing"

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
		if port.Type != value.TypePort {
			t.Errorf("Expected TypePort, got %v", port.Type)
		}
		p, ok := port.AsPort()
		if !ok {
			t.Fatal("Failed to extract Port from value")
		}
		if p.State != value.PortOpen {
			t.Errorf("Expected PortOpen state, got %v", p.State)
		}
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
	t.Run("HTTPSWithValidCert", func(t *testing.T) {
		// Test opening HTTPS with valid certificate
		port, err := native.OpenPort("https://www.google.com", nil)
		if err != nil {
			t.Errorf("Expected success with valid HTTPS cert, got: %v", err)
		}
		if port.Type == value.TypePort {
			p, _ := port.AsPort()
			if p.State != value.PortOpen {
				t.Errorf("Expected PortOpen, got %v", p.State)
			}
		}
	})

	t.Run("HTTPSWithInvalidCertNoFlag", func(t *testing.T) {
		// Test opening HTTPS with invalid cert (should fail without --insecure)
		_, err := native.OpenPort("https://self-signed.badssl.com/", nil)
		if err == nil {
			t.Error("Expected error with self-signed cert when --insecure not set")
		}
	})

	t.Run("HTTPSWithInvalidCertAndInsecureFlag", func(t *testing.T) {
		// Test opening HTTPS with invalid cert and --insecure flag
		opts := map[string]value.Value{
			"insecure": value.LogicVal(true),
		}
		port, err := native.OpenPort("https://self-signed.badssl.com/", opts)
		if err != nil {
			t.Errorf("Expected success with --insecure flag, got: %v", err)
		}
		if port.Type == value.TypePort {
			p, _ := port.AsPort()
			if p.State != value.PortOpen {
				t.Errorf("Expected PortOpen, got %v", p.State)
			}
		}
	})
}

// T054: open TCP port with timeout
func TestTCPPortTimeout(t *testing.T) {
	t.Run("TCPWithTimeout", func(t *testing.T) {
		// Test TCP connection with custom timeout
		opts := map[string]value.Value{
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

	t.Run("WriteAndReadFile", func(t *testing.T) {
		// Write data to file
		data := value.StrVal("Hello, Viro!")
		err := native.WritePort("test-write.txt", data, nil)
		if err != nil {
			t.Fatalf("Failed to write file: %v", err)
		}

		// Read data back
		content, err := native.ReadPort("test-write.txt", nil)
		if err != nil {
			t.Fatalf("Failed to read file: %v", err)
		}

		str, ok := content.AsString()
		if !ok {
			t.Fatal("Expected string result from read")
		}
		if str.String() != "Hello, Viro!" {
			t.Errorf("Expected 'Hello, Viro!', got '%s'", str.String())
		}
	})

	t.Run("AppendToFile", func(t *testing.T) {
		// Write initial data
		data1 := value.StrVal("Line 1\n")
		if err := native.WritePort("test-append.txt", data1, nil); err != nil {
			t.Fatalf("Failed to write initial data: %v", err)
		}

		// Append more data
		data2 := value.StrVal("Line 2\n")
		opts := map[string]value.Value{
			"append": value.LogicVal(true),
		}
		if err := native.WritePort("test-append.txt", data2, opts); err != nil {
			t.Fatalf("Failed to append data: %v", err)
		}

		// Read and verify
		content, err := native.ReadPort("test-append.txt", nil)
		if err != nil {
			t.Fatalf("Failed to read file: %v", err)
		}

		str, _ := content.AsString()
		expected := "Line 1\nLine 2\n"
		if str.String() != expected {
			t.Errorf("Expected '%s', got '%s'", expected, str.String())
		}
	})
}

// T056: HTTP GET/POST/HEAD with redirects
func TestHTTPMethods(t *testing.T) {
	t.Run("HTTPGet", func(t *testing.T) {
		// Test HTTP GET request
		content, err := native.ReadPort("http://httpbin.org/get", nil)
		if err != nil {
			t.Errorf("HTTP GET failed: %v", err)
		}
		if content.Type != value.TypeString {
			t.Errorf("Expected string response, got %v", content.Type)
		}
	})

	t.Run("HTTPRedirect", func(t *testing.T) {
		// Test that redirects are followed automatically (max 10 hops)
		content, err := native.ReadPort("http://httpbin.org/redirect/3", nil)
		if err != nil {
			t.Errorf("HTTP redirect failed: %v", err)
		}
		if content.Type != value.TypeString {
			t.Errorf("Expected string response after redirects, got %v", content.Type)
		}
	})

	t.Run("HTTPPost", func(t *testing.T) {
		// Test HTTP POST request
		data := value.StrVal("test data")
		opts := map[string]value.Value{
			"method": value.WordVal("POST"),
		}
		err := native.WritePort("http://httpbin.org/post", data, opts)
		if err != nil {
			t.Errorf("HTTP POST failed: %v", err)
		}
	})
}

// T057: port query metadata
func TestPortQuery(t *testing.T) {
	tmpDir := t.TempDir()
	if err := eval.InitSandbox(tmpDir); err != nil {
		t.Fatalf("Failed to init sandbox: %v", err)
	}

	// Create test file
	testFile := "query-test.txt"
	testContent := "test content"
	if err := native.WritePort(testFile, value.StrVal(testContent), nil); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	t.Run("QueryFilePort", func(t *testing.T) {
		port, err := native.OpenPort(testFile, nil)
		if err != nil {
			t.Fatalf("Failed to open port: %v", err)
		}

		metadata, err := native.QueryPort(port)
		if err != nil {
			t.Errorf("Failed to query port: %v", err)
		}

		if metadata.Type != value.TypeObject {
			t.Errorf("Expected object metadata, got %v", metadata.Type)
		}

		// Metadata should contain: size, modified time, etc.
		obj, ok := metadata.AsObject()
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

		_, err := native.OpenPort("escape-link", nil)
		if err == nil {
			t.Error("Expected error when following symlink outside sandbox")
		}
	})
}

// T059: TLS --insecure flag behavior
func TestTLSInsecureFlag(t *testing.T) {
	t.Run("InsecureFlagOnHTTPS", func(t *testing.T) {
		opts := map[string]value.Value{
			"insecure": value.LogicVal(true),
		}
		_, err := native.OpenPort("https://self-signed.badssl.com/", opts)
		if err != nil {
			t.Errorf("Expected --insecure to allow invalid cert, got: %v", err)
		}
	})

	t.Run("InsecureFlagOnHTTP", func(t *testing.T) {
		// --insecure should be allowed but has no effect on HTTP
		opts := map[string]value.Value{
			"insecure": value.LogicVal(true),
		}
		_, err := native.OpenPort("http://httpbin.org/get", opts)
		if err != nil {
			t.Errorf("--insecure flag should be allowed on HTTP: %v", err)
		}
	})

	t.Run("InsecureFlagOnFile", func(t *testing.T) {
		tmpDir := t.TempDir()
		if err := eval.InitSandbox(tmpDir); err != nil {
			t.Fatalf("Failed to init sandbox: %v", err)
		}

		// --insecure on file:// should raise error
		opts := map[string]value.Value{
			"insecure": value.LogicVal(true),
		}
		_, err := native.OpenPort("test.txt", opts)
		if err == nil {
			t.Error("Expected error when using --insecure with file://")
		}
	})
}
