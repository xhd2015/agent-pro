## Expected

- `Screen` is `banner`.
- `Sendable` is `yes`.
- `Ready` is true.

## Side Effects

- None (pure).

## Errors

- None.

## Exit Code

N/A

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	assertEqual(t, "Screen", resp.Screen, "banner")
	assertEqual(t, "Sendable", resp.Sendable, "yes")
	assertEqual(t, "Ready", resp.Ready, true)
}
```
