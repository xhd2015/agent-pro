package main

import (
	"bytes"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/hinshun/vt10x"
)

var snapshotCUPRe = regexp.MustCompile(`\x1b\[(\d+);(\d+)[Hf]`)

var screenSnapshotMarker = []byte("\x1b[?25l")

func isScreenSnapshotFrame(data []byte) bool {
	return bytes.HasPrefix(data, screenSnapshotMarker) && bytes.Contains(data, []byte("\x1b[2J"))
}

// isObserverScreenFrame reports PTY output that represents a terminal screen state
// (full redraw or alternate-screen entry) suitable for vt10x snapshot rendering.
func isObserverScreenFrame(data []byte) bool {
	return needsVTRender(data)
}

// needsVTRender reports scrollback or frames that must be replayed through vt10x
// instead of ANSI-stripped (codex uses ?2026h + absolute cursor positioning, not 2J).
func needsVTRender(data []byte) bool {
	return bytes.Contains(data, []byte("\x1b[2J")) ||
		bytes.Contains(data, []byte("\x1b[?1049h")) ||
		bytes.Contains(data, []byte("\x1b[?2026h")) ||
		bytes.HasPrefix(data, screenSnapshotMarker)
}

// renderSnapshotOutput converts a ptywrap screen snapshot frame and/or raw scrollback
// into printable snapshot text, matching the screen attach pipeline used by run.
func renderSnapshotOutput(frame, scrollback string, cols, rows int) string {
	if cols <= 0 {
		cols = 80
	}
	if rows <= 0 {
		rows = 24
	}
	var scrollbackData []byte
	if scrollback != "" {
		scrollbackData = []byte(scrollback)
		data := preprocessSnapshotScrollback(scrollbackData)
		useRows := adequateSnapshotRows(data, cols, rows)
		if text, ok := scrollbackToScreenText(data, cols, useRows); ok {
			out := mergePlainTextPrefix(strings.TrimRight(string(text), "\n"), scrollbackData)
			if !snapshotMissingPlainPrefix(out, scrollbackData) {
				return out
			}
		}
	}
	if frame != "" {
		frameData := []byte(frame)
		if isScreenSnapshotFrame(frameData) {
			infCols, infRows := inferSnapshotDimensions(frameData, cols, rows)
			if rendered, ok := screenSnapshotFrameToText(frameData, infCols, infRows); ok {
				out := mergePlainTextPrefix(strings.TrimRight(string(rendered), "\n"), scrollbackData)
				if scrollback == "" || !snapshotMissingPlainPrefix(out, scrollbackData) {
					return out
				}
			}
		}
	}
	raw := scrollback
	if raw == "" {
		raw = frame
	}
	out := renderSnapshotScrollback(raw, cols, rows)
	return mergePlainTextPrefix(out, scrollbackData)
}

func inferSnapshotDimensions(frame []byte, cols, rows int) (int, int) {
	if cols <= 0 {
		cols = 80
	}
	if rows <= 0 {
		rows = 24
	}
	maxRow, maxCol := rows, cols
	for _, m := range snapshotCUPRe.FindAllSubmatch(frame, -1) {
		if r, err := strconv.Atoi(string(m[1])); err == nil && r > maxRow {
			maxRow = r
		}
		if c, err := strconv.Atoi(string(m[2])); err == nil && c > maxCol {
			maxCol = c
		}
	}
	if maxRow < rows {
		maxRow = rows
	}
	if maxCol < cols {
		maxCol = cols
	}
	return maxCol, maxRow
}

func adequateSnapshotRows(data []byte, cols, rows int) int {
	if cols <= 0 {
		cols = 80
	}
	if rows <= 0 {
		rows = 24
	}
	best := rows
	for try := rows; try <= 80; try += 4 {
		if text, ok := scrollbackToScreenText(data, cols, try); ok {
			best = try
			out := mergePlainTextPrefix(strings.TrimRight(string(text), "\n"), data)
			if !snapshotMissingPlainPrefix(out, data) {
				return try
			}
		}
	}
	return best
}

func mergePlainTextPrefix(text string, scrollback []byte) string {
	prefix := plainTextPrefixBeforeAlt(scrollback)
	if prefix == "" || text == "" {
		return text
	}
	needle := prefix
	if i := strings.IndexByte(needle, '\n'); i >= 0 {
		needle = needle[:i]
	}
	needle = strings.TrimSpace(needle)
	if needle == "" || strings.Contains(text, needle) {
		return text
	}
	if len(needle) > 48 {
		if strings.Contains(text, needle[:48]) {
			return text
		}
	}
	return prefix + "\n" + text
}

func plainTextPrefixBeforeAlt(data []byte) string {
	idx := bytes.Index(data, []byte("\x1b[?2026h"))
	if idx <= 0 {
		idx = bytes.Index(data, []byte("\x1b[?1049h"))
	}
	if idx <= 0 {
		return ""
	}
	return strings.TrimSpace(SanitizeForPrint(string(data[:idx])))
}

func snapshotMissingPlainPrefix(text string, data []byte) bool {
	prefix := plainTextPrefixBeforeAlt(data)
	if prefix == "" {
		return false
	}
	firstLine := prefix
	if i := strings.IndexByte(prefix, '\n'); i >= 0 {
		firstLine = prefix[:i]
	}
	firstLine = strings.TrimSpace(firstLine)
	if firstLine == "" {
		return false
	}
	needle := firstLine
	if len(needle) > 48 {
		needle = needle[:48]
	}
	return !strings.Contains(text, needle)
}

// renderSnapshotScrollback converts accumulated scrollback to printable snapshot text.
// Alternate-screen redraws are replayed through vt10x so only the latest visible screen
// is emitted; plain line-oriented output falls back to SanitizeForPrint.
func renderSnapshotScrollback(raw string, cols, rows int) string {
	if raw == "" {
		return ""
	}
	data := preprocessSnapshotScrollback([]byte(raw))
	if text, ok := scrollbackToScreenText(data, cols, rows); ok {
		return strings.TrimRight(string(text), "\n")
	}
	return strings.TrimRight(SanitizeForPrint(raw), "\n")
}

// preprocessSnapshotScrollback normalizes codex-style scrollback before vt replay.
func preprocessSnapshotScrollback(data []byte) []byte {
	if !bytes.Contains(data, []byte("\x1b[?2026h")) {
		return data
	}
	spuriousPrefix := []byte("\x1b[?1049l\x1b[0m")
	for bytes.HasPrefix(data, spuriousPrefix) {
		data = data[len(spuriousPrefix):]
	}
	return injectClearOnAltScreen2026(data)
}

func injectClearOnAltScreen2026(data []byte) []byte {
	marker := []byte("\x1b[?2026h")
	var out bytes.Buffer
	for {
		idx := bytes.Index(data, marker)
		if idx == -1 {
			out.Write(data)
			break
		}
		out.Write(data[:idx+len(marker)])
		out.WriteString("\x1b[2J")
		data = data[idx+len(marker):]
	}
	return out.Bytes()
}

// scrollbackToScreenText replays accumulated scrollback like ptywrap screen attach,
// then extracts the final visible screen as plain text.
func scrollbackToScreenText(scrollback []byte, cols, rows int) ([]byte, bool) {
	if len(scrollback) == 0 || !needsVTRender(scrollback) {
		return nil, false
	}
	if snapshot, ok := renderScreenSnapshotFrame(scrollback, cols, rows); ok {
		if text, ok := screenSnapshotFrameToText(snapshot, cols, rows); ok {
			return text, true
		}
	}
	return screenSnapshotToText(scrollback, cols, rows)
}

// renderObserverFrame converts observer-mode PTY bytes to visible text without CSI/C0 leaks.
func renderObserverFrame(data []byte, cols, rows int) []byte {
	if isObserverScreenFrame(data) {
		if text, ok := screenSnapshotToText(data, cols, rows); ok {
			if shouldPrependSnapshotNewline(text) {
				return append([]byte{'\n'}, text...)
			}
			return text
		}
	}
	cleaned := SanitizeForPrint(string(data))
	if cleaned == "" {
		return nil
	}
	return []byte(cleaned)
}

func screenSnapshotToText(data []byte, cols, rows int) ([]byte, bool) {
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
	return renderVTStateToText(vt, cols, rows)
}

func renderVTStateToText(vt vt10x.Terminal, cols, rows int) ([]byte, bool) {
	if cols <= 0 {
		cols = 80
	}
	if rows <= 0 {
		rows = 24
	}

	vt.Lock()
	defer vt.Unlock()

	var lines []string
	for y := 0; y < rows; y++ {
		line := normalizeSnapshotPrintableLine(renderSnapshotTextLine(vt, cols, y))
		if line != "" {
			lines = append(lines, line)
		}
	}
	if len(lines) == 0 {
		debugLogf("renderVTStateToText no non-empty lines cols=%d rows=%d", cols, rows)
		return nil, false
	}
	text := []byte(strings.Join(lines, "\n") + "\n")
	debugLogf("renderVTStateToText lines=%v", lines)
	debugLogBytes("renderVTStateToText out", text)
	return text, true
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
	out.Grow(len(scrollback) / 2)
	out.WriteString("\x1b[?25l")
	if vt.Mode()&vt10x.ModeAltScreen != 0 {
		out.WriteString("\x1b[?1049h")
	} else {
		out.WriteString("\x1b[?1049l")
	}
	out.WriteString("\x1b[0m\x1b[H\x1b[2J")
	for y := 0; y < rows; y++ {
		line := renderSnapshotTextLine(vt, cols, y)
		if line == "" {
			continue
		}
		fmt.Fprintf(&out, "\x1b[%d;1H%s", y+1, line)
	}
	cursor := vt.Cursor()
	cursorX := clampSnapshot(cursor.X+1, 1, cols)
	cursorY := clampSnapshot(cursor.Y+1, 1, rows)
	fmt.Fprintf(&out, "\x1b[%d;%dH", cursorY, cursorX)
	if vt.CursorVisible() {
		out.WriteString("\x1b[?25h")
	}
	return []byte(out.String()), true
}

func clampSnapshot(v, min, max int) int {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

func screenSnapshotFrameToText(data []byte, cols, rows int) ([]byte, bool) {
	if !bytes.HasPrefix(data, screenSnapshotMarker) || !bytes.Contains(data, []byte("\x1b[2J")) {
		return nil, false
	}
	return screenSnapshotToText(data, cols, rows)
}

func normalizeSnapshotPrintableLine(line string) string {
	if line == "" {
		return ""
	}
	return strings.Map(func(r rune) rune {
		if r == '\u2500' { // box drawings light horizontal ─
			return '-'
		}
		return r
	}, line)
}

func renderSnapshotTextLine(vt vt10x.Terminal, cols, y int) string {
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
	return string(runes[:lastNonSpace+1])
}
