# Scenario

**Feature**: pin normalizes tags (trim, drop empty, dedupe, sort)

```
BookmarkGrok(..., Tags=["backup"," migration ","backup","", "alpha"])
  -> stored tags ["alpha","backup","migration"] sorted unique
```

## Preconditions

- Live session exists.
- TagsSet true with messy input tags (duplicates, spaces, empty).

## Steps

1. Seed session.
2. Set pin Tags and TagsSet; Op=pin.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.SessionID = fixtureBookmarkSessionID
	req.Title = "pin-with-tags fixture"
	req.NumChatMessages = 7
	req.SessionDir = writeBookmarkSession(t, req.GrokHome, req.SessionID, req.FixtureCWD, req.Title, req.NumChatMessages)
	req.Op = "pin"
	req.TagsSet = true
	req.Tags = []string{"backup", " migration ", "backup", "", "alpha"}
	return nil
}
```
