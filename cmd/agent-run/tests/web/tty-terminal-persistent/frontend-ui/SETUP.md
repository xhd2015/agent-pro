# Scenario

**Feature**: browser UI for web-created grok-tty sessions and fixture-backed terminal modal

```
grok mock web harness -> POST grok-tty -> terminal_session_id + grok-tty-registry
browser chat -> terminal icon -> modal attaches live GROK_TTY_BANNER PTY
fixture leaves (modal-renders-*, finished-status-*) keep seeded ptywrap only
```

## Preconditions

- `playwright-debug` on PATH for `--label ui-automation` leaves.
- Web-created session leaves use root `createWebGrokTTYSessionThroughAPI` /
  `createRunningWebGrokTTYSessionThroughAPI` helpers (no fake `codex` shell binary).
- `real-codex-terminal-stale-input-follow-up` stays on real `codex-tty` (`label: codex`).

## Steps

1. Grouping setup sets `req.Mode = "ui"`.
2. Leaf setup creates grok-tty web session or writes fixture session/registry.
3. Leaf setup writes Playwright script.
4. `Run` executes `playwright-debug run`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Mode = "ui"
	return nil
}
```