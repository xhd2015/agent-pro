## Expected

- No error; Created=false.
- Tags equal `["fresh"]` only (old tags wiped).

## Errors

- None.

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertNoHarnessErr(t, err)
	assertNoError(t, resp)
	assertEqualBool(t, "Created", resp.Created, false)
	if resp.Bookmark == nil {
		t.Fatal("Bookmark is nil")
	}
	assertTagsEqual(t, resp.Bookmark.Tags, []string{"fresh"})
}
```
