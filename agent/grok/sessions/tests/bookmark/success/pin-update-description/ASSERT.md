## Expected

- No error; Created=false.
- Description is `new-desc`.
- Tags remain `["a"]`.

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
	assertEqualString(t, "Description", resp.Bookmark.Description, "new-desc")
	assertTagsEqual(t, resp.Bookmark.Tags, []string{"a"})
}
```
