package agenttty

import (
	"os"
	"strings"
	"time"

	"github.com/xhd2015/tty-watch/pkgs/ttywatch"
)

// CodexSubmitSettle is the pause between typing and a trailing Enter for codex-tty.
// Real Codex TUI often treats one-shot "text\r" as typing only.
const CodexSubmitSettle = 350 * time.Millisecond

// InjectMessage writes text into a live TTY session.
// When submit is false, bytes are typed without trailing Enter.
// When submit is true:
//   - codex-tty (real UI): type text, settle, bare \r (and \n for line discipline)
//   - codex-tty (fake TUI via AGENT_RUN_CODEX_TTY_COMMAND): text+\n+\r for shell `read`
//   - other runners: SendMessage with suffixCR=true (existing contract)
func InjectMessage(listenAddr, sessionID, runner, text string, submit bool) error {
	runner = strings.TrimSpace(runner)
	text = SanitizePromptForRunner(runner, text)
	if !submit {
		return injectSendRetry(listenAddr, sessionID, text, false)
	}
	if runner == "codex-tty" {
		// Fake interactive shell hooks used in doctests: one-shot text+\n+\r is enough
		// for `read` and avoids mid-turn inject races on short-lived headless serves.
		if strings.TrimSpace(os.Getenv(envCodexTTYCommand)) != "" {
			payload := text
			if payload != "" && !strings.HasSuffix(payload, "\n") {
				payload += "\n"
			}
			return injectSendRetry(listenAddr, sessionID, payload, true)
		}
		// Real Codex TUI: one-shot text+\r only fills the composer. Type, pause, Enter.
		if text != "" {
			if err := injectSendRetry(listenAddr, sessionID, text, false); err != nil {
				return err
			}
			time.Sleep(CodexSubmitSettle)
		}
		if err := injectSendRetry(listenAddr, sessionID, "", true); err != nil {
			return err
		}
		time.Sleep(200 * time.Millisecond)
		_ = ttywatch.SendMessage(listenAddr, sessionID, "", true)
		return nil
	}
	return injectSendRetry(listenAddr, sessionID, text, true)
}

// injectSendRetry retries briefly when the serve/PTY is not yet injectable.
// ptywrap maps "session exited" to HTTP 404 → "inject endpoint not found".
func injectSendRetry(listenAddr, sessionID, message string, suffixCR bool) error {
	var last error
	for i := 0; i < 8; i++ {
		last = ttywatch.SendMessage(listenAddr, sessionID, message, suffixCR)
		if last == nil {
			return nil
		}
		msg := last.Error()
		if !strings.Contains(msg, "endpoint not found") && !strings.Contains(msg, "session exited") &&
			!strings.Contains(msg, "session not found") && !strings.Contains(msg, "connection refused") {
			return last
		}
		time.Sleep(50 * time.Millisecond)
	}
	return last
}

// injectSessionGone reports whether inject failed because the PTY session is
// already gone (KeepAlive may still keep the HTTP serve up for scrollback).
func injectSessionGone(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "endpoint not found") ||
		strings.Contains(msg, "session exited") ||
		strings.Contains(msg, "session not found")
}

