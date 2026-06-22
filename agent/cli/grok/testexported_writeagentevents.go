package grok

import "io"

// TestExported_writeAgentEventsFromGrokLine wraps writeAgentEventsFromGrokLine for doctest access.
func TestExported_writeAgentEventsFromGrokLine(rawLog io.Writer, line string) {
	writeAgentEventsFromGrokLine(rawLog, line)
}