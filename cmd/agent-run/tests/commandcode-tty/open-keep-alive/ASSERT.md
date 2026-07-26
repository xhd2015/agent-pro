---
label: e2e
---

## Expected

- Exit code 0.
- Stderr matches `commandcode-tty: session-N`.
- After 10s, snapshot returns non-empty text containing `"Hello"`.

## Exit Code

0

```go
import (
	"strings"
	"testing"
	"time"

	"github.com/xhd2015/doctest/assert"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("expected exit 0, got %d\nstderr:\n%s", resp.ExitCode, resp.Stderr)
	}

	assert.Output(t, resp.Stderr, "" +
`<contains>
<regex>commandcode-tty:\s*session-\d+</regex>
</contains>`)

	sessionID := resp.SessionID
	if sessionID == "" {
		t.Fatalf("failed to extract session id from stderr:\n%s", resp.Stderr)
	}

	time.Sleep(10 * time.Second)
	snapshot := execSnapshot(t, req.AgentRun, sessionID)

	if strings.TrimSpace(snapshot) == "" {
		t.Fatalf("snapshot is empty for session %s", sessionID)
	}

	if !strings.Contains(snapshot, "Hello") {
		t.Fatalf("snapshot missing 'Hello', session %s:\n%s", sessionID, snapshot)
	}

	resp.Snapshot = snapshot
}
```
