## Expected

- Exit code 0 (status command succeeds even when TCP is unreachable).
- Output indicates tcp reachable is false or "unreachable".

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)
	assertOutput(t, resp, "stdout", "tcp")
}
```
