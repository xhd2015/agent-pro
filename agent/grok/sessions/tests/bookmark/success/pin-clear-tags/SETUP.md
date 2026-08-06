# Scenario

**Feature**: ClearTags wipes existing tags then merges any provided tags

```
# preseed tags [old,keep]
BookmarkGrok(..., ClearTags=true, Tags=[fresh])
  -> tags [fresh] only
```

## Preconditions

- Live session exists.
- Store preseeded with tags `["keep","old"]`.

## Steps

1. Seed session + preseed store.
2. Pin with ClearTags=true, TagsSet=true, Tags=["fresh"].

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.SessionID = fixtureBookmarkSessionID
	req.Title = "clear-tags fixture"
	req.NumChatMessages = 5
	req.SessionDir = writeBookmarkSession(t, req.GrokHome, req.SessionID, req.FixtureCWD, req.Title, req.NumChatMessages)

	created, updated := defaultPreseedTimes()
	writeStore(t, req.AgentProHome, []bookmarkEntry{{
		AgentRunner:     "grok",
		SessionID:       req.SessionID,
		SessionDir:      req.SessionDir,
		Title:           req.Title,
		NumChatMessages: req.NumChatMessages,
		Tags:            []string{"keep", "old"},
		Description:     "",
		CreatedAt:       created,
		UpdatedAt:       updated,
	}})

	req.Op = "pin"
	req.ClearTags = true
	req.TagsSet = true
	req.Tags = []string{"fresh"}
	return nil
}
```
