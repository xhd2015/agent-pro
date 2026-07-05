## Expected

- First stdout line matches `session-id: session-N` (auto-reserved id).
- Parent exits 0 within 5 seconds.
- No PTY bytes or escape sequences on host stdout.

## Exit Code

0

```go
import (
	"strings"
	"testing"
	"time"

	"github.com/xhd2015/agent-pro/script/tty-watch/ttywatchtest"
	"github.com/xhd2015/doctest/assert"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("expected exit 0, got %d stderr=%q", resp.ExitCode, resp.Stderr)
	}
	if resp.Elapsed > 5*time.Second {
		t.Fatalf("detach parent should exit promptly, took %s", resp.Elapsed)
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
	if !strings.HasPrefix(id, "session-") {
		t.Fatalf("expected auto session-N id, got %q", id)
	}
	if strings.Contains(resp.Stdout, "DETACH") || ttywatchtest.ContainsANSIEscape(resp.Stdout) {
		t.Fatalf("host stdout leaked PTY/escape bytes: %q", resp.Stdout)
	}
}
```