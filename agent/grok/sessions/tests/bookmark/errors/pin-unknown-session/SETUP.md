# Scenario

**Feature**: pin unknown session errors and does not create store

```
# empty sessions tree
BookmarkGrok(unknown-id)
  -> error containing not found; session_bookmarks.json absent
```

## Preconditions

- No session directory for the id.
- No pre-existing store file.

## Steps

1. Set unknown SessionID; Op=pin.
2. Do not write session fixtures.

```go
import "testing"

const unknownBookmarkSessionID = "019f283a-eeee-7eee-eeee-eeeeeeeeee88"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.SessionID = unknownBookmarkSessionID
	req.Op = "pin"
	req.NilOpts = true
	return nil
}
```
