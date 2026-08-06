# Scenario

**Feature**: ListFilter.Tags AND-match returns only bookmarks with all tags

```
# store: A tags [backup,migration]; B tags [backup]
ListBookmarks(Tags=[backup,migration])
  -> only A
```

## Preconditions

- Two grok sessions on disk (or preseed with dirs; hybrid optional).
- Store: session1 tags backup+migration; session2 tags backup only.
- FilterTags = backup, migration.

## Steps

1. Seed two sessions.
2. Preseed store.
3. Op=list with FilterTags AND.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.SessionID = fixtureBookmarkSessionID
	dir1 := writeBookmarkSession(t, req.GrokHome, fixtureBookmarkSessionID, req.FixtureCWD, "tag both", 2)
	dir2 := writeBookmarkSession(t, req.GrokHome, fixtureBookmarkSessionID2, req.FixtureCWD, "tag backup only", 3)
	created, updated := defaultPreseedTimes()
	writeStore(t, req.AgentProHome, []bookmarkEntry{
		{
			AgentRunner:     "grok",
			SessionID:       fixtureBookmarkSessionID,
			SessionDir:      dir1,
			Title:           "tag both",
			NumChatMessages: 2,
			Tags:            []string{"backup", "migration"},
			CreatedAt:       created,
			UpdatedAt:       updated,
		},
		{
			AgentRunner:     "grok",
			SessionID:       fixtureBookmarkSessionID2,
			SessionDir:      dir2,
			Title:           "tag backup only",
			NumChatMessages: 3,
			Tags:            []string{"backup"},
			CreatedAt:       created,
			UpdatedAt:       updated,
		},
	})
	req.Op = "list"
	req.FilterTags = []string{"backup", "migration"}
	return nil
}
```
