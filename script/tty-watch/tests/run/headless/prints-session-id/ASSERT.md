## Expected

- First stdout line matches `session-id: session-N` (auto-reserved id) or custom id when set.
- No PTY bytes or escape sequences on host stdout after the session-id line.

## Exit Code

0 (headless parent still running; harness terminates in cleanup)

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/agent-pro/script/tty-watch/ttywatchtest"
	"github.com/xhd2015/doctest/assert"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	firstLine := strings.TrimSpace(strings.Split(resp.Stdout, "\n")[0])
	assert.Output(t, firstLine, `---
version: 2
__SESSION_ID__: type=string, example=session-1, reserved registry id
---
session-id: __SESSION_ID__`)
	if !strings.HasPrefix(firstLine, "session-id: ") {
		t.Fatalf("expected session-id prefix, got %q", firstLine)
	}
	id := strings.TrimPrefix(firstLine, "session-id: ")
	if req.CustomSessionID != "" && id != req.CustomSessionID {
		t.Fatalf("expected custom id %q, got %q", req.CustomSessionID, id)
	}
	if req.CustomSessionID == "" && !strings.HasPrefix(id, "session-") {
		t.Fatalf("expected auto session-N id, got %q", id)
	}
	if strings.Contains(resp.Stdout, "HEADLESS") || ttywatchtest.ContainsANSIEscape(resp.Stdout) {
		t.Fatalf("host stdout leaked PTY/escape bytes: %q", resp.Stdout)
	}
}
```