---
label: unit
explanation: listen appends inbound messages.jsonl for session history
---

## Expected

- `messages.jsonl` exists under session dir.
- At least one line with `direction` in (or inbound equivalent).
- Text includes stripped user text `log me please` (or original with mention).
- Message id/ts present (may equal Slack ts `1710000730.000100`).

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
		t.Fatalf("want 1 agent launch, got %d: %v", len(resp.AgentInvocations), resp.AgentInvocations)
	}
	sessionID := "slack-channel-" + slackTestChannelID
	msgs, readErr := readMessagesJSONL(t, req.HomeDir, sessionID)
	if readErr != nil {
		t.Fatalf("expected messages.jsonl: %v", readErr)
	}
	if len(msgs) < 1 {
		t.Fatalf("want >=1 log line, got %d", len(msgs))
	}
	found := false
	for _, m := range msgs {
		if m.Direction != "" && m.Direction != "in" {
			continue
		}
		if strings.Contains(m.Text, "log me please") {
			found = true
			if m.MessageID == "" && m.TS == "" {
				t.Fatalf("inbound log line missing message_id and ts: %+v", m)
			}
			if m.User != "" && m.User != slackTestUserID {
				t.Fatalf("inbound user = %q, want %q (or empty)", m.User, slackTestUserID)
			}
			break
		}
	}
	if !found {
		t.Fatalf("no inbound line with text containing %q: %+v", "log me please", msgs)
	}
}
```
