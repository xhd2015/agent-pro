---
label: unit
explanation: session info human keys include agent_session_id, message_count, session_dir
---

## Expected

- Exit code 0.
- Human `key: value` lines (stable order) for session fields + derived counts/paths.
- Empty `dir` shown as `-`.
- `session_id` and `agent_session_id` both present and equal.
- Trailing newline; stderr empty.

## Exit Code

0

```go
import (
	"fmt"
	"path/filepath"
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
	sessionDir := filepath.Join(req.HomeDir, filepath.FromSlash(defaultSlackLocalBotRelDir), "sessions", sessionInfoFixtureID)
	want := fmt.Sprintf(`session_id: %s
agent_session_id: %s
channel_id: %s
thread_ts: 1710000900.000100
config_path: /tmp/slack-info-cfg.json
dir: -
kind: channel
reply_mode: channel
created_at: 2026-07-10T12:00:00Z
updated_at: 2026-07-13T08:00:00Z
last_message_preview: info preview
message_count: 2
session_dir: %s
`, sessionInfoFixtureID, sessionInfoFixtureID, slackTestChannelID, sessionDir)
	if resp.Stdout != want {
		t.Fatalf("stdout mismatch\nwant:\n%s\ngot:\n%s", want, resp.Stdout)
	}
}
```
