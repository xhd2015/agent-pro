# Scenario

**Feature**: EnrichHeavy still orphans when session is nowhere under grokHome

```
# store bookmark; no live session dir and Find cannot discover id
ListBookmarks(Enrich=heavy)
  -> Orphaned=true + warning; stored title kept
```

## Preconditions

- Catalog row for session id with no summary under grokHome.
- Op=list; Enrich=heavy (or enrich).

## Steps

1. Preseed store only (no writeBookmarkSession for that id).
2. List EnrichHeavy.

```go
import (
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.SessionID = fixtureBookmarkSessionID
	req.StoredTitle = "heavy still orphan title"
	req.StoredNumChatMessages = 8
	req.Title = req.StoredTitle
	req.NumChatMessages = req.StoredNumChatMessages
	// Point at a path that does not exist; also ensure no Find-able session.
	req.SessionDir = filepath.Join(req.TempDir, "no-such-session-dir")
	created, updated := defaultPreseedTimes()
	writeStore(t, req.AgentProHome, []bookmarkEntry{{
		AgentRunner:     "grok",
		SessionID:       req.SessionID,
		SessionDir:      req.SessionDir,
		Title:           req.StoredTitle,
		NumChatMessages: req.StoredNumChatMessages,
		Tags:            []string{"heavy-orphan"},
		Description:     "gone everywhere",
		CreatedAt:       created,
		UpdatedAt:       updated,
	}})
	req.Op = "list"
	req.Enrich = "heavy"
	return nil
}
```
