# Scenario

**Feature**: re-pin merges tags, keeps created_at, refreshes denorm, created=false

```
# preseed store: tags [backup], title old, created_at fixed
# live summary title/msgs refreshed
BookmarkGrok(..., Tags=[migration,backup])
  -> tags union sorted; created=false; created_at stable; title/msgs refreshed
```

## Preconditions

- Live session with **new** title and num_chat_messages (different from store snapshot).
- Store preseeded with older title/msgs, tags `["backup"]`, description `"keep-me"`,
  fixed created_at / updated_at.

## Steps

1. Seed live session with refreshed denorm values.
2. Write preseed store row for same grok session id.
3. Pin with TagsSet merge tags including existing + new; Description nil.

```go
import (
	"path/filepath"
	"testing"
	"time"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.SessionID = fixtureBookmarkSessionID
	req.Title = "refreshed title after re-pin"
	req.NumChatMessages = 99
	req.SessionDir = writeBookmarkSession(t, req.GrokHome, req.SessionID, req.FixtureCWD, req.Title, req.NumChatMessages)

	created, _ := defaultPreseedTimes()
	ts, err := time.Parse(time.RFC3339, created)
	if err != nil {
		t.Fatalf("parse created_at: %v", err)
	}
	req.PreseedCreatedAt = ts
	writeStore(t, req.AgentProHome, []bookmarkEntry{{
		AgentRunner:     "grok",
		SessionID:       req.SessionID,
		SessionDir:      filepath.Join(req.TempDir, "old-session-dir"),
		Title:           "stale title",
		NumChatMessages: 1,
		Tags:            []string{"backup"},
		Description:     "keep-me",
		CreatedAt:       created,
		UpdatedAt:       created,
	}})

	req.Op = "pin"
	req.TagsSet = true
	req.Tags = []string{"migration", "backup"}
	// Description nil → keep "keep-me"
	return nil
}
```

