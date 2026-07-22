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
