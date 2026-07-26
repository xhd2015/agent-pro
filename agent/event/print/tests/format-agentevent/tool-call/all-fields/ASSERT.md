## Expected
- Output contains all parts: RUN icon, text, output, file change, and FAILED.

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertContains(t, resp.Output, "RUN")
	assertContains(t, resp.Output, "Searching")
	assertContains(t, resp.Output, "found 3 matches")
	assertContains(t, resp.Output, "modify")
	assertContains(t, resp.Output, "a.txt")
	assertContains(t, resp.Output, "FAILED")
}
```
