package agenttty

import (
	"strings"

	"github.com/xhd2015/agent-pro/pkgs/utf8util"
)

// SanitizePromptForRunner rewrites a TTY inject / argv prompt for runner.
// Always: CR LF and CR → LF (CR is a submit key on Codex TUI).
// Codex only (runner "codex-tty" or "codex", case-trim): TAB → four spaces
// (Tab submits or queues a follow-up). Empty runner is not defaulted.
func SanitizePromptForRunner(runner, text string) string {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	if isCodexPromptRunner(runner) {
		text = strings.ReplaceAll(text, "\t", "    ")
	}
	return text
}

func isCodexPromptRunner(runner string) bool {
	switch strings.ToLower(strings.TrimSpace(runner)) {
	case "codex-tty", "codex":
		return true
	default:
		return false
	}
}

// normalizeRunnerPrompt prepares text for runner delivery (new-session argv
// PROMPT and inject paths). Returns valid UTF-8: Rust grok panics in
// std::env::args() when any argv element has invalid UTF-8 (seatalk local-bot
// agent-run incident: truncated multi-byte sequences on open inject).
func normalizeRunnerPrompt(s string) string {
	return utf8util.ToValid(s)
}

// prepareRunnerPrompt is normalizeRunnerPrompt after runner-aware CR/Tab rewrite.
func prepareRunnerPrompt(runner, s string) string {
	return normalizeRunnerPrompt(SanitizePromptForRunner(runner, s))
}
