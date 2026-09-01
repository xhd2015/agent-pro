package sessions

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"golang.org/x/term"
)

const (
	defaultPromptsListLimit = 10
	missingTimestampMarker  = "[—]" // em dash U+2014
	promptTruncateEllipsis  = "…"   // U+2026
	sessionBlockSeparator   = "────────────────────────────────────────"
	ansiReset               = "\x1b[0m"
	ansiDim                 = "\x1b[2m"
	ansiBoldRed             = "\x1b[1m\x1b[31m"
)

// UserPrompt is one coalesced user message from a Codex rollout JSONL.
type UserPrompt struct {
	Index     int       // 1-based chronological within session
	Timestamp time.Time // from wire; zero if unknown
	Text      string
}

// SessionPrompts is a session plus its user prompts in chronological order.
type SessionPrompts struct {
	Session
	UserPrompts   []UserPrompt
	OmittedBefore int
	OmittedAfter  int
}

// ListPromptsOptions controls multi-session selection and the per-session
// filter pipeline (grep → exclude → head|tail).
type ListPromptsOptions struct {
	Now       time.Time
	Recent    time.Duration
	RecentSet bool
	Limit     int
	LimitSet  bool
	Home      string

	OnlySessionIDs    []string
	OnlySessionIDsSet bool

	Grep       []string
	GrepSet    bool
	Exclude    string
	ExcludeSet bool
	Head       int
	HeadSet    bool
	Tail       int
	TailSet    bool
}

// FilterUserPromptsOptions is the pure in-memory filter pipeline.
type FilterUserPromptsOptions struct {
	Grep       []string
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
	Now          time.Time
	Home         string
	Location     *time.Location
	Window       time.Duration
	Limit        int
	RecentSet    bool
	LimitSet     bool
	ColorMode    string
	Grep         []string
	GrepSet      bool
	MaxBodyRunes int
	MaxBodySet   bool
}

var recentWindowRE = regexp.MustCompile(`(?i)^([0-9]+)([dhm])$`)

// ParseRecentWindow parses Nd|Nh|Nm (case-insensitive). 1d = 24h rolling.
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
// Does not apply text filters; callers use FilterUserPrompts when needed.
func Prompts(codexHome, sessionID string) (*SessionPrompts, error) {
	path, err := Find(codexHome, sessionID)
	if err != nil {
		return nil, err
	}
	session, err := sessionMetaFromFile(path)
	if err != nil {
		return nil, err
	}
	prompts, err := loadUserPrompts(session)
	if err != nil {
		return nil, err
	}
	return &SessionPrompts{Session: session, UserPrompts: prompts}, nil
}

// FilterUserPrompts applies grep keep → exclude drop → head|tail.
func FilterUserPrompts(prompts []UserPrompt, opts FilterUserPromptsOptions) (kept []UserPrompt, omittedBefore, omittedAfter int, err error) {
	if err := validateFilterUserPromptsOptions(opts); err != nil {
		return nil, 0, 0, err
	}

	kept = prompts
	if opts.GrepSet {
		patterns := opts.Grep
		var filtered []UserPrompt
		for _, p := range kept {
			if textContainsAllLiteralCI(p.Text, patterns) {
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
	if opts.GrepSet {
		if len(opts.Grep) == 0 {
			return fmt.Errorf("--grep pattern must not be empty")
		}
		for _, p := range opts.Grep {
			if strings.TrimSpace(p) == "" {
				return fmt.Errorf("--grep pattern must not be empty")
			}
		}
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

// StreamPromptsList walks sessions newest-first and writes each block immediately.
func StreamPromptsList(w io.Writer, codexHome string, opts ListPromptsOptions, fmtOpts FormatPromptsOptions) error {
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
	useColor := shouldColorPrompts(normalizePromptsColorMode(fmtOpts.ColorMode))
	now := fmtOpts.Now

	nSess := 0
	totalMsgs := 0
	var wroteAny bool

	err := forEachPromptSession(codexHome, opts, func(sp SessionPrompts) error {
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
			formatRelativeTime(sp.StartedAt, now),
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

func forEachPromptSession(codexHome string, opts ListPromptsOptions, fn func(SessionPrompts) error) error {
	filterOpts := filterOptsFromList(opts)
	if err := validateFilterUserPromptsOptions(filterOpts); err != nil {
		return err
	}

	var sessions []Session
	if opts.OnlySessionIDsSet {
		for _, id := range opts.OnlySessionIDs {
			id = strings.TrimSpace(id)
			if id == "" {
				continue
			}
			path, err := Find(codexHome, id)
			if err != nil {
				continue
			}
			s, err := sessionMetaFromFile(path)
			if err != nil {
				continue
			}
			sessions = append(sessions, s)
		}
		sort.SliceStable(sessions, func(i, j int) bool {
			if sessions[i].StartedAt.Equal(sessions[j].StartedAt) {
				return sessions[i].ID > sessions[j].ID
			}
			return sessions[i].StartedAt.After(sessions[j].StartedAt)
		})
	} else {
		var err error
		sessions, err = discoverSessions(codexHome)
		if err != nil {
			return err
		}
		sort.Slice(sessions, func(i, j int) bool {
			if sessions[i].StartedAt.Equal(sessions[j].StartedAt) {
				return sessions[i].ID > sessions[j].ID
			}
			return sessions[i].StartedAt.After(sessions[j].StartedAt)
		})
	}

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
	return nil
}

func listPromptsSessionCap(opts ListPromptsOptions) (int, bool) {
	if opts.OnlySessionIDsSet && !opts.LimitSet && !opts.RecentSet {
		return 0, false
	}
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
		if p.Timestamp.Before(start) || p.Timestamp.After(now) {
			continue
		}
		out = append(out, p)
	}
	return out
}

func loadUserPrompts(session Session) ([]UserPrompt, error) {
	all, err := loadChatMessagesFromRollout(session.Path)
	if err != nil {
		return nil, err
	}
	var out []UserPrompt
	idx := 0
	for _, m := range all {
		if m.Kind != MessageKindUser {
			continue
		}
		if strings.TrimSpace(m.Text) == "" {
			continue
		}
		idx++
		out = append(out, UserPrompt{
			Index:     idx,
			Timestamp: m.Timestamp,
			Text:      m.Text,
		})
	}
	return out, nil
}

// WritePromptsText streams compact prompt lines for one session to w.
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
	useColor := shouldColorPrompts(normalizePromptsColorMode(opts.ColorMode))
	return writeSessionPromptBody(w, sp, loc, useColor, opts)
}

func validateFormatPromptsOptions(opts FormatPromptsOptions) error {
	if opts.MaxBodySet && opts.MaxBodyRunes < 1 {
		return fmt.Errorf("--max-body must be >= 1 (got %d)", opts.MaxBodyRunes)
	}
	return nil
}

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

func normalizePromptsColorMode(mode string) string {
	mode = strings.TrimSpace(strings.ToLower(mode))
	if mode == "" {
		return "never"
	}
	return mode
}

func shouldColorPrompts(colorMode string) bool {
	switch strings.ToLower(strings.TrimSpace(colorMode)) {
	case "always":
		return true
	case "never":
		return false
	default:
		if os.Getenv("NO_COLOR") != "" {
			return false
		}
		return term.IsTerminal(int(os.Stdout.Fd()))
	}
}

func dimMeta(s string, on bool) string {
	if !on || s == "" {
		return s
	}
	return ansiDim + s + ansiReset
}

func formatPromptLine(p UserPrompt, loc *time.Location, useColor bool, opts FormatPromptsOptions) string {
	prefix := missingTimestampMarker
	if !p.Timestamp.IsZero() {
		if loc == nil {
			loc = time.Local
		}
		prefix = "[" + p.Timestamp.In(loc).Format("2006-01-02 15:04:05") + "]"
	}
	if useColor {
		prefix = dimMeta(prefix, true)
	}
	body := formatPromptBody(p.Text, opts, useColor)
	return prefix + " " + body
}

func formatPromptBody(text string, opts FormatPromptsOptions, useColor bool) string {
	collapsed := collapseWhitespace(text)

	patterns := opts.Grep
	if opts.GrepSet && len(patterns) > 0 {
		start, length, ok := findLiteralCI(collapsed, patterns[0])
		if ok {
			if opts.MaxBodySet {
				snippet := windowPromptBody(collapsed, start, length, opts.MaxBodyRunes)
				if useColor {
					return highlightAllLiteralCI(snippet, patterns)
				}
				return snippet
			}
			if useColor {
				return highlightAllLiteralCI(collapsed, patterns)
			}
			return collapsed
		}
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

func windowPromptBody(collapsed string, matchStart, matchLen, maxRunes int) string {
	runes := []rune(collapsed)
	total := len(runes)
	if maxRunes <= 0 || total <= maxRunes {
		return collapsed
	}

	matchStartRune := utf8.RuneCountInString(collapsed[:matchStart])
	matchEndRune := utf8.RuneCountInString(collapsed[:matchStart+matchLen])
	matchRuneCount := matchEndRune - matchStartRune

	if matchRuneCount >= maxRunes {
		return string(runes[matchStartRune : matchStartRune+maxRunes])
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
	if needBeforeEllipsis {
		b.WriteString(promptTruncateEllipsis)
	}
	beforeStart := matchStartRune - contentBefore
	b.WriteString(string(runes[beforeStart:matchStartRune]))
	b.WriteString(string(runes[matchStartRune:matchEndRune]))
	b.WriteString(string(runes[matchEndRune : matchEndRune+contentAfter]))
	if needAfterEllipsis {
		b.WriteString(promptTruncateEllipsis)
	}
	return b.String()
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

func collapseWhitespace(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

func textContainsAllLiteralCI(s string, patterns []string) bool {
	if len(patterns) == 0 {
		return true
	}
	for _, p := range patterns {
		if _, _, ok := findLiteralCI(s, p); !ok {
			return false
		}
	}
	return true
}

func findLiteralCI(s, pattern string) (start, length int, ok bool) {
	if pattern == "" || s == "" {
		return 0, 0, false
	}
	lowerS := strings.ToLower(s)
	lowerP := strings.ToLower(pattern)
	idx := strings.Index(lowerS, lowerP)
	if idx < 0 {
		return 0, 0, false
	}
	return idx, len(lowerP), true
}

func highlightAllLiteralCI(s string, patterns []string) string {
	if s == "" || len(patterns) == 0 {
		return s
	}
	var lowerPats []string
	for _, p := range patterns {
		if p == "" {
			continue
		}
		lowerPats = append(lowerPats, strings.ToLower(p))
	}
	if len(lowerPats) == 0 {
		return s
	}
	lower := strings.ToLower(s)
	var b strings.Builder
	i := 0
	for i < len(s) {
		bestStart := -1
		bestEnd := -1
		for _, lp := range lowerPats {
			rel := strings.Index(lower[i:], lp)
			if rel < 0 {
				continue
			}
			start := i + rel
			end := start + len(lp)
			if bestStart < 0 || start < bestStart || (start == bestStart && end > bestEnd) {
				bestStart = start
				bestEnd = end
			}
		}
		if bestStart < 0 {
			b.WriteString(s[i:])
			break
		}
		b.WriteString(s[i:bestStart])
		b.WriteString(ansiBoldRed)
		b.WriteString(s[bestStart:bestEnd])
		b.WriteString(ansiReset)
		i = bestEnd
	}
	return b.String()
}
