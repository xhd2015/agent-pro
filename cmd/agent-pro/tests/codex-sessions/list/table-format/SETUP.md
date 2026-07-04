# Scenario

**Feature**: list table output shows unified grok-shaped columns and relative times

```
# three rollout sessions with timestamp offsets from fixed now
writeRolloutSession x3 -> sessions.List -> FormatListTable(now)

# table includes SESSION ID, LAST ACTIVE, TITLE, MSGS, CWD with relative deltas
terminal table text
```

## Preconditions

- `req.Now` is fixed at `2026-07-03T15:00:00.000Z` by root Setup.
- Each session includes a `user_message` for TITLE derivation.

## Steps

1. Create session A at `req.Now` → `just now`.
2. Create session B five minutes before `req.Now` → `5m ago`.
3. Create session C two hours before `req.Now` → `2h ago`.

```go
import (
	"testing"
	"time"
)

func Setup(t *testing.T, req *Request) error {
	req.Limit = 10
	now := req.Now.UTC()

	writeRolloutSession(t, req.CodexHome,
		"01900004-aaaa-7aaa-aaaa-aaaaaaaaaaaa",
		now.Format("2006-01-02T15:04:05.000Z"), "/tmp/project-a",
		userMessageLine("Alpha refactor task"),
		agentMessageLine("working on alpha"))
	writeRolloutSession(t, req.CodexHome,
		"01900004-bbbb-7bbb-bbbb-bbbbbbbbbbbb",
		now.Add(-5*time.Minute).Format("2006-01-02T15:04:05.000Z"), "/tmp/project-b",
		userMessageLine("Beta bugfix task"),
		agentMessageLine("working on beta"))
	writeRolloutSession(t, req.CodexHome,
		"01900004-cccc-7ccc-cccc-cccccccccccc",
		now.Add(-2*time.Hour).Format("2006-01-02T15:04:05.000Z"), "/tmp/project-c",
		userMessageLine("Gamma cleanup task"),
		agentMessageLine("working on gamma"))
	return nil
}
```