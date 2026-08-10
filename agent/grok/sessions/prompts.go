package sessions

import (
	"fmt"
	"io"
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
	missingTimestampMarker  = "[—]" // em dash U+2014
	promptTruncateEllipsis  = "…"   // U+2026
	sessionBlockSeparator   = "────────────────────────────────────────"
	ansiReset               = "\x1b[0m"
	ansiDim                 = "\x1b[2m"
	ansiBold                = "\x1b[1m"
	ansiRed                 = "\x1b[31m"
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
	UserPrompts   []UserPrompt
	OmittedBefore int // tail clip; 0 if none
	OmittedAfter  int // head clip; 0 if none
}

// ListPromptsOptions controls multi-session selection, window filtering, and
// the per-session text filter pipeline (grep → exclude → head|tail).
// Now is required when RecentSet is true (relative window).
// Zero-value filter fields preserve pre-filter behavior.
type ListPromptsOptions struct {
	Now       time.Time
	Recent    time.Duration // 0 = no time window
	RecentSet bool          // true if --recent was provided
	Limit     int           // session cap; 0 + !LimitSet → default 10 when !RecentSet
	LimitSet  bool
	Home      string // optional path shorten for formatters

	Grep       string
	GrepSet    bool
	Exclude    string
	ExcludeSet bool
	Head       int // N >= 1 when HeadSet
	HeadSet    bool
	Tail       int // N >= 1 when TailSet
	TailSet    bool
}

// FilterUserPromptsOptions is the pure in-memory filter pipeline for one
// session's prompt slice (no FS). Zero-value = identity (keep all).
type FilterUserPromptsOptions struct {
	Grep       string
	GrepSet    bool
	Exclude    string
	ExcludeSet bool
	Head       int
	HeadSet    bool
	Tail       int
	TailSet    bool
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
	// ColorMode is "auto" | "always" | "never"; empty treated as "never" for
	// deterministic Format* string helpers (CLI passes "auto" by default).
	ColorMode string
	// Grep / GrepSet: when set, bold-red highlights the first match when color
	// is on. Without MaxBodySet the full collapsed body is kept; with MaxBodySet
	// the body is windowed around the match within MaxBodyRunes.
	Grep    string
	GrepSet bool
	// MaxBodyRunes soft-caps each collapsed body to N content runes + "…"
	// (ellipsis outside the N budget) when MaxBodySet is true. N must be >= 1.
	MaxBodyRunes int
	MaxBodySet   bool // true if --max-body provided
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
// Does not apply text filters; callers use FilterUserPrompts when needed.
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

// FilterUserPrompts applies the pure filter pipeline on an in-memory prompt
// slice: grep keep → exclude drop → head|tail slice.
// Returns kept prompts and omission counts for formatter chrome.
// Invalid opts (head+tail, empty pattern when set, N < 1) → clear error.
func FilterUserPrompts(prompts []UserPrompt, opts FilterUserPromptsOptions) (kept []UserPrompt, omittedBefore, omittedAfter int, err error) {
	if err := validateFilterUserPromptsOptions(opts); err != nil {
		return nil, 0, 0, err
	}

	kept = prompts
	if opts.GrepSet {
		var filtered []UserPrompt
		for _, p := range kept {
			if _, _, ok := findLiteralCI(p.Text, opts.Grep); ok {
				filtered = append(filtered, p)
			}
		}
		kept = filtered
	}
	if opts.ExcludeSet {
		var filtered []UserPrompt
		for _, p := range kept {
			if _, _, ok := findLiteralCI(p.Text, opts.Exclude); ok {
				continue
			}
			filtered = append(filtered, p)
		}
		kept = filtered
	}
	if opts.HeadSet {
		if len(kept) > opts.Head {
			omittedAfter = len(kept) - opts.Head
			// copy slice head so caller cannot mutate underlying array unexpectedly
			kept = append([]UserPrompt(nil), kept[:opts.Head]...)
		}
	} else if opts.TailSet {
		if len(kept) > opts.Tail {
			omittedBefore = len(kept) - opts.Tail
			kept = append([]UserPrompt(nil), kept[len(kept)-opts.Tail:]...)
		}
	}
	if kept == nil {
		kept = []UserPrompt{}
	}
	return kept, omittedBefore, omittedAfter, nil
}

func validateFilterUserPromptsOptions(opts FilterUserPromptsOptions) error {
	if opts.HeadSet && opts.TailSet {
		return fmt.Errorf("--head and --tail are mutually exclusive")
	}
	if opts.HeadSet && opts.Head < 1 {
		return fmt.Errorf("--head must be >= 1 (got %d)", opts.Head)
	}
	if opts.TailSet && opts.Tail < 1 {
		return fmt.Errorf("--tail must be >= 1 (got %d)", opts.Tail)
	}
	if opts.GrepSet && opts.Grep == "" {
		return fmt.Errorf("--grep pattern must not be empty")
	}
	if opts.ExcludeSet && opts.Exclude == "" {
		return fmt.Errorf("--exclude pattern must not be empty")
	}
	return nil
}

func filterOptsFromList(opts ListPromptsOptions) FilterUserPromptsOptions {
	return FilterUserPromptsOptions{
		Grep:       opts.Grep,
		GrepSet:    opts.GrepSet,
		Exclude:    opts.Exclude,
		ExcludeSet: opts.ExcludeSet,
		Head:       opts.Head,
		HeadSet:    opts.HeadSet,
		Tail:       opts.Tail,
		TailSet:    opts.TailSet,
	}
}

// ListPrompts discovers sessions newest-first by last_active_at and applies
// the RecentSet × LimitSet selection matrix.
//
// Sessions with zero (in-window when RecentSet) user prompts are always
// skipped and do not count toward the limit — so --limit N means N sessions
// that actually have prompts to show.
//
// For progressive CLI output, prefer StreamPromptsList (load+print per session).
func ListPrompts(grokHome string, opts ListPromptsOptions) ([]SessionPrompts, error) {
	var out []SessionPrompts
	err := forEachPromptSession(grokHome, opts, func(sp SessionPrompts) error {
		out = append(out, sp)
		return nil
	})
	return out, err
}

// StreamPromptsList walks sessions newest-first, loads each session's user
// prompts, and writes that block to w immediately (separator + header + lines).
// Footer is written at the end. Partial stdout may exist if a later session fails.
//
// This is the progressive path: discovery of summary.json is one pass; heavy
// updates.jsonl reads happen only for candidates and each ready session is
// flushed before the next load.
func StreamPromptsList(w io.Writer, grokHome string, opts ListPromptsOptions, fmtOpts FormatPromptsOptions) error {
	if err := validateFormatPromptsOptions(fmtOpts); err != nil {
		return err
	}
	if fmtOpts.Now.IsZero() {
		fmtOpts.Now = opts.Now
	}
	if fmtOpts.Now.IsZero() {
		fmtOpts.Now = time.Now()
	}
	if fmtOpts.Home == "" {
		fmtOpts.Home = opts.Home
	}
	if !fmtOpts.RecentSet {
		fmtOpts.RecentSet = opts.RecentSet
	}
	if fmtOpts.Window == 0 && opts.RecentSet {
		fmtOpts.Window = opts.Recent
	}
	if !fmtOpts.LimitSet {
		fmtOpts.LimitSet = opts.LimitSet
		fmtOpts.Limit = opts.Limit
	}

	loc := fmtOpts.Location
	if loc == nil {
		loc = time.Local
	}
	useColor := shouldColor(normalizePromptsColorMode(fmtOpts.ColorMode))
	now := fmtOpts.Now

	nSess := 0
	totalMsgs := 0
	var wroteAny bool

	err := forEachPromptSession(grokHome, opts, func(sp SessionPrompts) error {
		if nSess > 0 {
			if _, err := io.WriteString(w, "\n"); err != nil {
				return err
			}
			rule := sessionBlockSeparator
			if useColor {
				rule = dimMeta(rule, true)
			}
			if _, err := fmt.Fprintln(w, rule); err != nil {
				return err
			}
			if _, err := io.WriteString(w, "\n"); err != nil {
				return err
			}
		}

		title := strings.TrimSpace(sp.Title)
		if title == "" {
			title = "(untitled)"
		}
		header := fmt.Sprintf(
			"── %s  ·  %s  ·  %s  ·  %s",
			sp.ID,
			formatRelativeTime(sp.LastActiveAt, now),
			title,
			shortenPath(sp.CWD, fmtOpts.Home),
		)
		if useColor {
			header = dimMeta(header, true)
		}
		if _, err := fmt.Fprintln(w, header); err != nil {
			return err
		}
		if err := writeSessionPromptBody(w, &sp, loc, useColor, fmtOpts); err != nil {
			return err
		}
		// Force the session block to the terminal before loading the next file.
		if err := flushWriter(w); err != nil {
			return err
		}

		nSess++
		totalMsgs += len(sp.UserPrompts)
		wroteAny = true
		return nil
	})
	if err != nil {
		return err
	}
	if !wroteAny {
		_, err := io.WriteString(w, "No user prompts found\n")
		_ = flushWriter(w)
		return err
	}

	footer := fmt.Sprintf("%d sessions, %d user messages", nSess, totalMsgs)
	if fmtOpts.RecentSet && fmtOpts.Window > 0 {
		footer += fmt.Sprintf(" (recent %s)", formatWindowShort(fmtOpts.Window))
	}
	if fmtOpts.LimitSet && fmtOpts.Limit > 0 {
		footer += fmt.Sprintf(" (limit %d)", fmtOpts.Limit)
	}
	if useColor {
		footer = dimMeta(footer, true)
	}
	if _, err := fmt.Fprintln(w, footer); err != nil {
		return err
	}
	return flushWriter(w)
}

// forEachPromptSession discovers/sorts sessions, loads user prompts per session,
// applies recent window then grep/exclude/head|tail, skips empties, applies the
// limit matrix (survivors only), and invokes fn for each kept session in order
// (before the next load when the caller streams).
func forEachPromptSession(grokHome string, opts ListPromptsOptions, fn func(SessionPrompts) error) error {
	filterOpts := filterOptsFromList(opts)
	// Validate filter opts up front so empty homes still reject invalid flags.
	if err := validateFilterUserPromptsOptions(filterOpts); err != nil {
		return err
	}

	sessions, err := discoverSessions(grokHome)
	if err != nil {
		return err
	}
	sort.Slice(sessions, func(i, j int) bool {
		if sessions[i].LastActiveAt.Equal(sessions[j].LastActiveAt) {
			return sessions[i].ID > sessions[j].ID
		}
		return sessions[i].LastActiveAt.After(sessions[j].LastActiveAt)
	})

	sessionCap, hasCap := listPromptsSessionCap(opts)
	kept := 0
	for _, s := range sessions {
		prompts, err := loadUserPrompts(s)
		if err != nil {
			return err
		}
		if opts.RecentSet {
			prompts = filterPromptsInWindow(prompts, opts.Now, opts.Recent)
		}
		prompts, omittedBefore, omittedAfter, err := FilterUserPrompts(prompts, filterOpts)
		if err != nil {
			return err
		}
		if len(prompts) == 0 {
			continue
		}
		if err := fn(SessionPrompts{
			Session:       s,
			UserPrompts:   prompts,
			OmittedBefore: omittedBefore,
			OmittedAfter:  omittedAfter,
		}); err != nil {
			return err
		}
		kept++
		if hasCap && kept >= sessionCap {
			break
		}
	}
	return nil
}

func flushWriter(w io.Writer) error {
	type flusher interface{ Flush() error }
	if f, ok := w.(flusher); ok {
		return f.Flush()
	}
	// *os.File has no Flush; Sync is too heavy. Prefer bufio from CLI.
	return nil
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
	var b strings.Builder
	_ = WritePromptsText(&b, sp, opts)
	return b.String()
}

// WritePromptsText streams compact prompt lines for one session to w.
// When OmittedBefore/OmittedAfter > 0, prints virtual omission markers
// (not counted as user messages).
func WritePromptsText(w io.Writer, sp *SessionPrompts, opts FormatPromptsOptions) error {
	if err := validateFormatPromptsOptions(opts); err != nil {
		return err
	}
	if sp == nil || len(sp.UserPrompts) == 0 {
		_, err := io.WriteString(w, "No user prompts found\n")
		return err
	}
	loc := opts.Location
	if loc == nil {
		loc = time.Local
	}
	useColor := shouldColor(normalizePromptsColorMode(opts.ColorMode))
	return writeSessionPromptBody(w, sp, loc, useColor, opts)
}

// validateFormatPromptsOptions rejects invalid MaxBody when MaxBodySet.
func validateFormatPromptsOptions(opts FormatPromptsOptions) error {
	if opts.MaxBodySet && opts.MaxBodyRunes < 1 {
		return fmt.Errorf("--max-body must be >= 1 (got %d)", opts.MaxBodyRunes)
	}
	return nil
}

// writeSessionPromptBody writes omission markers + prompt lines for one session.
func writeSessionPromptBody(w io.Writer, sp *SessionPrompts, loc *time.Location, useColor bool, opts FormatPromptsOptions) error {
	if sp.OmittedBefore > 0 {
		if _, err := fmt.Fprintln(w, formatOmissionMarker(sp.OmittedBefore, useColor)); err != nil {
			return err
		}
	}
	for _, p := range sp.UserPrompts {
		if _, err := fmt.Fprintln(w, formatPromptLine(p, loc, useColor, opts)); err != nil {
			return err
		}
	}
	if sp.OmittedAfter > 0 {
		if _, err := fmt.Fprintln(w, formatOmissionMarker(sp.OmittedAfter, useColor)); err != nil {
			return err
		}
	}
	return nil
}

func formatOmissionMarker(m int, useColor bool) string {
	s := fmt.Sprintf("(...%d omitted...)", m)
	if useColor {
		return dimMeta(s, true)
	}
	return s
}

// FormatPromptsListText renders multi-session compact output with headers,
// blank-line + rule separators between sessions, and a footer.
// Empty list → "No user prompts found\n". Always ends with trailing newline.
func FormatPromptsListText(list []SessionPrompts, opts FormatPromptsOptions) string {
	var b strings.Builder
	_ = WritePromptsList(&b, list, opts)
	return b.String()
}

// WritePromptsList streams multi-session compact output to w (session-by-session).
// Prefer this from CLI for progressive output; FormatPromptsListText wraps it.
func WritePromptsList(w io.Writer, list []SessionPrompts, opts FormatPromptsOptions) error {
	if err := validateFormatPromptsOptions(opts); err != nil {
		return err
	}
	// Filter to sessions with prompts for counting / emission.
	var with []SessionPrompts
	totalMsgs := 0
	for i := range list {
		if len(list[i].UserPrompts) == 0 {
			continue
		}
		with = append(with, list[i])
		totalMsgs += len(list[i].UserPrompts)
	}
	if len(with) == 0 {
		_, err := io.WriteString(w, "No user prompts found\n")
		return err
	}

	loc := opts.Location
	if loc == nil {
		loc = time.Local
	}
	now := opts.Now
	if now.IsZero() {
		now = time.Now()
	}
	useColor := shouldColor(normalizePromptsColorMode(opts.ColorMode))

	for i := range with {
		if i > 0 {
			// Separator between sessions: blank line + rule + blank line feel
			if _, err := io.WriteString(w, "\n"); err != nil {
				return err
			}
			rule := sessionBlockSeparator
			if useColor {
				rule = dimMeta(rule, true)
			}
			if _, err := fmt.Fprintln(w, rule); err != nil {
				return err
			}
			if _, err := io.WriteString(w, "\n"); err != nil {
				return err
			}
		}

		sp := &with[i]
		title := strings.TrimSpace(sp.Title)
		if title == "" {
			title = "(untitled)"
		}
		cwd := shortenPath(sp.CWD, opts.Home)
		header := fmt.Sprintf(
			"── %s  ·  %s  ·  %s  ·  %s",
			sp.ID,
			formatRelativeTime(sp.LastActiveAt, now),
			title,
			cwd,
		)
		if useColor {
			header = dimMeta(header, true)
		}
		if _, err := fmt.Fprintln(w, header); err != nil {
			return err
		}
		if err := writeSessionPromptBody(w, sp, loc, useColor, opts); err != nil {
			return err
		}
		// Flush if bufio so each session appears promptly when streaming.
		if f, ok := w.(interface{ Flush() error }); ok {
			_ = f.Flush()
		}
	}

	footer := fmt.Sprintf("%d sessions, %d user messages", len(with), totalMsgs)
	if opts.RecentSet && opts.Window > 0 {
		footer += fmt.Sprintf(" (recent %s)", formatWindowShort(opts.Window))
	}
	if opts.LimitSet && opts.Limit > 0 {
		footer += fmt.Sprintf(" (limit %d)", opts.Limit)
	}
	if useColor {
		footer = dimMeta(footer, true)
	}
	if _, err := fmt.Fprintln(w, footer); err != nil {
		return err
	}
	if f, ok := w.(interface{ Flush() error }); ok {
		_ = f.Flush()
	}
	return nil
}

func normalizePromptsColorMode(mode string) string {
	mode = strings.TrimSpace(strings.ToLower(mode))
	if mode == "" {
		// String helpers used by tests default to never (no ANSI in asserts).
		return "never"
	}
	return mode
}

func dimMeta(s string, on bool) string {
	if !on || s == "" {
		return s
	}
	// Prefer dim; gray as fallback styling family (SGR 2 is widely supported).
	return ansiDim + s + ansiReset
}

func formatPromptLine(p UserPrompt, loc *time.Location, useColor bool, opts FormatPromptsOptions) string {
	prefix := missingTimestampMarker
	if !p.Timestamp.IsZero() {
		prefix = "[" + p.Timestamp.In(loc).Format("2006-01-02 15:04:05") + "]"
	}
	if useColor {
		prefix = dimMeta(prefix, true)
	}
	body := formatPromptBody(p.Text, opts, useColor)
	return prefix + " " + body
}

// formatPromptBody collapses whitespace. Default (!MaxBodySet): full body.
// MaxBodySet: soft-truncate to MaxBodyRunes + "…". With GrepSet: highlight
// first match; window around match only when MaxBodySet (budget = MaxBodyRunes).
func formatPromptBody(text string, opts FormatPromptsOptions, useColor bool) string {
	collapsed := collapseWhitespace(text)

	if opts.GrepSet && opts.Grep != "" {
		start, length, ok := findLiteralCI(collapsed, opts.Grep)
		if ok {
			if opts.MaxBodySet {
				snippet, newStart, newLen := windowPromptBody(collapsed, start, length, opts.MaxBodyRunes)
				return highlightMatchSpan(snippet, newStart, newLen, useColor)
			}
			// Full body + optional highlight (no window / soft-cap).
			return highlightMatchSpan(collapsed, start, length, useColor)
		}
		// No match: still apply MaxBody soft-cap if set.
		if opts.MaxBodySet {
			return softTruncateRunes(collapsed, opts.MaxBodyRunes)
		}
		return collapsed
	}

	if opts.MaxBodySet {
		return softTruncateRunes(collapsed, opts.MaxBodyRunes)
	}
	return collapsed
}

// highlightMatchSpan wraps [start, start+length) bytes of s in bold-red when
// useColor is true. start/length are byte offsets into s.
func highlightMatchSpan(s string, start, length int, useColor bool) string {
	if !useColor || length <= 0 || start < 0 || start+length > len(s) {
		return s
	}
	return s[:start] + ansiBold + ansiRed + s[start:start+length] + ansiReset + s[start+length:]
}

// windowPromptBody builds a ≤maxRunes window around a match in collapsed text.
// Match offsets are byte positions into the returned snippet.
// Cut sides get a single Unicode ellipsis (…); same family as soft-truncate.
func windowPromptBody(collapsed string, matchStart, matchLen, maxRunes int) (snippet string, newStart, newLen int) {
	runes := []rune(collapsed)
	total := len(runes)
	if maxRunes <= 0 || total <= maxRunes {
		return collapsed, matchStart, matchLen
	}

	matchStartRune := utf8.RuneCountInString(collapsed[:matchStart])
	matchEndRune := utf8.RuneCountInString(collapsed[:matchStart+matchLen])
	matchRuneCount := matchEndRune - matchStartRune

	if matchRuneCount >= maxRunes {
		s := string(runes[matchStartRune : matchStartRune+maxRunes])
		return s, 0, len(s)
	}

	remaining := maxRunes - matchRuneCount
	beforeBudget := remaining / 2
	afterBudget := remaining - beforeBudget

	beforeAvail := matchStartRune
	afterAvail := total - matchEndRune

	takeBefore := beforeBudget
	takeAfter := afterBudget
	if takeBefore > beforeAvail {
		takeAfter += takeBefore - beforeAvail
		takeBefore = beforeAvail
	}
	if takeAfter > afterAvail {
		takeBefore += takeAfter - afterAvail
		takeAfter = afterAvail
		if takeBefore > beforeAvail {
			takeBefore = beforeAvail
		}
	}

	// Ellipsis is 1 rune (…); reserve from the cut side's budget.
	const ellipsisRunes = 1
	needBeforeEllipsis := takeBefore < beforeAvail
	needAfterEllipsis := takeAfter < afterAvail
	contentBefore := takeBefore
	contentAfter := takeAfter
	if needBeforeEllipsis {
		contentBefore = takeBefore - ellipsisRunes
		if contentBefore < 0 {
			contentBefore = 0
		}
	}
	if needAfterEllipsis {
		contentAfter = takeAfter - ellipsisRunes
		if contentAfter < 0 {
			contentAfter = 0
		}
	}

	var b strings.Builder
	b.Grow(maxRunes * utf8.UTFMax)
	if needBeforeEllipsis {
		b.WriteString(promptTruncateEllipsis)
	}
	beforeStart := matchStartRune - contentBefore
	b.WriteString(string(runes[beforeStart:matchStartRune]))
	newStart = b.Len()
	matchStr := string(runes[matchStartRune:matchEndRune])
	b.WriteString(matchStr)
	newLen = len(matchStr)
	b.WriteString(string(runes[matchEndRune : matchEndRune+contentAfter]))
	if needAfterEllipsis {
		b.WriteString(promptTruncateEllipsis)
	}
	return b.String(), newStart, newLen
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
