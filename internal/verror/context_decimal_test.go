package verror

import (
	"strings"
	"testing"

	"github.com/ericlagergren/decimal"
	"github.com/marcin-radoszewski/viro/internal/core"
	"github.com/marcin-radoszewski/viro/internal/value"
)

// TestCaptureNear_DecimalMetadata verifies that decimal values include
// scale metadata in error context per T048 requirements.
func TestCaptureNear_DecimalMetadata(t *testing.T) {
	tests := []struct {
		name       string
		values     []core.Value
		index      int
		wantSubstr []string
	}{
		{
			name: "decimal with scale in error position",
			values: []core.Value{
				value.NewIntVal(1),
				value.NewIntVal(2),
				value.DecimalVal(decimal.New(1999, -2), 2),
			},
			index: 2,
			wantSubstr: []string{
				"scale:2",
				">>>",
				"<<<",
			},
		},
		{
			name: "decimal with zero scale",
			values: []core.Value{
				value.DecimalVal(decimal.New(42, 0), 0),
				value.NewIntVal(10),
			},
			index: 0,
			wantSubstr: []string{
				"42",
				"scale:0",
				">>>",
				"<<<",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CaptureNear(tt.values, tt.index)

			for _, want := range tt.wantSubstr {
				if !strings.Contains(got, want) {
					t.Errorf("CaptureNear() output missing expected substring\nwant substring: %q\ngot: %q", want, got)
				}
			}

			if !strings.Contains(got, ">>>") || !strings.Contains(got, "<<<") {
				t.Errorf("CaptureNear() missing error position markers\ngot: %q", got)
			}
		})
	}
}
