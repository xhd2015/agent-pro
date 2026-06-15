package subagent

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/xhd2015/agent-pro/agent/event/print"
)

type meta struct {
	ExplicitSessionID        string    `json:"explicit_session_id,omitempty"`
	SubagentRoleSessionID    string    `json:"subagent_role_session_id,omitempty"`
	MainAgentCodexThreadID   string    `json:"main_agent_codex_thread_id,omitempty"`
	OpencodeSessionID        string    `json:"opencode_session_id,omitempty"`
	AgentRunner              string    `json:"agent_runner,omitempty"`
	CreatedAt                time.Time `json:"created_at"`
}

func sessionsBase(c Config, opts Options) (string, error) {
	debugEnv := c.DebugSessionEnv
	if debugEnv == "" {
		debugEnv = "AGENT_PRO_SUBAGENT_DEBUG_SESSION_HOME"
	}
	if v := os.Getenv(debugEnv); v != "" {
		return v, nil
	}
	if opts.SessionBase != "" {
		return filepath.Join(opts.SessionBase, c.RoleName, "sessions"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("get home dir: %w", err)
	}
	return filepath.Join(home, ".agent-pro", "subagent", c.RoleName, "sessions"), nil
}

func findOrCreateSession(c Config, opts Options, threadID string, srcs *sessionIDSources) (dir string, isNew bool, err error) {
	base, err := sessionsBase(c, opts)
	if err != nil {
		return "", false, err
	}

	dir, _ = findSession(c, base, threadID, srcs)
	if dir != "" {
		return dir, false, nil
	}

	dir, err = createSession(c, base, threadID, srcs)
	if err != nil {
		return "", false, err
	}
	return dir, true, nil
}

func findSession(c Config, base, threadID string, srcs *sessionIDSources) (string, bool) {
	matchField := sessionMatchField(c, srcs)

	if dir := findSessionByField(base, threadID, matchField); dir != "" {
		return dir, false
	}

	if srcs.explicitSessionID != "" && matchField != "main_agent_codex_thread_id" {
		if dir := findSessionByField(base, threadID, "main_agent_codex_thread_id"); dir != "" {
			return dir, true
		}
	}

	return "", false
}

func findSessionByField(base, threadID, matchField string) string {
	entries, err := os.ReadDir(base)
	if err == nil {
		for _, entry := range entries {
			if !entry.IsDir() || !strings.HasPrefix(entry.Name(), "sess_") {
				continue
			}
			sessDir := filepath.Join(base, entry.Name())
			metaPath := filepath.Join(sessDir, "meta.json")
			data, err := os.ReadFile(metaPath)
			if err != nil {
				continue
			}
			var m map[string]any
			if err := json.Unmarshal(data, &m); err != nil {
				continue
			}
			if v, ok := m[matchField]; ok {
				if s, ok := v.(string); ok && s == threadID {
					return sessDir
				}
			}
		}
	}

	today := time.Now()
	for i := 0; i < 7; i++ {
		dateDir := today.AddDate(0, 0, -i).Format("2006/01/02")
		dayPath := filepath.Join(base, dateDir)
		entries, err := os.ReadDir(dayPath)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if !entry.IsDir() || !strings.HasPrefix(entry.Name(), "sess_") {
				continue
			}
			sessDir := filepath.Join(dayPath, entry.Name())
			metaPath := filepath.Join(sessDir, "meta.json")
			data, err := os.ReadFile(metaPath)
			if err != nil {
				continue
			}
			var m map[string]any
			if err := json.Unmarshal(data, &m); err != nil {
				continue
			}
			if v, ok := m[matchField]; ok {
				if s, ok := v.(string); ok && s == threadID {
					return sessDir
				}
			}
		}
	}
	return ""
}

func sessionMatchField(c Config, srcs *sessionIDSources) string {
	if srcs.explicitSessionID != "" {
		return "explicit_session_id"
	}
	if srcs.implSessionID != "" {
		return c.metaSessionField()
	}
	return "main_agent_codex_thread_id"
}

func createSession(c Config, base, threadID string, srcs *sessionIDSources) (string, error) {
	now := time.Now()
	dateDir := now.Format("2006/01/02")
	sessID := fmt.Sprintf("sess_%s_%d", now.Format("150405"), now.UnixNano())
	sessDir := filepath.Join(base, dateDir, sessID)

	if err := os.MkdirAll(sessDir, 0755); err != nil {
		return "", fmt.Errorf("create session dir: %w", err)
	}

	m := map[string]any{
		"agent_runner": srcs.agentRunner,
		"created_at":   now,
	}
	if srcs.explicitSessionID != "" {
		m["explicit_session_id"] = srcs.explicitSessionID
	}
	if srcs.codexThreadID != "" {
		m["main_agent_codex_thread_id"] = srcs.codexThreadID
	}
	if srcs.implSessionID != "" {
		m[c.metaSessionField()] = srcs.implSessionID
	}

	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal meta: %w", err)
	}
	metaPath := filepath.Join(sessDir, "meta.json")
	if err := os.WriteFile(metaPath, append(data, '\n'), 0644); err != nil {
		return "", fmt.Errorf("write meta.json: %w", err)
	}

	return sessDir, nil
}

func writeSessionPID(sessionDir string) error {
	pid := os.Getpid()
	return os.WriteFile(filepath.Join(sessionDir, "pid"), []byte(fmt.Sprintf("%d", pid)), 0644)
}

func removeSessionPID(sessionDir string) {
	os.Remove(filepath.Join(sessionDir, "pid"))
}

func isSessionLive(sessionDir string) bool {
	data, err := os.ReadFile(filepath.Join(sessionDir, "pid"))
	if err != nil {
		return false
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		return false
	}
	return processExists(pid)
}

func readOpencodeSessionID(sessionDir string) string {
	metaPath := filepath.Join(sessionDir, "meta.json")
	data, err := os.ReadFile(metaPath)
	if err != nil {
		return ""
	}
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		return ""
	}
	s, _ := m["opencode_session_id"].(string)
	return s
}

func updateSessionMeta(sessionDir, innerSessionID string, srcs *sessionIDSources) error {
	metaPath := filepath.Join(sessionDir, "meta.json")
	data, err := os.ReadFile(metaPath)
	if err != nil {
		return err
	}
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}
	m["opencode_session_id"] = innerSessionID
	if srcs != nil && srcs.codexThreadID != "" {
		m["main_agent_codex_thread_id"] = srcs.codexThreadID
	}
	newData, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(metaPath, append(newData, '\n'), 0644)
}

func showStatus(c Config, opts Options) error {
	srcs, err := resolveSessionID(c, opts.SessionID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return nil
	}

	base, err := sessionsBase(c, opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return nil
	}

	sessionDir, codexFallback := findSession(c, base, srcs.sessionID, srcs)
	if sessionDir == "" {
		fmt.Fprintf(os.Stderr, "error: session not found: %s\n", srcs.sessionID)
		return nil
	}

	if codexFallback {
		fmt.Fprintf(os.Stdout, "from --session-id, matching CODEX_THREAD_ID\n")
	}

	metaPath := filepath.Join(sessionDir, "meta.json")
	metaData, err := os.ReadFile(metaPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: read meta.json: %v\n", err)
		return nil
	}

	var metaMap map[string]any
	if err := json.Unmarshal(metaData, &metaMap); err != nil {
		fmt.Fprintf(os.Stderr, "error: invalid meta.json: %v\n", err)
		return nil
	}

	sesID, _ := metaMap["explicit_session_id"].(string)
	if sesID == "" {
		sesID = srcs.sessionID
	}

	runner, _ := metaMap["agent_runner"].(string)
	if runner == "" {
		runner = "opencode"
	}

	codex, _ := metaMap["main_agent_codex_thread_id"].(string)
	opencodeSID, _ := metaMap["opencode_session_id"].(string)

	createdAtStr, _ := metaMap["created_at"].(string)
	if t, err := time.Parse(time.RFC3339, createdAtStr); err == nil {
		createdAtStr = t.Format("2006-01-02 15:04:05")
	}

	eventsPath := filepath.Join(sessionDir, "events.jsonl")
	var eventLines []string
	var lastTimestampMs int64
	eventsData, evErr := os.ReadFile(eventsPath)
	if evErr == nil {
		for _, line := range strings.Split(string(eventsData), "\n") {
			line = strings.TrimSpace(line)
			if line != "" && strings.HasPrefix(line, "{") {
				eventLines = append(eventLines, line)
			}
		}
		if len(eventLines) > 0 {
			lastTimestampMs = parseEventTimestamp(eventLines[len(eventLines)-1])
		}
	}

	eventCount := len(eventLines)
	lastRelative := "\u2014"
	if lastTimestampMs > 0 {
		lastRelative = relativeTime(lastTimestampMs)
	}

	status := "finished"
	if isSessionLive(sessionDir) {
		pidData, _ := os.ReadFile(filepath.Join(sessionDir, "pid"))
		pidStr := strings.TrimSpace(string(pidData))
		status = fmt.Sprintf("running (PID %s)", pidStr)
	}

	fmt.Fprintf(os.Stdout, "\n\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\n")
	fmt.Fprintf(os.Stdout, "  Session:  %s\n", sesID)
	fmt.Fprintf(os.Stdout, "  Status:   %s\n", status)
	fmt.Fprintf(os.Stdout, "  Runner:   %s\n", runner)
	fmt.Fprintf(os.Stdout, "  Created:  %s\n", createdAtStr)

	codexDisplay := codex
	if codexDisplay == "" {
		codexDisplay = "\u2014"
	}
	fmt.Fprintf(os.Stdout, "  Codex:    %s\n", codexDisplay)
	opencodeDisplay := opencodeSID
	if opencodeDisplay == "" {
		opencodeDisplay = "\u2014"
	}
	fmt.Fprintf(os.Stdout, "  Opencode: %s\n", opencodeDisplay)
	fmt.Fprintf(os.Stdout, "  Events:  %d lines (last: %s)\n", eventCount, lastRelative)
	fmt.Fprintf(os.Stdout, "\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\n\n")

	if eventCount == 0 {
		Logf("No events yet")
	} else {
		lastEvents := eventLines
		if len(eventLines) > 3 {
			lastEvents = eventLines[len(eventLines)-3:]
		}
		for i, line := range lastEvents {
			formatted := print.FormatTraceLine(line)
			if formatted == "" {
				formatted = line
			}
			ts := parseEventTimestamp(line)
			rel := "\u2014"
			if ts > 0 {
				rel = relativeTime(ts)
			}
			lines := strings.Split(formatted, "\n")
			Logf("  [%d] %s \u2014 %s", eventCount-len(lastEvents)+i+1, lines[0], rel)
			for _, l := range lines[1:] {
				Logf("       %s", l)
			}
		}
	}

	return nil
}

func listSessions(c Config, opts Options) error {
	base, err := sessionsBase(c, opts)
	if err != nil {
		return err
	}

	type sessionInfo struct {
		ID        string
		Runner    string
		CreatedAt time.Time
	}

	var sessions []sessionInfo

	collectSessions := func(dirs []os.DirEntry, basePath string) {
		for _, entry := range dirs {
			if !entry.IsDir() || !strings.HasPrefix(entry.Name(), "sess_") {
				continue
			}
			sessDir := filepath.Join(basePath, entry.Name())
			metaPath := filepath.Join(sessDir, "meta.json")
			data, err := os.ReadFile(metaPath)
			if err != nil {
				continue
			}
			var m map[string]any
			if err := json.Unmarshal(data, &m); err != nil {
				continue
			}

			sid, _ := m["explicit_session_id"].(string)
			if sid == "" {
				sid = entry.Name()
			}
			runner, _ := m["agent_runner"].(string)
			if runner == "" {
				runner = "opencode"
			}

			var createdAt time.Time
			if ts, ok := m["created_at"].(string); ok {
				createdAt, _ = time.Parse(time.RFC3339, ts)
			}

			sessions = append(sessions, sessionInfo{
				ID:        sid,
				Runner:    runner,
				CreatedAt: createdAt,
			})
		}
	}

	entries, err := os.ReadDir(base)
	if err == nil {
		collectSessions(entries, base)
	}

	today := time.Now()
	for i := 0; i < 7; i++ {
		dateDir := today.AddDate(0, 0, -i).Format("2006/01/02")
		dayPath := filepath.Join(base, dateDir)
		entries, err := os.ReadDir(dayPath)
		if err != nil {
			continue
		}
		collectSessions(entries, dayPath)
	}

	if len(sessions) == 0 {
		fmt.Println("No sessions found")
		return nil
	}

	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].CreatedAt.After(sessions[j].CreatedAt)
	})

	fmt.Printf("\nSessions (%d):\n\n", len(sessions))
	for _, s := range sessions {
		timeStr := s.CreatedAt.Format("2006-01-02 15:04:05")
		fmt.Printf("%-15s %-10s %s\n", s.ID, s.Runner, timeStr)
	}

	return nil
}
