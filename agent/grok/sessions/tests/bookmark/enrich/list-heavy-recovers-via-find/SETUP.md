# Scenario

**Feature**: EnrichHeavy recovers via Find when session_dir is wrong/empty

```
# store empty or wrong session_dir; live session exists under grokHome
ListBookmarks(Enrich=heavy)
  -> Find recovers Title/SessionDir/NumChatMessages; Orphaned=false
```

## Preconditions

- Live session under normal Find layout.
- Catalog SessionDir empty (or missing path) so light would orphan; heavy Find succeeds.
- Op=list; Enrich=heavy.

## Steps

1. Write live session with known live title/msgs.
2. Preseed store with empty SessionDir and different stored title.
3. List EnrichHeavy — must recover via Find after light fails.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.SessionID = fixtureBookmarkSessionID
	req.StoredTitle = "heavy stored before recover"
	req.StoredNumChatMessages = 1
	req.LiveTitle = "heavy recovered live title"
	req.LiveNumChatMessages = 55
	req.Title = req.LiveTitle
	req.NumChatMessages = req.LiveNumChatMessages
	// Live session discoverable only via Find when session_dir is empty.
	req.SessionDir = writeBookmarkSession(t, req.GrokHome, req.SessionID, req.FixtureCWD, req.LiveTitle, req.LiveNumChatMessages)
	created, updated := defaultPreseedTimes()
	writeStore(t, req.AgentProHome, []bookmarkEntry{{
		AgentRunner:     "grok",
		SessionID:       req.SessionID,
		SessionDir:      "", // empty → light cannot refresh; heavy must Find
		Title:           req.StoredTitle,
		NumChatMessages: req.StoredNumChatMessages,
		Tags:            []string{"heavy-recover"},
		Description:     "recover via find",
		CreatedAt:       created,
		UpdatedAt:       updated,
	}})
	req.Op = "list"
	req.Enrich = "heavy"
	return nil
}
```
