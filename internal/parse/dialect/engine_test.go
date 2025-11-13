package dialect

import (
	"testing"

	"github.com/marcin-radoszewski/viro/internal/core"
	"github.com/marcin-radoszewski/viro/internal/value"
)

func TestEngine_StringLiteral(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		rules   []core.Value
		want    bool
		wantPos int
	}{
		{
			name:    "simple match",
			input:   "hello",
			rules:   []core.Value{value.NewStrVal("hello")},
			want:    true,
			wantPos: 5,
		},
		{
			name:    "partial match",
			input:   "hello world",
			rules:   []core.Value{value.NewStrVal("hello")},
			want:    true,
			wantPos: 5,
		},
		{
			name:    "no match",
			input:   "hello",
			rules:   []core.Value{value.NewStrVal("world")},
			want:    false,
			wantPos: 0,
		},
		{
			name:    "case insensitive match",
			input:   "Hello",
			rules:   []core.Value{value.NewStrVal("hello")},
			want:    true,
			wantPos: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := value.NewStrVal(tt.input)
			opts := DefaultOptions()
			opts.MatchAll = false // Don't require consuming entire input

			engine, err := NewEngine(input, tt.rules, opts, nil)
			if err != nil {
				t.Fatalf("NewEngine() error = %v", err)
			}

			got, err := engine.Parse(tt.rules)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			if got != tt.want {
				t.Errorf("Parse() = %v, want %v", got, tt.want)
			}

			if engine.GetPosition() != tt.wantPos {
				t.Errorf("GetPosition() = %d, want %d", engine.GetPosition(), tt.wantPos)
			}
		})
	}
}

func TestEngine_Sequence(t *testing.T) {
	input := value.NewStrVal("hello world")
	rules := []core.Value{
		value.NewStrVal("hello"),
		value.NewStrVal(" "),
		value.NewStrVal("world"),
	}
	opts := DefaultOptions()

	engine, err := NewEngine(input, rules, opts, nil)
	if err != nil {
		t.Fatalf("NewEngine() error = %v", err)
	}

	got, err := engine.Parse(rules)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if !got {
		t.Error("Parse() should succeed for sequence match")
	}

	if engine.GetPosition() != 11 {
		t.Errorf("GetPosition() = %d, want 11", engine.GetPosition())
	}
}

func TestEngine_Skip(t *testing.T) {
	input := value.NewStrVal("abc")
	rules := []core.Value{
		value.NewWordVal("skip"), // skip 'a'
		value.NewStrVal("b"),
	}
	opts := DefaultOptions()
	opts.MatchAll = false

	engine, err := NewEngine(input, rules, opts, nil)
	if err != nil {
		t.Fatalf("NewEngine() error = %v", err)
	}

	got, err := engine.Parse(rules)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if !got {
		t.Error("Parse() should succeed")
	}

	if engine.GetPosition() != 2 {
		t.Errorf("GetPosition() = %d, want 2", engine.GetPosition())
	}
}

func TestEngine_End(t *testing.T) {
	input := value.NewStrVal("hello")
	rules := []core.Value{
		value.NewStrVal("hello"),
		value.NewWordVal("end"),
	}
	opts := DefaultOptions()

	engine, err := NewEngine(input, rules, opts, nil)
	if err != nil {
		t.Fatalf("NewEngine() error = %v", err)
	}

	got, err := engine.Parse(rules)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if !got {
		t.Error("Parse() should succeed when 'end' matches at end of input")
	}
}

func TestEngine_Bitset(t *testing.T) {
	input := value.NewStrVal("abc")
	bs := value.NewBitsetFromString("abc")
	rules := []core.Value{
		bs,
		bs,
		bs,
	}
	opts := DefaultOptions()

	engine, err := NewEngine(input, rules, opts, nil)
	if err != nil {
		t.Fatalf("NewEngine() error = %v", err)
	}

	got, err := engine.Parse(rules)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if !got {
		t.Error("Parse() should succeed with bitset matching")
	}
}

func TestEngine_Alternation(t *testing.T) {
	input := value.NewStrVal("hello")
	rules := []core.Value{
		value.NewBlockVal([]core.Value{
			value.NewStrVal("hi"),
			value.NewWordVal("|"),
			value.NewStrVal("hello"),
		}),
	}
	opts := DefaultOptions()

	engine, err := NewEngine(input, rules, opts, nil)
	if err != nil {
		t.Fatalf("NewEngine() error = %v", err)
	}

	got, err := engine.Parse(rules)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if !got {
		t.Error("Parse() should succeed with alternation")
	}
}

func TestEngine_Block(t *testing.T) {
	// Test parsing block input
	input := value.NewBlockVal([]core.Value{
		value.NewIntVal(1),
		value.NewIntVal(2),
		value.NewIntVal(3),
	})
	rules := []core.Value{
		value.NewWordVal("integer!"),
		value.NewWordVal("integer!"),
		value.NewWordVal("integer!"),
	}
	opts := DefaultOptions()

	engine, err := NewEngine(input, rules, opts, nil)
	if err != nil {
		t.Fatalf("NewEngine() error = %v", err)
	}

	got, err := engine.Parse(rules)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if !got {
		t.Error("Parse() should succeed matching block elements")
	}
}
