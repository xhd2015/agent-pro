## Expected

- `Screen` is `banner` (trimmed value after colon).
- `Sendable` is `yes`.
- `Ready` is true.

## Errors

- None.

## Exit Code

N/A

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	assertEqual(t, "Screen", resp.Screen, "banner")
	assertEqual(t, "Sendable", resp.Sendable, "yes")
	assertEqual(t, "Ready", resp.Ready, true)
}
```
