## Expected

- `ReadIdlePolicy` returns a non-empty error for truncated JSON.

## Side Effects

- Raw `{` file exists under the session dir.

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
		t.Fatal("ReadIdlePolicy must error on truncated JSON")
	}
}
```
