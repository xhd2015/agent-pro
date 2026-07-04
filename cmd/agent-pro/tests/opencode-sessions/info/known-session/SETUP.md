# Scenario

**Feature**: info returns session fields, message paths, and summed tokens/cost

```
# session JSON + three message files with tokens and cost
writeOpencodeSession + writeOpencodeMessage x3 -> sessions.Info -> FormatInfoText(now)

# output includes metadata, Messages block, Files block, Tokens/Cost block
terminal key-value text
```

## Preconditions

- Three message JSON files exist under `storage/message/{sessionID}/`.
- `req.Now` is fixed for relative last-active time.

## Steps

1. Write a session with title and directory.
2. Write three messages with distinct token and cost values.
3. Set `req.SessionID` to the fixture session id.

```go
import (
	"testing"
	"time"
)

const knownSessionID = "ses_known_alpha"

func Setup(t *testing.T, req *Request) error {
	req.SessionID = knownSessionID
	updated, err := time.Parse(time.RFC3339, "2026-07-03T13:00:00.000Z")
	if err != nil {
		t.Fatalf("parse updated time: %v", err)
	}
	writeOpencodeSession(t, req.DataDir, "proj_known", knownSessionID,
		"Refactor auth module", "/tmp/opencode-known-project", updated)
	writeOpencodeMessage(t, req.DataDir, knownSessionID, "msg_known_01", opencodeMessageOpts{
		InputTokens: 100, OutputTokens: 50, Cost: 0.01,
	})
	writeOpencodeMessage(t, req.DataDir, knownSessionID, "msg_known_02", opencodeMessageOpts{
		InputTokens: 200, OutputTokens: 80, Cost: 0.02,
	})
	writeOpencodeMessage(t, req.DataDir, knownSessionID, "msg_known_03", opencodeMessageOpts{
		InputTokens: 50, OutputTokens: 20, Cost: 0.005,
	})
	return nil
}
```