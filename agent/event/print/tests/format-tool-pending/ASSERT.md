## Expected
- Output contains the RUN label for bash with the command text.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertContains(t, resp.Output, "RUN")
	assertContains(t, resp.Output, "sleep 10")
}
```
