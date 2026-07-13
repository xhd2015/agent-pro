package explain

import (
	"fmt"
	"os"
	"strings"

	"github.com/xhd2015/less-gen/flags"
	"golang.org/x/term"
)

const (
	defaultListLimit = 10
	maxListLimit     = 100
)

var listHelp = `Usage: explain list [--limit N] [--color]

List recent explain sessions (newest first) with Q/A cards.

Options:
  --limit N   Max sessions to show (default 10, max 100; <=0 uses default)
  --color     Force ANSI color on (overrides NO_COLOR and non-TTY)
  -h, --help  Show this help message
`

// RunList handles `explain list` with optional --limit and --color.
// It never starts or resumes an agent.
func RunList(args []string) error {
	var limit int
	var color bool
	_, err := flags.
		Int("--limit", &limit).
		Bool("--color", &color).
		Help("-h,--help", listHelp).
		Parse(args)
	if err != nil {
		return err
	}

	limit = normalizeListLimit(limit)
	useColor := resolveListColor(color)

	sessions, total, err := ListRecentSessions(limit)
	if err != nil {
		return err
	}

	fmt.Print(formatListOutput(sessions, total, limit, useColor))
	return nil
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

func formatListOutput(sessions []sessionDir, total, limit int, useColor bool) string {
	if len(sessions) == 0 {
		return "No explain sessions yet.\n"
	}

	var b strings.Builder
	fmt.Fprintf(&b, "Recent explain sessions (%d shown of %d, limit %d)\n\n", len(sessions), total, limit)

	for i, s := range sessions {
		if i > 0 {
			b.WriteByte('\n')
		}
		b.WriteString(formatSessionCard(i+1, s, useColor))
	}
	return b.String()
}

func formatSessionCard(index int, s sessionDir, useColor bool) string {
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
		b.WriteString(formatMessageBody(colored, m.Message))
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
