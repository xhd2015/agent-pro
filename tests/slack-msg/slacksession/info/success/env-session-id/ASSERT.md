---
label: unit
explanation: session info uses SLACK_MSG_SESSION_ID without --session-id
---

## Expected

- Exit code 0.
- Stdout includes fixture session_id and message_count: 2.
- Stderr empty.

## Exit Code

0

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertExitCode(t, resp, 0)
	if resp.Stderr != "" {
		t.Fatalf("expected empty stderr, got:\n%s", resp.Stderr)
	}
	if !strings.Contains(resp.Stdout, "session_id: "+sessionInfoFixtureID) &&
		!strings.Contains(resp.Stdout, `"session_id":"`+sessionInfoFixtureID+`"`) {
		// human default
		if !strings.Contains(resp.Stdout, sessionInfoFixtureID) {
			t.Fatalf("stdout missing session id %q:\n%s", sessionInfoFixtureID, resp.Stdout)
		}
	}
	if !strings.Contains(resp.Stdout, "message_count: 2") &&
		!strings.Contains(resp.Stdout, `"message_count":2`) {
		t.Fatalf("stdout missing message_count 2:\n%s", resp.Stdout)
	}
	if !strings.Contains(resp.Stdout, "agent_session_id") {
		t.Fatalf("stdout missing agent_session_id:\n%s", resp.Stdout)
	}
}
```
