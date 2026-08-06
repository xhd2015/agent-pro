# Scenario

**Feature**: corrupt store JSON errors list/pin and does not clobber file

```
# write invalid JSON to session_bookmarks.json
ListBookmarks / BookmarkGrok
  -> error; file body unchanged (not replaced with empty valid store)
```

## Preconditions

- Corrupt marker body written to store path (not valid JSON object).
- Live session exists so pin would otherwise succeed.
- This leaf uses Op=list first path; Assert also verifies pin would not wipe
  (pin via Assert call after list).

## Steps

1. Seed live session.
2. Write corrupt store body.
3. Op=list (must error).

```go
import "testing"

const corruptStoreBody = "{not valid json for bookmarks!!!"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.SessionID = fixtureBookmarkSessionID
	req.Title = "corrupt store fixture"
	req.NumChatMessages = 1
	req.SessionDir = writeBookmarkSession(t, req.GrokHome, req.SessionID, req.FixtureCWD, req.Title, req.NumChatMessages)
	req.CorruptMarker = corruptStoreBody
	mustWriteFile(t, storePath(req.AgentProHome), corruptStoreBody)
	req.Op = "list"
	return nil
}
```
