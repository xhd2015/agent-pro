package agenttty

import "strings"

// normalizeRunnerPrompt prepares text for runner delivery (new-session argv
// PROMPT and inject paths). Returns valid UTF-8: Rust grok panics in
// std::env::args() when any argv element has invalid UTF-8 (seatalk local-bot
// agent-run incident: truncated multi-byte sequences on open inject).
//
// Invalid sequences are replaced with U+FFFD so surrounding text is preserved.
func normalizeRunnerPrompt(s string) string {
	return strings.ToValidUTF8(s, "\uFFFD")
}
