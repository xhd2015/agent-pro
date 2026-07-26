---
label: unit
explanation: open inject markers + strip bot mention + SYSTEM.md path in prompt
---

## Expected

- One agent launch (interactive open profile still used).
- Launch argv/prompt line includes open-inject markers:
  - `Slack listen session open` (or equivalent session-open header)
  - `session-id:` with `slack-channel-{channelID}`
  - `channel:`
  - `thread_ts:`
  - `from:` including user id (display name preferred)
  - `Instructions:` path to SYSTEM.md
  - `User message:` section with stripped text `clean me`
- Prompt must **not** contain raw `<@U023BECGF>` (bot self-mention).
- Still uses `--open` / `--auto-send-or-resume` / `--session-id=`.

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
	if len(resp.AgentInvocations) != 1 {
		t.Fatalf("want 1 invocation, got %d: %v", len(resp.AgentInvocations), resp.AgentInvocations)
	}
	line := resp.AgentInvocations[0]
	sessionID := "slack-channel-" + slackTestChannelID
	sysPath := expectedSessionSystemMDPath(req.HomeDir, sessionID)
	rawMention := "<@" + slackTestBotUserID + ">"

	for _, want := range []string{
		"run",
		"--session-id=" + sessionID,
		"--auto-send-or-resume",
		"--open",
		"session-id",
		sessionID,
		"channel",
		slackTestChannelID,
		"thread_ts",
		"1710000710.000100",
		"from:",
		slackTestUserID,
		"Instructions:",
		sysPath,
		"User message",
		"clean me",
	} {
		if !strings.Contains(line, want) {
			t.Fatalf("agent inject missing %q in %q", want, line)
		}
	}
	// Header phrase preferred; allow "session open" if wording differs slightly.
	if !strings.Contains(line, "Slack listen session open") && !strings.Contains(strings.ToLower(line), "session open") {
		t.Fatalf("expected session open header in inject prompt: %q", line)
	}
	if strings.Contains(line, rawMention) {
		t.Fatalf("agent prompt must strip bot mention %q; got %q", rawMention, line)
	}
	if strings.Contains(line, "--keep-tty") {
		t.Fatalf("interactive open must not use --keep-tty: %q", line)
	}
}
```
