# Scenario

**Feature**: GetBookmark with empty runner errors when session_id matches 2+ runners

```
# store: same session_id for grok and codex
GetBookmark(runner="", sessionID)
  -> error asking for runner (ambiguous)
```

## Preconditions

- Two store rows share `fixtureBookmarkCollideID` with runners grok and codex.
- Op=show, Runner="".

## Steps

1. Optionally seed grok session for the collide id.
2. Preseed two runners same session_id.
3. Op=show with empty runner.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.SessionID = fixtureBookmarkCollideID
	req.Title = "ambiguous fixture"
	req.NumChatMessages = 1
	req.SessionDir = writeBookmarkSession(t, req.GrokHome, req.SessionID, req.FixtureCWD, req.Title, req.NumChatMessages)
	created, updated := defaultPreseedTimes()
	writeStore(t, req.AgentProHome, []bookmarkEntry{
		{
			AgentRunner:     "grok",
			SessionID:       req.SessionID,
			SessionDir:      req.SessionDir,
			Title:           req.Title,
			NumChatMessages: req.NumChatMessages,
			CreatedAt:       created,
			UpdatedAt:       updated,
		},
		{
			AgentRunner:     "codex",
			SessionID:       req.SessionID,
			SessionDir:      "/tmp/codex/ambiguous",
			Title:           "codex same id",
			NumChatMessages: 2,
			CreatedAt:       created,
			UpdatedAt:       updated,
		},
	})
	req.Op = "show"
	req.Runner = ""
	return nil
}
```
