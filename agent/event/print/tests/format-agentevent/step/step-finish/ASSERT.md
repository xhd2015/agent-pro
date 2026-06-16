## Expected
- Output contains `◼ STEP FINISH`.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertContains(t, resp.Output, "◼ STEP FINISH")
}
```
