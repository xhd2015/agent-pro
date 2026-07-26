## Expected
- Output contains `[custom_event]` (type in brackets).
- Output contains the text `some arbitrary event`.

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertContains(t, resp.Output, "[custom_event]")
	assertContains(t, resp.Output, "some arbitrary event")
}
```
