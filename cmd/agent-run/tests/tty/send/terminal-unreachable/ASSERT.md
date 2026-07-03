## Expected

- Exit code 1 (terminal unreachable).
- Stderr mentions connection error or terminal unavailable.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertExitCode(t, resp, 1)
	assertOutput(t, resp, "stderr", "terminal")
}
```
