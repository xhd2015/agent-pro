# Scenario

**Feature**: GetBookmark + FormatBookmarkShow includes full detail fields

```
# preseed store + live session
GetBookmark(runner=grok) -> FormatBookmarkShow
  -> output contains runner, id, title, msgs, session_dir, tags, description
```

## Preconditions

- Live session and preseed store with tags + description.
- Op=format-show, Runner=grok.

## Steps

1. Seed session + store with description and tags.
2. Set Op=format-show.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.SessionID = fixtureBookmarkSessionID
	req.Title = "show-detail fixture"
	req.NumChatMessages = 15
	req.SessionDir = writeBookmarkSession(t, req.GrokHome, req.SessionID, req.FixtureCWD, req.Title, req.NumChatMessages)
	created, updated := defaultPreseedTimes()
	writeStore(t, req.AgentProHome, []bookmarkEntry{{
		AgentRunner:     "grok",
		SessionID:       req.SessionID,
		SessionDir:      req.SessionDir,
		Title:           req.Title,
		NumChatMessages: req.NumChatMessages,
		Tags:            []string{"show", "detail"},
		Description:     "user note for show",
		CreatedAt:       created,
		UpdatedAt:       updated,
	}})
	req.Op = "format-show"
	req.Runner = "grok"
	return nil
}
```
