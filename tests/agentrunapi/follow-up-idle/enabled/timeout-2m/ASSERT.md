## Expected

- `NormalizeIdle(true, 2m)` succeeds with enabled=true and duration `2m`.
- No emit error.
- Follow-up line contains `--exit-on-idle` and exact token `--idle-timeout=2m`
  (not `10m`, not `2m0s`) before `--`.
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
	assertNormalized(t, resp, true, 2*time.Minute)
	assertEmitsIdle(t, resp.FollowUp, "--idle-timeout=2m")
	if hasExactToken(resp.FollowUp, "--idle-timeout=10m") {
		t.Fatalf("explicit 2m must not emit default 10m; got %q", resp.FollowUp)
	}
	if hasExactToken(resp.FollowUp, "--idle-timeout=2m0s") {
		t.Fatalf("must emit compact 2m, not 2m0s; got %q", resp.FollowUp)
	}
	assertOpenProfile(t, resp.FollowUp, "sess-idle-on-2m")
}
```
