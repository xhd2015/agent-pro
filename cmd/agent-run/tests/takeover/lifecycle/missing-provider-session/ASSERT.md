## Expected

- Exit code non-zero.
- Action body is not the P1 stub alone: error indicates provider session not found
  under GROK_HOME (or equivalent missing/unknown session wording).
- No iTerm script written.
- No kill log entries.
- No new agent-run session meta created for the provider id.

## Exit Code

non-zero

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero exit for missing provider session, got 0\nstdout:\n%s\nstderr:\n%s",
			resp.Stdout, resp.Stderr)
	}
	combined := combinedOut(resp)
	assertTakeoverActionImplemented(t, combined)
	assertContainsAny(t, combined,
		"not found",
		"no such",
		"unknown session",
		"session not found",
		"missing",
		"does not exist",
	)
	assertNoItermScript(t, req)
	assertNoKillLog(t, req)
	// Soft: store should remain empty of real sessions.
	if ids := listAgentSessionIDs(t, req.Home); len(ids) > 0 {
		// Meta create on missing-session is a contract violation.
		t.Fatalf("missing provider session must not create agent-run meta; sessions=%v", ids)
	}
	_ = strings.ToLower(combined)
}
```
