# Scenario

**Feature**: log redacts encrypted-only reasoning content

```
# rollout with reasoning item containing only encrypted_content
writeRolloutSession -> sessions.PrintLog

# REASONING block shows [Redacted] instead of ciphertext
terminal log with REASONING [Redacted]
```

## Preconditions

- Reasoning event has `encrypted_content` and no plaintext summary field.

## Steps

1. Set session id `01900011-2222-7222-8222-222222222222`.
2. Write reasoning response_item with encrypted_content only.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.SessionID = "01900011-2222-7222-8222-222222222222"
	reasonLine := `{"type":"response_item","payload":{"type":"reasoning","call_id":"call_reason_1","encrypted_content":"gAAAAABsecretblob"}}`
	writeRolloutSession(t, req.CodexHome, req.SessionID,
		"2026-06-23T16:00:00.000Z", "/tmp/log-reason", reasonLine)
	return nil
}
```