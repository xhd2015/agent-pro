# Scenario

**Feature**: same UUID present under both GROK_HOME and CODEX_HOME is ambiguous

```
seed shared UUID under GROK_HOME and CODEX_HOME
  -> agent-run takeover SHARED-UUID   # no runner flags
  -> exit ≠ 0
  -> error: ambiguous / both providers (mention grok and codex)
  -> no kill; no iTerm; no meta create
```

## Preconditions

- Both provider homes contain the same id (`takeoverAutoDetectSharedID`).
- No agent-run mapping; empty hooks.

## Steps

1. Seed Grok + Codex with the shared UUID.
2. Empty hooks + iTerm capture (must stay empty).
3. Args = `takeover <shared-uuid>`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Mode = "handle"
	providerID := takeoverAutoDetectSharedID
	ws := absPath(t, req.TempDir)
	seedGrokSession(t, takeoverGrokHome(req), ws, providerID)
	seedCodexSession(t, takeoverCodexHome(req), ws, providerID)
	writeEmptyTakeoverHooks(t, req)
	applyIterm2TestHooks(req)
	req.Args = []string{
		"takeover",
		providerID,
	}
	return nil
}
```
