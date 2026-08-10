# Scenario

**Feature**: not-running, unmapped Grok session → import under agent-run + iTerm ForceNew

```
seed Grok provider UUID (no agent-run meta)
empty ListProcs hooks (not running)
  -> agent-run takeover --grok UUID
  -> exit 0
  -> creates agent-run session meta with runner_session_id=UUID
  -> iTerm ModeForceNew script; follow-up includes agent-run open/resume/import path
  -> no kill log
```

## Preconditions

- Provider session exists; store has no mapping for UUID.
- Empty hooks = not running (no kill).
- iTerm hooks enabled for script capture.
- Parent only launches iTerm (does not attach in-process).

## Steps

1. Seed Grok only.
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
	ws := absPath(t, req.TempDir)
	seedGrokSession(t, takeoverGrokHome(req), ws, providerID)
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
