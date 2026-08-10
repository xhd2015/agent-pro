## Expected

- Exit code non-zero.
- Error indicates session/provider not found or cannot resolve (not only the
  empty-runner flag gate forever).
- No iTerm script; no kill log; no agent-run meta.

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
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero exit when neither provider has the id, got 0\nstdout:\n%s\nstderr:\n%s",
			resp.Stdout, resp.Stderr)
	}
	combined := combinedOut(resp)
	assertTakeoverActionImplemented(t, combined)
	lower := strings.ToLower(combined)
	// Reject the pre-auto-detect empty-runner message as the sole outcome.
	if strings.Contains(lower, "requires --grok") || strings.Contains(lower, "requires --codex") ||
		(strings.Contains(lower, "requires") && strings.Contains(lower, "agent-runner")) {
		t.Fatalf("want not-found after auto-detect lookup, still empty-runner flag gate:\n%s", combined)
	}
	assertContainsAny(t, combined,
		"not found",
		"no such",
		"unknown session",
		"session not found",
		"missing",
		"does not exist",
		"cannot resolve",
		"could not resolve",
		"unable to resolve",
		"no provider",
	)
	assertNoItermScript(t, req)
	assertNoKillLog(t, req)
	if ids := listAgentSessionIDs(t, req.Home); len(ids) > 0 {
		t.Fatalf("neither-hit must not create meta; sessions=%v", ids)
	}
}
```
