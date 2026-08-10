# Scenario

**Feature**: auto-detect chooses Grok when only GROK_HOME has the UUID

```
seed Grok UUID under GROK_HOME; CODEX_HOME empty of that id
empty hooks (not running)
  -> agent-run takeover UUID   # no --grok/--codex/--agent-runner
  -> exit 0
  -> import meta runner=grok-tty, runner_session_id=UUID
  -> iTerm ForceNew; no kill log
```

## Preconditions

- Only Grok has the provider session.
- Unmapped agent-run store; not running.

## Steps

1. Seed Grok only (default Grok fixture UUID).
2. Empty hooks + iTerm capture.
3. Args = `takeover <uuid>` (no runner flags).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Mode = "handle"
	providerID := takeoverFixtureSessionID
	ws := absPath(t, req.TempDir)
	seedGrokSession(t, takeoverGrokHome(req), ws, providerID)
	// CODEX_HOME stays empty of this UUID (root only mkdir).
	writeEmptyTakeoverHooks(t, req)
	applyIterm2TestHooks(req)
	req.Args = []string{
		"takeover",
		providerID,
	}
	return nil
}
```
