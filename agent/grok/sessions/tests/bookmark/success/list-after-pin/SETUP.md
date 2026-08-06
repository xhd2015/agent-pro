# Scenario

**Feature**: pin then list roundtrip returns the bookmarked session

```
writeBookmarkSession -> BookmarkGrok then ListBookmarks
  -> one view: grok, matching id/title/msgs, Orphaned=false
```

## Preconditions

- Live session exists.
- Op=pin-list performs pin then list in one Run.

## Steps

1. Seed session with known title/tags via pin Tags.
2. Set Op=pin-list.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.SessionID = fixtureBookmarkSessionID
	req.Title = "list-after-pin fixture"
	req.NumChatMessages = 11
	req.SessionDir = writeBookmarkSession(t, req.GrokHome, req.SessionID, req.FixtureCWD, req.Title, req.NumChatMessages)
	req.Op = "pin-list"
	req.TagsSet = true
	req.Tags = []string{"roundtrip"}
	return nil
}
```
