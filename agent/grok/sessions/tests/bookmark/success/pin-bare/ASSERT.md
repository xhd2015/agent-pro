## Expected

- No error; Created=true.
- Tags empty; Description empty; AgentRunner=grok.

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
	assertEqualString(t, "AgentRunner", resp.Bookmark.AgentRunner, "grok")
	if len(resp.Bookmark.Tags) != 0 {
		t.Fatalf("Tags=%v want empty", resp.Bookmark.Tags)
	}
	assertEqualString(t, "Description", resp.Bookmark.Description, "")
	assertFileExists(t, storePath(req.AgentProHome))
}
```
