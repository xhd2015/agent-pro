## Expected
- Output is empty (non-assistant message events produce no formatted output).

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertEmpty(t, resp.Output)
}
```
