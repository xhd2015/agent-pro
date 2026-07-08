package agenttty

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	grok_session "github.com/xhd2015/agent-pro/agent/event/grok_session"
	types "github.com/xhd2015/agent-pro/agent/event/types"
	"github.com/xhd2015/dot-pkgs/go-pkgs/logs"
)

const (
	envGrokTTYGrokSessionID  = "AGENT_RUN_GROK_TTY_GROK_SESSION_ID"
	sessionDiscoveryInterval = 150 * time.Millisecond
	tailPollInterval         = 150 * time.Millisecond
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

func updatesTailStartOffset(path string, runStart time.Time) int64 {
	info, err := os.Stat(path)
	if err != nil {
		return 0
	}
	cutoff := runStart.Add(-sessionDiscoveryGrace)
	if info.ModTime().Before(cutoff) {
		return info.Size()
	}
	return 0
}

// TailUpdatesFromOffset tails updates.jsonl starting at the given byte offset.
// onInitialSynced, when non-nil, is invoked after the first batch of updates is emitted.
func TailUpdatesFromOffset(ctx context.Context, updatesPath string, startOffset int64, emit func(types.AgentEvent) error, onInitialSynced ...func()) error {
	converter := grok_session.NewConverter()
	emitEvent := func(ev types.AgentEvent) error {
		return emit(trimACPAgentEventNewlines(ev))
	}
	var synced func()
	if len(onInitialSynced) > 0 {
		synced = onInitialSynced[0]
	}

	if _, err := readNewACPUpdates(updatesPath, startOffset, converter, emitEvent); err != nil {
		return err
	}
	for _, ev := range converter.Flush() {
		if err := emitEvent(ev); err != nil {
			return err
		}
	}
	if synced != nil {
		synced()
	}

	watchErr := logs.WatchLine(ctx, updatesPath, logs.WatchLineOptions{DisableDebounce: true}, func(line string) error {
		return processACPUpdateLine(line, converter, emitEvent)
	})

	for _, ev := range converter.Flush() {
		if err := emitEvent(ev); err != nil {
			return err
		}
	}

	if watchErr != nil {
		return watchErr
	}
	return ctx.Err()
}

func trimACPAgentEventNewlines(ev types.AgentEvent) types.AgentEvent {
	ev.Text = strings.TrimRight(ev.Text, "\n")
	ev.Output = strings.TrimRight(ev.Output, "\n")
	return ev
}

func processACPUpdateLine(line string, converter *grok_session.Converter, emit func(types.AgentEvent) error) error {
	line = strings.TrimRight(line, "\n")
	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return nil
	}
	if _, ok := grok_session.ParseLine(trimmed); !ok {
		return nil
	}
	for _, ev := range converter.ProcessLine(line) {
		if err := emit(ev); err != nil {
			return err
		}
	}
	for _, ev := range converter.Flush() {
		if err := emit(ev); err != nil {
			return err
		}
	}
	return nil
}

func readNewACPUpdates(path string, offset int64, converter *grok_session.Converter, emit func(types.AgentEvent) error) (int64, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return offset, nil
		}
		return offset, err
	}
	defer f.Close()

	if _, err := f.Seek(offset, 0); err != nil {
		return offset, err
	}

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 256*1024), 2*1024*1024)
	for scanner.Scan() {
		if err := processACPUpdateLine(scanner.Text(), converter, emit); err != nil {
			return offset, err
		}
	}
	if err := scanner.Err(); err != nil {
		return offset, err
	}
	newOffset, err := f.Seek(0, 1)
	if err != nil {
		return offset, err
	}
	return newOffset, nil
}

func sessionRoot(grokHome, workspace string) (string, error) {
	encoded, err := encodedCwd(workspace)
	if err != nil {
		return "", err
	}
	return filepath.Join(grokHome, "sessions", encoded), nil
}

func encodedCwd(workspace string) (string, error) {
	abs, err := filepath.Abs(canonicalWorkspacePath(workspace))
	if err != nil {
		return "", err
	}
	return pathEscape(abs), nil
}

func sessionDir(grokHome, workspace, sessionID string) (string, error) {
	root, err := sessionRoot(grokHome, workspace)
	if err != nil {
		return "", err
	}
	return filepath.Join(root, sessionID), nil
}

// FindUpdatesBySessionID locates updates.jsonl for an existing grok session id.
func FindUpdatesBySessionID(grokHome, workspace, sessionID string) (string, bool) {
	return findUpdatesBySessionID(grokHome, workspace, sessionID)
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

func updatesJSONLPath(grokHome, workspace, sessionID string) (string, error) {
	dir, err := sessionDir(grokHome, workspace, sessionID)
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "updates.jsonl"), nil
}

// UpdatesTailOffset returns the current byte size of a grok session updates.jsonl.
func UpdatesTailOffset(grokHome, workspace, grokSessionID string) int64 {
	path, ok := findUpdatesBySessionID(grokHome, workspace, grokSessionID)
	if !ok {
		return 0
	}
	info, err := os.Stat(path)
	if err != nil {
		return 0
	}
	return info.Size()
}

// TailGrokSessionFromOffset tails updates.jsonl for an existing grok session id.
func TailGrokSessionFromOffset(ctx context.Context, grokHome, workspace, grokSessionID string, startOffset int64, emit func(types.AgentEvent) error) error {
	path, ok := findUpdatesBySessionID(grokHome, workspace, grokSessionID)
	if !ok {
		return fmt.Errorf("grok updates not found for session %s", grokSessionID)
	}
	return TailUpdatesFromOffset(ctx, path, startOffset, emit, nil)
}

// DiscoverSession locates the grok on-disk session matching workspace cwd and prompt.
func DiscoverSession(ctx context.Context, grokHome, workspace, prompt string, runStart time.Time) (sessionID string, updatesPath string, err error) {
	hook := strings.TrimSpace(os.Getenv(envGrokTTYGrokSessionID))
	if hook != "" {
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

		// Re-check a pinned grok session id each poll — updates.jsonl may appear after the worker starts.
		if hook != "" {
			if path, ok := findUpdatesBySessionID(grokHome, workspace, hook); ok {
				return hook, path, nil
			}
		}

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