---
label: codex
explanation: Real Codex TUI via llm-mock-run-codex; verifies trust auto-accept and idle sendable.
---

## Expected

- Open exits 0.
- Snapshot does **not** show directory trust modal.
- Status is sendable, or snapshot shows idle Codex chrome with `›` and no trust text.

```go
import (
	"strings"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	if err != nil {
		t.Fatalf("open idle: %v\nsnapshot:\n%s\nstatus:\n%s", err, resp.Snapshot, resp.StatusText)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("open exit %d stderr:\n%s", resp.ExitCode, resp.Stderr)
	}
	plain := plainSnapshotText(resp.Snapshot)
	low := strings.ToLower(plain)
	if strings.Contains(low, "do you trust the contents") || strings.Contains(low, "yes, continue") {
		t.Fatalf("trust modal still visible after open:\n%s\nstatus:\n%s", plain, resp.StatusText)
	}
	if strings.Contains(strings.ToLower(resp.StatusText), "update available") {
		t.Fatalf("trust misclassified as update available:\n%s", resp.StatusText)
	}
	sendable := strings.Contains(resp.StatusText, "sendable: yes")
	hasPrompt := strings.Contains(plain, "›") || strings.Contains(plain, "OpenAI Codex")
	if !sendable && !hasPrompt {
		t.Fatalf("expected sendable idle prompt; status:\n%s\nsnapshot:\n%s", resp.StatusText, plain)
	}
}
```
