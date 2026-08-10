## Expected

- Exit code 0.
- Stdout/stderr mentions killed pid `9201` (and preferably `codex`).
- Kill log records pid `9201` (TERM and/or KILL).
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
	const nativePID = 9201
	combined := combinedOut(resp)
	assertTakeoverActionImplemented(t, combined)
	assertExitCode(t, resp, 0)

	if !strings.Contains(combined, strconv.Itoa(nativePID)) {
		t.Fatalf("expected output to mention killed pid %d, got:\n%s", nativePID, combined)
	}
	assertContainsAny(t, combined, "kill", "killed")
	assertContainsAny(t, combined, "codex", "pid")

	assertKillLogMentionsPID(t, req, nativePID)

	script := readItermScript(t, req)
	assertItermForceNewScript(t, script)
	assertContainsAny(t, script, "agent-run", "resume", "open", takeoverCodexFixtureSessionID)

	assertContainsAny(t, combined, "session-id", "opened", "iterm", "iTerm", "provider")
}
```
