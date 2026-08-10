# Scenario

**Feature**: not-running, unmapped Codex session → import under agent-run + iTerm ForceNew

```
seed Codex provider UUID (no agent-run meta)
empty ListProcs hooks (not running)
  -> agent-run takeover --codex UUID
  -> exit 0
  -> creates agent-run session meta with runner=codex-tty, runner_session_id=UUID
  -> iTerm ModeForceNew script; follow-up includes agent-run open/resume path
  -> no kill log
```

## Preconditions

- Provider session exists; store has no mapping for UUID.
- Empty hooks = not running (no kill).
- iTerm hooks enabled for script capture.
- Parent only launches iTerm (does not attach in-process).

## Steps

1. Seed Codex only.
2. Empty hooks + iTerm capture.
3. Args = `takeover --codex <uuid>`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Mode = "handle"
	providerID := takeoverCodexFixtureSessionID
	ws := absPath(t, req.TempDir)
	seedCodexSession(t, takeoverCodexHome(req), ws, providerID)
	writeEmptyTakeoverHooks(t, req)
	applyIterm2TestHooks(req)
	req.Args = []string{
		"takeover",
		"--codex",
		providerID,
	}
	return nil
}
```
