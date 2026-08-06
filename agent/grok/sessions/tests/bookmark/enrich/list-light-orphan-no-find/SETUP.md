# Scenario

**Feature**: EnrichLight never calls Find — wrong session_dir orphans even if Find decoy exists

```
# live session under real grok path (Find would recover)
# store session_dir points at missing/wrong path
ListBookmarks(Enrich=light)
  -> Orphaned=true + warning; stored title kept (no Find recovery)
```

## Preconditions

- Real session written under normal `sessions/<encode(cwd)>/<id>/` so `Find`
  would discover it.
- Catalog `session_dir` is a path that does not contain summary.json.
- Op=list; Enrich=light (explicit).

## Steps

1. Write live session under grokHome (Find-able decoy).
2. Preseed store with wrong SessionDir and snapshot title.
3. List EnrichLight — must orphan without walking GROK_HOME.

```go
import (
	"os"
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.SessionID = fixtureBookmarkSessionID
	req.StoredTitle = "light orphan snapshot"
	req.StoredNumChatMessages = 12
	req.LiveTitle = "find decoy live title"
	req.LiveNumChatMessages = 33
	req.Title = req.StoredTitle
	req.NumChatMessages = req.StoredNumChatMessages
	// Real session Find can discover (decoy for proving light does not walk).
	liveDir := writeBookmarkSession(t, req.GrokHome, req.SessionID, req.FixtureCWD, req.LiveTitle, req.LiveNumChatMessages)
	_ = liveDir
	// Wrong session_dir: empty dir under temp, no summary.json.
	wrongDir := filepath.Join(req.TempDir, "wrong-session-dir-no-summary")
	if err := os.MkdirAll(wrongDir, 0o755); err != nil {
		t.Fatalf("mkdir wrong session dir: %v", err)
	}
	req.SessionDir = wrongDir
	created, updated := defaultPreseedTimes()
	writeStore(t, req.AgentProHome, []bookmarkEntry{{
		AgentRunner:     "grok",
		SessionID:       req.SessionID,
		SessionDir:      req.SessionDir,
		Title:           req.StoredTitle,
		NumChatMessages: req.StoredNumChatMessages,
		Tags:            []string{"light-orphan"},
		Description:     "wrong path",
		CreatedAt:       created,
		UpdatedAt:       updated,
	}})
	req.Op = "list"
	req.Enrich = "light"
	return nil
}
```
