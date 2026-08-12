## Expected

- `EnsureCodexRunnerBound` returns `bound=true`.
- Returned and stored `runner_session_id` equal the resume footer id
  (`019fef18-0225-7d10-a08d-8250e433e045`).

## Side Effects

- Persist via `UpdateSessionRunnerSessionID`.

## Errors

- None.

## Exit Code

- N/A (library)

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	assertNoError(t, err)
	assertBoundPersisted(t, resp, fixtureCodexIDOther)
}
```
