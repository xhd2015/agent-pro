## Expected

- Exit code 1.
- Stderr mentions session not found or expired.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertExitCode(t, resp, 1)
	assertOutput(t, resp, "stderr", "not found")
}
```
