# Scenario

**Feature**: bare pin with empty PinOptions still succeeds

```
writeBookmarkSession -> BookmarkGrok(..., &PinOptions{})
  -> created=true; empty tags; empty description
```

## Preconditions

- Live session exists.
- NilOpts false with empty PinOptions (no TagsSet, no Description, ClearTags false).

## Steps

1. Seed session.
2. Op=pin with default empty opts fields.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.SessionID = fixtureBookmarkSessionID
	req.Title = "pin-bare fixture"
	req.NumChatMessages = 2
	req.SessionDir = writeBookmarkSession(t, req.GrokHome, req.SessionID, req.FixtureCWD, req.Title, req.NumChatMessages)
	req.Op = "pin"
	req.NilOpts = false
	req.TagsSet = false
	req.ClearTags = false
	req.Description = nil
	return nil
}
```
