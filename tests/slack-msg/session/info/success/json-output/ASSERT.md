---
label: unit
explanation: session info --json object with session_id, agent_session_id, message_count, session_dir
---

## Expected

- Exit code 0.
- JSON object includes both session id fields, map fields, message_count, session_dir.
- Empty dir is `""`.
- Trailing newline; stderr empty.

## Exit Code

0

```go
import (
	"encoding/json"
	"path/filepath"
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
		MessageCount       int    `json:"message_count"`
		SessionDir         string `json:"session_dir"`
	}
	if err := json.Unmarshal([]byte(resp.Stdout), &doc); err != nil {
		t.Fatalf("stdout is not JSON: %v\n%s", err, resp.Stdout)
	}
	if doc.SessionID != sessionInfoFixtureID {
		t.Fatalf("session_id = %q, want %q", doc.SessionID, sessionInfoFixtureID)
	}
	if doc.AgentSessionID != doc.SessionID {
		t.Fatalf("agent_session_id = %q, want equal to session_id %q", doc.AgentSessionID, doc.SessionID)
	}
	if doc.ChannelID != slackTestChannelID {
		t.Fatalf("channel_id = %q, want %q", doc.ChannelID, slackTestChannelID)
	}
	if doc.Dir != "" {
		t.Fatalf("dir = %q, want empty string", doc.Dir)
	}
	if doc.MessageCount != 2 {
		t.Fatalf("message_count = %d, want 2", doc.MessageCount)
	}
	wantDir := filepath.Join(req.HomeDir, filepath.FromSlash(defaultSlackLocalBotRelDir), "sessions", sessionInfoFixtureID)
	if doc.SessionDir != wantDir {
		t.Fatalf("session_dir = %q, want %q", doc.SessionDir, wantDir)
	}
	if doc.LastMessagePreview != "info preview" {
		t.Fatalf("last_message_preview = %q, want info preview", doc.LastMessagePreview)
	}
	if resp.Stdout == "" || resp.Stdout[len(resp.Stdout)-1] != '\n' {
		t.Fatalf("stdout must end with trailing newline, got %q", resp.Stdout)
	}
}
```
