// Package profile implements the mvncfg profile operations.
package profile

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Thomas-PEYROT/mvncfg/internal/config"
)

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
	target, err := os.Readlink(cfg.SettingsPath())
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", fmt.Errorf("no active settings.xml")
		}
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
	profilePath := cfg.ProfilePath(name)
	if _, err := os.Stat(profilePath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("unknown profile: %s", name)
		}
		return fmt.Errorf("cannot access profile %s: %w", name, err)
	}

	settingsPath := cfg.SettingsPath()
	if err := os.Remove(settingsPath); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("cannot replace current settings.xml: %w", err)
	}

	if err := os.Symlink(profilePath, settingsPath); err != nil {
		return fmt.Errorf("cannot activate profile %s: %w", name, err)
	}

	return nil
}

const defaultProfileContent = `<?xml version="1.0" encoding="UTF-8"?>
<settings xmlns="http://maven.apache.org/SETTINGS/1.0.0"
          xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
          xsi:schemaLocation="http://maven.apache.org/SETTINGS/1.0.0 https://maven.apache.org/xsd/settings-1.0.0.xsd">
</settings>
`

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
			if err := os.WriteFile(defaultProfilePath, []byte(defaultProfileContent), 0o644); err != nil {
				return fmt.Errorf("cannot create default profile: %w", err)
			}
		} else {
			return fmt.Errorf("cannot access default profile: %w", err)
		}
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
