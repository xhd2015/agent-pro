## Expected

- `NormalizeIdle(false, 0)` succeeds with enabled=false and duration 0.
- No emit error.
- Follow-up line has neither `--exit-on-idle` nor any `--idle-timeout*` token.
- Open profile still present.

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
	assertNormalized(t, resp, false, time.Duration(0))
	assertOmitsIdleFlags(t, resp.FollowUp)
	assertOpenProfile(t, resp.FollowUp, "sess-idle-off-zero")
}
```
