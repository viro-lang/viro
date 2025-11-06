package value

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/marcin-radoszewski/viro/internal/core"
)

// Port represents a unified I/O abstraction for files, TCP, and HTTP (Feature 002).
//
// Design per data-model.md and research.md:
// - Scheme: identifies port type (file://, tcp://, http://, https://)
// - Driver: pluggable implementation for scheme-specific operations
// - State: lifecycle tracking (open, closed)
// - Timeout: optional operation timeout (nil = OS defaults per clarification)
//
// Per FR-007: unified abstraction with open/close/read/write/query operations
type Port struct {
	Scheme  string         // "file", "tcp", "http", "https"
	Spec    string         // Original port specification (URL/path)
	Driver  PortDriver     // Scheme-specific implementation
	State   PortState      // Current lifecycle state
	Timeout *time.Duration // Optional timeout (nil = OS default)
}

// PortState tracks port lifecycle.
type PortState int

const (
	PortClosed PortState = iota // Initial/final state
	PortOpen                    // Ready for I/O operations
	PortError                   // Error state (requires close/reopen)
)

func (s PortState) String() string {
	switch s {
	case PortClosed:
		return "closed"
	case PortOpen:
		return "open"
	case PortError:
		return "error"
	default:
		return "unknown"
	}
}

// PortDriver defines the interface that all port implementations must satisfy.
// Allows pluggable file, TCP, and HTTP drivers per research decision.
type PortDriver interface {
	Open(ctx context.Context, spec string) error
	Read(buf []byte) (int, error)
	Write(buf []byte) (int, error)
	Close() error
	Query() (map[string]interface{}, error) // Returns port metadata
}

// NewPort creates a Port with the given scheme and specification.
func NewPort(scheme, spec string, driver PortDriver) *Port {
	return &Port{
		Scheme:  scheme,
		Spec:    spec,
		Driver:  driver,
		State:   PortClosed,
		Timeout: nil, // Default to OS timeout behavior
	}
}

// String returns a debug representation of the port.
func (p *Port) String() string {
	return p.Mold()
}

// Mold returns the mold-formatted port representation.
func (p *Port) Mold() string {
	if p == nil {
		return "port[closed]"
	}
	return fmt.Sprintf("port[%s %s %s]", p.Scheme, p.State, p.Spec)
}

// Form returns the form-formatted port representation (same as mold for ports).
func (p *Port) Form() string {
	return p.Mold()
}

// PortVal creates a Value wrapping a Port.
func PortVal(port *Port) core.Value {
	return port
}

// AsPort extracts the Port from a Value, or returns nil if wrong type.
func AsPort(v core.Value) (*Port, bool) {
	if v.GetType() != TypePort {
		return nil, false
	}
	port, ok := v.GetPayload().(*Port)
	return port, ok
}

// Ensure Port implements io.ReadWriteCloser for standard library compatibility
var _ io.ReadWriteCloser = (*PortAdapter)(nil)

// PortAdapter adapts Port to io.ReadWriteCloser interface.
type PortAdapter struct {
	Port *Port
}

func (a *PortAdapter) Read(p []byte) (n int, err error) {
	if a.Port.Driver == nil {
		return 0, fmt.Errorf("port driver not initialized")
	}
	return a.Port.Driver.Read(p)
}

func (a *PortAdapter) Write(p []byte) (n int, err error) {
	if a.Port.Driver == nil {
		return 0, fmt.Errorf("port driver not initialized")
	}
	return a.Port.Driver.Write(p)
}

func (a *PortAdapter) Close() error {
	if a.Port.Driver == nil {
		return fmt.Errorf("port driver not initialized")
	}
	return a.Port.Driver.Close()
}

func (p *Port) GetType() core.ValueType {
	return TypePort
}

func (p *Port) GetPayload() any {
	return p
}

func (p *Port) Equals(other core.Value) bool {
	if other.GetType() != TypePort {
		return false
	}
	return other.GetPayload() == p
}
