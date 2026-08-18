## Expected

- `NormalizeIdle(true, 0)` succeeds with enabled=true and `DefaultIdleTimeout` (`10m`).
- No emit error.
- Follow-up line contains exact tokens `--exit-on-idle` and `--idle-timeout=10m`
  (not `10m0s`) before `--` / prompt.
- Session id and `--open` still present; no `--new-terminal`.

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

	"github.com/xhd2015/agent-pro/pkgs/agentrunapi"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	_ = d
	_ = req
	assertNoError(t, err)
	assertNoAPIError(t, resp)
	if agentrunapi.DefaultIdleTimeout != 10*time.Minute {
		t.Fatalf("DefaultIdleTimeout: got %s, want 10m", agentrunapi.DefaultIdleTimeout)
	}
	assertNormalized(t, resp, true, agentrunapi.DefaultIdleTimeout)
	assertEmitsIdle(t, resp.FollowUp, "--idle-timeout=10m")
	if hasExactToken(resp.FollowUp, "--idle-timeout=10m0s") {
		t.Fatalf("must emit compact 10m, not 10m0s; got %q", resp.FollowUp)
	}
	assertOpenProfile(t, resp.FollowUp, "sess-idle-on-default")
}
```
