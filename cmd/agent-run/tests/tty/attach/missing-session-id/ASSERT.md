## Expected

- Exit code 1.
- Stderr mentions missing session id or usage.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertExitCode(t, resp, 1)
	assertOutput(t, resp, "stderr", "session")
}
```
