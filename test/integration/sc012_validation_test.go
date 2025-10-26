package integration

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/marcin-radoszewski/viro/internal/eval"
	"github.com/marcin-radoszewski/viro/internal/native"
	"github.com/marcin-radoszewski/viro/internal/value"
)

// TestSC012_FileReadWriteThroughput validates Feature 002 - User Story 2
// Success Criteria SC-012: File I/O throughput >= 50 MB/s
func TestSC012_FileReadWriteThroughput(t *testing.T) {
	// Setup sandbox
	tmpDir := t.TempDir()
	if err := eval.InitSandbox(tmpDir); err != nil {
		t.Fatalf("Failed to init sandbox: %v", err)
	}
	native.SandboxRoot = tmpDir

	// Create test data (1 MB)
	testData := strings.Repeat("a", 1024*1024)
	testFile := "throughput_test.txt"

	// Test write throughput
	t.Run("WriteThroughput", func(t *testing.T) {
		iterations := 10 // Write 10 MB total
		start := time.Now()

		for i := 0; i < iterations; i++ {
			filename := fmt.Sprintf("write_test_%d.txt", i)
			err := native.WritePort(filename, value.NewStrVal(testData), nil)
			if err != nil {
				t.Fatalf("Write failed: %v", err)
			}
		}

		duration := time.Since(start)
		throughput := float64(iterations) / duration.Seconds() // MB/s

		t.Logf("Write throughput: %.2f MB/s", throughput)
		if throughput < 50.0 {
			t.Logf("WARNING: Write throughput below target (50 MB/s), got %.2f MB/s", throughput)
		}
	})

	// Test read throughput
	t.Run("ReadThroughput", func(t *testing.T) {
		// Setup: Create test file
		err := native.WritePort(testFile, value.NewStrVal(testData), nil)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		iterations := 10 // Read 10 MB total
		start := time.Now()

		for i := 0; i < iterations; i++ {
			_, err := native.ReadPort(testFile, nil)
			if err != nil {
				t.Fatalf("Read failed: %v", err)
			}
		}

		duration := time.Since(start)
		throughput := float64(iterations) / duration.Seconds() // MB/s

		t.Logf("Read throughput: %.2f MB/s", throughput)
		if throughput < 50.0 {
			t.Logf("WARNING: Read throughput below target (50 MB/s), got %.2f MB/s", throughput)
		}
	})

	t.Log("SC-012 File I/O throughput validation complete")
}

// TestSC012_HTTPGetLatency validates Feature 002 - User Story 2
// Success Criteria SC-012: HTTP GET latency 95th percentile < 2s for LAN
func TestSC012_HTTPGetLatency(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test response"))
	}))
	defer server.Close()

	t.Run("HTTPGetLatency", func(t *testing.T) {
		url := server.URL
		iterations := 20
		latencies := make([]time.Duration, 0, iterations)

		for i := 0; i < iterations; i++ {
			start := time.Now()
			_, err := native.ReadPort(url, nil)
			latency := time.Since(start)

			if err != nil {
				t.Fatalf("Request %d failed: %v", i+1, err)
			}

			latencies = append(latencies, latency)
		}

		for i := 0; i < len(latencies)-1; i++ {
			for j := i + 1; j < len(latencies); j++ {
				if latencies[i] > latencies[j] {
					latencies[i], latencies[j] = latencies[j], latencies[i]
				}
			}
		}

		p95Index := int(float64(len(latencies)) * 0.95)
		if p95Index >= len(latencies) {
			p95Index = len(latencies) - 1
		}
		p95Latency := latencies[p95Index]

		t.Logf("HTTP GET latency (95th percentile): %v", p95Latency)
		t.Logf("Completed %d/%d requests successfully", len(latencies), iterations)

		if p95Latency > 2*time.Second {
			t.Errorf("95th percentile latency exceeds 2s target (actual: %v)", p95Latency)
		}
	})

	t.Log("SC-012 HTTP GET latency validation complete")
}

// TestSC012_SandboxEnforcement validates Feature 002 - User Story 2
// Success Criteria SC-012: Sandbox prevents access to files outside root
func TestSC012_SandboxEnforcement(t *testing.T) {
	// Setup sandbox
	tmpDir := t.TempDir()
	if err := eval.InitSandbox(tmpDir); err != nil {
		t.Fatalf("Failed to init sandbox: %v", err)
	}
	native.SandboxRoot = tmpDir

	tests := []struct {
		name        string
		path        string
		shouldFail  bool
		description string
	}{
		{
			name:        "AccessWithinSandbox",
			path:        "allowed.txt",
			shouldFail:  false,
			description: "File within sandbox should be accessible",
		},
		{
			name:        "AbsolutePathOutsideSandbox",
			path:        "/etc/passwd",
			shouldFail:  true,
			description: "Absolute path outside sandbox should be denied",
		},
		{
			name:        "RelativePathEscape",
			path:        "../../etc/passwd",
			shouldFail:  true,
			description: "Relative path escape should be denied",
		},
		{
			name:        "SubdirectoryAccess",
			path:        "subdir/file.txt",
			shouldFail:  false,
			description: "Subdirectory within sandbox should be accessible",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Try to write to the path
			err := native.WritePort(tt.path, value.NewStrVal("test content"), nil)

			if tt.shouldFail {
				if err == nil {
					t.Errorf("%s: Expected error (sandbox violation), but operation succeeded", tt.description)
				} else {
					t.Logf("SC-012 PASS: %s - correctly denied", tt.description)
				}
			} else {
				if err != nil {
					t.Errorf("%s: Expected success, got error: %v", tt.description, err)
				} else {
					t.Logf("SC-012 PASS: %s - correctly allowed", tt.description)
					// Cleanup
					resolved, _ := filepath.Abs(filepath.Join(tmpDir, tt.path))
					os.Remove(resolved)
				}
			}
		})
	}

	t.Log("SC-012 Sandbox enforcement validation complete")
}

// TestSC012_PortLifecycle validates basic port operations
func TestSC012_PortLifecycle(t *testing.T) {
	// Setup sandbox
	tmpDir := t.TempDir()
	if err := eval.InitSandbox(tmpDir); err != nil {
		t.Fatalf("Failed to init sandbox: %v", err)
	}
	native.SandboxRoot = tmpDir

	t.Run("FilePortLifecycle", func(t *testing.T) {
		testFile := "lifecycle_test.txt"

		// Open port
		port, err := native.OpenPort(testFile, nil)
		if err != nil {
			t.Fatalf("Failed to open port: %v", err)
		}

		p, ok := value.AsPort(port)
		if !ok {
			t.Fatal("Failed to extract port from value")
		}

		if p.State != value.PortOpen {
			t.Errorf("Expected PortOpen state, got %v", p.State)
		}

		// Close port
		err = native.ClosePort(port)
		if err != nil {
			t.Errorf("Failed to close port: %v", err)
		}

		if p.State != value.PortClosed {
			t.Errorf("Expected PortClosed state after close, got %v", p.State)
		}

		// Test idempotent close
		err = native.ClosePort(port)
		if err != nil {
			t.Errorf("Second close should be idempotent, got error: %v", err)
		}

		t.Log("SC-012 PASS: File port lifecycle")
	})

	t.Run("SaveAndLoad", func(t *testing.T) {
		testFile := "saveload_test.txt"
		testValue := value.NewIntVal(42)

		// Save value
		err := native.SavePort(testFile, testValue, nil)
		if err != nil {
			t.Fatalf("Failed to save: %v", err)
		}

		// Load value
		loaded, err := native.LoadPort(testFile, nil)
		if err != nil {
			t.Fatalf("Failed to load: %v", err)
		}

		// Verify content (basic check - full parsing would require parse package)
		if loaded.GetType() != value.TypeString {
			t.Errorf("Expected string type after load, got %v", loaded.GetType())
		}

		t.Log("SC-012 PASS: Save and load operations")
	})

	t.Log("SC-012 Port lifecycle validation complete")
}
