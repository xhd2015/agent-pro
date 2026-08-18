## Expected

- `SampleIsIdle` is false (screen unknown).

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
		t.Fatal("screen unknown must not be idle")
	}
}
```
