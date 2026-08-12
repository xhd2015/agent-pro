## Expected

- `bound=false`.
- Returned and stored `runner_session_id` remain empty.
- Matching codex fixtures must not be applied to a non-codex runner.

## Side Effects

- No bind persist.

## Errors

- None (no panic / no hard fail).

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
