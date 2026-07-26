## Expected

- API error mentioning session.
- Zero StatusFn polls (validation before poll).

## Side Effects

- None.

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
	assertEqual(t, "StatusPollCalls", resp.StatusPollCalls, 0)
}
```
