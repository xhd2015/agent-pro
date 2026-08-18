## Expected

- `ReadIdlePolicy` returns a non-empty error (`nope` is not a duration).

## Side Effects

- Seed file exists.

## Errors

- Read error required.

## Exit Code

N/A

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	_ = d
	_ = req
	assertNoError(t, err)
	if resp.ErrString == "" {
		t.Fatal("ReadIdlePolicy must error on idle_timeout=nope")
	}
}
```
