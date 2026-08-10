# Scenario

**Feature**: takeover fails when the Grok provider session is missing under GROK_HOME

```
# GROK_HOME isolated and empty of sessions
agent-run takeover --grok 550e8400-e29b-41d4-a716-446655440abc
  -> exit ≠ 0
  -> not found / missing (not "not implemented" forever once action lands)
```

## Preconditions

- No `seedGrokSession`.
- Empty agent-run store.
- Empty takeover hooks (no live PIDs).

## Steps

1. Enable empty hooks + iTerm capture (negative: must not open).
2. Args = `takeover --grok <uuid>`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Mode = "handle"
	writeEmptyTakeoverHooks(t, req)
	applyIterm2TestHooks(req)
	req.Args = []string{
		"takeover",
		"--grok",
		takeoverFixtureSessionID,
	}
	return nil
}
```
