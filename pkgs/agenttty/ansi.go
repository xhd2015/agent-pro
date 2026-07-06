package agenttty

import (
	"regexp"
	"strings"
)

var ansiEscapeRe = regexp.MustCompile(`\x1b\[[0-9;?=>]*[ -/]*[@-~]|\x1b\][^\x07\x1b]*(?:\x07|\x1b\\)|\x1b[78]|\x1b[@-_]|\x1b[PX^_][^\x1b]*\x1b\\`)

// StripANSI removes ANSI escape sequences from terminal scrollback.
func StripANSI(data []byte) string {
	s := string(data)
	s = ansiEscapeRe.ReplaceAllString(s, "")
	s = strings.ReplaceAll(s, "\r", "")
	return s
}

func stripPlain(scrollback []byte) string {
	return StripANSI(scrollback)
}