package ttywatchtest

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/hinshun/vt10x"
)

var screenSnapshotMarker = []byte("\x1b[?25l")

// ScreenSnapshotTextFromScrollback renders scrollback the way ptywrap screen attach
// does, then converts the snapshot frame to plain text the way tty-watch attach does.
func ScreenSnapshotTextFromScrollback(scrollback []byte, cols, rows int) (string, bool) {
	snapshot, ok := renderScreenSnapshotFrame(scrollback, cols, rows)
	if !ok {
		return "", false
	}
	text, ok := screenSnapshotFrameToText(snapshot, cols, rows)
	if !ok {
		return "", false
	}
	return string(text), true
}

// ContentLinesAtColumnZero returns trimmed logical content lines that must start at
// column 0. Unlike VisibleContentLines, leading whitespace is preserved and rejected.
func ContentLinesAtColumnZero(s string) ([]string, error) {
	s = csiStripRe.ReplaceAllString(s, "")
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	var out []string
	for _, line := range strings.Split(s, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if strings.HasPrefix(line, " ") || strings.HasPrefix(line, "\t") {
			return nil, fmt.Errorf("line has leading whitespace: %q", line)
		}
		out = append(out, trimmed)
	}
	return out, nil
}

// NoLeadingBlankLine reports an error when s begins with a blank line before the first
// visible content (e.g. relayTerminalOutput prepending "\n" before screen snapshot text).
func NoLeadingBlankLine(s string) error {
	if s == "" {
		return nil
	}
	normalized := csiStripRe.ReplaceAllString(s, "")
	normalized = strings.ReplaceAll(normalized, "\r\n", "\n")
	normalized = strings.ReplaceAll(normalized, "\r", "\n")
	if strings.HasPrefix(normalized, "\n") {
		return fmt.Errorf("output has leading blank line before first content")
	}
	return nil
}

// HasLeadingBlankLine reports whether s begins with a blank line before first content.
func HasLeadingBlankLine(s string) bool {
	return NoLeadingBlankLine(s) != nil
}

// HasIndentedExitMarker reports output where [Terminal exited] is not left-aligned.
func HasIndentedExitMarker(s string) bool {
	if strings.Contains(s, "\n   [Terminal exited]") {
		return true
	}
	if strings.Contains(s, "\r   [Terminal exited]") {
		return true
	}
	for _, line := range strings.Split(s, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "[Terminal exited]" && line != trimmed {
			return true
		}
	}
	return false
}

// EndsWithNewline reports whether s ends with a newline (host prompt must start fresh).
func EndsWithNewline(s string) bool {
	return len(s) > 0 && s[len(s)-1] == '\n'
}

func renderScreenSnapshotFrame(scrollback []byte, cols, rows int) ([]byte, bool) {
	if len(scrollback) == 0 {
		return nil, false
	}
	if cols <= 0 {
		cols = 80
	}
	if rows <= 0 {
		rows = 24
	}

	vt := vt10x.New(vt10x.WithSize(cols, rows))
	if _, err := vt.Write(scrollback); err != nil {
		return nil, false
	}

	vt.Lock()
	defer vt.Unlock()

	var out strings.Builder
	out.WriteString("\x1b[?25l\x1b[?1049l\x1b[0m\x1b[H\x1b[2J")
	for y := 0; y < rows; y++ {
		line := renderSnapshotGridLine(vt, cols, y)
		if line == "" {
			continue
		}
		fmt.Fprintf(&out, "\x1b[%d;1H%s", y+1, line)
	}
	cursor := vt.Cursor()
	fmt.Fprintf(&out, "\x1b[%d;%dH", cursor.Y+1, cursor.X+1)
	return []byte(out.String()), true
}

func screenSnapshotFrameToText(data []byte, cols, rows int) ([]byte, bool) {
	if !bytes.HasPrefix(data, screenSnapshotMarker) || !bytes.Contains(data, []byte("\x1b[2J")) {
		return nil, false
	}
	if cols <= 0 {
		cols = 80
	}
	if rows <= 0 {
		rows = 24
	}

	vt := vt10x.New(vt10x.WithSize(cols, rows))
	if _, err := vt.Write(data); err != nil {
		return nil, false
	}

	vt.Lock()
	defer vt.Unlock()

	var lines []string
	for y := 0; y < rows; y++ {
		line := renderSnapshotGridLine(vt, cols, y)
		if line != "" {
			lines = append(lines, line)
		}
	}
	if len(lines) == 0 {
		return nil, false
	}
	return []byte(strings.Join(lines, "\n") + "\n"), true
}

func renderSnapshotGridLine(vt vt10x.Terminal, cols, y int) string {
	runes := make([]rune, cols)
	lastNonSpace := -1
	for x := 0; x < cols; x++ {
		ch := vt.Cell(x, y).Char
		if ch == 0 {
			ch = ' '
		}
		runes[x] = ch
		if ch != ' ' {
			lastNonSpace = x
		}
	}
	if lastNonSpace < 0 {
		return ""
	}
	firstNonSpace := 0
	for firstNonSpace <= lastNonSpace && (runes[firstNonSpace] == ' ' || runes[firstNonSpace] == '\t') {
		firstNonSpace++
	}
	return string(runes[firstNonSpace : lastNonSpace+1])
}