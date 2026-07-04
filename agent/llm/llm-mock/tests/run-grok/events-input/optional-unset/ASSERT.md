## Expected

- Exit code 0.
- Output contains `from-config` from the sole config exchange.
- Output does not contain `from-events` (events file was not set).

## Exit Code

0

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)
	combined := resp.Stdout + resp.Stderr
	assertContains(t, combined, "from-config")
	assertNotContains(t, combined, "from-events")
}
```
