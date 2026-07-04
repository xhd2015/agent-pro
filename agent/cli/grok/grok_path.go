package grok

import (
	"os"
	"path/filepath"
	"runtime"
)

// grokInstallCandidates returns well-known grok binary locations for GUI-launched
// processes that inherit a minimal PATH (e.g. macOS Finder / menu-bar apps).
func grokInstallCandidates(home string) []string {
	if home == "" {
		return []string{
			"/opt/homebrew/bin/grok",
			"/usr/local/bin/grok",
		}
	}
	return []string{
		filepath.Join(home, ".grok", "bin", "grok"),
		filepath.Join(home, "go", "bin", "grok"),
		"/opt/homebrew/bin/grok",
		"/usr/local/bin/grok",
	}
}

func findExecutableGrok(candidates []string) (string, bool) {
	for _, candidate := range candidates {
		if isExecutableFile(candidate) {
			return candidate, true
		}
	}
	return "", false
}

func probeGrokInstallPath() (string, bool) {
	home, err := os.UserHomeDir()
	if err != nil {
		return findExecutableGrok(grokInstallCandidates(""))
	}
	return findExecutableGrok(grokInstallCandidates(home))
}

func isExecutableFile(path string) bool {
	info, err := os.Stat(path)
	if err != nil || info.IsDir() {
		return false
	}
	if runtime.GOOS == "windows" {
		return true
	}
	return info.Mode()&0o111 != 0
}