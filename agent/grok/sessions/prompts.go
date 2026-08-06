package sessions

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/xhd2015/agent-pro/agent/event/grok_session"
	types "github.com/xhd2015/agent-pro/agent/event/types"
)

const (
	defaultPromptsListLimit = 10
	promptBodyMaxRunes      = 200
	missingTimestampMarker  = "[—]" // em dash U+2014
	promptTruncateEllipsis  = "…"   // U+2026
)

// UserPrompt is one coalesced user message from a session's updates.jsonl.
type UserPrompt struct {
	Index     int       // 1-based chronological within session
	Timestamp time.Time // from wire; zero if unknown
	Text      string    // raw coalesced user text (before format collapse)
}

// SessionPrompts is a session plus its user prompts in chronological order.
type SessionPrompts struct {
	Session
	UserPrompts []UserPrompt
}

// ListPromptsOptions controls multi-session selection and window filtering.
// Now is required when RecentSet is true (relative window).
type ListPromptsOptions struct {
	Now       time.Time
	Recent    time.Duration // 0 = no time window
	RecentSet bool          // true if --recent was provided
	Limit     int           // session cap; 0 + !LimitSet → default 10 when !RecentSet
	LimitSet  bool
	Home      string // optional path shorten for formatters
}

// FormatPromptsOptions controls compact text rendering for CLI stdout.
type FormatPromptsOptions struct {
	Now       time.Time
	Home      string
	Location  *time.Location // nil → time.Local; tests pass time.UTC
	Window    time.Duration  // footer only
	Limit     int            // footer only
	RecentSet bool
	LimitSet  bool
}

var recentWindowRE = regexp.MustCompile(`(?i)^([0-9]+)([dhm])$`)

// ParseRecentWindow parses Nd|Nh|Nm (case-insensitive). 1d = 24h rolling.
// Rejects empty, bare numbers, 0*, and unsupported units (e.g. 2w).
func ParseRecentWindow(s string) (time.Duration, error) {
	s = strings.TrimSpace(s)
	m := recentWindowRE.FindStringSubmatch(s)
	if m == nil {
		return 0, fmt.Errorf("invalid recent window %q; use Nd, Nh, or Nm (e.g. 1d, 2h, 30m)", s)
	}
	n, err := strconv.Atoi(m[1])
	if err != nil || n <= 0 {
		return 0, fmt.Errorf("invalid recent window %q; use Nd, Nh, or Nm with a positive number", s)
	}
	switch strings.ToLower(m[2]) {
	case "d":
		return time.Duration(n) * 24 * time.Hour, nil
	case "h":
		return time.Duration(n) * time.Hour, nil
	case "m":
		return time.Duration(n) * time.Minute, nil
	default:
		return 0, fmt.Errorf("invalid recent window %q; use Nd, Nh, or Nm (e.g. 1d, 2h, 30m)", s)
	}
}

// Prompts returns all user prompts for one session (full history).
// Unknown / empty id → error containing "grok session not found".
// Missing updates.jsonl → empty UserPrompts, no error.
func Prompts(grokHome, sessionID string) (*SessionPrompts, error) {
	session, err := Find(grokHome, sessionID)
	if err != nil {
		return nil, err
	}
	prompts, err := loadUserPrompts(session)
	if err != nil {
		return nil, err
	}
	return &SessionPrompts{Session: session, UserPrompts: prompts}, nil
}

// ListPrompts discovers sessions newest-first by last_active_at and applies
// the RecentSet × LimitSet selection matrix. When RecentSet, only in-window
// user prompts are kept and sessions with zero in-window prompts are skipped
// (do not count toward limit).
func ListPrompts(grokHome string, opts ListPromptsOptions) ([]SessionPrompts, error) {
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

	sessionCap, hasCap := listPromptsSessionCap(opts)

	var out []SessionPrompts
	for _, s := range sessions {
		prompts, err := loadUserPrompts(s)
		if err != nil {
			return nil, err
		}
		if opts.RecentSet {
			prompts = filterPromptsInWindow(prompts, opts.Now, opts.Recent)
			if len(prompts) == 0 {
				continue
			}
		}
		out = append(out, SessionPrompts{Session: s, UserPrompts: prompts})
		if hasCap && len(out) >= sessionCap {
			break
		}
	}
	return out, nil
}

// listPromptsSessionCap returns (limit, hasCap) per the selection matrix:
//
//	!RecentSet && !LimitSet → default 10
//	!RecentSet && LimitSet  → Limit
//	RecentSet  && !LimitSet → no default cap
//	RecentSet  && LimitSet  → Limit
func listPromptsSessionCap(opts ListPromptsOptions) (int, bool) {
	if opts.RecentSet {
		if opts.LimitSet {
			return opts.Limit, true
		}
		return 0, false
	}
	if opts.LimitSet {
		return opts.Limit, true
	}
	return defaultPromptsListLimit, true
}

func filterPromptsInWindow(prompts []UserPrompt, now time.Time, window time.Duration) []UserPrompt {
	if window <= 0 {
		// Zero window with RecentSet: only prompts exactly at Now (inclusive ends).
		// Treat as [now-window, now] with window=0 → only exact Now.
	}
	start := now.Add(-window)
	var out []UserPrompt
	for _, p := range prompts {
		if p.Timestamp.IsZero() {
			continue
		}
		// Inclusive ends: Timestamp ∈ [Now-Recent, Now]
		if p.Timestamp.Before(start) || p.Timestamp.After(now) {
			continue
		}
		out = append(out, p)
	}
	return out
}

func loadUserPrompts(session Session) ([]UserPrompt, error) {
	updatesPath := filepath.Join(filepath.Dir(session.Path), "updates.jsonl")
	data, err := os.ReadFile(updatesPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	if len(data) == 0 {
		return nil, nil
	}
	// Split preserving last empty segment handling via ProcessLine skip.
	raw := string(data)
	lines := strings.Split(raw, "\n")
	events := grok_session.FromUpdatesJSONL(lines)

	var out []UserPrompt
	idx := 0
	for _, ev := range events {
		if ev.Type != types.ActionMessage || ev.Role != "user" {
			continue
		}
		if ev.Text == "" {
			continue
		}
		idx++
		var ts time.Time
		if ev.Timestamp > 0 {
			ts = time.UnixMilli(ev.Timestamp).UTC()
		}
		out = append(out, UserPrompt{
			Index:     idx,
			Timestamp: ts,
			Text:      ev.Text,
		})
	}
	return out, nil
}

// FormatPromptsText renders compact prompt lines for one session.
// Empty → "No user prompts found\n". Always ends with trailing newline.
func FormatPromptsText(sp *SessionPrompts, opts FormatPromptsOptions) string {
	if sp == nil || len(sp.UserPrompts) == 0 {
		return "No user prompts found\n"
	}
	loc := opts.Location
	if loc == nil {
		loc = time.Local
	}
	var b strings.Builder
	for _, p := range sp.UserPrompts {
		b.WriteString(formatPromptLine(p, loc))
		b.WriteByte('\n')
	}
	return b.String()
}

// FormatPromptsListText renders multi-session compact output with headers.
// Empty list → "No user prompts found\n". Always ends with trailing newline.
func FormatPromptsListText(list []SessionPrompts, opts FormatPromptsOptions) string {
	if len(list) == 0 {
		return "No user prompts found\n"
	}
	// Count messages; if all empty, still friendly empty.
	totalMsgs := 0
	for i := range list {
		totalMsgs += len(list[i].UserPrompts)
	}
	if totalMsgs == 0 {
		return "No user prompts found\n"
	}

	loc := opts.Location
	if loc == nil {
		loc = time.Local
	}
	now := opts.Now
	if now.IsZero() {
		now = time.Now()
	}

	var b strings.Builder
	for i := range list {
		sp := &list[i]
		if len(sp.UserPrompts) == 0 {
			continue
		}
		title := strings.TrimSpace(sp.Title)
		if title == "" {
			title = "(untitled)"
		}
		cwd := shortenPath(sp.CWD, opts.Home)
		fmt.Fprintf(
			&b,
			"── %s  ·  %s  ·  %s  ·  %s\n",
			sp.ID,
			formatRelativeTime(sp.LastActiveAt, now),
			title,
			cwd,
		)
		for _, p := range sp.UserPrompts {
			b.WriteString(formatPromptLine(p, loc))
			b.WriteByte('\n')
		}
	}

	// Optional footer
	nSess := 0
	for i := range list {
		if len(list[i].UserPrompts) > 0 {
			nSess++
		}
	}
	fmt.Fprintf(&b, "%d sessions, %d user messages", nSess, totalMsgs)
	if opts.RecentSet && opts.Window > 0 {
		fmt.Fprintf(&b, " (recent %s)", formatWindowShort(opts.Window))
	}
	if opts.LimitSet && opts.Limit > 0 {
		fmt.Fprintf(&b, " (limit %d)", opts.Limit)
	}
	b.WriteByte('\n')
	return b.String()
}

func formatPromptLine(p UserPrompt, loc *time.Location) string {
	prefix := missingTimestampMarker
	if !p.Timestamp.IsZero() {
		prefix = "[" + p.Timestamp.In(loc).Format("2006-01-02 15:04:05") + "]"
	}
	body := softTruncateRunes(collapseWhitespace(p.Text), promptBodyMaxRunes)
	return prefix + " " + body
}

func softTruncateRunes(s string, max int) string {
	if max <= 0 {
		return ""
	}
	if utf8.RuneCountInString(s) <= max {
		return s
	}
	runes := []rune(s)
	return string(runes[:max]) + promptTruncateEllipsis
}

func formatWindowShort(d time.Duration) string {
	if d <= 0 {
		return d.String()
	}
	if d%(24*time.Hour) == 0 {
		return fmt.Sprintf("%dd", int(d/(24*time.Hour)))
	}
	if d%time.Hour == 0 {
		return fmt.Sprintf("%dh", int(d/time.Hour))
	}
	if d%time.Minute == 0 {
		return fmt.Sprintf("%dm", int(d/time.Minute))
	}
	return d.String()
}
