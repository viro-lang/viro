package contract

import (
	"testing"

	"github.com/marcin-radoszewski/viro/internal/value"
)

func TestBinaryMold(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want string
	}{
		{
			name: "empty binary",
			data: []byte{},
			want: "#{}",
		},
		{
			name: "single byte zero",
			data: []byte{0x00},
			want: "#{00}",
		},
		{
			name: "single byte FF",
			data: []byte{0xFF},
			want: "#{FF}",
		},
		{
			name: "single byte 42",
			data: []byte{0x42},
			want: "#{42}",
		},
		{
			name: "two bytes",
			data: []byte{0xDE, 0xAD},
			want: "#{DEAD}",
		},
		{
			name: "four bytes DEADBEEF",
			data: []byte{0xDE, 0xAD, 0xBE, 0xEF},
			want: "#{DEADBEEF}",
		},
		{
			name: "eight bytes",
			data: []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07},
			want: "#{0001020304050607}",
		},
		{
			name: "all zeros",
			data: []byte{0x00, 0x00, 0x00, 0x00},
			want: "#{00000000}",
		},
		{
			name: "all FF",
			data: []byte{0xFF, 0xFF, 0xFF, 0xFF},
			want: "#{FFFFFFFF}",
		},
		{
			name: "mixed case data",
			data: []byte{0xAB, 0xCD, 0xEF, 0x01, 0x23, 0x45, 0x67, 0x89},
			want: "#{ABCDEF0123456789}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bin := value.NewBinaryVal(tt.data)
			binValue := bin.(*value.BinaryValue)
			got := binValue.Mold()
			if got != tt.want {
				t.Errorf("Mold() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestBinaryForm(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want string
	}{
		{
			name: "empty binary",
			data: []byte{},
			want: "#{}",
		},
		{
			name: "single byte",
			data: []byte{0xFF},
			want: "#{FF}",
		},
		{
			name: "four bytes",
			data: []byte{0xDE, 0xAD, 0xBE, 0xEF},
			want: "#{DEADBEEF}",
		},
		{
			name: "exactly 64 bytes",
			data: func() []byte {
				b := make([]byte, 64)
				for i := range b {
					b[i] = byte(i)
				}
				return b
			}(),
			want: "#{000102030405060708090A0B0C0D0E0F101112131415161718191A1B1C1D1E1F202122232425262728292A2B2C2D2E2F303132333435363738393A3B3C3D3E3F}",
		},
		{
			name: "65 bytes - just over threshold",
			data: func() []byte {
				b := make([]byte, 65)
				for i := range b {
					b[i] = byte(i)
				}
				return b
			}(),
			want: "#{0001020304050607...(65 bytes)}",
		},
		{
			name: "128 bytes",
			data: func() []byte {
				b := make([]byte, 128)
				for i := range b {
					b[i] = byte(i % 256)
				}
				return b
			}(),
			want: "#{0001020304050607...(128 bytes)}",
		},
		{
			name: "256 bytes",
			data: func() []byte {
				b := make([]byte, 256)
				for i := range b {
					b[i] = byte(i)
				}
				return b
			}(),
			want: "#{0001020304050607...(256 bytes)}",
		},
		{
			name: "1024 bytes",
			data: func() []byte {
				b := make([]byte, 1024)
				for i := range b {
					b[i] = byte(i % 256)
				}
				return b
			}(),
			want: "#{0001020304050607...(1024 bytes)}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bin := value.NewBinaryVal(tt.data)
			binValue := bin.(*value.BinaryValue)
			got := binValue.Form()
			if got != tt.want {
				t.Errorf("Form() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestBinaryString(t *testing.T) {
	bin := value.NewBinaryVal([]byte{0xDE, 0xAD, 0xBE, 0xEF})
	binValue := bin.(*value.BinaryValue)
	got := binValue.String()
	want := "#{DEADBEEF}"
	if got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
}
