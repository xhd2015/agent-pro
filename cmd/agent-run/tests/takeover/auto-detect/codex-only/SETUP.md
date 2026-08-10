# Scenario

**Feature**: auto-detect chooses Codex when only CODEX_HOME has the UUID

```
seed Codex UUID under CODEX_HOME; GROK_HOME empty of that id
empty hooks (not running)
  -> agent-run takeover UUID   # no runner flags
  -> exit 0
  -> import meta runner=codex-tty, runner_session_id=UUID
  -> iTerm ForceNew; no kill log
```

## Preconditions

- Only Codex has the provider session (default Codex fixture UUID).
- Unmapped store; not running.

## Steps

1. Seed Codex only.
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
	providerID := takeoverCodexFixtureSessionID
	ws := absPath(t, req.TempDir)
	seedCodexSession(t, takeoverCodexHome(req), ws, providerID)
	// GROK_HOME stays empty of this UUID.
	writeEmptyTakeoverHooks(t, req)
	applyIterm2TestHooks(req)
	req.Args = []string{
		"takeover",
		providerID,
	}
	return nil
}
```
