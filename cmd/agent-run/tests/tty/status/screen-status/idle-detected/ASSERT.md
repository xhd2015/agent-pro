## Expected

- Exit code 0.
- Stdout contains "screen" with idle indicator (e.g., "idle" or "input prompt" visible).

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)
	assertOutput(t, resp, "stdout", "screen", "idle")
}
```
