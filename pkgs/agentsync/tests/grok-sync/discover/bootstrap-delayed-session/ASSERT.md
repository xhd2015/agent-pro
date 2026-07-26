## Expected

- `events.jsonl` contains user message with `initial_prompt` text.
- Assistant reply `discover-bootstrap-reply-marker` present.
- At least one `done` event.
- `meta.json` `runner_session_id` equals `bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb`.
- `grok-sync.json` checkpoint exists with discovered `updates_path`.

## Errors

- Empty `events.jsonl` after timeout indicates discovery bootstrap not implemented
  (PRIMARY RED before fix).

## Exit Code

N/A (direct package call)

```go
import (
	"os"
	"strings"
	"testing"

	types "github.com/xhd2015/agent-pro/agent/event/types"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.EnsureErr != nil {
		t.Fatalf("EnsureGrokSync: %v", resp.EnsureErr)
	}
	if count := countUserMessagesByText(resp.Events, discoverBootstrapPrompt); count != 1 {
		t.Fatalf("user prompt %q: want 1 got %d; events=%d", discoverBootstrapPrompt, count, len(resp.Events))
	}
	foundReply := false
	for _, ev := range resp.Events {
		if ev.Type == types.ActionMessage && ev.Role == "assistant" && strings.Contains(ev.Text, discoverBootstrapReply) {
			foundReply = true
			break
		}
	}
	if !foundReply {
		t.Fatalf("missing assistant reply %q; events=%d", discoverBootstrapReply, len(resp.Events))
	}
	if countActionDone(resp.Events) < 1 {
		t.Fatalf("want >= 1 ActionDone, got %d", countActionDone(resp.Events))
	}
	if resp.RunnerSessionID != discoverBootstrapGrokUUID {
		t.Fatalf("runner_session_id: got %q want %q", resp.RunnerSessionID, discoverBootstrapGrokUUID)
	}
	if _, err := os.Stat(grokSyncJSONPath(req.SessionDir)); err != nil {
		t.Fatalf("grok-sync.json missing after discovery: %v", err)
	}
	cp := readGrokSyncCheckpoint(t, req.SessionDir)
	if strings.TrimSpace(cp.UpdatesPath) == "" {
		t.Fatal("checkpoint updates_path empty after discovery")
	}
}
```