package groktty

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"

	grok_session "github.com/xhd2015/agent-pro/agent/event/grok_session"
)

const (
	envGrokHome              = "GROK_HOME"
	envGrokTTYGrokSessionID  = "AGENT_RUN_GROK_TTY_GROK_SESSION_ID"
	sessionDiscoveryInterval = 150 * time.Millisecond
	// Grace before runStart so grok session dirs created just before our clock still match.
	sessionDiscoveryGrace = 2 * time.Second
)

// GrokHome returns the grok data directory ($GROK_HOME or $HOME/.grok).
func GrokHome() string {
	if v := strings.TrimSpace(os.Getenv(envGrokHome)); v != "" {
		return v
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(os.TempDir(), ".grok")
	}
	return filepath.Join(home, ".grok")
}

// EncodedCwd returns url.PathEscape(absWorkspace) matching grok session layout.
func EncodedCwd(workspace string) (string, error) {
	abs, err := filepath.Abs(canonicalWorkspacePath(workspace))
	if err != nil {
		return "", err
	}
	return encodeGrokCwd(abs), nil
}

func encodeGrokCwd(abs string) string {
	// url.PathEscape is imported via net/url in discover — use inline to avoid
	// pulling net/url into every caller; implemented in discoverSession.
	return pathEscape(abs)
}

func sessionRoot(grokHome, workspace string) (string, error) {
	encoded, err := EncodedCwd(workspace)
	if err != nil {
		return "", err
	}
	return filepath.Join(grokHome, "sessions", encoded), nil
}

func sessionDir(grokHome, workspace, sessionID string) (string, error) {
	root, err := sessionRoot(grokHome, workspace)
	if err != nil {
		return "", err
	}
	return filepath.Join(root, sessionID), nil
}

func findUpdatesBySessionID(grokHome, workspace, sessionID string) (string, bool) {
	for _, ws := range workspacePathVariants(workspace) {
		path, err := updatesJSONLPath(grokHome, ws, sessionID)
		if err != nil {
			continue
		}
		if _, err := os.Stat(path); err == nil {
			return path, true
		}
	}
	sessionsRoot := filepath.Join(grokHome, "sessions")
	entries, err := os.ReadDir(sessionsRoot)
	if err != nil {
		return "", false
	}
	for _, ent := range entries {
		if !ent.IsDir() {
			continue
		}
		path := filepath.Join(sessionsRoot, ent.Name(), sessionID, "updates.jsonl")
		if _, err := os.Stat(path); err == nil {
			return path, true
		}
	}
	return "", false
}

func workspacePathVariants(workspace string) []string {
	seen := make(map[string]bool)
	var out []string
	for _, candidate := range []string{workspace, canonicalWorkspacePath(workspace)} {
		candidate = strings.TrimSpace(candidate)
		if candidate == "" || seen[candidate] {
			continue
		}
		seen[candidate] = true
		out = append(out, candidate)
		if abs, err := filepath.Abs(candidate); err == nil && !seen[abs] {
			seen[abs] = true
			out = append(out, abs)
		}
	}
	return out
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

func updatesJSONLPath(grokHome, workspace, sessionID string) (string, error) {
	dir, err := sessionDir(grokHome, workspace, sessionID)
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "updates.jsonl"), nil
}

// DiscoverSession locates the grok on-disk session matching workspace cwd and prompt.
// When AGENT_RUN_GROK_TTY_GROK_SESSION_ID is set, discovery is skipped.
func DiscoverSession(ctx context.Context, grokHome, workspace, prompt string, runStart time.Time) (sessionID string, updatesPath string, err error) {
	if hook := strings.TrimSpace(os.Getenv(envGrokTTYGrokSessionID)); hook != "" {
		if path, ok := findUpdatesBySessionID(grokHome, workspace, hook); ok {
			return hook, path, nil
		}
	}

	root, err := sessionRoot(grokHome, workspace)
	if err != nil {
		return "", "", err
	}

	for {
		select {
		case <-ctx.Done():
			return "", "", ctx.Err()
		default:
		}

		// Prefer time-bounded active_sessions (newest opened_at >= runStart) before
		// scanning on-disk dirs — avoids matching a prior session with the same prompt.
		if id, path, ok := discoverFromActiveSessions(grokHome, workspace, prompt, runStart); ok {
			return id, path, nil
		}
		if id, path, ok := scanSessionsForMatch(root, workspace, prompt, runStart); ok {
			return id, path, nil
		}

		select {
		case <-ctx.Done():
			return "", "", ctx.Err()
		case <-time.After(sessionDiscoveryInterval):
		}
	}
}

func sessionNotBefore(runStart, sessionTime time.Time) bool {
	if sessionTime.IsZero() {
		return false
	}
	cutoff := runStart.Add(-sessionDiscoveryGrace)
	return !sessionTime.Before(cutoff)
}

func scanSessionsForMatch(root, workspace, prompt string, runStart time.Time) (sessionID, updatesPath string, ok bool) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return "", "", false
	}
	absWorkspace, _ := filepath.Abs(canonicalWorkspacePath(workspace))
	wantPrompt := strings.TrimSpace(prompt)

	var bestID, bestPath string
	var bestCreated time.Time
	for _, ent := range entries {
		if !ent.IsDir() {
			continue
		}
		id := ent.Name()
		dir := filepath.Join(root, id)
		if !summaryCwdMatches(dir, absWorkspace) {
			continue
		}
		created, hasCreated := summaryCreatedAt(dir)
		if !hasCreated || !sessionNotBefore(runStart, created) {
			continue
		}
		path := filepath.Join(dir, "updates.jsonl")
		if !promptMatchesSession(path, wantPrompt) {
			continue
		}
		if bestID == "" || created.After(bestCreated) {
			bestID = id
			bestPath = path
			bestCreated = created
		}
	}
	if bestID == "" {
		return "", "", false
	}
	return bestID, bestPath, true
}

type grokSummary struct {
	Info struct {
		Cwd string `json:"cwd"`
	} `json:"info"`
	CreatedAt string `json:"created_at"`
}

func readSummary(sessionDir string) (grokSummary, bool) {
	data, err := os.ReadFile(filepath.Join(sessionDir, "summary.json"))
	if err != nil {
		return grokSummary{}, false
	}
	var summary grokSummary
	if err := json.Unmarshal(data, &summary); err != nil {
		return grokSummary{}, false
	}
	return summary, true
}

func summaryCwdMatches(sessionDir, absWorkspace string) bool {
	summary, ok := readSummary(sessionDir)
	if !ok {
		return false
	}
	cwd, err := filepath.Abs(canonicalWorkspacePath(strings.TrimSpace(summary.Info.Cwd)))
	if err != nil {
		cwd = canonicalWorkspacePath(strings.TrimSpace(summary.Info.Cwd))
	}
	return cwd == absWorkspace
}

func summaryCreatedAt(sessionDir string) (time.Time, bool) {
	summary, ok := readSummary(sessionDir)
	if !ok {
		return time.Time{}, false
	}
	return parseGrokTimestamp(summary.CreatedAt)
}

func parseGrokTimestamp(raw string) (time.Time, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return time.Time{}, false
	}
	if t, err := time.Parse(time.RFC3339Nano, raw); err == nil {
		return t, true
	}
	if t, err := time.Parse(time.RFC3339, raw); err == nil {
		return t, true
	}
	return time.Time{}, false
}

func promptMatchesSession(updatesPath, wantPrompt string) bool {
	first, found := firstUserMessageChunk(updatesPath)
	return found && strings.TrimSpace(first) == strings.TrimSpace(wantPrompt)
}

func firstUserMessageChunk(updatesPath string) (string, bool) {
	data, err := os.ReadFile(updatesPath)
	if err != nil {
		return "", false
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		upd, ok := grok_session.ParseLine(line)
		if !ok || upd.SessionUpdate != "user_message_chunk" {
			continue
		}
		text := grok_session.TextContent(upd.Content)
		if text != "" {
			return text, true
		}
	}
	return "", false
}

type activeSessionsFile struct {
	Sessions []activeSessionEntry `json:"sessions"`
}

type activeSessionEntry struct {
	SessionID      string `json:"sessionId"`
	SessionIDSnake string `json:"session_id"`
	Cwd            string `json:"cwd"`
	OpenedAt       string `json:"openedAt"`
	OpenedAtSnake  string `json:"opened_at"`
}

func (e activeSessionEntry) resolvedSessionID() string {
	if id := strings.TrimSpace(e.SessionID); id != "" {
		return id
	}
	return strings.TrimSpace(e.SessionIDSnake)
}

func (e activeSessionEntry) resolvedOpenedAt() string {
	if t := strings.TrimSpace(e.OpenedAt); t != "" {
		return t
	}
	return strings.TrimSpace(e.OpenedAtSnake)
}

func discoverFromActiveSessions(grokHome, workspace, prompt string, runStart time.Time) (sessionID, updatesPath string, ok bool) {
	path := filepath.Join(grokHome, "active_sessions.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return "", "", false
	}
	var file activeSessionsFile
	if err := json.Unmarshal(data, &file); err != nil {
		// tolerate array root
		var entries []activeSessionEntry
		if err2 := json.Unmarshal(data, &entries); err2 != nil {
			return "", "", false
		}
		file.Sessions = entries
	}

	absWorkspace, _ := filepath.Abs(canonicalWorkspacePath(workspace))
	var bestID string
	var bestOpened time.Time
	for _, ent := range file.Sessions {
		cwd := canonicalWorkspacePath(strings.TrimSpace(ent.Cwd))
		if cwd != absWorkspace {
			continue
		}
		openedAt := ent.resolvedOpenedAt()
		if openedAt == "" {
			continue
		}
		opened, err := time.Parse(time.RFC3339Nano, openedAt)
		if err != nil {
			opened, err = time.Parse(time.RFC3339, openedAt)
			if err != nil {
				continue
			}
		}
		if !sessionNotBefore(runStart, opened) {
			continue
		}
		sessionID := ent.resolvedSessionID()
		if sessionID == "" {
			continue
		}
		if bestID == "" || opened.After(bestOpened) {
			bestID = sessionID
			bestOpened = opened
		}
	}
	if bestID == "" {
		return "", "", false
	}
	updates, ok := findUpdatesBySessionID(grokHome, workspace, bestID)
	if !ok {
		return "", "", false
	}
	if !promptMatchesSession(updates, prompt) {
		return "", "", false
	}
	return bestID, updates, true
}