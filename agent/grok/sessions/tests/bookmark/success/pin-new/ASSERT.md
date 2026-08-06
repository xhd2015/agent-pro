## Expected

- No error.
- `Created` = true.
- Bookmark: `AgentRunner=grok`, `SessionID` = fixture id, `Title` = fixture title,
  `NumChatMessages=42`, `SessionDir` absolute equals seeded dir, empty tags,
  empty description.
- Store file exists at `{agentProHome}/session_bookmarks.json` with version 1
  and one bookmark matching the key.

## Errors

- None.

```go
import (
	"path/filepath"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertNoHarnessErr(t, err)
	assertNoError(t, resp)
	assertEqualBool(t, "Created", resp.Created, true)
	if resp.Bookmark == nil {
		t.Fatal("Bookmark is nil")
	}
	bm := resp.Bookmark
	assertEqualString(t, "AgentRunner", bm.AgentRunner, "grok")
	assertEqualString(t, "SessionID", bm.SessionID, req.SessionID)
	assertEqualString(t, "Title", bm.Title, req.Title)
	assertEqualInt(t, "NumChatMessages", bm.NumChatMessages, req.NumChatMessages)
	if filepath.Clean(bm.SessionDir) != filepath.Clean(req.SessionDir) {
		t.Fatalf("SessionDir=%q want %q", bm.SessionDir, req.SessionDir)
	}
	if !filepath.IsAbs(bm.SessionDir) {
		t.Fatalf("SessionDir not absolute: %q", bm.SessionDir)
	}
	if len(bm.Tags) != 0 {
		t.Fatalf("Tags=%v, want empty", bm.Tags)
	}
	assertEqualString(t, "Description", bm.Description, "")
	if bm.CreatedAt.IsZero() || bm.UpdatedAt.IsZero() {
		t.Fatalf("timestamps zero: created=%v updated=%v", bm.CreatedAt, bm.UpdatedAt)
	}

	assertFileExists(t, storePath(req.AgentProHome))
	entries := readStoreEntries(t, req.AgentProHome)
	if len(entries) != 1 {
		t.Fatalf("store entries=%d, want 1", len(entries))
	}
	e, ok := findEntry(entries, "grok", req.SessionID)
	if !ok {
		t.Fatalf("store missing grok/%s: %+v", req.SessionID, entries)
	}
	assertEqualString(t, "store.title", e.Title, req.Title)
	assertEqualInt(t, "store.num_chat_messages", e.NumChatMessages, req.NumChatMessages)
}
```
