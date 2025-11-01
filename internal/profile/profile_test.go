package profile

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestNewProfiler(t *testing.T) {
	p := NewProfiler()
	if p == nil {
		t.Fatal("NewProfiler returned nil")
	}
	if p.enabled {
		t.Error("New profiler should be disabled by default")
	}
	if p.functionStats == nil {
		t.Error("functionStats map should be initialized")
	}
}

func TestProfilerEnableDisable(t *testing.T) {
	p := NewProfiler()

	if p.IsEnabled() {
		t.Error("Profiler should start disabled")
	}

	p.Enable()
	if !p.IsEnabled() {
		t.Error("Profiler should be enabled after Enable()")
	}

	p.Disable()
	if p.IsEnabled() {
		t.Error("Profiler should be disabled after Disable()")
	}

	if p.totalTime == 0 {
		t.Error("totalTime should be set after Disable()")
	}
}

func TestRecordEvent(t *testing.T) {
	p := NewProfiler()
	p.Enable()

	p.RecordEvent("test-function", 100*time.Microsecond)
	p.RecordEvent("test-function", 200*time.Microsecond)
	p.RecordEvent("other-function", 50*time.Microsecond)

	p.Disable()

	stats := p.functionStats["test-function"]
	if stats == nil {
		t.Fatal("test-function stats not recorded")
	}

	if stats.CallCount != 2 {
		t.Errorf("Expected 2 calls, got %d", stats.CallCount)
	}

	expectedTotal := 300 * time.Microsecond
	if stats.TotalTime != expectedTotal {
		t.Errorf("Expected total time %v, got %v", expectedTotal, stats.TotalTime)
	}

	if stats.MinTime != 100*time.Microsecond {
		t.Errorf("Expected min time 100µs, got %v", stats.MinTime)
	}

	if stats.MaxTime != 200*time.Microsecond {
		t.Errorf("Expected max time 200µs, got %v", stats.MaxTime)
	}

	expectedAvg := 150 * time.Microsecond
	if stats.AverageTime != expectedAvg {
		t.Errorf("Expected average time %v, got %v", expectedAvg, stats.AverageTime)
	}
}

func TestRecordEventWhenDisabled(t *testing.T) {
	p := NewProfiler()

	p.RecordEvent("test-function", 100*time.Microsecond)

	if len(p.functionStats) != 0 {
		t.Error("Events should not be recorded when profiler is disabled")
	}
}

func TestGetReport(t *testing.T) {
	p := NewProfiler()
	p.Enable()

	p.RecordEvent("fast-func", 10*time.Microsecond)
	p.RecordEvent("slow-func", 1000*time.Microsecond)
	p.RecordEvent("medium-func", 100*time.Microsecond)
	p.RecordEvent("slow-func", 900*time.Microsecond)

	p.Disable()

	report := p.GetReport()

	if report == nil {
		t.Fatal("GetReport returned nil")
	}

	if report.TotalEvents != 4 {
		t.Errorf("Expected 4 total events, got %d", report.TotalEvents)
	}

	if len(report.Functions) != 3 {
		t.Errorf("Expected 3 functions, got %d", len(report.Functions))
	}

	if report.Functions[0].Name != "slow-func" {
		t.Errorf("Expected functions sorted by total time, first should be 'slow-func', got '%s'", report.Functions[0].Name)
	}
}

func TestFormatText(t *testing.T) {
	p := NewProfiler()
	p.Enable()

	p.RecordEvent("test-func", 100*time.Microsecond)
	p.RecordEvent("test-func", 200*time.Microsecond)

	time.Sleep(10 * time.Millisecond)
	p.Disable()

	report := p.GetReport()

	var buf bytes.Buffer
	report.FormatText(&buf)

	output := buf.String()

	if !strings.Contains(output, "EXECUTION PROFILE") {
		t.Error("Output should contain header")
	}

	if !strings.Contains(output, "test-func") {
		t.Error("Output should contain function name")
	}

	if !strings.Contains(output, "Total Execution Time") {
		t.Error("Output should contain total execution time")
	}

	if !strings.Contains(output, "Total Events:") {
		t.Error("Output should contain total events count")
	}
}

func TestFormatTextEmpty(t *testing.T) {
	p := NewProfiler()
	p.Enable()
	p.Disable()

	report := p.GetReport()

	var buf bytes.Buffer
	report.FormatText(&buf)

	output := buf.String()

	if !strings.Contains(output, "No function calls recorded") {
		t.Error("Output should indicate no function calls")
	}
}

func TestFormatJSON(t *testing.T) {
	p := NewProfiler()
	p.Enable()

	p.RecordEvent("test-func", 100*time.Microsecond)

	p.Disable()

	report := p.GetReport()

	var buf bytes.Buffer
	err := report.FormatJSON(&buf)

	if err != nil {
		t.Fatalf("FormatJSON failed: %v", err)
	}

	output := buf.String()

	if !strings.Contains(output, "test-func") {
		t.Error("JSON output should contain function name")
	}

	if !strings.Contains(output, "TotalExecutionTime") {
		t.Error("JSON output should contain TotalExecutionTime field")
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		duration time.Duration
		want     string
	}{
		{500 * time.Nanosecond, "500ns"},
		{1500 * time.Nanosecond, "1.50µs"},
		{1500 * time.Microsecond, "1.50ms"},
		{1500 * time.Millisecond, "1.500s"},
	}

	for _, tt := range tests {
		got := formatDuration(tt.duration)
		if got != tt.want {
			t.Errorf("formatDuration(%v) = %s, want %s", tt.duration, got, tt.want)
		}
	}
}

func TestRecordEventEmptyWord(t *testing.T) {
	p := NewProfiler()
	p.Enable()

	p.RecordEvent("", 100*time.Microsecond)
	p.RecordEvent("valid-func", 100*time.Microsecond)

	p.Disable()

	if len(p.functionStats) != 1 {
		t.Errorf("Expected 1 function stat, got %d (empty words should be ignored)", len(p.functionStats))
	}

	if _, exists := p.functionStats["valid-func"]; !exists {
		t.Error("valid-func should be recorded")
	}
}

func TestConcurrentRecordEvent(t *testing.T) {
	p := NewProfiler()
	p.Enable()

	done := make(chan bool)

	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				p.RecordEvent("concurrent-func", 10*time.Microsecond)
			}
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	p.Disable()

	stats := p.functionStats["concurrent-func"]
	if stats == nil {
		t.Fatal("concurrent-func stats not recorded")
	}

	if stats.CallCount != 1000 {
		t.Errorf("Expected 1000 calls, got %d", stats.CallCount)
	}
}
