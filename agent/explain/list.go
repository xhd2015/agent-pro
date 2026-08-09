package explain

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/xhd2015/less-gen/flags"
	"golang.org/x/term"
)

const (
	defaultListLimit = 10
	maxListLimit     = 100

	// boldRedOpen / sgrReset wrap grep match spans when color is on.
	boldRedOpen = "\x1b[1;31m"
	sgrReset    = "\x1b[0m"
)

var listHelp = `Usage: explain list [--limit N] [--grep PATTERN]... [--or|--and] [--color]

List recent explain sessions (newest first) with Q/A cards.
Optional --grep filters sessions by case-insensitive literal substring
match against Q/A message bodies. Multiple --grep patterns OR by default;
use --or for explicit OR or --and to require every pattern (session-level).

Options:
  --limit N   Max sessions to show (default 10, max 100; <=0 uses default)
  --grep P    Keep sessions whose Q/A bodies contain P (repeatable; case-insensitive literal)
  --or        Combine multiple --grep with OR (default when greps are set)
  --and       Combine multiple --grep with AND (patterns may hit different messages)
  --color     Force ANSI color on (overrides NO_COLOR and non-TTY)
  -h, --help  Show this help message
`

// RunList handles `explain list` with optional --limit, --grep, --or/--and, and --color.
// It never starts or resumes an agent.
func RunList(args []string) error {
	var limit int
	var color bool
	var greps []string
	var orFlag, andFlag bool
	_, err := flags.
		Int("--limit", &limit).
		StringSlice("--grep", &greps).
		Bool("--or", &orFlag).
		Bool("--and", &andFlag).
		Bool("--color", &color).
		Help("-h,--help", listHelp).
		Parse(args)
	if err != nil {
		return err
	}

	if err := validateListGrepFlags(greps, orFlag, andFlag); err != nil {
		return err
	}

	limit = normalizeListLimit(limit)
	useColor := resolveListColor(color)
	requireAll := andFlag // default and explicit --or are OR (requireAll=false)

	all, err := listSessions()
	if err != nil {
		return err
	}
	sort.Slice(all, func(i, j int) bool {
		return all[i].timestamp.After(all[j].timestamp)
	})
	storeCount := len(all)

	// Pipeline: list → sort newest-first → filter → total=match count → limit.
	sessions := all
	if len(greps) > 0 {
		sessions = filterSessionsByGrep(all, greps, requireAll)
	}
	total := len(sessions)
	if len(sessions) > limit {
		sessions = sessions[:limit]
	}

	fmt.Print(formatListOutput(sessions, total, limit, useColor, greps, storeCount))
	return nil
}

func validateListGrepFlags(greps []string, orFlag, andFlag bool) error {
	if orFlag && andFlag {
		return fmt.Errorf("cannot use both --or and --and")
	}
	if (orFlag || andFlag) && len(greps) == 0 {
		return fmt.Errorf("%s requires at least one --grep", modeFlagName(orFlag, andFlag))
	}
	for _, g := range greps {
		if g == "" {
			return fmt.Errorf("--grep pattern must be non-empty")
		}
	}
	return nil
}

func modeFlagName(orFlag, andFlag bool) string {
	if orFlag {
		return "--or"
	}
	if andFlag {
		return "--and"
	}
	return "--or/--and"
}

func normalizeListLimit(limit int) int {
	if limit <= 0 {
		return defaultListLimit
	}
	if limit > maxListLimit {
		return maxListLimit
	}
	return limit
}

func resolveListColor(forceColor bool) bool {
	if forceColor {
		return true
	}
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	return term.IsTerminal(int(os.Stdout.Fd()))
}

// filterSessionsByGrep keeps sessions matching greps (already sorted).
// requireAll=true → AND (every pattern somewhere in message bodies);
// requireAll=false → OR (any pattern).
func filterSessionsByGrep(sessions []sessionDir, greps []string, requireAll bool) []sessionDir {
	if len(greps) == 0 {
		return sessions
	}
	var out []sessionDir
	for _, s := range sessions {
		if sessionMatchesGrep(s, greps, requireAll) {
			out = append(out, s)
		}
	}
	return out
}

// sessionMatchesGrep reports whether session Q/A message bodies match greps.
// Only Messages[].Message is searched (not runner/model/dirname/role).
func sessionMatchesGrep(s sessionDir, greps []string, requireAll bool) bool {
	if len(greps) == 0 {
		return true
	}
	if requireAll {
		for _, p := range greps {
			if !sessionHasPattern(s, p) {
				return false
			}
		}
		return true
	}
	for _, p := range greps {
		if sessionHasPattern(s, p) {
			return true
		}
	}
	return false
}

func sessionHasPattern(s sessionDir, pattern string) bool {
	if pattern == "" {
		return false
	}
	pl := strings.ToLower(pattern)
	for _, m := range s.data.Messages {
		if strings.Contains(strings.ToLower(m.Message), pl) {
			return true
		}
	}
	return false
}

// formatListOutput builds list stdout.
// storeCount is the number of sessions before grep filter (empty store vs no-match).
// greps are used for match-span highlighting when useColor is true.
func formatListOutput(sessions []sessionDir, total, limit int, useColor bool, greps []string, storeCount int) string {
	if len(sessions) == 0 {
		if storeCount == 0 || len(greps) == 0 {
			return "No explain sessions yet.\n"
		}
		return "No matching explain sessions.\n"
	}

	var b strings.Builder
	fmt.Fprintf(&b, "Recent explain sessions (%d shown of %d, limit %d)\n\n", len(sessions), total, limit)

	highlight := useColor && len(greps) > 0
	for i, s := range sessions {
		if i > 0 {
			b.WriteByte('\n')
		}
		b.WriteString(formatSessionCard(i+1, s, useColor, greps, highlight))
	}
	return b.String()
}

func formatSessionCard(index int, s sessionDir, useColor bool, greps []string, highlight bool) string {
	turns := countUserTurns(s.data)
	turnLabel := "turns"
	if turns == 1 {
		turnLabel = "turn"
	}

	ts := s.timestamp.Format("2006-01-02 15:04:05")
	meta := formatRunnerModel(s.data.AgentRunner, s.data.Model)
	header := fmt.Sprintf("── %d ──  %s  ·  %s  ·  %d %s", index, ts, meta, turns, turnLabel)
	if useColor {
		header = "\x1b[2m" + header + "\x1b[0m"
	}

	var b strings.Builder
	b.WriteString(header)
	b.WriteByte('\n')

	for _, m := range s.data.Messages {
		label, ok := roleLabel(m.Role)
		if !ok {
			continue
		}
		colored := label
		if useColor {
			switch label {
			case "Q":
				colored = "\x1b[1;36mQ\x1b[0m"
			case "A":
				colored = "\x1b[1;32mA\x1b[0m"
			}
		}
		body := m.Message
		if highlight {
			body = highlightGrepMatches(body, greps)
		}
		b.WriteString(formatMessageBody(colored, body))
	}
	return b.String()
}

// formatMessageBody prints a Q/A body in full (no truncate/collapse).
// First line: three spaces + label + two spaces + first line.
// Continuation non-empty lines: exactly 6 spaces + line.
// Empty segments: pure "\n" (no spaces-only blank lines).
func formatMessageBody(coloredLabel, body string) string {
	lines := strings.Split(body, "\n")
	var b strings.Builder
	// Three spaces before label, two spaces after label.
	fmt.Fprintf(&b, "   %s  %s\n", coloredLabel, lines[0])
	for _, line := range lines[1:] {
		if line == "" {
			b.WriteByte('\n')
		} else {
			b.WriteString("      ")
			b.WriteString(line)
			b.WriteByte('\n')
		}
	}
	return b.String()
}

// highlightGrepMatches wraps case-insensitive literal matches of patterns in
// bold red SGR, preserving original casing. Processing is per-line;
// leftmost-first, no nested/overlapping spans. All patterns are considered.
func highlightGrepMatches(text string, patterns []string) string {
	if text == "" || len(patterns) == 0 {
		return text
	}
	// Pre-lower patterns once; skip empties (validated earlier, still defensive).
	var lowerPats []string
	for _, p := range patterns {
		if p == "" {
			continue
		}
		lowerPats = append(lowerPats, strings.ToLower(p))
	}
	if len(lowerPats) == 0 {
		return text
	}

	lines := strings.Split(text, "\n")
	for i, line := range lines {
		lines[i] = highlightLine(line, lowerPats)
	}
	return strings.Join(lines, "\n")
}

// highlightLine highlights non-overlapping match spans on a single line.
// lowerPats are already lowercased. Original casing of line is preserved.
func highlightLine(line string, lowerPats []string) string {
	if line == "" || len(lowerPats) == 0 {
		return line
	}
	lower := strings.ToLower(line)
	var b strings.Builder
	i := 0
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
			// Leftmost-first; on equal start prefer longer span (no nested leftover).
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
		b.WriteString(boldRedOpen)
		b.WriteString(line[bestStart:bestEnd])
		b.WriteString(sgrReset)
		i = bestEnd
	}
	return b.String()
}

func formatRunnerModel(runner, model string) string {
	runner = strings.TrimSpace(runner)
	model = strings.TrimSpace(model)
	switch {
	case runner != "" && model != "":
		return runner + " / " + model
	case runner != "":
		return runner
	case model != "":
		return model
	default:
		return "unknown"
	}
}

func roleLabel(role string) (string, bool) {
	switch role {
	case "user":
		return "Q", true
	case "assistant":
		return "A", true
	default:
		return "", false
	}
}

func countUserTurns(data SessionData) int {
	n := 0
	for _, m := range data.Messages {
		if m.Role == "user" {
			n++
		}
	}
	return n
}
