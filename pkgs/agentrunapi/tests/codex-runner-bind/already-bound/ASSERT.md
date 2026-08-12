## Expected

- `bound=true` (already bound counts as bound).
- Returned `meta.runner_session_id` remains the original id A
  (`aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee`).
- Store re-read still has A — must **not** become offer id B.

## Side Effects

- No overwrite of `runner_session_id` (A preserved).

## Errors

- None.

## Exit Code

- N/A (library)

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	assertNoError(t, err)
	assertBoundPersisted(t, resp, fixtureCodexIDBound)
	// Explicit no-overwrite guard against offer id.
	if strings.TrimSpace(resp.Meta.RunnerSessionID) == fixtureCodexIDOffer ||
		strings.TrimSpace(resp.StoredRunnerSessionID) == fixtureCodexIDOffer {
		t.Fatalf("runner_session_id was overwritten with offer id %q", fixtureCodexIDOffer)
	}
}
```
