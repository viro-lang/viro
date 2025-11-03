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
			want: "#{DE AD}",
		},
		{
			name: "four bytes DEADBEEF",
			data: []byte{0xDE, 0xAD, 0xBE, 0xEF},
			want: "#{DE AD BE EF}",
		},
		{
			name: "eight bytes",
			data: []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07},
			want: "#{00 01 02 03 04 05 06 07}",
		},
		{
			name: "all zeros",
			data: []byte{0x00, 0x00, 0x00, 0x00},
			want: "#{00 00 00 00}",
		},
		{
			name: "all FF",
			data: []byte{0xFF, 0xFF, 0xFF, 0xFF},
			want: "#{FF FF FF FF}",
		},
		{
			name: "mixed case data",
			data: []byte{0xAB, 0xCD, 0xEF, 0x01, 0x23, 0x45, 0x67, 0x89},
			want: "#{AB CD EF 01 23 45 67 89}",
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
			want: "#{DE AD BE EF}",
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
			want: "#{00 01 02 03 04 05 06 07 08 09 0A 0B 0C 0D 0E 0F 10 11 12 13 14 15 16 17 18 19 1A 1B 1C 1D 1E 1F 20 21 22 23 24 25 26 27 28 29 2A 2B 2C 2D 2E 2F 30 31 32 33 34 35 36 37 38 39 3A 3B 3C 3D 3E 3F}",
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
			want: "#{00 01 02 03 04 05 06 07 ... (65 bytes)}",
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
			want: "#{00 01 02 03 04 05 06 07 ... (128 bytes)}",
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
			want: "#{00 01 02 03 04 05 06 07 ... (256 bytes)}",
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
			want: "#{00 01 02 03 04 05 06 07 ... (1024 bytes)}",
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
	want := "#{DE AD BE EF}"
	if got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
}
