# Scenario

**Feature**: takeover --codex fails when the Codex provider session is missing under CODEX_HOME

```
# CODEX_HOME isolated and empty of rollouts
agent-run takeover --codex 660e8400-e29b-41d4-a716-446655440c0d
  -> exit ≠ 0
  -> not found / missing (not forever "codex support is not implemented yet")
```

## Preconditions

- No `seedCodexSession`.
- Empty agent-run store.
- Empty takeover hooks.

## Steps

1. Enable empty hooks + iTerm capture (negative: must not open).
2. Args = `takeover --codex <uuid>`.

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
		"--codex",
		takeoverCodexFixtureSessionID,
	}
	return nil
}
```
