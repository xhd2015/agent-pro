package sessions

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const (
	defaultListLimit = 20
	maxListLimit     = 100
)

type Session struct {
	ID              string
	LastActiveAt    time.Time
	CWD             string
	Title           string
	Path            string
	NumChatMessages int
}

type SessionInfo struct {
	Session
	CreatedAt, UpdatedAt time.Time
	NumMessages          int
	CurrentModelID       string
	AgentName            string
	SandboxProfile       string
	GitRootDir           string
	HeadBranch           string
	HeadCommit           string
	SessionDir           string
	SummaryPath          string
	UpdatesPath          string
	SignalsPath          string
	PromptContextPath    string
	UpdatesExists        bool
	SignalsExists        bool
	PromptContextExists  bool
	ContextTokensUsed    int
	ContextWindowTokens  int
	ContextWindowUsage   int
	TotalTokensBeforeCompaction int
}

type grokSummary struct {
	Info struct {
		ID  string `json:"id"`
		CWD string `json:"cwd"`
	} `json:"info"`
	SessionSummary   string `json:"session_summary"`
	GeneratedTitle   string `json:"generated_title"`
	CreatedAt        string `json:"created_at"`
	UpdatedAt        string `json:"updated_at"`
	LastActiveAt     string `json:"last_active_at"`
	NumMessages      int    `json:"num_messages"`
	NumChatMessages  int    `json:"num_chat_messages"`
	CurrentModelID   string `json:"current_model_id"`
	AgentName        string `json:"agent_name"`
	SandboxProfile   string `json:"sandbox_profile"`
	GitRootDir       string `json:"git_root_dir"`
	HeadBranch       string `json:"head_branch"`
	HeadCommit       string `json:"head_commit"`
}

type grokSignals struct {
	ContextTokensUsed           int `json:"contextTokensUsed"`
	ContextWindowTokens         int `json:"contextWindowTokens"`
	ContextWindowUsage          int `json:"contextWindowUsage"`
	TotalTokensBeforeCompaction int `json:"totalTokensBeforeCompaction"`
}

func List(grokHome string, limit int) ([]Session, error) {
	if limit <= 0 {
		limit = defaultListLimit
	}
	if limit > maxListLimit {
		limit = maxListLimit
	}

	sessions, err := discoverSessions(grokHome)
	if err != nil {
		return nil, err
	}

	sort.Slice(sessions, func(i, j int) bool {
		if sessions[i].LastActiveAt.Equal(sessions[j].LastActiveAt) {
			return sessions[i].ID > sessions[j].ID
		}
		return sessions[i].LastActiveAt.After(sessions[j].LastActiveAt)
	})

	if len(sessions) > limit {
		sessions = sessions[:limit]
	}
	return sessions, nil
}

func Find(grokHome, sessionID string) (Session, error) {
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return Session{}, sessionNotFoundError(sessionID)
	}

	root := filepath.Join(grokHome, "sessions")
	var found Session
	foundOK := false

	err := filepath.WalkDir(root, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			if os.IsNotExist(walkErr) {
				return nil
			}
			return walkErr
		}
		if d.IsDir() || d.Name() != "summary.json" {
			return nil
		}

		session, ok := parseSummaryFile(path)
		if !ok {
			return nil
		}
		if session.ID == sessionID {
			found = session
			foundOK = true
			return filepath.SkipAll
		}
		return nil
	})
	if err != nil {
		if os.IsNotExist(err) {
			return Session{}, sessionNotFoundError(sessionID)
		}
		return Session{}, err
	}
	if !foundOK {
		return Session{}, sessionNotFoundError(sessionID)
	}
	return found, nil
}

func Info(grokHome, sessionID string) (*SessionInfo, error) {
	session, err := Find(grokHome, sessionID)
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(session.Path)
	if err != nil {
		return nil, err
	}
	data = []byte(strings.TrimSpace(string(data)))
	if len(data) == 0 {
		return nil, sessionNotFoundError(sessionID)
	}

	var summary grokSummary
	if err := json.Unmarshal(data, &summary); err != nil {
		return nil, fmt.Errorf("parse summary for %s: %w", sessionID, err)
	}

	createdAt, _ := parseTimestamp(summary.CreatedAt)
	updatedAt, _ := parseTimestamp(summary.UpdatedAt)

	sessionDir := filepath.Dir(session.Path)
	summaryPath := session.Path
	updatesPath := filepath.Join(sessionDir, "updates.jsonl")
	signalsPath := filepath.Join(sessionDir, "signals.json")
	promptContextPath := filepath.Join(sessionDir, "prompt_context.json")

	info := &SessionInfo{
		Session:           session,
		CreatedAt:         createdAt,
		UpdatedAt:         updatedAt,
		NumMessages:       summary.NumMessages,
		CurrentModelID:    strings.TrimSpace(summary.CurrentModelID),
		AgentName:         strings.TrimSpace(summary.AgentName),
		SandboxProfile:    strings.TrimSpace(summary.SandboxProfile),
		GitRootDir:        strings.TrimSpace(summary.GitRootDir),
		HeadBranch:        strings.TrimSpace(summary.HeadBranch),
		HeadCommit:        strings.TrimSpace(summary.HeadCommit),
		SessionDir:        sessionDir,
		SummaryPath:       summaryPath,
		UpdatesPath:       updatesPath,
		SignalsPath:       signalsPath,
		PromptContextPath: promptContextPath,
	}
	info.NumChatMessages = summary.NumChatMessages

	info.UpdatesExists = fileExists(updatesPath)
	info.SignalsExists = fileExists(signalsPath)
	info.PromptContextExists = fileExists(promptContextPath)

	if info.SignalsExists {
		if signals, ok := parseSignalsFile(signalsPath); ok {
			info.ContextTokensUsed = signals.ContextTokensUsed
			info.ContextWindowTokens = signals.ContextWindowTokens
			info.ContextWindowUsage = signals.ContextWindowUsage
			info.TotalTokensBeforeCompaction = signals.TotalTokensBeforeCompaction
		}
	}

	return info, nil
}

func FormatListTable(sessions []Session, home string, now time.Time) string {
	if len(sessions) == 0 {
		return "No sessions found"
	}

	var b strings.Builder
	fmt.Fprintf(&b, "%-38s  %-12s  %-42s  %5s  %s\n", "SESSION ID", "LAST ACTIVE", "TITLE", "MSGS", "CWD")
	for _, session := range sessions {
		fmt.Fprintf(
			&b,
			"%-38s  %-12s  %-42s  %5d  %s\n",
			session.ID,
			formatRelativeTime(session.LastActiveAt, now),
			truncateTitle(session.Title),
			session.NumChatMessages,
			shortenPath(session.CWD, home),
		)
	}
	return strings.TrimRight(b.String(), "\n")
}

func FormatInfoText(info *SessionInfo, home string, now time.Time) string {
	if info == nil {
		return ""
	}

	var b strings.Builder
	fmt.Fprintf(&b, "Session: %s\n", info.ID)

	title := strings.TrimSpace(info.Title)
	if title == "" {
		title = "(untitled)"
	}
	fmt.Fprintf(&b, "Title: %s\n", title)

	fmt.Fprintf(
		&b,
		"Last active: %s (%s)\n",
		formatRelativeTime(info.LastActiveAt, now),
		formatAbsoluteTime(info.LastActiveAt),
	)
	if !info.CreatedAt.IsZero() {
		fmt.Fprintf(
			&b,
			"Created: %s (%s)\n",
			formatRelativeTime(info.CreatedAt, now),
			formatAbsoluteTime(info.CreatedAt),
		)
	}
	if info.CWD != "" {
		fmt.Fprintf(&b, "CWD: %s\n", info.CWD)
	}

	if info.CurrentModelID != "" {
		fmt.Fprintf(&b, "Model: %s\n", info.CurrentModelID)
	}
	if info.AgentName != "" {
		fmt.Fprintf(&b, "Agent: %s\n", info.AgentName)
	}
	fmt.Fprintf(&b, "Messages: %d total, %d chat\n", info.NumMessages, info.NumChatMessages)

	if info.HeadBranch != "" || info.HeadCommit != "" {
		fmt.Fprintf(&b, "Git: %s @ %s\n", info.HeadBranch, info.HeadCommit)
	}
	if info.SandboxProfile != "" {
		fmt.Fprintf(&b, "Sandbox: %s\n", info.SandboxProfile)
	}

	fmt.Fprintf(&b, "\nFiles:\n")
	fmt.Fprintf(&b, "  Dir:      %s\n", shortenPath(info.SessionDir, home))
	fmt.Fprintf(&b, "  Summary:  %s\n", shortenPath(info.SummaryPath, home))
	if info.UpdatesExists {
		fmt.Fprintf(&b, "  Updates:  %s\n", shortenPath(info.UpdatesPath, home))
	} else {
		fmt.Fprintf(&b, "  Updates:  %s (missing)\n", shortenPath(info.UpdatesPath, home))
	}
	if info.SignalsExists {
		fmt.Fprintf(&b, "  Signals:  %s\n", shortenPath(info.SignalsPath, home))
	}
	if info.PromptContextExists {
		fmt.Fprintf(&b, "  Prompt:   %s\n", shortenPath(info.PromptContextPath, home))
	}

	if hasTokenUsage(info) {
		fmt.Fprintf(&b, "\nTokens:\n")
		fmt.Fprintf(
			&b,
			"  Context: %d / %d (%d%%)\n",
			info.ContextTokensUsed,
			info.ContextWindowTokens,
			info.ContextWindowUsage,
		)
		fmt.Fprintf(&b, "  Before compaction: %d\n", info.TotalTokensBeforeCompaction)
	}

	return strings.TrimRight(b.String(), "\n")
}

func discoverSessions(grokHome string) ([]Session, error) {
	root := filepath.Join(grokHome, "sessions")
	var sessions []Session

	err := filepath.WalkDir(root, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			if os.IsNotExist(walkErr) {
				return nil
			}
			return walkErr
		}
		// Session layout: sessions/{cwdKey}/{uuid}/summary.json (+ terminal/, updates, …).
		// When we enter a uuid dir that already has summary.json, parse it and SkipDir so
		// we never walk terminal/ or other heavy children (dominant cost with large homes).
		if d.IsDir() {
			if path == root {
				return nil
			}
			sumPath := filepath.Join(path, "summary.json")
			if st, err := os.Stat(sumPath); err == nil && !st.IsDir() {
				if session, ok := parseSummaryFile(sumPath); ok {
					sessions = append(sessions, session)
				}
				return filepath.SkipDir
			}
			return nil
		}
		// Files are only reached under dirs without summary.json (e.g. cwdKey placeholders).
		if d.Name() == "summary.json" {
			if session, ok := parseSummaryFile(path); ok {
				sessions = append(sessions, session)
			}
		}
		return nil
	})
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	return sessions, nil
}

func parseSummaryFile(path string) (Session, bool) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Session{}, false
	}
	data = []byte(strings.TrimSpace(string(data)))
	if len(data) == 0 {
		return Session{}, false
	}

	var summary grokSummary
	if err := json.Unmarshal(data, &summary); err != nil {
		return Session{}, false
	}

	id := strings.TrimSpace(summary.Info.ID)
	if id == "" {
		return Session{}, false
	}

	lastActive, ok := resolveLastActive(summary)
	if !ok {
		return Session{}, false
	}

	title := strings.TrimSpace(summary.GeneratedTitle)
	if title == "" {
		title = strings.TrimSpace(summary.SessionSummary)
	}

	return Session{
		ID:              id,
		LastActiveAt:    lastActive,
		CWD:             strings.TrimSpace(summary.Info.CWD),
		Title:           title,
		Path:            path,
		NumChatMessages: summary.NumChatMessages,
	}, true
}

func parseSignalsFile(path string) (grokSignals, bool) {
	data, err := os.ReadFile(path)
	if err != nil {
		return grokSignals{}, false
	}
	data = []byte(strings.TrimSpace(string(data)))
	if len(data) == 0 {
		return grokSignals{}, false
	}

	var signals grokSignals
	if err := json.Unmarshal(data, &signals); err != nil {
		return grokSignals{}, false
	}
	return signals, true
}

func resolveLastActive(summary grokSummary) (time.Time, bool) {
	for _, raw := range []string{
		summary.LastActiveAt,
		summary.UpdatedAt,
		summary.CreatedAt,
	} {
		if ts, ok := parseTimestamp(raw); ok {
			return ts, true
		}
	}
	return time.Time{}, false
}

func parseTimestamp(raw string) (time.Time, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return time.Time{}, false
	}
	for _, layout := range []string{
		time.RFC3339Nano,
		"2006-01-02T15:04:05.000Z",
		time.RFC3339,
	} {
		if ts, err := time.Parse(layout, raw); err == nil {
			return ts, true
		}
	}
	return time.Time{}, false
}

func formatRelativeTime(lastActive, now time.Time) string {
	diff := now.Sub(lastActive)
	if diff < time.Minute {
		return "just now"
	}
	if diff < time.Hour {
		return fmt.Sprintf("%dm ago", int(diff.Minutes()))
	}
	if diff < 24*time.Hour {
		return fmt.Sprintf("%dh ago", int(diff.Hours()))
	}
	if diff < 7*24*time.Hour {
		return fmt.Sprintf("%dd ago", int(diff.Hours()/24))
	}
	return fmt.Sprintf("%dw ago", int(diff.Hours()/(24*7)))
}

func formatAbsoluteTime(ts time.Time) string {
	if ts.IsZero() {
		return ""
	}
	return ts.UTC().Format("2006-01-02 15:04:05 UTC")
}

func truncateTitle(title string) string {
	title = strings.TrimSpace(title)
	if len(title) <= 40 {
		return title
	}
	return title[:40] + "..."
}

func shortenPath(path, home string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return ""
	}
	if home != "" && strings.HasPrefix(path, home) {
		rel := strings.TrimPrefix(path, home)
		rel = strings.TrimPrefix(rel, string(os.PathSeparator))
		if rel != "" {
			return "~/" + rel
		}
	}
	return path
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func hasTokenUsage(info *SessionInfo) bool {
	if info == nil || !info.SignalsExists {
		return false
	}
	return info.ContextTokensUsed != 0 ||
		info.ContextWindowTokens != 0 ||
		info.ContextWindowUsage != 0
}

func sessionNotFoundError(id string) error {
	return fmt.Errorf("grok session not found: %s", id)
}