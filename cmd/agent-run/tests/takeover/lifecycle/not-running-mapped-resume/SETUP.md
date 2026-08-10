# Scenario

**Feature**: not-running but already mapped → resume existing agent-run session in iTerm

```
seed Grok provider UUID
seed agent-run meta finished: runner=grok-tty, runner_session_id=UUID (mapped, dead)
empty hooks (not running)
  -> agent-run takeover --grok UUID
  -> exit 0
  -> reuses existing agent-run session id (no second-import / already-mapped error)
  -> iTerm ForceNew follow-up resume/open of existing id
  -> no kill log
```

## Preconditions

- Mapping exists; status not live (finished / idle).
- No live registry entry.
- Distinct from `run --resume-from-grok-session` "already mapped" **error** —
  takeover **reuses** the mapping.

## Steps

1. Seed Grok + mapped finished meta with known agent session id.
2. Empty hooks + iTerm capture.
3. Args = `takeover --grok <uuid>`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Mode = "handle"
	providerID := takeoverFixtureSessionID
	agentID := "takeover-mapped-resume-s1"
	ws := absPath(t, req.TempDir)
	seedGrokSession(t, takeoverGrokHome(req), ws, providerID)
	seedMappedMeta(t, req, "grok-tty", agentID, providerID, "finished")
	writeEmptyTakeoverHooks(t, req)
	applyIterm2TestHooks(req)
	req.Args = []string{
		"takeover",
		"--grok",
		providerID,
	}
	return nil
}
```
