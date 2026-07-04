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
	ID           string
	LastActiveAt time.Time
	CWD          string
	Title        string
	Path         string
	NumMessages  int
}

type SessionInfo struct {
	Session
	CreatedAt          time.Time
	UpdatedAt          time.Time
	NumMessages        int
	SessionPath        string
	MessageDir         string
	TotalInputTokens   int
	TotalOutputTokens  int
	TotalCost          float64
}

type sessionJSON struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Directory string `json:"directory"`
	Time      struct {
		Created int64 `json:"created"`
		Updated int64 `json:"updated"`
	} `json:"time"`
}

type messageJSON struct {
	Tokens struct {
		Input  int `json:"input"`
		Output int `json:"output"`
	} `json:"tokens"`
	Cost float64 `json:"cost"`
}

func OpenCodeDataHome(home string) string {
	if xdg := strings.TrimSpace(os.Getenv("XDG_DATA_HOME")); xdg != "" {
		return filepath.Join(xdg, "opencode")
	}
	return filepath.Join(home, ".local", "share", "opencode")
}

func List(dataDir string, limit int) ([]Session, error) {
	if limit <= 0 {
		limit = defaultListLimit
	}
	if limit > maxListLimit {
		limit = maxListLimit
	}

	sessions, err := discoverSessions(dataDir)
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

func Find(dataDir, sessionID string) (Session, error) {
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return Session{}, sessionNotFoundError(sessionID)
	}

	root := filepath.Join(dataDir, "storage", "session")
	var found Session
	foundOK := false

	err := filepath.WalkDir(root, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			if os.IsNotExist(walkErr) {
				return nil
			}
			return walkErr
		}
		if d.IsDir() || !strings.HasSuffix(d.Name(), ".json") {
			return nil
		}

		session, ok := parseSessionFile(path, dataDir)
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

func Info(dataDir, sessionID string) (*SessionInfo, error) {
	session, err := Find(dataDir, sessionID)
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

	var payload sessionJSON
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, fmt.Errorf("parse session for %s: %w", sessionID, err)
	}

	messageDir := filepath.Join(dataDir, "storage", "message", sessionID)
	numMessages := countMessageFiles(messageDir)
	inputTokens, outputTokens, totalCost := sumMessageTotals(messageDir)

	info := &SessionInfo{
		Session:           session,
		CreatedAt:         epochMSToTime(payload.Time.Created),
		UpdatedAt:         epochMSToTime(payload.Time.Updated),
		NumMessages:       numMessages,
		SessionPath:       session.Path,
		MessageDir:        messageDir,
		TotalInputTokens:  inputTokens,
		TotalOutputTokens: outputTokens,
		TotalCost:         totalCost,
	}
	info.NumMessages = numMessages

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
			session.NumMessages,
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
	fmt.Fprintf(&b, "Messages: %d\n", info.NumMessages)

	fmt.Fprintf(&b, "\nFiles:\n")
	fmt.Fprintf(&b, "  Session:   %s\n", shortenPath(info.SessionPath, home))
	fmt.Fprintf(&b, "  Messages:  %s\n", shortenPath(info.MessageDir, home))

	if hasTokenUsage(info) {
		fmt.Fprintf(&b, "\nTokens:\n")
		fmt.Fprintf(&b, "  Input:  %d\n", info.TotalInputTokens)
		fmt.Fprintf(&b, "  Output: %d\n", info.TotalOutputTokens)
		fmt.Fprintf(&b, "\nCost:\n")
		fmt.Fprintf(&b, "  Total: %g\n", info.TotalCost)
	}

	return strings.TrimRight(b.String(), "\n")
}

func discoverSessions(dataDir string) ([]Session, error) {
	root := filepath.Join(dataDir, "storage", "session")
	var sessions []Session

	err := filepath.WalkDir(root, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			if os.IsNotExist(walkErr) {
				return nil
			}
			return walkErr
		}
		if d.IsDir() || !strings.HasSuffix(d.Name(), ".json") {
			return nil
		}

		session, ok := parseSessionFile(path, dataDir)
		if !ok {
			return nil
		}
		sessions = append(sessions, session)
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

func parseSessionFile(path, dataDir string) (Session, bool) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Session{}, false
	}
	data = []byte(strings.TrimSpace(string(data)))
	if len(data) == 0 {
		return Session{}, false
	}

	var payload sessionJSON
	if err := json.Unmarshal(data, &payload); err != nil {
		return Session{}, false
	}

	id := strings.TrimSpace(payload.ID)
	if id == "" {
		return Session{}, false
	}

	updatedAt := epochMSToTime(payload.Time.Updated)
	if updatedAt.IsZero() {
		updatedAt = epochMSToTime(payload.Time.Created)
	}
	if updatedAt.IsZero() {
		return Session{}, false
	}

	messageDir := filepath.Join(dataDir, "storage", "message", id)
	numMessages := countMessageFiles(messageDir)

	return Session{
		ID:           id,
		LastActiveAt: updatedAt,
		CWD:          strings.TrimSpace(payload.Directory),
		Title:        strings.TrimSpace(payload.Title),
		Path:         path,
		NumMessages:  numMessages,
	}, true
}

func countMessageFiles(messageDir string) int {
	entries, err := os.ReadDir(messageDir)
	if err != nil {
		return 0
	}
	count := 0
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		count++
	}
	return count
}

func sumMessageTotals(messageDir string) (inputTokens, outputTokens int, totalCost float64) {
	entries, err := os.ReadDir(messageDir)
	if err != nil {
		return 0, 0, 0
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(messageDir, entry.Name()))
		if err != nil {
			continue
		}
		data = []byte(strings.TrimSpace(string(data)))
		if len(data) == 0 {
			continue
		}
		var msg messageJSON
		if err := json.Unmarshal(data, &msg); err != nil {
			continue
		}
		inputTokens += msg.Tokens.Input
		outputTokens += msg.Tokens.Output
		totalCost += msg.Cost
	}
	return inputTokens, outputTokens, totalCost
}

func epochMSToTime(ms int64) time.Time {
	if ms <= 0 {
		return time.Time{}
	}
	return time.UnixMilli(ms).UTC()
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

func hasTokenUsage(info *SessionInfo) bool {
	if info == nil {
		return false
	}
	return info.TotalInputTokens != 0 || info.TotalOutputTokens != 0
}

func sessionNotFoundError(id string) error {
	return fmt.Errorf("opencode session not found: %s", id)
}