package subagent

import (
	"fmt"
	"io"
)

const (
	// TraceHeaderRule separates trace/session status header blocks.
	TraceHeaderRule = "==================================================================="
	// TraceFooterRule separates trace footer messages.
	TraceFooterRule = "-------------------------------------------------------------------"
	// EmptyDisplay is shown when a trace/status field has no value.
	EmptyDisplay = "-"
)

// FprintTraceHeader writes the standard session/events trace header to w.
func FprintTraceHeader(w io.Writer, sessionID string, eventCount int) {
	fmt.Fprintf(w, "\n%s\n", TraceHeaderRule)
	fmt.Fprintf(w, "  Session: %s\n", sessionID)
	fmt.Fprintf(w, "  Events:  %d lines\n", eventCount)
	fmt.Fprintf(w, "%s\n\n", TraceHeaderRule)
}

// FprintTraceFooterFrame writes the footer rule line before and after a message.
func FprintTraceFooterFrame(w io.Writer) {
	fmt.Fprintf(w, "\n%s\n", TraceFooterRule)
}

// FprintlnTraceFooterRule closes a trace footer block.
func FprintlnTraceFooterRule(w io.Writer) {
	fmt.Fprintln(w, TraceFooterRule)
}