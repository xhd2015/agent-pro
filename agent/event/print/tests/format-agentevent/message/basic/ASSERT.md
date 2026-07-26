## Expected
- Output contains the 💬 icon.
- Output contains the message text `hello world`.

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertContains(t, resp.Output, "💬")
	assertContains(t, resp.Output, "hello world")
}
```
