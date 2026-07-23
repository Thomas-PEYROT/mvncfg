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
			return "", fmt.Errorf("no active settings.xml: %w", err)
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

const defaultSettingsXML = `<?xml version="1.0" encoding="UTF-8"?>
<!--
  Maven settings.xml
  Documentation: https://maven.apache.org/settings.html
-->
<settings xmlns="http://maven.apache.org/SETTINGS/1.0.0"
          xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
          xsi:schemaLocation="http://maven.apache.org/SETTINGS/1.0.0 https://maven.apache.org/xsd/settings-1.0.0.xsd">

  <!-- Local repository path -->
  <!-- <localRepository>${user.home}/.m2/repository</localRepository> -->

  <!-- Interactive mode: false disables prompting -->
  <interactiveMode>true</interactiveMode>

  <!-- Offline mode: set to true to prevent Maven from connecting to the network -->
  <offline>false</offline>

  <!-- Plugin groups searched when a plugin prefix is used -->
  <pluginGroups>
    <pluginGroup>org.apache.maven.plugins</pluginGroup>
    <pluginGroup>org.codehaus.mojo</pluginGroup>
  </pluginGroups>

  <!-- Servers: credentials for repositories and mirrors -->
  <servers>
    <!--
    <server>
      <id>my-server</id>
      <username>username</username>
      <password>{COQLCE6DU6GtcS5P=}</password>
    </server>
    -->
  </servers>

  <!-- Mirrors: redirect requests to a different server -->
  <mirrors>
    <!--
    <mirror>
      <id>mirror-central</id>
      <url>https://my-mirror.example.com/maven2</url>
      <mirrorOf>central</mirrorOf>
    </mirror>
    -->
  </mirrors>

  <!-- Proxies: network proxy configuration -->
  <proxies>
    <!--
    <proxy>
      <id>my-proxy</id>
      <active>true</active>
      <protocol>http</protocol>
      <host>proxy.example.com</host>
      <port>8080</port>
      <nonProxyHosts>localhost|*.example.com</nonProxyHosts>
    </proxy>
    -->
  </proxies>

  <!-- Profiles: environment-specific configuration -->
  <profiles>
    <!--
    <profile>
      <id>example</id>
      <activation>
        <activeByDefault>false</activeByDefault>
      </activation>
      <repositories>
        <repository>
          <id>example-repo</id>
          <url>https://example.com/maven2</url>
        </repository>
      </repositories>
    </profile>
    -->
  </profiles>

  <!-- Active profiles applied by default -->
  <activeProfiles>
    <!-- <activeProfile>example</activeProfile> -->
  </activeProfiles>

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
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("profile name cannot be empty")
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
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("profile name cannot be empty")
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
	if strings.TrimSpace(oldName) == "" || strings.TrimSpace(newName) == "" {
		return fmt.Errorf("profile names cannot be empty")
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
