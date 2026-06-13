## Expected
- Output contains the result text for the unknown tool.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertContains(t, resp.Output, "mystery result")
}
```
