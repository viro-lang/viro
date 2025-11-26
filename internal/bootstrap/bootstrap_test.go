package bootstrap

import (
	"errors"
	"io/fs"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/marcin-radoszewski/viro/internal/eval"
	"github.com/marcin-radoszewski/viro/internal/value"
	"github.com/marcin-radoszewski/viro/internal/verror"
)

func TestLoadAndExecuteBootstrapScriptsFromFS(t *testing.T) {
	mockFS := fstest.MapFS{
		"init.viro": {
			Data: []byte("test-word: 42"),
		},
		"alpha.viro": {
			Data: []byte("alpha-word: 100"),
		},
		"beta.viro": {
			Data: []byte("beta-word: 200"),
		},
	}

	evaluator := eval.NewEvaluator()

	err := LoadAndExecuteBootstrapScriptsFromFS(evaluator, mockFS)
	if err != nil {
		t.Fatalf("failed to load bootstrap scripts: %v", err)
	}

	testWord, ok := evaluator.GetFrameByIndex(0).Get("test-word")
	if !ok {
		t.Error("test-word from init.viro not found")
	} else if testWord.GetType() != value.TypeInteger {
		t.Errorf("expected test-word to be integer, got %s", value.TypeToString(testWord.GetType()))
	}

	alphaWord, ok := evaluator.GetFrameByIndex(0).Get("alpha-word")
	if !ok {
		t.Error("alpha-word from alpha.viro not found")
	}

	betaWord, ok := evaluator.GetFrameByIndex(0).Get("beta-word")
	if !ok {
		t.Error("beta-word from beta.viro not found")
	}

	_ = alphaWord
	_ = betaWord
}

func TestLoadAndExecuteBootstrapScriptsFromFS_InitFirst(t *testing.T) {
	mockFS := fstest.MapFS{
		"zinit.viro": {
			Data: []byte("z-word: 1"),
		},
		"init.viro": {
			Data: []byte("init-word: 2"),
		},
		"azinit.viro": {
			Data: []byte("az-word: 3"),
		},
	}

	evaluator := eval.NewEvaluator()

	err := LoadAndExecuteBootstrapScriptsFromFS(evaluator, mockFS)
	if err != nil {
		t.Fatalf("failed to load bootstrap scripts: %v", err)
	}

	_, ok := evaluator.GetFrameByIndex(0).Get("z-word")
	if !ok {
		t.Error("z-word not found")
	}

	_, ok = evaluator.GetFrameByIndex(0).Get("init-word")
	if !ok {
		t.Error("init-word not found")
	}

	_, ok = evaluator.GetFrameByIndex(0).Get("az-word")
	if !ok {
		t.Error("az-word not found")
	}
}

func TestLoadAndExecuteBootstrapScriptsFromFS_ParseError(t *testing.T) {
	mockFS := fstest.MapFS{
		"bad.viro": {
			Data: []byte("invalid syntax {{{"),
		},
	}

	evaluator := eval.NewEvaluator()

	err := LoadAndExecuteBootstrapScriptsFromFS(evaluator, mockFS)
	if err == nil {
		t.Error("expected parse error, got nil")
	}

	if !strings.Contains(err.Error(), "bad.viro") {
		t.Errorf("error should contain script path, got: %v", err)
	}
}

func TestLoadAndExecuteBootstrapScriptsFromFS_RuntimeError(t *testing.T) {
	mockFS := fstest.MapFS{
		"runtime-error.viro": {
			Data: []byte("undefined-word + 1"),
		},
	}

	evaluator := eval.NewEvaluator()

	err := LoadAndExecuteBootstrapScriptsFromFS(evaluator, mockFS)
	if err == nil {
		t.Error("expected runtime error, got nil")
	}

	if !strings.Contains(err.Error(), "runtime-error.viro") {
		t.Errorf("error should contain script path, got: %v", err)
	}
}

func TestLoadAndExecuteBootstrapScriptsFromFS_Ordering(t *testing.T) {
	mockFS := fstest.MapFS{
		"zscript.viro": {
			Data: []byte("z-word: 1"),
		},
		"ascript.viro": {
			Data: []byte("a-word: 2"),
		},
		"init.viro": {
			Data: []byte("init-word: 3"),
		},
		"bscript.viro": {
			Data: []byte("b-word: 4"),
		},
	}

	evaluator := eval.NewEvaluator()

	err := LoadAndExecuteBootstrapScriptsFromFS(evaluator, mockFS)
	if err != nil {
		t.Fatalf("failed to load bootstrap scripts: %v", err)
	}

	// Check that all words are defined
	checkWord := func(name string) {
		if _, ok := evaluator.GetFrameByIndex(0).Get(name); !ok {
			t.Errorf("word %s not found", name)
		}
	}

	checkWord("z-word")
	checkWord("a-word")
	checkWord("init-word")
	checkWord("b-word")
}

func TestLoadAndExecuteBootstrapScriptsFromFS_NoInit(t *testing.T) {
	mockFS := fstest.MapFS{
		"zscript.viro": {
			Data: []byte("z-word: 1"),
		},
		"ascript.viro": {
			Data: []byte("a-word: 2"),
		},
		"bscript.viro": {
			Data: []byte("b-word: 3"),
		},
	}

	evaluator := eval.NewEvaluator()

	err := LoadAndExecuteBootstrapScriptsFromFS(evaluator, mockFS)
	if err != nil {
		t.Fatalf("failed to load bootstrap scripts: %v", err)
	}

	checkWord := func(name string) {
		if _, ok := evaluator.GetFrameByIndex(0).Get(name); !ok {
			t.Errorf("word %s not found", name)
		}
	}

	checkWord("z-word")
	checkWord("a-word")
	checkWord("b-word")
}

func TestLoadAndExecuteBootstrapScriptsFromFS_WalkError(t *testing.T) {
	evaluator := eval.NewEvaluator()

	err := LoadAndExecuteBootstrapScriptsFromFS(evaluator, walkErrorFS{})
	if err == nil {
		t.Fatal("expected bootstrap error")
	}

	var verr *verror.Error
	if !errors.As(err, &verr) {
		t.Fatalf("expected structured error, got: %v", err)
	}

	if verr.Category != verror.ErrBootstrap {
		t.Fatalf("expected bootstrap category, got: %v", verr.Category)
	}

	if verr.ID != verror.ErrIDBootstrapFailure {
		t.Fatalf("expected bootstrap failure id, got: %s", verr.ID)
	}
}

func TestLoadAndExecuteBootstrapScriptsFromFS_ReadError(t *testing.T) {
	mockFS := readErrorFS{MapFS: fstest.MapFS{
		"init.viro": {
			Data: []byte("test-word: 1"),
		},
	}}

	evaluator := eval.NewEvaluator()

	err := LoadAndExecuteBootstrapScriptsFromFS(evaluator, mockFS)
	if err == nil {
		t.Fatal("expected bootstrap error")
	}

	var verr *verror.Error
	if !errors.As(err, &verr) {
		t.Fatalf("expected structured error, got: %v", err)
	}

	if verr.Category != verror.ErrBootstrap {
		t.Fatalf("expected bootstrap category, got: %v", verr.Category)
	}

	if verr.ID != verror.ErrIDBootstrapFailure {
		t.Fatalf("expected bootstrap failure id, got: %s", verr.ID)
	}
}

type walkErrorFS struct{}

func (walkErrorFS) Open(string) (fs.File, error) {
	return nil, errors.New("walk failure")
}

type readErrorFS struct {
	fstest.MapFS
}

func (r readErrorFS) ReadFile(name string) ([]byte, error) {
	if name == "init.viro" {
		return nil, errors.New("read failure")
	}
	return r.MapFS.ReadFile(name)
}
