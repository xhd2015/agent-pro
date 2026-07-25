package procresolve

import (
	"path/filepath"
	"regexp"
	"strings"
)

// Session id shape: hyphenated hex groups. First group is {8,} so IDs like
// fixture "019fabcdef-…" (10-char head) still match as a whole, not a suffix.
var uuidPattern = regexp.MustCompile(`(?i)[0-9a-f]{8,}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}`)

// parseSessionFromPath extracts a hard session hit from an open-file path.
// Returns kind ("grok"|"codex"), sessionID, ok.
func parseSessionFromPath(path string) (kind, sessionID string, ok bool) {
	if path == "" {
		return "", "", false
	}
	// Normalize for marker search; keep original segments for uuid extraction.
	slashPath := filepath.ToSlash(path)

	if id, found := parseGrokSessionPath(slashPath); found {
		return "grok", id, true
	}
	if id, found := parseCodexSessionPath(slashPath); found {
		return "codex", id, true
	}
	return "", "", false
}

// longestUUID returns the longest uuidPattern match in s (prefer full IDs
// over suffix submatches when the head group is longer than 8 hex chars).
func longestUUID(s string) string {
	matches := uuidPattern.FindAllString(s, -1)
	best := ""
	for _, m := range matches {
		if len(m) > len(best) {
			best = m
		}
	}
	return best
}

// Grok: …/.grok/sessions/…/<uuid>/…
func parseGrokSessionPath(path string) (string, bool) {
	const marker = "/.grok/sessions/"
	idx := strings.Index(path, marker)
	if idx < 0 {
		if strings.HasPrefix(path, ".grok/sessions/") {
			path = "/" + path
			idx = strings.Index(path, marker)
		}
		if idx < 0 {
			return "", false
		}
	}
	rest := path[idx+len(marker):]
	// Prefer a full path segment that is the session id.
	for _, seg := range strings.Split(rest, "/") {
		if seg == "" {
			continue
		}
		if m := longestUUID(seg); m != "" && len(m) == len(seg) {
			return strings.ToLower(m), true
		}
	}
	if m := longestUUID(rest); m != "" {
		return strings.ToLower(m), true
	}
	return "", false
}

// Codex: …/.codex/sessions/…/rollout-…-<uuid>[.jsonl]
func parseCodexSessionPath(path string) (string, bool) {
	const marker = "/.codex/sessions/"
	idx := strings.Index(path, marker)
	if idx < 0 {
		if strings.HasPrefix(path, ".codex/sessions/") {
			path = "/" + path
			idx = strings.Index(path, marker)
		}
		if idx < 0 {
			return "", false
		}
	}
	base := filepath.Base(path)
	// Prefer uuid from rollout-* filename (longest match).
	if strings.Contains(base, "rollout-") {
		if m := longestUUID(base); m != "" {
			return strings.ToLower(m), true
		}
	}
	rest := path[idx+len(marker):]
	if m := longestUUID(rest); m != "" {
		return strings.ToLower(m), true
	}
	return "", false
}
