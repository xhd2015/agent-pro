## Expected

- Exit code 0.
- Combined output contains `from-config` (first curl) and `from-events` (second curl from events file).

## Exit Code

0

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)
	combined := resp.Stdout + resp.Stderr
	assertContains(t, combined, "from-config")
	assertContains(t, combined, "from-events")
}
```
