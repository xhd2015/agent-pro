# Scenario

**Feature**: FormatBookmarksTable prints runner, session id, msgs, title, tags

```
# preseed store + live session
ListBookmarks -> FormatBookmarksTable
  -> header RUNNER / SESSION / MSGS / TITLE; row contains grok + id + title
```

## Preconditions

- Live session for hybrid enrich.
- Store preseeded with one grok bookmark.
- Op=format-list.

## Steps

1. Seed session and store.
2. Set Op=format-list.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.SessionID = fixtureBookmarkSessionID
	req.Title = "list-table fixture title"
	req.NumChatMessages = 42
	req.SessionDir = writeBookmarkSession(t, req.GrokHome, req.SessionID, req.FixtureCWD, req.Title, req.NumChatMessages)
	created, updated := defaultPreseedTimes()
	writeStore(t, req.AgentProHome, []bookmarkEntry{{
		AgentRunner:     "grok",
		SessionID:       req.SessionID,
		SessionDir:      req.SessionDir,
		Title:           req.Title,
		NumChatMessages: req.NumChatMessages,
		Tags:            []string{"backup", "migration"},
		Description:     "note",
		CreatedAt:       created,
		UpdatedAt:       updated,
	}})
	req.Op = "format-list"
	return nil
}
```
