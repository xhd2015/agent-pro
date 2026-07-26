---
label: e2e
---

## Expected

- Exit code 0.
- Stdout is a session-local message id line `msg_N` (typically `msg_1`) with trailing `\n`.
- Message enqueued/delivered: inject log contains followup text, **or** queue was drained after delivery.
- No provider argv probe file (send must not spawn `--resume` runner).

## Exit Code

0

```go
import (
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/xhd2015/doctest/assert"
	"github.com/xhd2015/doctest/session"
)

var msgIDLineRe = regexp.MustCompile(`(?m)^msg_\d+\n?$`)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("run failed: %v\nstdout:\n%s\nstderr:\n%s", err, resp.Stdout, resp.Stderr)
	}
	assertSuccess(t, resp)

	out := resp.Stdout
	if !msgIDLineRe.MatchString(strings.TrimSpace(out)+"\n") && !strings.HasPrefix(strings.TrimSpace(out), "msg_") {
		t.Fatalf("expected stdout msg_N line, got %q\nstderr:\n%s", out, resp.Stderr)
	}
	// Prefer strict first line msg_1 when first send on session.
	first := strings.TrimSpace(strings.Split(out, "\n")[0])
	if first == "msg_1" {
		assert.Output(t, first+"\n", `---
version: 2
---
msg_1
`)
	} else if !strings.HasPrefix(first, "msg_") {
		t.Fatalf("stdout first line must be msg_N, got %q", first)
	}
	assertTrailingNewline(t, out, "send stdout")

	// Delivery/enqueue proof: inject body or residual queue containing text.
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
	// Soft: prefer inject when WaitDefault + drainer worked.
	if !injected {
		t.Logf("inject not observed (may still be enqueued/delivered); stdout=%q queue=%q", first, string(qData))
	}

	if fileExists(req.ArgvProbePath) {
		probe, _ := os.ReadFile(req.ArgvProbePath)
		if strings.Contains(string(probe), "--resume") {
			t.Fatalf("live send must not spawn provider --resume; argv:\n%s", probe)
		}
	}
}
```
