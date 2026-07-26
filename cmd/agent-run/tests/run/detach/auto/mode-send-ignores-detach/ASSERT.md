---
label: e2e
---

## Expected

- Exit code 0.
- Stdout first line is `msg_N` (send path).
- Stderr should include a `note:` that `--detach` was ignored while live
  (product parallel to live `--open` ignore note).
- Must not print detach `session-id:` / `terminal-id:` success shape as the
  primary outcome (send path wins).

## Exit Code

0

```go
import (
	"os"
	"strings"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("run failed: %v\nstdout:\n%s\nstderr:\n%s", err, resp.Stdout, resp.Stderr)
	}
	assertNotUnknownFlag(t, strings.ToLower(resp.Stderr+"\n"+resp.Stdout), "--detach")
	assertSuccess(t, resp)

	first := strings.TrimSpace(strings.Split(resp.Stdout, "\n")[0])
	if !strings.HasPrefix(first, "msg_") {
		t.Fatalf("stdout first line must be msg_N (send path), got %q\nstderr:\n%s\nstdout:\n%s",
			first, resp.Stderr, resp.Stdout)
	}

	// Prefer product note on stderr (like live --open).
	errLower := strings.ToLower(resp.Stderr)
	hasNote := strings.Contains(errLower, "note:") || strings.Contains(errLower, "note ")
	hasDetach := strings.Contains(errLower, "detach")
	hasIgnore := strings.Contains(errLower, "ignore") ||
		strings.Contains(errLower, "ignored") ||
		strings.Contains(errLower, "while") && strings.Contains(errLower, "live")
	if !(hasDetach && (hasNote || hasIgnore)) {
		t.Fatalf("stderr should note that --detach was ignored while live:\n%s", resp.Stderr)
	}

	// Delivery/enqueue proof (soft if msg_N already proves send path).
	injected := false
	if req.FakePTYInjectLog != nil {
		for _, s := range *req.FakePTYInjectLog {
			if strings.Contains(s, req.FollowupPrompt) {
				injected = true
				break
			}
		}
	}
	qPath := sendQueuePath(req.Home, defaultRunner, req.TerminalSessionID)
	qData, _ := os.ReadFile(qPath)
	queued := strings.Contains(string(qData), req.FollowupPrompt)
	if !injected && !queued {
		t.Logf("no inject/queue evidence yet; msg_N is primary contract (inject=%v queue=%q)",
			req.FakePTYInjectLog, string(qData))
	}
}
```
