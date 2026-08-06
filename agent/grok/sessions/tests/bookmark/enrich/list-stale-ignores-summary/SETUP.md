# Scenario

**Feature**: EnrichOff (--stale) ignores live summary; returns catalog snapshot only

```
# preseed store; mutate summary.json under session_dir
ListBookmarks(Enrich=stale/off)
  -> Title/NumChatMessages stay stored; Orphaned=false; no warnings; no FS refresh
```

## Preconditions

- Same fixture pattern as list-default-light-refreshes (snapshot vs live diverge).
- Op=list; Enrich=stale.

## Steps

1. Write session + store with stored title/msgs.
2. Mutate summary to different live title/msgs.
3. List with Enrich=stale.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.SessionID = fixtureBookmarkSessionID
	req.StoredTitle = "stale stored title"
	req.StoredNumChatMessages = 5
	req.LiveTitle = "live title stale must ignore"
	req.LiveNumChatMessages = 99
	req.Title = req.StoredTitle
	req.NumChatMessages = req.StoredNumChatMessages
	req.SessionDir = writeBookmarkSession(t, req.GrokHome, req.SessionID, req.FixtureCWD, req.StoredTitle, req.StoredNumChatMessages)
	created, updated := defaultPreseedTimes()
	writeStore(t, req.AgentProHome, []bookmarkEntry{{
		AgentRunner:     "grok",
		SessionID:       req.SessionID,
		SessionDir:      req.SessionDir,
		Title:           req.StoredTitle,
		NumChatMessages: req.StoredNumChatMessages,
		Tags:            []string{"enrich-stale"},
		Description:     "catalog only",
		CreatedAt:       created,
		UpdatedAt:       updated,
	}})
	writeSessionSummary(t, req.SessionDir, req.SessionID, absPath(t, req.FixtureCWD), req.LiveTitle, req.LiveNumChatMessages)
	req.Op = "list"
	req.Enrich = "stale"
	return nil
}
```
