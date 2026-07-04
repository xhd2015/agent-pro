# Scenario

**Feature**: info for session without message files shows zero MSGS and no Tokens

```
# session JSON exists but storage/message/{id}/ is absent
writeOpencodeSession only -> sessions.Info -> FormatInfoText(now)

# NumMessages=0; no Tokens/Cost section in output
SessionInfo without message aggregates
```

## Preconditions

- Session JSON file is present.
- No message JSON files are written.

## Steps

1. Write one session fixture.
2. Set `req.SessionID` to that session id.

```go
import (
	"testing"
	"time"
)

const noMessagesSessionID = "ses_no_msgs"

func Setup(t *testing.T, req *Request) error {
	req.SessionID = noMessagesSessionID
	updated, err := time.Parse(time.RFC3339, "2026-07-03T14:00:00.000Z")
	if err != nil {
		t.Fatalf("parse updated time: %v", err)
	}
	writeOpencodeSession(t, req.DataDir, "proj_no_msgs", noMessagesSessionID,
		"Docs cleanup", "/tmp/opencode-no-msgs", updated)
	return nil
}
```