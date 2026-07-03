## Expected

- Exit code 0 (status command succeeds even when screen can't be read).
- Stdout contains "screen" with "unknown" or error indicator.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)
	assertOutput(t, resp, "stdout", "screen", "unknown")
}
```
