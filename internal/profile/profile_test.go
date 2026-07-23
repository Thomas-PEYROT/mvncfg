package profile

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Thomas-PEYROT/mvncfg/internal/config"
)

func setupTestConfig(t *testing.T) *config.M2Config {
	t.Helper()
	return config.NewWithRoot(t.TempDir())
}

func createProfileFile(t *testing.T, cfg *config.M2Config, name, content string) {
	t.Helper()
	if err := os.MkdirAll(cfg.ProfilesDir(), 0o755); err != nil {
		t.Fatalf("cannot create profiles dir: %v", err)
	}
	path := cfg.ProfilePath(name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("cannot write profile %s: %v", name, err)
	}
}

func TestList_EmptyWhenDirectoryMissing(t *testing.T) {
	cfg := setupTestConfig(t)

	profiles, err := List(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(profiles) != 0 {
		t.Errorf("expected empty list, got %v", profiles)
	}
}

func TestList_ReturnsSortedProfiles(t *testing.T) {
	cfg := setupTestConfig(t)
	createProfileFile(t, cfg, "work", "<settings/>")
	createProfileFile(t, cfg, "personal", "<settings/>")
	createProfileFile(t, cfg, "backup", "<settings/>")

	profiles, err := List(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []string{"backup", "personal", "work"}
	if len(profiles) != len(want) {
		t.Fatalf("expected %v, got %v", want, profiles)
	}
	for i := range want {
		if profiles[i] != want[i] {
			t.Errorf("expected %v, got %v", want, profiles)
			break
		}
	}
}

func TestCurrent_NoSettings(t *testing.T) {
	cfg := setupTestConfig(t)

	_, err := Current(cfg)
	if err == nil {
		t.Fatal("expected an error when settings.xml does not exist")
	}
	if !errors.Is(err, os.ErrNotExist) {
		t.Errorf("expected os.ErrNotExist in error chain, got %v", err)
	}
}

func TestCurrent_ReturnsActiveProfile(t *testing.T) {
	cfg := setupTestConfig(t)
	createProfileFile(t, cfg, "work", "<settings/>")
	if err := os.Symlink(cfg.ProfilePath("work"), cfg.SettingsPath()); err != nil {
		t.Fatalf("cannot create symlink: %v", err)
	}

	current, err := Current(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if current != "work" {
		t.Errorf("expected profile work, got %q", current)
	}
}

func TestCurrent_RefusesNonXMLTarget(t *testing.T) {
	cfg := setupTestConfig(t)
	if err := os.MkdirAll(cfg.ProfilesDir(), 0o755); err != nil {
		t.Fatalf("cannot create profiles dir: %v", err)
	}
	badFile := filepath.Join(cfg.ProfilesDir(), "notxml.txt")
	if err := os.WriteFile(badFile, []byte("hello"), 0o644); err != nil {
		t.Fatalf("cannot write file: %v", err)
	}
	if err := os.Symlink(badFile, cfg.SettingsPath()); err != nil {
		t.Fatalf("cannot create symlink: %v", err)
	}

	_, err := Current(cfg)
	if err == nil {
		t.Fatal("expected an error when symlink target is not an xml file")
	}
}

func TestUse_UnknownProfile(t *testing.T) {
	cfg := setupTestConfig(t)

	err := Use(cfg, "missing")
	if err == nil {
		t.Fatal("expected an error for unknown profile")
	}
	if !strings.Contains(err.Error(), "unknown profile") {
		t.Errorf("expected 'unknown profile' in error, got %q", err.Error())
	}
}

func TestUse_ActivatesProfile(t *testing.T) {
	cfg := setupTestConfig(t)
	createProfileFile(t, cfg, "work", "<settings/>")

	if err := Use(cfg, "work"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	target, err := os.Readlink(cfg.SettingsPath())
	if err != nil {
		t.Fatalf("settings.xml is not a symlink: %v", err)
	}
	if target != cfg.ProfilePath("work") {
		t.Errorf("settings.xml points to %q, want %q", target, cfg.ProfilePath("work"))
	}
}

func TestUse_ReplacesExistingSymlink(t *testing.T) {
	cfg := setupTestConfig(t)
	createProfileFile(t, cfg, "work", "<settings/>")
	createProfileFile(t, cfg, "personal", "<settings/>")

	if err := Use(cfg, "work"); err != nil {
		t.Fatalf("cannot activate work: %v", err)
	}
	if err := Use(cfg, "personal"); err != nil {
		t.Fatalf("cannot activate personal: %v", err)
	}

	target, err := os.Readlink(cfg.SettingsPath())
	if err != nil {
		t.Fatalf("settings.xml is not a symlink: %v", err)
	}
	if target != cfg.ProfilePath("personal") {
		t.Errorf("settings.xml points to %q, want %q", target, cfg.ProfilePath("personal"))
	}
}

func TestInit_CreatesDefaultProfileAndSymlink(t *testing.T) {
	cfg := setupTestConfig(t)

	if err := Init(cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := os.Stat(cfg.ProfilePath("default")); err != nil {
		t.Errorf("default profile was not created: %v", err)
	}
	target, err := os.Readlink(cfg.SettingsPath())
	if err != nil {
		t.Fatalf("settings.xml is not a symlink: %v", err)
	}
	if target != cfg.ProfilePath("default") {
		t.Errorf("settings.xml points to %q, want %q", target, cfg.ProfilePath("default"))
	}
}

func TestInit_BackupsRegularSettingsFile(t *testing.T) {
	cfg := setupTestConfig(t)
	original := "<settings>original</settings>"
	if err := os.WriteFile(cfg.SettingsPath(), []byte(original), 0o644); err != nil {
		t.Fatalf("cannot write settings.xml: %v", err)
	}

	if err := Init(cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content, err := os.ReadFile(cfg.ProfilePath("default"))
	if err != nil {
		t.Fatalf("cannot read default profile: %v", err)
	}
	if string(content) != original {
		t.Errorf("default profile content = %q, want %q", string(content), original)
	}
}

func TestInit_KeepsExistingSymlink(t *testing.T) {
	cfg := setupTestConfig(t)
	createProfileFile(t, cfg, "custom", "<settings/>")
	if err := os.Symlink(cfg.ProfilePath("custom"), cfg.SettingsPath()); err != nil {
		t.Fatalf("cannot create symlink: %v", err)
	}

	if err := Init(cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	target, err := os.Readlink(cfg.SettingsPath())
	if err != nil {
		t.Fatalf("settings.xml is not a symlink: %v", err)
	}
	if target != cfg.ProfilePath("custom") {
		t.Errorf("settings.xml points to %q, want %q", target, cfg.ProfilePath("custom"))
	}
}

func TestCreate_EmptyName(t *testing.T) {
	cfg := setupTestConfig(t)

	err := Create(cfg, "   ")
	if err == nil {
		t.Fatal("expected an error for empty profile name")
	}
}

func TestCreate_NewProfile(t *testing.T) {
	cfg := setupTestConfig(t)

	if err := Create(cfg, "work"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content, err := os.ReadFile(cfg.ProfilePath("work"))
	if err != nil {
		t.Fatalf("cannot read created profile: %v", err)
	}
	if !strings.Contains(string(content), "<settings") {
		t.Errorf("created profile does not look like a settings.xml: %q", string(content))
	}
}

func TestCreate_DuplicateProfile(t *testing.T) {
	cfg := setupTestConfig(t)
	createProfileFile(t, cfg, "work", "<settings/>")

	err := Create(cfg, "work")
	if err == nil {
		t.Fatal("expected an error for duplicate profile")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("expected 'already exists' in error, got %q", err.Error())
	}
}

func TestDelete_EmptyName(t *testing.T) {
	cfg := setupTestConfig(t)

	err := Delete(cfg, "")
	if err == nil {
		t.Fatal("expected an error for empty profile name")
	}
}

func TestDelete_UnknownProfile(t *testing.T) {
	cfg := setupTestConfig(t)

	err := Delete(cfg, "missing")
	if err == nil {
		t.Fatal("expected an error for unknown profile")
	}
}

func TestDelete_InactiveProfile(t *testing.T) {
	cfg := setupTestConfig(t)
	createProfileFile(t, cfg, "work", "<settings/>")

	if err := Delete(cfg, "work"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := os.Stat(cfg.ProfilePath("work")); !os.IsNotExist(err) {
		t.Error("profile file should have been deleted")
	}
}

func TestDelete_ActiveProfile(t *testing.T) {
	cfg := setupTestConfig(t)
	createProfileFile(t, cfg, "work", "<settings/>")
	if err := Use(cfg, "work"); err != nil {
		t.Fatalf("cannot activate work: %v", err)
	}

	err := Delete(cfg, "work")
	if err == nil {
		t.Fatal("expected an error when deleting active profile")
	}
	if !strings.Contains(err.Error(), "active profile") {
		t.Errorf("expected 'active profile' in error, got %q", err.Error())
	}
}

func TestRename_Validation(t *testing.T) {
	cfg := setupTestConfig(t)

	cases := []struct {
		name    string
		oldName string
		newName string
	}{
		{"empty old", "", "new"},
		{"empty new", "old", ""},
		{"same name", "same", "same"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := Rename(cfg, tc.oldName, tc.newName)
			if err == nil {
				t.Fatal("expected an error")
			}
		})
	}
}

func TestRename_UnknownSource(t *testing.T) {
	cfg := setupTestConfig(t)

	err := Rename(cfg, "missing", "new")
	if err == nil {
		t.Fatal("expected an error for unknown source profile")
	}
}

func TestRename_TargetExists(t *testing.T) {
	cfg := setupTestConfig(t)
	createProfileFile(t, cfg, "old", "<settings/>")
	createProfileFile(t, cfg, "new", "<settings/>")

	err := Rename(cfg, "old", "new")
	if err == nil {
		t.Fatal("expected an error when target profile exists")
	}
}

func TestRename_InactiveProfile(t *testing.T) {
	cfg := setupTestConfig(t)
	createProfileFile(t, cfg, "old", "<settings/>")
	createProfileFile(t, cfg, "other", "<settings/>")
	if err := Use(cfg, "other"); err != nil {
		t.Fatalf("cannot activate other: %v", err)
	}

	if err := Rename(cfg, "old", "renamed"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := os.Stat(cfg.ProfilePath("old")); !os.IsNotExist(err) {
		t.Error("old profile should have been removed")
	}
	if _, err := os.Stat(cfg.ProfilePath("renamed")); err != nil {
		t.Errorf("renamed profile was not created: %v", err)
	}

	target, err := os.Readlink(cfg.SettingsPath())
	if err != nil {
		t.Fatalf("cannot read settings.xml symlink: %v", err)
	}
	if target != cfg.ProfilePath("other") {
		t.Errorf("settings.xml points to %q, want %q", target, cfg.ProfilePath("other"))
	}
}

func TestRename_ActiveProfile(t *testing.T) {
	cfg := setupTestConfig(t)
	createProfileFile(t, cfg, "old", "<settings/>")
	if err := Use(cfg, "old"); err != nil {
		t.Fatalf("cannot activate old: %v", err)
	}

	if err := Rename(cfg, "old", "renamed"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := os.Stat(cfg.ProfilePath("old")); !os.IsNotExist(err) {
		t.Error("old profile should have been removed")
	}

	target, err := os.Readlink(cfg.SettingsPath())
	if err != nil {
		t.Fatalf("cannot read settings.xml symlink: %v", err)
	}
	if target != cfg.ProfilePath("renamed") {
		t.Errorf("settings.xml points to %q, want %q", target, cfg.ProfilePath("renamed"))
	}
}
