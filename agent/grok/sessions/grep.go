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
	"unicode/utf8"

	"golang.org/x/term"
)

const maxDisplayedHits = 5

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
}

// ListWithGrep discovers sessions, keeps those with ≥1 case-insensitive
// literal hit in summary.json or chat_history.jsonl, sorts by last_active_at
// desc, then applies limit.
func ListWithGrep(grokHome string, limit int, pattern string) ([]SessionMatch, error) {
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

	pattern = strings.TrimSpace(pattern)
	if pattern == "" {
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
		hits := searchSession(s, pattern)
		if len(hits) == 0 {
			continue
		}
		matches = append(matches, SessionMatch{Session: s, Hits: hits})
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
	fmt.Fprintf(&b, "%-38s  %-12s  %-42s  %5s  %s\n", "SESSION ID", "LAST ACTIVE", "TITLE", "MSGS", "CWD")
	for _, m := range matches {
		fmt.Fprintf(
			&b,
			"%-38s  %-12s  %-42s  %5d  %s\n",
			m.ID,
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
			b.WriteString(formatHitLine(h, useColor))
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

func formatHitLine(h MatchHit, useColor bool) string {
	if !useColor {
		return fmt.Sprintf("  %s:%d:%s: %s", h.File, h.Line, h.Part, h.Snippet)
	}

	const (
		reset  = "\x1b[0m"
		mag    = "\x1b[35m"
		green  = "\x1b[32m"
		bold   = "\x1b[1m"
		red    = "\x1b[31m"
	)

	snippet := h.Snippet
	if h.MatchLen > 0 && h.MatchStart >= 0 && h.MatchStart+h.MatchLen <= len(snippet) {
		before := snippet[:h.MatchStart]
		match := snippet[h.MatchStart : h.MatchStart+h.MatchLen]
		after := snippet[h.MatchStart+h.MatchLen:]
		snippet = before + bold + red + match + reset + after
	}

	return fmt.Sprintf(
		"  %s%s%s:%s%d%s:%s%s%s: %s",
		mag, h.File, reset,
		green, h.Line, reset,
		green, h.Part, reset,
		snippet,
	)
}

func searchSession(session Session, pattern string) []MatchHit {
	var hits []MatchHit
	hits = append(hits, searchSummaryFile(session.Path, pattern)...)
	chatPath := filepath.Join(filepath.Dir(session.Path), "chat_history.jsonl")
	hits = append(hits, searchChatHistory(chatPath, pattern)...)
	return hits
}

func searchSummaryFile(path, pattern string) []MatchHit {
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
		if h, ok := makeHit("summary.json", 1, f.part, f.text, pattern); ok {
			hits = append(hits, h)
		}
	}
	return hits
}

func searchChatHistory(path, pattern string) []MatchHit {
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
		if h, ok := makeHit("chat_history.jsonl", lineNum, part, text, pattern); ok {
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

func makeHit(file string, line int, part, text, pattern string) (MatchHit, bool) {
	snippet := collapseToOneLine(text)
	start, length, ok := findLiteralCI(snippet, pattern)
	if !ok {
		return MatchHit{}, false
	}
	return MatchHit{
		File:       file,
		Line:       line,
		Part:       part,
		Snippet:    snippet,
		MatchStart: start,
		MatchLen:   length,
	}, true
}

func collapseToOneLine(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	s = strings.ReplaceAll(s, "\n", " ")
	// collapse runs of spaces from newline replacement only lightly
	for strings.Contains(s, "  ") {
		s = strings.ReplaceAll(s, "  ", " ")
	}
	return s
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
