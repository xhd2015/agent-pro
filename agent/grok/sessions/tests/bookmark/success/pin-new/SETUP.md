# Scenario

**Feature**: first pin creates catalog with grok denormalized fields

```
writeBookmarkSession(title, num_chat_messages)
  -> BookmarkGrok(agentProHome, grokHome, id, opts=nil-ish)
  -> store created; agent_runner=grok; created=true
```

## Preconditions

- Session exists under grokHome with known title and num_chat_messages.
- No prior `session_bookmarks.json`.
- Bare pin opts: NilOpts true (or empty defaults).

## Steps

1. Seed live session for fixture id.
2. Set Op=pin, SessionID, NilOpts=true.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.SessionID = fixtureBookmarkSessionID
	req.Title = "pin-new fixture"
	req.NumChatMessages = 42
	req.SessionDir = writeBookmarkSession(t, req.GrokHome, req.SessionID, req.FixtureCWD, req.Title, req.NumChatMessages)
	req.Op = "pin"
	req.NilOpts = true
	return nil
}
```
