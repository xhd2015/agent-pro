## Expected

- `NormalizeIdle(true, 2s)` succeeds with enabled=true and duration `2s`.
- No emit error.
- Follow-up line contains `--exit-on-idle` and exact token `--idle-timeout=2s`
  before `--`.
- Open profile still present; no `--new-terminal`.

## Side Effects

- None (pure).

## Errors

- None.

## Exit Code

N/A

```go
import (
	"testing"
	"time"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	_ = d
	_ = req
	assertNoError(t, err)
	assertNoAPIError(t, resp)
	assertNormalized(t, resp, true, 2*time.Second)
	assertEmitsIdle(t, resp.FollowUp, "--idle-timeout=2s")
	assertOpenProfile(t, resp.FollowUp, "sess-idle-on-2s")
}
```
