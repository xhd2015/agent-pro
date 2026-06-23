# Scenario

**Feature**: log skips non-displayable rollout noise events

```
# rollout with session_meta, token_count, turn_context only
writeRolloutSession -> sessions.PrintLog

# no compact trace lines emitted
empty or whitespace-only output
```

## Preconditions

- `session_meta`, `token_count`, and `turn_context` produce no log output.

## Steps

1. Set session id `01900012-3333-7333-8333-333333333333`.
2. Write only skipped event types after session_meta.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.SessionID = "01900012-3333-7333-8333-333333333333"
	lines := []string{
		`{"type":"event_msg","payload":{"type":"token_count","input_tokens":10,"output_tokens":5}}`,
		`{"type":"turn_context","payload":{"cwd":"/tmp","approval_policy":"never"}}`,
	}
	writeRolloutSession(t, req.CodexHome, req.SessionID,
		"2026-06-23T17:00:00.000Z", "/tmp/log-noise", lines...)
	return nil
}
```