## Expected

- `bound=false`.
- Returned and stored `runner_session_id` remain empty.
- Must not bind the other-workspace Codex id.

## Side Effects

- No `runner_session_id` write (or empty no-op).

## Errors

- None (best-effort miss is not an error).

## Exit Code

- N/A (library)

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	assertNoError(t, err)
	assertUnbound(t, resp)
}
```
