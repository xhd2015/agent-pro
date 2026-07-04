package main

import (
	"os"
	"path/filepath"
)

const (
	envTTYWatchHome = "TTY_WATCH_HOME"
	defaultHomeDir  = ".tty-watch"
	registrySubdir  = "registry"
)

// TTYWatchHome returns the tty-watch data directory.
func TTYWatchHome() (string, error) {
	if v := os.Getenv(envTTYWatchHome); v != "" {
		return v, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, defaultHomeDir), nil
}

func registryDir(home string) string {
	return filepath.Join(home, registrySubdir)
}

func registryPath(home, sessionID string) string {
	return filepath.Join(registryDir(home), sessionID+".json")
}