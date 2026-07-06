package agenttty

import (
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const sessionDiscoveryGrace = 2 * time.Second

func pathEscape(s string) string {
	return url.PathEscape(s)
}

func canonicalWorkspacePath(path string) string {
	path = filepath.Clean(path)
	if !filepath.IsAbs(path) {
		if abs, err := filepath.Abs(path); err == nil {
			path = filepath.Clean(abs)
		}
	}
	if strings.HasPrefix(path, "/private/") {
		alt := strings.TrimPrefix(path, "/private")
		if alt != "" && alt[0] == '/' {
			if a, errA := os.Lstat(path); errA == nil {
				if b, errB := os.Lstat(alt); errB == nil && os.SameFile(a, b) {
					return filepath.Clean(alt)
				}
			}
		}
	}
	return path
}

func sessionNotBefore(runStart, sessionTime time.Time) bool {
	if sessionTime.IsZero() {
		return false
	}
	cutoff := runStart.Add(-sessionDiscoveryGrace)
	return !sessionTime.Before(cutoff)
}