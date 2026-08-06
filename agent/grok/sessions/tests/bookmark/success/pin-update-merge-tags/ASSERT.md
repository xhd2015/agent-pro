## Expected

- No error; `Created` = false.
- Tags are union sorted: `["backup","migration"]`.
- Description still `keep-me` (Description opts nil).
- Title and NumChatMessages refreshed from live session.
- SessionDir refreshed to live session dir (not preseed old path).
- `CreatedAt` equals preseed created_at (stable).
- `UpdatedAt` is at or after CreatedAt / not equal to stale preseed if clock moves
  (at least non-zero and not before CreatedAt).

## Errors

- None.

```go
import (
	"path/filepath"
	"testing"
	"time"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertNoHarnessErr(t, err)
	assertNoError(t, resp)
	assertEqualBool(t, "Created", resp.Created, false)
	if resp.Bookmark == nil {
		t.Fatal("Bookmark is nil")
	}
	bm := resp.Bookmark
	assertTagsEqual(t, bm.Tags, []string{"backup", "migration"})
	assertEqualString(t, "Description", bm.Description, "keep-me")
	assertEqualString(t, "Title", bm.Title, req.Title)
	assertEqualInt(t, "NumChatMessages", bm.NumChatMessages, req.NumChatMessages)
	if filepath.Clean(bm.SessionDir) != filepath.Clean(req.SessionDir) {
		t.Fatalf("SessionDir=%q want live %q", bm.SessionDir, req.SessionDir)
	}
	// created_at stable (allow second resolution equality)
	if !bm.CreatedAt.Equal(req.PreseedCreatedAt) && !bm.CreatedAt.Truncate(time.Second).Equal(req.PreseedCreatedAt.Truncate(time.Second)) {
		t.Fatalf("CreatedAt=%v want preseed %v", bm.CreatedAt, req.PreseedCreatedAt)
	}
	if bm.UpdatedAt.Before(bm.CreatedAt) {
		t.Fatalf("UpdatedAt %v before CreatedAt %v", bm.UpdatedAt, bm.CreatedAt)
	}
}
```
