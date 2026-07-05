package config

import (
	"os"
	"path/filepath"
)

// SettingsPath returns the path to Claude Code's settings file.
func SettingsPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join("~", ".claude", "settings.json")
	}
	return filepath.Join(home, ".claude", "settings.json")
}

// JSONConfigPath returns the path to Claude Code's JSON config file.
func JSONConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join("~", ".claude.json")
	}
	return filepath.Join(home, ".claude.json")
}

// GlobalSkillsDir returns the path to Claude Code's global skills directory.
func GlobalSkillsDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join("~", ".claude", "skills") + string(os.PathSeparator)
	}
	return filepath.Join(home, ".claude", "skills") + string(os.PathSeparator)
}
