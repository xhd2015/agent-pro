package agenttty

import "github.com/xhd2015/agent-pro/pkgs/utf8util"

// normalizeRunnerPrompt prepares text for runner delivery (new-session argv
// PROMPT and inject paths). Returns valid UTF-8: Rust grok panics in
// std::env::args() when any argv element has invalid UTF-8 (seatalk local-bot
// agent-run incident: truncated multi-byte sequences on open inject).
func normalizeRunnerPrompt(s string) string {
	return utf8util.ToValid(s)
}
