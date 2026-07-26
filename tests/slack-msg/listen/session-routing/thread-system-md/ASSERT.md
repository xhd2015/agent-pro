---
label: unit
explanation: thread mode writes SYSTEM.md with session reply/history recipes
---

## Expected

- SYSTEM.md path:
  `$HOME/.agent-pro/slack-local-bot/sessions/slack-channel-{channelID}/SYSTEM.md`
- File exists after first open.
- Contents include:
  - session id
  - channel id
  - thread_ts
  - `slack-msg session history` recipe
  - `slack-msg session history --after-msg-id` recipe (or after-msg-id flag)
  - `slack-msg session reply` recipe
- Contents must **not** include:
  - raw bot token (`xoxb-slacktest-token` / `xoxb-`)
  - raw `slack-msg send --channel` / `send --thread` recipes
  - `slack-msg history --channel` / `history --thread` as the primary recipe

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
		t.Fatal(err)
	}
	if len(resp.AgentInvocations) != 1 {
		t.Fatalf("want 1 agent launch, got %d: %v", len(resp.AgentInvocations), resp.AgentInvocations)
	}
	sessionID := "slack-channel-" + slackTestChannelID
	path := expectedSessionSystemMDPath(req.HomeDir, sessionID)
	data, readErr := os.ReadFile(path)
	if readErr != nil {
		t.Fatalf("expected SYSTEM.md at %s: %v\nagent: %v", path, readErr, resp.AgentInvocations)
	}
	body := string(data)
	for _, want := range []string{
		sessionID,
		slackTestChannelID,
		"1710000700.000100",
		"slack-msg session history",
		"slack-msg session reply",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("SYSTEM.md missing %q\npath=%s\nbody:\n%s", want, path, body)
		}
	}
	if !strings.Contains(body, "after-msg-id") && !strings.Contains(body, "--after-msg-id") {
		t.Fatalf("SYSTEM.md should document session history --after-msg-id\npath=%s\nbody:\n%s", path, body)
	}
	if strings.Contains(body, slackTestBotToken) || strings.Contains(body, "xoxb-slacktest") {
		t.Fatalf("SYSTEM.md must not embed bot token secrets:\n%s", body)
	}
	// No raw channel/thread send/history recipes (SeaTalk session CLI).
	for _, forbidden := range []string{
		"slack-msg send --channel",
		"slack-msg send --thread",
		"send --channel",
		"history --channel",
		"history --thread",
	} {
		if strings.Contains(body, forbidden) {
			t.Fatalf("SYSTEM.md must not teach raw %q recipe:\n%s", forbidden, body)
		}
	}
}
```
