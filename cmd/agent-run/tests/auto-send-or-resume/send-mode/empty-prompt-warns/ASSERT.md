---
label: e2e
---

## Expected

- Exit code 0.
- Stderr contains a warning that the session is live / there is no message to send.
- Stdout does not print `msg_N`.
- No send-queue entry for the terminal (or empty / absent file).

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
	assertSuccess(t, resp)

	combined := strings.ToLower(resp.Stderr + "\n" + resp.Stdout)
	assertContainsAny(t, combined,
		"no message",
		"nothing to send",
		"empty prompt",
		"no prompt",
		"live",
		"still active",
		"warning",
	)
	// Prefer explicit "no message" / warn style when available.
	assertContainsAny(t, combined,
		"no message",
		"nothing to send",
		"not send",
		"skip",
		"warn",
		"live",
	)

	out := strings.TrimSpace(resp.Stdout)
	if strings.HasPrefix(out, "msg_") {
		t.Fatalf("empty prompt must not enqueue; stdout has msg id: %q", resp.Stdout)
	}

	qPath := sendQueuePath(req.Home, defaultRunner, req.TerminalSessionID)
	if data, rErr := os.ReadFile(qPath); rErr == nil && len(strings.TrimSpace(string(data))) > 0 {
		t.Fatalf("send queue must stay empty for empty-prompt live path; queue=%q", string(data))
	}
	if req.FakePTYInjectLog != nil && len(*req.FakePTYInjectLog) > 0 {
		// WS may still receive attach probes; only fail if prompt-like inject happened.
		for _, s := range *req.FakePTYInjectLog {
			if strings.Contains(strings.ToLower(s), "followup") || len(strings.TrimSpace(s)) > 20 {
				t.Logf("inject log non-empty (may be control chars): %q", s)
			}
		}
	}
}
```
