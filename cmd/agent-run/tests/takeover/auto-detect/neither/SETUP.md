# Scenario

**Feature**: auto-detect fails when neither GROK_HOME nor CODEX_HOME has the UUID

```
# both provider homes empty of the id
agent-run takeover 880e8400-e29b-41d4-a716-446655440000
  -> exit ≠ 0
  -> not found / cannot resolve provider
  -> no kill; no iTerm; no meta
```

## Preconditions

- No Grok/Codex seed for the requested UUID.
- Empty hooks.

## Steps

1. Do not seed provider sessions.
2. Empty hooks + iTerm capture (negative).
3. Args = `takeover <unknown-uuid>`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

const takeoverAutoDetectMissingID = "880e8400-e29b-41d4-a716-446655440000"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Mode = "handle"
	writeEmptyTakeoverHooks(t, req)
	applyIterm2TestHooks(req)
	req.Args = []string{
		"takeover",
		takeoverAutoDetectMissingID,
	}
	return nil
}
```
