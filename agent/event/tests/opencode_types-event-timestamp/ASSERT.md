## Expected
- Timestamp is parsed from JSON correctly.
- When re-marshaled, the timestamp is included in the output.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)
	assertContains(t, resp.Stdout, "type=text")
	assertContains(t, resp.Stdout, "sessionID=sess_ts")
	assertContains(t, resp.Stdout, "timestamp=1718200000999")
	assertContains(t, resp.Stdout, "has_timestamp=true")
}
```
