## Expected

- `EnsureCodexRunnerBound` returns `bound=true`.
- Returned `meta.runner_session_id` equals the matching rollout Codex id
  (`019fef17-ea39-7623-9ef6-b2376b1556c0`).
- Store re-read shows the same `runner_session_id` (persisted, not return-only).

## Side Effects

- `store.UpdateSessionRunnerSessionID` wrote meta under `sessions/<id>/meta.json`.

## Errors

- None (best-effort API; harness error must be nil).

## Exit Code

- N/A (library)

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	assertNoError(t, err)
	assertBoundPersisted(t, resp, fixtureCodexIDMatching)
}
```
