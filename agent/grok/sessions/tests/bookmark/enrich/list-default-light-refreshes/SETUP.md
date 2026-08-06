# Scenario

**Feature**: default list enrich (light) refreshes title/msgs from session_dir summary

```
# preseed store snapshot; mutate summary.json under stored session_dir
ListBookmarks(Enrich="" / light)
  -> Title/NumChatMessages from live summary; Orphaned=false; no warnings
```

## Preconditions

- Live session under grokHome; store session_dir points at that dir.
- Store snapshot title/msgs intentionally older than live summary after mutate.
- Op=list; Enrich empty (default light).

## Steps

1. Write session with stored title/msgs; preseed catalog with those values.
2. Mutate summary.json to live title/msgs.
3. List with default enrich (req.Enrich="").

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.SessionID = fixtureBookmarkSessionID
	req.StoredTitle = "stored light title"
	req.StoredNumChatMessages = 10
	req.LiveTitle = "live light title after mutate"
	req.LiveNumChatMessages = 77
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
		Tags:            []string{"enrich-light"},
		Description:     "light refresh",
		CreatedAt:       created,
		UpdatedAt:       updated,
	}})
	// Mutate live summary under the stored session_dir (light reads this path only).
	writeSessionSummary(t, req.SessionDir, req.SessionID, absPath(t, req.FixtureCWD), req.LiveTitle, req.LiveNumChatMessages)
	req.Op = "list"
	req.Enrich = "" // default EnrichLight
	return nil
}
```
