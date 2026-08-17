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
		return ttywatch.SendMessage(listenAddr, sessionID, text, false)
	}
	if runner == "codex-tty" {
		// Fake interactive shell hooks used in doctests: one-shot text+\n+\r is enough
		// for `read` and avoids mid-turn inject races on short-lived headless serves.
		if strings.TrimSpace(os.Getenv(envCodexTTYCommand)) != "" {
			payload := text
			if payload != "" && !strings.HasSuffix(payload, "\n") {
				payload += "\n"
			}
			return ttywatch.SendMessage(listenAddr, sessionID, payload, true)
		}
		// Real Codex TUI: one-shot text+\r only fills the composer. Type, pause, Enter.
		if text != "" {
			if err := ttywatch.SendMessage(listenAddr, sessionID, text, false); err != nil {
				return err
			}
			time.Sleep(CodexSubmitSettle)
		}
		if err := ttywatch.SendMessage(listenAddr, sessionID, "", true); err != nil {
			return err
		}
		time.Sleep(200 * time.Millisecond)
		_ = ttywatch.SendMessage(listenAddr, sessionID, "", true)
		return nil
	}
	return ttywatch.SendMessage(listenAddr, sessionID, text, true)
}
