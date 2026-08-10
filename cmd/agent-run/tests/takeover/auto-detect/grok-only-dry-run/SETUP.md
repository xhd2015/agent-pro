# Scenario

**Feature**: auto-detect Grok with `--dry-run` plans import/open without side effects

```
seed Grok UUID only under GROK_HOME
  -> agent-run takeover --dry-run UUID   # no --grok
  -> exit 0
  -> dry-run plan mentions open iTerm (and not empty-runner flag error)
  -> no kill log; no iTerm script file; no meta create
```

## Preconditions

- Optional P4 leaf: dry-run + auto-detect.
- Not running; unmapped.

## Steps

1. Seed Grok only.
2. Empty hooks + iTerm capture (must remain empty).
3. Args = `takeover --dry-run <uuid>`.

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
		"--dry-run",
		providerID,
	}
	return nil
}
```
