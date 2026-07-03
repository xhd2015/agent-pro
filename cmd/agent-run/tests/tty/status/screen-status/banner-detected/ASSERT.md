## Expected

- Exit code 0.
- Stdout contains "screen" with indication that banner was detected (or "banner" appears in screen status field).

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)
	assertOutput(t, resp, "stdout", "screen", "banner")
}
```
