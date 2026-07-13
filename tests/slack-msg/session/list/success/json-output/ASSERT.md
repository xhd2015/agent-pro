---
label: unit
explanation: session list --json emits sessions with session_id and agent_session_id
---

## Expected

- Exit code 0.
- JSON document `{"sessions":[…]}` sorted by `updated_at` desc.
- Each session includes both `session_id` and `agent_session_id` (equal today).
- Empty `dir` is `""`.
- Trailing newline; stderr empty.

## Exit Code

0

```go
import (
	"encoding/json"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertExitCode(t, resp, 0)
	if resp.Stderr != "" {
		t.Fatalf("expected empty stderr, got:\n%s", resp.Stderr)
	}
	var doc struct {
		Sessions []struct {
			SessionID          string `json:"session_id"`
			AgentSessionID     string `json:"agent_session_id"`
			ChannelID          string `json:"channel_id"`
			ThreadTS           string `json:"thread_ts"`
			ConfigPath         string `json:"config_path"`
			Dir                string `json:"dir"`
			Kind               string `json:"kind"`
			ReplyMode          string `json:"reply_mode"`
			CreatedAt          string `json:"created_at"`
			UpdatedAt          string `json:"updated_at"`
			LastMessagePreview string `json:"last_message_preview"`
		} `json:"sessions"`
	}
	if err := json.Unmarshal([]byte(resp.Stdout), &doc); err != nil {
		t.Fatalf("stdout is not JSON: %v\n%s", err, resp.Stdout)
	}
	if len(doc.Sessions) != 2 {
		t.Fatalf("sessions len = %d, want 2: %+v", len(doc.Sessions), doc.Sessions)
	}
	// Sorted updated_at desc: newer channel first.
	s0, s1 := doc.Sessions[0], doc.Sessions[1]
	if s0.SessionID != sessionListNewerID {
		t.Fatalf("sessions[0].session_id = %q, want %q", s0.SessionID, sessionListNewerID)
	}
	if s1.SessionID != sessionListOlderID {
		t.Fatalf("sessions[1].session_id = %q, want %q", s1.SessionID, sessionListOlderID)
	}
	for i, s := range doc.Sessions {
		if s.AgentSessionID == "" {
			t.Fatalf("sessions[%d] missing agent_session_id", i)
		}
		if s.AgentSessionID != s.SessionID {
			t.Fatalf("sessions[%d] agent_session_id %q != session_id %q", i, s.AgentSessionID, s.SessionID)
		}
	}
	if s0.ChannelID != "C01ABCDEFF0" || s0.LastMessagePreview != "hello from slack" {
		t.Fatalf("sessions[0] unexpected fields: %+v", s0)
	}
	if s0.Dir != "" {
		t.Fatalf("sessions[0].dir = %q, want empty string", s0.Dir)
	}
	if s1.Dir != "" {
		t.Fatalf("sessions[1].dir = %q, want empty string", s1.Dir)
	}
	if resp.Stdout == "" || resp.Stdout[len(resp.Stdout)-1] != '\n' {
		t.Fatalf("stdout must end with trailing newline, got %q", resp.Stdout)
	}
}
```
