// Package config resolves the Maven configuration directories used by mvncfg.
package config

import (
	"fmt"
	"os"
	"path/filepath"
)

// M2Config holds the paths used by mvncfg.
type M2Config struct {
	root string
}

// New returns a M2Config resolved from the environment.
// It honors the M2_HOME environment variable and falls back to $HOME/.m2.
func New() (*M2Config, error) {
	root := os.Getenv("M2_HOME")
	if root == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("cannot determine home directory: %w", err)
		}
		root = filepath.Join(home, ".m2")
	}

	return &M2Config{root: root}, nil
}

// NewWithRoot returns a M2Config with an explicit root directory.
// It is mainly useful for tests.
func NewWithRoot(root string) *M2Config {
	return &M2Config{root: root}
}

// Root returns the resolved M2_HOME directory.
func (c *M2Config) Root() string {
	return c.root
}

// ProfilesDir returns the directory where profile XML files are stored.
func (c *M2Config) ProfilesDir() string {
	return filepath.Join(c.root, "profiles")
}

// SettingsPath returns the path to the active settings.xml symlink/file.
func (c *M2Config) SettingsPath() string {
	return filepath.Join(c.root, "settings.xml")
}

// ProfilePath returns the path to the XML file for the given profile name.
func (c *M2Config) ProfilePath(name string) string {
	return filepath.Join(c.ProfilesDir(), name+".xml")
}
