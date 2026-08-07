package agenttty

import (
	"regexp"
	"strings"
)

// codexResumeCmdRe matches "codex resume" or "codex --resume" (optional id).
// Used only together with the "to continue this session" phrase (conservative AND).
var codexResumeCmdRe = regexp.MustCompile(`(?i)\bcodex\s+(?:--)?resume\b`)

// TerminalExitedMarkerPresent reports the shared keep-alive / shell footer.
func TerminalExitedMarkerPresent(scrollback string) bool {
	return strings.Contains(scrollback, "[Terminal exited]")
}

// CodexExitFooterPresent reports a real Codex /exit footer.
//
// Conservative AND (not two standalone cases): both must appear in scrollback:
//  1. phrase "to continue this session"
//  2. "codex resume" or "codex --resume"
//
// Real Codex (e.g. v0.147): "To continue this session, run codex resume <uuid>"
func CodexExitFooterPresent(scrollback string) bool {
	if strings.TrimSpace(scrollback) == "" {
		return false
	}
	lower := strings.ToLower(scrollback)
	if !strings.Contains(lower, "to continue this session") {
		return false
	}
	return codexResumeCmdRe.MatchString(scrollback)
}

// GrokExitFooterPresent reports Grok post-/exit resume footers (grok-tty only).
func GrokExitFooterPresent(scrollback string) bool {
	if strings.TrimSpace(scrollback) == "" {
		return false
	}
	lower := strings.ToLower(scrollback)
	if strings.Contains(lower, "grok --resume") || strings.Contains(lower, "grok resume") {
		return true
	}
	if strings.Contains(lower, "resume this session with") {
		return true
	}
	return false
}

// isCodexRunnerID reports codex-tty / codex provider ids.
func isCodexRunnerID(runner string) bool {
	r := strings.ToLower(strings.TrimSpace(runner))
	return r == "codex-tty" || r == "codex" || strings.HasPrefix(r, "codex")
}

// isGrokRunnerID reports grok-tty / grok provider ids.
func isGrokRunnerID(runner string) bool {
	r := strings.ToLower(strings.TrimSpace(runner))
	return r == "grok-tty" || r == "grok" || strings.HasPrefix(r, "grok")
}

// ScrollbackSuggestsAgentExited reports post-/exit markers for Classify.
//
// Shared: [Terminal exited]
// codex-tty: CodexExitFooterPresent (phrase AND resume cmd)
// grok-tty: GrokExitFooterPresent
// other: shared marker only
func ScrollbackSuggestsAgentExited(scrollback, runnerID string) bool {
	if strings.TrimSpace(scrollback) == "" {
		return false
	}
	if TerminalExitedMarkerPresent(scrollback) {
		return true
	}
	switch {
	case isCodexRunnerID(runnerID):
		return CodexExitFooterPresent(scrollback)
	case isGrokRunnerID(runnerID):
		return GrokExitFooterPresent(scrollback)
	default:
		// Unknown runner: do not apply agent-specific footers.
		return false
	}
}
