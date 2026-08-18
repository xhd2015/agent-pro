## Expected

- `SampleIsIdle` is false (queue length > 0).

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
		t.Fatal("queue length 1 must not be idle")
	}
}
```
