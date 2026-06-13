## Expected
- Output contains the RUN label for bash and the tool output.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertContains(t, resp.Output, "RUN")
	assertContains(t, resp.Output, "hi there")
}
```
