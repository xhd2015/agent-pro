# Scenario

**Feature**: SSE tail of `events.jsonl` delivers new session events to the SPA

```
POST session -> GET .../events/stream?after=0 -> SSE data lines (user + assistant events)
```

## Preconditions

- Route `GET /api/agent-run/sessions/:runner/:id/events/stream?after=<byte_offset>` exists.
- Each SSE message is `data: <json>\n\n` for one `AgentEvent` line.

## Steps

1. Leaf sets `req.Mode = "sse"`, starts web, creates session.
2. `Run` opens SSE subscription and collects parsed event payloads until session finishes or minimum events received.

```go
import (
	"fmt"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	if req.Mode != "" && req.Mode != "sse" {
		return fmt.Errorf("web/stream group: unexpected Mode %q", req.Mode)
	}
	req.Mode = "sse"
	if req.WebTokenMode == "" {
		req.WebTokenMode = "explicit"
	}
	if req.WebTokenMode == "explicit" && req.WebToken == "" {
		req.WebToken = "test"
	}
	return nil
}

func sseHasUserPrompt(events []map[string]any, prompt string) bool {
	for _, ev := range events {
		if ev["type"] == "message" && ev["role"] == "user" && ev["text"] == prompt {
			return true
		}
	}
	return false
}

func sseHasAssistantMessage(events []map[string]any) bool {
	for _, ev := range events {
		if ev["type"] == "message" && ev["role"] == "assistant" {
			return true
		}
	}
	return false
}
```