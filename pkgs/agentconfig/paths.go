package agentconfig

import (
	"os"
	"path/filepath"
	"strings"
)

func isSensitive(zipPath string) bool {
	return zipPath == "pi/auth.json" ||
		zipPath == "opencode/opencode.jsonc" ||
		zipPath == "crush/config/crush.json"
}

type zipDest struct {
	destPath string
	perm     os.FileMode
}

func resolveDest(zipPath string, homeDir string) *zipDest {
	if strings.Contains(zipPath, "..") {
		return nil
	}
	if strings.HasPrefix(zipPath, "opencode/") {
		rest := strings.TrimPrefix(zipPath, "opencode/")
		if zipPath == "opencode/opencode.jsonc" {
			return &zipDest{
				destPath: filepath.Join(homeDir, ".config", "opencode", "opencode.jsonc"),
				perm:     0600,
			}
		}
		if strings.HasPrefix(rest, "plugins/") {
			rel := strings.TrimPrefix(rest, "plugins/")
			return &zipDest{
				destPath: filepath.Join(homeDir, ".config", "opencode", "plugins", rel),
				perm:     0644,
			}
		}
		if strings.HasPrefix(rest, "skills/") {
			rel := strings.TrimPrefix(rest, "skills/")
			return &zipDest{
				destPath: filepath.Join(homeDir, ".config", "opencode", "skills", rel),
				perm:     0644,
			}
		}
		return &zipDest{
			destPath: filepath.Join(homeDir, ".local", "share", "opencode", rest),
			perm:     0644,
		}
	}
	if strings.HasPrefix(zipPath, "pi/") {
		rel := strings.TrimPrefix(zipPath, "pi/")
		perm := os.FileMode(0644)
		if isSensitive(zipPath) {
			perm = 0600
		}
		return &zipDest{
			destPath: filepath.Join(homeDir, ".pi", "agent", rel),
			perm:     perm,
		}
	}
	if strings.HasPrefix(zipPath, "crush/") {
		rel := strings.TrimPrefix(zipPath, "crush/")
		perm := os.FileMode(0644)
		if isSensitive(zipPath) {
			perm = 0600
		}
		var destPath string
		switch rel {
		case "config/crush.json":
			destPath = filepath.Join(homeDir, ".config", "crush", "crush.json")
		case "data/crush.json":
			destPath = filepath.Join(homeDir, ".local", "share", "crush", "crush.json")
		default:
			return nil
		}
		return &zipDest{
			destPath: destPath,
			perm:     perm,
		}
	}
	return nil
}
