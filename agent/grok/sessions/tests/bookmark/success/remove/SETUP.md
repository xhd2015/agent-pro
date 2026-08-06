# Scenario

**Feature**: RemoveBookmark deletes entry; second remove errors

```
# preseed one bookmark
RemoveBookmark(runner=grok, id)
  -> store empty / entry gone
  -> second RemoveBookmark -> not found error
```

## Preconditions

- Store preseeded with one grok bookmark (session may exist).
- Op=remove, Runner=grok.

## Steps

1. Seed session + store.
2. Op=remove.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.SessionID = fixtureBookmarkSessionID
	req.Title = "remove fixture"
	req.NumChatMessages = 1
	req.SessionDir = writeBookmarkSession(t, req.GrokHome, req.SessionID, req.FixtureCWD, req.Title, req.NumChatMessages)
	created, updated := defaultPreseedTimes()
	writeStore(t, req.AgentProHome, []bookmarkEntry{{
		AgentRunner:     "grok",
		SessionID:       req.SessionID,
		SessionDir:      req.SessionDir,
		Title:           req.Title,
		NumChatMessages: req.NumChatMessages,
		CreatedAt:       created,
		UpdatedAt:       updated,
	}})
	req.Op = "remove"
	req.Runner = "grok"
	return nil
}
```
