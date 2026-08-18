## Expected

- `SampleIsIdle` is false (not sendable).

## Side Effects

- None (pure).

## Errors

- None.

## Exit Code

N/A

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	_ = d
	_ = req
	assertNoError(t, err)
	if resp.Idle {
		t.Fatal("not-sendable sample must not be idle")
	}
}
```
