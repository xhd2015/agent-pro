package sessions

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"golang.org/x/term"
)

const maxDisplayedHits = 5
const maxSnippetRunes = 1024
const snippetEllipsis = "..."
const snippetEllipsisRunes = 3

// MatchHit is one content match inside a session file.
type MatchHit struct {
	File       string // "summary.json", "chat_history.jsonl"
	Line       int    // 1-based
	Part       string // title, session_summary, cwd, user, assistant, ...
	Snippet    string // one-line excerpt
	MatchStart int    // byte offset into Snippet
	MatchLen   int    // length of match in Snippet
}

// SessionMatch is a session that matched a grep pattern, with all hits.
type SessionMatch struct {
	Session
	Hits []MatchHit // full hit list; formatter caps display at 5
	Grep []string   // patterns used for this search (multi-span highlight)
}

// ListWithGrep discovers sessions, keeps those with ≥1 case-insensitive
// literal hit in summary.json or chat_history.jsonl, sorts by last_active_at
// desc, then applies limit.
//
// patterns is AND on the same field/line: a hit requires every pattern as a
// substring of that unit. Empty/nil patterns → no filter (all sessions, empty
// hits); CLI rejects empty --grep before calling.
func ListWithGrep(grokHome string, limit int, patterns []string) ([]SessionMatch, error) {
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

	patterns, err = normalizeGrepPatternsOptional(patterns)
	if err != nil {
		return nil, err
	}
	if len(patterns) == 0 {
		// No filter: wrap all sessions with empty hits (CLI rejects empty --grep).
		matches := make([]SessionMatch, 0, len(sessions))
		for _, s := range sessions {
			matches = append(matches, SessionMatch{Session: s})
		}
		sortSessionMatches(matches)
		if len(matches) > limit {
			matches = matches[:limit]
		}
		return matches, nil
	}

	var matches []SessionMatch
	for _, s := range sessions {
		hits := searchSession(s, patterns)
		if len(hits) == 0 {
			continue
		}
		matches = append(matches, SessionMatch{Session: s, Hits: hits, Grep: patterns})
	}

	sortSessionMatches(matches)
	if len(matches) > limit {
		matches = matches[:limit]
	}
	return matches, nil
}

// FormatListTableWithHits formats matching sessions like FormatListTable, with
// indented hit lines under each row. colorMode is "never" | "always" | "auto".
// Output has no trailing newline (same TrimRight style as FormatListTable).
func FormatListTableWithHits(matches []SessionMatch, home string, now time.Time, colorMode string) string {
	if len(matches) == 0 {
		return "No sessions found"
	}

	useColor := shouldColor(colorMode)
	var b strings.Builder
	fmt.Fprintf(&b, "%-38s  %-5s  %-12s  %-42s  %5s  %s\n", "SESSION ID", "KIND", "LAST ACTIVE", "TITLE", "MSGS", "CWD")
	for _, m := range matches {
		fmt.Fprintf(
			&b,
			"%-38s  %-5s  %-12s  %-42s  %5d  %s\n",
			m.ID,
			sessionKindOrMain(m.Session),
			formatRelativeTime(m.LastActiveAt, now),
			truncateTitle(m.Title),
			m.NumChatMessages,
			shortenPath(m.CWD, home),
		)

		hits := m.Hits
		show := hits
		if len(show) > maxDisplayedHits {
			show = hits[:maxDisplayedHits]
		}
		for _, h := range show {
			b.WriteString(formatHitLine(h, useColor, m.Grep))
			b.WriteByte('\n')
		}
		if len(hits) > maxDisplayedHits {
			fmt.Fprintf(&b, "  ... and %d more matches\n", len(hits)-maxDisplayedHits)
		}
	}
	return strings.TrimRight(b.String(), "\n")
}

func sortSessionMatches(matches []SessionMatch) {
	sort.Slice(matches, func(i, j int) bool {
		if matches[i].LastActiveAt.Equal(matches[j].LastActiveAt) {
			return matches[i].ID > matches[j].ID
		}
		return matches[i].LastActiveAt.After(matches[j].LastActiveAt)
	})
}

func shouldColor(colorMode string) bool {
	switch strings.ToLower(strings.TrimSpace(colorMode)) {
	case "always":
		return true
	case "never":
		return false
	default: // auto
		if os.Getenv("NO_COLOR") != "" {
			return false
		}
		return term.IsTerminal(int(os.Stdout.Fd()))
	}
}

func formatHitLine(h MatchHit, useColor bool, patterns []string) string {
	if !useColor {
		return fmt.Sprintf("  %s:%d:%s: %s", h.File, h.Line, h.Part, h.Snippet)
	}

	const (
		reset = "\x1b[0m"
		mag   = "\x1b[35m"
		green = "\x1b[32m"
	)

	snippet := h.Snippet
	if len(patterns) > 0 {
		snippet = highlightAllLiteralCI(snippet, patterns)
	} else if h.MatchLen > 0 && h.MatchStart >= 0 && h.MatchStart+h.MatchLen <= len(snippet) {
		before := snippet[:h.MatchStart]
		match := snippet[h.MatchStart : h.MatchStart+h.MatchLen]
		after := snippet[h.MatchStart+h.MatchLen:]
		snippet = before + "\x1b[1m\x1b[31m" + match + reset + after
	}

	return fmt.Sprintf(
		"  %s%s%s:%s%d%s:%s%s%s: %s",
		mag, h.File, reset,
		green, h.Line, reset,
		green, h.Part, reset,
		snippet,
	)
}

func searchSession(session Session, patterns []string) []MatchHit {
	var hits []MatchHit
	hits = append(hits, searchSummaryFile(session.Path, patterns)...)
	chatPath := filepath.Join(filepath.Dir(session.Path), "chat_history.jsonl")
	hits = append(hits, searchChatHistory(chatPath, patterns)...)
	return hits
}

func searchSummaryFile(path string, patterns []string) []MatchHit {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	data = []byte(strings.TrimSpace(string(data)))
	if len(data) == 0 {
		return nil
	}

	var summary grokSummary
	if err := json.Unmarshal(data, &summary); err != nil {
		return nil
	}

	// Prefer order: title, session_summary, cwd, then other metadata.
	fields := []struct {
		part string
		text string
	}{
		{"title", strings.TrimSpace(summary.GeneratedTitle)},
		{"session_summary", strings.TrimSpace(summary.SessionSummary)},
		{"cwd", strings.TrimSpace(summary.Info.CWD)},
		{"model", strings.TrimSpace(summary.CurrentModelID)},
		{"agent", strings.TrimSpace(summary.AgentName)},
	}

	var hits []MatchHit
	for _, f := range fields {
		if f.text == "" {
			continue
		}
		if h, ok := makeHit("summary.json", 1, f.part, f.text, patterns); ok {
			hits = append(hits, h)
		}
	}
	return hits
}

func searchChatHistory(path string, patterns []string) []MatchHit {
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()

	var hits []MatchHit
	scanner := bufio.NewScanner(f)
	// chat lines can be large; raise limit above default 64K
	scanner.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)

	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}

		msgType, text := extractChatLineText(line)
		if text == "" {
			continue
		}
		part := msgType
		if part == "" {
			part = "message"
		}
		if h, ok := makeHit("chat_history.jsonl", lineNum, part, text, patterns); ok {
			hits = append(hits, h)
		}
	}
	return hits
}

func extractChatLineText(line string) (msgType, text string) {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal([]byte(line), &raw); err != nil {
		return "", ""
	}

	if t, ok := raw["type"]; ok {
		_ = json.Unmarshal(t, &msgType)
	}
	msgType = strings.TrimSpace(msgType)

	switch msgType {
	case "user":
		return msgType, extractContentText(raw["content"])
	case "reasoning":
		return msgType, extractSummaryText(raw["summary"])
	default:
		// system, assistant, tool_result, or string content types
		return msgType, extractContentText(raw["content"])
	}
}

func extractContentText(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}
	// string content
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		return s
	}
	// array of {type, text}
	var parts []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	}
	if err := json.Unmarshal(raw, &parts); err == nil {
		var b strings.Builder
		for _, p := range parts {
			if p.Text == "" {
				continue
			}
			if b.Len() > 0 {
				b.WriteByte(' ')
			}
			b.WriteString(p.Text)
		}
		return b.String()
	}
	return ""
}

func extractSummaryText(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}
	var parts []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	}
	if err := json.Unmarshal(raw, &parts); err == nil {
		var b strings.Builder
		for _, p := range parts {
			if p.Text == "" {
				continue
			}
			if b.Len() > 0 {
				b.WriteByte(' ')
			}
			b.WriteString(p.Text)
		}
		return b.String()
	}
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		return s
	}
	return ""
}

func makeHit(file string, line int, part, text string, patterns []string) (MatchHit, bool) {
	snippet := collapseWhitespace(text)
	if !textContainsAllLiteralCI(snippet, patterns) {
		return MatchHit{}, false
	}
	// Window around the first pattern's first hit (stable with single --grep).
	start, length, ok := findLiteralCI(snippet, patterns[0])
	if !ok {
		return MatchHit{}, false
	}
	snippet, start, length = windowSnippet(snippet, start, length)
	return MatchHit{
		File:       file,
		Line:       line,
		Part:       part,
		Snippet:    snippet,
		MatchStart: start,
		MatchLen:   length,
	}, true
}

// normalizeGrepPatternsOptional trims patterns. Empty input → nil, nil (no filter).
// Any empty-after-trim entry → error. Does not require GrepSet.
func normalizeGrepPatternsOptional(patterns []string) ([]string, error) {
	if len(patterns) == 0 {
		return nil, nil
	}
	return normalizeGrepPatterns(patterns)
}

// normalizeGrepPatterns trims each pattern and rejects empties. Requires ≥1 pattern.
func normalizeGrepPatterns(patterns []string) ([]string, error) {
	if len(patterns) == 0 {
		return nil, fmt.Errorf("grep pattern must not be empty")
	}
	out := make([]string, 0, len(patterns))
	for _, p := range patterns {
		p = strings.TrimSpace(p)
		if p == "" {
			return nil, fmt.Errorf("grep pattern must not be empty")
		}
		out = append(out, p)
	}
	return out, nil
}

// validateGrepPatterns returns normalized patterns when set is true.
// set && empty/invalid → error. !set → nil, nil.
func validateGrepPatterns(set bool, patterns []string) ([]string, error) {
	if !set {
		return nil, nil
	}
	return normalizeGrepPatterns(patterns)
}

// textContainsAllLiteralCI reports whether s contains every pattern as a
// case-insensitive literal substring (AND on the same string).
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

// highlightAllLiteralCI wraps case-insensitive literal matches of patterns in
// bold-red SGR. Leftmost-first, non-overlapping; longer span wins on equal start.
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
	return highlightLiteralLine(s, lowerPats)
}

func highlightLiteralLine(line string, lowerPats []string) string {
	if line == "" || len(lowerPats) == 0 {
		return line
	}
	lower := strings.ToLower(line)
	var b strings.Builder
	i := 0
	const (
		boldRed = "\x1b[1m\x1b[31m"
		reset   = "\x1b[0m"
	)
	for i < len(line) {
		bestStart := -1
		bestEnd := -1
		for _, lp := range lowerPats {
			if lp == "" {
				continue
			}
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
			b.WriteString(line[i:])
			break
		}
		b.WriteString(line[i:bestStart])
		b.WriteString(boldRed)
		b.WriteString(line[bestStart:bestEnd])
		b.WriteString(reset)
		i = bestEnd
	}
	return b.String()
}

// collapseWhitespace replaces runs of Unicode space (spaces/tabs/newlines/etc.)
// with a single ASCII space and trims edges.
func collapseWhitespace(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	prevSpace := true // skip leading whitespace
	for _, r := range s {
		if unicode.IsSpace(r) {
			if !prevSpace {
				b.WriteByte(' ')
				prevSpace = true
			}
			continue
		}
		b.WriteRune(r)
		prevSpace = false
	}
	out := b.String()
	if prevSpace && len(out) > 0 {
		// trailing space written before end
		out = out[:len(out)-1]
	}
	return out
}

// windowSnippet builds a ≤maxSnippetRunes window around the match in collapsed
// text. MatchStart/MatchLen are byte offsets into the returned snippet.
func windowSnippet(collapsed string, matchStart, matchLen int) (snippet string, newStart, newLen int) {
	runes := []rune(collapsed)
	total := len(runes)
	if total <= maxSnippetRunes {
		return collapsed, matchStart, matchLen
	}

	// Byte offsets → rune indices for the match span.
	matchStartRune := utf8.RuneCountInString(collapsed[:matchStart])
	matchEndRune := utf8.RuneCountInString(collapsed[:matchStart+matchLen])
	matchRuneCount := matchEndRune - matchStartRune

	// Match alone fills (or exceeds) the budget: first maxSnippetRunes of match.
	if matchRuneCount >= maxSnippetRunes {
		s := string(runes[matchStartRune : matchStartRune+maxSnippetRunes])
		return s, 0, len(s)
	}

	remaining := maxSnippetRunes - matchRuneCount
	// ~50/50 before/after; odd remainder prefers after.
	beforeBudget := remaining / 2
	afterBudget := remaining - beforeBudget

	beforeAvail := matchStartRune
	afterAvail := total - matchEndRune

	takeBefore := beforeBudget
	takeAfter := afterBudget

	// Reallocate unused budget from a short side to the long side.
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

	// Ellipsis only on sides that were cut; reserve 3 runes of that side's budget.
	needBeforeEllipsis := takeBefore < beforeAvail
	needAfterEllipsis := takeAfter < afterAvail

	contentBefore := takeBefore
	contentAfter := takeAfter
	if needBeforeEllipsis {
		contentBefore = takeBefore - snippetEllipsisRunes
		if contentBefore < 0 {
			contentBefore = 0
		}
	}
	if needAfterEllipsis {
		contentAfter = takeAfter - snippetEllipsisRunes
		if contentAfter < 0 {
			contentAfter = 0
		}
	}

	var b strings.Builder
	// Rough capacity: all ASCII in tests, but grow for multi-byte safely.
	b.Grow(maxSnippetRunes * utf8.UTFMax)

	if needBeforeEllipsis {
		b.WriteString(snippetEllipsis)
	}
	beforeStart := matchStartRune - contentBefore
	b.WriteString(string(runes[beforeStart:matchStartRune]))
	newStart = b.Len()
	matchStr := string(runes[matchStartRune:matchEndRune])
	b.WriteString(matchStr)
	newLen = len(matchStr)
	b.WriteString(string(runes[matchEndRune : matchEndRune+contentAfter]))
	if needAfterEllipsis {
		b.WriteString(snippetEllipsis)
	}
	return b.String(), newStart, newLen
}

// findLiteralCI finds the first case-insensitive literal occurrence of pattern
// in s. MatchLen is the byte length of the matched span in the original string.
func findLiteralCI(s, pattern string) (start, length int, ok bool) {
	if pattern == "" || s == "" {
		return 0, 0, false
	}
	// Fast path: ASCII-only case fold via ToLower (tests and typical tokens are ASCII).
	lowerS := strings.ToLower(s)
	lowerP := strings.ToLower(pattern)
	idx := strings.Index(lowerS, lowerP)
	if idx < 0 {
		return 0, 0, false
	}
	// For ASCII, lowered length equals original match length.
	// Guard multi-byte: take the same number of bytes as lowerP from s at idx
	// when both are valid UTF-8 of equal rune lengths.
	matchLen := len(lowerP)
	if idx+matchLen > len(s) {
		// Fallback: measure by runes if lower changed byte length (rare).
		runes := []rune(s[idx:])
		patRunes := []rune(lowerP)
		if len(runes) < len(patRunes) {
			return 0, 0, false
		}
		matchLen = 0
		for i := 0; i < len(patRunes); i++ {
			matchLen += utf8.RuneLen(runes[i])
		}
	}
	return idx, matchLen, true
}
