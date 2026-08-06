# Scenario

**Feature**: GetBookmark with empty runner succeeds when session_id is unique

```
# single store row for id
GetBookmark(runner="", sessionID)
  -> view returned; AgentRunner=grok
```

## Preconditions

- Exactly one bookmark matches SessionID (no multi-runner collision).
- Op=show, Runner="".

## Steps

1. Seed session + single store row.
2. Op=show without runner.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.SessionID = fixtureBookmarkSessionID
	req.Title = "unique show fixture"
	req.NumChatMessages = 8
	req.SessionDir = writeBookmarkSession(t, req.GrokHome, req.SessionID, req.FixtureCWD, req.Title, req.NumChatMessages)
	created, updated := defaultPreseedTimes()
	writeStore(t, req.AgentProHome, []bookmarkEntry{{
		AgentRunner:     "grok",
		SessionID:       req.SessionID,
		SessionDir:      req.SessionDir,
		Title:           req.Title,
		NumChatMessages: req.NumChatMessages,
		Tags:            nil,
		CreatedAt:       created,
		UpdatedAt:       updated,
	}})
	req.Op = "show"
	req.Runner = ""
	return nil
}
```
