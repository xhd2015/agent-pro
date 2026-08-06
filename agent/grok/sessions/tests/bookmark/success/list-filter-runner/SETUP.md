# Scenario

**Feature**: ListFilter.Runner returns only matching agent_runner

```
# store: grok row + codex row
ListBookmarks(filter Runner=grok)
  -> only grok view
```

## Preconditions

- Store preseeded with grok and codex bookmarks (different session ids).
- Live grok session exists for hybrid enrich of the grok row.
- Runner filter = `grok`.

## Steps

1. Seed grok session.
2. Preseed two store rows.
3. Op=list, Runner=grok.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.SessionID = fixtureBookmarkSessionID
	req.Title = "filter-runner grok"
	req.NumChatMessages = 4
	req.SessionDir = writeBookmarkSession(t, req.GrokHome, req.SessionID, req.FixtureCWD, req.Title, req.NumChatMessages)
	created, updated := defaultPreseedTimes()
	writeStore(t, req.AgentProHome, []bookmarkEntry{
		{
			AgentRunner:     "grok",
			SessionID:       req.SessionID,
			SessionDir:      req.SessionDir,
			Title:           req.Title,
			NumChatMessages: req.NumChatMessages,
			Tags:            []string{"g"},
			CreatedAt:       created,
			UpdatedAt:       updated,
		},
		{
			AgentRunner:     "codex",
			SessionID:       fixtureBookmarkSessionID2,
			SessionDir:      "/tmp/codex-session-dir",
			Title:           "codex only",
			NumChatMessages: 1,
			Tags:            []string{"c"},
			CreatedAt:       created,
			UpdatedAt:       updated,
		},
	})
	req.Op = "list"
	req.Runner = "grok"
	return nil
}
```
