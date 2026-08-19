package agenttty

import "strings"

// appendNewSessionPrompt attaches the initial prompt for a new TTY session
// (no --resume). Mirrors the RunHeadless argv rules:
//
//   - NoSubmit: never put draft on argv (real Grok auto-submits positional PROMPT).
//   - codex-tty: never put prompt on argv (inject after banner).
//   - commandcode-tty headless (!open): -p <prompt>.
//   - grok-tty (and others): trailing positional PROMPT.
//
// Callers must pass a non-empty ResumeSessionID check before calling (only for
// new sessions). prompt is TrimSpace'd here.
//
// Contract (enforced by tests): every string appended for runner delivery must
// be valid UTF-8. Rust runners (grok) panic in std::env::args() on invalid UTF-8
// (see seatalk local-bot agent-run incident: env.rs unwrap on OsString).
func appendNewSessionPrompt(argv []string, runnerID, prompt string, noSubmit, open bool) []string {
	if noSubmit {
		return argv
	}
	p := prepareRunnerPrompt(runnerID, strings.TrimSpace(prompt))
	if p == "" {
		return argv
	}
	// commandcode-tty: headless uses -p. Open omits argv prompt and injects
	// after ready (positional "cmd Hello" one-shots and exits before attach).
	if runnerID == "commandcode-tty" {
		if open {
			return argv
		}
		return append(argv, "-p", p)
	}
	if runnerID == "codex-tty" {
		return argv
	}
	return append(argv, p)
}
