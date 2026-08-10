# Scenario

**Feature**: not-running but already mapped Codex → resume existing agent-run session in iTerm

```
seed Codex provider UUID
seed agent-run meta finished: runner=codex-tty, runner_session_id=UUID (mapped, dead)
empty hooks (not running)
  -> agent-run takeover --codex UUID
  -> exit 0
  -> reuses existing agent-run session id (no second-import / already-mapped error)
  -> iTerm ForceNew follow-up resume/open of existing id
  -> no kill log
```

## Preconditions

- Mapping exists; status not live (finished / idle).
- No live registry entry.
- Takeover **reuses** the mapping (does not error as already-mapped).

## Steps

1. Seed Codex + mapped finished meta with known agent session id.
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
	agentID := "takeover-codex-mapped-resume-s1"
	ws := absPath(t, req.TempDir)
	seedCodexSession(t, takeoverCodexHome(req), ws, providerID)
	seedMappedMeta(t, req, "codex-tty", agentID, providerID, "finished")
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
