# Scenario

**Feature**: list returns preseeded non-grok rows alongside grok

```
# store: agent_runner=codex + agent_runner=grok
ListBookmarks(filter all)
  -> both rows present
```

## Preconditions

- Live grok session for hybrid enrich of grok row.
- Codex row only in store (no live enrich required).
- Op=list, Runner="".

## Steps

1. Seed grok session.
2. Preseed codex + grok store rows.
3. Op=list.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.SessionID = fixtureBookmarkSessionID
	req.Title = "preseed grok title"
	req.NumChatMessages = 9
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
			SessionDir:      "/tmp/codex/preseed-session",
			Title:           "codex preseed title",
			NumChatMessages: 3,
			Tags:            []string{"c"},
			CreatedAt:       created,
			UpdatedAt:       updated,
		},
	})
	req.Op = "list"
	req.Runner = ""
	return nil
}
```
