## Expected
- Output contains the RUN icon/label.
- Output contains the text `Running command...`.
- Output contains the tool result `command result`.

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertContains(t, resp.Output, "RUN")
	assertContains(t, resp.Output, "Running command")
	assertContains(t, resp.Output, "command result")
}
```
