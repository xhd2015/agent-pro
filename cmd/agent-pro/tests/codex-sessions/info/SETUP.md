# Scenario

**Feature**: session info for a single Codex rollout (replaces CLI brief view)

```
# locate rollout by full UUID
sessions.Find -> sessions.Info(codexHome, sessionID, lastN)

# emit metadata, status, line count, recent messages, rollout path, token totals
SessionInfo -> FormatInfoText(now) -> key-value output
```

## Preconditions

- This branch tests the `info` operation (`agent-pro codex session info <id>`).
- Session ID must be a full UUID match.

## Steps

1. Set `req.Operation = "info"`.
2. Leaf Setup writes a rollout fixture for `req.SessionID`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Operation = "info"
	if req.LastN == 0 {
		req.LastN = 3
	}
	return nil
}
```