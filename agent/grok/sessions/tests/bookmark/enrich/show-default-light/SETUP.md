# Scenario

**Feature**: GetBookmark default light refreshes from session_dir summary

```
# preseed store; mutate summary under session_dir
GetBookmark(enrich=light / "")
  -> View Title/NumChatMessages from live summary; Orphaned=false
```

## Preconditions

- Live session + store with matching session_dir.
- Summary mutated after preseed (same as list light refresh).
- Op=show; Runner=grok; Enrich empty (default light).

## Steps

1. Seed session + store snapshot.
2. Mutate summary.json title/msgs.
3. Show with default enrich.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.SessionID = fixtureBookmarkSessionID
	req.StoredTitle = "show stored title"
	req.StoredNumChatMessages = 4
	req.LiveTitle = "show live title light"
	req.LiveNumChatMessages = 21
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
		Tags:            []string{"show-light"},
		Description:     "show light path",
		CreatedAt:       created,
		UpdatedAt:       updated,
	}})
	writeSessionSummary(t, req.SessionDir, req.SessionID, absPath(t, req.FixtureCWD), req.LiveTitle, req.LiveNumChatMessages)
	req.Op = "show"
	req.Runner = "grok"
	req.Enrich = "" // EnrichLight default
	return nil
}
```
