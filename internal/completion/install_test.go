package completion

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAppendMissingLines_MissingFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "rc")

	lines := []string{"# mvncfg completion", "source <(mvncfg completion bash)"}
	added, err := appendMissingLines(path, lines)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(added) != len(lines) {
		t.Errorf("expected %d added lines, got %d", len(lines), len(added))
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("cannot read file: %v", err)
	}
	want := "# mvncfg completion\nsource <(mvncfg completion bash)\n"
	if string(content) != want {
		t.Errorf("unexpected file content:\ngot:  %q\nwant: %q", string(content), want)
	}
}

func TestAppendMissingLines_AllLinesPresent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "rc")
	initial := "# mvncfg completion\nsource <(mvncfg completion bash)\n"
	if err := os.WriteFile(path, []byte(initial), 0o644); err != nil {
		t.Fatalf("cannot write file: %v", err)
	}

	lines := []string{"# mvncfg completion", "source <(mvncfg completion bash)"}
	added, err := appendMissingLines(path, lines)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(added) != 0 {
		t.Errorf("expected no added lines, got %v", added)
	}
}

func TestAppendMissingLines_SubstringIsNotAMatch(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "rc")
	initial := "fpath+=/some/other/dir\n"
	if err := os.WriteFile(path, []byte(initial), 0o644); err != nil {
		t.Fatalf("cannot write file: %v", err)
	}

	lines := []string{"# mvncfg completion", "fpath+=/tmp/mvncfg/completions"}
	added, err := appendMissingLines(path, lines)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(added) != len(lines) {
		t.Errorf("expected %d added lines, got %d", len(lines), len(added))
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("cannot read file: %v", err)
	}
	want := "fpath+=/some/other/dir\n\n# mvncfg completion\nfpath+=/tmp/mvncfg/completions\n"
	if string(content) != want {
		t.Errorf("unexpected file content:\ngot:  %q\nwant: %q", string(content), want)
	}
}

func TestAppendMissingLines_PartialMatch(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "rc")
	initial := "# mvncfg completion\nsource <(mvncfg completion bash)\n"
	if err := os.WriteFile(path, []byte(initial), 0o644); err != nil {
		t.Fatalf("cannot write file: %v", err)
	}

	lines := []string{"# mvncfg completion", "fpath+=/tmp/mvncfg/completions"}
	added, err := appendMissingLines(path, lines)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(added) != 1 {
		t.Errorf("expected 1 added line, got %d: %v", len(added), added)
	}
}

func TestAppendMissingLines_NoTrailingNewline(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "rc")
	initial := "existing line"
	if err := os.WriteFile(path, []byte(initial), 0o644); err != nil {
		t.Fatalf("cannot write file: %v", err)
	}

	lines := []string{"new line"}
	added, err := appendMissingLines(path, lines)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(added) != 1 {
		t.Errorf("expected 1 added line, got %d", len(added))
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("cannot read file: %v", err)
	}
	want := "existing line\n\nnew line\n"
	if string(content) != want {
		t.Errorf("unexpected file content:\ngot:  %q\nwant: %q", string(content), want)
	}
}
