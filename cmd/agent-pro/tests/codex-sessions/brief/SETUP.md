# Scenario

**Feature**: brief summary for a single Codex session

```
# locate rollout file by full UUID
sessions.Find -> sessions.Brief(codexHome, sessionID, lastN)

# emit metadata, status, line count, last N displayable messages
SessionBrief -> FormatBriefText / FormatBriefJSON -> output
```

## Preconditions

- This branch tests the `brief` operation (no `--log`).
- Session ID must be a full UUID match.

## Steps

1. Set `req.Operation = "brief"`.
2. Leaf Setup writes a rollout fixture for `req.SessionID`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Operation = "brief"
	if req.LastN == 0 {
		req.LastN = 3
	}
	return nil
}
```