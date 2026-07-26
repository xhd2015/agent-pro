---
label: e2e
---

## Expected

- Exit code **0** (soft unbound is success for detach).
- Both `session-id:` and `terminal-id:` on stdout.
- Must **not** hard-fail with unresolved grok session wording.
- `meta.runner_session_id` may be empty / absent.

## Errors

- None for soft bind miss under detach.

## Exit Code

0

```go
import (
	"strings"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("run failed: %v\nstdout:\n%s\nstderr:\n%s", err, resp.Stdout, resp.Stderr)
	}
	combined := strings.ToLower(resp.Stderr + "\n" + resp.Stdout)
	assertNotUnknownFlag(t, combined, "--detach")

	// Hard-require path (open-style) must not fire on detach soft miss.
	if strings.Contains(combined, "not resolved") ||
		strings.Contains(combined, "session id not resolved") ||
		strings.Contains(combined, "grok session id not resolved") {
		t.Fatalf("detach soft bind miss must not hard-fail unresolved; got:\nstderr:\n%s\nstdout:\n%s",
			resp.Stderr, resp.Stdout)
	}
	assertSuccess(t, resp)
	sid, _ := assertDetachIDsOnStdout(t, resp)

	// Soft unbound: runner_session_id empty is OK.
	if resp.MetaAfter == nil && sid != "" {
		path := metaJSONPath(req.Home, sid)
		if fileExists(path) {
			resp.MetaAfter = readMetaJSON(t, req.Home, sid)
		}
	}
	if resp.MetaAfter != nil {
		if rs, _ := resp.MetaAfter["runner_session_id"].(string); strings.TrimSpace(rs) != "" {
			// Soft miss isolation should typically leave unbound; if a bind
			// somehow hit, still exit 0 is fine — log only.
			t.Logf("unexpected runner_session_id=%q under soft-miss isolation (still exit 0 OK)", rs)
		}
	}
}
```
