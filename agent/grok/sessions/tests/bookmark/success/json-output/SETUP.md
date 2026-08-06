# Scenario

**Feature**: FormatBookmarkJSON emits key fields without ANSI

```
# preseed store + live session
ListBookmarks -> FormatBookmarkJSON
  -> JSON array/object with agent_runner, session_id, title; no ANSI
```

## Preconditions

- One bookmarked live session.
- Op=format-json.

## Steps

1. Seed session + store.
2. Op=format-json.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.SessionID = fixtureBookmarkSessionID
	req.Title = "json-output fixture"
	req.NumChatMessages = 6
	req.SessionDir = writeBookmarkSession(t, req.GrokHome, req.SessionID, req.FixtureCWD, req.Title, req.NumChatMessages)
	created, updated := defaultPreseedTimes()
	writeStore(t, req.AgentProHome, []bookmarkEntry{{
		AgentRunner:     "grok",
		SessionID:       req.SessionID,
		SessionDir:      req.SessionDir,
		Title:           req.Title,
		NumChatMessages: req.NumChatMessages,
		Tags:            []string{"json"},
		Description:     "json note",
		CreatedAt:       created,
		UpdatedAt:       updated,
	}})
	req.Op = "format-json"
	return nil
}
```
