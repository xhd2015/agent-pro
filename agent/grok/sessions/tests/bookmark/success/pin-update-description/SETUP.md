# Scenario

**Feature**: Description pointer sets new text; nil Tags keeps existing tags

```
# preseed: description old-desc, tags [a]
BookmarkGrok(..., Description=&"new-desc", Tags nil)
  -> description=new-desc; tags still [a]
```

## Preconditions

- Live session exists.
- Store preseeded with description `old-desc` and tags `["a"]`.

## Steps

1. Seed session.
2. Preseed store.
3. Pin with Description pointing to new string; TagsSet false.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.SessionID = fixtureBookmarkSessionID
	req.Title = "desc update fixture"
	req.NumChatMessages = 3
	req.SessionDir = writeBookmarkSession(t, req.GrokHome, req.SessionID, req.FixtureCWD, req.Title, req.NumChatMessages)

	created, updated := defaultPreseedTimes()
	writeStore(t, req.AgentProHome, []bookmarkEntry{{
		AgentRunner:     "grok",
		SessionID:       req.SessionID,
		SessionDir:      req.SessionDir,
		Title:           req.Title,
		NumChatMessages: req.NumChatMessages,
		Tags:            []string{"a"},
		Description:     "old-desc",
		CreatedAt:       created,
		UpdatedAt:       updated,
	}})

	desc := "new-desc"
	req.Description = &desc
	req.TagsSet = false
	req.Op = "pin"
	return nil
}
```
