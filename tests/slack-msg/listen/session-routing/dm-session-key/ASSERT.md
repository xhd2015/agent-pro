---
label: unit
explanation: DM session id is slack-dm-{userID}, not slack-channel-{D channel}
---

## Expected

- Exactly one launch invocation.
- Argv contains `--session-id=slack-dm-{userID}` (fixture user `W012A3CDE`).
- Argv must **not** use `slack-channel-{DM channel id}` (`slack-channel-D…`).
- Argv must **not** use legacy `slack-{channel}-{ts}` form.
- Interactive open profile still used (`run`, `--auto-send-or-resume`, `--open`).

## Exit Code

0

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.AgentInvocations) != 1 {
		t.Fatalf("want 1 invocation, got %d: %v", len(resp.AgentInvocations), resp.AgentInvocations)
	}
	line := resp.AgentInvocations[0]
	sessionID := "slack-dm-" + slackTestUserID
	wrongChannelKey := "slack-channel-" + slackTestDMChannelID
	legacyTS := "slack-" + slackTestDMChannelID + "-1710001100.000100"
	for _, want := range []string{
		"run",
		"--session-id=" + sessionID,
		"--auto-send-or-resume",
		"--open",
	} {
		if !strings.Contains(line, want) {
			t.Fatalf("agent argv missing %q in %q", want, line)
		}
	}
	if strings.Contains(line, "--session-id="+wrongChannelKey) {
		t.Fatalf("DM must not use channel-key session id %q: %q", wrongChannelKey, line)
	}
	if strings.Contains(line, "--session-id="+legacyTS) {
		t.Fatalf("DM must not use legacy per-ts session id: %q", line)
	}
	if strings.Contains(line, "slack-channel-"+slackTestDMChannelID) {
		t.Fatalf("DM session must not be keyed by DM channel id: %q", line)
	}
}
```
