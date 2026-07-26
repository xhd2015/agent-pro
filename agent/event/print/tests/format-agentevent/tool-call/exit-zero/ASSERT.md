## Expected
- Output contains RUN icon/label and the output text.
- Output does NOT contain FAILED (exit_code 0 is success).

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertContains(t, resp.Output, "RUN")
	assertContains(t, resp.Output, "ok")
	assertNotContains(t, resp.Output, "FAILED")
}
```
