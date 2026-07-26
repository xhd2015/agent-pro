## Expected
- Output contains `[sleep]` (type in brackets).
- Output contains the text `waiting 5000ms`.

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertContains(t, resp.Output, "[sleep]")
	assertContains(t, resp.Output, "waiting 5000ms")
}
```
