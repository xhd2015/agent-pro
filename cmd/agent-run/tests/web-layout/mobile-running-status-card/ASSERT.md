---
label: chromium
explanation: playwright mobile session running card
---

## Expected

- `playwright-debug` exits 0.
- Viewport 390×844; no horizontal document scroll.
- `[data-testid="agent-running-card"]` visible above the message list.
- `[data-testid="agent-running-duration"]` visible with non-empty text containing a digit and a time cue (`:`, `m`, `s`, or “running”/“for”).
- `[data-testid="composer"]` remains pinned to the viewport bottom.

## Side Effects

- Background `agent-run web` started during Setup and stopped on cleanup.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("Run error: %v\nstderr:\n%s", err, resp.PlaywrightStderr)
	}
	if resp.PlaywrightExit != 0 {
		t.Fatalf("playwright-debug exit %d\nstdout:\n%s\nstderr:\n%s",
			resp.PlaywrightExit, resp.PlaywrightStdout, resp.PlaywrightStderr)
	}
	if req.Layout != "running-card" {
		t.Fatalf("expected layout running-card, got %q", req.Layout)
	}
	if strings.TrimSpace(resp.PlaywrightStderr) != "" {
		t.Logf("playwright stderr: %s", resp.PlaywrightStderr)
	}
}
```