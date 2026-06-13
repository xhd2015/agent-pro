## Expected
- Every opencode event type round-trips unchanged: the output must contain `ALL MATCH`.
- If any type mismatches, the output shows `[FAIL] <type>` with orig and rt JSON for diagnosis.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertContains(t, resp.Output, "ALL MATCH")
}
```
