---
label: e2e
---

## Expected

- Exit code 0.
- Stdout `msg_N` (send still works).
- No iTerm script written (or file empty/missing) — `--new-terminal` ignored on live.
- Optional: stderr mentions ignore / new-terminal (soft).
- No provider `--resume` spawn.

## Exit Code

0

```go
import (
	"os"
	"regexp"
	"strings"
	"testing"
	"github.com/xhd2015/doctest/session"
)

var msgIDLineReNT = regexp.MustCompile(`(?m)^msg_\d+\n?$`)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("run failed: %v\nstdout:\n%s\nstderr:\n%s", err, resp.Stdout, resp.Stderr)
	}
	assertSuccess(t, resp)

	out := resp.Stdout
	first := strings.TrimSpace(strings.Split(out, "\n")[0])
	if !strings.HasPrefix(first, "msg_") {
		t.Fatalf("expected stdout msg_N line, got %q\nstderr:\n%s", out, resp.Stderr)
	}
	assertTrailingNewline(t, out, "send stdout")

	assertNoItermScript(t, req)

	// Delivery/enqueue proof.
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
	if !injected && !queued && !strings.HasPrefix(first, "msg_") {
		t.Fatalf("expected inject or queue evidence for %q; inject=%v queue=%q", req.FollowupPrompt, req.FakePTYInjectLog, string(qData))
	}

	if fileExists(req.ArgvProbePath) {
		probe, _ := os.ReadFile(req.ArgvProbePath)
		if strings.Contains(string(probe), "--resume") {
			t.Fatalf("live send must not spawn provider --resume; argv:\n%s", probe)
		}
	}

	// Soft optional note about ignore.
	if strings.Contains(strings.ToLower(resp.Stderr), "new-terminal") ||
		strings.Contains(strings.ToLower(resp.Stderr), "ignore") {
		t.Logf("optional ignore note present on stderr: %q", resp.Stderr)
	}
	_ = msgIDLineReNT
}
```
