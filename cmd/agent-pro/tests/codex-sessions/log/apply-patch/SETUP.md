# Scenario

**Feature**: log prints EDIT for apply_patch custom_tool_call

```
# rollout with custom_tool_call name=apply_patch
writeRolloutSession -> sessions.PrintLog

# compact trace shows EDIT block for patch application
terminal log with EDIT label
```

## Preconditions

- `response_item.custom_tool_call` with `name=apply_patch` maps to EDIT output.

## Steps

1. Set session id `01900010-1111-7111-8111-111111111111`.
2. Write apply_patch custom_tool_call event.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.SessionID = "01900010-1111-7111-8111-111111111111"
	patchLine := `{"type":"response_item","payload":{"type":"custom_tool_call","name":"apply_patch","call_id":"call_patch_1","input":"{\"path\":\"src/main.go\",\"patch\":\"+added line\"}"}}`
	writeRolloutSession(t, req.CodexHome, req.SessionID,
		"2026-06-23T15:00:00.000Z", "/tmp/log-patch", patchLine)
	return nil
}
```