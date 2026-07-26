## Expected
- Output contains the `SEARCH` header.
- Output includes the searched glob pattern below the header.

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertContains(t, resp.Output, "SEARCH")
	assertContains(t, resp.Output, ".agents/skills/git-fetch/**/*")
}
```
