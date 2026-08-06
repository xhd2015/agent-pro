# Scenario

**Feature**: format never uses emoji / multi-line USER card chrome from session view

```
# normal prompts formatted
Output must not contain "👤" or "USER" card-style headers
```

## Preconditions

- One session with a normal user prompt.
- Op format-single.

## Steps

1. Write simple session.
2. Format and assert absence of session-view chrome.

```go
import (
	"testing"
	"time"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Op = "format-single"
	req.SessionID = idFormatSingle
	ts := atFixed(-12 * time.Minute)
	writePromptSession(t, req.GrokHome, promptSessionOpts{
		ID:           idFormatSingle,
		Title:        "no emoji",
		LastActiveAt: ts,
		Updates:      updatesJSONL(userChunkAt("plain prompt", ts)),
	})
	return nil
}
```
