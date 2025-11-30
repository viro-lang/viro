//go:build webui

package native

import (
	"errors"
	"sync"
	"testing"

	ui "github.com/webui-dev/go-webui/v2"
)

type mockWindow struct {
	showCalled         bool
	showContent        string
	showBrowserCalled  bool
	showBrowserContent string
	showBrowserBrowser ui.Browser
	showError          error
}

func (m *mockWindow) Show(content string) error {
	m.showCalled = true
	m.showContent = content
	return m.showError
}

func (m *mockWindow) ShowBrowser(content string, browser ui.Browser) error {
	m.showBrowserCalled = true
	m.showBrowserContent = content
	m.showBrowserBrowser = browser
	return m.showError
}

func TestWebUIRuntimeInitialization(t *testing.T) {
	// Reset the once for testing - this is not ideal but necessary for testing
	// In a real scenario, we'd use dependency injection
	webuiInitOnce = sync.Once{}

	mockWin := &mockWindow{}

	// Test ShowWindow
	mockWin.showError = nil
	success, err := ShowWindow(mockWin, "test content")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if !success {
		t.Errorf("Expected success=true, got false")
	}
	if !mockWin.showCalled {
		t.Errorf("Expected Show to be called")
	}
	if mockWin.showContent != "test content" {
		t.Errorf("Expected content 'test content', got '%s'", mockWin.showContent)
	}

	// Reset for next test
	mockWin = &mockWindow{}
	mockWin.showError = nil

	// Test ShowBrowserWindow
	success, err = ShowBrowserWindow(mockWin, "browser content", ui.Chrome)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if !success {
		t.Errorf("Expected success=true, got false")
	}
	if !mockWin.showBrowserCalled {
		t.Errorf("Expected ShowBrowser to be called")
	}
	if mockWin.showBrowserContent != "browser content" {
		t.Errorf("Expected content 'browser content', got '%s'", mockWin.showBrowserContent)
	}
	if mockWin.showBrowserBrowser != ui.Chrome {
		t.Errorf("Expected browser Chrome, got %v", mockWin.showBrowserBrowser)
	}
}

func TestWebUIRuntimeErrorHandling(t *testing.T) {
	// Reset the once for testing
	webuiInitOnce = sync.Once{}

	mockWin := &mockWindow{}
	mockWin.showError = errors.New("mock error")

	// Test ShowWindow with error
	success, err := ShowWindow(mockWin, "test content")
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
	if success {
		t.Errorf("Expected success=false, got true")
	}
}
