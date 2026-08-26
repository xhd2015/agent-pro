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
	// Kind is the list display token: main | sub | sub+ | sub-f | fork.
	Kind string

	// Raw summary fields used by role/forked filters (not part of public table API).
	rawSessionKind  string
	parentSessionID string
	forkedAt        string
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
	SessionKind      string `json:"session_kind"`
	ParentSessionID  string `json:"parent_session_id"`
	ForkedAt         string `json:"forked_at"`
}

type grokSignals struct {
	ContextTokensUsed           int `json:"contextTokensUsed"`
	ContextWindowTokens         int `json:"contextWindowTokens"`
	ContextWindowUsage          int `json:"contextWindowUsage"`
	TotalTokensBeforeCompaction int `json:"totalTokensBeforeCompaction"`
}

// ListOptions configures ListWithOptions. Zero value matches List(grokHome, 0):
// default limit 20, no place/recent/active/role/forked/grep filters.
type ListOptions struct {
	Limit     int
	PlaceCWDs []string // resolved abs paths; OR match on Session.CWD; empty = no place filter
	Recent    time.Duration
	RecentSet bool
	Active    bool
	Now       time.Time // for recent window; zero → time.Now()
	Grep      []string // --grep (repeatable); AND on same field/line
	GrepSet   bool
	MainAgent bool // --main-agent; mutually exclusive with SubAgent
	SubAgent  bool // --sub-agent
	Forked    bool // --forked
}

func List(grokHome string, limit int) ([]Session, error) {
	return ListWithOptions(grokHome, ListOptions{Limit: limit})
}

// ListWithOptions discovers all sessions under grokHome, then applies filters
// in locked order: place → recent → active → role → forked → grep → sort → limit.
func ListWithOptions(grokHome string, opts ListOptions) ([]Session, error) {
	if opts.MainAgent && opts.SubAgent {
		return nil, fmt.Errorf("--main-agent and --sub-agent are mutually exclusive")
	}
	if opts.RecentSet && opts.Recent <= 0 {
		return nil, fmt.Errorf("invalid recent window: must be positive")
	}
	grepPatterns, err := validateGrepPatterns(opts.GrepSet, opts.Grep)
	if err != nil {
		return nil, err
	}

	limit := opts.Limit
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

	// 1. place: OR across PlaceCWDs; Abs+Clean equality; empty session CWD never matches
	if len(opts.PlaceCWDs) > 0 {
		placeSet := make(map[string]struct{}, len(opts.PlaceCWDs))
		for _, p := range opts.PlaceCWDs {
			cp, err := canonicalPath(p)
			if err != nil {
				return nil, fmt.Errorf("place cwd %q: %w", p, err)
			}
			placeSet[cp] = struct{}{}
		}
		var filtered []Session
		for _, s := range sessions {
			if strings.TrimSpace(s.CWD) == "" {
				continue
			}
			cp, err := canonicalPath(s.CWD)
			if err != nil {
				continue
			}
			if _, ok := placeSet[cp]; ok {
				filtered = append(filtered, s)
			}
		}
		sessions = filtered
	}

	// 2. recent: last_active >= Now - Recent (inclusive lower bound)
	if opts.RecentSet {
		now := opts.Now
		if now.IsZero() {
			now = time.Now()
		}
		cutoff := now.Add(-opts.Recent)
		var filtered []Session
		for _, s := range sessions {
			if !s.LastActiveAt.Before(cutoff) {
				filtered = append(filtered, s)
			}
		}
		sessions = filtered
	}

	// 3. active: IsFileActive / active_sessions.json membership
	if opts.Active {
		var filtered []Session
		for _, s := range sessions {
			ok, err := IsFileActive(grokHome, s.ID)
			if err != nil {
				return nil, err
			}
			if ok {
				filtered = append(filtered, s)
			}
		}
		sessions = filtered
	}

	// 4. role: main-agent class XOR sub-agent class
	if opts.MainAgent || opts.SubAgent {
		var filtered []Session
		for _, s := range sessions {
			isSub := isSubAgentClass(s)
			if opts.MainAgent && !isSub {
				filtered = append(filtered, s)
			} else if opts.SubAgent && isSub {
				filtered = append(filtered, s)
			}
		}
		sessions = filtered
	}

	// 5. forked: fork kinds or non-empty forked_at
	if opts.Forked {
		var filtered []Session
		for _, s := range sessions {
			if isForkedSession(s) {
				filtered = append(filtered, s)
			}
		}
		sessions = filtered
	}

	// 6. grep: presence filter (same search family as ListWithGrep; AND per unit)
	if opts.GrepSet {
		var filtered []Session
		for _, s := range sessions {
			if len(searchSession(s, grepPatterns)) > 0 {
				filtered = append(filtered, s)
			}
		}
		sessions = filtered
	}

	// 7. sort last_active desc (ID desc tie-break, same as List)
	sort.Slice(sessions, func(i, j int) bool {
		if sessions[i].LastActiveAt.Equal(sessions[j].LastActiveAt) {
			return sessions[i].ID > sessions[j].ID
		}
		return sessions[i].LastActiveAt.After(sessions[j].LastActiveAt)
	})

	// 8. limit
	if len(sessions) > limit {
		sessions = sessions[:limit]
	}
	return sessions, nil
}

// isSubAgentClass reports whether s is in the sub-agent filter class:
// session_kind ∈ {subagent, subagent_resume, subagent_fork}
// OR (kind empty/absent AND parent_session_id non-empty).
func isSubAgentClass(s Session) bool {
	k := strings.TrimSpace(s.rawSessionKind)
	switch k {
	case "subagent", "subagent_resume", "subagent_fork":
		return true
	}
	if k == "" && strings.TrimSpace(s.parentSessionID) != "" {
		return true
	}
	return false
}

// isForkedSession reports whether s is a forked session for --forked:
// session_kind ∈ {fork, subagent_fork} OR forked_at is non-empty non-whitespace.
func isForkedSession(s Session) bool {
	k := strings.TrimSpace(s.rawSessionKind)
	if k == "fork" || k == "subagent_fork" {
		return true
	}
	return strings.TrimSpace(s.forkedAt) != ""
}

// displayKindToken maps raw session_kind to list KIND display tokens.
// Priority: subagent_fork → fork → subagent_resume → subagent → main.
func displayKindToken(rawKind string) string {
	switch strings.TrimSpace(rawKind) {
	case "subagent_fork":
		return "sub-f"
	case "fork":
		return "fork"
	case "subagent_resume":
		return "sub+"
	case "subagent":
		return "sub"
	default:
		return "main"
	}
}

// sessionKindOrMain returns Kind for table rendering; empty Kind defaults to main.
func sessionKindOrMain(s Session) string {
	if k := strings.TrimSpace(s.Kind); k != "" {
		return k
	}
	return "main"
}

// canonicalPath returns filepath.Clean(filepath.Abs(p)) for place comparisons.
func canonicalPath(p string) (string, error) {
	abs, err := filepath.Abs(p)
	if err != nil {
		return "", err
	}
	return filepath.Clean(abs), nil
}

// Find locates a session under grokHome/sessions/<place>/<sessionID>/summary.json.
// It matches the session directory basename to sessionID, parses only that
// summary, and skips descending into other session dirs (avoids walking large
// per-session payloads like updates.jsonl).
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
		if !d.IsDir() {
			return nil
		}
		// Never walk the sessions root as a "session"; only its children.
		if path == root {
			return nil
		}

		summaryPath := filepath.Join(path, "summary.json")
		if d.Name() == sessionID {
			session, ok := parseSummaryFile(summaryPath)
			if ok && session.ID == sessionID {
				found = session
				foundOK = true
				return filepath.SkipAll
			}
			// Dir name matched but summary missing/mismatched: do not descend.
			return filepath.SkipDir
		}

		// Other session leaf (has summary.json): skip its payload files.
		if _, statErr := os.Lstat(summaryPath); statErr == nil {
			return filepath.SkipDir
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
	fmt.Fprintf(&b, "%-38s  %-5s  %-12s  %-42s  %5s  %s\n", "SESSION ID", "KIND", "LAST ACTIVE", "TITLE", "MSGS", "CWD")
	for _, session := range sessions {
		fmt.Fprintf(
			&b,
			"%-38s  %-5s  %-12s  %-42s  %5d  %s\n",
			session.ID,
			sessionKindOrMain(session),
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

	rawKind := strings.TrimSpace(summary.SessionKind)
	parentID := strings.TrimSpace(summary.ParentSessionID)
	forkedAt := summary.ForkedAt // keep raw; filters TrimSpace themselves

	return Session{
		ID:              id,
		LastActiveAt:    lastActive,
		CWD:             strings.TrimSpace(summary.Info.CWD),
		Title:           title,
		Path:            path,
		NumChatMessages: summary.NumChatMessages,
		Kind:            displayKindToken(rawKind),
		rawSessionKind:  rawKind,
		parentSessionID: parentID,
		forkedAt:        forkedAt,
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