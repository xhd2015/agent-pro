## Expected

- `Screen` is `starting`.
- `Sendable` is `no`.
- `Ready` is false.

## Errors

- None.

## Exit Code

N/A

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	assertEqual(t, "Screen", resp.Screen, "starting")
	assertEqual(t, "Sendable", resp.Sendable, "no")
	assertEqual(t, "Ready", resp.Ready, false)
}
```
