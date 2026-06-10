## Expected
- Notify rejects the event.

```go
import "testing"
func Assert(t *testing.T, req *Request, resp *Response, err error) { expectErrContains(t, resp, "event_type") }
```

