package agenttty

import (
	"fmt"
	"strings"

	"github.com/xhd2015/tty-watch/pkgs/ttywatch"
)

// formatOpenAgentExitedError builds a user-facing error when the PTY agent
// dies before --open attach. Best-effort: pulls scrollback for the real
// provider error and appends actionable hints (e.g. corrupt Codex SQLite).
func formatOpenAgentExitedError(runnerID, listenAddr, sessionID, phase string, cause error) error {
	runnerID = strings.TrimSpace(runnerID)
	if runnerID == "" {
		runnerID = "agent"
	}
	phase = strings.TrimSpace(phase)
	if phase == "" {
		phase = "open attach"
	}

	scrollback := ""
	if strings.TrimSpace(listenAddr) != "" && strings.TrimSpace(sessionID) != "" {
		if text, err := ttywatch.SnapshotText(listenAddr, sessionID); err == nil {
			scrollback = text
		}
	}
	providerMsg := extractProviderExitMessage(scrollback)
	hint := suggestFromProviderExit(runnerID, providerMsg, scrollback)

	var b strings.Builder
	fmt.Fprintf(&b, "%s: agent exited before %s", runnerID, phase)
	if cause != nil && strings.TrimSpace(cause.Error()) != "" {
		fmt.Fprintf(&b, " (%v)", cause)
	}
	if providerMsg != "" {
		fmt.Fprintf(&b, "\n  provider: %s", providerMsg)
	}
	if hint != "" {
		fmt.Fprintf(&b, "\n  hint: %s", hint)
	}
	if providerMsg == "" && hint == "" {
		b.WriteString("\n  hint: re-run with the same --session-id and inspect TTY scrollback; for codex-tty also try: codex resume <runner_session_id>")
	}
	return fmt.Errorf("%s", b.String())
}

// extractProviderExitMessage picks the most useful error line from scrollback.
func extractProviderExitMessage(scrollback string) string {
	plain := stripPlain([]byte(scrollback))
	if strings.TrimSpace(plain) == "" {
		return ""
	}
	lines := strings.Split(plain, "\n")
	var best string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		lower := strings.ToLower(line)
		switch {
		case strings.HasPrefix(lower, "error:"):
			best = line
		case strings.Contains(lower, "database disk image is malformed"):
			best = line
		case strings.Contains(lower, "failed to resume session"):
			if best == "" || !strings.HasPrefix(strings.ToLower(best), "error:") {
				best = line
			}
		case strings.Contains(lower, "thread-store") && strings.Contains(lower, "error"):
			if best == "" {
				best = line
			}
		}
	}
	// Collapse whitespace for multi-line paste noise.
	best = strings.Join(strings.Fields(best), " ")
	if len(best) > 400 {
		best = best[:400] + "…"
	}
	return best
}

// suggestFromProviderExit returns a short operator fix for known failures.
func suggestFromProviderExit(runnerID, providerMsg, scrollback string) string {
	blob := strings.ToLower(providerMsg + "\n" + scrollback)
	isCodex := isCodexRunnerID(runnerID) || strings.Contains(blob, "codex")

	if strings.Contains(blob, "database disk image is malformed") ||
		(strings.Contains(blob, "thread-store") && strings.Contains(blob, "malformed")) ||
		(strings.Contains(blob, "failed to read thread metadata") && strings.Contains(blob, "malformed")) {
		if isCodex {
			return "Codex local DB is corrupt (usually ~/.codex/state_5.sqlite). " +
				"Backup then reset: " +
				"mv ~/.codex/state_5.sqlite ~/.codex/state_5.sqlite.bak.$(date +%Y%m%d%H%M%S) " +
				"(also move state_5.sqlite-wal/-shm if present), then retry. " +
				"Old threads may be lost; rollout jsonl under ~/.codex/sessions may still exist."
		}
		return "provider local database is corrupted; repair or reset its on-disk store, then retry"
	}
	if strings.Contains(blob, "failed to resume session") || strings.Contains(blob, "error resuming thread") {
		if isCodex {
			return "codex resume failed for this runner_session_id. " +
				"Try: codex resume <id> in a normal terminal to see the full error; " +
				"or start a new agent-run session if the thread is unrecoverable."
		}
		return "provider failed to resume the bound session; verify the runner session id still exists"
	}
	if strings.Contains(blob, "stdin is not a terminal") {
		return "provider refused non-TTY stdin (unexpected under agent-run PTY); report as agent-run bug with this log"
	}
	return ""
}
