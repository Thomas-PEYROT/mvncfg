// Package profile implements the mvncfg profile operations.
package profile

import (
	_ "embed"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/Thomas-PEYROT/mvncfg/internal/config"
)

// profileNameRegexp matches valid profile names: letters, digits, '.', '_', and '-'.
var profileNameRegexp = regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)

// validateProfileName returns an error if name is empty, reserved, or contains forbidden characters.
func validateProfileName(name string) error {
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("profile name cannot be empty")
	}
	if name == "." || name == ".." {
		return fmt.Errorf("invalid profile name %q: reserved name", name)
	}
	if !profileNameRegexp.MatchString(name) {
		return fmt.Errorf("invalid profile name %q: only letters, digits, '.', '_', and '-' are allowed", name)
	}
	return nil
}

// List returns the names of all available profiles, sorted alphabetically.
func List(cfg *config.M2Config) ([]string, error) {
	entries, err := os.ReadDir(cfg.ProfilesDir())
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("cannot read profiles directory: %w", err)
	}

	var profiles []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, ".xml") {
			continue
		}
		profiles = append(profiles, strings.TrimSuffix(name, ".xml"))
	}

	sort.Strings(profiles)
	return profiles, nil
}

// Current returns the name of the profile currently active via the settings.xml symlink.
// If there is no symlink or it cannot be resolved, it returns an empty string and an error.
func Current(cfg *config.M2Config) (string, error) {
	settingsPath := cfg.SettingsPath()
	info, err := os.Lstat(settingsPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", fmt.Errorf("no active settings.xml; run mvncfg init: %w", err)
		}
		return "", fmt.Errorf("cannot inspect settings.xml: %w", err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		if info.Mode().IsRegular() {
			return "", fmt.Errorf("settings.xml is a regular file, not a symlink; run mvncfg init to migrate")
		}
		return "", fmt.Errorf("settings.xml is not a symlink; run mvncfg init to migrate")
	}

	target, err := os.Readlink(settingsPath)
	if err != nil {
		return "", fmt.Errorf("cannot read settings.xml symlink: %w", err)
	}

	base := filepath.Base(target)
	if !strings.HasSuffix(base, ".xml") {
		return "", fmt.Errorf("settings.xml does not point to a profile file")
	}

	return strings.TrimSuffix(base, ".xml"), nil
}

// Use activates the given profile by symlinking settings.xml to it.
func Use(cfg *config.M2Config, name string) error {
	if err := validateProfileName(name); err != nil {
		return err
	}

	profilePath := cfg.ProfilePath(name)
	if _, err := os.Stat(profilePath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("unknown profile: %s", name)
		}
		return fmt.Errorf("cannot access profile %s: %w", name, err)
	}

	settingsPath := cfg.SettingsPath()
	info, err := os.Lstat(settingsPath)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("cannot inspect current settings.xml: %w", err)
	}
	if err == nil {
		if info.Mode()&os.ModeSymlink == 0 {
			if info.Mode().IsRegular() {
				return fmt.Errorf("settings.xml is a regular file, not a symlink; run mvncfg init to migrate")
			}
			return fmt.Errorf("settings.xml is not a symlink; run mvncfg init to migrate")
		}
	}

	if err := os.Remove(settingsPath); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("cannot replace current settings.xml: %w", err)
	}

	if err := os.Symlink(profilePath, settingsPath); err != nil {
		return fmt.Errorf("cannot activate profile %s: %w", name, err)
	}

	return nil
}

//go:embed default_settings.xml
var defaultSettingsXML string

// Init creates the ~/.m2/profiles directory and a default profile.
// If an existing settings.xml is a regular file, it is backed up as the default profile.
// Finally, settings.xml is symlinked to the default profile.
func Init(cfg *config.M2Config) error {
	profilesDir := cfg.ProfilesDir()
	if err := os.MkdirAll(profilesDir, 0o755); err != nil {
		return fmt.Errorf("cannot create profiles directory: %w", err)
	}

	settingsPath := cfg.SettingsPath()
	defaultProfilePath := cfg.ProfilePath("default")

	info, err := os.Lstat(settingsPath)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("cannot inspect settings.xml: %w", err)
	}

	if err == nil {
		switch {
		case info.Mode()&os.ModeSymlink != 0:
			// settings.xml is already a symlink. Keep the existing target profile intact
			// and just ensure the symlink stays in place.
		case info.Mode().IsRegular():
			// Backup the existing settings.xml as the default profile.
			if err := os.Rename(settingsPath, defaultProfilePath); err != nil {
				return fmt.Errorf("cannot backup existing settings.xml: %w", err)
			}
		default:
			return fmt.Errorf("settings.xml exists but is not a regular file or symlink")
		}
	}

	// Create a default profile if none exists yet.
	if _, err := os.Stat(defaultProfilePath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			if err := os.WriteFile(defaultProfilePath, []byte(defaultSettingsXML), 0o644); err != nil {
				return fmt.Errorf("cannot create default profile: %w", err)
			}
		} else {
			return fmt.Errorf("cannot access default profile: %w", err)
		}
	}

	// If settings.xml is already a symlink, keep the existing target profile intact
	// and just ensure the default profile exists for future use.
	if err == nil && info.Mode()&os.ModeSymlink != 0 {
		return nil
	}

	// Ensure settings.xml points to the default profile.
	if err := os.Remove(settingsPath); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("cannot replace settings.xml: %w", err)
	}
	if err := os.Symlink(defaultProfilePath, settingsPath); err != nil {
		return fmt.Errorf("cannot link settings.xml to default profile: %w", err)
	}

	return nil
}

// Create creates a new profile with a default settings.xml template.
// It returns an error if the profile already exists.
func Create(cfg *config.M2Config, name string) error {
	if err := validateProfileName(name); err != nil {
		return err
	}

	profilesDir := cfg.ProfilesDir()
	if err := os.MkdirAll(profilesDir, 0o755); err != nil {
		return fmt.Errorf("cannot create profiles directory: %w", err)
	}

	profilePath := cfg.ProfilePath(name)
	if _, err := os.Stat(profilePath); err == nil {
		return fmt.Errorf("profile already exists: %s", name)
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("cannot access profile %s: %w", name, err)
	}

	if err := os.WriteFile(profilePath, []byte(defaultSettingsXML), 0o644); err != nil {
		return fmt.Errorf("cannot create profile %s: %w", name, err)
	}

	return nil
}

// Delete removes a profile file.
// It refuses to delete the currently active profile to avoid breaking the settings.xml symlink.
func Delete(cfg *config.M2Config, name string) error {
	if err := validateProfileName(name); err != nil {
		return err
	}

	current, err := Current(cfg)
	if err == nil && current == name {
		return fmt.Errorf("cannot delete the active profile: %s (switch to another profile first)", name)
	}

	profilePath := cfg.ProfilePath(name)
	if _, err := os.Stat(profilePath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("unknown profile: %s", name)
		}
		return fmt.Errorf("cannot access profile %s: %w", name, err)
	}

	if err := os.Remove(profilePath); err != nil {
		return fmt.Errorf("cannot delete profile %s: %w", name, err)
	}

	return nil
}

// Rename renames a profile. If the renamed profile is the active one,
// the settings.xml symlink is updated to point to the new name.
func Rename(cfg *config.M2Config, oldName, newName string) error {
	if err := validateProfileName(oldName); err != nil {
		return err
	}
	if err := validateProfileName(newName); err != nil {
		return err
	}
	if oldName == newName {
		return fmt.Errorf("old and new profile names are the same")
	}

	oldPath := cfg.ProfilePath(oldName)
	if _, err := os.Stat(oldPath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("unknown profile: %s", oldName)
		}
		return fmt.Errorf("cannot access profile %s: %w", oldName, err)
	}

	newPath := cfg.ProfilePath(newName)
	if _, err := os.Stat(newPath); err == nil {
		return fmt.Errorf("profile already exists: %s", newName)
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("cannot access profile %s: %w", newName, err)
	}

	if err := os.Rename(oldPath, newPath); err != nil {
		return fmt.Errorf("cannot rename profile %s to %s: %w", oldName, newName, err)
	}

	current, err := Current(cfg)
	if err == nil && current == oldName {
		settingsPath := cfg.SettingsPath()
		if err := os.Remove(settingsPath); err != nil {
			return fmt.Errorf("cannot update settings.xml symlink: %w", err)
		}
		if err := os.Symlink(newPath, settingsPath); err != nil {
			return fmt.Errorf("cannot relink settings.xml to %s: %w", newName, err)
		}
	}

	return nil
}
