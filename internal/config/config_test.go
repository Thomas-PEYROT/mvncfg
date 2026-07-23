package config

import (
	"path/filepath"
	"testing"
)

func TestNew_UsesM2Home(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("M2_HOME", dir)

	cfg, err := New()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Root() != dir {
		t.Errorf("expected root %q, got %q", dir, cfg.Root())
	}
}

func TestNew_FallsBackToHome(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("M2_HOME", "")

	cfg, err := New()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := filepath.Join(home, ".m2")
	if cfg.Root() != want {
		t.Errorf("expected root %q, got %q", want, cfg.Root())
	}
}

func TestNew_ErrorsWithoutHome(t *testing.T) {
	t.Setenv("M2_HOME", "")
	t.Setenv("HOME", "")

	_, err := New()
	if err == nil {
		t.Fatal("expected an error when home directory cannot be determined")
	}
}

func TestNewWithRoot(t *testing.T) {
	dir := t.TempDir()
	cfg := NewWithRoot(dir)
	if cfg.Root() != dir {
		t.Errorf("expected root %q, got %q", dir, cfg.Root())
	}
}

func TestPaths(t *testing.T) {
	root := t.TempDir()
	cfg := NewWithRoot(root)

	if got, want := cfg.ProfilesDir(), filepath.Join(root, "profiles"); got != want {
		t.Errorf("ProfilesDir() = %q, want %q", got, want)
	}
	if got, want := cfg.SettingsPath(), filepath.Join(root, "settings.xml"); got != want {
		t.Errorf("SettingsPath() = %q, want %q", got, want)
	}
	if got, want := cfg.ProfilePath("work"), filepath.Join(root, "profiles", "work.xml"); got != want {
		t.Errorf("ProfilePath(work) = %q, want %q", got, want)
	}
}
