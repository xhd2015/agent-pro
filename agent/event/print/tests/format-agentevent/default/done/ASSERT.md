## Expected
- Output contains `[done]` (type in brackets).
- Output contains the text `all tasks completed`.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertContains(t, resp.Output, "[done]")
	assertContains(t, resp.Output, "all tasks completed")
}
```
