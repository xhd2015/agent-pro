package main

import (
	"bytes"
	"strings"

	"github.com/hinshun/vt10x"
)

var screenSnapshotMarker = []byte("\x1b[?25l")

func isScreenSnapshotFrame(data []byte) bool {
	return bytes.HasPrefix(data, screenSnapshotMarker) && bytes.Contains(data, []byte("\x1b[2J"))
}

// isObserverScreenFrame reports PTY output that represents a terminal screen state
// (full redraw or alternate-screen entry) suitable for vt10x snapshot rendering.
func isObserverScreenFrame(data []byte) bool {
	return bytes.Contains(data, []byte("\x1b[2J")) ||
		bytes.Contains(data, []byte("\x1b[?1049h")) ||
		bytes.HasPrefix(data, screenSnapshotMarker)
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
		line := renderSnapshotTextLine(vt, cols, y)
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
	firstNonSpace := 0
	for firstNonSpace <= lastNonSpace && (runes[firstNonSpace] == ' ' || runes[firstNonSpace] == '\t') {
		firstNonSpace++
	}
	return string(runes[firstNonSpace : lastNonSpace+1])
}
