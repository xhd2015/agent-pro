package sessions

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/xhd2015/agent-pro/agent/event/print"
	_ "github.com/xhd2015/agent-pro/agent/event/traceparse"
)

const (
	defaultListLimit = 20
	maxListLimit     = 100
)

type Session struct {
	ID               string    `json:"id"`
	StartedAt        time.Time `json:"started_at"`
	CWD              string    `json:"cwd"`
	Title            string    `json:"title,omitempty"`
	Path             string    `json:"path"`
	NumDisplayEvents int       `json:"num_display_events,omitempty"`
}

type SessionBrief struct {
	Session
	Status         string         `json:"status"`
	LineCount      int            `json:"line_count"`
	RecentMessages []DisplayEvent `json:"recent_messages"`
}

type SessionInfo struct {
	Session
	Status            string         `json:"status"`
	LineCount         int            `json:"line_count"`
	NumDisplayEvents  int            `json:"num_display_events"`
	RecentMessages    []DisplayEvent `json:"recent_messages"`
	TotalInputTokens  int            `json:"total_input_tokens"`
	TotalOutputTokens int            `json:"total_output_tokens"`
}

type DisplayEvent struct {
	Kind      string `json:"kind"`
	Text      string `json:"text"`
	Formatted string `json:"formatted"`
}

func CodexHomeFromEnv(home string) string {
	return filepath.Join(home, ".codex")
}

func List(codexHome string, limit int) ([]Session, error) {
	if limit <= 0 {
		limit = defaultListLimit
	}
	if limit > maxListLimit {
		limit = maxListLimit
	}

	sessions, err := discoverSessions(codexHome)
	if err != nil {
		return nil, err
	}

	sort.Slice(sessions, func(i, j int) bool {
		if sessions[i].StartedAt.Equal(sessions[j].StartedAt) {
			return sessions[i].ID > sessions[j].ID
		}
		return sessions[i].StartedAt.After(sessions[j].StartedAt)
	})

	if len(sessions) > limit {
		sessions = sessions[:limit]
	}

	for i := range sessions {
		if err := enrichSessionForList(&sessions[i]); err != nil {
			return nil, err
		}
	}
	return sessions, nil
}

func Find(codexHome string, sessionID string) (string, error) {
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return "", sessionNotFoundError(sessionID)
	}

	// Fast path: rollout files are named rollout-*-<sessionID>.jsonl under
	// sessions/YYYY/MM/DD/. Avoid discoverSessions (full tree open+parse).
	pattern := filepath.Join(codexHome, "sessions", "*", "*", "*", "rollout-*"+sessionID+".jsonl")
	if matches, gerr := filepath.Glob(pattern); gerr == nil && len(matches) > 0 {
		sort.Strings(matches)
		return matches[len(matches)-1], nil
	}

	sessions, err := discoverSessions(codexHome)
	if err != nil {
		return "", err
	}
	for _, session := range sessions {
		if session.ID == sessionID {
			return session.Path, nil
		}
	}
	return "", sessionNotFoundError(sessionID)
}

func Info(codexHome string, sessionID string, lastN int) (*SessionInfo, error) {
	if lastN <= 0 {
		lastN = 3
	}

	path, err := Find(codexHome, sessionID)
	if err != nil {
		return nil, err
	}

	lines, err := readLines(path)
	if err != nil {
		return nil, err
	}

	session, err := sessionFromFile(path, lines)
	if err != nil {
		return nil, err
	}

	events := collectDisplayEvents(lines)
	recent := events
	if len(recent) > lastN {
		recent = recent[len(recent)-lastN:]
	}
	inputTokens, outputTokens := sumTokenCounts(lines)

	return &SessionInfo{
		Session:           session,
		Status:            inferSessionStatus(lines),
		LineCount:         len(lines),
		NumDisplayEvents:  len(events),
		RecentMessages:    recent,
		TotalInputTokens:  inputTokens,
		TotalOutputTokens: outputTokens,
	}, nil
}

func Brief(codexHome string, sessionID string, lastN int) (*SessionBrief, error) {
	if lastN <= 0 {
		lastN = 3
	}

	path, err := Find(codexHome, sessionID)
	if err != nil {
		return nil, err
	}

	lines, err := readLines(path)
	if err != nil {
		return nil, err
	}

	session, err := sessionFromFile(path, lines)
	if err != nil {
		return nil, err
	}

	events := collectDisplayEvents(lines)
	if len(events) > lastN {
		events = events[len(events)-lastN:]
	}

	return &SessionBrief{
		Session:        session,
		Status:         inferSessionStatus(lines),
		LineCount:      len(lines),
		RecentMessages: events,
	}, nil
}

func PrintLog(path string, w io.Writer, tail int) error {
	lines, err := readLines(path)
	if err != nil {
		return err
	}
	events := collectDisplayEvents(lines)
	if tail > 0 && len(events) > tail {
		events = events[len(events)-tail:]
	}
	for _, event := range events {
		formatted := strings.TrimSpace(event.Formatted)
		if formatted == "" {
			continue
		}
		fmt.Fprintln(w, formatted)
	}
	return nil
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
			formatRelativeTime(session.StartedAt, now),
			truncateTitle(session.Title),
			session.NumDisplayEvents,
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
		formatRelativeTime(info.StartedAt, now),
		formatAbsoluteTime(info.StartedAt),
	)
	if info.CWD != "" {
		fmt.Fprintf(&b, "CWD: %s\n", info.CWD)
	}
	if info.Status != "" {
		fmt.Fprintf(&b, "Status: %s\n", info.Status)
	}
	fmt.Fprintf(&b, "Lines: %d\n", info.LineCount)
	if info.Path != "" {
		fmt.Fprintf(&b, "File: %s\n", shortenPath(info.Path, home))
	}
	if len(info.RecentMessages) > 0 {
		fmt.Fprintln(&b, "\nRecent messages:")
		for _, msg := range info.RecentMessages {
			fmt.Fprintln(&b, strings.TrimSpace(msg.Formatted))
		}
	}
	if info.TotalInputTokens != 0 || info.TotalOutputTokens != 0 {
		fmt.Fprintf(&b, "\nTokens:\n")
		fmt.Fprintf(&b, "  Input:  %d\n", info.TotalInputTokens)
		fmt.Fprintf(&b, "  Output: %d\n", info.TotalOutputTokens)
	}

	return strings.TrimRight(b.String(), "\n")
}

func FormatListJSON(sessions []Session) ([]byte, error) {
	type sessionJSON struct {
		ID        string `json:"id"`
		StartedAt string `json:"started_at"`
		CWD       string `json:"cwd"`
		Path      string `json:"path"`
	}
	out := struct {
		Sessions []sessionJSON `json:"sessions"`
	}{Sessions: make([]sessionJSON, 0, len(sessions))}

	for _, session := range sessions {
		out.Sessions = append(out.Sessions, sessionJSON{
			ID:        session.ID,
			StartedAt: session.StartedAt.UTC().Format(time.RFC3339Nano),
			CWD:       session.CWD,
			Path:      session.Path,
		})
	}
	return json.Marshal(out)
}

func FormatBriefText(brief *SessionBrief, home string) string {
	if brief == nil {
		return ""
	}

	var b strings.Builder
	fmt.Fprintf(&b, "Session: %s\n", brief.ID)
	fmt.Fprintf(&b, "Started: %s\n", brief.StartedAt.UTC().Format("2006-01-02 15:04:05"))
	fmt.Fprintf(&b, "CWD: %s\n", brief.CWD)
	if brief.Path != "" {
		fmt.Fprintf(&b, "File: %s\n", shortenPath(brief.Path, home))
	}
	if brief.Status != "" {
		fmt.Fprintf(&b, "Status: %s\n", brief.Status)
	}
	fmt.Fprintf(&b, "Lines: %d\n", brief.LineCount)
	if len(brief.RecentMessages) > 0 {
		fmt.Fprintln(&b, "\nRecent messages:")
		for _, msg := range brief.RecentMessages {
			fmt.Fprintln(&b, strings.TrimSpace(msg.Formatted))
		}
	}
	return strings.TrimRight(b.String(), "\n")
}

func FormatBriefJSON(brief *SessionBrief) ([]byte, error) {
	if brief == nil {
		return json.Marshal(struct{}{})
	}

	type briefJSON struct {
		ID             string         `json:"id"`
		StartedAt      string         `json:"started_at"`
		CWD            string         `json:"cwd"`
		Path           string         `json:"path"`
		Status         string         `json:"status,omitempty"`
		LineCount      int            `json:"line_count"`
		RecentMessages []DisplayEvent `json:"recent_messages"`
	}

	payload := briefJSON{
		ID:             brief.ID,
		StartedAt:      brief.StartedAt.UTC().Format(time.RFC3339Nano),
		CWD:            brief.CWD,
		Path:           brief.Path,
		Status:         brief.Status,
		LineCount:      brief.LineCount,
		RecentMessages: brief.RecentMessages,
	}
	if payload.RecentMessages == nil {
		payload.RecentMessages = []DisplayEvent{}
	}
	return json.Marshal(payload)
}

func formatTraceLine(traceLine string) string {
	return print.FormatTraceLine(traceLine)
}

type discoveredSession struct {
	Session
}

func discoverSessions(codexHome string) ([]Session, error) {
	root := filepath.Join(codexHome, "sessions")
	var sessions []Session

	err := filepath.WalkDir(root, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			if os.IsNotExist(walkErr) {
				return nil
			}
			return walkErr
		}
		if d.IsDir() || !strings.HasSuffix(d.Name(), ".jsonl") {
			return nil
		}

		session, err := sessionMetaFromFile(path)
		if err != nil {
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

func sessionMetaFromFile(path string) (Session, error) {
	info, err := os.Stat(path)
	if err != nil {
		return Session{}, err
	}

	file, err := os.Open(path)
	if err != nil {
		return Session{}, err
	}
	defer file.Close()

	session := Session{
		Path:      path,
		StartedAt: info.ModTime(),
	}

	scanner := newRolloutScanner(file)
	if scanner.Scan() {
		if meta, ok := parseSessionMeta(scanner.Text()); ok {
			session.ID = meta.ID
			session.CWD = meta.CWD
			if ts, err := parseTimestamp(meta.Timestamp); err == nil {
				session.StartedAt = ts
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return Session{}, err
	}

	if session.ID == "" {
		session.ID = uuidFromFilename(filepath.Base(path))
	}
	if session.ID == "" {
		return Session{}, fmt.Errorf("unable to determine session id for %s", path)
	}
	return session, nil
}

func enrichSessionForList(session *Session) error {
	file, err := os.Open(session.Path)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := newRolloutScanner(file)
	titleFound := session.Title != ""
	count := 0
	for scanner.Scan() {
		line := scanner.Text()
		if !titleFound {
			if title := titleFromRolloutLine(line); title != "" {
				session.Title = title
				titleFound = true
			}
		}
		count += countDisplayableFromRolloutLine(line)
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	session.NumDisplayEvents = count
	return nil
}

func sessionFromFile(path string, lines []string) (Session, error) {
	if lines == nil {
		var err error
		lines, err = readLines(path)
		if err != nil {
			return Session{}, err
		}
	}

	info, err := os.Stat(path)
	if err != nil {
		return Session{}, err
	}

	session := Session{
		Path:      path,
		StartedAt: info.ModTime(),
	}

	if len(lines) > 0 {
		if meta, ok := parseSessionMeta(lines[0]); ok {
			session.ID = meta.ID
			session.CWD = meta.CWD
			if ts, err := parseTimestamp(meta.Timestamp); err == nil {
				session.StartedAt = ts
			}
		}
	}

	if session.ID == "" {
		session.ID = uuidFromFilename(filepath.Base(path))
	}
	if session.ID == "" {
		return Session{}, fmt.Errorf("unable to determine session id for %s", path)
	}

	session.Title = extractTitle(lines)
	session.NumDisplayEvents = len(collectDisplayEvents(lines))
	return session, nil
}

// maxRolloutScanToken matches tip.go: Codex rollout lines can exceed the
// default bufio.Scanner 64KiB limit (base_instructions / large tool payloads).
const maxRolloutScanToken = 16 * 1024 * 1024

func newRolloutScanner(r io.Reader) *bufio.Scanner {
	sc := bufio.NewScanner(r)
	buf := make([]byte, 0, 1024*1024)
	sc.Buffer(buf, maxRolloutScanToken)
	return sc
}

func readLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := newRolloutScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return lines, nil
}

func collectDisplayEvents(lines []string) []DisplayEvent {
	var events []DisplayEvent
	for _, line := range lines {
		for _, traceLine := range rolloutToTraceLines(line) {
			if event, ok := displayEventFromTraceLine(traceLine); ok {
				events = append(events, event)
			}
		}
	}
	return events
}

func parseTimestamp(value string) (time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, fmt.Errorf("empty timestamp")
	}
	for _, layout := range []string{
		time.RFC3339Nano,
		"2006-01-02T15:04:05.000Z",
		time.RFC3339,
	} {
		if ts, err := time.Parse(layout, value); err == nil {
			return ts, nil
		}
	}
	return time.Time{}, fmt.Errorf("unsupported timestamp: %s", value)
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
	title = strings.Join(strings.Fields(strings.TrimSpace(title)), " ")
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
	if home != "" {
		if strings.HasPrefix(path, home) {
			rel := strings.TrimPrefix(path, home)
			rel = strings.TrimPrefix(rel, string(os.PathSeparator))
			if rel != "" {
				return "~/" + rel
			}
		}
	}
	return path
}