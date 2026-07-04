# Scenario

**Feature**: info without token_count events omits Tokens section

```
# rollout with messages but no token_count events
writeRolloutSession -> sessions.Info -> FormatInfoText(now)

# info succeeds; output has no Tokens: block
SessionInfo without token aggregates
```

## Preconditions

- Displayable messages exist.
- No `token_count` events in the rollout.

## Steps

1. Write a minimal session with user and agent messages only.
2. Set `req.SessionID` to the fixture UUID.

```go
import "testing"

const noTokensSessionID = "01900009-ffff-7fff-ffff-ffffffffffff"

func Setup(t *testing.T, req *Request) error {
	req.SessionID = noTokensSessionID
	lines := []string{
		userMessageLine("Docs cleanup"),
		agentMessageLine("Cleaning docs"),
	}
	writeRolloutSession(t, req.CodexHome, noTokensSessionID,
		"2026-07-03T14:00:00.000Z", "/tmp/codex-no-tokens", lines...)
	return nil
}
```