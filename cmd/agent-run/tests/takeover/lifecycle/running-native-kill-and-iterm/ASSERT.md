## Expected

- Exit code 0.
- Stdout/stderr mentions killed pid `9101` (and preferably `grok`).
- Kill log records pid `9101` (TERM and/or KILL).
- iTerm ForceNew script written with agent-run follow-up.
- Session/provider summary lines acceptable (`session-id:`, `provider:`, opened window).

## Exit Code

0

```go
import (
	"strconv"
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	if err != nil {
		t.Fatal(err)
	}
	const nativePID = 9101
	combined := combinedOut(resp)
	assertTakeoverActionImplemented(t, combined)
	assertExitCode(t, resp, 0)

	if !strings.Contains(combined, strconv.Itoa(nativePID)) {
		t.Fatalf("expected output to mention killed pid %d, got:\n%s", nativePID, combined)
	}
	assertContainsAny(t, combined, "kill", "killed")
	assertContainsAny(t, combined, "grok", "pid")

	assertKillLogMentionsPID(t, req, nativePID)

	script := readItermScript(t, req)
	assertItermForceNewScript(t, script)
	assertContainsAny(t, script, "agent-run", "resume", "open", "--resume-from-grok-session", takeoverFixtureSessionID)

	assertContainsAny(t, combined, "session-id", "opened", "iterm", "iTerm", "provider")
}
```
