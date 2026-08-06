# Scenario

**Feature**: list marks orphan when bookmarked grok session dir is gone

```
# pin live session, then remove session directory tree
ListBookmarks
  -> Orphaned=true; warning contains session id + not found; stored title kept
```

## Preconditions

- Store preseeded (or written) with grok bookmark and known title snapshot.
- Session directory deleted after preseed so Find fails.
- Op=list.

## Steps

1. Seed session; write store with snapshot title.
2. Remove the session directory (and summary) so Find fails.
3. Op=list.

```go
import (
	"os"
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.SessionID = fixtureBookmarkSessionID
	req.Title = "orphan snapshot title"
	req.NumChatMessages = 20
	req.SessionDir = writeBookmarkSession(t, req.GrokHome, req.SessionID, req.FixtureCWD, req.Title, req.NumChatMessages)
	created, updated := defaultPreseedTimes()
	writeStore(t, req.AgentProHome, []bookmarkEntry{{
		AgentRunner:     "grok",
		SessionID:       req.SessionID,
		SessionDir:      req.SessionDir,
		Title:           req.Title,
		NumChatMessages: req.NumChatMessages,
		Tags:            []string{"orphan"},
		Description:     "still bookmarked",
		CreatedAt:       created,
		UpdatedAt:       updated,
	}})
	// Delete session so hybrid Find fails.
	if err := os.RemoveAll(req.SessionDir); err != nil {
		t.Fatalf("remove session dir: %v", err)
	}
	req.Op = "list"
	return nil
}
```
