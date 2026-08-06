## Expected

- No error; created=true.
- Bookmark.Tags equals `["alpha","backup","migration"]` (sorted, unique, trimmed).
- Store entry tags match the same ordered list.

## Errors

- None.

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertNoHarnessErr(t, err)
	assertNoError(t, resp)
	assertEqualBool(t, "Created", resp.Created, true)
	if resp.Bookmark == nil {
		t.Fatal("Bookmark is nil")
	}
	want := []string{"alpha", "backup", "migration"}
	assertTagsEqual(t, resp.Bookmark.Tags, want)
	if !tagsSortedUnique(resp.Bookmark.Tags) {
		t.Fatalf("tags not sorted unique: %v", resp.Bookmark.Tags)
	}
	entries := readStoreEntries(t, req.AgentProHome)
	e, ok := findEntry(entries, "grok", req.SessionID)
	if !ok {
		t.Fatal("store entry missing")
	}
	assertTagsEqual(t, e.Tags, want)
}
```
