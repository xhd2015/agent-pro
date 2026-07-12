package agenttty

import (
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// sessionDiscoveryGrace is how far before runStart a grok session's created_at
// may still be accepted by DiscoverSession / prompt fallback.
//
// Must tolerate:
//   - Setup preseed → build/copy lag → agent-run start (parallel doctests)
//   - Delayed materialize finishing slightly before a starved process sets runStart
//   - Real-world process start lag after parent schedules discovery
//
// Too tight (e.g. 2s) causes parallel open-bind flakes: session exists on disk
// but is filtered forever → post-detach grace cancels with "not resolved".
// 1m is enough for Setup/build lag and delayed materialize under parallel load
// without a multi-hour same-prompt collision window.
const sessionDiscoveryGrace = 1 * time.Minute

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