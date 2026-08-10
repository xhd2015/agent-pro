## Expected

- Exit code 0.
- Stderr/stdout `warning:` already managed by agent-run; nothing to take over.
- No kill log entries (must not kill the managed codex child).
- No iTerm script.

## Exit Code

0

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
	combined := combinedOut(resp)
	assertTakeoverActionImplemented(t, combined)
	assertExitCode(t, resp, 0)
	lower := strings.ToLower(combined)
	if !strings.Contains(lower, "warning") {
		t.Fatalf("expected warning: for already-managed process, got:\n%s", combined)
	}
	assertContainsAny(t, combined,
		"already managed",
		"already under agent-run",
		"nothing to take over",
		"managed by agent-run",
	)
	assertNoKillLog(t, req)
	assertNoItermScript(t, req)
}
```
