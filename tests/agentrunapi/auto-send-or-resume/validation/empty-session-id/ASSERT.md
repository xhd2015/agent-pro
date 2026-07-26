## Expected

- API error whose message mentions session (id).
- Zero Run/Send/Resume hook calls (no side effects).

## Side Effects

- None (no classify dispatch, no session create).

## Errors

- Expected API error only.

## Exit Code

N/A

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	assertAPIError(t, resp)
	assertContainsFold(t, resp.ErrString, "session")
	assertZeroHooks(t, resp)
}
```
